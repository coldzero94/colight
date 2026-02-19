package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent/deletionrequest"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/jobs"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var adminID = uuid.New()

func makeJob() *river.Job[jobs.DeleteExpiredUsersArgs] {
	return &river.Job[jobs.DeleteExpiredUsersArgs]{
		JobRow: &rivertype.JobRow{},
	}
}

func TestDeleteExpiredUsersWorker_ProcessesDueRequests(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("delete-me@test.com").
		SetNickname("DeleteMe").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	db.DeletionRequest.Create().
		SetUserID(user.ID).
		SetStatus(deletionrequest.StatusPending).
		SetScheduledAt(time.Now().Add(-1 * time.Hour)).
		SetRequestedBy(adminID).
		SaveX(ctx)

	worker := jobs.NewDeleteExpiredUsersWorker(db)
	err := worker.Work(ctx, makeJob())
	require.NoError(t, err)

	// User deleted
	userCount, err := db.UserProfile.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, userCount)

	// DeletionRequest also gone (deleted by worker before user delete)
	reqCount, err := db.DeletionRequest.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, reqCount)
}

func TestDeleteExpiredUsersWorker_SkipsNotDueRequests(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("not-yet@test.com").
		SetNickname("NotYet").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	db.DeletionRequest.Create().
		SetUserID(user.ID).
		SetStatus(deletionrequest.StatusPending).
		SetScheduledAt(time.Now().Add(24 * time.Hour)).
		SetRequestedBy(adminID).
		SaveX(ctx)

	worker := jobs.NewDeleteExpiredUsersWorker(db)
	err := worker.Work(ctx, makeJob())
	require.NoError(t, err)

	// User still exists
	userCount, err := db.UserProfile.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, userCount)

	// Request still pending
	req, err := db.DeletionRequest.Query().First(ctx)
	require.NoError(t, err)
	assert.Equal(t, deletionrequest.StatusPending, req.Status)
}

func TestDeleteExpiredUsersWorker_SkipsCancelledRequests(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("cancelled@test.com").
		SetNickname("Cancelled").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	db.DeletionRequest.Create().
		SetUserID(user.ID).
		SetStatus(deletionrequest.StatusCancelled).
		SetScheduledAt(time.Now().Add(-1 * time.Hour)).
		SetCancelledAt(time.Now().Add(-30 * time.Minute)).
		SetRequestedBy(adminID).
		SaveX(ctx)

	worker := jobs.NewDeleteExpiredUsersWorker(db)
	err := worker.Work(ctx, makeJob())
	require.NoError(t, err)

	// User still exists (cancelled request ignored)
	userCount, err := db.UserProfile.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, userCount)
}

func TestDeleteExpiredUsersWorker_ProcessesMultiple(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	// 2 due users
	for i, email := range []string{"due1@test.com", "due2@test.com"} {
		u := db.UserProfile.Create().
			SetEmail(email).
			SetNickname("Due" + string(rune('1'+i))).
			SetAuthProvider(userprofile.AuthProviderEmail).
			SaveX(ctx)
		db.DeletionRequest.Create().
			SetUserID(u.ID).
			SetStatus(deletionrequest.StatusPending).
			SetScheduledAt(time.Now().Add(-1 * time.Hour)).
			SetRequestedBy(adminID).
			SaveX(ctx)
	}

	// 1 future user
	uFuture := db.UserProfile.Create().
		SetEmail("future@test.com").
		SetNickname("Future").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)
	db.DeletionRequest.Create().
		SetUserID(uFuture.ID).
		SetStatus(deletionrequest.StatusPending).
		SetScheduledAt(time.Now().Add(24 * time.Hour)).
		SetRequestedBy(adminID).
		SaveX(ctx)

	worker := jobs.NewDeleteExpiredUsersWorker(db)
	err := worker.Work(ctx, makeJob())
	require.NoError(t, err)

	// Only 1 future user remains
	userCount, err := db.UserProfile.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, userCount)

	// Only future request remains
	reqCount, err := db.DeletionRequest.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, reqCount)
}

func TestDeleteExpiredUsersWorker_EmptyQueue(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	worker := jobs.NewDeleteExpiredUsersWorker(db)
	err := worker.Work(ctx, makeJob())
	assert.NoError(t, err)
}
