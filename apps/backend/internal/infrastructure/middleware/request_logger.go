package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger logs each HTTP request with trace_id, user_id, status, and duration.
// It generates a trace_id (or accepts one from the client via X-Trace-ID header)
// and sets it as a response header for frontend correlation.
func RequestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Trace ID: accept from client or generate
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" || len(traceID) > 36 {
			traceID = uuid.New().String()[:8]
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		start := time.Now()

		// 2. Process request
		c.Next()

		// 3. Log completed request
		duration := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("trace_id", traceID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.String("duration", duration.String()),
		}

		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uuid.UUID); ok {
				attrs = append(attrs, slog.String("user_id", uid.String()))
			}
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		base.LogAttrs(c.Request.Context(), level, "request", attrs...)
	}
}
