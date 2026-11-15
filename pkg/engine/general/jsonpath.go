package general

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
)

// jsonPath executes jsonPath query on an object and returns the result
func (e *GeneralEngine) jsonPath(jsonPath string, v interface{}) (interface{}, error) {

	if v, enabled := e.LintResult(); enabled {
		// validate jsonpath
		_, err := jsonpath.Language().NewEvaluableWithContext(e.Ctx, jsonPath)
		return v, err
	}

	ret, err := jsonpath.Get(jsonPath, v)
	if err != nil {
		return nil, fmt.Errorf(`unable to execute jsonpath '%v': %w`, jsonPath, err)
	}

	return ret, nil
}
