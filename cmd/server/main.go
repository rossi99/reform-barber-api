package main

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	devMode, err := strconv.ParseBool(mustEnv("DEV_MODE"))
	if err != nil {
		devMode = false // default to dev mode false if parsing fails
	}

	logger, err = initLogger(mustEnv("LOG_LEVEL"), devMode)
	exitOnErr(err)

	pool := initDatabaseConn(ctx)
	defer pool.Close()

	store := buildStore(devMode)

	notifier := buildNotifier()

	// start chi router
	router := initRouter(devMode, pool, store, notifier)

	startSever(router)
}

func startSever(r *chi.Mux) {
	port := mustEnv("API_PORT")
	addr := ":" + port

	logger.Info().Msgf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		exitOnErr(err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		logger.Error().Msgf("required env var %s not set", key)
		os.Exit(1)
	}
	return v
}

func exitOnErr(err error) {
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		os.Exit(1)
	}
}
