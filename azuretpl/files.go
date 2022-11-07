package azuretpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (e *AzureTemplateExecutor) filesGet(path string) (string, error) {
	var sourcePath string
	if !filepath.IsAbs(path) {
		sourcePath = filepath.Clean(fmt.Sprintf("%s/%s", e.TemplateBasePath, path))
	} else {
		sourcePath = filepath.Clean(path)
	}

	if val, err := filepath.Abs(sourcePath); err == nil {
		sourcePath = val
	} else {
		return "", fmt.Errorf(`unable to resolve include referance: %w`, err)
	}

	if !strings.HasPrefix(sourcePath, e.TemplateBasePath) {
		return "", fmt.Errorf(
			`'%v' must be in same directory or below (expected prefix: %v, got: %v)`,
			path,
			e.TemplateBasePath,
			filepath.Dir(sourcePath),
		)
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf(`unable to read file: %w`, err)
	}

	return string(content), nil
}

func (e *AzureTemplateExecutor) filesGlob(pattern string) (interface{}, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf(
			`failed to parse glob pattern '%v': %w`,
			pattern,
			err,
		)
	}

	ret := []string{}
	for _, path := range matches {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			ret = append(ret, path)
		}
	}

	return ret, err
}
