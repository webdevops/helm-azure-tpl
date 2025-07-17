package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/webdevops/helm-azure-tpl/azuretpl"
)

type (
	TemplateFile struct {
		Context         context.Context
		SourceFile      string
		TargetFile      string
		TemplateBaseDir string
		Logger          *slog.Logger
	}
)

func (f *TemplateFile) Lint() {
	var buf strings.Builder
	f.Logger.Info(`linting file`)
	f.parse(&buf)
	f.Logger.Info(`file successfully linted`)
}

func (f *TemplateFile) Apply() {
	var buf strings.Builder
	f.Logger.Info(`process file`)
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

	azureTemplate := azuretpl.New(ctx, opts.AzureTpl, contextLogger)
	azureTemplate.SetUserAgent(UserAgent + gitTag)
	azureTemplate.SetLintMode(lintMode)
	azureTemplate.SetTemplateRootPath(f.TemplateBaseDir)
	azureTemplate.SetTemplateRelPath(filepath.Dir(f.SourceFile))
	err := azureTemplate.Parse(f.SourceFile, templateData, buf)
	if err != nil {
		f.Logger.Error(err.Error())
		os.Exit(1)
	}
}

func (f *TemplateFile) write(buf *strings.Builder) {
	f.Logger.Info(`writing file`, slog.String("path", f.TargetFile))
	err := os.WriteFile(f.TargetFile, []byte(buf.String()), 0600)
	if err != nil {
		f.Logger.Error(`unable to write target file`, slog.String("path", f.TargetFile), slog.Any("error", err.Error()))
		os.Exit(1)
	}
}
