package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

// msGraphUserByUserPrincipalName fetches one user from MsGraph API using userPrincipalName
func (e *AzureTemplateExecutor) msGraphUserByUserPrincipalName(userPrincipalName string) interface{} {
	e.logger.Infof(`fetching MsGraph user by userPrincipalName '%v'`, userPrincipalName)

	if val, enabled := e.lintResult(); enabled {
		return val
	}

	cacheKey := generateCacheKey(`msGraphUserByUserPrincipalName`, userPrincipalName)
	return e.cacheResult(cacheKey, func() interface{} {
		client := msgraph.NewUsersClient(e.msGraphClient.GetTenantID())
		client.BaseClient.Authorizer = e.msGraphClient.Authorizer()

		queryOpts := odata.Query{
			Filter: fmt.Sprintf(
				`userPrincipalName eq '%v'`,
				escapeMsGraphFilter(userPrincipalName),
			),
		}
		list, _, err := client.List(e.ctx, queryOpts)
		if err != nil {
			e.logger.Fatalf(`failed to query MsGraph user: %v`, err.Error())
		}

		if list == nil {
			e.logger.Fatalf(`user '%v' was not found in AzureAD`, userPrincipalName)
		}

		if len(*list) == 1 {
			return (*list)[0]
		} else {
			e.logger.Fatalf(`found more then one user '%v'`, userPrincipalName)
		}

		return ""
	})
}

// msGraphUserList fetches list of users from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphUserList(filter string) interface{} {
	e.logger.Infof(`fetching MsGraph user list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val
	}

	cacheKey := generateCacheKey(`msGraphUserList`, filter)
	return e.cacheResult(cacheKey, func() interface{} {
		client := msgraph.NewUsersClient(e.msGraphClient.GetTenantID())
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
