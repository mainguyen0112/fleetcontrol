package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testJWTSecret   = "0123456789abcdef0123456789abcdef"
	testJWTIssuer   = "fleetcontrol"
	testJWTAudience = "fleetcontrol-api"
)

var testJWTNow = time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)

func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name    string
		config  JWTConfig
		wantErr bool
	}{
		{
			name: "valid",
			config: JWTConfig{
				Secret: testJWTSecret, Issuer: testJWTIssuer, Audience: testJWTAudience,
			},
		},
		{
			name: "missing secret",
			config: JWTConfig{
				Issuer: testJWTIssuer, Audience: testJWTAudience,
			},
			wantErr: true,
		},
		{
			name: "short secret",
			config: JWTConfig{
				Secret: strings.Repeat("x", minimumJWTSecretBytes-1), Issuer: testJWTIssuer, Audience: testJWTAudience,
			},
			wantErr: true,
		},
		{
			name: "whitespace secret",
			config: JWTConfig{
				Secret: strings.Repeat(" ", minimumJWTSecretBytes), Issuer: testJWTIssuer, Audience: testJWTAudience,
			},
			wantErr: true,
		},
		{
			name: "missing issuer",
			config: JWTConfig{
				Secret: testJWTSecret, Audience: testJWTAudience,
			},
			wantErr: true,
		},
		{
			name: "missing audience",
			config: JWTConfig{
				Secret: testJWTSecret, Issuer: testJWTIssuer,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewJWTManager(tt.config)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidJWTConfig) {
					t.Fatalf("expected ErrInvalidJWTConfig, got %v", err)
				}
				if manager != nil {
					t.Fatal("expected no manager for invalid config")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewJWTManager returned an error: %v", err)
			}
			if manager == nil {
				t.Fatal("expected a JWT manager")
			}
		})
	}
}

func TestJWTManager_GenerateAndParseHumanToken(t *testing.T) {
	manager := newTestJWTManager(t)
	tokenString, err := manager.GenerateHumanToken("user-1", RoleAdmin)
	if err != nil {
		t.Fatalf("GenerateHumanToken returned an error: %v", err)
	}

	principal, err := manager.ParseHumanToken(tokenString)
	if err != nil {
		t.Fatalf("ParseHumanToken returned an error: %v", err)
	}
	if principal.Kind() != ActorHuman {
		t.Errorf("expected human actor, got %q", principal.Kind())
	}
	if principal.Subject() != "user-1" {
		t.Errorf("expected subject user-1, got %q", principal.Subject())
	}
	role, ok := principal.HumanRole()
	if !ok || role != RoleAdmin {
		t.Errorf("expected admin role, got %q (human: %t)", role, ok)
	}

	claims := &Claims{}
	parsed, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	if err != nil {
		t.Fatalf("failed to inspect generated token: %v", err)
	}
	if parsed.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		t.Errorf("expected HS256, got %s", parsed.Method.Alg())
	}
	if claims.Issuer != testJWTIssuer {
		t.Errorf("expected issuer %q, got %q", testJWTIssuer, claims.Issuer)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != testJWTAudience {
		t.Errorf("expected audience %q, got %v", testJWTAudience, claims.Audience)
	}
	if claims.Subject != "user-1" {
		t.Errorf("expected registered subject user-1, got %q", claims.Subject)
	}
	if claims.IssuedAt == nil || !claims.IssuedAt.Time.Equal(testJWTNow) {
		t.Errorf("expected issued-at %v, got %v", testJWTNow, claims.IssuedAt)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(testJWTNow.Add(HumanTokenTTL)) {
		t.Errorf("expected expiry %v, got %v", testJWTNow.Add(HumanTokenTTL), claims.ExpiresAt)
	}
	if claims.ActorKind != ActorHuman {
		t.Errorf("expected actor kind human, got %q", claims.ActorKind)
	}
	if claims.Role != RoleAdmin {
		t.Errorf("expected role admin, got %q", claims.Role)
	}

	mapClaims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(tokenString, mapClaims); err != nil {
		t.Fatalf("failed to inspect raw token claims: %v", err)
	}
	if _, exists := mapClaims["user_id"]; exists {
		t.Error("generated token must not contain the legacy user_id claim")
	}
}

func TestJWTManager_GenerateHumanTokenRejectsInvalidPrincipal(t *testing.T) {
	manager := newTestJWTManager(t)
	tests := []struct {
		name    string
		subject string
		role    HumanRole
	}{
		{name: "missing subject", subject: "", role: RoleAdmin},
		{name: "invalid role", subject: "user-1", role: HumanRole("owner")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := manager.GenerateHumanToken(tt.subject, tt.role); !errors.Is(err, ErrInvalidPrincipal) {
				t.Fatalf("expected ErrInvalidPrincipal, got %v", err)
			}
		})
	}
}

