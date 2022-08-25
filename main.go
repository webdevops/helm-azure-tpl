package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"

	"github.com/webdevops/helm-azure-tpl/azuretpl"
	"github.com/webdevops/helm-azure-tpl/config"
)

const (
	Author    = "webdevops.io"
	UserAgent = "helm-azure-tpl/"
)

var (
	argparser *flags.Parser

	AzureClient *armclient.ArmClient

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

	log.Infof("connecting to Azure")
	initAzureConnection()

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

		azureTemplate := azuretpl.New(ctx, AzureClient, contextLogger)
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

	// check if there is an parse error
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
	if opts.Logger.Debug {
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

func initAzureConnection() {
	var err error

	// we're going to use az cli auth here
	err = os.Setenv("AZURE_AUTH", "az")
	if err != nil {
		log.Panic(err.Error())
	}

	AzureClient, err = armclient.NewArmClientWithCloudName(*opts.Azure.Environment, log.StandardLogger())
	if err != nil {
		log.Panic(err.Error())
	}

	AzureClient.SetUserAgent(UserAgent + gitTag)
}
