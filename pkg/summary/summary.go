package summary

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/webdevops/go-common/log/slogger"

	"github.com/webdevops/helm-azure-tpl/config"
)

const (
	Header      = "# helm-cloud-tpl summary\n"
	ValueNotSet = "<n/a>"
)

type (
	SummarySection struct {
		Name   string
		Header []string
		Body   []string
		Footer []string
	}
)

var (
	Summary map[string]*SummarySection = map[string]*SummarySection{}
)

func Section(section string) *SummarySection {
	if _, ok := Summary[section]; !ok {
		Summary[section] = &SummarySection{
			Name:   section,
			Header: []string{},
			Body:   []string{},
			Footer: []string{},
		}
	}

	return Summary[section]
}

func (e *SummarySection) SetHeader(val ...string) *SummarySection {
	e.Header = val
	return e
}

func (e *SummarySection) AddHeader(val ...string) *SummarySection {
	e.Header = append(e.Header, val...)
	return e
}

func (e *SummarySection) AddBody(val ...string) *SummarySection {
	e.Body = append(e.Body, val...)
	return e
}

func Build(opts config.Opts) string {
	output := []string{Header}

	output = append(output, "templates:\n")
	for _, file := range opts.Args.Files {
		output = append(output, fmt.Sprintf("- %s", file))
	}

	for _, section := range Summary {
		if len(section.Header) > 0 && len(section.Body) > 0 && len(section.Footer) > 0 {
			output = append(output, fmt.Sprintf("\n### %s\n", section.Name))

			if len(section.Header) > 0 {
				output = append(output, strings.Join(section.Header, "\n"))
			}
			if len(section.Body) > 0 {
				output = append(output, strings.Join(section.Body, "\n"))
			}
			if len(section.Footer) > 0 {
				output = append(output, strings.Join(section.Footer, "\n"))
			}
		}
	}

	return "\n" + strings.Join(output, "\n") + "\n"
}

func PostSummary(logger *slogger.Logger, opts config.Opts) {
	if val := os.Getenv("AZURETPL_EXPERIMENTAL_SUMMARY"); val != "true" && val != "1" {
		return
	}

	// skip empty summary
	if len(Summary) == 0 {
		return
	}

	// github summary
	if summaryPath := os.Getenv("GITHUB_STEP_SUMMARY"); summaryPath != "" {
		content := Build(opts)

		// If the file doesn't exist, create it, or append to the file
		f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gosec // only active when experimental active and GITHUB_STEP_SUMMARY available
		if err != nil {
			logger.Warn(`unable to post GITHUB step summary`, slog.Any("error", err))
			return
		}
		defer f.Close() // nolint: errcheck

		if _, err := f.Write([]byte(content)); err != nil {
			logger.Warn(`unable to post GITHUB step summary`, slog.Any("error", err))
		}
	}
}
