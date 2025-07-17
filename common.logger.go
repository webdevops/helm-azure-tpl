package main

import (
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
	"gitlab.com/greyxor/slogor"
)

var (
	logger      *slog.Logger
	loggerLevel = new(slog.LevelVar)
)

func initLogger() *slog.Logger {
	if opts.Logger.Json {
		ReplaceAttr := func(group []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}
			return slog.Attr{Key: a.Key, Value: a.Value}
		}

		loggerOpts := slog.HandlerOptions{
			AddSource:   true,
			Level:       loggerLevel,
			ReplaceAttr: ReplaceAttr,
		}
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &loggerOpts))
	} else {
		logOpts := []slogor.OptionFn{
			slogor.SetLevel(loggerLevel),
		}

		if opts.Logger.Debug {
			logOpts = append(logOpts, slogor.ShowSource())
		}

		if !isatty.IsTerminal(os.Stderr.Fd()) {
			logOpts = append(logOpts, slogor.DisableColor())
		}

		logger = slog.New(slogor.NewHandler(os.Stderr, logOpts...))
	}

	if opts.Logger.Debug {
		loggerLevel.Set(slog.LevelDebug)
	}

	slog.SetDefault(logger)

	return logger
}
