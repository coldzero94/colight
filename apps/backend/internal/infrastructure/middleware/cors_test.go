package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCORSRouter(frontendURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware(frontendURL))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	r := setupCORSRouter("http://localhost:4000")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://localhost:4000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCORSMiddleware_PreflightOptions(t *testing.T) {
	r := setupCORSRouter("https://colight.kr")

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "https://colight.kr", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	// Body should be empty for preflight
	assert.Empty(t, w.Body.String())
}

func TestCORSMiddleware_DifferentOrigins(t *testing.T) {
	tests := []struct {
		name        string
		frontendURL string
	}{
		{"localhost", "http://localhost:4000"},
		{"production", "https://colight.kr"},
		{"staging", "https://staging.colight.kr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupCORSRouter(tt.frontendURL)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.frontendURL, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
