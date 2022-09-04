package azuretpl

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"
)

// azurePublicIpAddress fetches ipAddress from Azure Public IP Address
func (e *AzureTemplateExecutor) azurePublicIpAddress(resourceID string) (interface{}, error) {
	e.logger.Infof(`fetching Azure PublicIpAddress '%v'`, resourceID)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azurePublicIpAddress`, resourceID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewPublicIPAddressesClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		pipAddress, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		return to.String(pipAddress.Properties.IPAddress), nil
	})
}

// azurePublicIpPrefixAddressPrefix fetches ipAddress prefix from Azure Public IP Address prefix
func (e *AzureTemplateExecutor) azurePublicIpPrefixAddressPrefix(resourceID string) (interface{}, error) {
	e.logger.Infof(`fetching Azure PublicIpPrefix '%v'`, resourceID)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azurePublicIpPrefixAddressPrefix`, resourceID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewPublicIPPrefixesClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		pipAddress, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		return to.String(pipAddress.Properties.IPPrefix), nil
	})
}

// azureVirtualNetworkAddressPrefixes fetches ipAddress prefixes (array) from Azure VirtualNetwork
func (e *AzureTemplateExecutor) azureVirtualNetworkAddressPrefixes(resourceID string) (interface{}, error) {
	e.logger.Infof(`fetching AddressPrefixes from Azure VirtualNetwork '%v'`, resourceID)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azureVirtualNetworkAddressPrefixes`, resourceID)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewVirtualNetworksClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		vnet, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
		}

		if vnet.Properties.AddressSpace != nil {
			return to.Slice(vnet.Properties.AddressSpace.AddressPrefixes), nil
		}
		return []string{}, nil
	})
}

// azureVirtualNetworkSubnetAddressPrefixes fetches ipAddress prefixes (array) from Azure VirtualNetwork subnet
func (e *AzureTemplateExecutor) azureVirtualNetworkSubnetAddressPrefixes(resourceID string, subnetName string) (interface{}, error) {
	e.logger.Infof(`fetching AddressPrefixes from Azure VirtualNetwork '%v' subnet '%v'`, resourceID, subnetName)

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	cacheKey := generateCacheKey(`azureVirtualNetworkSubnetAddressPrefixes`, resourceID, subnetName)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
		}

		client, err := armnetwork.NewVirtualNetworksClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		vnet, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
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
