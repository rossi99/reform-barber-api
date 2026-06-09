package main

import (
	"fmt"
	"os"

	"github.com/reform-barber/api/internal/storage"
)

func buildStore() storage.Store {
	backend := os.Getenv("STORAGE_BACKEND")
	if backend == "r2" {
		return storage.NewR2Store(
			mustEnv("R2_ACCOUNT_ID"),
			mustEnv("R2_ACCESS_KEY"),
			mustEnv("R2_SECRET_KEY"),
			mustEnv("R2_BUCKET"),
			mustEnv("R2_PUBLIC_URL"),
		)
	}
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	baseURL := fmt.Sprintf("http://localhost:%s/uploads", func() string {
		if p := os.Getenv("API_PORT"); p != "" {
			return p
		}
		return "8080"
	}())
	return storage.NewLocalStore(uploadsDir, baseURL)
}
