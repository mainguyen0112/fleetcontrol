package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware_NoToken_Returns401(t *testing.T) {
	handler := Middleware("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMiddleware_ValidToken_StoresHumanPrincipal(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateToken(secret, "user-1", "admin", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := Middleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("expected an authenticated principal in the request context")
		}
		if principal.Kind() != ActorHuman {
			t.Errorf("expected human principal, got %q", principal.Kind())
		}
		if principal.Subject() != "user-1" {
			t.Errorf("expected subject user-1, got %q", principal.Subject())
		}
		role, ok := principal.HumanRole()
		if !ok || role != RoleAdmin {
			t.Errorf("expected admin human role, got %q (human: %t)", role, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_InvalidHumanClaims_Returns401(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		role   string
	}{
		{name: "missing subject", userID: "", role: string(RoleAdmin)},
		{name: "invalid role", userID: "user-1", role: "owner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const secret = "test-secret"
			token, err := GenerateToken(secret, tt.userID, tt.role, time.Hour)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			nextCalled := false
			handler := Middleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
			if nextCalled {
				t.Error("expected invalid human claims to stop the middleware chain")
			}
		})
	}
}

func TestPrincipalFromContext(t *testing.T) {
	principal, err := NewHumanPrincipal("user-1", RoleViewer)
	if err != nil {
		t.Fatalf("failed to create principal: %v", err)
	}

	tests := []struct {
		name   string
		ctx    context.Context
		wantOK bool
	}{
		{name: "missing", ctx: context.Background(), wantOK: false},
		{name: "wrong value type", ctx: context.WithValue(context.Background(), principalContextKey, "raw claims"), wantOK: false},
		{name: "invalid", ctx: withPrincipal(context.Background(), Principal{}), wantOK: false},
		{name: "valid", ctx: withPrincipal(context.Background(), principal), wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := PrincipalFromContext(tt.ctx)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%t, got %t", tt.wantOK, ok)
			}
			if tt.wantOK && got.Subject() != principal.Subject() {
				t.Errorf("expected subject %q, got %q", principal.Subject(), got.Subject())
			}
		})
	}
}

func TestRequireHumanRole(t *testing.T) {
	admin, err := NewHumanPrincipal("admin-1", RoleAdmin)
	if err != nil {
		t.Fatalf("failed to create admin principal: %v", err)
	}
	viewer, err := NewHumanPrincipal("viewer-1", RoleViewer)
	if err != nil {
		t.Fatalf("failed to create viewer principal: %v", err)
	}
	operator, err := NewOperatorPrincipal("operator-1")
	if err != nil {
		t.Fatalf("failed to create operator principal: %v", err)
	}

	tests := []struct {
		name       string
		principal  *Principal
		required   HumanRole
		wantStatus int
		wantNext   bool
	}{
		{name: "admin allowed", principal: &admin, required: RoleAdmin, wantStatus: http.StatusOK, wantNext: true},
		{name: "viewer denied", principal: &viewer, required: RoleAdmin, wantStatus: http.StatusForbidden},
		{name: "machine actor denied", principal: &operator, required: RoleAdmin, wantStatus: http.StatusForbidden},
		{name: "missing principal denied", principal: nil, required: RoleAdmin, wantStatus: http.StatusForbidden},
		{name: "unknown required role denied", principal: &admin, required: HumanRole("owner"), wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			handler := RequireHumanRole(tt.required)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.principal != nil {
				req = req.WithContext(withPrincipal(req.Context(), *tt.principal))
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
			if nextCalled != tt.wantNext {
				t.Errorf("expected next called=%t, got %t", tt.wantNext, nextCalled)
			}
		})
	}
}
