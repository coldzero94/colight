package logger

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestFromGinContext_WithTraceAndUserID(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	userID := uuid.New()

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("trace_id", "abc12345")
		c.Set("user_id", userID)

		log := FromGinContext(c)
		log.Info("test message")
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()
	assert.Contains(t, output, "trace_id=abc12345")
	assert.Contains(t, output, "user_id="+userID.String())
	assert.Contains(t, output, "test message")
}

func TestFromGinContext_TraceIDOnly(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("trace_id", "def67890")

		log := FromGinContext(c)
		log.Info("trace only")
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()
	assert.Contains(t, output, "trace_id=def67890")
	assert.NotContains(t, output, "user_id")
}

func TestFromGinContext_EmptyContext(t *testing.T) {
	r := gin.New()
	var gotLogger *slog.Logger
	r.GET("/test", func(c *gin.Context) {
		gotLogger = FromGinContext(c)
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.NotNil(t, gotLogger)
}
