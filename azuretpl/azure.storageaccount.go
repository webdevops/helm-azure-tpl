package azuretpl

import (
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// azStorageAccountContainerBlob fetches container blob from StorageAccount
func (e *AzureTemplateExecutor) azStorageAccountContainerBlob(containerBlobUrl string) (interface{}, error) {
	e.logger.Infof(`fetching Azure StorageAccount container blob '%v'`, containerBlobUrl)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	pathUrl, err := azblob.ParseURL(containerBlobUrl)
	if err != nil {
		return nil, err
	}

	cacheKey := generateCacheKey(`azStorageAccountContainerBlob`, containerBlobUrl)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		azblobOpts := azblob.ClientOptions{ClientOptions: *e.azureClient().NewAzCoreClientOptions()}

		storageAccountUrl := fmt.Sprintf("%s://%s", pathUrl.Scheme, pathUrl.Host)
		client, err := azblob.NewClient(storageAccountUrl, e.azureClient().GetCred(), &azblobOpts)
		if err != nil {
			return nil, err
		}

		response, err := client.DownloadStream(e.ctx, pathUrl.ContainerName, pathUrl.BlobName, nil)
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
