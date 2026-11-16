package cloudtpl

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	textTemplate "text/template"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Masterminds/sprig/v3"
	"github.com/webdevops/go-common/log/slogger"

	"github.com/webdevops/helm-azure-tpl/pkg/config"
	"github.com/webdevops/helm-azure-tpl/pkg/engine/azure"
	"github.com/webdevops/helm-azure-tpl/pkg/engine/base"
	"github.com/webdevops/helm-azure-tpl/pkg/engine/general"
	"github.com/webdevops/helm-azure-tpl/pkg/engine/helm"
	"github.com/webdevops/helm-azure-tpl/pkg/engine/k8s"
)

const (
	EngineGeneral int = iota
	EngineHelm
	EngineAzure
	EngineKubernetes
)

type (
	CloudTpl struct {
		*base.BaseEngine
		engines map[int]base.Engine
	}
)

func New(ctx context.Context, opts config.Opts, logger *slogger.Logger) *CloudTpl {
	e := &CloudTpl{
		BaseEngine: &base.BaseEngine{
			Logger: logger,
			Opts:   opts,
			Ctx:    ctx,
			Cache:  globalCache,
		},
		engines: map[int]base.Engine{},
	}

	e.init()
	return e
}

func (e *CloudTpl) init() {
	e.engines = map[int]base.Engine{
		EngineGeneral:    general.New(e.BaseEngine.Clone()), // nolint:staticcheck // want to make obvious what we're cloning here
		EngineHelm:       helm.New(e.BaseEngine.Clone()),    // nolint:staticcheck // want to make obvious what we're cloning here
		EngineAzure:      azure.New(e.BaseEngine.Clone()),   // nolint:staticcheck // want to make obvious what we're cloning here
		EngineKubernetes: k8s.New(e.BaseEngine.Clone()),     // nolint:staticcheck // want to make obvious what we're cloning here
	}
}

// SetUserAgent sets user agent for clients
func (e *CloudTpl) SetUserAgent(val string) {
	e.UserAgent = val
}

func (e *CloudTpl) SetTemplateRootPath(val string) {
	path, err := filepath.Abs(filepath.Clean(val))
	if err != nil {
		e.Logger.Error(`invalid base path`, slog.String("path", val), slog.Any("error", err))
		os.Exit(1)
	}
	e.TemplateRootPath = path
}

func (e *CloudTpl) SetTemplateRelPath(val string) {
	path, err := filepath.Abs(filepath.Clean(val))
	if err != nil {
		e.Logger.Error(`invalid base path`, slog.String("path", val), slog.Any("error", err))
		os.Exit(1)
	}
	e.TemplateRelPath = path
}

func (e *CloudTpl) SetLintMode(val bool) {
	e.LintMode = val
}

func (e *CloudTpl) TxtTemplate(name string) *textTemplate.Template {
	tmpl := textTemplate.New(name).Funcs(sprig.TxtFuncMap())
	for _, engine := range e.engines {
		tmpl = tmpl.Funcs(engine.TxtFuncMap(tmpl))
	}

	if !e.LintMode {
		tmpl.Option("missingkey=error")
	} else {
		tmpl.Option("missingkey=zero")
	}

	return tmpl
}

func (e *CloudTpl) Parse(path string, templateData interface{}, buf *strings.Builder) error {
	e.CurrentPath = path

	tmpl := e.TxtTemplate(path)

	content, err := os.ReadFile(path)
	if err != nil {
		return e.HandleCicdError(fmt.Errorf(`unable to read file: '%w'`, err))
	}

	parsedContent, err := tmpl.Parse(string(content))
	if err != nil {
		return e.HandleCicdError(fmt.Errorf(`unable to parse file: %w`, err))
	}

	oldPwd, err := os.Getwd()
	if err != nil {
		return e.HandleCicdError(err)
	}

	if err = os.Chdir(e.TemplateRootPath); err != nil {
		return e.HandleCicdError(err)
	}

	if err = parsedContent.Execute(buf, templateData); err != nil {
		return fmt.Errorf(`unable to process template: '%w'`, err)
	}

	if err = os.Chdir(oldPwd); err != nil {
		return e.HandleCicdError(err)
	}

	return nil
}

func (e *CloudTpl) SetAzureTokenCredentialCallback(callback func() *azcore.TokenCredential) {
	e.engines[EngineAzure].(*azure.AzureEngine).SetAzureTokenCredentialCallback(callback)
}
