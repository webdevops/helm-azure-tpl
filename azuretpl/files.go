package azuretpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (e *AzureTemplateExecutor) fileMakePathAbs(path string) string {
	path = filepath.Clean(path)

	if filepath.IsAbs(path) {
		path = fmt.Sprintf(
			"%s/%s",
			e.TemplateRootPath,
			strings.TrimLeft(path, string(os.PathSeparator)),
		)
	} else {
		path = fmt.Sprintf(
			"%s/%s",
			e.TemplateRelPath,
			strings.TrimLeft(path, string(os.PathSeparator)),
		)
	}

	return path
}

func (e *AzureTemplateExecutor) filesGet(path string) (string, error) {
	sourcePath := e.fileMakePathAbs(path)

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf(`unable to read file: %w`, err)
	}

	return string(content), nil
}

func (e *AzureTemplateExecutor) filesGlob(pattern string) (interface{}, error) {
	pattern = e.fileMakePathAbs(pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf(
			`failed to parse glob pattern '%v': %w`,
			pattern,
			err,
		)
	}

	var ret []string
	for _, path := range matches {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			// make path relative
			path = fmt.Sprintf(".%s%s", string(os.PathSeparator), strings.TrimLeft(strings.TrimPrefix(path, e.TemplateRootPath), string(os.PathSeparator)))
			ret = append(ret, path)
		}
	}

	return ret, err
}
