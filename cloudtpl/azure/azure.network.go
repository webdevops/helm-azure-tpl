package azure

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"
)

// azPublicIpAddress fetches ipAddress from Azure Public IP Address
func (e *AzureEngine) azPublicIpAddress(resourceID string) (interface{}, error) {
	e.Logger.Info(`fetching Azure PublicIpAddress`, slog.String("resourceID", resourceID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azPublicIpAddress`, resourceID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewPublicIPAddressesClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		pipAddress, err := client.Get(e.Ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		return to.String(pipAddress.Properties.IPAddress), nil
	})
}

// azPublicIpPrefixAddressPrefix fetches ipAddress prefix from Azure Public IP Address prefix
func (e *AzureEngine) azPublicIpPrefixAddressPrefix(resourceID string) (interface{}, error) {
	e.Logger.Info(`fetching Azure PublicIpPrefix`, slog.String("resourceID", resourceID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azPublicIpPrefixAddressPrefix`, resourceID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewPublicIPPrefixesClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		pipAddress, err := client.Get(e.Ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		return to.String(pipAddress.Properties.IPPrefix), nil
	})
}

// azVirtualNetworkAddressPrefixes fetches ipAddress prefixes (array) from Azure VirtualNetwork
func (e *AzureEngine) azVirtualNetworkAddressPrefixes(resourceID string) (interface{}, error) {
	e.Logger.Info(`fetching AddressPrefixes from Azure VirtualNetwork '%v'`, slog.String("resourceID", resourceID))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azVirtualNetworkAddressPrefixes`, resourceID)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewVirtualNetworksClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		vnet, err := client.Get(e.Ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		if vnet.Properties.AddressSpace != nil {
			return to.Slice(vnet.Properties.AddressSpace.AddressPrefixes), nil
		}
		return []string{}, nil
	})
}

// azVirtualNetworkSubnetAddressPrefixes fetches ipAddress prefixes (array) from Azure VirtualNetwork subnet
func (e *AzureEngine) azVirtualNetworkSubnetAddressPrefixes(resourceID string, subnetName string) (interface{}, error) {
	e.Logger.Info(`fetching AddressPrefixes from Azure VirtualNetwork`, slog.String("resourceID", resourceID), slog.String("subnet", subnetName))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`azVirtualNetworkSubnetAddressPrefixes`, resourceID, subnetName)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewVirtualNetworksClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
		if err != nil {
			return nil, err
		}

		vnet, err := client.Get(e.Ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		if vnet.Properties.Subnets != nil {
			for _, subnet := range vnet.Properties.Subnets {
				if strings.EqualFold(to.String(subnet.Name), subnetName) {
					if subnet.Properties.AddressPrefixes != nil {
						return to.Slice(subnet.Properties.AddressPrefixes), nil
					} else if subnet.Properties.AddressPrefix != nil {
						return []string{to.String(subnet.Properties.AddressPrefix)}, nil
					}
				}
			}
		}

		return nil, fmt.Errorf(`unable to find Azure VirtualNetwork '%v' subnet '%v'`, resourceID, subnetName)
	})
}
