package azuretpl

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"
)

// azurePublicIpAddress fetches ipAddress from Azure Public IP Address
func (e *AzureTemplateExecutor) azurePublicIpAddress(resourceID string) interface{} {
	e.logger.Infof(`fetching Azure PublicIpAddress "%v"`, resourceID)

	cacheKey := generateCacheKey(`azurePublicIpAddress`, resourceID)
	return e.cacheResult(cacheKey, func() interface{} {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			e.logger.Fatalf(`unable to parse Azure resourceID "%v": %v`, resourceID, err.Error())
		}

		client, err := armnetwork.NewPublicIPAddressesClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			e.logger.Fatalf(err.Error())
		}

		pipAddress, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			e.logger.Fatalf(`unable to fetch Azure resource "%v": %v`, resourceID, err.Error())
		}

		return to.String(pipAddress.Properties.IPAddress)
	})
}

// azurePublicIpPrefixAddressPrefix fetches ipAddress prefix from Azure Public IP Address prefix
func (e *AzureTemplateExecutor) azurePublicIpPrefixAddressPrefix(resourceID string) interface{} {
	e.logger.Infof(`fetching Azure PublicIpPrefix "%v"`, resourceID)

	cacheKey := generateCacheKey(`azurePublicIpPrefixAddressPrefix`, resourceID)
	return e.cacheResult(cacheKey, func() interface{} {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			e.logger.Fatalf(`unable to parse Azure resourceID "%v": %v`, resourceID, err.Error())
		}

		client, err := armnetwork.NewPublicIPPrefixesClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			e.logger.Fatalf(err.Error())
		}

		pipAddress, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			e.logger.Fatalf(`unable to fetch Azure resource "%v": %v`, resourceID, err.Error())
		}

		return to.String(pipAddress.Properties.IPPrefix)
	})
}

// azureVirtualNetworkAddressPrefixes fetches ipAddress prefixes (array) from Azure VirtualNetwork
func (e *AzureTemplateExecutor) azureVirtualNetworkAddressPrefixes(resourceID string) interface{} {
	e.logger.Infof(`fetching AddressPrefixes from Azure VirtualNetwork "%v"`, resourceID)

	cacheKey := generateCacheKey(`azureVirtualNetworkAddressPrefixes`, resourceID)
	return e.cacheResult(cacheKey, func() interface{} {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			e.logger.Fatalf(`unable to parse Azure resourceID "%v": %v`, resourceID, err.Error())
		}

		client, err := armnetwork.NewVirtualNetworksClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			e.logger.Fatalf(err.Error())
		}

		vnet, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			e.logger.Fatalf(`unable to fetch Azure resource "%v": %v`, resourceID, err.Error())
		}

		if vnet.Properties.AddressSpace != nil {
			return to.Slice(vnet.Properties.AddressSpace.AddressPrefixes)
		}
		return []string{}
	})
}

// azureVirtualNetworkSubnetAddressPrefixes fetches ipAddress prefixes (array) from Azure VirtualNetwork subnet
func (e *AzureTemplateExecutor) azureVirtualNetworkSubnetAddressPrefixes(resourceID string, subnetName string) interface{} {
	e.logger.Infof(`fetching AddressPrefixes from Azure VirtualNetwork "%v" subnet "%v"`, resourceID, subnetName)

	cacheKey := generateCacheKey(`azureVirtualNetworkSubnetAddressPrefixes`, resourceID, subnetName)
	return e.cacheResult(cacheKey, func() interface{} {
		resourceInfo, err := armclient.ParseResourceId(resourceID)
		if err != nil {
			e.logger.Fatalf(`unable to parse Azure resourceID "%v": %v`, resourceID, err.Error())
		}

		client, err := armnetwork.NewVirtualNetworksClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
		if err != nil {
			e.logger.Fatalf(err.Error())
		}

		vnet, err := client.Get(e.ctx, resourceInfo.ResourceGroup, resourceInfo.ResourceName, nil)
		if err != nil {
			e.logger.Fatalf(`unable to fetch Azure resource "%v": %v`, resourceID, err.Error())
		}

		if vnet.Properties.Subnets != nil {
			for _, subnet := range vnet.Properties.Subnets {
				if strings.EqualFold(to.String(subnet.Name), subnetName) {
					if subnet.Properties.AddressPrefixes != nil {
						return to.Slice(subnet.Properties.AddressPrefixes)
					} else if subnet.Properties.AddressPrefix != nil {
						return []string{to.String(subnet.Properties.AddressPrefix)}
					}
				}
			}
		}

		e.logger.Fatalf(`unable to find Azure VirtualNetwork "%v" subnet "%v"`, resourceID, subnetName)

		return []string{}
	})
}
