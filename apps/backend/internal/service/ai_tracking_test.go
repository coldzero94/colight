package service_test

import (
	"context"
	"errors"
	"testing"

	entaicallerror "github.com/coby/colight/apps/backend/ent/aicallerror"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogAICall_SuccessCreatesUsageLog(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("log-success@test.com").
		SetNickname("LogSuccess").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	err := service.LogAICall(ctx, db, service.AICallParams{
		UserID:       user.ID,
		Feature:      "draft",
		Model:        "gemini-2.0-flash",
		InputTokens:  100,
		OutputTokens: 50,
		LatencyMs:    500,
	})
	require.NoError(t, err)

	// usage_log created with status=success
	log, err := db.UsageLog.Query().First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "draft", log.Feature)
	assert.Equal(t, "success", log.Status)
	assert.Equal(t, "gemini", *log.Provider)
	assert.Equal(t, "gemini-2.0-flash", *log.Model)
	assert.Equal(t, 100, log.InputTokens)
	assert.Equal(t, 50, log.OutputTokens)

	// No ai_call_error
	count, err := db.AICallError.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestLogAICall_ErrorCreatesAICallError(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("log-error@test.com").
		SetNickname("LogError").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	callErr := errors.New("HTTP 503: upstream provider unavailable")
	err := service.LogAICall(ctx, db, service.AICallParams{
		UserID:  user.ID,
		Feature: "analysis",
		Model:   "gemini-2.0-flash",
		Err:     callErr,
	})
	require.NoError(t, err)

	// usage_log with status=error
	log, err := db.UsageLog.Query().First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "error", log.Status)

	// ai_call_error created
	errRecord, err := db.AICallError.Query().First(ctx)
	require.NoError(t, err)
	assert.Equal(t, entaicallerror.ErrorTypeProviderError, errRecord.ErrorType)
	assert.Equal(t, log.ID, errRecord.UsageLogID)
}

func TestLogAICall_RateLimitCreatesQuotaHitEvent(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("rl-test@test.com").
		SetNickname("RLTest").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	rlErr := errors.New("HTTP 429: quota exceeded for gemini-2.0-flash")
	err := service.LogAICall(ctx, db, service.AICallParams{
		UserID:  user.ID,
		Feature: "draft",
		Model:   "gemini-2.0-flash",
		Err:     rlErr,
	})
	require.NoError(t, err)

	// ai_call_error with rate_limit type
	errRecord, err := db.AICallError.Query().First(ctx)
	require.NoError(t, err)
	assert.Equal(t, entaicallerror.ErrorTypeRateLimit, errRecord.ErrorType)

	// quota_hit_event also created
	evt, err := db.QuotaHitEvent.Query().First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "gemini", evt.Provider)
	assert.Equal(t, "gemini-2.0-flash", evt.Model)
	assert.Equal(t, "draft", *evt.Feature)
	assert.Equal(t, errRecord.UsageLogID, *evt.UsageLogID)
}
