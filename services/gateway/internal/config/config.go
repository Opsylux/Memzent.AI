package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// insecureDefaultJWTSecret is the historical placeholder secret. It's fine
// for local development but must never be used in production, since it is
// public (visible in this source file / git history).
const insecureDefaultJWTSecret = "memzent-enterprise-secret-2026"

type Config struct {
	Port                   string
	ValkeyURL              string
	RouterURL              string
	PostgresURL            string
	MCPURL                 string
	AnthropicAPIKey        string
	OpenAIAPIKey           string
	OpenAIModel            string
	GeminiAPIKey           string
	GeminiModel            string
	OllamaEnabled          bool
	OllamaURL              string
	OllamaModel            string
	JWTSecret              string
	JWKSURL                string
	SupabaseKey            string
	LLMCacheTTL            time.Duration
	ToolRelevanceThreshold float64
	Environment            string
	AllowedOrigins         []string
	DevAdminBypass         bool
}

func LoadConfig() *Config {
	return &Config{
		Port:                   getEnv("PORT", ":8080"),
		ValkeyURL:              getEnv("VALKEY_URL", "http://localhost:6379"),
		RouterURL:              getEnv("ROUTER_URL", "router:50051"),
		PostgresURL:            getEnv("POSTGRES_URL", "postgres://user:password@postgres:5432/memzent_db?sslmode=disable"),
		MCPURL:                 getEnv("MCP_SERVER_URL", "http://memzent-mcp-server:50052/mcp"),
		AnthropicAPIKey:        getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIAPIKey:           getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:            getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		GeminiAPIKey:           getEnv("GEMINI_API_KEY", ""),
		GeminiModel:            getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		OllamaEnabled:          getEnv("OLLAMA_ENABLED", "true") == "true",
		OllamaURL:              getEnv("OLLAMA_URL", "http://host.docker.internal:11434"),
		OllamaModel:            getEnv("OLLAMA_MODEL", "llama3.2"),
		JWTSecret:              getEnv("JWT_SECRET", insecureDefaultJWTSecret),
		JWKSURL:                getEnv("JWKS_URL", ""),
		SupabaseKey:            getEnv("SUPABASE_ANON_KEY", ""),
		LLMCacheTTL:            getEnvDuration("LLM_CACHE_TTL", 1*time.Hour),
		ToolRelevanceThreshold: getEnvFloat("TOOL_RELEVANCE_THRESHOLD", 0.7),
		Environment:            getEnv("ENVIRONMENT", "development"),
		AllowedOrigins:         getEnvList("ALLOWED_ORIGINS", "*"),
		DevAdminBypass:         getEnv("MEMZENT_DEV_ADMIN_BYPASS", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvList(key string, fallback string) []string {
	value := getEnv(key, fallback)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Validate fails closed on insecure configuration when running in production.
// It intentionally does not block development/test environments so local
// setup keeps working without extra ceremony.
func (c *Config) Validate() error {
	if c.Environment != "production" {
		return nil
	}

	var problems []string

	if c.JWTSecret == insecureDefaultJWTSecret || len(c.JWTSecret) < 32 {
		problems = append(problems, "JWT_SECRET must be set to a unique value of at least 32 characters in production")
	}
	for _, origin := range c.AllowedOrigins {
		if origin == "*" {
			problems = append(problems, "ALLOWED_ORIGINS must not be \"*\" in production")
			break
		}
	}
	if strings.Contains(c.PostgresURL, "sslmode=disable") {
		problems = append(problems, "POSTGRES_URL must not use sslmode=disable in production")
	}
	if c.DevAdminBypass {
		problems = append(problems, "MEMZENT_DEV_ADMIN_BYPASS must not be enabled in production")
	}

	if len(problems) > 0 {
		return fmt.Errorf("insecure production configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}
