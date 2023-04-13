package azuretpl

import (
	"encoding/json"
	"strings"
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
