package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRoleLevel(t *testing.T) {
	tests := []struct {
		role  string
		level int
	}{
		{"user", 1},
		{"manager", 2},
		{"admin", 3},
		{"super_admin", 4},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			assert.Equal(t, tt.level, RoleLevel(tt.role))
		})
	}
}

func setupRoleRouter(minRole string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if role := c.GetHeader("X-Test-Role"); role != "" {
			c.Set("role", role)
		}
		c.Next()
	})
	r.Use(RequireRole(minRole))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestRequireRole_ManagerAccess(t *testing.T) {
	r := setupRoleRouter("manager")

	tests := []struct {
		name   string
		role   string
		status int
	}{
		{"manager allowed", "manager", 200},
		{"admin allowed", "admin", 200},
		{"super_admin allowed", "super_admin", 200},
		{"user denied", "user", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Test-Role", tt.role)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestRequireRole_AdminAccess(t *testing.T) {
	r := setupRoleRouter("admin")

	tests := []struct {
		name   string
		role   string
		status int
	}{
		{"admin allowed", "admin", 200},
		{"super_admin allowed", "super_admin", 200},
		{"manager denied", "manager", 403},
		{"user denied", "user", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Test-Role", tt.role)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestRequireRole_SuperAdminAccess(t *testing.T) {
	r := setupRoleRouter("super_admin")

	tests := []struct {
		name   string
		role   string
		status int
	}{
		{"super_admin allowed", "super_admin", 200},
		{"admin denied", "admin", 403},
		{"manager denied", "manager", 403},
		{"user denied", "user", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Test-Role", tt.role)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestRequireRole_InsufficientRole(t *testing.T) {
	r := setupRoleRouter("admin")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-Role", "user")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "AUTH_003")
}

func TestRequireRole_NoRole(t *testing.T) {
	r := setupRoleRouter("admin")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}
