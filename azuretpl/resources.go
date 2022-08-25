package azuretpl

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/webdevops/go-common/azuresdk/armclient"
)

// azureResource fetches resource json from Azure REST API using the specified apiVersion
func (e *AzureTemplateExecutor) azureResource(resourceID string, apiVersion string) interface{} {
	e.logger.Infof(`fetching Azure ResourceInfo "%v" in apiVersion "%v"`, resourceID, apiVersion)

	cacheKey := generateCacheKey(`azureResource`, resourceID, apiVersion)
	return e.cacheResult(cacheKey, func() interface{} {

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
	})
}
