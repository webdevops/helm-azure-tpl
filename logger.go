package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
)

func initLogger() *zap.SugaredLogger {
	var config zap.Config
	if opts.Logger.Development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
	}

	config.Encoding = "console"
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.CallerKey = ""
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// running as build tool, should not log times (only in debug mode)
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// debug level
	if opts.Debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		config.EncoderConfig.TimeKey = "ts"
		config.EncoderConfig.CallerKey = "caller"
		config.DisableStacktrace = false
	}

	// json log format
	if opts.Logger.Json {
		config.Encoding = "json"

		// if running in containers, logs already enriched with timestamp by the container runtime
		config.EncoderConfig.TimeKey = ""
	}

	switch {
	case os.Getenv("SYSTEM_TEAMFOUNDATIONSERVERURI") != "":
		// Azure DevOps
		fallthrough
	case os.Getenv("GITLAB_CI") != "":
		// GitLab
		fallthrough
	case os.Getenv("JENKINS_URL") != "":
		// Jenkins
		fallthrough
	case os.Getenv("GITHUB_ACTION") != "":
		// GitHub
		config.EncoderConfig.TimeKey = ""
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// build logger
	log, err := config.Build()
	if err != nil {
		panic(err)
	}
	logger = log.Sugar()
	return logger
}
