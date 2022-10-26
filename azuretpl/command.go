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
	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/msgraphsdk/msgraphclient"
)

type (
	AzureTemplateExecutor struct {
		ctx           context.Context
		azureClient   *armclient.ArmClient
		msGraphClient *msgraphclient.MsGraphClient
		logger        *log.Entry

		cache    *cache.Cache
		cacheTtl time.Duration

		TemplateBasePath string

		LintMode bool

		azureCliAccountInfo map[string]interface{}
	}
)

func New(ctx context.Context, azureClient *armclient.ArmClient, msGraphClient *msgraphclient.MsGraphClient, logger *log.Entry) *AzureTemplateExecutor {
	e := &AzureTemplateExecutor{
		ctx:           ctx,
		azureClient:   azureClient,
		msGraphClient: msGraphClient,
		logger:        logger,

		cacheTtl: 15 * time.Minute,
	}
	e.init()
	return e
}

func (e *AzureTemplateExecutor) init() {
	e.cache = globalCache
}

func (e *AzureTemplateExecutor) SetAzureCliAccountInfo(accountInfo map[string]interface{}) {
	e.azureCliAccountInfo = accountInfo
}

func (e *AzureTemplateExecutor) SetTemplateBasePath(val string) {
	e.TemplateBasePath = val
}

func (e *AzureTemplateExecutor) SetLintMode(val bool) {
	e.LintMode = val
}

func (e *AzureTemplateExecutor) TxtFuncMap(tmpl *template.Template) template.FuncMap {
	includedNames := make(map[string]int)

	funcMap := map[string]interface{}{
		// azure
		`azureKeyVaultSecret`:                      e.azureKeyVaultSecret,
		`azureResource`:                            e.azureResource,
		`azureSubscription`:                        e.azureSubscription,
		`azureSubscriptionList`:                    e.azureSubscriptionList,
		`azurePublicIpAddress`:                     e.azurePublicIpAddress,
		`azurePublicIpPrefixAddressPrefix`:         e.azurePublicIpPrefixAddressPrefix,
		`azureVirtualNetworkAddressPrefixes`:       e.azureVirtualNetworkAddressPrefixes,
		`azureVirtualNetworkSubnetAddressPrefixes`: e.azureVirtualNetworkSubnetAddressPrefixes,
		`azureAccountInfo`:                         e.azureAccountInfo,

		// msGraph
		`msGraphUserByUserPrincipalName`:       e.msGraphUserByUserPrincipalName,
		`msGraphUserList`:                      e.msGraphUserList,
		`msGraphGroupByDisplayName`:            e.msGraphGroupByDisplayName,
		`msGraphGroupList`:                     e.msGraphGroupList,
		`msGraphServicePrincipalByDisplayName`: e.msGraphServicePrincipalByDisplayName,
		`msGraphServicePrincipalList`:          e.msGraphServicePrincipalList,
		`msGraphApplicationByDisplayName`:      e.msGraphApplicationByDisplayName,
		`msGraphApplicationList`:               e.msGraphApplicationList,

		// misc
		`jsonPath`: e.jsonPath,

		// borrowed from github.com/helm/helm
		"toYaml":        toYAML,
		"fromYaml":      fromYAML,
		"fromYamlArray": fromYAMLArray,
		"toJson":        toJSON,
		"fromJson":      fromJSON,
		"fromJsonArray": fromJSONArray,

		// files
		"filesGet": e.filesGet,

		"include": func(path string, data interface{}) (string, error) {
			var sourcePath string
			if !filepath.IsAbs(path) {
				sourcePath = filepath.Clean(fmt.Sprintf("%s/%s", e.TemplateBasePath, path))
			} else {
				sourcePath = filepath.Clean(path)
			}

			if val, err := filepath.Abs(sourcePath); err == nil {
				sourcePath = val
			} else {
				return "", fmt.Errorf(`unable to resolve include referance: %w`, err)
			}

			if !strings.HasPrefix(sourcePath, e.TemplateBasePath) {
				return "", fmt.Errorf(
					`'%v' must be in same directory or below (expected prefix: %v, got: %v)`,
					path,
					e.TemplateBasePath,
					filepath.Dir(sourcePath),
				)
			}

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
			err = parsedContent.Execute(&buf, nil)
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

	return funcMap
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

func (e *AzureTemplateExecutor) fetchAzureResource(resourceID string, apiVersion string) (interface{}, error) {
	resourceInfo, err := armclient.ParseResourceId(resourceID)
	if err != nil {
		return nil, fmt.Errorf(`unable to parse Azure resourceID '%v': %w`, resourceID, err)
	}

	if val, enabled := e.lintResult(); enabled {
		return val, nil
	}

	client, err := armresources.NewClient(resourceInfo.Subscription, e.azureClient.GetCred(), e.azureClient.NewArmClientOptions())
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
