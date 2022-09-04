package azuretpl

import (
	"fmt"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/webdevops/go-common/utils/to"
)

// msGraphServicePrincipalByDisplayName fetches one servicePrincipal from MsGraph API using displayName
func (e *AzureTemplateExecutor) msGraphServicePrincipalByDisplayName(displayName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph servicePrincipal by displayName '%v'`, displayName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphServicePrincipalByDisplayName`, displayName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
			QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(`displayName eq '%v'`,
					escapeMsGraphFilter(displayName))),
			},
		}
		result, err := e.msGraphClient.ServiceClient().ServicePrincipals().Get(e.ctx, requestOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
		}

		list, err := e.msGraphServicePrincipalCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
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

// msGraphServicePrincipalList fetches list of servicePrincipals from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphServicePrincipalList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph servicePrincipal list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphServicePrincipalList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.msGraphClient.ServiceClient().ServicePrincipals().Get(e.ctx, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
		}

		list, err := e.msGraphServicePrincipalCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph servicePrincipal: %v`, err.Error())
		}

		return list, nil
	})
}

func (e *AzureTemplateExecutor) msGraphServicePrincipalCreateListFromResult(result models.ServicePrincipalCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator(result, e.msGraphClient.RequestAdapter(), models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.ctx, func(pageItem interface{}) bool {
		servicePrincipal := pageItem.(models.ServicePrincipalable)

		obj, serializeErr := e.msGraphSerializeObject(servicePrincipal)
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
