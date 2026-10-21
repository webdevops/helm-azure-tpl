package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/webdevops/go-common/azuresdk/azidentity"
)

var (
	azureCred     *azcore.TokenCredential
	azAccountInfo map[string]interface{}
)

// detectAzAccountInfo tries to detect azure account information from az cli and sets the correct environment settings
func detectAzAccountInfo() {
	logger.Info("detecting Azure account information")

	cmd := exec.Command("az", "account", "show", "-o", "json")
	cmd.Stderr = os.Stderr

	accountInfo, err := cmd.Output()
	if err != nil {
		logger.Error(`unable to detect Azure TenantID via 'az account show'`, slog.Any("error", err))
		os.Exit(1)
	}

	err = json.Unmarshal(accountInfo, &azAccountInfo)
	if err != nil {
		logger.Error(`unable to parse 'az account show' output`, slog.Any("error", err))
		os.Exit(1)
	}

	// auto set azure tenant id
	if opts.Azure.Environment == nil || *opts.Azure.Environment == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["environmentName"].(string); ok {
			logger.Info(`detected Azure Environment from 'az account show'`, slog.String("azureEnvironment", val))
			opts.Azure.Environment = &val
		}
	}

	// auto set azure tenant id
	if opts.Azure.Tenant == nil || *opts.Azure.Tenant == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["tenantId"].(string); ok {
			logger.Info(`detected Azure TenantID from 'az account show'`, slog.String("azureTenant", val))
			opts.Azure.Tenant = &val
		}
	}

	setOsEnvIfUnset(azidentity.EnvAzureEnvironment, *opts.Azure.Environment)
	setOsEnvIfUnset(azidentity.EnvAzureTenantID, *opts.Azure.Tenant)
}

// getAzureTokenCredential generates and returns a globally cached azcore.TokenCredential
func getAzureTokenCredential() *azcore.TokenCredential {
	if azureCred == nil {
		logger.Info("establishing Azure connection")
		detectAzAccountInfo()

		cred, err := azidentity.NewAzCliCachedCredential()
		if err != nil {
			panic(err)
		}
		azureCred = &cred
	}
	return azureCred
}
