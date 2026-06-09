package main

import (
	"fmt"

	"github.com/reform-barber/api/internal/storage"
)

func buildStore(devMode bool) storage.Store {
	if devMode {
		// init local store when in dev mode
		logger.Info().Msg("initiating local storage")
		baseURL := fmt.Sprintf("http://localhost:%s/uploads", mustEnv("API_PORT"))
		return storage.NewLocalStore(mustEnv("UPLOADS_DIR"), baseURL)
	}

	// init Cloudflare R2 store
	logger.Info().Msg("initiating cloudflare R2 storage")
	return storage.NewR2Store(mustEnv("R2_ACCOUNT_ID"), mustEnv("R2_ACCESS_KEY"), mustEnv("R2_SECRET_KEY"), mustEnv("R2_BUCKET"), mustEnv("R2_PUBLIC_URL"))
}
