package azuretpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/webdevops/go-common/azuresdk/cloudconfig"
	"github.com/webdevops/go-common/utils/to"

	"github.com/webdevops/helm-azure-tpl/azuretpl/models"
)

// buildAppConfigUrl builds Azure AppConfig url in case value is supplied as AppConfig name only
func (e *AzureTemplateExecutor) buildAppConfigUrl(appConfigUrl string) (string, error) {
	// do not build keyvault url in lint mode
	if e.LintMode {
		return appConfigUrl, nil
	}

	// vault url generation (if only vault name is specified)
	if !strings.HasPrefix(strings.ToLower(appConfigUrl), "https://") {
		switch cloudName := e.azureClient().GetCloudName(); cloudName {
		case cloudconfig.AzurePublicCloud:
			appConfigUrl = fmt.Sprintf(`https://%s.azconfig.io`, appConfigUrl)
		case cloudconfig.AzureChinaCloud:
			appConfigUrl = fmt.Sprintf(`https://%s.azconfig.azure.cn`, appConfigUrl)
		case cloudconfig.AzureGovernmentCloud:
			appConfigUrl = fmt.Sprintf(`https://%s.azconfig.azure.us`, appConfigUrl)
		default:
			return appConfigUrl, fmt.Errorf(`cannot build Azure AppConfig url for "%s" and Azure cloud "%s", please use full url`, appConfigUrl, cloudName)
		}
	}

	// improve caching by removing trailing slash
	appConfigUrl = strings.TrimSuffix(appConfigUrl, "/")

	return appConfigUrl, nil
}

// azAppConfigSetting fetches secret object from Azure KeyVault
func (e *AzureTemplateExecutor) azAppConfigSetting(appConfigUrl string, settingName string, label string) (interface{}, error) {
	// azure keyvault url detection
	if val, err := e.buildAppConfigUrl(appConfigUrl); err == nil {
		appConfigUrl = val
	} else {
		return nil, err
	}

	e.logger.Infof(`fetching AppConfig value '%v' -> '%v'`, appConfigUrl, settingName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`azAppConfigSetting`, appConfigUrl, settingName, label)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		client, err := azappconfig.NewClient(appConfigUrl, e.azureClient().GetCred(), nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to create appconfig client for instance "%v": %w`, appConfigUrl, err)
		}

		options := azappconfig.GetSettingOptions{}
		if label != "" {
			options.Label = to.StringPtr(label)
		}

		appConfigValue, err := client.GetSetting(e.ctx, settingName, &options)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch app setting value "%[2]v" from appconfig instance "%[1]v": %[3]w`, appConfigUrl, settingName, err)
		}

		switch to.String(appConfigValue.ContentType) {
		case "application/vnd.microsoft.appconfig.keyvaultref+json;charset=utf-8":
			keyVaultResult := struct {
				Uri string
			}{}

			data := to.String(appConfigValue.Value)
			if err := json.Unmarshal([]byte(data), &keyVaultResult); err != nil {
				return nil, fmt.Errorf(`unable to parse keyvault reference from app setting value "%[2]v" from appconfig instance "%[1]v": %[3]w`, appConfigUrl, settingName, err)
			}

			if keyVaultResult.Uri == "" {
				return nil, fmt.Errorf(`unable to parse keyvault reference from app setting value "%[2]v" from appconfig instance "%[1]v": %[3]w`, appConfigUrl, settingName, errors.New("keyvault uri is empty"))
			}

			keyVaultRefUrl, err := url.Parse(keyVaultResult.Uri)
			if err != nil {
				return nil, fmt.Errorf(`unable to parse keyvault reference from app setting value "%[2]v" from appconfig instance "%[1]v": %[3]w`, appConfigUrl, settingName, err)
			}

			vaultUrl := fmt.Sprintf(`https://%s`, keyVaultRefUrl.Host)
			vaultSecretPathParts := strings.Split(strings.TrimPrefix(keyVaultRefUrl.Path, "/"), "/")

			secretName := vaultSecretPathParts[1]
			secretVersion := ""
			if len(vaultSecretPathParts) >= 3 {
				secretVersion = vaultSecretPathParts[2]
			}

			secret, err := e.azKeyVaultSecret(vaultUrl, secretName, secretVersion)
			if err != nil {
				return nil, fmt.Errorf(`unable to fetch keyvault reference from app setting value "%[2]v" from appconfig instance "%[1]v": %[3]w`, appConfigUrl, settingName, err)
			}

			secretMap := secret.(map[string]interface{})
			appConfigValue.Value = to.StringPtr(secretMap["value"].(string))
		}

		return transformToInterface(models.NewAzAppconfigSettingFromReponse(appConfigValue))
	})
}
