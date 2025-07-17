package main

import (
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
	"gitlab.com/greyxor/slogor"
)

var (
	logger *slog.Logger
)

func initLogger() *slog.Logger {
	if opts.Logger.Json {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		if opts.Logger.Debug {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		} else {
			slog.SetLogLoggerLevel(slog.LevelInfo)
		}
	} else {
		logOpts := []slogor.OptionFn{}

		if opts.Logger.Debug {
			logOpts = append(logOpts, slogor.SetLevel(slog.LevelDebug))
			logOpts = append(logOpts, slogor.ShowSource())
		} else {
			logOpts = append(logOpts, slogor.SetLevel(slog.LevelInfo))
		}

		if !isatty.IsTerminal(os.Stderr.Fd()) {
			logOpts = append(logOpts, slogor.DisableColor())
		}

		logger = slog.New(slogor.NewHandler(os.Stderr, logOpts...))
	}

	slog.SetDefault(logger)

	logger.Debug("foobar")

	return logger
}
