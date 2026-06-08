package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/reform-barber/api/internal/handler"
	"github.com/reform-barber/api/internal/middleware"
	"github.com/reform-barber/api/internal/notify"
	"github.com/reform-barber/api/internal/storage"
	"github.com/rs/zerolog"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

var (
	err    error
	logger zerolog.Logger
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	logger, err = initLogger(mustEnv("LOG_LEVEL"), mustEnv("DEV_MODE"))
	exitOnErr(err)

	pool := initDatabaseConn(ctx)
	defer pool.Close()

	migrateDatabase(pool)

	store := buildStore()

	notifier := buildNotifier()

	// Handlers
	jwtSecret := mustEnv("JWT_SECRET")
	authH := handler.NewAuthHandler(pool, jwtSecret)
	barbersH := handler.NewBarbersHandler(pool)
	servicesH := handler.NewServicesHandler(pool)
	productsH := handler.NewProductsHandler(pool)
	bookingsH := handler.NewBookingsHandler(pool, notifier)
	mediaH := handler.NewMediaHandler(pool, store)

	authn := middleware.Authenticate(jwtSecret)
	requireBarber := middleware.RequireRole("barber", "founder")
	requireFounder := middleware.RequireRole("founder")

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RequestID)

	// Serve local uploads in dev
	if os.Getenv("STORAGE_BACKEND") == "local" || os.Getenv("STORAGE_BACKEND") == "" {
		uploadsDir := os.Getenv("UPLOADS_DIR")
		if uploadsDir == "" {
			uploadsDir = "./uploads"
		}
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))
	}

	r.Route("/api", func(r chi.Router) {
		// Public
		r.Get("/barbers", barbersH.List)
		r.Get("/services", servicesH.List)
		r.Get("/products", productsH.List)
		r.Get("/availability", handler.GetAvailability(pool))
		r.Get("/media/carousel", mediaH.ListCarousel)
		r.Get("/media/gallery", mediaH.ListGallery)

		// Auth
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/refresh", authH.Refresh)
		r.Post("/auth/logout", authH.Logout)

		// Authenticated
		r.Group(func(r chi.Router) {
			r.Use(authn)
			r.Get("/me", authH.Me)
			r.Post("/bookings", bookingsH.Create)
			r.Get("/me/bookings", bookingsH.ListMine)
			r.Post("/me/bookings/{id}/cancel", bookingsH.Cancel)

			// Barber diary
			r.Group(func(r chi.Router) {
				r.Use(requireBarber)
				r.Get("/barber/appointments", bookingsH.BarberAppointments)
			})

			// Founder admin
			r.Group(func(r chi.Router) {
				r.Use(requireFounder)
				r.Get("/admin/bookings", bookingsH.AdminListBookings)

				r.Put("/admin/barbers/{id}", barbersH.Update)
				r.Post("/admin/barbers/{id}/photo", mediaH.UploadBarberPhoto)

				r.Post("/admin/services", servicesH.Create)
				r.Put("/admin/services/{id}", servicesH.Update)

				r.Post("/admin/products", productsH.Create)
				r.Put("/admin/products/{id}", productsH.Update)
				r.Post("/admin/products/{id}/image", mediaH.UploadProductImage)

				r.Post("/admin/media/carousel", mediaH.UploadCarousel)
				r.Post("/admin/media/gallery", mediaH.UploadGallery)
				r.Delete("/admin/media/{id}", mediaH.Delete)
				r.Patch("/admin/media/{id}/order", mediaH.UpdateOrder)
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func exitOnErr(err error) {
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		os.Exit(1)
	}
	return
}

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
		if p := os.Getenv("PORT"); p != "" {
			return p
		}
		return "8080"
	}())
	return storage.NewLocalStore(uploadsDir, baseURL)
}

func buildNotifier() notify.Notifier {
	var notifiers []notify.Notifier

	if key := os.Getenv("RESEND_API_KEY"); key != "" {
		from := os.Getenv("EMAIL_FROM")
		if from == "" {
			from = "bookings@reformbarber.co.uk"
		}
		notifiers = append(notifiers, notify.NewEmailNotifier(key, from, "RE:FORM"))
	}

	if sid := os.Getenv("TWILIO_ACCOUNT_SID"); sid != "" {
		notifiers = append(notifiers, notify.NewSMSNotifier(
			sid,
			mustEnv("TWILIO_AUTH_TOKEN"),
			mustEnv("TWILIO_FROM_NUMBER"),
		))
	}

	if len(notifiers) == 0 {
		slog.Warn("no notification providers configured — using noop")
		return notify.Noop{}
	}
	return notify.NewMulti(notifiers...)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required env var not set", "key", key)
		os.Exit(1)
	}
	return v
}
