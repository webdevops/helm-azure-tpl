package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/msgraphsdk/hamiltonclient"

	"github.com/webdevops/helm-azure-tpl/azuretpl"
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
	MsGraphClient *hamiltonclient.MsGraphClient

	azAccountInfo AzureCliAccount

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

var (
	opts config.Opts
	args []string
)

func main() {
	initArgparser()

	log.Infof("helm-azure-tpl v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	log.Info(string(opts.GetJson()))

	// we're going to use az cli auth here
	if err := os.Setenv("AZURE_AUTH", "az"); err != nil {
		log.Panic(err.Error())
	}

	log.Infof("detecting Azure account information")
	fetchAzAccountInfo()

	log.Infof("connecting to Azure")
	initAzureConnection()

	log.Infof("connecting to MsGraph")
	initMsGraphConnection()

	ctx := context.Background()

	for _, filePath := range args {
		sourcePath := filePath
		targetPath := filePath
		if strings.Contains(sourcePath, ":") {
			parts := strings.SplitN(sourcePath, ":", 2)
			sourcePath = parts[0]
			targetPath = parts[1]
		}

		contextLogger := log.WithFields(log.Fields{
			`sourcePath`: sourcePath,
			`targetPath`: targetPath,
		})
		contextLogger.Infof(`processing file`)

		azureTemplate := azuretpl.New(ctx, AzureClient, MsGraphClient, contextLogger)
		tmpl := template.New("helm-azuretpl-tpl").Funcs(sprig.TxtFuncMap()).Funcs(azureTemplate.TxtFuncMap())

		content, err := os.ReadFile(sourcePath) // #nosec G304 passed as parameter
		if err != nil {
			contextLogger.Fatalf(`unable to read file: %v`, err.Error())
		}

		parsedContent, err := tmpl.Parse(string(content))
		if err != nil {
			contextLogger.Fatalf(`unable to parse file: %v`, err.Error())
		}

		var buf bytes.Buffer
		err = parsedContent.Execute(&buf, nil)
		if err != nil {
			contextLogger.Fatalf(`unable to process template: %v`, err.Error())
		}

		if opts.Debug {
			fmt.Println()
			fmt.Println(strings.Repeat("-", TermColumns))
			fmt.Printf("--- %v\n", targetPath)
			fmt.Println(strings.Repeat("-", TermColumns))
			fmt.Println(buf.String())
		}

		if !opts.DryRun {
			contextLogger.Infof(`writing file "%v"`, targetPath)
			err := os.WriteFile(targetPath, buf.Bytes(), 0600)
			if err != nil {
				contextLogger.Fatalf(`unable to write target file "%v": %v`, targetPath, err.Error())
			}
		}

	}
}

func initArgparser() {
	var err error
	argparser = flags.NewParser(&opts, flags.Default)
	args, err = argparser.Parse()

	// check if there is a parse error
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

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
	cmd := exec.Command("az", "account", "show", "-o", "json")
	cmd.Stderr = os.Stderr

	accountInfo, err := cmd.Output()
	if err != nil {
		log.Fatalf(`unable to detect Azure TenantID via "az account show": %v`, err.Error())
	}

	err = json.Unmarshal(accountInfo, &azAccountInfo)
	if err != nil {
		log.Fatalf(`unable to parse "az account show" output: %v`, err.Error())
	}
}

func initAzureConnection() {
	var err error

	if opts.Azure.Environment == nil || *opts.Azure.Environment == "" {
		// autodetect tenant
		log.Infof(`use Azure Environment "%v" from "az account show"`, azAccountInfo.EnvironmentName)
		opts.Azure.Environment = &azAccountInfo.EnvironmentName
	}

	AzureClient, err = armclient.NewArmClientWithCloudName(*opts.Azure.Environment, log.StandardLogger())
	if err != nil {
		log.Panic(err.Error())
	}

	AzureClient.SetUserAgent(UserAgent + gitTag)
}

func initMsGraphConnection() {
	var err error

	if opts.Azure.Tenant == nil || *opts.Azure.Tenant == "" {
		// autodetect tenant
		log.Infof(`use Azure TenantID "%v" from "az account show"`, azAccountInfo.TenantID)
		opts.Azure.Tenant = &azAccountInfo.TenantID
	}

	if MsGraphClient == nil {
		MsGraphClient, err = hamiltonclient.NewMsGraphClientWithCloudName(*opts.Azure.Environment, *opts.Azure.Tenant, log.StandardLogger())
		if err != nil {
			log.Panic(err.Error())
		}

		MsGraphClient.SetUserAgent(UserAgent + gitTag)
	}
}