func TestJWTManager_ParseHumanTokenRejectsInvalidContract(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Claims)
		method jwt.SigningMethod
		secret string
	}{
		{name: "wrong algorithm", method: jwt.SigningMethodHS384},
		{name: "wrong signature", secret: strings.Repeat("z", minimumJWTSecretBytes)},
		{name: "expired", mutate: func(c *Claims) { c.ExpiresAt = jwt.NewNumericDate(testJWTNow.Add(-time.Minute)) }},
		{name: "future issued-at", mutate: func(c *Claims) { c.IssuedAt = jwt.NewNumericDate(testJWTNow.Add(time.Minute)) }},
		{name: "missing issuer", mutate: func(c *Claims) { c.Issuer = "" }},
		{name: "wrong issuer", mutate: func(c *Claims) { c.Issuer = "other" }},
		{name: "missing audience", mutate: func(c *Claims) { c.Audience = nil }},
		{name: "wrong audience", mutate: func(c *Claims) { c.Audience = jwt.ClaimStrings{"other"} }},
		{name: "missing subject", mutate: func(c *Claims) { c.Subject = "" }},
		{name: "missing issued-at", mutate: func(c *Claims) { c.IssuedAt = nil }},
		{name: "missing expiry", mutate: func(c *Claims) { c.ExpiresAt = nil }},
		{name: "missing actor kind", mutate: func(c *Claims) { c.ActorKind = "" }},
		{name: "wrong actor kind", mutate: func(c *Claims) { c.ActorKind = ActorOperator }},
		{name: "missing role", mutate: func(c *Claims) { c.Role = "" }},
		{name: "invalid role", mutate: func(c *Claims) { c.Role = HumanRole("owner") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := validTestClaims()
			mutate := tt.mutate
			if mutate != nil {
				mutate(&claims)
			}
			method := tt.method
			if method == nil {
				method = jwt.SigningMethodHS256
			}
			secret := tt.secret
			if secret == "" {
				secret = testJWTSecret
			}
			tokenString := signTestClaims(t, claims, method, secret)

			if _, err := newTestJWTManager(t).ParseHumanToken(tokenString); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func TestJWTManager_ParseHumanTokenRejectsMalformedToken(t *testing.T) {
	if _, err := newTestJWTManager(t).ParseHumanToken("not-a-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func newTestJWTManager(t *testing.T) *JWTManager {
	t.Helper()
	manager, err := NewJWTManager(JWTConfig{
		Secret:   testJWTSecret,
		Issuer:   testJWTIssuer,
		Audience: testJWTAudience,
	})
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}
	manager.now = func() time.Time { return testJWTNow }
	return manager
}

func validTestClaims() Claims {
	return Claims{
		ActorKind: ActorHuman,
		Role:      RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testJWTIssuer,
			Subject:   "user-1",
			Audience:  jwt.ClaimStrings{testJWTAudience},
			ExpiresAt: jwt.NewNumericDate(testJWTNow.Add(HumanTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(testJWTNow),
		},
	}
}

func signTestHumanClaims(t *testing.T, claims Claims) string {
	t.Helper()
	return signTestClaims(t, claims, jwt.SigningMethodHS256, testJWTSecret)
}

func signTestClaims(t *testing.T, claims Claims, method jwt.SigningMethod, secret string) string {
	t.Helper()
	tokenString, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return tokenString
}
