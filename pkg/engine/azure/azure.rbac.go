package azure

import (
	"fmt"
	"log/slog"

	armauthorization "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/webdevops/go-common/utils/to"
)

// azRoleDefinition fetches Azure RoleDefinition by roleName
func (e *AzureEngine) azRoleDefinition(scope string, roleName string) (interface{}, error) {
	e.Logger.Info(`fetching Azure RoleDefinition`, slog.String("scope", scope), slog.String("role", roleName))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azRoleDefinition`, scope, roleName)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		filter := fmt.Sprintf(
			`roleName eq '%s'`,
			roleName,
		)

		result, err := e.fetchAzureRoleDefinitions(scope, filter)
		if err != nil {
			return nil, err
		}

		if len(result) == 0 {
			return nil, fmt.Errorf(`no Azure RoleDefinition with roleName '%v' for scope '%v' found`, roleName, scope)
		}

		if len(result) > 1 {
			return nil, fmt.Errorf(`multiple Azure RoleDefinitions for roleName '%v' for scope '%v' found`, roleName, scope)
		}

		return e.TransformToInterface(result[0])
	})
}

// azRoleDefinitionList fetches list of roleDefinitions using $filter
func (e *AzureEngine) azRoleDefinitionList(scope string, filter ...string) (interface{}, error) {
	var roleDefinitionFilter string

	if len(filter) == 1 {
		roleDefinitionFilter = filter[0]
	}

	e.Logger.Info(`fetching Azure RoleDefinitions`, slog.String("scope", scope), slog.String("filter", roleDefinitionFilter))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azRoleDefinitionList`, scope, roleDefinitionFilter)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.fetchAzureRoleDefinitions(scope, roleDefinitionFilter)
		if err != nil {
			return nil, err
		}
		return e.TransformToInterface(result)
	})
}

func (e *AzureEngine) fetchAzureRoleDefinitions(scope string, filter string) ([]armauthorization.RoleDefinition, error) {
	client, err := armauthorization.NewRoleDefinitionsClient(e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
	if err != nil {
		return nil, err
	}

	listOpts := armauthorization.RoleDefinitionsClientListOptions{
		Filter: to.StringPtr(filter),
	}
	pager := client.NewListPager(scope, &listOpts)

	list := []armauthorization.RoleDefinition{}
	for pager.More() {
		result, err := pager.NextPage(e.Ctx)
		if err != nil {
			return nil, err
		}

		for _, roleDefinition := range result.Value {
			list = append(list, *roleDefinition)
		}
	}

	return list, nil
}
