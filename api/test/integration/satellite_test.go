package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/config"
	"github.com/mainguyen0112/fleetcontrol/api/internal/db"
	"github.com/mainguyen0112/fleetcontrol/api/internal/satellite"
	"go.uber.org/zap"
)

var testRouter *chi.Mux
var testToken string

func TestMain(m *testing.M) {
	cfg := config.Load()

	log, _ := zap.NewDevelopment()

	pool, err := db.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}

	authHandler := &auth.Handler{DB: pool, Secret: cfg.JWTSecret}
	satRepo := satellite.NewPostgresRepository(pool)
	satService := satellite.NewService(satRepo)
	satHandler := satellite.NewHandler(satService)

	r := chi.NewRouter()
	r.Post("/auth/login", authHandler.Login)
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Post("/satellites", satHandler.Create)
		r.Get("/satellites", satHandler.List)
		r.Get("/satellites/{id}", satHandler.GetByID)
		r.Patch("/satellites/{id}", satHandler.Update)
		r.Delete("/satellites/{id}", satHandler.Delete)
		r.Post("/satellites/{id}/heartbeat", satHandler.Heartbeat)
	})

	testRouter = r
	os.Exit(m.Run())
}

func TestSatelliteCRUD(t *testing.T) {
	// Step 1: login
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}

	var loginResp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&loginResp)
	token := loginResp["token"].(string)

	// Step 2: create satellite
	body, _ = json.Marshal(map[string]string{"name": "test-edge", "region": "test-region"})
	req = httptest.NewRequest(http.MethodPost, "/satellites", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create satellite failed: %d %s", rec.Code, rec.Body.String())
	}

	var sat map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&sat)
	satID := sat["id"].(string)

	// Step 3: heartbeat
	req = httptest.NewRequest(http.MethodPost, "/satellites/"+satID+"/heartbeat", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("heartbeat failed: %d", rec.Code)
	}

	// Step 4: get by id
	req = httptest.NewRequest(http.MethodGet, "/satellites/"+satID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("get satellite failed: %d", rec.Code)
	}

	// Step 5: update
	body, _ = json.Marshal(map[string]string{"region": "updated-region"})
	req = httptest.NewRequest(http.MethodPatch, "/satellites/"+satID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("update satellite failed: %d", rec.Code)
	}

	// Step 6: delete
	req = httptest.NewRequest(http.MethodDelete, "/satellites/"+satID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete satellite failed: %d", rec.Code)
	}
}

func TestSatelliteList_NoToken_Returns401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/satellites", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
