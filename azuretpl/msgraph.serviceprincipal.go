package azuretpl

import (
	"fmt"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/webdevops/go-common/utils/to"
)

// mgServicePrincipalByDisplayName fetches one servicePrincipal from MsGraph API using displayName
func (e *AzureTemplateExecutor) mgServicePrincipalByDisplayName(displayName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph servicePrincipal by displayName '%v'`, displayName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`mgServicePrincipalByDisplayName`, displayName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
			QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(`displayName eq '%v'`,
					escapeMsGraphFilter(displayName))),
			},
		}
		result, err := e.msGraphClient().ServiceClient().ServicePrincipals().Get(e.ctx, requestOpts)
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
func (e *AzureTemplateExecutor) mgServicePrincipalList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph servicePrincipal list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`mgServicePrincipalList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.msGraphClient().ServiceClient().ServicePrincipals().Get(e.ctx, nil)
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

func (e *AzureTemplateExecutor) mgServicePrincipalCreateListFromResult(result models.ServicePrincipalCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator[models.ServicePrincipalable](result, e.msGraphClient().RequestAdapter(), models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.ctx, func(servicePrincipal models.ServicePrincipalable) bool {
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
