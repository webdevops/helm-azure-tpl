package azure

import (
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/webdevops/go-common/azuresdk/armclient"
)

// azManagedClusterUserCredentials fetches user credentials object from managed cluster (AKS)
func (e *AzureEngine) azManagedClusterUserCredentials(resourceID string) (interface{}, error) {
	e.Logger.Info(`fetching Azure ManagedCluster user credentials`, slog.String("resourceID", resourceID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}
	cacheKey := e.GenerateCacheKey(`azManagedClusterUserCredentials`, resourceID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armcontainerservice.NewManagedClustersClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(`failed to create ManagedCluster client for cluster "%v": %w`, resourceID, err)
		}

		userCreds, err := client.ListClusterUserCredentials(e.Ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch ManagedCluster user credentials for cluster "%v": %w`, resourceID, err)
		}

		return e.TransformToInterface(userCreds)
	})
}
