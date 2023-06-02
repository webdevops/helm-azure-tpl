package azuretpl

import (
	"fmt"
	"strings"
)

// azResourceGraphQuery executes ResourceGraph query and returns result
func (e *AzureTemplateExecutor) azResourceGraphQuery(subscriptionID interface{}, query string) (interface{}, error) {
	subscriptionIdList := []string{}

	switch v := subscriptionID.(type) {
	case string:
		for _, val := range strings.Split(v, ",") {
			val, err := parseSubscriptionId(val)
			if err != nil {
				return nil, err
			}
			subscriptionIdList = append(subscriptionIdList, val)
		}
	case []string:
		for _, val := range v {
			val, err := parseSubscriptionId(val)
			if err != nil {
				return nil, err
			}
			subscriptionIdList = append(subscriptionIdList, val)
		}
	default:
		return nil, fmt.Errorf(`invalid subscription ID type, expected string or string array, got "%v"`, v)
	}

	fmt.Println(subscriptionIdList)

	if len(subscriptionIdList) > 1 {
		return nil, fmt.Errorf(`{{azResourceGraphQuery}} needs at least one subscription ID`)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	e.logger.Infof(`executing ResourceGraph query '%v' for subscriptions '%v'`, query, subscriptionID)

	cacheKey := generateCacheKey(`azResourceGraphQuery`, query, strings.Join(subscriptionIdList, ","))
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		result, err := e.azureClient().ExecuteResourceGraphQuery(e.ctx, subscriptionIdList, query)
		if err != nil {
			return nil, err
		}

		return transformToInterface(result)
	})
}
