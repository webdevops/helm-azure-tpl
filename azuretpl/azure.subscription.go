package azuretpl

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/webdevops/go-common/utils/to"
)

// azureSubscription fetches current or defined Azure subscription
func (e *AzureTemplateExecutor) azureSubscription(subscriptionID ...string) interface{} {
	var selectedSubscriptionId string
	if len(subscriptionID) > 1 {
		e.logger.Fatalf(`{{azureSubscription}} only supports zero or one subscriptionIDs`)
	}

	if len(subscriptionID) == 1 {
		//
		selectedSubscriptionId = subscriptionID[0]
	} else {
		// use current subscription id
		if val, exists := e.azureCliAccountInfo["id"].(string); exists {
			selectedSubscriptionId = val
		} else {
			e.logger.Fatalf(`{{azureSubscription}} is unable to find current subscription from "az account show" output`)
		}
	}

	e.logger.Infof(`fetching Azure subscription "%v"`, selectedSubscriptionId)

	cacheKey := generateCacheKey(`azureSubscription`, selectedSubscriptionId)
	return e.cacheResult(cacheKey, func() interface{} {
		client, err := armsubscriptions.NewClient(e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			e.logger.Fatalf(err.Error())
		}

		resource, err := client.Get(e.ctx, selectedSubscriptionId, nil)
		if err != nil {
			e.logger.Fatalf(`unable to fetch Azure subscription "%v": %v`, selectedSubscriptionId, err.Error())
		}

		subscriptionData, err := transformToInterface(resource)
		if err != nil {
			e.logger.Fatalf(`unable to transform Azure subscription "%v": %v`, selectedSubscriptionId, err.Error())
		}
		return subscriptionData
	})
}

// azureSubscriptionList fetches list of visible Azure subscriptions
func (e *AzureTemplateExecutor) azureSubscriptionList() interface{} {
	e.logger.Infof(`fetching Azure subscriptions`)

	cacheKey := generateCacheKey(`azureSubscriptionList`)
	return e.cacheResult(cacheKey, func() interface{} {
		client, err := armsubscriptions.NewClient(e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			e.logger.Panic(err.Error())
		}

		pager := client.NewListPager(nil)
		var ret []interface{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				e.logger.Panic(err)
			}

			for _, subscription := range result.Value {
				subscriptionData, err := transformToInterface(subscription)
				if err != nil {
					e.logger.Fatalf(`unable to transform Azure subscription "%v": %v`, to.String(subscription.SubscriptionID), err.Error())
				}
				ret = append(ret, subscriptionData)
			}
		}

		return ret
	})
}
