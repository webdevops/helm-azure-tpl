package azuretpl

import (
	"fmt"
	"log/slog"

	armauthorization "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/webdevops/go-common/utils/to"
)

// azRoleDefinition fetches Azure RoleDefinition by roleName
func (e *AzureTemplateExecutor) azRoleDefinition(scope string, roleName string) (interface{}, error) {
	e.logger.Info(`fetching Azure RoleDefinition`, slog.String("scope", scope), slog.String("role", roleName))

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azRoleDefinition`, scope, roleName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
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

		return transformToInterface(result[0])
	})
}

// azRoleDefinitionList fetches list of roleDefinitions using $filter
func (e *AzureTemplateExecutor) azRoleDefinitionList(scope string, filter ...string) (interface{}, error) {
	var roleDefinitionFilter string

	if len(filter) == 1 {
		roleDefinitionFilter = filter[0]
	}

	e.logger.Info(`fetching Azure RoleDefinitions`, slog.String("scope", scope), slog.String("filter", roleDefinitionFilter))

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azRoleDefinitionList`, scope, roleDefinitionFilter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.fetchAzureRoleDefinitions(scope, roleDefinitionFilter)
		if err != nil {
			return nil, err
		}
		return transformToInterface(result)
	})
}

func (e *AzureTemplateExecutor) fetchAzureRoleDefinitions(scope string, filter string) ([]armauthorization.RoleDefinition, error) {
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
		result, err := pager.NextPage(e.ctx)
		if err != nil {
			return nil, err
		}

		for _, roleDefinition := range result.Value {
			list = append(list, *roleDefinition)
		}
	}

	return list, nil
}
