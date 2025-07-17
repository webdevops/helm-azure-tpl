package azuretpl

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/webdevops/go-common/utils/to"
)

// azManagementGroup fetches Azure ManagementGroup
func (e *AzureTemplateExecutor) azManagementGroup(groupID string) (interface{}, error) {
	e.logger.Info(`fetching Azure ManagementGroup`, slog.String("mgmtgroup", groupID))

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azManagementGroup`, groupID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client, err := armmanagementgroups.NewClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(`failed to create ManagementGroup "%v": %w`, groupID, err)
		}

		managementGroup, err := client.Get(e.ctx, groupID, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch ManagementGroup "%v": %w`, groupID, err)
		}

		return transformToInterface(managementGroup)
	})
}

// azManagementGroupSubscriptionList fetches list of Azure Subscriptions under Azure ManagementGroup
func (e *AzureTemplateExecutor) azManagementGroupSubscriptionList(groupID string) (interface{}, error) {
	e.logger.Info(`fetching subscriptions from Azure ManagementGroup`, slog.String("mgmtgroup", groupID))

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azManagementGroupSubscriptionList`, groupID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client, err := armmanagementgroups.NewClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(`failed to create ManagementGroup "%v": %w`, groupID, err)
		}

		pager := client.NewGetDescendantsPager(groupID, nil)
		ret := []interface{}{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				panic(err)
			}

			for _, resource := range result.Value {
				if strings.EqualFold(to.String(resource.Type), "Microsoft.Management/managementGroups/subscriptions") {
					ret = append(ret, resource)
				}
			}
		}
		return transformToInterface(ret)
	})
}
