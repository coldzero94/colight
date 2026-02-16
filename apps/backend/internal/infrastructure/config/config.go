package config

import (
	"os"
	"time"
)

type Config struct {
	// Server
	APIPort string

	// Database
	DatabaseURL string

	// Naver OAuth
	NaverClientID     string
	NaverClientSecret string
	NaverCallbackURL  string

	// JWT
	JWTSecret          string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration

	// Frontend
	FrontendURL string

	// AI
	LLMLightProvider string // "gemini" (default) or "groq"
	GroqModel        string // Groq model ID or alias (default: llama-3.3-70b-versatile)
	GeminiAPIKey     string
	GroqAPIKey       string
	AnthropicAPIKey  string
	OpenAIAPIKey     string // embeddings only
}

func Load() *Config {
	return &Config{
		APIPort:            getEnv("API_PORT", "9000"),
		DatabaseURL:        mustGetEnv("DATABASE_URL"),
		NaverClientID:      getEnv("NAVER_CLIENT_ID", ""),
		NaverClientSecret:  getEnv("NAVER_CLIENT_SECRET", ""),
		NaverCallbackURL:   getEnv("NAVER_CALLBACK_URL", "http://localhost:9000/v1/auth/naver/callback"),
		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTAccessTokenTTL:  parseDuration("JWT_ACCESS_TOKEN_TTL", 1*time.Hour),
		JWTRefreshTokenTTL: parseDuration("JWT_REFRESH_TOKEN_TTL", 168*time.Hour),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:4000"),
		LLMLightProvider:   getEnv("LLM_LIGHT_PROVIDER", "gemini"),
		GroqModel:          getEnv("GROQ_MODEL", ""),
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		GroqAPIKey:         getEnv("GROQ_API_KEY", ""),
		AnthropicAPIKey:    getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env: " + key)
	}
	return v
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
