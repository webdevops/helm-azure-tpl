package azure

import (
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/webdevops/go-common/utils/to"
)

// azSubscription fetches current or defined Azure subscription
func (e *AzureEngine) azSubscription(subscriptionID ...string) (interface{}, error) {
	var selectedSubscriptionId string
	if len(subscriptionID) > 1 {
		return nil, fmt.Errorf(`{{azSubscription}} only supports zero or one subscriptionIDs`)
	}

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	if len(subscriptionID) == 1 {
		//
		selectedSubscriptionId = subscriptionID[0]
	} else {
		// load az account info
		_, err := e.azAccountInfo()
		if err != nil {
			return nil, err
		}

		// try to read subscription id
		if val, exists := e.azureCliAccountInfo["id"].(string); exists {
			selectedSubscriptionId = val
		} else {
			return nil, fmt.Errorf(`{{azSubscription}} is unable to find current subscription from "az account show" output`)
		}
	}

	e.Logger.Info(`fetching Azure subscription`, slog.String("subscriptionID", selectedSubscriptionId))

	cacheKey := e.GenerateCacheKey(`azSubscription`, selectedSubscriptionId)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := armsubscriptions.NewClient(e.AzureClient().GetCred(), e.AzureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		resource, err := client.Get(e.Ctx, selectedSubscriptionId, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure subscription '%v': %w`, selectedSubscriptionId, err)
		}

		subscriptionData, err := e.TransformToInterface(resource)
		if err != nil {
			return nil, fmt.Errorf(`unable to transform Azure subscription '%v': %w`, selectedSubscriptionId, err)
		}
		return subscriptionData, nil
	})
}

// azSubscriptionList fetches list of visible Azure subscriptions
func (e *AzureEngine) azSubscriptionList() (interface{}, error) {
	e.Logger.Info(`fetching Azure subscriptions`)

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azSubscriptionList`)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := armsubscriptions.NewClient(e.AzureClient().GetCred(), e.AzureClient().NewArmClientOptions())
		if err != nil {
			panic(err.Error())
		}

		pager := client.NewListPager(nil)
		var ret []interface{}
		for pager.More() {
			result, err := pager.NextPage(e.Ctx)
			if err != nil {
				panic(err)
			}

			for _, subscription := range result.Value {
				subscriptionData, err := e.TransformToInterface(subscription)
				if err != nil {
					return nil, fmt.Errorf(`unable to transform Azure subscription '%v': %w`, to.String(subscription.SubscriptionID), err)
				}
				ret = append(ret, subscriptionData)
			}
		}

		return ret, nil
	})
}
