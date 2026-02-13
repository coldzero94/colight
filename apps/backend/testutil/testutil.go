package testutil

import (
	"time"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
)

const (
	TestJWTSecret       = "test-secret-key-minimum-32-chars!!"
	TestAccessTokenTTL  = 1 * time.Hour
	TestRefreshTokenTTL = 7 * 24 * time.Hour
)

// NewTestConfig creates a Config suitable for testing.
func NewTestConfig() *config.Config {
	return &config.Config{
		APIPort:            "9000",
		NaverClientID:      "test-naver-client-id",
		NaverClientSecret:  "test-naver-client-secret",
		NaverCallbackURL:   "http://localhost:9000/v1/auth/naver/callback",
		JWTSecret:          TestJWTSecret,
		JWTAccessTokenTTL:  TestAccessTokenTTL,
		JWTRefreshTokenTTL: TestRefreshTokenTTL,
		FrontendURL:        "http://localhost:4000",
	}
}
