package azuretpl

import (
	"encoding/json"
	"os"
	"os/exec"
)

func (e *AzureTemplateExecutor) azAccountInfo() (interface{}, error) {
	if e.azureCliAccountInfo == nil {
		cmd := exec.Command("az", "account", "show", "-o", "json")
		cmd.Stderr = os.Stderr

		accountInfo, err := cmd.Output()
		if err != nil {
			e.logger.Fatalf(`unable to detect Azure TenantID via 'az account show': %v`, err)
		}

		err = json.Unmarshal(accountInfo, &e.azureCliAccountInfo)
		if err != nil {
			e.logger.Fatalf(`unable to parse 'az account show' output: %v`, err)
		}
	}
	return e.azureCliAccountInfo, nil
}
