package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/config"
	"github.com/mainguyen0112/fleetcontrol/api/internal/db"
	"github.com/mainguyen0112/fleetcontrol/api/internal/docs"
	"github.com/mainguyen0112/fleetcontrol/api/internal/health"
	"github.com/mainguyen0112/fleetcontrol/api/internal/httpserver"
	"github.com/mainguyen0112/fleetcontrol/api/internal/satellite"
	"github.com/mainguyen0112/fleetcontrol/api/internal/user"
)

// This test exercises the FULL Phase 3 pipeline:
//
//	OpenAPI spec -> generated ClientWithResponses -> HTTP -> gen.ServerInterface
//	(via ServerInterfaceWrapper) -> Server adapter -> domain Handler -> Service
//	-> Repository -> PostgreSQL
//
// It intentionally builds the router the same way cmd/server/main.go does,
// so a bug in the Server adapter or route wiring fails this test instead of
// only showing up at runtime.

var testServerURL string
var apiClient *gen.ClientWithResponses

func TestMain(m *testing.M) {
	cfg := config.Load()
	log, _ := zap.NewDevelopment()

	pool, err := db.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}

	tokens, err := auth.NewJWTManager(auth.JWTConfig{
		Secret:   cfg.JWTSecret,
		Issuer:   cfg.JWTIssuer,
		Audience: cfg.JWTAudience,
	})
	if err != nil {
		log.Fatal("invalid JWT configuration", zap.Error(err))
	}

	authHandler := &auth.Handler{DB: pool, Tokens: tokens}

	satRepo := satellite.NewPostgresRepository(pool)
	satService := satellite.NewService(satRepo)
	satHandler := satellite.NewHandler(satService)

	userRepo := user.NewPostgresRepository(pool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	healthHandler := &health.Handler{DB: pool}

	server := httpserver.NewServer(satHandler, userHandler, authHandler, healthHandler)
	r := httpserver.NewRouter(server, docs.NewHandler(), tokens, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()
	testServerURL = ts.URL

	client, err := gen.NewClientWithResponses(testServerURL)
	if err != nil {
		log.Fatal("failed to create generated client", zap.Error(err))
	}
	apiClient = client

	os.Exit(m.Run())
}

func TestSatelliteCRUD_ThroughGeneratedClient(t *testing.T) {
	ctx := context.Background()

	// Step 1: login
	loginResp, err := apiClient.PostAuthLoginWithResponse(ctx, gen.PostAuthLoginJSONRequestBody{
		Username: "admin",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if loginResp.StatusCode() != http.StatusOK || loginResp.JSON200 == nil || loginResp.JSON200.Token == nil {
		t.Fatalf("login failed: status=%d body=%s", loginResp.StatusCode(), string(loginResp.Body))
	}
	token := *loginResp.JSON200.Token
	authEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	// Step 2: create
	createResp, err := apiClient.PostSatellitesWithResponse(ctx, gen.PostSatellitesJSONRequestBody{
		Name:   "test-edge",
		Region: "test-region",
	}, authEditor)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	if createResp.StatusCode() != http.StatusCreated || createResp.JSON201 == nil {
		t.Fatalf("create satellite failed: status=%d body=%s", createResp.StatusCode(), string(createResp.Body))
	}
	if createResp.JSON201.Id == nil {
		t.Fatalf("created satellite has no id")
	}
	satID := *createResp.JSON201.Id

	// Step 3: heartbeat
	hbResp, err := apiClient.PostSatellitesIdHeartbeatWithResponse(ctx, satID, authEditor)
	if err != nil {
		t.Fatalf("heartbeat request failed: %v", err)
	}
	if hbResp.StatusCode() != http.StatusOK || hbResp.JSON200 == nil {
		t.Fatalf("heartbeat failed: status=%d body=%s", hbResp.StatusCode(), string(hbResp.Body))
	}

	// Step 4: get by id
	getResp, err := apiClient.GetSatellitesIdWithResponse(ctx, satID, authEditor)
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	if getResp.StatusCode() != http.StatusOK || getResp.JSON200 == nil {
		t.Fatalf("get satellite failed: status=%d body=%s", getResp.StatusCode(), string(getResp.Body))
	}

	// Step 5: update
	updateResp, err := apiClient.PatchSatellitesIdWithResponse(ctx, satID, gen.PatchSatellitesIdJSONRequestBody{
		Region: "updated-region",
	}, authEditor)
	if err != nil {
		t.Fatalf("update request failed: %v", err)
	}
	if updateResp.StatusCode() != http.StatusOK || updateResp.JSON200 == nil {
		t.Fatalf("update satellite failed: status=%d body=%s", updateResp.StatusCode(), string(updateResp.Body))
	}
	if updateResp.JSON200.Region == nil || *updateResp.JSON200.Region != "updated-region" {
		t.Errorf("expected region to be updated-region, got %+v", updateResp.JSON200.Region)
	}

	// Step 6: delete
	deleteResp, err := apiClient.DeleteSatellitesIdWithResponse(ctx, satID, authEditor)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	if deleteResp.StatusCode() != http.StatusNoContent {
		t.Fatalf("delete satellite failed: status=%d body=%s", deleteResp.StatusCode(), string(deleteResp.Body))
	}
}

func TestSatelliteList_NoToken_Returns401(t *testing.T) {
	resp, err := apiClient.GetSatellitesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode() != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode())
	}
}

// sanity check that raw JSON bytes are also valid JSON, catching cases
// where JSON200/JSON201 might be nil due to a content-type mismatch that
// the strict typed field wouldn't otherwise surface clearly.
func TestSatelliteCreate_ResponseBodyIsValidJSON(t *testing.T) {
	ctx := context.Background()
	loginResp, _ := apiClient.PostAuthLoginWithResponse(ctx, gen.PostAuthLoginJSONRequestBody{
		Username: "admin", Password: "admin123",
	})
	token := *loginResp.JSON200.Token
	authEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	resp, err := apiClient.PostSatellitesWithResponse(ctx, gen.PostSatellitesJSONRequestBody{
		Name: "json-check", Region: "r1",
	}, authEditor)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		t.Fatalf("response body is not valid JSON: %v\nbody: %s", err, string(resp.Body))
	}
	if _, ok := raw["id"]; !ok {
		t.Errorf("response JSON missing 'id' field: %s", string(resp.Body))
	}

	// cleanup
	_, _ = apiClient.DeleteSatellitesIdWithResponse(ctx, *resp.JSON201.Id, authEditor)
}
