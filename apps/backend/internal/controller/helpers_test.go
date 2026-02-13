package controller

import (
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAuthTokens(t *testing.T) {
	pair := &service.TokenPair{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresIn:    3600,
	}

	result := toAuthTokens(pair)

	assert.Equal(t, "access-123", result["access_token"])
	assert.Equal(t, "refresh-456", result["refresh_token"])
	assert.Equal(t, int64(3600), result["expires_in"])
}

func TestParseDate_ValidDate(t *testing.T) {
	parsed, err := parseDate("2024-06-15")
	require.NoError(t, err)
	assert.Equal(t, 2024, parsed.Year())
	assert.Equal(t, time.June, parsed.Month())
	assert.Equal(t, 15, parsed.Day())
}

func TestParseDate_InvalidDate(t *testing.T) {
	_, err := parseDate("not-a-date")
	require.Error(t, err)
}

func TestParseDate_EmptyString(t *testing.T) {
	_, err := parseDate("")
	require.Error(t, err)
}

func TestParseDate_DifferentFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid ISO", "2025-01-15", false},
		{"wrong separator", "2025/01/15", true},
		{"datetime", "2025-01-15T10:00:00Z", true},
		{"partial", "2025-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
