package azuretpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/webdevops/go-common/utils/to"
)

func (e *AzureTemplateExecutor) addSummaryLine(section, val string) {
	if _, ok := e.summary[section]; !ok {
		e.summary[section] = []string{}
	}

	e.summary[section] = append(e.summary[section], val)
}

func (e *AzureTemplateExecutor) addSummaryKeyvaultSecret(vaultUrl string, secret azsecrets.GetSecretResponse) {
	section := "Azure Keyvault Secrets"
	if _, ok := e.summary[section]; !ok {
		e.summary[section] = []string{
			"| KeyVault | Secret | Version | ContentType | Expiry |",
			"|----------|--------|---------|-------------|--------|",
		}
	}

	expiryDate := "<not set>"
	if secret.Attributes != nil && secret.Attributes.Expires != nil {
		expiryDate = secret.Attributes.Expires.Format(time.RFC3339)
	}

	val := fmt.Sprintf(
		"| %s | %s | %s | %s | %s |",
		vaultUrl,
		secret.ID.Name(),
		secret.ID.Version(),
		to.String(secret.ContentType),
		expiryDate,
	)

	e.summary[section] = append(e.summary[section], val)
}

func (e *AzureTemplateExecutor) buildSummary() string {
	output := []string{
		"# helm-azure-tpl summary",
	}

	for section, rows := range e.summary {
		output = append(output, fmt.Sprintf("\n### %s\n", section))
		output = append(output, strings.Join(rows, "\n"))
	}

	return "\n" + strings.Join(output, "\n") + "\n"
}

func (e *AzureTemplateExecutor) postSummary() {
	if val := os.Getenv("AZURETPL_EXPERIMENTAL_SUMMARY"); val != "true" && val != "1" {
		return
	}

	// github summary
	if summaryPath := os.Getenv("GITHUB_STEP_SUMMARY"); summaryPath != "" && filepath.IsLocal(summaryPath) {
		content := e.buildSummary()

		// If the file doesn't exist, create it, or append to the file
		f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			e.logger.Warnf(`unable to post GITHUB step summary: %w`, err)
			return
		}
		defer f.Close() // nolint: errcheck

		if _, err := f.Write([]byte(content)); err != nil {
			e.logger.Warnf(`unable to post GITHUB step summary: %w`, err)
		}
	}
}
