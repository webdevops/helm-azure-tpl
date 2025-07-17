package azuretpl

import (
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/webdevops/go-common/utils/to"
)

// azSubscription fetches current or defined Azure subscription
func (e *AzureTemplateExecutor) azSubscription(subscriptionID ...string) (interface{}, error) {
	var selectedSubscriptionId string
	if len(subscriptionID) > 1 {
		return nil, fmt.Errorf(`{{azSubscription}} only supports zero or one subscriptionIDs`)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	if len(subscriptionID) == 1 {
		//
		selectedSubscriptionId = subscriptionID[0]
	} else {
		// use current subscription id
		if val, exists := e.azureCliAccountInfo["id"].(string); exists {
			selectedSubscriptionId = val
		} else {
			return nil, fmt.Errorf(`{{azSubscription}} is unable to find current subscription from "az account show" output`)
		}
	}

	e.logger.Info(`fetching Azure subscription`, slog.String("subscriptionID", selectedSubscriptionId))

	cacheKey := generateCacheKey(`azSubscription`, selectedSubscriptionId)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client, err := armsubscriptions.NewClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		resource, err := client.Get(e.ctx, selectedSubscriptionId, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure subscription '%v': %w`, selectedSubscriptionId, err)
		}

		subscriptionData, err := transformToInterface(resource)
		if err != nil {
			return nil, fmt.Errorf(`unable to transform Azure subscription '%v': %w`, selectedSubscriptionId, err)
		}
		return subscriptionData, nil
	})
}

// azSubscriptionList fetches list of visible Azure subscriptions
func (e *AzureTemplateExecutor) azSubscriptionList() (interface{}, error) {
	e.logger.Info(`fetching Azure subscriptions`)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azSubscriptionList`)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client, err := armsubscriptions.NewClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			panic(err.Error())
		}

		pager := client.NewListPager(nil)
		var ret []interface{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				panic(err)
			}

			for _, subscription := range result.Value {
				subscriptionData, err := transformToInterface(subscription)
				if err != nil {
					return nil, fmt.Errorf(`unable to transform Azure subscription '%v': %w`, to.String(subscription.SubscriptionID), err)
				}
				ret = append(ret, subscriptionData)
			}
		}

		return ret, nil
	})
}
