package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

func initLogger(logLevel, devMode string) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("error: failed to set logger level, %w", err)
	}
	return zerolog.New(os.Stdout).Level(level), nil
}
