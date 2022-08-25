package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

// msGraphServicePrincipalByDisplayName fetches one servicePrincipal from MsGraph API using displayName
func (e *AzureTemplateExecutor) msGraphServicePrincipalByDisplayName(displayName string) interface{} {
	e.logger.Infof(`fetching MsGraph servicePrincipal by displayName "%v"`, displayName)

	cacheKey := generateCacheKey(`msGraphServicePrincipalByDisplayName`, displayName)
	return e.cacheResult(cacheKey, func() interface{} {
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
			e.logger.Fatalf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
		}

		if list == nil {
			e.logger.Fatalf(`servicePrincipal "%v" was not found in AzureAD`, displayName)
		}

		if len(*list) == 1 {
			return (*list)[0]
		} else {
			e.logger.Fatalf(`found more then one servicePrincipal "%v"`, displayName)
		}

		return ""
	})
}

// msGraphServicePrincipalList fetches list of servicePrincipals from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphServicePrincipalList(filter string) interface{} {
	e.logger.Infof(`fetching MsGraph servicePrincipal list with $filter "%v"`, filter)

	cacheKey := generateCacheKey(`msGraphServicePrincipalList`, filter)
	return e.cacheResult(cacheKey, func() interface{} {
		client := msgraph.NewServicePrincipalsClient(e.msGraphClient.GetTenantID())
		client.BaseClient.Authorizer = e.msGraphClient.Authorizer()

		queryOpts := odata.Query{
			Filter: filter,
		}
		list, _, err := client.List(e.ctx, queryOpts)
		if err != nil {
			e.logger.Fatalf(`failed to query MsGraph group: %v`, err.Error())
		}

		return list
	})

}
