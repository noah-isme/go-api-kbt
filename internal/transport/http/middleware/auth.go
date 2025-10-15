package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-api-kbt/internal/config"
	userDomain "go-api-kbt/internal/domain/user"
	serviceAuth "go-api-kbt/internal/service/auth"
	httputil "go-api-kbt/internal/transport/httputil"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ( // Context keys
	ContextKeyUser contextKey = "user"
)

// AuthMiddleware is a middleware for JWT authentication.
func AuthMiddleware(cfg *config.AuthConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.RespondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				httputil.RespondError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenString := headerParts[1]
			claims := &serviceAuth.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
				return []byte(cfg.Secret), nil
			})

			if err != nil || !token.Valid || claims.Subject != "access" {
				httputil.RespondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Store user claims in context
			ctx := context.WithValue(r.Context(), ContextKeyUser, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole is a middleware that checks if the authenticated user has one of the required roles.
func RequireRole(roles ...userDomain.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKeyUser).(*serviceAuth.Claims)
			if !ok {
				httputil.RespondError(w, http.StatusForbidden, "access denied: user claims not found")
				return
			}

			hasRole := false
			for _, requiredRole := range roles {
				if claims.Role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				httputil.RespondError(w, http.StatusForbidden, "access denied: insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
