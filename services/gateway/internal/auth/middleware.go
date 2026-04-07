package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserClaimsKey contextKey = "user_claims"
)

func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for health, metrics, etc.
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized: Missing Authorization Header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized: Invalid Authorization Header Format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			// 1. Parse and Validate Supabase JWT (RS256)
			// In production, we'd use the Supabase Project's JWT Public Key
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// For Supabase, the alg is typically HS256 (project secret) or RS256 (public key)
				return []byte(secret), nil // Placeholder for project secret validation
			})

			if err != nil || !token.Valid {
				http.Error(w, fmt.Sprintf("unauthorized: %v", err), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "unauthorized: invalid claims", http.StatusUnauthorized)
				return
			}

			// 2. Extract Multi-Tenant Identity (Stateless)
			// Supabase Auth Hooks can inject 'org_id' and 'tier' into app_metadata
			appMetadata, _ := claims["app_metadata"].(map[string]interface{})
			userID := claims["sub"].(string)
			orgID := "default-org" // Fallback
			if oid, ok := appMetadata["org_id"].(string); ok {
				orgID = oid
			}
			tier := "free"
			if t, ok := appMetadata["tier"].(string); ok {
				tier = t
			}

			// 3. Inject Tenant Context
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			ctx = context.WithValue(ctx, "user_id", userID)
			ctx = context.WithValue(ctx, "org_id", orgID)
			ctx = context.WithValue(ctx, "tier", tier)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GenerateJWT creates a signed JWT token with standard claims
func GenerateJWT(userID, role, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(duration).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
