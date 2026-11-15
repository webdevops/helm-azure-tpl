package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/strvals"
)

const (
	CommandHelp    = "help"
	CommandVersion = "version"
	CommandLint    = "lint"
	CommandProcess = "apply"
)


func printAppHeader() {
	logger.Info(fmt.Sprintf("%v v%s (%s; %s; by %v at %v)", argparser.Name, gitTag, gitCommit, runtime.Version(), Author, buildDate))
	logger.Info(string(opts.GetJson()))
}

// borrowed from helm/helm
// https://github.com/helm/helm/blob/main/pkg/cli/values/options.go
// Apache License, Version 2.0
func readValuesFiles() error {
	templateDataValues := map[string]interface{}{}

	for _, filePath := range opts.AzureTpl.ValuesFiles {
		currentMap := map[string]interface{}{}

		contextLogger := logger.With(slog.String(`path`, filePath))

		contextLogger.Info("using .Values file")
		data, err := os.ReadFile(filePath)
		if err != nil {
			contextLogger.Error(`unable to read values file: %v`, slog.Any("error", err))
			os.Exit(1)
		}
		err = yaml.Unmarshal(data, &currentMap)
		if err != nil {
			logger.Error("error: %v", slog.Any("error", err))
			os.Exit(1)
		}
		// Merge with the previous map
		templateDataValues = mergeMaps(templateDataValues, currentMap)
	}

	// User specified a value via --set-json
	for _, value := range opts.AzureTpl.JSONValues {
		if err := strvals.ParseJSON(value, templateDataValues); err != nil {
			return fmt.Errorf(`failed parsing --set-json data %s`, value)
		}
	}

	// User specified a value via --set
	for _, value := range opts.AzureTpl.Values {
		if err := strvals.ParseInto(value, templateDataValues); err != nil {
			return fmt.Errorf(`failed parsing --set data: %w`, err)
		}
	}

	// User specified a value via --set-string
	for _, value := range opts.AzureTpl.StringValues {
		if err := strvals.ParseIntoString(value, templateDataValues); err != nil {
			return fmt.Errorf(`failed parsing --set-string data: %w`, err)
		}
	}

	// User specified a value via --set-file
	for _, value := range opts.AzureTpl.FileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := os.ReadFile(string(rs))
			if err != nil {
				return nil, err
			}
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, templateDataValues, reader); err != nil {
			return fmt.Errorf(`failed parsing --set-file data: %w`, err)
		}
	}

	templateData["Values"] = templateDataValues

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

		contextLogger := logger.With(slog.String(`template`, sourcePath))

		if !opts.Stdout {
			contextLogger = contextLogger.With(slog.String(`targetPath`, targetPath))

			if targetPath == "" || targetPath == "." || targetPath == "/" {
				contextLogger.Error(`invalid path detected`, slog.String("path", targetPath))
				os.Exit(1)
			}
		}

		if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
			logger.Error(err.Error())
			os.Exit(1)
		}

		var templateBasePath string
		if opts.Template.BasePath != nil {
			templateBasePath = *opts.Template.BasePath
		} else {
			if val, err := filepath.Abs(sourcePath); err == nil {
				templateBasePath = filepath.Dir(val)
			} else {
				logger.Error(`unable to resolve file`, slog.Any("error", err))
				os.Exit(1)
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
