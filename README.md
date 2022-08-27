# Helm plugin for Azure template processing

(also works as standalone go template processor with Azure functions)

[![license](https://img.shields.io/github/license/webdevops/helm-azure-tpl.svg)](https://github.com/webdevops/helm-azure-tpl/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fhelm--azure--tpl-blue)](https://hub.docker.com/r/webdevops/helm-azure-tpl/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fhelm--azure--tpl-blue)](https://quay.io/repository/webdevops/helm-azure-tpl)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-azure-tpl)](https://artifacthub.io/packages/search?repo=helm-azure-tpl)

## Installation

requires `sed` and `curl`

```
helm plugin install https://github.com/webdevops/helm-azure-tpl
```

## Usage

`helm azure-tpl` uses AzureCLI authentication to talk to Azure

Process one file and overwrite it:
```
helm azure-tpl template.tpl
```

Process one file and saves generated content as another file:
```
helm azure-tpl template.tpl:template.yaml
```

General usage:
```
Usage:
  helm azure-tpl [OPTIONS]

Application Options:
      --debug              debug mode [$DEBUG]
  -v, --verbose            verbose mode [$VERBOSE]
      --log.json           Switch log output to json format [$LOG_JSON]
      --azure.tenant=      Azure tenant id [$AZURE_TENANT_ID]
      --azure.environment= Azure environment name (default: AzurePublicCloud) [$AZURE_ENVIRONMENT]
      --dry-run            dry run [$DRY_RUN]

Help Options:
  -h, --help               Show this help message
```



## Template functions

### Azure template functions

| Function                                   | Parameters                                   | Description                                                                    |
|--------------------------------------------|----------------------------------------------|--------------------------------------------------------------------------------|
| `azureAccountInfo`                         |                                              | Output of `az account show`                                                    |
| `azureSubscription`                        | `subscriptionID` (string, optional)          | Fetches Azure subscription (current selected one if `subscriptionID` is empty) |
| `azureSubscriptionList`                    |                                              | Fetches list of all visible azure subscriptions                                |
| `azureKeyVaultSecret`                      | `vaultUrl` (string), `secretName` (string)   | Fetches secret object from Azure KeyVault                                      |
| `azureResource`                            | `resourceID` (string), `apiVersion` (string) | Fetches Azure resource information (interface object)                          |
| `azurePublicIpAddress`                     | `resourceID` (string)                        | Fetches ip address from Azure Public IP                                        |
| `azurePublicIpPrefixAddressPrefix`         | `resourceID` (string)                        | Fetches ip address prefix from Azure Public IP prefix                          |
| `azureVirtualNetworkAddressPrefixes`       | `resourceID` (string)                        | Fetches address prefix (string array) from Azure VirtualNetwork                |
| `azureVirtualNetworkSubnetAddressPrefixes` | `resourceID` (string), `subnetName` (string) | Fetches address prefix (string array) from Azure VirtualNetwork subnet         |

### MsGraph (AzureAD) functions

| Function                               | Parameters             | Description                                                                                                                                                          |
|----------------------------------------|------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `msGraphUserByUserPrincipalName`       | `userPrincipalName`    | Fetches one user by UserPrincipalName                                                                                                                                |
| `msGraphUserList`                      | `filter` (string)      | Fetches list of users based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query             |
| `msGraphGroupByDisplayName`            | `displayName` (string) | Fetches one group by displayName                                                                                                                                     |
| `msGraphGroupList`                     | `filter` (string)      | Fetches list of groups based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query            |
| `msGraphServicePrincipalByDisplayName` | `displayName` (string) | Fetches one serviceprincipal by displayName                                                                                                                          |
| `msGraphServicePrincipalList`          | `filter` (string)      | Fetches list of servicePrincipals based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query |

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
    | toYaml
}}

```


### Helm template functions (borrowed from [helm project](https://github.com/helm/helm))

| Function        | Parameters                          | Description                                  |
|-----------------|-------------------------------------|----------------------------------------------|
| `include`       | `path` (string), `data` (interface) | Parses and includes template file            |
| `required`      | `message` (string)                  | Throws error if passed object/value is empty |
| `fail`          | `message` (string)                  | Throws error                                 |
| `toYaml`        |                                     | Convert object to yaml                       |
| `fromYaml`      |                                     | Convert yaml to object                       |
| `fromYamlArray` |                                     | Convert yaml array to array                  |
| `toJson`        |                                     | Convert object to json                       |
| `fromJson`      |                                     | Convert json to object                       |
| `fromJsonArray` |                                     | Convert json array to array                  |

## Sprig template functions

[Sprig template functions](https://masterminds.github.io/sprig/) are also available


## Examples

```gotemplate

## Fetch resource as object and convert to yaml
{{ azureResource
   "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
   "2022-01-01"
   | toYaml
}}


## Fetch resource as object, select .properties.aadProfile via jsonPath and convert to yaml
{{ azureResource
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

## Fetch current environmentName
{{ azureAccountInfo.environmentName }}

## Fetch current tenantId
{{ azureAccountInfo.tenantId }}

## Fetch current selected subscription displayName
{{ azureSubscription.displayName }}

```
