package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

// msGraphServicePrincipalByDisplayName fetches one servicePrincipal from MsGraph API using displayName
func (e *AzureTemplateExecutor) msGraphServicePrincipalByDisplayName(displayName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph servicePrincipal by displayName '%v'`, displayName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphServicePrincipalByDisplayName`, displayName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client := msgraph.NewGroupsClient(e.msGraphClient.GetTenantID())
		client.BaseClient.Authorizer = e.msGraphClient.Authorizer()

		queryOpts := odata.Query{
			Filter: fmt.Sprintf(
				`displayName eq '%v'`,
				escapeMsGraphFilter(displayName),
			),
		}
		list, _, err := client.List(e.ctx, queryOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
		}

		if list == nil {
			return nil, fmt.Errorf(`servicePrincipal '%v' was not found in AzureAD`, displayName)
		}

		if len(*list) == 1 {
			return (*list)[0], nil
		} else {
			return nil, fmt.Errorf(`found more then one servicePrincipal '%v'`, displayName)
		}
	})
}

// msGraphServicePrincipalList fetches list of servicePrincipals from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphServicePrincipalList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph servicePrincipal list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphServicePrincipalList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client := msgraph.NewServicePrincipalsClient(e.msGraphClient.GetTenantID())
		client.BaseClient.Authorizer = e.msGraphClient.Authorizer()

		queryOpts := odata.Query{
			Filter: filter,
		}
		list, _, err := client.List(e.ctx, queryOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph group: %v`, err.Error())
		}

		return list, nil
	})

}
