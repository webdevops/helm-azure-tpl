package azuretpl

import (
	"fmt"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

func (e *AzureTemplateExecutor) msGraphServicePrincipalByDisplayName(servicePrincipalName string) interface{} {
	client := msgraph.NewGroupsClient(e.msGraphClient.GetTenantID())
	client.BaseClient.Authorizer = e.msGraphClient.Authorizer()

	queryOpts := odata.Query{
		Filter: fmt.Sprintf(
			`displayName eq '%v'`,
			escapeMsGraphFilter(servicePrincipalName),
		),
	}
	list, _, err := client.List(e.ctx, queryOpts)
	if err != nil {
		e.logger.Fatalf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
	}

	if list == nil {
		e.logger.Fatalf(`servicePrincipal "%v" was not found in AzureAD`, servicePrincipalName)
	}

	if len(*list) == 1 {
		return (*list)[0]
	} else {
		e.logger.Fatalf(`found more then one servicePrincipal "%v"`, servicePrincipalName)
	}

	return ""
}

func (e *AzureTemplateExecutor) msGraphServicePrincipalList(filter string) interface{} {
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
}
