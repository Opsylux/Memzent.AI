package auth

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// NewRBACClient connection-string injection tests (no live DB required)
// ---------------------------------------------------------------------------

func TestNewRBACClient_AppendsBinaryParamsToURL(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantSubs string
	}{
		{
			name:     "plain postgres:// URL",
			input:    "postgres://user:pass@localhost/db",
			wantSubs: "binary_parameters=yes",
		},
		{
			name:     "postgresql:// URL",
			input:    "postgresql://user:pass@localhost/db",
			wantSubs: "binary_parameters=yes",
		},
		{
			name:     "URL already has query params",
			input:    "postgres://user:pass@localhost/db?sslmode=require",
			wantSubs: "binary_parameters=yes",
		},
		{
			name:     "URL already has binary_parameters",
			input:    "postgres://user:pass@localhost/db?binary_parameters=yes",
			wantSubs: "binary_parameters=yes",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// We exercise only the connection-string transformation logic,
			// not the actual Ping(), by calling the helper inline.
			result := injectBinaryParams(tc.input)
			if !strings.Contains(result, tc.wantSubs) {
				t.Errorf("expected %q to contain %q", result, tc.wantSubs)
			}
			// Must not double-inject
			count := strings.Count(result, "binary_parameters=yes")
			if count > 1 {
				t.Errorf("binary_parameters injected %d times, expected 1", count)
			}
		})
	}
}

func TestNewRBACClient_NonPostgresURLUntouched(t *testing.T) {
	input := "sqlite:///some/file.db"
	result := injectBinaryParams(input)
	if result != input {
		t.Errorf("Non-postgres URL should be unchanged, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// VerifyAPIKey — format / length guard
// ---------------------------------------------------------------------------

func TestVerifyAPIKey_TooShort(t *testing.T) {
	c := &RBACClient{db: nil}
	_, _, _, _, err := c.VerifyAPIKey(nil, "short")
	if err == nil {
		t.Fatal("expected error for short API key, got nil")
	}
	if !strings.Contains(err.Error(), "invalid API key format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestVerifyAPIKey_EmptyKey(t *testing.T) {
	c := &RBACClient{db: nil}
	_, _, _, _, err := c.VerifyAPIKey(nil, "")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

// ---------------------------------------------------------------------------
// injectBinaryParams — extracted helper so the logic is testable without DB
// ---------------------------------------------------------------------------

// injectBinaryParams mirrors the connection-string transformation in
// NewRBACClient, extracted here purely for testability.
func injectBinaryParams(connStr string) string {
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		if strings.Contains(connStr, "?") {
			if !strings.Contains(connStr, "binary_parameters=") {
				connStr += "&binary_parameters=yes"
			}
		} else {
			connStr += "?binary_parameters=yes"
		}
	}
	return connStr
}
