package ai

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordQuotaHit_AppendsEvent(t *testing.T) {
	throttler := NewModelThrottler()

	throttler.RecordQuotaHit("gemini-3-pro", fmt.Errorf("429 rate limit"))
	throttler.RecordQuotaHit("llama-3.3-70b-versatile", fmt.Errorf("quota exceeded"))

	summary := throttler.GetQuotaHitSummary(time.Hour)
	assert.Len(t, summary, 2)
}

func TestRecordQuotaHit_RingBufferLimit(t *testing.T) {
	throttler := NewModelThrottler()

	// Fill beyond 200-event buffer
	for i := 0; i < 210; i++ {
		throttler.RecordQuotaHit("gemini-3-pro", fmt.Errorf("429 rate limit #%d", i))
	}

	// Internal buffer should be capped at 200
	throttler.hitMu.Lock()
	assert.Len(t, throttler.quotaHits, 200)
	throttler.hitMu.Unlock()

	summary := throttler.GetQuotaHitSummary(time.Hour)
	require.Len(t, summary, 1)
	assert.Equal(t, 200, summary[0].HitCount)
}

func TestGetQuotaHitSummary_FiltersByDuration(t *testing.T) {
	throttler := NewModelThrottler()

	// Record a hit
	throttler.RecordQuotaHit("gemini-3-pro", fmt.Errorf("429"))

	// Manually insert an old event
	throttler.hitMu.Lock()
	throttler.quotaHits = append([]QuotaHitEvent{{
		Model:     "old-model",
		Error:     "old error",
		Timestamp: time.Now().Add(-2 * time.Hour),
	}}, throttler.quotaHits...)
	throttler.hitMu.Unlock()

	// Query last 1 hour - should only return gemini-3-pro
	summary := throttler.GetQuotaHitSummary(time.Hour)
	assert.Len(t, summary, 1)
	assert.Equal(t, "gemini-3-pro", summary[0].Model)

	// Query last 3 hours - should return both
	summary = throttler.GetQuotaHitSummary(3 * time.Hour)
	assert.Len(t, summary, 2)
}

func TestGetQuotaHitSummary_AggregatesByModel(t *testing.T) {
	throttler := NewModelThrottler()

	throttler.RecordQuotaHit("gemini-3-pro", fmt.Errorf("429 first"))
	throttler.RecordQuotaHit("gemini-3-pro", fmt.Errorf("429 second"))
	throttler.RecordQuotaHit("llama-3.3-70b-versatile", fmt.Errorf("quota exceeded"))

	summary := throttler.GetQuotaHitSummary(time.Hour)
	assert.Len(t, summary, 2)

	for _, s := range summary {
		if s.Model == "gemini-3-pro" {
			assert.Equal(t, 2, s.HitCount)
			assert.Equal(t, "gemini", s.Provider)
			assert.Equal(t, "429 second", s.LastError)
		} else if s.Model == "llama-3.3-70b-versatile" {
			assert.Equal(t, 1, s.HitCount)
			assert.Equal(t, "groq", s.Provider)
		}
	}
}

func TestGetQuotaHitSummary_EmptyReturnsEmpty(t *testing.T) {
	throttler := NewModelThrottler()
	summary := throttler.GetQuotaHitSummary(time.Hour)
	assert.Len(t, summary, 0)
}

func TestGetModelLimits_KnownModel(t *testing.T) {
	limits := GetModelLimits("gemini-3-pro")
	assert.Equal(t, 15, limits.RPM)
	assert.Equal(t, 1500, limits.RPD)
	assert.Equal(t, "gemini", limits.Provider)
}

func TestGetModelLimits_ShortAlias(t *testing.T) {
	limits := GetModelLimits("llama-3.3-70b")
	assert.Equal(t, 30, limits.RPM)
	assert.Equal(t, "groq", limits.Provider)
}

func TestGetModelLimits_Unknown(t *testing.T) {
	limits := GetModelLimits("nonexistent-model")
	assert.Equal(t, 0, limits.RPM)
	assert.Equal(t, "unknown", limits.Provider)
}

func TestIsSelectableModel(t *testing.T) {
	assert.True(t, IsSelectableModel("gemini-3-pro"))
	assert.True(t, IsSelectableModel("llama-3.3-70b-versatile"))
	assert.True(t, IsSelectableModel("llama-3.3-70b")) // short alias
	assert.False(t, IsSelectableModel("gpt-4"))
	assert.False(t, IsSelectableModel("claude-sonnet"))
}

func TestGetSelectableModels_ReturnsAll(t *testing.T) {
	models := GetSelectableModels()
	assert.Greater(t, len(models), 10) // We have 16 models

	// Verify each has required fields
	for _, m := range models {
		assert.NotEmpty(t, m.ID)
		assert.NotEmpty(t, m.Provider)
		assert.True(t, m.Provider == "gemini" || m.Provider == "groq")
	}
}
