package azuretpl

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"
)

// azResource fetches resource json from Azure REST API using the specified apiVersion
func (e *AzureTemplateExecutor) azResource(resourceID string, apiVersion string) (interface{}, error) {
	e.logger.Infof(`fetching Azure Resource '%v' in apiVersion '%v'`, resourceID, apiVersion)

	cacheKey := generateCacheKey(`azResource`, resourceID, apiVersion)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		return e.fetchAzureResource(resourceID, apiVersion)
	})
}

// azResourceList fetches list of resources by scope (either subscription or resourcegroup) from Azure REST API
func (e *AzureTemplateExecutor) azResourceList(scope string, opts ...string) (interface{}, error) {
	filter := ""
	if len(opts) >= 1 {
		filter = opts[0]
	}

	e.logger.Infof(`fetching Azure Resource list for scope '%v' and filter '%v'`, scope, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azResourceList`, scope, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		scopeInfo, err := armclient.ParseResourceId(scope)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, scope, err)
		}

		client, err := armresources.NewClient(scopeInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		ret := []interface{}{}
		if scopeInfo.ResourceGroup != "" {
			// list by ResourceGroup
			options := armresources.ClientListByResourceGroupOptions{}
			if filter != "" {
				options.Filter = to.StringPtr(filter)
			}

			pager := client.NewListByResourceGroupPager(scopeInfo.ResourceGroup, &options)
			for pager.More() {
				result, err := pager.NextPage(e.ctx)
				if err != nil {
					e.logger.Panic(err)
				}

				for _, resource := range result.Value {
					resourceData, err := transformToInterface(resource)
					if err != nil {
						return nil, fmt.Errorf(`unable to transform Azure resource '%v': %w`, to.String(resource.ID), err)
					}
					ret = append(ret, resourceData)
				}
			}
		} else {
			// list by Subscription
			options := armresources.ClientListOptions{}
			if filter != "" {
				options.Filter = to.StringPtr(filter)
			}

			pager := client.NewListPager(&options)
			for pager.More() {
				result, err := pager.NextPage(e.ctx)
				if err != nil {
					e.logger.Panic(err)
				}

				for _, resource := range result.Value {
					resourceData, err := transformToInterface(resource)
					if err != nil {
						return nil, fmt.Errorf(`unable to transform Azure resource '%v': %w`, to.String(resource.ID), err)
					}
					ret = append(ret, resourceData)
				}
			}
		}

		return ret, nil
	})
}
