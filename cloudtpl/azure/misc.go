package azure

import (
	"strings"

	"github.com/webdevops/go-common/azuresdk/armclient"
)

func escapeMsGraphFilter(val string) string {
	return strings.ReplaceAll(val, `''`, `\'`)
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
