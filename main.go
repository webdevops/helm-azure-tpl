package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/webdevops/helm-azure-tpl/config"
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
