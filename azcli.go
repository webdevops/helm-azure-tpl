package main

type (
	AzureCliAccount struct {
		EnvironmentName  string        `json:"environmentName"`
		HomeTenantID     string        `json:"homeTenantId"`
		ID               string        `json:"id"`
		IsDefault        bool          `json:"isDefault"`
		ManagedByTenants []interface{} `json:"managedByTenants"`
		Name             string        `json:"name"`
		State            string        `json:"state"`
		TenantID         string        `json:"tenantId"`
		User             struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"user"`
	}
)
