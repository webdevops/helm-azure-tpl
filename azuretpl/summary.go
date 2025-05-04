package azuretpl

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/webdevops/go-common/utils/to"
	"go.uber.org/zap"

	"github.com/webdevops/helm-azure-tpl/config"
)

const (
	SummaryHeader      = "# helm-azure-tpl summary\n"
	SummaryValueNotSet = "<n/a>"
)

var (
	summary map[string][]string = map[string][]string{}
)

func (e *AzureTemplateExecutor) addSummaryLine(section, val string) {
	if _, ok := summary[section]; !ok {
		summary[section] = []string{}
	}

	summary[section] = append(summary[section], val)
}

func (e *AzureTemplateExecutor) addSummaryKeyvaultSecret(vaultUrl string, secret azsecrets.GetSecretResponse) {
	section := "Azure Keyvault Secrets"
	if _, ok := summary[section]; !ok {
		summary[section] = []string{
			"| KeyVault | Secret | Version | ContentType | Expiry |",
			"|----------|--------|---------|-------------|--------|",
		}
	}

	expiryDate := SummaryValueNotSet
	if secret.Attributes != nil && secret.Attributes.Expires != nil {
		expiryDate = secret.Attributes.Expires.Format(time.RFC3339)
	}

	contentType := to.String(secret.ContentType)
	if contentType == "" {
		contentType = SummaryValueNotSet
	}

	val := fmt.Sprintf(
		"| %s | %s | %s | %s | %s |",
		vaultUrl,
		secret.ID.Name(),
		secret.ID.Version(),
		contentType,
		expiryDate,
	)

	summary[section] = append(summary[section], val)
}

func buildSummary(opts config.Opts) string {
	output := []string{SummaryHeader}

	output = append(output, "templates:\n")
	for _, file := range opts.Args.Files {
		output = append(output, fmt.Sprintf("- %s", file))
	}

	for section, rows := range summary {
		output = append(output, fmt.Sprintf("\n### %s\n", section))
		output = append(output, strings.Join(rows, "\n"))
	}

	return "\n" + strings.Join(output, "\n") + "\n"
}

func PostSummary(logger *zap.SugaredLogger, opts config.Opts) {
	if val := os.Getenv("AZURETPL_EXPERIMENTAL_SUMMARY"); val != "true" && val != "1" {
		return
	}

	// skip empty summary
	if len(summary) == 0 {
		return
	}

	// github summary
	if summaryPath := os.Getenv("GITHUB_STEP_SUMMARY"); summaryPath != "" {
		content := buildSummary(opts)

		// If the file doesn't exist, create it, or append to the file
		f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Warnf(`unable to post GITHUB step summary: %w`, err)
			return
		}
		defer f.Close() // nolint: errcheck

		if _, err := f.Write([]byte(content)); err != nil {
			logger.Warnf(`unable to post GITHUB step summary: %w`, err)
		}
	}
}
