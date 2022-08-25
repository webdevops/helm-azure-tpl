package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

// msGraphGroupByDisplayName fetches one group from MsGraph API using displayName
func (e *AzureTemplateExecutor) msGraphGroupByDisplayName(displayName string) interface{} {
	e.logger.Infof(`fetching MsGraph group by displayName "%v"`, displayName)

	cacheKey := generateCacheKey(`msGraphGroupByDisplayName`, displayName)
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
			e.logger.Fatalf(`failed to query MsGraph group: %v`, err.Error())
		}

		if list == nil {
			e.logger.Fatalf(`group "%v" was not found in AzureAD`, displayName)
		}

		if len(*list) == 1 {
			return (*list)[0]
		} else {
			e.logger.Fatalf(`found more then one group "%v"`, displayName)
		}

		return ""
	})
}

// msGraphGroupList fetches list of groups from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphGroupList(filter string) interface{} {
	e.logger.Infof(`fetching MsGraph group list with $filter "%v"`, filter)

	cacheKey := generateCacheKey(`msGraphGroupList`, filter)
	return e.cacheResult(cacheKey, func() interface{} {
		client := msgraph.NewGroupsClient(e.msGraphClient.GetTenantID())
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
