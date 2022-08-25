package azuretpl

import (
	"strings"
)

func escapeMsGraphFilter(val string) string {
	return strings.ReplaceAll(val, `''`, `\'`)
}

func generateCacheKey(val ...string) string {
	return strings.Join(val, ":")
}
