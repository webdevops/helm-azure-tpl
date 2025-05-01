package azuretpl

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"
)

// azRedisAccessKeys fetches accesskeys from redis cache
func (e *AzureTemplateExecutor) azRedisAccessKeys(resourceID string) (interface{}, error) {
	e.logger.Infof(`fetching Azure Redis accesskey '%v'`, resourceID)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azRedisAccessKeys`, resourceID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armredis.NewClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		result, err := client.ListKeys(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, err
		}

		val := []string{
			to.String(result.PrimaryKey),
			to.String(result.SecondaryKey),
		}

		return transformToInterface(val)
	})
}
