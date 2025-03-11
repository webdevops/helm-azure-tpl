package azuretpl

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/webdevops/go-common/azuresdk/armclient"
)

// azEventHubListByNamespace fetches list of Azure EventHubs by Namespace
func (e *AzureTemplateExecutor) azEventHubListByNamespace(resourceID string) (interface{}, error) {
	e.logger.Infof(`fetching EventHub list by namespace '%v'`, resourceID)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azEventHubListByNamespace`, resourceID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse resourceID "%v": %w`, resourceID, err)
		}

		client, err := armeventhub.NewEventHubsClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(`failed to create EventHubsClient "%v": %w`, resourceID, err)
		}

		pager := client.NewListByNamespacePager(resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)

		ret := []*armeventhub.Eventhub{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				return nil, err
			}

			ret = append(ret, result.Value...)
		}

		return transformToInterface(ret)
	})
}
