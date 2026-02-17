package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_DevelopmentUsesTextHandler(t *testing.T) {
	l := New("development", "debug")
	assert.NotNil(t, l)
	assert.IsType(t, &slog.TextHandler{}, l.Handler())
}

func TestNew_ProductionUsesJSONHandler(t *testing.T) {
	l := New("production", "info")
	assert.NotNil(t, l)
	assert.IsType(t, &slog.JSONHandler{}, l.Handler())
}

func TestNew_StagingUsesJSONHandler(t *testing.T) {
	l := New("staging", "info")
	assert.NotNil(t, l)
	assert.IsType(t, &slog.JSONHandler{}, l.Handler())
}

func TestNew_UnknownEnvUsesTextHandler(t *testing.T) {
	l := New("", "info")
	assert.NotNil(t, l)
	assert.IsType(t, &slog.TextHandler{}, l.Handler())
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseLevel(tt.input))
		})
	}
}
