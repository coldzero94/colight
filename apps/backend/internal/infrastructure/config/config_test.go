package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_KEY", "hello")
	assert.Equal(t, "hello", getEnv("TEST_KEY", "fallback"))
}

func TestGetEnv_Fallback(t *testing.T) {
	assert.Equal(t, "fallback", getEnv("NONEXISTENT_KEY_XYZ", "fallback"))
}

func TestMustGetEnv_WithValue(t *testing.T) {
	t.Setenv("MUST_KEY", "value")
	assert.Equal(t, "value", mustGetEnv("MUST_KEY"))
}

func TestMustGetEnv_Panics(t *testing.T) {
	assert.Panics(t, func() {
		mustGetEnv("NONEXISTENT_MUST_KEY_XYZ")
	})
}

func TestParseDuration_WithValue(t *testing.T) {
	t.Setenv("DUR_KEY", "30m")
	assert.Equal(t, 30*time.Minute, parseDuration("DUR_KEY", time.Hour))
}

func TestParseDuration_Fallback(t *testing.T) {
	assert.Equal(t, time.Hour, parseDuration("NONEXISTENT_DUR_KEY", time.Hour))
}

func TestParseDuration_InvalidFormat(t *testing.T) {
	t.Setenv("BAD_DUR", "notaduration")
	assert.Equal(t, time.Hour, parseDuration("BAD_DUR", time.Hour))
}

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg := Load()

	assert.Equal(t, "9000", cfg.APIPort)
	assert.Equal(t, "postgres://test:test@localhost/test", cfg.DatabaseURL)
	assert.Equal(t, "test-secret", cfg.JWTSecret)
	assert.Equal(t, "http://localhost:4000", cfg.FrontendURL)
	assert.Equal(t, 1*time.Hour, cfg.JWTAccessTokenTTL)
	assert.Equal(t, 168*time.Hour, cfg.JWTRefreshTokenTTL)
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom/db")
	t.Setenv("JWT_SECRET", "custom-secret")
	t.Setenv("API_PORT", "8080")
	t.Setenv("FRONTEND_URL", "https://colight.kr")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "30m")

	cfg := Load()

	assert.Equal(t, "8080", cfg.APIPort)
	assert.Equal(t, "https://colight.kr", cfg.FrontendURL)
	assert.Equal(t, 30*time.Minute, cfg.JWTAccessTokenTTL)
}

func TestLoad_PanicsWithoutRequired(t *testing.T) {
	// Unset required env vars
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "some-secret")

	assert.Panics(t, func() {
		Load()
	})
}
