package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultJWTSecret is the development-only fallback. Must not be used in production.
const DefaultJWTSecret = "memzent-enterprise-secret-2026"

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
	CORSAllowedOrigins     []string
}

func normalizeEnvironment(env string) string {
	return strings.ToLower(strings.TrimSpace(env))
}

func LoadConfig() *Config {
	env := normalizeEnvironment(getEnv("ENVIRONMENT", "development"))
	cfg := &Config{
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
		JWTSecret:              getEnv("JWT_SECRET", DefaultJWTSecret),
		JWKSURL:                getEnv("JWKS_URL", ""),
		SupabaseKey:            getEnv("SUPABASE_ANON_KEY", ""),
		LLMCacheTTL:            getEnvDuration("LLM_CACHE_TTL", 1*time.Hour),
		ToolRelevanceThreshold: getEnvFloat("TOOL_RELEVANCE_THRESHOLD", 0.7),
		Environment:            env,
		CORSAllowedOrigins:     parseCORSOrigins(env),
	}
	cfg.validateProduction()
	return cfg
}

func (c *Config) IsProduction() bool {
	return normalizeEnvironment(c.Environment) == "production"
}

func (c *Config) UsesDefaultJWTSecret() bool {
	return c.JWTSecret == DefaultJWTSecret
}

func (c *Config) validateProduction() {
	if !c.IsProduction() {
		return
	}
	if c.UsesDefaultJWTSecret() {
		slog.Error("Refusing to start: JWT_SECRET is the development default. Set a strong unique secret before production deployment.")
		os.Exit(1)
	}
	if len(c.CORSAllowedOrigins) == 0 {
		slog.Error("Refusing to start: CORS_ALLOWED_ORIGINS must be set in production (comma-separated list of allowed origins).")
		os.Exit(1)
	}
}

func parseCORSOrigins(environment string) []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		if normalizeEnvironment(environment) == "production" {
			return nil
		}
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
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
