package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
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
		log.Info("enabling lint mode, all functions are in dry mode")
		lintMode = true
		fallthrough
	case CommandProcess:
		printAppHeader()

		if len(opts.Args.Files) == 0 {
			log.Fatal(`no files specified as arguments`)
		}

		templateFileList := buildSourceTargetList()

		if !lintMode {
			log.Infof("detecting Azure account information")
			fetchAzAccountInfo()

			log.Infof("connecting to Azure")
			initAzureConnection()

			log.Infof("connecting to MsGraph")
			initMsGraphConnection()
		}

		for _, templateFile := range templateFileList {
			if lintMode {
				templateFile.Lint()
			} else {
				templateFile.Apply()
			}
		}

		log.Info("finished")
	default:
		fmt.Printf("invalid command '%v'\n", opts.Args.Command)
		fmt.Println()
		argparser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
}

func printAppHeader() {
	log.Infof("%v v%s (%s; %s; by %v)", argparser.Command.Name, gitTag, gitCommit, runtime.Version(), Author)
	log.Info(string(opts.GetJson()))
}

func buildSourceTargetList() (list []TemplateFile) {
	ctx := context.Background()

	for _, filePath := range opts.Args.Files {
		var targetPath string
		sourcePath := filePath

		if strings.Contains(sourcePath, ":") {
			// explicit target path set in argument (source:target)
			parts := strings.SplitN(sourcePath, ":", 2)
			sourcePath = parts[0]
			targetPath = parts[1]
		} else {
			targetPath = sourcePath

			// target not set explicit
			if opts.Target.FileExt != nil {
				// remove file extension
				targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath))
				// adds new file extension
				targetPath = fmt.Sprintf("%s%s", targetPath, *opts.Target.FileExt)
			}

			// automatic target path
			targetPath = fmt.Sprintf(
				"%s%s%s",
				opts.Target.Prefix,
				targetPath,
				opts.Target.Suffix,
			)
		}

		sourcePath = filepath.Clean(sourcePath)
		targetPath = filepath.Clean(targetPath)

		contextLogger := log.WithFields(log.Fields{
			`sourcePath`: sourcePath,
			`targetPath`: targetPath,
		})

		if targetPath == "" || targetPath == "." || targetPath == "/" {
			contextLogger.Fatalf(`invalid path '%v' detected`, targetPath)
		}

		if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
			log.Fatalf(err.Error())
		}

		var templateBasePath string
		if opts.Template.BasePath != nil {
			templateBasePath = *opts.Template.BasePath
		} else {
			if val, err := filepath.Abs(sourcePath); err == nil {
				templateBasePath = filepath.Dir(val)
			} else {
				log.Fatalf(`unable to resolve file: %v`, err)
			}
		}

		list = append(
			list,
			TemplateFile{
				Context:         ctx,
				SourceFile:      sourcePath,
				TargetFile:      targetPath,
				TemplateBaseDir: templateBasePath,
				Logger:          contextLogger,
			},
		)
	}

	return
}
