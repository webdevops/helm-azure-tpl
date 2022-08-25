package azuretpl

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

func (e *AzureTemplateExecutor) azureKeyVaultSecret(vaultUrl string, secretName string) interface{} {
	e.logger.Infof(`fetching Azure KeyVault secret "%v" -> "%v"`, vaultUrl, secretName)
	secretClient := azsecrets.NewClient(vaultUrl, e.azureClient.GetCred(), nil)

	secret, err := secretClient.GetSecret(e.ctx, secretName, "", nil)
	if err != nil {
		e.logger.Fatalf(`unable to fetch secret "%[2]v" from vault "%[1]v": %[3]v`, vaultUrl, secretName, err.Error())
	}

	if !*secret.Attributes.Enabled {
		e.logger.Fatalf(`Azure KeyVault secret "%v" -> "%v" is not enabled`, vaultUrl, secretName)
	}

	if secret.Attributes.NotBefore != nil && time.Now().Before(*secret.Attributes.NotBefore) {
		e.logger.Fatalf(`Azure KeyVault secret "%v" -> "%v" is not yet active (notBefore: %v)`, vaultUrl, secretName, secret.Attributes.NotBefore.Format(time.RFC3339))
	}

	if secret.Attributes.Expires != nil && time.Now().After(*secret.Attributes.Expires) {
		e.logger.Fatalf(`Azure KeyVault secret "%v" -> "%v" is expired (expires: %v)`, vaultUrl, secretName, secret.Attributes.Expires.Format(time.RFC3339))
	}

	return secret
}
