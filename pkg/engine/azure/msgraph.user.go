package azure

import (
	"fmt"
	"log/slog"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/webdevops/go-common/utils/to"
)

// mgUserByUserPrincipalName fetches one user from MsGraph API using userPrincipalName
func (e *AzureEngine) mgUserByUserPrincipalName(userPrincipalName string) (interface{}, error) {
	e.Logger.Info(`fetching MsGraph user by userPrincipalName`, slog.String("upn", userPrincipalName))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`mgUserByUserPrincipalName`, userPrincipalName)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(
					`userPrincipalName eq '%v'`,
					escapeMsGraphFilter(userPrincipalName),
				)),
			},
		}
		result, err := e.MsGraphClient().ServiceClient().Users().Get(e.Ctx, requestOpts)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph user: %w`, err)
		}

		list, err := e.mgUserCreateListFromResult(result)
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

// mgUserList fetches list of users from MsGraph API using $filter query
func (e *AzureEngine) mgUserList(filter string) (interface{}, error) {
	e.Logger.Info(`fetching MsGraph user list with $filter`, slog.String("filter", filter))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`mgUserList`, filter)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.MsGraphClient().ServiceClient().Users().Get(e.Ctx, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph users: %w`, err)
		}

		list, err := e.mgUserCreateListFromResult(result)
		if err != nil {
			return nil, fmt.Errorf(`failed to query MsGraph users: %w`, err)
		}

		return list, nil
	})
}

func (e *AzureEngine) mgUserCreateListFromResult(result models.UserCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator[models.Userable](result, e.MsGraphClient().RequestAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.Ctx, func(user models.Userable) bool {
		obj, serializeErr := e.mgSerializeObject(user)
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
