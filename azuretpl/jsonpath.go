package azuretpl

import (
	"github.com/PaesslerAG/jsonpath"
)

// jsonPath executes jsonPath query on an object and returns the result
func (e *AzureTemplateExecutor) jsonPath(jsonPath string, v interface{}) interface{} {
	ret, err := jsonpath.Get(jsonPath, v)
	if err != nil {
		e.logger.Fatalf(`unable to execute jsonpath '%v': %v`, jsonPath, err.Error())
	}

	return ret
}
