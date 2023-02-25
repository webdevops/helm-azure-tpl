package azuretpl

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

// azureKeyVaultSecret fetches secret object from Azure KeyVault
func (e *AzureTemplateExecutor) azureKeyVaultSecret(vaultUrl string, secretName string) (interface{}, error) {
	e.logger.Infof(`fetching Azure KeyVault secret '%v' -> '%v'`, vaultUrl, secretName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`azureKeyVaultSecret`, vaultUrl, secretName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		secretClient, err := azsecrets.NewClient(vaultUrl, e.azureClient().GetCred(), nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to create keyvault client for vault "%v": %w`, vaultUrl, err)
		}

		secret, err := secretClient.GetSecret(e.ctx, secretName, "", nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch secret "%[2]v" from vault "%[1]v": %[3]w`, vaultUrl, secretName, err)
		}

		if !*secret.Attributes.Enabled {
			return nil, fmt.Errorf(`unable to use Azure KeyVault secret '%v' -> '%v': secret is disabled`, vaultUrl, secretName)
		}

		if secret.Attributes.NotBefore != nil && time.Now().Before(*secret.Attributes.NotBefore) {
			return nil, fmt.Errorf(`unable to use Azure KeyVault secret '%v' -> '%v': secret is not yet active (notBefore: %v)`, vaultUrl, secretName, secret.Attributes.NotBefore.Format(time.RFC3339))
		}

		if secret.Attributes.Expires != nil && time.Now().After(*secret.Attributes.Expires) {
			return nil, fmt.Errorf(`unable to useAzure KeyVault secret '%v' -> '%v': secret is expired (expires: %v)`, vaultUrl, secretName, secret.Attributes.Expires.Format(time.RFC3339))
		}

		return transformToInterface(secret)
	})
}

// azureKeyVaultSecretList fetches secrets from Azure KeyVault
func (e *AzureTemplateExecutor) azureKeyVaultSecretList(vaultUrl string, secretNamePattern string) (interface{}, error) {
	e.logger.Infof(`fetching Azure KeyVault secret list from vault '%v'`, vaultUrl)

	secretNamePatternRegExp, err := regexp.Compile(secretNamePattern)
	if err != nil {
		return nil, fmt.Errorf(`unable to compile Regular Expression "%v": %w`, secretNamePattern, err)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`azureKeyVaultSecretList`, vaultUrl)
	list, err := e.cacheResult(cacheKey, func() (interface{}, error) {
		secretClient, err := azsecrets.NewClient(vaultUrl, e.azureClient().GetCred(), nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to create keyvault client for vault "%v": %w`, vaultUrl, err)
		}

		pager := secretClient.NewListSecretsPager(nil)

		ret := map[string]interface{}{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				e.logger.Panic(err)
			}

			for _, secret := range result.Value {
				secretData, err := transformToInterface(secret)
				if err != nil {
					return nil, fmt.Errorf(`unable to transform KeyVault secret '%v': %w`, secret.ID.Name(), err)
				}
				ret[secret.ID.Name()] = secretData
			}
		}

		return transformToInterface(ret)
	})
	if err != nil {
		return list, err
	}

	// filter list
	if secretList, ok := list.(map[string]interface{}); ok {
		ret := map[string]interface{}{}
		for secretName, secret := range secretList {
			if secretNamePatternRegExp.MatchString(secretName) {
				ret[secretName] = secret
			}
		}
		list = ret
	}

	return list, nil
}
