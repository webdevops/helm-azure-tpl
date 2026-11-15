package azure

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/webdevops/go-common/utils/to"
)

// azManagementGroup fetches Azure ManagementGroup
func (e *AzureEngine) azManagementGroup(groupID string) (interface{}, error) {
	e.Logger.Info(`fetching Azure ManagementGroup`, slog.String("mgmtgroup", groupID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azManagementGroup`, groupID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := armmanagementgroups.NewClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(`failed to create ManagementGroup "%v": %w`, groupID, err)
		}

		managementGroup, err := client.Get(e.Ctx, groupID, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch ManagementGroup "%v": %w`, groupID, err)
		}

		return e.TransformToInterface(managementGroup)
	})
}

// azManagementGroupSubscriptionList fetches list of Azure Subscriptions under Azure ManagementGroup
func (e *AzureEngine) azManagementGroupSubscriptionList(groupID string) (interface{}, error) {
	e.Logger.Info(`fetching subscriptions from Azure ManagementGroup`, slog.String("mgmtgroup", groupID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azManagementGroupSubscriptionList`, groupID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := armmanagementgroups.NewClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(`failed to create ManagementGroup "%v": %w`, groupID, err)
		}

		pager := client.NewGetDescendantsPager(groupID, nil)
		ret := []interface{}{}
		for pager.More() {
			result, err := pager.NextPage(e.Ctx)
			if err != nil {
				panic(err)
			}

			for _, resource := range result.Value {
				if strings.EqualFold(to.String(resource.Type), "Microsoft.Management/managementGroups/subscriptions") {
					ret = append(ret, resource)
				}
			}
		}
		return e.TransformToInterface(ret)
	})
}
