package azuretpl

import (
	"fmt"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/webdevops/go-common/utils/to"
)

// mgApplicationByDisplayName fetches one application from MsGraph API using displayName
func (e *AzureTemplateExecutor) mgApplicationByDisplayName(displayName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph application by displayName '%v'`, displayName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`mgApplicationByDisplayName`, displayName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
			QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(`displayName eq '%v'`,
					escapeMsGraphFilter(displayName))),
			},
		}
		result, err := e.msGraphClient().ServiceClient().Applications().Get(e.ctx, requestOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph application: %w`, err)
		}

		list, err := e.mgApplicationCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph application: %w`, err)
		}

		switch len(list) {
		case 0:
			return nil, nil
		case 1:
			return list[0], nil
		default:
			return nil, fmt.Errorf(`found more then one application '%v'`, displayName)
		}
	})
}

// mgApplicationList fetches list of applications from MsGraph API using $filter query
func (e *AzureTemplateExecutor) mgApplicationList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph application list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`mgApplicationList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.msGraphClient().ServiceClient().Applications().Get(e.ctx, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph applications: %w`, err)
		}

		list, err := e.mgApplicationCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph applications: %w`, err)
		}

		return list, nil
	})
}

func (e *AzureTemplateExecutor) mgApplicationCreateListFromResult(result models.ApplicationCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator(result, e.msGraphClient().RequestAdapter(), models.CreateApplicationCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.ctx, func(pageItem interface{}) bool {
		application := pageItem.(models.Applicationable)

		obj, serializeErr := e.mgSerializeObject(application)
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
