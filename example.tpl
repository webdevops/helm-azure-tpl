{{
    azureResource
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
    "2022-01-01"
    | toYaml | nindent 2
}}


{{
    azureResource
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
    "2022-01-01"
    | jsonPath "$.properties.aadProfile"
    | toYaml | nindent 2
}}

{{ azureVirtualNetworkAddressPrefixes
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
}}


{{ azureVirtualNetworkSubnetAddressPrefixes
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
    "default2"
    | join ","
}}

{{ (azureKeyVaultSecret "https://examplevault.vault.azure.net/" "secretname").Value }}
