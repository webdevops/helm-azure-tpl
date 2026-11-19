package hashicorp

import (
	"fmt"
	textTemplate "text/template"

	vault "github.com/hashicorp/vault/api"

	"github.com/webdevops/helm-azure-tpl/pkg/engine/base"
)

type (
	HashicorpEngine struct {
		base.Engine
		*base.BaseEngine

		vaultClient *vault.Client
	}
)

func New(base *base.BaseEngine) *HashicorpEngine {
	ret := HashicorpEngine{
		BaseEngine: base,
	}
	ret.Cache = ret.Cache.Sub("hashicorp")
	return &ret
}

func (e *HashicorpEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	funcMap := textTemplate.FuncMap{
		// Vault
		`HashicorpVaultSecret`: e.HashicorpVaultSecret,
		`HashicorpVaultGet`:    e.HashicorpVaultGet,
	}

	return funcMap
}

// VaultClient creates on-demand hashicorp vault client and returns it
func (e *HashicorpEngine) VaultClient() *vault.Client {
	if e.vaultClient == nil {
		config := vault.DefaultConfig()

		client, err := vault.NewClient(config)
		if err != nil {
			panic(fmt.Errorf("unable to initialize Vault client: %w", err))
		}
		e.vaultClient = client
	}

	return e.vaultClient
}
