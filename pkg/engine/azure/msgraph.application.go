package azure

import (
	"fmt"
	"log/slog"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/webdevops/go-common/utils/to"
)

// mgApplicationByDisplayName fetches one application from MsGraph API using displayName
func (e *AzureEngine) mgApplicationByDisplayName(displayName string) (interface{}, error) {
	e.Logger.Info(`fetching MsGraph application by displayName`, slog.String("displayName", displayName))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`mgApplicationByDisplayName`, displayName)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
			QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(`displayName eq '%v'`,
					escapeMsGraphFilter(displayName))),
			},
		}
		result, err := e.MsGraphClient().ServiceClient().Applications().Get(e.Ctx, requestOpts)
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
func (e *AzureEngine) mgApplicationList(filter string) (interface{}, error) {
	e.Logger.Info(`fetching MsGraph application list with $filter`, slog.String("filter", filter))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`mgApplicationList`, filter)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.MsGraphClient().ServiceClient().Applications().Get(e.Ctx, nil)
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

func (e *AzureEngine) mgApplicationCreateListFromResult(result models.ApplicationCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator[models.Applicationable](result, e.MsGraphClient().RequestAdapter(), models.CreateApplicationCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.Ctx, func(application models.Applicationable) bool {
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
