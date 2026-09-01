package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	minimumJWTSecretBytes = 32

	// HumanTokenTTL is the lifetime returned by the login endpoint for human
	// access tokens.
	HumanTokenTTL = time.Hour
)

var (
	// ErrInvalidJWTConfig is returned when JWT signing or validation settings
	// do not meet FleetControl's minimum security requirements.
	ErrInvalidJWTConfig = errors.New("invalid JWT configuration")

	// ErrInvalidToken is returned when a bearer token cannot be authenticated
	// as a FleetControl human principal.
	ErrInvalidToken = errors.New("invalid human token")
)

// JWTConfig defines the shared contract used to issue and validate human JWTs.
type JWTConfig struct {
	Secret   string
	Issuer   string
	Audience string
}

// Claims is the wire representation of a FleetControl human JWT.
type Claims struct {
	ActorKind ActorKind `json:"actor_kind"`
	Role      HumanRole `json:"role"`
	jwt.RegisteredClaims
}

// Validate enforces FleetControl-specific claims after the standard JWT
// registered claims have been checked by golang-jwt.
func (c Claims) Validate() error {
	if c.IssuedAt == nil {
		return errors.New("issued-at claim is required")
	}
	if c.ActorKind != ActorHuman {
		return errors.New("actor kind must be human")
	}
	if _, err := NewHumanPrincipal(c.Subject, c.Role); err != nil {
		return err
	}

	return nil
}

// JWTManager issues and authenticates human JWTs using one validated config.
type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
	now      func() time.Time
}

// NewJWTManager validates JWT configuration before any token can be issued or
// accepted. The secret must contain at least 32 bytes.
func NewJWTManager(config JWTConfig) (*JWTManager, error) {
	if len([]byte(config.Secret)) < minimumJWTSecretBytes || strings.TrimSpace(config.Secret) == "" {
		return nil, fmt.Errorf("%w: secret must contain at least %d bytes", ErrInvalidJWTConfig, minimumJWTSecretBytes)
	}

	issuer := strings.TrimSpace(config.Issuer)
	if issuer == "" {
		return nil, fmt.Errorf("%w: issuer is required", ErrInvalidJWTConfig)
	}

	audience := strings.TrimSpace(config.Audience)
	if audience == "" {
		return nil, fmt.Errorf("%w: audience is required", ErrInvalidJWTConfig)
	}

	return &JWTManager{
		secret:   append([]byte(nil), []byte(config.Secret)...),
		issuer:   issuer,
		audience: audience,
		now:      time.Now,
	}, nil
}

// GenerateHumanToken issues an HS256 token for a validated human principal.
func (m *JWTManager) GenerateHumanToken(subject string, role HumanRole) (string, error) {
	principal, err := NewHumanPrincipal(subject, role)
	if err != nil {
		return "", err
	}

	now := m.now().UTC()
	claims := Claims{
		ActorKind: principal.Kind(),
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   principal.Subject(),
			Audience:  jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(HumanTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseHumanToken validates the complete human JWT contract and returns the
// server-created principal. Tokens for other actor kinds are rejected.
func (m *JWTManager) ParseHumanToken(tokenString string) (Principal, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(*jwt.Token) (interface{}, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || token == nil || !token.Valid {
		return Principal{}, fmt.Errorf("%w: validation failed", ErrInvalidToken)
	}

	principal, err := NewHumanPrincipal(claims.Subject, claims.Role)
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	return principal, nil
}
