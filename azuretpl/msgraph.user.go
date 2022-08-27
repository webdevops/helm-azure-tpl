package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

// msGraphUserByUserPrincipalName fetches one user from MsGraph API using userPrincipalName
func (e *AzureTemplateExecutor) msGraphUserByUserPrincipalName(userPrincipalName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph user by userPrincipalName '%v'`, userPrincipalName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphUserByUserPrincipalName`, userPrincipalName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
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
			return nil, fmt.Errorf(`failed to query MsGraph user: %v`, err.Error())
		}

		if list == nil {
			return nil, fmt.Errorf(`user '%v' was not found in AzureAD`, userPrincipalName)
		}

		if len(*list) == 1 {
			return (*list)[0], nil
		} else {
			return nil, fmt.Errorf(`found more then one user '%v'`, userPrincipalName)
		}
	})
}

// msGraphUserList fetches list of users from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphUserList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph user list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphUserList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client := msgraph.NewUsersClient(e.msGraphClient.GetTenantID())
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
