package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/webdevops/helm-azure-tpl/config"
	"github.com/webdevops/helm-azure-tpl/pkg/summary"
)

const (
	Author    = "webdevops.io"
	UserAgent = "helm-azure-tpl/"

	TermColumns = 80
)

var (
	argparser *flags.Parser

	startTime time.Time

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
	buildDate = "<unknown>"
)

var (
	opts config.Opts

	templateData map[string]interface{}
	lintMode     = false
)

func main() {
	startTime = time.Now()
	initArgparser()
	initLogger()
	run()
}

func initArgparser() {
	var err error
	argparser = flags.NewParser(&opts, flags.Default)

	// check if run by helm
	if helmCmd := os.Getenv("HELM_BIN"); helmCmd != "" {
		if pluginName := os.Getenv("HELM_PLUGIN_NAME"); pluginName != "" {
			argparser.Name = fmt.Sprintf(`%v %v`, helmCmd, pluginName)
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

func run() {

	// init template data
	templateData = make(map[string]interface{})

	lintMode := false
	switch opts.Args.Command {
	case CommandHelp:
		argparser.WriteHelp(os.Stdout)
		os.Exit(0)
	case CommandVersion:
		versionPayload := map[string]string{
			"version":   gitTag,
			"gitTag":    gitTag,
			"gitCommit": gitCommit,
			"goVersion": runtime.Version(),
			"buildDate": buildDate,
			"os":        runtime.GOOS,
			"arch":      runtime.GOARCH,
			"compiler":  runtime.Compiler,
			"author":    Author,
		}

		version, _ := json.Marshal(versionPayload) // nolint: errcheck
		fmt.Println(string(version))
		os.Exit(0)
	case CommandLint:
		lintMode = true
		fallthrough
	case CommandProcess:
		printAppHeader()
		initSystem()

		if lintMode {
			logger.Info("enabling lint mode, all functions are in dry mode")
		}

		if len(opts.Args.Files) == 0 {
			logger.Error(`no files specified as arguments`)
			os.Exit(1)
		}

		if err := readValuesFiles(); err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		templateFileList := buildSourceTargetList()

		for _, templateFile := range templateFileList {
			if lintMode {
				templateFile.Lint()
			} else {
				templateFile.Apply()
			}
		}

		summary.PostSummary(logger, opts)

		logger.With(slog.String("duration", time.Since(startTime).String())).Info("finished")
	default:
		fmt.Printf("invalid command '%v'\n", opts.Args.Command)
		fmt.Println()
		argparser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
}
