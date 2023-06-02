package azuretpl

import (
	"encoding/json"
	"strings"

	"github.com/webdevops/go-common/azuresdk/armclient"
)

func escapeMsGraphFilter(val string) string {
	return strings.ReplaceAll(val, `''`, `\'`)
}

func generateCacheKey(val ...string) string {
	return strings.Join(val, ":")
}

func transformToInterface(obj interface{}) (interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var ret interface{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func parseSubscriptionId(val string) (string, error) {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(strings.ToLower(val), "/subscriptions/") {
		info, err := armclient.ParseResourceId(val)
		if err != nil {
			return "", err
		}

		return info.Subscription, nil
	} else {
		return val, nil
	}
}
