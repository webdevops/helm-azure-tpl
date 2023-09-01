# Helm plugin for Azure template processing

Plugin for [Helm](https://github.com/helm/helm) to inject Azure information (subscriptions, resources, msgraph) and Azure KeyVault secrets.
Also works as standalone executable outside of Helm.

[![license](https://img.shields.io/github/license/webdevops/helm-azure-tpl.svg)](https://github.com/webdevops/helm-azure-tpl/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fhelm--azure--tpl-blue)](https://hub.docker.com/r/webdevops/helm-azure-tpl/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fhelm--azure--tpl-blue)](https://quay.io/repository/webdevops/helm-azure-tpl)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-azure-tpl)](https://artifacthub.io/packages/search?repo=helm-azure-tpl)

## Installation

requires `sed` and `curl` for installation

```bash
# Installation of latest version
helm plugin install https://github.com/webdevops/helm-azure-tpl

# Installation of specific version
helm plugin install https://github.com/webdevops/helm-azure-tpl --version=0.43.0

# Update to latest version
helm plugin update azure-tpl

# Deinstallation
helm plugin uninstall azure-tpl
```

## Usage

### Helm (downloader mode)

you can use helm in "downloader" mode to process files eg:

```gotemplate
helm upgrade foobar123 -f azuretpl://config/values.yaml .
```

for additional values files for azure-tpl you can use environment variabels:

```gotemplate
AZURETPL_VALUES=./path/to/azuretpl.yaml helm upgrade foobar123 -f azuretpl://config/values.yaml .
```

### File processing

`helm azure-tpl` uses AzureCLI authentication to talk to Azure

Process one file and overwrite it:
```
helm azure-tpl apply template.tpl
```

Process one file and saves generated content as another file:
```
helm azure-tpl apply template.tpl:template.yaml
```

Processes all `.tpl` files and saves them as `.yaml` files
```
helm azure-tpl apply --target.fileext=.yaml *.tpl
```

General usage:
```
Usage:
  helm-azure-tpl [OPTIONS] [command] [files...]

Application Options:
      --log.devel                        development mode [$LOG_DEVEL]
      --log.json                         Switch log output to json format [$LOG_JSON]
      --dry-run                          dry run, do not write any files [$AZURETPL_DRY_RUN]
      --debug                            debug run, print generated content to stdout (WARNING: can expose secrets!) [$HELMHELM_DEBUG_DEBUG]
      --stdout                           Print parsed content to stdout instead of file (logs will be written to stderr) [$AZURETPL_STDOUT]
      --template.basepath=               sets custom base path (if empty, base path is set by base directory for each file. will be
                                         appended to all root paths inside templates) [$AZURETPL_TEMPLATE_BASEPATH]
      --target.prefix=                   adds this value as prefix to filename on save (not used if targetfile is specified in argument)
                                         [$AZURETPL_TARGET_PREFIX]
      --target.suffix=                   adds this value as suffix to filename on save (not used if targetfile is specified in argument)
                                         [$AZURETPL_TARGET_SUFFIX]
      --target.fileext=                  replaces file extension (or adds if empty) with this value (eg. '.yaml') [$AZURETPL_TARGET_FILEEXT]
      --keyvault.expiry.warningduration= warn before soon expiring Azure KeyVault entries (default: 168h)
                                         [$AZURETPL_KEYVAULT_EXPIRY_WARNING_DURATION]
      --keyvault.expiry.ignore           ignore expiry date of Azure KeyVault entries and don't fail' [$AZURETPL_KEYVAULT_EXPIRY_IGNORE]
      --values=                          path to yaml files for .Values [$AZURETPL_VALUES]
      --set-json=                        set JSON values on the command line (can specify multiple or separate values with commas:
                                         key1=jsonval1,key2=jsonval2)
      --set=                             set values on the command line (can specify multiple or separate values with commas:
                                         key1=val1,key2=val2)
      --set-string=                      set STRING values on the command line (can specify multiple or separate values with commas:
                                         key1=val1,key2=val2)
      --set-file=                        set values from respective files specified via the command line (can specify multiple or separate
                                         values with commas: key1=path1,key2=path2)

Help Options:
  -h, --help                             Show this help message

Arguments:
  command:                               specifies what to do (help, version, lint, apply)
  files:                                 list of files to process (will overwrite files, different target file can be specified as
                                         sourcefile:targetfile)
```

## Build-in objects

| Object    | Description                                                                                                   |
|-----------|---------------------------------------------------------------------------------------------------------------|
| `.Values` | Additional data can be passed via `--values=values.yaml` files which is available under `.Values` (like Helm) |

## Template functions

### Azure template functions

:information_source: Functions can also be used starting with `azure` prefix instead of `az`

| Function                                | Parameters                                    | Description                                                                                                                                                                                                                             |
|-----------------------------------------|-----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `azAccountInfo`                         |                                               | Output of `az account show`                                                                                                                                                                                                             |
| `azSubscription`                        | `subscriptionID` (string, optional)           | Fetches Azure subscription (current selected one if `subscriptionID` is empty)                                                                                                                                                          |
| `azSubscriptionList`                    |                                               | Fetches list of all visible azure subscriptions                                                                                                                                                                                         |
| `azResource`                            | `resourceID` (string), `apiVersion` (string)  | Fetches Azure resource information (json representation, interface object)                                                                                                                                                              |
| `azResourceList`                        | `scope` (string), `filter` (string, optional) | Fetches list of Azure resources and filters it by using [$filter](https://learn.microsoft.com/en-us/rest/api/resources/resources/list), scope can be subscription ID or resourceGroup ID (array, json representation, interface object) |
| `azPublicIpAddress`                     | `resourceID` (string)                         | Fetches ip address from Azure Public IP                                                                                                                                                                                                 |
| `azPublicIpPrefixAddressPrefix`         | `resourceID` (string)                         | Fetches ip address prefix from Azure Public IP prefix                                                                                                                                                                                   |
| `azVirtualNetworkAddressPrefixes`       | `resourceID` (string)                         | Fetches address prefix (string array) from Azure VirtualNetwork                                                                                                                                                                         |
| `azVirtualNetworkSubnetAddressPrefixes` | `resourceID` (string), `subnetName` (string)  | Fetches address prefix (string array) from Azure VirtualNetwork subnet                                                                                                                                                                  |

### Azure Keyvault functions
| Function                   | Parameters                                                               | Description                                                                                                                           |
|----------------------------|--------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| `azKeyVaultSecret`         | `vaultUrl` (string), `secretName` (string), `version` (string, optional) | Fetches secret object from Azure KeyVault                                                                                             |
| `azKeyVaultSecretVersions` | `vaultUrl` (string), `secretName` (string), `count` (integer)            | Fetches the list of `count` secret versions (as array, excluding disabled secrets) from Azure KeyVault                                |
| `azKeyVaultSecretList`     | `vaultUrl` (string), `secretNamePattern` (string, regexp)                | Fetche the list of secret objects (without secret value) from Azure KeyVault and filters list by regular expression secretNamePattern |

response format:
```json
{
  "attributes": {
    "created": 1620236104,
    "enabled": true,
    "exp": 1724593377,
    "nbf": 1661362977,
    "recoverableDays": 0,
    "recoveryLevel": "Purgeable",
    "updated": 1661449616
  },
  "contentType": "...",
  "id": "https://xxx.vault.azure.net/secrets/xxx/xxxxxxxxxx",
  "managed": false,
  "name": "xxx",
  "tags": {},
  "value": "...",
  "version": "xxxxxxxxxx"
}
```

### Azure Redis cache functions
| Function                        | Parameters                  | Description                                         |
|---------------------------------|-----------------------------|-----------------------------------------------------|
| `azRedisAccessKeys`             | `resourceID` (string)       | Fetches access keys from Azure Redis Cache as array |

### Azure StorageAccount functions
| Function                         | Parameters                  | Description                                                |
|----------------------------------|-----------------------------|------------------------------------------------------------|
| `azStorageAccountAccessKeys`     | `resourceID` (string)       | Fetches access keys from Azure StorageAccount as array     |
| `azStorageAccountContainerBlob`  | `containerBlobUrl` (string) | Fetches container blob from Azure StorageAccount as string |

### Azure AppConfig functions
| Function               | Parameters                                                         | Description                                                                          |
|------------------------|--------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| `azAppConfigSetting`   | `appConfigUrl` (string), `settingName` (string), `label` (string)  | Fetches setting value from app configuration instance (resolves keyvault references) |

response format:
```json
{
  "ContentType": "...",
  "ETag": "...",
  "IsReadOnly": false,
  "Key":" ...",
  "Label": null,
  "LastModified": null,
  "SyncToken": "...",
  "Tags": {},
  "Value": "..."
}

```

### Azure RBAC functions
| Function               | Parameters                                   | Description                                                                                              |
|------------------------|----------------------------------------------|----------------------------------------------------------------------------------------------------------|
| `azRoleDefinition`     | `scope` (string), `roleName` (string)        | Fetches Azure RoleDefinition using scope (eg `/subscriptions/xxx`) and roleName                          |
| `azRoleDefinitionList` | `scope` (string), `filter` (string,optional) | Fetches list of Azure RoleDefinitions using scope (eg `/subscriptions/xxx`) and optional `$filter` query |

### Azure ResourceGraph functions
| Function               | Parameters                                              | Description                                                                                                   |
|------------------------|---------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| `azResourceGraphQuery` | `subscriptionID` (string or []string), `query` (string) | Executes Azure ResourceGraph query against selected subscriptions (as string comma separated or string array) |

### MsGraph (AzureAD) functions

:information_source: Functions can also be used starting with `msGraph` prefix instead of `mg`

| Function                          | Parameters             | Description                                                                                                                                                          |
|-----------------------------------|------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `mgUserByUserPrincipalName`       | `userPrincipalName`    | Fetches one user by UserPrincipalName                                                                                                                                |
| `mgUserList`                      | `filter` (string)      | Fetches list of users based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query             |
| `mgGroupByDisplayName`            | `displayName` (string) | Fetches one group by displayName                                                                                                                                     |
| `mgGroupList`                     | `filter` (string)      | Fetches list of groups based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query            |
| `mgServicePrincipalByDisplayName` | `displayName` (string) | Fetches one serviceprincipal by displayName                                                                                                                          |
| `mgServicePrincipalList`          | `filter` (string)      | Fetches list of servicePrincipals based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query |
| `mgApplicationByDisplayName`      | `displayName` (string) | Fetches one application by displayName                                                                                                                               |
| `mgApplicationList`               | `filter` (string)      | Fetches list of applications based on [`$filter`](https://docs.microsoft.com/en-us/graph/filter-query-parameter#examples-using-the-filter-query-operator) query      |

## Time template functions

| Function       | Parameters                     | Description                                 |
|----------------|--------------------------------|---------------------------------------------|
| `fromUnixtime` | `timestamp` (int/float/string) | Converts unixtimestamp to Time object       |
| `toRFC3339`    | `time` (time.Time)             | Converts time object to RFC3339 time string |

## Misc template functions

| Function    | Parameters          | Description                                                                          |
|-------------|---------------------|--------------------------------------------------------------------------------------|
| `jsonPath`  | `jsonPath` (string) | Fetches object information using jsonPath (useful to process `azureResource` output) |
| `filesGet`  | `path` (string)     | Fetches content of file and returns content as string                                |
| `filesGlob` | `pattern` (string)  | Lists files using glob pattern                                                       |

```gotemplate

{{
    azResource
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
{{ azResource
   "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
   "2022-01-01"
   | toYaml
}}

## Fetch resource as object, select .properties.aadProfile via jsonPath and convert to yaml
{{ azResource
   "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
   "2022-01-01"
   | jsonPath "$.properties.aadProfile"
   | toYaml
}}

## Fetches all resources from subscription
{{ (azResourceList "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx") | toYaml }}

## Fetches all virtualNetwork resources from subscription
{{ (azResourceList "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx" "resourceType eq 'Microsoft.Network/virtualNetworks'") | toYaml }}

## Fetches all resources from resourceGroup
{{ (azResourceList "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg") | toYaml }}

## Fetch Azure VirtualNetwork address prefixes
{{ azVirtualNetworkAddressPrefixes
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
}}xxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
   "default2"
   | join ","
}}

## Fetch first storageaccount key
{{ (index (azStorageAccountAccessKeys "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourceGroups/example-rg/providers/Microsoft.Storage/storageAccounts/foobar") 0).value }}

## fetch blob from storageaccount container
{{ azureStorageAccountContainerBlob "https://foobar.blob.core.windows.net/examplecontainer/file.json" }}

## Fetch secret value from Azure KeyVault (using only name; only AzurePublicCloud, AzureChinaCloud and AzureGovernmentCloud)
{{ (azKeyVaultSecret "examplevault" "secretname").value }}
{{ (azKeyVaultSecret "examplevault" "secretname").attributes.exp | fromUnixtime | toRFC3339 }}

## Fetch secret value from Azure KeyVault (using full url)
{{ (azKeyVaultSecret "https://examplevault.vault.azure.net/" "secretname").value }}

## Fetch current environmentName
{{ azAccountInfo.environmentName }}

## Fetch current tenantId
{{ azAccountInfo.tenantId }}

## Fetch current selected subscription displayName
{{ azSubscription.displayName }}

## Fetch RoleDefinition id for "owner" role
{{ (azRoleDefinition "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx" "Owner").name }}

## Executes ResourceGraph query and returns result as yaml
{{ azResourceGraphQuery "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx"  `resources | where resourceGroup contains "xxxx"` | toYaml }}

```


PS: some code is borrowed from [Helm](https://github.com/helm/helm)
