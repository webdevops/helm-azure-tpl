package azuretpl

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/webdevops/go-common/azuresdk/cloudconfig"
	"github.com/webdevops/go-common/utils/to"

	"github.com/webdevops/helm-azure-tpl/azuretpl/models"
)

// buildAzKeyVaulUrl builds Azure KeyVault url in case value is supplied as KeyVault name only
func (e *AzureTemplateExecutor) buildAzKeyVaulUrl(vaultUrl string) (string, error) {
	// do not build keyvault url in lint mode
	if e.LintMode {
		return vaultUrl, nil
	}

	// vault url generation (if only vault name is specified)
	if !strings.HasPrefix(strings.ToLower(vaultUrl), "https://") {
		switch cloudName := e.azureClient().GetCloudName(); cloudName {
		case cloudconfig.AzurePublicCloud:
			vaultUrl = fmt.Sprintf(`https://%s.vault.azure.net`, vaultUrl)
		case cloudconfig.AzureChinaCloud:
			vaultUrl = fmt.Sprintf(`https://%s.vault.azure.cn`, vaultUrl)
		case cloudconfig.AzureGovernmentCloud:
			vaultUrl = fmt.Sprintf(`https://%s.vault.usgovcloudapi.net`, vaultUrl)
		default:
			return vaultUrl, fmt.Errorf(`cannot build Azure KeyVault url for "%s" and Azure cloud "%s", please use full url`, vaultUrl, cloudName)
		}
	}

	// improve caching by removing trailing slash
	vaultUrl = strings.TrimSuffix(vaultUrl, "/")

	return vaultUrl, nil
}

// azKeyVaultSecret fetches secret object from Azure KeyVault
func (e *AzureTemplateExecutor) azKeyVaultSecret(vaultUrl string, secretName string, opts ...string) (interface{}, error) {
	// azure keyvault url detection
	if val, err := e.buildAzKeyVaulUrl(vaultUrl); err == nil {
		vaultUrl = val
	} else {
		return nil, err
	}

	e.logger.Infof(`fetching Azure KeyVault secret '%v' -> '%v'`, vaultUrl, secretName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`azKeyVaultSecret`, vaultUrl, secretName, strings.Join(opts, ";"))
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		secretClient, err := azsecrets.NewClient(vaultUrl, e.azureClient().GetCred(), nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to create keyvault client for vault "%v": %w`, vaultUrl, err)
		}

		version := ""
		if len(opts) == 1 {
			version = opts[0]
		}

		secret, err := secretClient.GetSecret(e.ctx, secretName, version, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch secret "%[2]v" from vault "%[1]v": %[3]w`, vaultUrl, secretName, err)
		}

		if !*secret.Attributes.Enabled {
			return nil, fmt.Errorf(`unable to use Azure KeyVault secret '%v' -> '%v': secret is disabled`, vaultUrl, secretName)
		}

		if secret.Attributes.NotBefore != nil && time.Now().Before(*secret.Attributes.NotBefore) {
			return nil, fmt.Errorf(`unable to use Azure KeyVault secret '%v' -> '%v': secret is not yet active (notBefore: %v)`, vaultUrl, secretName, secret.Attributes.NotBefore.Format(time.RFC3339))
		}

		if secret.Attributes.Expires != nil {
			// secret has expiry date, let's check it

			if time.Now().After(*secret.Attributes.Expires) {
				// secret is expired
				if !e.opts.Keyvault.IgnoreExpiry {
					return nil, fmt.Errorf(`unable to use Azure KeyVault secret '%v' -> '%v': secret is expired (expires: %v, set env AZURETPL_KEYVAULT_EXPIRY_IGNORE=1 to ignore)`, vaultUrl, secretName, secret.Attributes.Expires.Format(time.RFC3339))
				} else {
					e.logger.Warnln(
						e.handleCicdWarning(
							fmt.Errorf(`Azure KeyVault secret '%v' -> '%v': secret is expired, but env AZURETPL_KEYVAULT_EXPIRY_IGNORE=1 is active (expires: %v)`, vaultUrl, secretName, secret.Attributes.Expires.Format(time.RFC3339)),
						),
					)
				}
			} else if time.Now().Add(e.opts.Keyvault.ExpiryWarning).After(*secret.Attributes.Expires) {
				// secret is expiring soon
				e.logger.Warnln(
					e.handleCicdWarning(
						fmt.Errorf(`Azure KeyVault secret '%v' -> '%v': secret is expiring soon (expires: %v)`, vaultUrl, secretName, secret.Attributes.Expires.Format(time.RFC3339)),
					),
				)
			}
		}

		e.logger.Infof(`using Azure KeyVault secret '%v' -> '%v' (version: %v)`, vaultUrl, secretName, secret.ID.Version())
		e.handleCicdMaskSecret(to.String(secret.Secret.Value))

		return transformToInterface(models.NewAzSecretItem(secret.Secret))
	})
}

