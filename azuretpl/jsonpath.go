package azuretpl

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
)

// jsonPath executes jsonPath query on an object and returns the result
func (e *AzureTemplateExecutor) jsonPath(jsonPath string, v interface{}) (interface{}, error) {

	if v, enabled := e.lintResult(); enabled {
		// validate jsonpath
		_, err := jsonpath.Language().NewEvaluableWithContext(e.ctx, jsonPath)
		return v, err
	}

	ret, err := jsonpath.Get(jsonPath, v)
	if err != nil {
		return nil, fmt.Errorf(`unable to execute jsonpath '%v': %v`, jsonPath, err.Error())
	}

	return ret, nil
}
