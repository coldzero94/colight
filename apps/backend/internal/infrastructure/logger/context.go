package logger

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FromGinContext returns a logger enriched with trace_id and user_id
// from the gin context. Controllers use this for contextual logging:
//
//	log := logger.FromGinContext(c)
//	log.Error("something failed", "error", err)
func FromGinContext(c *gin.Context) *slog.Logger {
	var attrs []any

	if traceID, ok := c.Get("trace_id"); ok {
		attrs = append(attrs, "trace_id", traceID)
	}
	if userID, ok := c.Get("user_id"); ok {
		if uid, ok := userID.(uuid.UUID); ok {
			attrs = append(attrs, "user_id", uid.String())
		}
	}

	if len(attrs) == 0 {
		return slog.Default()
	}
	return slog.Default().With(attrs...)
}
