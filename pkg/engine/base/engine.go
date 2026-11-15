package base

import (
	"context"
	"encoding/json"
	textTemplate "text/template"

	"github.com/patrickmn/go-cache"
	"github.com/webdevops/go-common/log/slogger"

	"github.com/webdevops/helm-azure-tpl/pkg/config"
)

type (
	BaseEngine struct {
		Ctx       context.Context
		Cache     *cache.Cache
		Logger    *slogger.Logger
		UserAgent string
		Opts      config.Opts
		LintMode  bool

		TemplateRootPath string
		TemplateRelPath  string

		CurrentPath string
	}

	Engine interface {
		TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap
	}
)

// lintResult checks if lint mode is active and returns example value
func (e *BaseEngine) LintResult() (interface{}, bool) {
	if e.LintMode {
		return nil, true
	}
	return nil, false
}

func (e *BaseEngine) TransformToInterface(obj interface{}) (interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var ret interface{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
