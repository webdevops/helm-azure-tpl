package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/strvals"
)

const (
	CommandHelp    = "help"
	CommandVersion = "version"
	CommandLint    = "lint"
	CommandProcess = "apply"
)

type (
	TemplatePayload struct {
		Values map[string]interface{}
	}
)

var (
	templateData TemplatePayload
	lintMode     = false
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

		if err := readValuesFiles(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		templateFileList := buildSourceTargetList()

		if !lintMode {
			log.Infof("detecting Azure account information")
			fetchAzAccountInfo()

			azAccountInfoJson, err := json.Marshal(azAccountInfo)
			if err == nil {
				log.Infof(string(azAccountInfoJson))
			}
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

// borrowed from helm/helm
// https://github.com/helm/helm/blob/main/pkg/cli/values/options.go
// Apache License, Version 2.0
func readValuesFiles() error {
	templateData.Values = map[string]interface{}{}
	for _, filePath := range opts.ValuesFiles {
		currentMap := map[string]interface{}{}

		contextLogger := log.WithFields(log.Fields{
			`valuesPath`: filePath,
		})

		contextLogger.Info("using .Values file")
		data, err := os.ReadFile(filePath)
		if err != nil {
			contextLogger.Fatalf(`unable to read values file: %v`, err)
		}
		err = yaml.Unmarshal(data, &currentMap)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		// Merge with the previous map
		templateData.Values = mergeMaps(templateData.Values, currentMap)
	}

	// User specified a value via --set-json
	for _, value := range opts.JSONValues {
		if err := strvals.ParseJSON(value, templateData.Values); err != nil {
			return fmt.Errorf(`failed parsing --set-json data %s`, value)
		}
	}

	// User specified a value via --set
	for _, value := range opts.Values {
		if err := strvals.ParseInto(value, templateData.Values); err != nil {
			return fmt.Errorf(`failed parsing --set data: %w`, err)
		}
	}

	// User specified a value via --set-string
	for _, value := range opts.StringValues {
		if err := strvals.ParseIntoString(value, templateData.Values); err != nil {
			return fmt.Errorf(`failed parsing --set-string data: %w`, err)
		}
	}

	// User specified a value via --set-file
	for _, value := range opts.FileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := os.ReadFile(string(rs))
			if err != nil {
				return nil, err
			}
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, templateData.Values, reader); err != nil {
			return fmt.Errorf(`failed parsing --set-file data: %w`, err)
		}
	}

	if opts.Debug {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, strings.Repeat("-", TermColumns))
		fmt.Fprintln(os.Stderr, "--- VALUES")
		fmt.Fprintln(os.Stderr, strings.Repeat("-", TermColumns))
		values, _ := yaml.Marshal(templateData)
		fmt.Fprintln(os.Stderr, string(values))
	}

	return nil
}

// borrowed from helm/helm
// https://github.com/helm/helm/blob/main/pkg/cli/values/options.go
// Apache License, Version 2.0
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func buildSourceTargetList() (list []TemplateFile) {
	ctx := context.Background()

	for _, filePath := range opts.Args.Files {
		var targetPath string
		sourcePath := filePath

		// remove protocol prefix (when using helm downloader)
		sourcePath = strings.TrimPrefix(sourcePath, "azuretpl://")
		sourcePath = strings.TrimPrefix(sourcePath, "azure-tpl://")

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
		})

		if !opts.Stdout {
			contextLogger = contextLogger.WithFields(log.Fields{
				`targetPath`: targetPath,
			})

			if targetPath == "" || targetPath == "." || targetPath == "/" {
				contextLogger.Fatalf(`invalid path '%v' detected`, targetPath)
			}
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
