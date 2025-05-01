package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/webdevops/go-common/azuresdk/azidentity"

	"github.com/webdevops/helm-azure-tpl/config"
)

const (
	Author    = "webdevops.io"
	UserAgent = "helm-azure-tpl/"

	TermColumns = 80
)

var (
	argparser *flags.Parser

	azAccountInfo map[string]interface{}

	startTime time.Time

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

var (
	opts config.Opts
)

func main() {
	startTime = time.Now()
	initArgparser()
	initLogger()
	initSystem()
	initAzureEnvironment()
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
}

func fetchAzAccountInfo() {
	cmd := exec.Command("az", "account", "show", "-o", "json")
	cmd.Stderr = os.Stderr

	accountInfo, err := cmd.Output()
	if err != nil {
		logger.Fatalf(`unable to detect Azure TenantID via 'az account show': %v`, err)
	}

	err = json.Unmarshal(accountInfo, &azAccountInfo)
	if err != nil {
		logger.Fatalf(`unable to parse 'az account show' output: %v`, err)
	}

	// auto set azure tenant id
	if opts.Azure.Environment == nil || *opts.Azure.Environment == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["environmentName"].(string); ok {
			logger.Infof(`use Azure Environment '%v' from 'az account show'`, val)
			opts.Azure.Environment = &val
		}
	}

	// auto set azure tenant id
	if opts.Azure.Tenant == nil || *opts.Azure.Tenant == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["tenantId"].(string); ok {
			logger.Infof(`use Azure TenantID '%v' from 'az account show'`, val)
			opts.Azure.Tenant = &val
		}
	}

	setOsEnvIfUnset(azidentity.EnvAzureEnvironment, *opts.Azure.Environment)
	setOsEnvIfUnset(azidentity.EnvAzureTenantID, *opts.Azure.Tenant)
}

func initAzureEnvironment() {
	if opts.Azure.Environment == nil || *opts.Azure.Environment == "" {
		// autodetect tenant
		if val, ok := azAccountInfo["environmentName"].(string); ok {
			logger.Infof(`use Azure Environment '%v' from 'az account show'`, val)
			opts.Azure.Environment = &val
		}
	}

	if opts.Azure.Environment != nil {
		if err := os.Setenv(azidentity.EnvAzureEnvironment, *opts.Azure.Environment); err != nil {
			logger.Warnf(`unable to set envvar AZURE_ENVIRONMENT: %v`, err.Error())
		}
	}
}

func setOsEnvIfUnset(name, value string) {
	if envVal := os.Getenv(name); envVal == "" {
		if err := os.Setenv(name, value); err != nil {
			panic(err)
		}
	}
}
