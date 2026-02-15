package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFeedbackService(t *testing.T) *FeedbackService {
	t.Helper()
	db := testutil.NewTestClient(t)
	return NewFeedbackService(db)
}

func TestFeedbackSubmit_Success(t *testing.T) {
	svc := newTestFeedbackService(t)
	ctx := context.Background()

	user := svc.db.UserProfile.Create().
		SetEmail("feedback-test@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	fb, err := svc.Submit(ctx, user.ID, FeedbackInput{
		Category:  "bug",
		Content:   "Something is broken",
		PageURL:   "/experiences",
		UserAgent: "Mozilla/5.0",
	})

	require.NoError(t, err)
	require.NotNil(t, fb)
	assert.Equal(t, "bug", string(fb.Category))
	assert.Equal(t, "Something is broken", fb.Content)
	assert.Equal(t, "/experiences", fb.PageURL)
	assert.Equal(t, "Mozilla/5.0", fb.UserAgent)
	assert.Equal(t, user.ID, fb.UserID)
}

func TestFeedbackSubmit_EmptyContent(t *testing.T) {
	svc := newTestFeedbackService(t)
	ctx := context.Background()

	user := svc.db.UserProfile.Create().
		SetEmail("feedback-empty@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	fb, err := svc.Submit(ctx, user.ID, FeedbackInput{
		Category: "improvement",
		Content:  "",
	})

	assert.ErrorIs(t, err, ErrEmptyFeedbackContent)
	assert.Nil(t, fb)
}
