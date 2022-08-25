# Helm plugin for Azure template processing

[![license](https://img.shields.io/github/license/webdevops/helm-azure-tpl.svg)](https://github.com/webdevops/helm-azure-tpl/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fhelm--azure--tpl-blue)](https://hub.docker.com/r/webdevops/helm-azure-tpl/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fhelm--azure--tpl-blue)](https://quay.io/repository/webdevops/helm-azure-tpl)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-azure-tpl)](https://artifacthub.io/packages/search?repo=helm-azure-tpl)


## Usage

TODO

## Template functions

### Azure template functions

| Function                                   | Parameters                                     | Description                                                             |
|--------------------------------------------|------------------------------------------------|-------------------------------------------------------------------------|
| `azureKeyVaultSecret`                      | `vaultUrl` (string), `secretName` (string)     | Fetches secret object from Azure KeyVault                               |
| `azureResource`                            | `resourceID` (string), `apiVersion` (string)   | Fetches Azure resource information (interface object)                   |
| `azurePublicIpAddress`                     | `resourceID` (string)                          | Fetches ip address from Azure Public IP                                 |
| `azurePublicIpPrefixAddressPrefix`         | `resourceID` (string)                          | Fetches ip address prefix from Azure Public IP prefix                   |
| `azureVirtualNetworkAddressPrefixes`       | `resourceID` (string)                          | Fetches address prefix (string array) from Azure VirtualNetwork         |
| `azureVirtualNetworkSubnetAddressPrefixes` | `resourceID` (string), `subnetName` (string)   | Fetches address prefix (string array) from Azure VirtualNetwork subnet  |


### Misc template functions

| Function   | Parameters          | Description                                                                          |
|------------|---------------------|--------------------------------------------------------------------------------------|
| `jsonPath` | `jsonPath` (string) | Fetches object information using jsonPath (useful to process `azureResource` output) |

```gotemplate

{{
    azureResource
    "/subscriptions/d86bcf13-ddf7-45ea-82f1-6f656767a318/resourcegroups/k8s/providers/Microsoft.ContainerService/managedClusters/mblaschke"
    "2022-01-01"
    | jsonPath "$.properties.aadProfile"
    | toYaml | nindent 2
}}

```


### Helm template functions (borrowed from helm project)

| Function        | Parameters | Description                 |
|-----------------|------------|-----------------------------|
| `toYaml`        |            | Convert object to yaml      |
| `fromYaml`      |            | Convert yaml to object      |
| `fromYamlArray` |            | Convert yaml array to array |
| `toJson`        |            | Convert object to json      |
| `fromJson`      |            | Convert json to object      |
| `fromJsonArray` |            | Convert json array to array |

## Sprig template functions

[Sprig template functions](http://masterminds.github.io/sprig/defaults.html) are also available


## Examples

```gotemplate

## Fetch resource as object and convert to yaml
{{
azureResource
"/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
"2022-01-01"
| toYaml
}}


## Fetch resource as object, select .properties.aadProfile via jsonPath and convert to yaml
{{
azureResource
"/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
"2022-01-01"
| jsonPath "$.properties.aadProfile"
| toYaml
}}

## Fetch Azure VirtualNetwork address prefixes
{{ azureVirtualNetworkAddressPrefixes
"/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
}}


## Fetch Azure VirtualNetwork subnet address prefixes and join them to a string list
{{ azureVirtualNetworkSubnetAddressPrefixes
"/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
"default2"
| join ","
}}

## Fetch secret value from Azure KeyVault
{{ (azureKeyVaultSecret "https://examplevault.vault.azure.net/" "secretname").Value }}

```
