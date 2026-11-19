package azure

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/webdevops/go-common/azuresdk/armclient"
)

// azStorageAccountAccessKeys fetches container blob from StorageAccount
func (e *AzureEngine) azStorageAccountAccessKeys(resourceID string) (interface{}, error) {
	e.Logger.Info(`fetching Azure StorageAccount accesskey`, slog.String("resourceID", resourceID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azStorageAccountAccessKeys`, resourceID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armstorage.NewAccountsClient(resourceInfo.Subscription, e.AzureClient().GetCred(), e.AzureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		result, err := client.ListKeys(e.Ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, err
		}

		return e.TransformToInterface(result.Keys)
	})
}

// azStorageAccountContainerBlob fetches container blob from StorageAccount
func (e *AzureEngine) azStorageAccountContainerBlob(containerBlobUrl string) (interface{}, error) {
	e.Logger.Info(`fetching Azure StorageAccount container blob`, slog.String("containerBlobUrl", containerBlobUrl))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	pathUrl, err := azblob.ParseURL(containerBlobUrl)
	if err != nil {
		return nil, err
	}

	cacheKey := e.GenerateCacheKey(`azStorageAccountContainerBlob`, containerBlobUrl)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		azblobOpts := azblob.ClientOptions{ClientOptions: *e.AzureClient().NewAzCoreClientOptions()}

		storageAccountUrl := fmt.Sprintf("%s://%s", pathUrl.Scheme, pathUrl.Host)
		client, err := azblob.NewClient(storageAccountUrl, e.AzureClient().GetCred(), &azblobOpts)
		if err != nil {
			return nil, err
		}

		response, err := client.DownloadStream(e.Ctx, pathUrl.ContainerName, pathUrl.BlobName, nil)
		if err != nil {
			return nil, err
		}

		if content, err := io.ReadAll(response.Body); err == nil {
			return string(content), nil
		} else {
			return nil, err
		}
	})
}
