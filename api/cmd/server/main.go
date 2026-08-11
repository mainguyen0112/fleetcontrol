package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/config"
	"github.com/mainguyen0112/fleetcontrol/api/internal/db"
	"github.com/mainguyen0112/fleetcontrol/api/internal/health"
	"github.com/mainguyen0112/fleetcontrol/api/internal/satellite"
	"github.com/mainguyen0112/fleetcontrol/api/internal/user"
	"github.com/mainguyen0112/fleetcontrol/api/pkg/logger"
	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

func main() {
	cfg := config.Load()

	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	pool, err := db.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}
	defer pool.Close()

	authHandler := &auth.Handler{DB: pool, Secret: cfg.JWTSecret}

	satRepo := satellite.NewPostgresRepository(pool)
	satService := satellite.NewService(satRepo)
	satHandler := satellite.NewHandler(satService)

	userRepo := user.NewPostgresRepository(pool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	healthHandler := &health.Handler{DB: pool}

	server := NewServer(satHandler, userHandler, authHandler, healthHandler)
	wrapper := &gen.ServerInterfaceWrapper{Handler: server}

	r := chi.NewRouter()
	r.Use(logger.RequestLogger(log))

	r.Post("/auth/login", wrapper.PostAuthLogin)
	r.Get("/health", wrapper.GetHealth)
	r.Get("/version", wrapper.GetVersion)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Post("/satellites", wrapper.PostSatellites)
		r.Get("/satellites", wrapper.GetSatellites)
		r.Get("/satellites/{id}", wrapper.GetSatellitesId)
		r.Patch("/satellites/{id}", wrapper.PatchSatellitesId)
		r.Delete("/satellites/{id}", wrapper.DeleteSatellitesId)
		r.Post("/satellites/{id}/heartbeat", wrapper.PostSatellitesIdHeartbeat)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Use(auth.RequireRole("admin"))
		r.Post("/users", wrapper.PostUsers)
		r.Get("/users", wrapper.GetUsers)
		r.Delete("/users/{id}", wrapper.DeleteUsersId)
	})

	log.Info("server listening", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}