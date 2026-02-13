package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-minimum-32-chars!!"

func newTestTokenService() *service.TokenService {
	return service.NewTokenService(testSecret, 1*time.Hour, 7*24*time.Hour)
}

func setupAuthRouter(tokenService *service.TokenService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(tokenService))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		c.JSON(200, gin.H{"user_id": userID, "role": role})
	})
	return r
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	ts := newTestTokenService()
	userID := uuid.New()

	pair, err := ts.IssueTokenPair(userID, userprofile.RoleUser)
	require.NoError(t, err)

	r := setupAuthRouter(ts)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	ts := newTestTokenService()
	r := setupAuthRouter(ts)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "AUTH_001")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	ts := newTestTokenService()
	r := setupAuthRouter(ts)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "AUTH_002")
}
