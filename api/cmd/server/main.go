package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/config"
	"github.com/mainguyen0112/fleetcontrol/api/internal/db"
	"github.com/mainguyen0112/fleetcontrol/api/internal/docs"
	"github.com/mainguyen0112/fleetcontrol/api/internal/health"
	"github.com/mainguyen0112/fleetcontrol/api/internal/httpserver"
	"github.com/mainguyen0112/fleetcontrol/api/internal/satellite"
	"github.com/mainguyen0112/fleetcontrol/api/internal/user"
	"github.com/mainguyen0112/fleetcontrol/api/pkg/logger"
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
	docsHandler := docs.NewHandler()

	server := httpserver.NewServer(satHandler, userHandler, authHandler, healthHandler)
	r := httpserver.NewRouter(server, docsHandler, cfg.JWTSecret, logger.RequestLogger(log))

	log.Info("server listening", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}
