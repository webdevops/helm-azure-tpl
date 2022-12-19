package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/msgraphsdk/msgraphclient"

	"github.com/webdevops/helm-azure-tpl/config"
)

const (
	Author    = "webdevops.io"
	UserAgent = "helm-azure-tpl/"

	TermColumns = 80
)

var (
	argparser *flags.Parser

	AzureClient   *armclient.ArmClient
	MsGraphClient *msgraphclient.MsGraphClient

	azAccountInfo map[string]interface{}

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

var (
	opts config.Opts
)

func main() {
	initArgparser()
	run()
}

func initArgparser() {
	var err error
	argparser = flags.NewParser(&opts, flags.Default)

	// check if run by helm
	if helmCmd := os.Getenv("HELM_BIN"); helmCmd != "" {
		if pluginName := os.Getenv("HELM_PLUGIN_NAME"); pluginName != "" {
			argparser.Command.Name = fmt.Sprintf(`%v %v`, helmCmd, pluginName)
		}
	}
	_, err = argparser.Parse()

	// check if there is a parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	if opts.Stdout {
		// send all logs to stderr if we're using stdout output
		log.SetOutput(os.Stderr)
	}

	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		DisableQuote:     true,
	})

	// verbose level
	if opts.Logger.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// debug level
	if opts.Debug {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
		log.SetFormatter(&log.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	}

	// json log format
	if opts.Logger.LogJson {
		log.SetReportCaller(true)
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	}
}

func fetchAzAccountInfo() {
	// enforce Azure AZ auth
	if err := os.Setenv("AZURE_AUTH", "az"); err != nil {
		log.Fatalf(`unable to set AZURE_AUTH: %v`, err.Error())
	}

	cmd := exec.Command("az", "account", "show", "-o", "json")
	cmd.Stderr = os.Stderr

	accountInfo, err := cmd.Output()
	if err != nil {
		log.Fatalf(`unable to detect Azure TenantID via 'az account show': %v`, err)
	}

	err = json.Unmarshal(accountInfo, &azAccountInfo)
	if err != nil {
		log.Fatalf(`unable to parse 'az account show' output: %v`, err)
	}

	// auto set azure tenant id
	if opts.Azure.Tenant == nil || *opts.Azure.Tenant == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["tenantId"].(string); ok {
			log.Infof(`use Azure TenantID '%v' from 'az account show'`, val)
			opts.Azure.Tenant = &val

			// ensure that tenant id is also populated in env settings
			if err := os.Setenv("AZURE_TENANT_ID", val); err != nil {
				log.Fatalf(`unable to set AZURE_TENANT_ID: %v`, err.Error())
			}
		}
	}
}

func initAzureConnection() {
	var err error

	if opts.Azure.Environment == nil || *opts.Azure.Environment == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["environmentName"].(string); ok {
			log.Infof(`use Azure Environment '%v' from 'az account show'`, val)
			opts.Azure.Environment = &val
		}
	}

	AzureClient, err = armclient.NewArmClientWithCloudName(*opts.Azure.Environment, log.StandardLogger())
	if err != nil {
		log.Panic(err.Error())
	}

	AzureClient.SetUserAgent(UserAgent + gitTag)
}

func initMsGraphConnection() {
	var err error

	if MsGraphClient == nil {
		MsGraphClient, err = msgraphclient.NewMsGraphClientWithCloudName(*opts.Azure.Environment, *opts.Azure.Tenant, log.StandardLogger())
		if err != nil {
			log.Panic(err.Error())
		}

		MsGraphClient.SetUserAgent(UserAgent + gitTag)
	}
}
