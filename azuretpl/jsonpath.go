package azuretpl

import (
	"github.com/PaesslerAG/jsonpath"
)

func (e *AzureTemplateExecutor) jsonPath(jsonPath string, v interface{}) interface{} {
	ret, err := jsonpath.Get(jsonPath, v)
	if err != nil {
		e.logger.Fatalf(`unable to execute jsonpath "%v": %v`, jsonPath, err.Error())
	}

	return ret
}
