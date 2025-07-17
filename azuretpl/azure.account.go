package azuretpl

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
)

func (e *AzureTemplateExecutor) azAccountInfo() (interface{}, error) {
	if e.azureCliAccountInfo == nil {
		cmd := exec.Command("az", "account", "show", "-o", "json")
		cmd.Stderr = os.Stderr

		accountInfo, err := cmd.Output()
		if err != nil {
			e.logger.Error(`unable to detect Azure TenantID via 'az account show'`, slog.Any("error", err))
			os.Exit(1)
		}

		err = json.Unmarshal(accountInfo, &e.azureCliAccountInfo)
		if err != nil {
			e.logger.Error(`unable to parse 'az account show' output`, slog.Any("error", err))
			os.Exit(1)
		}
	}
	return e.azureCliAccountInfo, nil
}
