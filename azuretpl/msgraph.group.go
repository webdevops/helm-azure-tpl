package azuretpl

import (
	"fmt"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/webdevops/go-common/utils/to"
)

// msGraphGroupByDisplayName fetches one group from MsGraph API using displayName
func (e *AzureTemplateExecutor) msGraphGroupByDisplayName(displayName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph group by displayName '%v'`, displayName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`msGraphGroupByDisplayName`, displayName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &groups.GroupsRequestBuilderGetRequestConfiguration{
			QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(`displayName eq '%v'`,
					escapeMsGraphFilter(displayName))),
			},
		}
		result, err := e.msGraphClient.ServiceClient().Groups().Get(e.ctx, requestOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph group: %w`, err)
		}

		list, err := e.msGraphGroupCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph group: %w`, err)
		}

		switch len(list) {
		case 0:
			return nil, nil
		case 1:
			return list[0], nil
		default:
			return nil, fmt.Errorf(`found more then one group '%v'`, displayName)
		}
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
		result, err := e.msGraphClient.ServiceClient().Groups().Get(e.ctx, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph group: %w`, err)
		}

		list, err := e.msGraphGroupCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph groups: %w`, err)
		}

		return list, nil
	})

}

func (e *AzureTemplateExecutor) msGraphGroupCreateListFromResult(result models.GroupCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator(result, e.msGraphClient.RequestAdapter(), models.CreateGroupCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.ctx, func(pageItem interface{}) bool {
		group := pageItem.(models.Groupable)

		obj, serializeErr := e.msGraphSerializeObject(group)
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
