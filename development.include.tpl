{{ azureSubscription | required "need subscription" }}

{{
    azureResource
    "/subscriptions/d86bcf13-ddf7-45ea-82f1-6f656767a318/resourcegroups/k8s/providers/Microsoft.ContainerService/managedClusters/mblaschke"
    "2022-01-01"
    | jsonPath "$.properties.aadProfile"
    | toYaml | nindent 2
}}


{{ azureSubscriptionList }}

{{ (azureKeyVaultSecret "https://blascma-testvault.vault.azure.net/" "qweqwe").Value }}
{{ (azureKeyVaultSecret "https://blascma-testvault.vault.azure.net/" "qweqwe").Value }}
{{ (azureKeyVaultSecret "https://blascma-testvault.vault.azure.net/" "qweqwe").Value }}
