package azuretpl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Masterminds/sprig/v3"
	cache "github.com/patrickmn/go-cache"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/msgraphsdk/msgraphclient"
	"go.uber.org/zap"
)

type (
	AzureTemplateExecutor struct {
		ctx    context.Context
		logger *zap.SugaredLogger

		cache    *cache.Cache
		cacheTtl time.Duration

		UserAgent string

		TemplateRootPath string
		TemplateRelPath  string

		LintMode bool

		azureCliAccountInfo map[string]interface{}
	}
)

var (
	azureClient   *armclient.ArmClient
	msGraphClient *msgraphclient.MsGraphClient
)

func New(ctx context.Context, logger *zap.SugaredLogger) *AzureTemplateExecutor {
	e := &AzureTemplateExecutor{
		ctx:    ctx,
		logger: logger,

		cacheTtl: 15 * time.Minute,
	}
	e.init()
	return e
}

func (e *AzureTemplateExecutor) init() {
	e.cache = globalCache
}

func (e *AzureTemplateExecutor) azureClient() *armclient.ArmClient {
	var err error
	if azureClient == nil {
		azureClient, err = armclient.NewArmClientFromEnvironment(e.logger)
		if err != nil {
			e.logger.Fatal(err.Error())
		}

		azureClient.SetUserAgent(e.UserAgent)
		azureClient.UseAzCliAuth()
		if err := azureClient.Connect(); err != nil {
			e.logger.Fatal(err.Error())
		}
	}
	return azureClient
}

func (e *AzureTemplateExecutor) msGraphClient() *msgraphclient.MsGraphClient {
	var err error
	if msGraphClient == nil {
		// ensure azureclient init
		if azureClient == nil {
			e.azureClient()
		}

		msGraphClient, err = msgraphclient.NewMsGraphClientFromEnvironment(e.logger)
		if err != nil {
			e.logger.Fatal(err.Error())
		}

		msGraphClient.SetUserAgent(e.UserAgent)
		msGraphClient.UseAzCliAuth()
	}
	return msGraphClient
}

func (e *AzureTemplateExecutor) SetUserAgent(val string) {
	e.UserAgent = val
}

func (e *AzureTemplateExecutor) SetAzureCliAccountInfo(accountInfo map[string]interface{}) {
	e.azureCliAccountInfo = accountInfo
}

func (e *AzureTemplateExecutor) SetTemplateRootPath(val string) {
	path, err := filepath.Abs(filepath.Clean(val))
	if err != nil {
		e.logger.Fatalf(`invalid base path '%v': %v`, val, err.Error())
	}
	e.TemplateRootPath = path
}

func (e *AzureTemplateExecutor) SetTemplateRelPath(val string) {
	path, err := filepath.Abs(filepath.Clean(val))
	if err != nil {
		e.logger.Fatalf(`invalid base path '%v': %v`, val, err.Error())
	}
	e.TemplateRelPath = path
}

func (e *AzureTemplateExecutor) SetLintMode(val bool) {
	e.LintMode = val
}

