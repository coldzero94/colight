package service

import (
	"context"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageService_CheckLimit_FreeUser_BelowLimit(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("free@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanFree).
		SaveX(ctx)

	svc := NewUsageService(client)

	// Create 2 experience usage logs (limit is 3)
	for i := 0; i < 2; i++ {
		err := svc.TrackUsage(ctx, user.ID, "experience", nil)
		require.NoError(t, err)
	}

	status, err := svc.CheckLimit(ctx, user.ID, "experience")
	require.NoError(t, err)
	assert.True(t, status.Allowed)
	assert.Equal(t, 2, status.Used)
	assert.Equal(t, 3, status.Limit)
	assert.Equal(t, 1, status.Remaining)
}

func TestUsageService_CheckLimit_FreeUser_AtLimit(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("free2@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanFree).
		SaveX(ctx)

	svc := NewUsageService(client)

	// Create 3 experience usage logs (at limit)
	for i := 0; i < 3; i++ {
		err := svc.TrackUsage(ctx, user.ID, "experience", nil)
		require.NoError(t, err)
	}

	status, err := svc.CheckLimit(ctx, user.ID, "experience")
	require.NoError(t, err)
	assert.False(t, status.Allowed)
	assert.Equal(t, 3, status.Used)
	assert.Equal(t, 3, status.Limit)
	assert.Equal(t, 0, status.Remaining)
}

func TestUsageService_CheckLimit_PaidUser(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("paid@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanStarter).
		SaveX(ctx)

	svc := NewUsageService(client)

	status, err := svc.CheckLimit(ctx, user.ID, "experience")
	require.NoError(t, err)
	assert.True(t, status.Allowed)
	assert.Equal(t, -1, status.Limit) // unlimited
}

func TestUsageService_DailyReset(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("daily@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanFree).
		SaveX(ctx)

	svc := NewUsageService(client)

	// Insert a usage log with yesterday's timestamp
	yesterday := time.Now().Add(-24 * time.Hour)
	client.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("analysis").
		SetCreatedAt(yesterday).
		SaveX(ctx)

	// Yesterday's usage should not count for daily limit
	status, err := svc.CheckLimit(ctx, user.ID, "analysis")
	require.NoError(t, err)
	assert.True(t, status.Allowed)
	assert.Equal(t, 0, status.Used)
	assert.Equal(t, 1, status.Remaining)
}

func TestUsageService_TrackUsage(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("track@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	svc := NewUsageService(client)

	err := svc.TrackUsage(ctx, user.ID, "draft", map[string]interface{}{
		"cover_letter_id": "test-cl-id",
	})
	require.NoError(t, err)

	// Verify log exists
	count, err := client.UsageLog.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUsageService_InvalidFeature(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("invalid@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	svc := NewUsageService(client)

	_, err := svc.CheckLimit(ctx, user.ID, "nonexistent")
	assert.ErrorIs(t, err, ErrInvalidFeature)

	err = svc.TrackUsage(ctx, user.ID, "nonexistent", nil)
	assert.ErrorIs(t, err, ErrInvalidFeature)
}

func TestUsageService_GetAllUsage(t *testing.T) {
	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("all@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanFree).
		SaveX(ctx)

	svc := NewUsageService(client)

	// Track some usage
	_ = svc.TrackUsage(ctx, user.ID, "experience", nil)
	_ = svc.TrackUsage(ctx, user.ID, "draft", nil)

	features, plan, err := svc.GetAllUsage(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "free", plan)
	assert.Len(t, features, 5) // all 5 features

	assert.Equal(t, 1, features["experience"].Used)
	assert.Equal(t, 3, features["experience"].Limit)

	assert.Equal(t, 1, features["draft"].Used)
	assert.Equal(t, 1, features["draft"].Limit)

	assert.Equal(t, 0, features["analysis"].Used)
	assert.True(t, features["analysis"].Allowed)
}
