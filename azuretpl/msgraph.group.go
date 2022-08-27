package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

// msGraphGroupByDisplayName fetches one group from MsGraph API using displayName
func (e *AzureTemplateExecutor) msGraphGroupByDisplayName(displayName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph group by displayName '%v'`, displayName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`msGraphGroupByDisplayName`, displayName)
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
			return nil, fmt.Errorf(`failed to query MsGraph group: %v`, err.Error())
		}

		if list == nil {
			return nil, fmt.Errorf(`group '%v' was not found in AzureAD`, displayName)
		}

		if len(*list) == 1 {
			return (*list)[0], nil
		} else {
			return nil, fmt.Errorf(`found more then one group '%v'`, displayName)
		}

		return "", nil
	})
}

// msGraphGroupList fetches list of groups from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphGroupList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph group list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`msGraphGroupList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client := msgraph.NewGroupsClient(e.msGraphClient.GetTenantID())
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
