package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/config"
	"github.com/mainguyen0112/fleetcontrol/api/internal/db"
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

	r := chi.NewRouter()
	r.Post("/auth/login", authHandler.Login)

	log.Info("server listening", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}
