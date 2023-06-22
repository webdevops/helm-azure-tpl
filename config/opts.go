package config

import (
	"encoding/json"
	"time"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Development bool `long:"log.devel"    env:"LOG_DEVEL"  description:"development mode"`
			Json        bool `long:"log.json"     env:"LOG_JSON"   description:"Switch log output to json format"`
		}

		// Api option
		Azure struct {
			Tenant      *string `env:"AZURE_TENANT_ID"           description:"Azure tenant id"`
			Environment *string `env:"AZURE_ENVIRONMENT"         description:"Azure environment name"`
		}

		DryRun bool `long:"dry-run" env:"AZURETPL_DRY_RUN"      description:"dry run, do not write any files"`
		Debug  bool `long:"debug"   env:"HELMHELM_DEBUG_DEBUG"  description:"debug run, print generated content to stdout (WARNING: can expose secrets!)"`
		Stdout bool `long:"stdout"  env:"AZURETPL_STDOUT"       description:"Print parsed content to stdout instead of file (logs will be written to stderr)"`

		Template struct {
			BasePath *string `long:"template.basepath"  env:"AZURETPL_TEMPLATE_BASEPATH"  description:"sets custom base path (if empty, base path is set by base directory for each file. will be appended to all root paths inside templates)"`
		}

		Target struct {
			Prefix  string  `long:"target.prefix"   env:"AZURETPL_TARGET_PREFIX"   description:"adds this value as prefix to filename on save (not used if targetfile is specified in argument)"`
			Suffix  string  `long:"target.suffix"   env:"AZURETPL_TARGET_SUFFIX"   description:"adds this value as suffix to filename on save (not used if targetfile is specified in argument)"`
			FileExt *string `long:"target.fileext"  env:"AZURETPL_TARGET_FILEEXT"  description:"replaces file extension (or adds if empty) with this value (eg. '.yaml')"`
		}

		Keyvault struct {
			ExpiryWarning time.Duration `long:"keyvault.expiry.warningduration"   env:"AZURETPL_KEYVAULT_EXPIRY_WARNING_DURATION"   description:"warn before soon expiring Azure KeyVault entries" default:"168h"`
			IgnoreExpiry  bool          `long:"keyvault.expiry.ignore"   env:"AZURETPL_KEYVAULT_EXPIRY_IGNORE"   description:"ignore expiry date of Azure KeyVault entries and don't fail'"`
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
		panic(err)
	}
	return jsonBytes
}
