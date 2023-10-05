package models

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/webdevops/go-common/utils/to"
)

type (
	AzAppconfigSetting struct {
		// The primary identifier of the configuration setting.
		// A Key is used together with a Label to uniquely identify a configuration setting.
		Key *string `json:"key"`

		// The configuration setting's value.
		Value *string `json:"value"`

		// A value used to group configuration settings.
		// A Label is used together with a Key to uniquely identify a configuration setting.
		Label *string `json:"label"`

		// The content type of the configuration setting's value.
		// Providing a proper content-type can enable transformations of values when they are retrieved by applications.
		ContentType *string `json:"contentType"`

		// An ETag indicating the state of a configuration setting within a configuration store.
		ETag *azcore.ETag `json:"eTag"`

		// A dictionary of tags used to assign additional properties to a configuration setting.
		// These can be used to indicate how a configuration setting may be applied.
		Tags map[string]string `json:"tags"`

		// The last time a modifying operation was performed on the given configuration setting.
		LastModified *time.Time `json:"lastModified"`

		// A value indicating whether the configuration setting is read only.
		// A read only configuration setting may not be modified until it is made writable.
		IsReadOnly bool `json:"isReadOnly"`

		// Sync token for the Azure App Configuration client, corresponding to the current state of the client.
		SyncToken *string `json:"syncToken"`
	}
)

func NewAzAppconfigSetting(setting azappconfig.Setting) *AzAppconfigSetting {
	return &AzAppconfigSetting{
		Key:          setting.Key,
		Value:        setting.Value,
		Label:        setting.Label,
		ContentType:  setting.ContentType,
		ETag:         setting.ETag,
		Tags:         setting.Tags,
		LastModified: setting.LastModified,
		IsReadOnly:   to.Bool(setting.IsReadOnly),
	}
}

func NewAzAppconfigSettingFromReponse(setting azappconfig.GetSettingResponse) *AzAppconfigSetting {
	ret := NewAzAppconfigSetting(setting.Setting)
	ret.SyncToken = to.StringPtr(string(setting.SyncToken))
	return ret
}
