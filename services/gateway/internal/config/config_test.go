package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	keys := []string{
		"PORT", "VALKEY_URL", "ROUTER_URL", "POSTGRES_URL", "MCP_SERVER_URL",
		"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "OPENAI_MODEL", "GEMINI_API_KEY",
		"GEMINI_MODEL", "OLLAMA_ENABLED", "OLLAMA_URL", "OLLAMA_MODEL",
		"JWT_SECRET", "JWKS_URL", "SUPABASE_ANON_KEY", "LLM_CACHE_TTL",
		"TOOL_RELEVANCE_THRESHOLD", "ENVIRONMENT",
	}

	originalValues := make(map[string]string)
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			originalValues[key] = val
			os.Unsetenv(key)
		}
	}

	defer func() {
		for key, val := range originalValues {
			os.Setenv(key, val)
		}
	}()

	cfg := LoadConfig()

	if cfg.Port != ":8080" {
		t.Errorf("expected default Port :8080, got %q", cfg.Port)
	}
	if cfg.ValkeyURL != "http://localhost:6379" {
		t.Errorf("expected default ValkeyURL http://localhost:6379, got %q", cfg.ValkeyURL)
	}
	if cfg.OllamaEnabled != true {
		t.Errorf("expected default OllamaEnabled true, got %t", cfg.OllamaEnabled)
	}
	if cfg.LLMCacheTTL != 1*time.Hour {
		t.Errorf("expected default LLMCacheTTL 1h, got %v", cfg.LLMCacheTTL)
	}
	if cfg.ToolRelevanceThreshold != 0.7 {
		t.Errorf("expected default ToolRelevanceThreshold 0.7, got %f", cfg.ToolRelevanceThreshold)
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	overrides := map[string]string{
		"PORT":                     ":9090",
		"VALKEY_URL":               "valkey://custom:6379",
		"ROUTER_URL":               "custom-router:50051",
		"POSTGRES_URL":             "postgres://custom-db",
		"MCP_SERVER_URL":           "http://custom-mcp",
		"ANTHROPIC_API_KEY":        "anthropic-key",
		"OPENAI_API_KEY":           "openai-key",
		"OPENAI_MODEL":             "gpt-4o",
		"GEMINI_API_KEY":           "gemini-key",
		"GEMINI_MODEL":             "gemini-ultra",
		"OLLAMA_ENABLED":           "false",
		"OLLAMA_URL":               "http://custom-ollama",
		"OLLAMA_MODEL":             "llama-custom",
		"JWT_SECRET":               "custom-secret",
		"JWKS_URL":                 "http://custom-jwks",
		"SUPABASE_ANON_KEY":        "custom-supabase",
		"LLM_CACHE_TTL":            "30m",
		"TOOL_RELEVANCE_THRESHOLD": "0.85",
		"ENVIRONMENT":              "production",
	}

	originalValues := make(map[string]string)
	for key := range overrides {
		if val, ok := os.LookupEnv(key); ok {
			originalValues[key] = val
		}
	}

	for key, val := range overrides {
		os.Setenv(key, val)
	}

	defer func() {
		for key := range overrides {
			if val, ok := originalValues[key]; ok {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	cfg := LoadConfig()

	if cfg.Port != ":9090" {
		t.Errorf("expected custom Port :9090, got %q", cfg.Port)
	}
	if cfg.ValkeyURL != "valkey://custom:6379" {
		t.Errorf("expected custom ValkeyURL valkey://custom:6379, got %q", cfg.ValkeyURL)
	}
	if cfg.OllamaEnabled != false {
		t.Errorf("expected custom OllamaEnabled false, got %t", cfg.OllamaEnabled)
	}
	if cfg.LLMCacheTTL != 30*time.Minute {
		t.Errorf("expected custom LLMCacheTTL 30m, got %v", cfg.LLMCacheTTL)
	}
	if cfg.ToolRelevanceThreshold != 0.85 {
		t.Errorf("expected custom ToolRelevanceThreshold 0.85, got %f", cfg.ToolRelevanceThreshold)
	}
}

func TestLoadConfig_InvalidFormats(t *testing.T) {
	overrides := map[string]string{
		"LLM_CACHE_TTL":            "invalid-duration",
		"TOOL_RELEVANCE_THRESHOLD": "invalid-float",
	}

	originalValues := make(map[string]string)
	for key := range overrides {
		if val, ok := os.LookupEnv(key); ok {
			originalValues[key] = val
		}
	}

	for key, val := range overrides {
		os.Setenv(key, val)
	}

	defer func() {
		for key := range overrides {
			if val, ok := originalValues[key]; ok {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	cfg := LoadConfig()

	if cfg.LLMCacheTTL != 1*time.Hour {
		t.Errorf("expected default LLMCacheTTL 1h when overridden with invalid format, got %v", cfg.LLMCacheTTL)
	}
	if cfg.ToolRelevanceThreshold != 0.7 {
		t.Errorf("expected default ToolRelevanceThreshold 0.7 when overridden with invalid format, got %f", cfg.ToolRelevanceThreshold)
	}
}

func TestParseCORSOrigins_DevelopmentDefault(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	origins := parseCORSOrigins("development")
	if len(origins) != 1 || origins[0] != "*" {
		t.Fatalf("expected dev wildcard, got %v", origins)
	}
}

func TestParseCORSOrigins_ProductionEmpty(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	origins := parseCORSOrigins("production")
	if len(origins) != 0 {
		t.Fatalf("expected no default origins in production, got %v", origins)
	}
}

func TestParseCORSOrigins_ExplicitList(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3002, https://app.memzent.ai ")
	origins := parseCORSOrigins("production")
	if len(origins) != 2 || origins[0] != "http://localhost:3002" || origins[1] != "https://app.memzent.ai" {
		t.Fatalf("unexpected origins: %v", origins)
	}
}

func TestConfig_IsProduction_CaseInsensitive(t *testing.T) {
	for _, env := range []string{"production", "Production", "PRODUCTION", " production "} {
		c := &Config{Environment: env}
		if !c.IsProduction() {
			t.Fatalf("IsProduction() should be true for %q", env)
		}
	}
	if (&Config{Environment: "staging"}).IsProduction() {
		t.Fatal("staging should not be production")
	}
}

func TestParseCORSOrigins_ProductionCaseInsensitive(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	origins := parseCORSOrigins("Production")
	if len(origins) != 0 {
		t.Fatalf("expected no default origins for Production, got %v", origins)
	}
}

func TestConfig_UsesDefaultJWTSecret(t *testing.T) {
	c := &Config{JWTSecret: DefaultJWTSecret}
	if !c.UsesDefaultJWTSecret() {
		t.Fatal("expected default secret detection")
	}
	c.JWTSecret = "custom"
	if c.UsesDefaultJWTSecret() {
		t.Fatal("expected non-default secret")
	}
}

func TestLoadConfig_ReadsEnvironment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("JWT_SECRET", "test-secret-not-default")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3002")
	cfg := LoadConfig()
	if cfg.Environment != "staging" {
		t.Fatalf("expected staging, got %s", cfg.Environment)
	}
	_ = os.Unsetenv("ENVIRONMENT")
}
