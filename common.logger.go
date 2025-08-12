package main

import (
	"os"

	"github.com/webdevops/go-common/log/slogger"
)

var (
	logger *slogger.Logger
)

func initLogger() *slogger.Logger {
	loggerOpts := []slogger.LoggerOptionFunc{
		slogger.WithLevelText(opts.Logger.Level),
		slogger.WithFormat(slogger.FormatMode(opts.Logger.Format)),
		slogger.WithSourceMode(slogger.SourceMode(opts.Logger.Source)),
		slogger.WithTime(opts.Logger.Time),
		slogger.WithColor(slogger.ColorMode(opts.Logger.Color)),
	}

	logger = slogger.NewCliLogger(
		os.Stderr, loggerOpts...,
	)

	return logger
}
