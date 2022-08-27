package azuretpl

func (e *AzureTemplateExecutor) azureAccountInfo() (interface{}, error) {
	return e.azureCliAccountInfo, nil
}
