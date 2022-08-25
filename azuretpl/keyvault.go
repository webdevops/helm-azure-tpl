package azuretpl

import (
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

func (e *AzureTemplateExecutor) azureKeyVaultSecret(vaultUrl string, secretName string) interface{} {
	e.logger.Infof(`fetching Azure KeyVault secret "%v" -> "%v"`, vaultUrl, secretName)
	secretClient := azsecrets.NewClient(vaultUrl, e.client.GetCred(), nil)

	secret, err := secretClient.GetSecret(e.ctx, secretName, "", nil)
	if err != nil {
		e.logger.Fatalf(`unable to fetch secret "%[2]v" from vault "%[1]v": %[3]v`, vaultUrl, secretName, err.Error())
	}

	return secret
}
