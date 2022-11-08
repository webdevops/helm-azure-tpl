package config

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Verbose bool `short:"v"  long:"verbose"      env:"AZURETPL_VERBOSE"  description:"verbose mode"`
			LogJson bool `           long:"log.json"     env:"AZURETPL_LOG_JSON" description:"Switch log output to json format"`
		}

		// Api option
		Azure struct {
			Tenant      *string `long:"azure.tenant"                   env:"AZURE_TENANT_ID"           description:"Azure tenant id"`
			Environment *string `long:"azure.environment"              env:"AZURE_ENVIRONMENT"         description:"Azure environment name"`
		}

		DryRun bool `long:"dry-run" env:"AZURETPL_DRY_RUN"    description:"dry run, do not write any files"`
		Debug  bool `long:"debug"   env:"HELM_DEBUG"          description:"debug run, print generated content to stdout (WARNING: can expose secrets!)"`
		Stdout bool `long:"stdout"  env:"AZURETPL_STDOUT"     description:"Print parsed content to stdout instead of file (logs will be written to stderr)"`

		Template struct {
			BasePath *string `long:"template.basepath"  env:"AZURETPL_TEMPLATE_BASEPATH"  description:"sets custom base path (if empty, base path is set by base directory for each file. will be appended to all root paths inside templates)"`
		}

		Target struct {
			Prefix  string  `long:"target.prefix"   env:"AZURETPL_TARGET_PREFIX"   description:"adds this value as prefix to filename on save (not used if targetfile is specified in argument)"`
			Suffix  string  `long:"target.suffix"   env:"AZURETPL_TARGET_SUFFIX"   description:"adds this value as suffix to filename on save (not used if targetfile is specified in argument)"`
			FileExt *string `long:"target.fileext"  env:"AZURETPL_TARGET_FILEEXT"  description:"replaces file extension (or adds if empty) with this value (eg. '.yaml')"`
		}

		ValuesFiles  []string `long:"values"  env:"AZURETPL_VALUES" env-delim:":" description:"path to yaml files for .Values"`
		JSONValues   []string `long:"set-json"                           description:"set JSON values on the command line (can specify multiple or separate values with commas: key1=jsonval1,key2=jsonval2)"`
		Values       []string `long:"set"                                description:"set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)"`
		StringValues []string `long:"set-string"                         description:"set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)"`
		FileValues   []string `long:"set-file"                           description:"set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)"`

		Args struct {
			Command string   `positional-arg-name:"command" description:"specifies what to do (help, version, lint, apply)" choice:"help" choice:"version" choice:"lint" choice:"apply" required:"yes"` // nolint:staticcheck
			Files   []string `positional-arg-name:"files" description:"list of files to process (will overwrite files, different target file can be specified as sourcefile:targetfile)"`
		} `positional-args:"yes" `
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