func (e *AzureTemplateExecutor) TxtFuncMap(tmpl *template.Template) template.FuncMap {
	includedNames := make(map[string]int)

	funcMap := map[string]interface{}{
		// azure
		`azResource`:                            e.azResource,
		`azResourceList`:                        e.azResourceList,
		`azSubscription`:                        e.azSubscription,
		`azSubscriptionList`:                    e.azSubscriptionList,
		`azPublicIpAddress`:                     e.azPublicIpAddress,
		`azPublicIpPrefixAddressPrefix`:         e.azPublicIpPrefixAddressPrefix,
		`azVirtualNetworkAddressPrefixes`:       e.azVirtualNetworkAddressPrefixes,
		`azVirtualNetworkSubnetAddressPrefixes`: e.azVirtualNetworkSubnetAddressPrefixes,
		`azAccountInfo`:                         e.azAccountInfo,

		// azure keyvault
		`azKeyVaultSecret`:     e.azKeyVaultSecret,
		`azKeyVaultSecretList`: e.azKeyVaultSecretList,

		// azure storageAccount
		`azStorageAccountAccessKeys`:    e.azStorageAccountAccessKeys,
		`azStorageAccountContainerBlob`: e.azStorageAccountContainerBlob,

		// azure app config
		`azAppConfigSetting`: e.azAppConfigSetting,

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

		// misc
		`jsonPath`: e.jsonPath,

		// time
		`fromUnixtime`: fromUnixtime,
		`toRFC3339`:    toRFC3339,

		// borrowed from github.com/helm/helm
		"toYaml":        toYAML,
		"fromYaml":      fromYAML,
		"fromYamlArray": fromYAMLArray,
		"toJson":        toJSON,
		"fromJson":      fromJSON,
		"fromJsonArray": fromJSONArray,

		// files
		"filesGet":  e.filesGet,
		"filesGlob": e.filesGlob,

		"include": func(path string, data interface{}) (string, error) {
			sourcePath := e.fileMakePathAbs(path)

			if v, ok := includedNames[sourcePath]; ok {
				if v > recursionMaxNums {
					return "", fmt.Errorf(`too many recursions for inclusion of '%v'`, path)
				}
				includedNames[sourcePath]++
			} else {
				includedNames[sourcePath] = 1
			}

			content, err := os.ReadFile(sourcePath)
			if err != nil {
				return "", fmt.Errorf(`unable to read file: %w`, err)
			}

			parsedContent, err := tmpl.Parse(string(content))
			if err != nil {
				return "", fmt.Errorf(`unable to parse file: %w`, err)
			}

			var buf bytes.Buffer
			err = parsedContent.Execute(&buf, data)
			if err != nil {
				return "", fmt.Errorf("unable to process template:\n%w", err)
			}

			includedNames[sourcePath]--
			return buf.String(), nil
		},

		"required": func(message string, val interface{}) (interface{}, error) {
			if val == nil {
				if e.LintMode {
					// Don't fail on missing required values when linting
					e.logger.Infof("[TPL::required] missing required value: %s", message)
					return "", nil
				}
				return val, errors.New(message)
			} else if _, ok := val.(string); ok {
				if val == "" {
					if e.LintMode {
						// Don't fail on missing required values when linting
						e.logger.Infof("[TPL::required] missing required value: %s", message)
						return "", nil
					}
					return val, errors.New(message)
				}
			}
			return val, nil
		},

		"fail": func(message string) (string, error) {
			if e.LintMode {
				// Don't fail when linting
				e.logger.Infof("[TPL::fail] fail: %s", message)
				return "", nil
			}
			return "", errors.New(message)
		},
	}

	// automatic add legacy funcs
	tmp := map[string]interface{}{}
	for funcName, funcCallback := range funcMap {
		// azFunc -> azureFunc
		if strings.HasPrefix(funcName, "az") {
			tmp["azure"+strings.TrimPrefix(funcName, "az")] = funcCallback
		}

		// mgFunc -> msGraphFunc
		if strings.HasPrefix(funcName, "mg") {
			tmp["msGraph"+strings.TrimPrefix(funcName, "mg")] = funcCallback
		}

		tmp[funcName] = funcCallback
	}
	funcMap = tmp

	return funcMap
}

func (e *AzureTemplateExecutor) Parse(path string, templateData interface{}, buf *strings.Builder) error {
	tmpl := template.New(path).Funcs(sprig.TxtFuncMap())
	tmpl = tmpl.Funcs(e.TxtFuncMap(tmpl))

	if !e.LintMode {
		tmpl.Option("missingkey=error")
	} else {
		tmpl.Option("missingkey=zero")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(`unable to read file: '%w'`, err)
	}

	parsedContent, err := tmpl.Parse(string(content))
	if err != nil {
		return fmt.Errorf(`unable to parse file: %w`, err)
	}

	oldPwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err = os.Chdir(e.TemplateRootPath); err != nil {
		return err
	}

	if err = parsedContent.Execute(buf, templateData); err != nil {
		return fmt.Errorf(`unable to process template: '%w'`, err)
	}

	if err = os.Chdir(oldPwd); err != nil {
		return err
	}

	return nil
}

// lintResult checks if lint mode is active and returns example value
func (e *AzureTemplateExecutor) lintResult() (interface{}, bool) {
	if e.LintMode {
		return nil, true
	}
	return nil, false
}

// cacheResult caches template function results (eg. Azure REST API resource information)
func (e *AzureTemplateExecutor) cacheResult(cacheKey string, callback func() (interface{}, error)) (interface{}, error) {
	if val, ok := e.cache.Get(cacheKey); ok {
		e.logger.Infof("found in cache (%v)", cacheKey)
		return val, nil
	}

	ret, err := callback()
	if err != nil {
		return nil, err
	}

	e.cache.SetDefault(cacheKey, ret)

	return ret, nil
}

// fetchAzureResource fetches json representation of Azure resource by resourceID and apiVersion
func (e *AzureTemplateExecutor) fetchAzureResource(resourceID string, apiVersion string) (interface{}, error) {
	resourceInfo, err := armclient.ParseResourceId(resourceID)
	if err != nil {
		return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	client, err := armresources.NewClient(resourceInfo.Subscription, e.azureClient().GetCred(), e.azureClient().NewArmClientOptions())
	if err != nil {
		return nil, err
	}

	resource, err := client.GetByID(e.ctx, resourceID, apiVersion, nil)
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
