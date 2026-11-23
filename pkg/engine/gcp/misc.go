package gcp

import (
	"strings"
)

// escapeQueryValue wraps values in quotes for Cloud Asset search queries.
func escapeQueryValue(val string) string {
	val = strings.ReplaceAll(val, `\`, `\\`)
	val = strings.ReplaceAll(val, `"`, `\"`)
	return `"` + val + `"`
}
