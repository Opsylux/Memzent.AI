package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWT(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateJWT("user123", "admin", secret, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["sub"] != "user123" {
		t.Errorf("expected sub=user123, got %v", claims["sub"])
	}
	if claims["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", claims["role"])
	}
}

func TestHasScope(t *testing.T) {
	// JWT auth method always has scope
	ctx := context.WithValue(context.Background(), "auth_method", "jwt")
	if !HasScope(ctx, "any:scope") {
		t.Errorf("JWT auth method should have any scope")
	}

	// API key with specific scope
	ctx = context.WithValue(context.Background(), "auth_method", "api_key")
	ctx = context.WithValue(ctx, "key_scopes", []string{"read:tool", "write:tool"})
	
	if !HasScope(ctx, "read:tool") {
		t.Errorf("Expected to have read:tool scope")
	}
	if HasScope(ctx, "admin:all") {
		t.Errorf("Should not have admin:all scope")
	}

	// API key with wildcard scope
	ctx = context.WithValue(context.Background(), "auth_method", "api_key")
	ctx = context.WithValue(ctx, "key_scopes", []string{"*"})
	
	if !HasScope(ctx, "admin:all") {
		t.Errorf("Wildcard scope should allow any scope")
	}

	// No scopes
	ctx = context.WithValue(context.Background(), "auth_method", "api_key")
	if HasScope(ctx, "read") {
		t.Errorf("Should not have scope when no key_scopes are set")
	}
}

func TestUnifiedAuthMiddleware_NoIdentity(t *testing.T) {
	middleware := UnifiedAuthMiddleware("secret", nil, nil)
	
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestUnifiedAuthMiddleware_SkipHealth(t *testing.T) {
	middleware := UnifiedAuthMiddleware("secret", nil, nil)
	
	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Errorf("Handler should be called for /healthz")
	}
}

func TestUnifiedAuthMiddleware_ValidJWT(t *testing.T) {
	secret := "test-secret"
	
	// Create a token with org_id in user_metadata
	claims := jwt.MapClaims{
		"sub": "user1",
		"user_metadata": map[string]interface{}{
			"org_id": "org1",
		},
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	middleware := UnifiedAuthMiddleware(secret, nil, nil)
	
	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Context().Value("org_id") != "org1" {
			t.Errorf("Expected org_id=org1 in context")
		}
		if r.Context().Value("user_id") != "user1" {
			t.Errorf("Expected user_id=user1 in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Errorf("Handler should be called for valid JWT")
	}
}
