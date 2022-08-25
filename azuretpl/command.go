package azuretpl

import (
	"context"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/msgraphsdk/hamiltonclient"
)

type (
	AzureTemplateExecutor struct {
		ctx           context.Context
		azureClient   *armclient.ArmClient
		msGraphClient *hamiltonclient.MsGraphClient
		logger        *log.Entry
	}
)

func New(ctx context.Context, azureClient *armclient.ArmClient, msGraphClient *hamiltonclient.MsGraphClient, logger *log.Entry) *AzureTemplateExecutor {
	return &AzureTemplateExecutor{
		ctx:           ctx,
		azureClient:   azureClient,
		msGraphClient: msGraphClient,
		logger:        logger,
	}
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

		// msGraph
		`msGraphUserByUserPrincipalName`:       e.msGraphUserByUserPrincipalName,
		`msGraphUserList`:                      e.msGraphUserList,
		`msGraphGroupByDisplayName`:            e.msGraphGroupByDisplayName,
		`msGraphGroupList`:                     e.msGraphGroupList,
		`msGraphServicePrincipalByDisplayName`: e.msGraphServicePrincipalByDisplayName,
		`msGraphServicePrincipalList`:          e.msGraphServicePrincipalList,

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
