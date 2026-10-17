{{
    azResource
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
    "2022-01-01"
    | toYaml | nindent 2
}}


{{
    azResource
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.ContainerService/managedClusters/k8scluster"
    "2022-01-01"
    | jsonPath "$.properties.aadProfile"
    | toYaml | nindent 2
}}

{{ azVirtualNetworkAddressPrefixes
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
}}


{{ azVirtualNetworkSubnetAddressPrefixes
    "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx/resourcegroups/example-rg/providers/Microsoft.Network/virtualNetworks/k8s-vnet"
    "default2"
    | join ","
}}

{{ (azKeyVaultSecret "https://examplevault.vault.azure.net/" "secretname").Value }}
