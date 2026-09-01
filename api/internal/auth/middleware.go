package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const principalContextKey contextKey = "principal"

// PrincipalFromContext returns the authenticated principal stored by
// Middleware. The context key is private so callers cannot depend on raw
// credential claims or replace the server-created identity accidentally.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(Principal)
	if !ok || principal.Validate() != nil {
		return Principal{}, false
	}

	return principal, true
}

func withPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func Middleware(tokens *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"missing or invalid token"}}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			principal, err := tokens.ParseHumanToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid token"}}`, http.StatusUnauthorized)
				return
			}

			ctx := withPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireHumanRole allows only human principals with the requested role.
func RequireHumanRole(role HumanRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			principalRole, isHuman := principal.HumanRole()
			if !ok || !isHuman || principalRole != role {
				http.Error(w, `{"error":{"code":"FORBIDDEN","message":"insufficient role"}}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
