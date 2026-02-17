package middleware

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

func setupLoggerRouter(logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestLogger(logger))
	return r
}

func captureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return l, &buf
}

func TestRequestLogger_GeneratesTraceID(t *testing.T) {
	l, buf := captureLogger()
	r := setupLoggerRouter(l)
	r.GET("/test", func(c *gin.Context) {
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Response should have X-Trace-ID header
	traceID := w.Header().Get("X-Trace-ID")
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 8)

	// Log should contain the trace_id
	assert.Contains(t, buf.String(), "trace_id="+traceID)
}

func TestRequestLogger_AcceptsClientTraceID(t *testing.T) {
	l, buf := captureLogger()
	r := setupLoggerRouter(l)
	r.GET("/test", func(c *gin.Context) {
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", "client01")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "client01", w.Header().Get("X-Trace-ID"))
	assert.Contains(t, buf.String(), "trace_id=client01")
}

func TestRequestLogger_RejectsOversizedTraceID(t *testing.T) {
	l, _ := captureLogger()
	r := setupLoggerRouter(l)
	r.GET("/test", func(c *gin.Context) {
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", "this-is-a-very-long-trace-id-that-exceeds-36-chars-limit")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should generate a new 8-char trace ID instead
	traceID := w.Header().Get("X-Trace-ID")
	assert.Len(t, traceID, 8)
}

func TestRequestLogger_LogsInfoFor200(t *testing.T) {
	l, buf := captureLogger()
	r := setupLoggerRouter(l)
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()
	assert.Contains(t, output, "level=INFO")
	assert.Contains(t, output, "method=GET")
	assert.Contains(t, output, "path=/test")
	assert.Contains(t, output, "status=200")
}

func TestRequestLogger_LogsWarnFor4xx(t *testing.T) {
	l, buf := captureLogger()
	r := setupLoggerRouter(l)
	r.GET("/test", func(c *gin.Context) {
		c.Status(404)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Contains(t, buf.String(), "level=WARN")
	assert.Contains(t, buf.String(), "status=404")
}

func TestRequestLogger_LogsErrorFor5xx(t *testing.T) {
	l, buf := captureLogger()
	r := setupLoggerRouter(l)
	r.GET("/test", func(c *gin.Context) {
		c.Status(500)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Contains(t, buf.String(), "level=ERROR")
	assert.Contains(t, buf.String(), "status=500")
}

func TestRequestLogger_IncludesUserID(t *testing.T) {
	l, buf := captureLogger()
	r := setupLoggerRouter(l)
	userID := uuid.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Contains(t, buf.String(), "user_id="+userID.String())
}

func TestRequestLogger_SetsTraceIDInContext(t *testing.T) {
	l, _ := captureLogger()
	r := setupLoggerRouter(l)

	var ctxTraceID string
	r.GET("/test", func(c *gin.Context) {
		if v, ok := c.Get("trace_id"); ok {
			ctxTraceID = v.(string)
		}
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.NotEmpty(t, ctxTraceID)
	assert.Equal(t, ctxTraceID, w.Header().Get("X-Trace-ID"))
}
