package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAdminRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Simulate auth middleware setting role
	r.Use(func(c *gin.Context) {
		if role := c.GetHeader("X-Test-Role"); role != "" {
			c.Set("role", role)
		}
		c.Next()
	})
	r.Use(AdminMiddleware())
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestAdminMiddleware_AdminRole(t *testing.T) {
	r := setupAdminRouter()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestAdminMiddleware_UserRole(t *testing.T) {
	r := setupAdminRouter()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Test-Role", "user")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "AUTH_003")
}

func TestAdminMiddleware_NoRole(t *testing.T) {
	r := setupAdminRouter()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "AUTH_003")
}
