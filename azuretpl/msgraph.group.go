package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

func (e *AzureTemplateExecutor) msGraphGroupByDisplayName(groupName string) interface{} {
	e.logger.Infof(`fetching MsGraph group by displayName "%v"`, groupName)

	cacheKey := generateCacheKey(`msGraphGroupByDisplayName`, groupName)
	return e.cacheResult(cacheKey, func() interface{} {
		client := msgraph.NewGroupsClient(e.msGraphClient.GetTenantID())
		client.BaseClient.Authorizer = e.msGraphClient.Authorizer()

		queryOpts := odata.Query{
			Filter: fmt.Sprintf(
				`displayName eq '%v'`,
				escapeMsGraphFilter(groupName),
			),
		}
		list, _, err := client.List(e.ctx, queryOpts)
		if err != nil {
			e.logger.Fatalf(`failed to query MsGraph group: %v`, err.Error())
		}

		if list == nil {
			e.logger.Fatalf(`group "%v" was not found in AzureAD`, groupName)
		}

		if len(*list) == 1 {
			return (*list)[0]
		} else {
			e.logger.Fatalf(`found more then one group "%v"`, groupName)
		}

		return ""
	})
}

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
