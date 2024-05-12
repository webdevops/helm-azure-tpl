package azuretpl

import (
	"fmt"
	"strings"

	"github.com/webdevops/go-common/azuresdk/armclient"
)

// azResourceGraphQuery executes ResourceGraph query and returns result
func (e *AzureTemplateExecutor) azResourceGraphQuery(scope interface{}, query string) (interface{}, error) {
	resourceGraphOptions := armclient.ResourceGraphOptions{}

	scopeList := []string{}

	switch v := scope.(type) {
	case string:
		for _, val := range strings.Split(v, ",") {
			scopeList = append(scopeList, val)

			err := parseResourceGraphScope(val, &resourceGraphOptions)
			if err != nil {
				panic(err)
			}
		}
	case []string:
		scopeList = v

		for _, val := range v {
			err := parseResourceGraphScope(val, &resourceGraphOptions)
			if err != nil {
				panic(err)
			}
		}
	default:
		return nil, fmt.Errorf(`invalid scope type, expected string or string array, got "%v"`, v)
	}

	if len(resourceGraphOptions.ManagementGroups) == 0 && len(resourceGraphOptions.Subscriptions) == 0 {
		return nil, fmt.Errorf(`{{azResourceGraphQuery}} needs at least one subscription ID or managementGroup ID`)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	e.logger.Infof(`executing ResourceGraph query '%v' for scopes '%v'`, query, scopeList)

	cacheKey := generateCacheKey(`azResourceGraphQuery`, query, strings.Join(scopeList, ","))
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.azureClient().ExecuteResourceGraphQuery(e.ctx, query, resourceGraphOptions)
		if err != nil {
			return nil, err
		}

		return transformToInterface(result)
	})
}

func parseResourceGraphScope(scope string, resourceGraphOptions *armclient.ResourceGraphOptions) error {
	scope = strings.TrimSpace(scope)
	if strings.HasPrefix(strings.ToLower(scope), "/providers/microsoft.management/managementgroups/") {
		// seems to be a mgmtgroup id
		managementGroupId := strings.TrimPrefix(strings.ToLower(scope), "/providers/microsoft.management/managementgroups/")
		resourceGraphOptions.ManagementGroups = append(resourceGraphOptions.Subscriptions, managementGroupId)
	} else {
		// might be a subscription id
		val, err := parseSubscriptionId(scope)
		if err != nil {
			return err
		}
		resourceGraphOptions.Subscriptions = append(resourceGraphOptions.Subscriptions, val)
	}

	return nil
}
