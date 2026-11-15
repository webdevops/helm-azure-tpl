package helm

import (
	textTemplate "text/template"

	"github.com/webdevops/helm-azure-tpl/cloudtpl/base"
)

type (
	HelmEngine struct {
		base.Engine
		*base.BaseEngine
	}
)

func New(base *base.BaseEngine) *HelmEngine {
	ret := HelmEngine{
		BaseEngine: base,
	}
	return &ret
}

func (e *HelmEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	funcMap := textTemplate.FuncMap{
		// time
		`fromUnixtime`: fromUnixtime,
		`toRFC3339`:    toRFC3339,

		// borrowed from github.com/helm/helm
		"toToml":        toTOML,
		"mustToToml":    mustToTOML,
		"fromToml":      fromTOML,
		"toYaml":        toYAML,
		"mustToYaml":    mustToYAML,
		"toYamlPretty":  toYAMLPretty,
		"fromYaml":      fromYAML,
		"fromYamlArray": fromYAMLArray,
		"toJson":        toJSON,
		"mustToJson":    mustToJSON,
		"fromJson":      fromJSON,
		"fromJsonArray": fromJSONArray,

		// Duration helpers
		"mustToDuration":       mustToDuration,
		"durationSeconds":      durationSeconds,
		"durationMilliseconds": durationMilliseconds,
		"durationMicroseconds": durationMicroseconds,
		"durationNanoseconds":  durationNanoseconds,
		"durationMinutes":      durationMinutes,
		"durationHours":        durationHours,
		"durationDays":         durationDays,
		"durationWeeks":        durationWeeks,
		"durationRoundTo":      durationRoundTo,
		"durationTruncateTo":   durationTruncateTo,
	}

	return funcMap
}
