package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/webdevops/helm-azure-tpl/azuretpl"
)

type (
	TemplateFile struct {
		Context         context.Context
		SourceFile      string
		TargetFile      string
		TemplateBaseDir string
		Logger          *log.Entry
	}
)

func (f *TemplateFile) Lint() {
	var buf strings.Builder
	f.Logger.Infof(`linting file`)
	f.parse(&buf)
	f.Logger.Info(`file successfully linted`)
}

func (f *TemplateFile) Apply() {
	var buf strings.Builder
	f.Logger.Infof(`process file`)
	f.parse(&buf)

	if opts.Debug {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, strings.Repeat("-", TermColumns))
		fmt.Fprintf(os.Stderr, "--- %v\n", f.TargetFile)
		fmt.Fprintln(os.Stderr, strings.Repeat("-", TermColumns))
		fmt.Fprintln(os.Stderr, buf.String())
	}

	if opts.Stdout {
		fmt.Println("--- # src: " + f.SourceFile)
		fmt.Println(buf.String())
		fmt.Println()
		return
	}

	if !opts.DryRun {
		f.write(&buf)
	} else {
		f.Logger.Warn(`not writing file, DRY RUN active`)
	}
}

func (f *TemplateFile) parse(buf *strings.Builder) {
	ctx := f.Context
	contextLogger := f.Logger

	azureTemplate := azuretpl.New(ctx, contextLogger)
	azureTemplate.SetUserAgent(UserAgent + gitTag)
	azureTemplate.SetAzureCliAccountInfo(azAccountInfo)
	azureTemplate.SetLintMode(lintMode)
	azureTemplate.SetTemplateRootPath(f.TemplateBaseDir)
	azureTemplate.SetTemplateRelPath(filepath.Dir(f.SourceFile))
	err := azureTemplate.Parse(f.SourceFile, templateData, buf)
	if err != nil {
		contextLogger.Fatalf(err.Error())
	}
}

func (f *TemplateFile) write(buf *strings.Builder) {
	f.Logger.Infof(`writing file '%v'`, f.TargetFile)
	err := os.WriteFile(f.TargetFile, []byte(buf.String()), 0600)
	if err != nil {
		f.Logger.Fatalf(`unable to write target file '%v': %v`, f.TargetFile, err.Error())
	}
}