// azKeyVaultSecretVersions fetches older versions of one secret from Azure KeyVault
func (e *AzureTemplateExecutor) azKeyVaultSecretVersions(vaultUrl string, secretName string, count int) (interface{}, error) {
	// azure keyvault url detection
	if val, err := e.buildAzKeyVaulUrl(vaultUrl); err == nil {
		vaultUrl = val
	} else {
		return nil, err
	}

	e.logger.Infof(`fetching Azure KeyVault secret history '%v' -> '%v' with %d versions`, vaultUrl, secretName, count)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`azKeyVaultSecretHistory`, vaultUrl, secretName, strconv.Itoa(count))
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		secretClient, err := azsecrets.NewClient(vaultUrl, e.azureClient().GetCred(), nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to create keyvault client for vault "%v": %w`, vaultUrl, err)
		}

		pager := secretClient.NewListSecretPropertiesVersionsPager(secretName, nil)

		// get secrets first
		secretList := []*azsecrets.SecretProperties{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				e.logger.Panic(err)
			}

			for _, secretVersion := range result.Value {
				if !*secretVersion.Attributes.Enabled {
					continue
				}

				secretList = append(secretList, secretVersion)
			}

			// WARNING: secrets are ordered by version instead of creation date
			// so we cannot limit paging to just a few pages as even the current secrets
			// could be on the next or last page.
			// this was an awful design decision from Azure not to order the entries by creation date.
			// so we have to get the full list of versions for filtering,
			// otherwise we might miss important entries.
			// luckily versions are limited to just 500 entries but it's still an awful trap.
			// the same applies to the Azure KeyVault secret listing in the Azure Portal,
			// it's a mess, horrible to debug and a trap for all developers.
			//
			// if count >= 0 && len(secretList) >= count {
			// 	break
			// }
		}

		// sort results
		sort.Slice(secretList, func(i, j int) bool {
			return secretList[i].Attributes.Created.UTC().After(secretList[j].Attributes.Created.UTC())
		})

		// process list
		ret := []interface{}{}
		for _, secretVersion := range secretList {
			secret, err := secretClient.GetSecret(e.ctx, secretVersion.ID.Name(), secretVersion.ID.Version(), nil)
			if err != nil {
				return nil, fmt.Errorf(`unable to fetch secret "%[2]v" with version "%[3]v" from vault "%[1]v": %[4]w`, vaultUrl, secretVersion.ID.Name(), secretVersion.ID.Version(), err)
			}

			e.handleCicdMaskSecret(to.String(secret.Secret.Value))

			if val, err := transformToInterface(models.NewAzSecretItem(secret.Secret)); err == nil {
				ret = append(ret, val)
			} else {
				return nil, err
			}

			if count >= 0 && len(ret) >= count {
				break
			}
		}

		return ret, nil
	})
}

// azKeyVaultSecretList fetches secrets from Azure KeyVault
func (e *AzureTemplateExecutor) azKeyVaultSecretList(vaultUrl string, secretNamePattern string) (interface{}, error) {
	// azure keyvault url detection
	if val, err := e.buildAzKeyVaulUrl(vaultUrl); err == nil {
		vaultUrl = val
	} else {
		return nil, err
	}

	e.logger.Infof(`fetching Azure KeyVault secret list from vault '%v'`, vaultUrl)

	secretNamePatternRegExp, err := regexp.Compile(secretNamePattern)
	if err != nil {
		return nil, fmt.Errorf(`unable to compile Regular Expression "%v": %w`, secretNamePattern, err)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}
	cacheKey := generateCacheKey(`azKeyVaultSecretList`, vaultUrl)
	list, err := e.cacheResult(cacheKey, func() (interface{}, error) {
		secretClient, err := azsecrets.NewClient(vaultUrl, e.azureClient().GetCred(), nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to create keyvault client for vault "%v": %w`, vaultUrl, err)
		}

		pager := secretClient.NewListSecretPropertiesPager(nil)

		ret := map[string]interface{}{}
		for pager.More() {
			result, err := pager.NextPage(e.ctx)
			if err != nil {
				e.logger.Panic(err)
			}

			for _, secret := range result.Value {
				secretData, err := transformToInterface(models.NewAzSecretItemFromSecretproperties(*secret))
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
