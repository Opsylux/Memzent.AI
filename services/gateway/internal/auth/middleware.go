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

func UnifiedAuthMiddleware(secret string, jwks *JWKSProvider, rbac *RBACClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for health, metrics, etc.
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Try API Key Authentication (X-API-Key)
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" && rbac != nil {
				orgID, userID, scopes, role, err := rbac.VerifyAPIKey(r.Context(), apiKey)
				if err == nil {
					ctx := context.WithValue(r.Context(), "user_id", userID)
					ctx = context.WithValue(ctx, "org_id", orgID)
					ctx = context.WithValue(ctx, "tier", "pro") // Default tier for API access
					ctx = context.WithValue(ctx, "user_role", role)
					ctx = context.WithValue(ctx, "key_role", role)
					ctx = context.WithValue(ctx, "key_scopes", scopes)
					ctx = context.WithValue(ctx, "auth_method", "api_key")

					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// If key was provided but invalid, we continue to check JWT or return error later
			}

			// 2. Try JWT Authentication (Authorization: Bearer <token>)
			authHeader := r.Header.Get("Authorization")
			xOrgID := r.Header.Get("X-Org-ID") // Get X-Org-ID as a potential fallback/override

			if authHeader == "" {
				http.Error(w, "Unauthorized: Missing identity (JWT or API Key)", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized: Invalid Authorization Header Format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			// Parse and Validate token using dynamic JWKS or static secret
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// 1. If JWKS is enabled and token has a Key ID (`kid`), try discovery
				if jwks != nil {
					if kid, ok := token.Header["kid"].(string); ok {
						return jwks.GetKey(kid)
					}
				}

				// 2. Fallback to Algorithm-Aware Static Secret (HMAC/ECDSA/RSA)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
					return []byte(secret), nil
				}

				if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok {
					pub, err := jwt.ParseECPublicKeyFromPEM([]byte(secret))
					if err != nil {
						return nil, fmt.Errorf("token uses ECDSA but secret is not a valid PEM Public Key: %v", err)
					}
					return pub, nil
				}

				if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
					pub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(secret))
					if err != nil {
						return nil, fmt.Errorf("token uses RSA but secret is not a valid PEM Public Key: %v", err)
					}
					return pub, nil
				}

				return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
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

			// Extract Multi-Tenant Identity
			appMetadata, _ := claims["app_metadata"].(map[string]interface{})
			userMetadata, _ := claims["user_metadata"].(map[string]interface{})
			userID := claims["sub"].(string)

			// Resolve Org ID (Priority: JWT app_metadata -> JWT user_metadata -> X-Org-ID Header)
			var orgID string
			
			if oid, ok := appMetadata["org_id"].(string); ok {
				orgID = oid
			} else if oid, ok := userMetadata["org_id"].(string); ok {
				orgID = oid
			} else if xOrgID != "" {
				// Trust X-Org-ID if provided by the dashboard as long as the JWT is valid
				orgID = xOrgID
			}

			// Block requests that lack an organizational context (required for RBAC and Audit Logging)
			if orgID == "" {
				http.Error(w, "Forbidden: Organizational context missing. Please ensure your account has a workspace.", http.StatusForbidden)
				return
			}
			
			tier := "free"
			if t, ok := appMetadata["tier"].(string); ok {
				tier = t
			} else if t, ok := userMetadata["tier"].(string); ok {
				tier = t
			}

			// Get initial role from JWT claims
			role, _ := claims["role"].(string)

			// Resolve verified Role from Database (Persistent RBAC)
			if rbac != nil && userID != "" {
				dbRole, err := rbac.GetMemberRole(r.Context(), orgID, userID)
				if err == nil && dbRole != "guest" {
					role = dbRole
				} else if role == "authenticated" || role == "" {
					// If mapping from DB failed and JWT is just 'authenticated', treat as guest
					role = "guest"
				}
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			ctx = context.WithValue(ctx, "user_id", userID)
			ctx = context.WithValue(ctx, "org_id", orgID)
			ctx = context.WithValue(ctx, "tier", tier)
			ctx = context.WithValue(ctx, "user_role", role)
			ctx = context.WithValue(ctx, "key_role", role)
			ctx = context.WithValue(ctx, "key_scopes", []string{"*"})
			ctx = context.WithValue(ctx, "auth_method", "jwt")

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

// HasScope checks if the authenticated context has a specific permission scope.
func HasScope(ctx context.Context, requiredScope string) bool {
	authMethod, _ := ctx.Value("auth_method").(string)
	if authMethod == "jwt" {
		// JWT users have full scope permission bypass, governed by their org membership/role checks.
		return true
	}

	scopes, ok := ctx.Value("key_scopes").([]string)
	if !ok {
		return false
	}

	for _, s := range scopes {
		if s == "*" || s == requiredScope {
			return true
		}
	}
	return false
}
