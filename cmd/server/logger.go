package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func initLogger(logLevel string, devMode bool) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("error: failed to set logger level, %w", err)
	}

	if devMode {
		return createDevLogger(level), nil
	}
	return zerolog.New(os.Stdout).Level(level), nil
}

func createDevLogger(level zerolog.Level) zerolog.Logger {
	writerOps := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.TimeOnly,
	}

	l := zerolog.New(writerOps).Level(level).With().Timestamp().Caller().Logger()
	l.Info().Msg("app running in dev mode")
	return l
}
