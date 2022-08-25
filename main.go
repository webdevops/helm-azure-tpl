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

	log.Infof("starting helm-azuretpl-tpl v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	log.Info(string(opts.GetJson()))

	log.Infof("init Azure connection")
	initAzureConnection()

	ctx := context.Background()

	for _, filePath := range args {
		targetFile := filePath
		if strings.Contains(filePath, ":") {
			parts := strings.SplitN(filePath, ":", 2)
			filePath = parts[0]
			targetFile = parts[1]
		}

		contextLogger := log.WithField(`file`, filePath)
		azureTemplate := azuretpl.New(ctx, AzureClient, contextLogger)
		tmpl := template.New("helm-azuretpl-tpl").Funcs(sprig.TxtFuncMap()).Funcs(azureTemplate.TxtFuncMap())

		content, err := os.ReadFile(filePath) // #nosec G304 passed as parameter
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
			contextLogger.Infof(`writing file "%v"`, targetFile)
			err := os.WriteFile(targetFile, buf.Bytes(), 0600)
			if err != nil {
				contextLogger.Fatalf(`unable to write target file "%v": %v`, targetFile, err.Error())
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
	AzureClient, err = armclient.NewArmClientWithCloudName(*opts.Azure.Environment, log.StandardLogger())
	if err != nil {
		log.Panic(err.Error())
	}

	AzureClient.SetUserAgent(UserAgent + gitTag)
}
