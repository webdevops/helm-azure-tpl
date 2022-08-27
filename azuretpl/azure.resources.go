package azuretpl

// azureResource fetches resource json from Azure REST API using the specified apiVersion
func (e *AzureTemplateExecutor) azureResource(resourceID string, apiVersion string) (interface{}, error) {
	e.logger.Infof(`fetching Azure Resource '%v' in apiVersion '%v'`, resourceID, apiVersion)

	cacheKey := generateCacheKey(`azureResource`, resourceID, apiVersion)
	return e.cacheResult(cacheKey, func() (interface{}, error) {
		return e.fetchAzureResource(resourceID, apiVersion)
	})
}
