package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/helm-azure-tpl/azuretpl"
)

type (
	TemplateFile struct {
		Context         context.Context
		SourceFile      string
		TargetFile      string
		TemplateBaseDir string
		Logger          *log.Entry
	}
)

func (f *TemplateFile) Lint() {
	var buf strings.Builder
	f.Logger.Infof(`linting file`)
	f.parse(&buf)
	f.Logger.Info(`file successfully linted`)
}

func (f *TemplateFile) Apply() {
	var buf strings.Builder
	f.Logger.Infof(`process file`)
	f.parse(&buf)

	if opts.Debug {
		fmt.Println()
		fmt.Println(strings.Repeat("-", TermColumns))
		fmt.Printf("--- %v\n", f.TargetFile)
		fmt.Println(strings.Repeat("-", TermColumns))
		fmt.Println(buf.String())
	}

	if !opts.DryRun {
		f.write(&buf)
	} else {
		f.Logger.Warn(`not writing file, DRY RUN active`)
	}
}

func (f *TemplateFile) parse(buf *strings.Builder) {
	ctx := f.Context
	contextLogger := f.Logger

	azureTemplate := azuretpl.New(ctx, AzureClient, MsGraphClient, contextLogger)
	azureTemplate.SetAzureCliAccountInfo(azAccountInfo)
	azureTemplate.SetLintMode(lintMode)
	azureTemplate.SetTemplateBasePath(f.TemplateBaseDir)

	tmpl := template.New(f.SourceFile).Funcs(sprig.TxtFuncMap())
	tmpl = tmpl.Funcs(azureTemplate.TxtFuncMap(tmpl))

	content, err := os.ReadFile(f.SourceFile) // #nosec G304 passed as parameter
	if err != nil {
		contextLogger.Fatalf(`unable to read file: '%v'`, err.Error())
	}

	parsedContent, err := tmpl.Parse(string(content))
	if err != nil {
		contextLogger.Fatalf(`unable to parse file: %v`, err.Error())
	}

	err = parsedContent.Execute(buf, nil)
	if err != nil {
		contextLogger.Fatalf(`unable to process template: '%v'`, err.Error())
	}
}

func (f *TemplateFile) write(buf *strings.Builder) {
	f.Logger.Infof(`writing file '%v'`, f.TargetFile)
	err := os.WriteFile(f.TargetFile, []byte(buf.String()), 0600)
	if err != nil {
		f.Logger.Fatalf(`unable to write target file '%v': %v`, f.TargetFile, err.Error())
	}
}
