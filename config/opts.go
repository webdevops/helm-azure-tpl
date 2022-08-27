package config

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
			LogJson bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`
		}

		// Api option
		Azure struct {
			Tenant      *string `long:"azure.tenant"                   env:"AZURE_TENANT_ID"           description:"Azure tenant id"`
			Environment *string `long:"azure.environment"              env:"AZURE_ENVIRONMENT"         description:"Azure environment name"`
		}

		DryRun bool `long:"dry-run"      env:"DRY_RUN"  description:"dry run"`
		Debug  bool `long:"debug"                       description:"debug run (WARNING: can expose secrets!)"`

		Template struct {
			BasePath *string `long:"template.basepath"  env:"TEMPLATE_BASEPATH"  description:"sets custom base path (if empty, base path is set by base directory for each file)"`
		}

		Target struct {
			Prefix string `long:"target.prefix"  env:"TARGET_PREFIX"  description:"adds this value as prefix to filename on save (not used if targetfile is specified in argument)"`
			Suffix string `long:"target.suffix"  env:"TARGET_SUFFIX"  description:"adds this value as suffix to filename on save (not used if targetfile is specified in argument)"`
		}

		Args struct {
			Command string   `description:"specifies what to do (help, version, lint, apply)" choice:"help" choice:"version" choice:"lint" choice:"apply" required:"yes"` // nolint:staticcheck
			Files   []string `description:"list of files to process (will overwrite files, different target file can be specified as sourcefile:targetfile)"`
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
