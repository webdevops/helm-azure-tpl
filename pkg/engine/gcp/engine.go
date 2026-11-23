package gcp

import (
	textTemplate "text/template"

	"github.com/webdevops/helm-azure-tpl/pkg/engine/base"
)

type (
	GcpEngine struct {
		base.Engine
		*base.BaseEngine
	}
)

func New(base *base.BaseEngine) *GcpEngine {
	ret := GcpEngine{
		BaseEngine: base,
	}
	ret.Cache = ret.Cache.Sub("gcp")
	return &ret
}

func (e *GcpEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	funcMap := textTemplate.FuncMap{
		// Asset
		`gcpAssetResource`: e.gcpAssetResource,
		`gcpAssetQuery`:    e.gcpAssetQuery,

		// SecretsManager
		`gcpSecretsManagerSecret`:         e.gcpSecretsManagerSecret,
		`gcpSecretsManagerSecretVersions`: e.gcpSecretsManagerSecretVersions,
		`gcpSecretsManagerSecretList`:     e.gcpSecretsManagerSecretList,
	}

	return funcMap
}
