package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/auth"
	"github.com/reform-barber/api/internal/handler"
	"github.com/reform-barber/api/internal/middleware"
	"github.com/reform-barber/api/internal/notify"
	"github.com/reform-barber/api/internal/storage"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func initRouter(devMode bool, db *pgxpool.Pool, store storage.Store, notifier notify.Notifier) *chi.Mux {
	logger.Info().Msg("initialising router")

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(chiMiddleware.RequestID)

	// set auth so it can be applied to routes that require it
	jwtSecret := mustEnv("JWT_SECRET")
	authn := middleware.Authenticate(jwtSecret)
	mustBeBarber := middleware.RequireRole(auth.RoleBarber)
	// Admin is a super-user above founder: it inherits every founder
	// (shop-management) route in addition to its own admin-only routes.
	mustBeFounderOrAdmin := middleware.RequireRole(auth.RoleFounder, auth.RoleAdmin)
	mustBeAdmin := middleware.RequireRole(auth.RoleAdmin)

	// create handlers for router
	authH, barbersH, servicesH, productsH, bookingsH, mediaH, adminH := initHandlers(devMode, db, jwtSecret, store, notifier)

	// add routes
	r.Route("/api", func(r chi.Router) {
		// - public routes (ones that require no auth)
		r.Get("/barbers", barbersH.List)
		r.Get("/services", servicesH.List)
		r.Get("/products", productsH.List)
		r.Get("/availability", handler.GetAvailability(db))
		r.Get("/media/carousel", mediaH.ListCarousel)
		r.Get("/media/gallery", mediaH.ListGallery)

		// - auth routes
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/refresh", authH.Refresh)
		r.Post("/auth/logout", authH.Logout)

		// - private routes (ones that users will need authentication to access)
		r.Group(func(r chi.Router) {
			r.Use(authn)
			r.Get("/me", authH.Me)
			r.Post("/bookings", bookingsH.Create)
			r.Get("/me/bookings", bookingsH.ListMine)
			r.Post("/me/bookings/{id}/cancel", bookingsH.Cancel)

			// Barber diary
			r.Group(func(r chi.Router) {
				r.Use(mustBeBarber)
				r.Get("/barber/appointments", bookingsH.BarberAppointments)
			})

			// Founder admin — also reachable by admin, since admin is a
			// super-user that inherits all shop-management capabilities.
			r.Group(func(r chi.Router) {
				r.Use(mustBeFounderOrAdmin)
				r.Get("/founder/bookings", bookingsH.AdminListBookings)

				r.Put("/founder/barbers/{id}", barbersH.Update)
				r.Post("/founder/barbers/{id}/photo", mediaH.UploadBarberPhoto)

				r.Get("/founder/services", servicesH.ListAll)
				r.Post("/founder/services", servicesH.Create)
				r.Put("/founder/services/{id}", servicesH.Update)
				r.Patch("/founder/services/{id}/publish", servicesH.Publish)

				r.Post("/founder/products", productsH.Create)
				r.Put("/founder/products/{id}", productsH.Update)
				r.Post("/founder/products/{id}/image", mediaH.UploadProductImage)

				r.Post("/founder/media/carousel", mediaH.UploadCarousel)
				r.Post("/founder/media/gallery", mediaH.UploadGallery)
				r.Delete("/founder/media/{id}", mediaH.Delete)
				r.Patch("/founder/media/{id}/order", mediaH.UpdateOrder)
			})

			// Admin-only: user/role management and site-wide analytics.
			r.Group(func(r chi.Router) {
				r.Use(mustBeAdmin)
				r.Get("/admin/users", adminH.ListUsers)
				r.Patch("/admin/users/{id}/role", adminH.UpdateUserRole)
				r.Get("/admin/stats", adminH.Stats)
			})
		})
	})

	// handle upload (in local dev)
	if devMode {
		uploadsDir := mustEnv("UPLOADS_DIR")
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))
	}
	return r
}

func initHandlers(devMode bool, db *pgxpool.Pool, jwtSecret string, store storage.Store, notifier notify.Notifier) (
	*handler.AuthHandler, *handler.BarbersHandler, *handler.ServicesHandler, *handler.ProductsHandler, *handler.BookingsHandler, *handler.MediaHandler, *handler.AdminHandler) {
	authH := handler.NewAuthHandler(db, jwtSecret, devMode)
	barbers := handler.NewBarbersHandler(db)
	services := handler.NewServicesHandler(db)
	products := handler.NewProductsHandler(db)
	bookings := handler.NewBookingsHandler(db, notifier)
	media := handler.NewMediaHandler(db, store)
	admin := handler.NewAdminHandler(db)

	return authH, barbers, services, products, bookings, media, admin
}
