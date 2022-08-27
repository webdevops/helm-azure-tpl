package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/helm-azure-tpl/azuretpl"
)

const (
	CommandHelp    = "help"
	CommandVersion = "version"
	CommandLint    = "lint"
	CommandProcess = "apply"
)

var (
	lintMode = false
)

func run() {
	switch opts.Args.Command {
	case CommandHelp:
		argparser.WriteHelp(os.Stdout)
		os.Exit(0)
	case CommandVersion:
		fmt.Printf("helm-azure-tpl version: %v (%v, %v)\n", gitTag, gitCommit, runtime.Version())
		os.Exit(0)
	case CommandLint:
		lintMode = true
		fallthrough
	case CommandProcess:
		printAppHeader()

		if len(opts.Args.Files) == 0 {
			log.Fatal(`no files specified as arguments`)
		}

		if !lintMode {
			log.Infof("detecting Azure account information")
			fetchAzAccountInfo()

			log.Infof("connecting to Azure")
			initAzureConnection()

			log.Infof("connecting to MsGraph")
			initMsGraphConnection()
		}
		process()
	default:
		log.Fatalf(`invalid command "%v"`, opts.Args.Command)
	}
}

func printAppHeader() {
	log.Infof("helm-azure-tpl v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	log.Info(string(opts.GetJson()))
}

func process() {
	ctx := context.Background()

	for _, filePath := range opts.Args.Files {
		sourcePath := filePath
		targetPath := sourcePath
		if strings.Contains(sourcePath, ":") {
			parts := strings.SplitN(sourcePath, ":", 2)
			sourcePath = parts[0]
			targetPath = parts[1]
		}

		sourcePath = filepath.Clean(sourcePath)
		targetPath = filepath.Clean(targetPath)

		contextLogger := log.WithFields(log.Fields{
			`sourcePath`: sourcePath,
			`targetPath`: targetPath,
		})

		var templateBasePath string
		if val, err := filepath.Abs(sourcePath); err == nil {
			templateBasePath = filepath.Dir(val)
		} else {
			contextLogger.Fatalf(`unable to resolve file: %v`, err)
		}

		if lintMode {
			contextLogger.Infof(`linting file`)
		} else {
			contextLogger.Infof(`processing file`)
		}

		azureTemplate := azuretpl.New(ctx, AzureClient, MsGraphClient, contextLogger)
		azureTemplate.SetAzureCliAccountInfo(azAccountInfo)
		azureTemplate.SetLintMode(lintMode)
		azureTemplate.SetTemplateBasePath(templateBasePath)
		tmpl := template.New("helm-azuretpl-tpl").Funcs(sprig.TxtFuncMap())
		tmpl = tmpl.Funcs(azureTemplate.TxtFuncMap(tmpl))

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

		if lintMode {
			continue
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
		} else {
			contextLogger.Warn(`not writing file, DRY RUN active`)
		}
	}
}
