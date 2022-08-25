package azuretpl

import (
	"strings"
)

func escapeMsGraphFilter(val string) string {
	return strings.ReplaceAll(val, `''`, `\'`)
}
