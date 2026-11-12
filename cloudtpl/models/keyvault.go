package models

import (
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/webdevops/go-common/utils/to"
)

type (
	AzSecret struct {
		// The secret management attributes.
		Attributes *azsecrets.SecretAttributes `json:"attributes"`

		// The content type of the secret.
		ContentType *string `json:"contentType"`

		// The secret id.
		ID string `json:"id"`

		// Application specific metadata in the form of key-value pairs.
		Tags map[string]*string `json:"tags"`

		// The secret value.
		Value *string `json:"value"`

		Managed bool `json:"managed"`

		Version string `json:"version" yaml:"version"`
		Name    string `json:"name" yaml:"name"`
	}
)

func NewAzSecretItem(secret azsecrets.Secret) *AzSecret {
	return &AzSecret{
		Attributes:  secret.Attributes,
		ContentType: secret.ContentType,
		ID:          string(*secret.ID),
		Tags:        secret.Tags,
		Value:       secret.Value,
		Managed:     to.Bool(secret.Managed),
		Version:     secret.ID.Version(),
		Name:        secret.ID.Name(),
	}
}

func NewAzSecretItemFromSecretproperties(secret azsecrets.SecretProperties) *AzSecret {
	return &AzSecret{
		Attributes:  secret.Attributes,
		ContentType: secret.ContentType,
		ID:          string(*secret.ID),
		Tags:        secret.Tags,
		Value:       nil,
		Managed:     to.Bool(secret.Managed),
		Version:     secret.ID.Version(),
		Name:        secret.ID.Name(),
	}
}
