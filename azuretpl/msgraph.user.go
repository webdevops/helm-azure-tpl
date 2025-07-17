package azuretpl

import (
	"fmt"
	"log/slog"

	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/webdevops/go-common/utils/to"
)

// mgUserByUserPrincipalName fetches one user from MsGraph API using userPrincipalName
func (e *AzureTemplateExecutor) mgUserByUserPrincipalName(userPrincipalName string) (interface{}, error) {
	e.logger.Info(`fetching MsGraph user by userPrincipalName`, slog.String("upn", userPrincipalName))

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`mgUserByUserPrincipalName`, userPrincipalName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		requestOpts := &users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Filter: to.StringPtr(fmt.Sprintf(
					`userPrincipalName eq '%v'`,
					escapeMsGraphFilter(userPrincipalName),
				)),
			},
		}
		result, err := e.msGraphClient().ServiceClient().Users().Get(e.ctx, requestOpts)
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
func (e *AzureTemplateExecutor) mgUserList(filter string) (interface{}, error) {
	e.logger.Info(`fetching MsGraph user list with $filter`, slog.String("filter", filter))

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`mgUserList`, filter)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.msGraphClient().ServiceClient().Users().Get(e.ctx, nil)
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

func (e *AzureTemplateExecutor) mgUserCreateListFromResult(result models.UserCollectionResponseable) (list []interface{}, err error) {
	pageIterator, pageIteratorErr := msgraphcore.NewPageIterator[models.Userable](result, e.msGraphClient().RequestAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
	if pageIteratorErr != nil {
		return list, pageIteratorErr
	}

	iterateErr := pageIterator.Iterate(e.ctx, func(user models.Userable) bool {
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
