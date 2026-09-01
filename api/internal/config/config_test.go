package config

import "testing"

func TestLoadJWTSettings(t *testing.T) {
	t.Run("requires an explicit secret and provides identity defaults", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		t.Setenv("JWT_ISSUER", "")
		t.Setenv("JWT_AUDIENCE", "")

		cfg := Load()

		if cfg.JWTSecret != "" {
			t.Errorf("expected no JWT secret fallback, got %q", cfg.JWTSecret)
		}
		if cfg.JWTIssuer != "fleetcontrol" {
			t.Errorf("expected default issuer fleetcontrol, got %q", cfg.JWTIssuer)
		}
		if cfg.JWTAudience != "fleetcontrol-api" {
			t.Errorf("expected default audience fleetcontrol-api, got %q", cfg.JWTAudience)
		}
	})

	t.Run("honors explicit settings", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
		t.Setenv("JWT_ISSUER", "test-issuer")
		t.Setenv("JWT_AUDIENCE", "test-audience")

		cfg := Load()

		if cfg.JWTSecret != "0123456789abcdef0123456789abcdef" {
			t.Errorf("expected configured JWT secret, got %q", cfg.JWTSecret)
		}
		if cfg.JWTIssuer != "test-issuer" {
			t.Errorf("expected configured issuer, got %q", cfg.JWTIssuer)
		}
		if cfg.JWTAudience != "test-audience" {
			t.Errorf("expected configured audience, got %q", cfg.JWTAudience)
		}
	})
}
