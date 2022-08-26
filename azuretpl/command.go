package azuretpl

import (
	"context"
	"encoding/json"
	"text/template"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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

		azureCliAccountInfo map[string]interface{}
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
	e.cache = globalCache
}

func (e *AzureTemplateExecutor) SetAzureCliAccountInfo(accountInfo map[string]interface{}) {
	e.azureCliAccountInfo = accountInfo
}

func (e *AzureTemplateExecutor) TxtFuncMap() template.FuncMap {
	funcMap := map[string]interface{}{
		// azure
		`azureKeyVaultSecret`:                      e.azureKeyVaultSecret,
		`azureResource`:                            e.azureResource,
		`azureSubscription`:                        e.azureSubscription,
		`azureSubscriptionList`:                    e.azureSubscriptionList,
		`azurePublicIpAddress`:                     e.azurePublicIpAddress,
		`azurePublicIpPrefixAddressPrefix`:         e.azurePublicIpPrefixAddressPrefix,
		`azureVirtualNetworkAddressPrefixes`:       e.azureVirtualNetworkAddressPrefixes,
		`azureVirtualNetworkSubnetAddressPrefixes`: e.azureVirtualNetworkSubnetAddressPrefixes,
		`azureAccountInfo`:                         e.azureAccountInfo,

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

// cacheResult caches template function results (eg. Azure REST API resource information)
func (e *AzureTemplateExecutor) cacheResult(cacheKey string, callback func() interface{}) interface{} {
	if val, ok := e.cache.Get(cacheKey); ok {
		e.logger.Infof("found in cache (%v)", cacheKey)
		return val
	}

	ret := callback()

	e.cache.Set(cacheKey, ret, e.cacheTtl)

	return ret
}

func (e *AzureTemplateExecutor) fetchAzureResource(resourceID string, apiVersion string) interface{} {
	resourceInfo, err := armclient.ParseResourceId(resourceID)
	if err != nil {
		e.logger.Fatalf(`unable to parse Azure resourceID "%v": %v`, resourceID, err.Error())
	}

	client, err := armresources.NewClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
	if err != nil {
		e.logger.Fatalf(err.Error())
	}

	resource, err := client.GetByID(e.ctx, resourceID, apiVersion, nil)
	if err != nil {
		e.logger.Fatalf(`unable to fetch Azure resource "%v": %v`, resourceID, err.Error())
	}

	data, err := resource.MarshalJSON()
	if err != nil {
		e.logger.Fatalf(`unable to marshal Azure resource "%v": %v`, resourceID, err.Error())
	}

	var resourceRawInfo map[string]interface{}
	err = json.Unmarshal(data, &resourceRawInfo)
	if err != nil {
		e.logger.Fatalf(`unable to unmarshal Azure resource "%v": %v`, resourceID, err.Error())
	}

	return resourceRawInfo
}
