package general

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	textTemplate "text/template"

	"github.com/webdevops/helm-azure-tpl/cloudtpl/base"
)

const (
	recursionMaxNums = 100
)

type (
	GeneralEngine struct {
		base.Engine
		*base.BaseEngine
	}
)

func New(base *base.BaseEngine) *GeneralEngine {
	ret := GeneralEngine{
		BaseEngine: base,
	}
	return &ret
}

func (e *GeneralEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	includedNames := make(map[string]int)

	funcMap := textTemplate.FuncMap{
		// json
		"jsonPath": e.jsonPath,

		// files
		"filesGet":  e.filesGet,
		"filesGlob": e.filesGlob,

		"include": func(path string, data interface{}) (string, error) {
			sourcePath := e.fileMakePathAbs(path)

			if v, ok := includedNames[sourcePath]; ok {
				if v > recursionMaxNums {
					return "", fmt.Errorf(`too many recursions for inclusion of '%v'`, path)
				}
				includedNames[sourcePath]++
			} else {
				includedNames[sourcePath] = 1
			}

			content, err := os.ReadFile(sourcePath)
			if err != nil {
				return "", fmt.Errorf(`unable to read file: %w`, err)
			}

			parsedContent, err := tmpl.Parse(string(content))
			if err != nil {
				return "", fmt.Errorf(`unable to parse file: %w`, err)
			}

			var buf bytes.Buffer
			err = parsedContent.Execute(&buf, data)
			if err != nil {
				return "", fmt.Errorf("unable to process template:\n%w", err)
			}

			includedNames[sourcePath]--
			return buf.String(), nil
		},

		"required": func(message string, val interface{}) (interface{}, error) {
			if val == nil {
				if e.LintMode {
					// Don't fail on missing required values when linting
					e.Logger.Info(fmt.Sprintf("[TPL::required] missing required value: %s", message))
					return "", nil
				}
				return val, errors.New(message)
			} else if _, ok := val.(string); ok {
				if val == "" {
					if e.LintMode {
						// Don't fail on missing required values when linting
						e.Logger.Info(fmt.Sprintf("[TPL::required] missing required value: %s", message))
						return "", nil
					}
					return val, errors.New(message)
				}
			}
			return val, nil
		},

		"fail": func(message string) (string, error) {
			if e.LintMode {
				// Don't fail when linting
				e.Logger.Info(fmt.Sprintf("[TPL::fail] fail: %s", message))
				return "", nil
			}
			return "", errors.New(message)
		},
	}

	return funcMap
}
