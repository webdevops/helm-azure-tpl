package azuretpl

import (
	"fmt"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/webdevops/go-common/utils/to"
)

// msGraphUserByUserPrincipalName fetches one user from MsGraph API using userPrincipalName
func (e *AzureTemplateExecutor) msGraphUserByUserPrincipalName(userPrincipalName string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph user by userPrincipalName '%v'`, userPrincipalName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphUserByUserPrincipalName`, userPrincipalName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(
					`userPrincipalName eq '%v'`,
					escapeMsGraphFilter(userPrincipalName),
				)),
			},
		}
		result, err := e.msGraphClient.ServiceClient().Users().Get(e.ctx, requestOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph user: %w`, err)
		}

		list, err := e.msGraphUserCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph user: %w`, err)
		}

		switch len(list) {
		case 0:
			return nil, nil
		case 1:
			return list[0], nil
		default:
			return nil, fmt.Errorf(`found more then one user '%v'`, userPrincipalName)
		}
	})
}

// msGraphUserList fetches list of users from MsGraph API using $filter query
func (e *AzureTemplateExecutor) msGraphUserList(filter string) (interface{}, error) {
	e.logger.Infof(`fetching MsGraph user list with $filter '%v'`, filter)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`msGraphUserList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.msGraphClient.ServiceClient().Users().Get(e.ctx, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph users: %w`, err)
		}

		list, err := e.msGraphUserCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph users: %w`, err)
		}

		return list, nil
	})
}

func (e *AzureTemplateExecutor) msGraphUserCreateListFromResult(result models.UserCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator(result, e.msGraphClient.RequestAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.ctx, func(pageItem interface{}) bool {
		user := pageItem.(models.Userable)

		obj, serializeErr := e.msGraphSerializeObject(user)
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
