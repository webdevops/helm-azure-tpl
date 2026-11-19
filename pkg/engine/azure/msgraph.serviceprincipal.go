package azure

import (
	"fmt"
	"log/slog"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/webdevops/go-common/utils/to"
)

// mgServicePrincipalByDisplayName fetches one servicePrincipal from MsGraph API using displayName
func (e *AzureEngine) mgServicePrincipalByDisplayName(displayName string) (interface{}, error) {
	e.Logger.Info(`fetching MsGraph servicePrincipal by displayName`, slog.String("displayName", displayName))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`mgServicePrincipalByDisplayName`, displayName)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
			QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(`displayName eq '%v'`,
					escapeMsGraphFilter(displayName))),
			},
		}
		result, err := e.MsGraphClient().ServiceClient().ServicePrincipals().Get(e.Ctx, requestOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %w`, err)
		}

		list, err := e.mgServicePrincipalCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %w`, err)
		}

		switch len(list) {
		case 0:
			return nil, nil
		case 1:
			return list[0], nil
		default:
			return nil, fmt.Errorf(`found more then one servicePrincipal '%v'`, displayName)
		}
	})
}

// mgServicePrincipalList fetches list of servicePrincipals from MsGraph API using $filter query
func (e *AzureEngine) mgServicePrincipalList(filter string) (interface{}, error) {
	e.Logger.Info(`fetching MsGraph servicePrincipal list with $filter`, slog.String("filter", filter))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`mgServicePrincipalList`, filter)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.MsGraphClient().ServiceClient().ServicePrincipals().Get(e.Ctx, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %w`, err)
		}

		list, err := e.mgServicePrincipalCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %w`, err)
		}

		return list, nil
	})
}

func (e *AzureEngine) mgServicePrincipalCreateListFromResult(result models.ServicePrincipalCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator[models.ServicePrincipalable](result, e.MsGraphClient().RequestAdapter(), models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.Ctx, func(servicePrincipal models.ServicePrincipalable) bool {
		obj, serializeErr := e.mgSerializeObject(servicePrincipal)
		if serializeErr != nil {
			err = serializeErr
			return false
		}

		list = append(list, obj)
		return true
	})
	if iterateErr != nil {
		return list, iterateErr
	}

	return
}
