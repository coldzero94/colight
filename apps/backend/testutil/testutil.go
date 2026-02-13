package testutil

import (
	"time"

	"github.com/coby/colight/apps/backend/internal/service"
)

const (
	TestJWTSecret = "test-secret-key-minimum-32-chars!!"
)

// NewTestTokenService creates a TokenService for testing with short TTLs.
func NewTestTokenService() *service.TokenService {
	return service.NewTokenService(TestJWTSecret, 1*time.Hour, 7*24*time.Hour)
}
