package azuretpl

import (
	"context"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"
)

type (
	AzureTemplateExecutor struct {
		ctx    context.Context
		client *armclient.ArmClient
		logger *log.Entry
	}
)

func New(ctx context.Context, client *armclient.ArmClient, logger *log.Entry) *AzureTemplateExecutor {
	return &AzureTemplateExecutor{ctx: ctx, client: client, logger: logger}
}

func (e *AzureTemplateExecutor) TxtFuncMap() template.FuncMap {
	funcMap := map[string]interface{}{
		// azure
		`azureKeyVaultSecret`:                      e.azureKeyVaultSecret,
		`azureResource`:                            e.azureResource,
		`azurePublicIpAddress`:                     e.azurePublicIpAddress,
		`azurePublicIpPrefixAddressPrefix`:         e.azurePublicIpPrefixAddressPrefix,
		`azureVirtualNetworkAddressPrefixes`:       e.azureVirtualNetworkAddressPrefixes,
		`azureVirtualNetworkSubnetAddressPrefixes`: e.azureVirtualNetworkSubnetAddressPrefixes,

		// misc
		`jsonPath`: e.jsonPath,

		// borrowed from helm
		"toYaml":        toYAML,
		"fromYaml":      fromYAML,
		"fromYamlArray": fromYAMLArray,
		"toJson":        toJSON,
		"fromJson":      fromJSON,
		"fromJsonArray": fromJSONArray,
	}

	return funcMap
}
