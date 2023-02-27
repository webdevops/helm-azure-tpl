package azuretpl

func (e *AzureTemplateExecutor) azAccountInfo() (interface{}, error) {
	return e.azureCliAccountInfo, nil
}
