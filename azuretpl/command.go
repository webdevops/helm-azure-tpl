package azuretpl

import (
	"context"
	"text/template"
	"time"

	cache "github.com/patrickmn/go-cache"
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

		cache    *cache.Cache
		cacheTtl time.Duration
	}
)

func New(ctx context.Context, azureClient *armclient.ArmClient, msGraphClient *hamiltonclient.MsGraphClient, logger *log.Entry) *AzureTemplateExecutor {
	e := &AzureTemplateExecutor{
		ctx:           ctx,
		azureClient:   azureClient,
		msGraphClient: msGraphClient,
		logger:        logger,

		cacheTtl: 15 * time.Minute,
	}
	e.init()
	return e
}

func (e *AzureTemplateExecutor) init() {
	e.cache = cache.New(e.cacheTtl, 1*time.Minute)
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

func (e *AzureTemplateExecutor) cacheResult(cacheKey string, callback func() interface{}) interface{} {
	if val, ok := e.cache.Get(cacheKey); ok {
		e.logger.Infof("found in cache (%v)", cacheKey)
		return val
	}

	ret := callback()

	e.cache.Set(cacheKey, ret, e.cacheTtl)

	return ret
}
