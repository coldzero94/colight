package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New creates a structured logger based on environment and level.
// production/staging → JSON, development → Text.
func New(env string, level string) *slog.Logger {
	var handler slog.Handler
	lvl := parseLevel(level)

	switch env {
	case "production", "staging":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
