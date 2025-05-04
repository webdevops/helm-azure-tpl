package models

import (
	"time"
)

type (
	Opts struct {
		Keyvault struct {
			ExpiryWarning time.Duration `long:"keyvault.expiry.warningduration"   env:"AZURETPL_KEYVAULT_EXPIRY_WARNING_DURATION"   description:"warn before soon expiring Azure KeyVault entries" default:"168h"`
			IgnoreExpiry  bool          `long:"keyvault.expiry.ignore"            env:"AZURETPL_KEYVAULT_EXPIRY_IGNORE"   description:"ignore expiry date of Azure KeyVault entries and don't fail'"`
		}

		ValuesFiles  []string `long:"values"  env:"AZURETPL_VALUES" env-delim:":" description:"path to yaml files for .Values"`
		JSONValues   []string `long:"set-json"                           description:"set JSON values on the command line (can specify multiple or separate values with commas: key1=jsonval1,key2=jsonval2)"`
		Values       []string `long:"set"                                description:"set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)"`
		StringValues []string `long:"set-string"                         description:"set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)"`
		FileValues   []string `long:"set-file"                           description:"set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)"`
	}
)
