package azure

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	textTemplate "text/template"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/msgraphsdk/msgraphclient"

	"github.com/webdevops/helm-azure-tpl/pkg/engine/base"
)

type (
	AzureEngine struct {
		base.Engine
		*base.BaseEngine

		azureTokenCredential         *azcore.TokenCredential
		azureTokenCredentialCallback func() *azcore.TokenCredential

		clients struct {
			azure   *armclient.ArmClient
			msGraph *msgraphclient.MsGraphClient
		}

		azureCliAccountInfo map[string]interface{}
	}
)

func New(base *base.BaseEngine) *AzureEngine {
	ret := AzureEngine{
		BaseEngine: base,
	}
	return &ret
}

func (e *AzureEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	funcMap := textTemplate.FuncMap{
		// azure
		`azResource`:                            e.azResource,
		`azResourceList`:                        e.azResourceList,
		`azManagementGroup`:                     e.azManagementGroup,
		`azManagementGroupSubscriptionList`:     e.azManagementGroupSubscriptionList,
		`azSubscription`:                        e.azSubscription,
		`azSubscriptionList`:                    e.azSubscriptionList,
		`azPublicIpAddress`:                     e.azPublicIpAddress,
		`azPublicIpPrefixAddressPrefix`:         e.azPublicIpPrefixAddressPrefix,
		`azVirtualNetworkAddressPrefixes`:       e.azVirtualNetworkAddressPrefixes,
		`azVirtualNetworkSubnetAddressPrefixes`: e.azVirtualNetworkSubnetAddressPrefixes,
		`azAccountInfo`:                         e.azAccountInfo,

		// azure keyvault
		`azKeyVaultSecret`:         e.azKeyVaultSecret,
		`azKeyVaultSecretVersions`: e.azKeyVaultSecretVersions,
		`azKeyVaultSecretList`:     e.azKeyVaultSecretList,

		// azure redis
		`azRedisAccessKeys`: e.azRedisAccessKeys,

		// azure storageAccount
		`azStorageAccountAccessKeys`:    e.azStorageAccountAccessKeys,
		`azStorageAccountContainerBlob`: e.azStorageAccountContainerBlob,

		// azure eventhub
		`azEventHubListByNamespace`: e.azEventHubListByNamespace,

		// azure app config
		`azAppConfigSetting`: e.azAppConfigSetting,

		// azure managedCluster
		`azManagedClusterUserCredentials`: e.azManagedClusterUserCredentials,

		// resourcegraph
		`azResourceGraphQuery`: e.azResourceGraphQuery,

		// rbac
		`azRoleDefinition`:     e.azRoleDefinition,
		`azRoleDefinitionList`: e.azRoleDefinitionList,

		// msGraph
		`mgUserByUserPrincipalName`:       e.mgUserByUserPrincipalName,
		`mgUserList`:                      e.mgUserList,
		`mgGroupByDisplayName`:            e.mgGroupByDisplayName,
		`mgGroupList`:                     e.mgGroupList,
		`mgServicePrincipalByDisplayName`: e.mgServicePrincipalByDisplayName,
		`mgServicePrincipalList`:          e.mgServicePrincipalList,
		`mgApplicationByDisplayName`:      e.mgApplicationByDisplayName,
		`mgApplicationList`:               e.mgApplicationList,
	}

	// automatic add legacy funcs
	tmp := textTemplate.FuncMap{}
	for funcName, funcCallback := range funcMap {
		// azFunc -> azureFunc
		if after, ok := strings.CutPrefix(funcName, "az"); ok {
			oldFuncName := "azure" + after
			tmp[oldFuncName] = e.Deprecate(oldFuncName, funcCallback)
		}

		// mgFunc -> msGraphFunc
		if after, ok := strings.CutPrefix(funcName, "mg"); ok {
			oldFuncName := "msGraph" + after
			tmp[oldFuncName] = e.Deprecate(oldFuncName, funcCallback)
		}

		tmp[funcName] = funcCallback
	}
	funcMap = tmp

	return funcMap
}

// getAzureTokenCredential creates on-demand azure armclient and returns it
func (e *AzureEngine) getAzureTokenCredential() *azcore.TokenCredential {
	// use callback first
	if e.azureTokenCredential == nil {
		e.azureTokenCredential = e.azureTokenCredentialCallback()
	}

	return e.azureTokenCredential
}

// azureClient creates on-demand azure armclient and returns it
func (e *AzureEngine) azureClient() *armclient.ArmClient {
	var err error

	if e.clients.azure == nil {
		// load az account details here
		e.clients.azure, err = armclient.NewArmClientFromEnvironment(e.Logger.Slog(), armclient.WithCred(e.getAzureTokenCredential()))
		if err != nil {
			e.Logger.Error(err.Error())
			os.Exit(1)
		}

		e.clients.azure.SetUserAgent(e.UserAgent)
	}
	return e.clients.azure
}

// msGraphClient creates on-demand azure msgraph and returns it
func (e *AzureEngine) msGraphClient() *msgraphclient.MsGraphClient {
	var err error
	if e.clients.msGraph == nil {
		// ensure azureclient init
		e.azureClient()

		e.clients.msGraph, err = msgraphclient.NewMsGraphClientFromEnvironment(e.Logger.Slog(), msgraphclient.WithCred(e.getAzureTokenCredential()))
		if err != nil {
			e.Logger.Error(err.Error())
			os.Exit(1)
		}

		e.clients.msGraph.SetUserAgent(e.UserAgent)
	}
	return e.clients.msGraph
}

// SetAzureTokenCredential sets Azure token credential (from azcore package)
func (e *AzureEngine) SetAzureTokenCredential(cred *azcore.TokenCredential) {
	e.azureTokenCredential = cred
	e.clients.azure = nil
	e.clients.msGraph = nil
}

// SetAzureTokenCredentialCallback sets Azure token credential (from azcore package) using a callback function
func (e *AzureEngine) SetAzureTokenCredentialCallback(callback func() *azcore.TokenCredential) {
	e.azureTokenCredentialCallback = callback
}

// fetchAzureResource fetches json representation of Azure resource by resourceID and apiVersion
func (e *AzureEngine) fetchAzureResource(resourceID string, apiVersion string) (interface{}, error) {
	resourceInfo, err := armclient.ParseResourceId(resourceID)
	if err != nil {
		return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
	}

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	client, err := armresources.NewClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
	if err != nil {
		return nil, err
	}

	resource, err := client.GetByID(e.Ctx, resourceID, apiVersion, nil)
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch Azure resource '%v': %w`, resourceID, err)
	}

	data, err := resource.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf(`unable to marshal Azure resource '%v': %w`, resourceID, err)
	}

	var resourceRawInfo map[string]interface{}
	err = json.Unmarshal(data, &resourceRawInfo)
	if err != nil {
		return nil, fmt.Errorf(`unable to unmarshal Azure resource '%v': %w`, resourceID, err)
	}

	return resourceRawInfo, nil
}
