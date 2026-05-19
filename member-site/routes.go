package main

import (
	"net/http"

	"github.com/atlantacoven/coven-platform/member-site/api"
	"github.com/atlantacoven/coven-platform/member-site/database"
	"github.com/atlantacoven/coven-platform/member-site/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouteBuilder func(r chi.Router)

var routers = []RouteBuilder{
	// ADD ROUTES HERE
	users.Router,
}

func NewServer(db database.DB) http.Handler {
	r := chi.NewRouter()

	// attach db to context
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := database.WithDB(db, r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Use(middleware.Logger)

	allowedOrigins := []string{"https://*", "http://*"}
	if api.IsEnv(api.Production) {
		allowedOrigins = []string{"https://api.thecoven.space"}
	}
  r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   allowedOrigins,
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: false,
    MaxAge:           300,
  }))

	// health check
	r.Get("/", HealthCheck)
	// routes from other packages
	for _, rb := range routers {
		rb(r)
	}
	return r
}
