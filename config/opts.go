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
		Debug  bool `long:"debug"                       description:"debug run (can expose secrets!)"`

		Args struct {
			Files []string
		} `positional-args:"yes" description:"List of files to process (will overwrite files, different target file can be specified as sourcefile:targetfile)'" required:"yes"`
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
