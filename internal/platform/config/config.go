package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        string
	Env         string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	JWTExpiry   time.Duration
	LogLevel    string
	GRPCPort    string

	// LLM provider (primary)
	LLMProvider string // LLM_PROVIDER: claude | openai | ollama
	LLMModel    string // LLM_MODEL
	LLMAPIKey   string // LLM_API_KEY

	// LLM provider (fallback — optional)
	LLMFallbackProvider string // LLM_FALLBACK_PROVIDER
	LLMFallbackModel    string // LLM_FALLBACK_MODEL
	LLMFallbackAPIKey   string // LLM_FALLBACK_API_KEY

	// Ollama server URL (used when provider = ollama)
	OllamaURL string // OLLAMA_URL, default http://localhost:11434
}

// Load reads configuration from environment variables with sensible defaults.
// Returns error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),

		LLMProvider: getEnv("LLM_PROVIDER", "ollama"),
		LLMModel:    getEnv("LLM_MODEL", "llama3"),
		LLMAPIKey:   os.Getenv("LLM_API_KEY"),

		LLMFallbackProvider: os.Getenv("LLM_FALLBACK_PROVIDER"),
		LLMFallbackModel:    os.Getenv("LLM_FALLBACK_MODEL"),
		LLMFallbackAPIKey:   os.Getenv("LLM_FALLBACK_API_KEY"),

		OllamaURL: getEnv("OLLAMA_URL", "http://localhost:11434"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	expiry := getEnv("JWT_EXPIRY", "24h")
	d, err := time.ParseDuration(expiry)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY %q: %w", expiry, err)
	}
	cfg.JWTExpiry = d

	return cfg, nil
}

// IsDev returns true if running in development mode.
func (c *Config) IsDev() bool {
	return c.Env == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
