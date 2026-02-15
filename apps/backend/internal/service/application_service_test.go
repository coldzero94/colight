package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createApplicationForTest(t *testing.T, ctx context.Context) (*ApplicationService, uuid.UUID) {
	t.Helper()
	client := testutil.NewTestClient(t)
	svc := NewApplicationService(client)

	email := fmt.Sprintf("app-%s@example.com", uuid.New().String()[:8])
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	return svc, user.ID
}

func TestApplicationService_ListApplications(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	// Create applications with cover letters
	deadline := time.Now().Add(48 * time.Hour)
	app1 := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SetPosition("백엔드 개발자").
		SetStatus(application.StatusPreparing).
		SetDeadline(deadline).
		SetNotes("메모").
		SaveX(ctx)

	// Add cover letters to app1
	svc.db.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(app1.ID).
		SetQuestionText("문항1").
		SetCurrentContent("내용1").
		SaveX(ctx)
	svc.db.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(app1.ID).
		SetQuestionText("문항2").
		SetCurrentContent("내용2").
		SaveX(ctx)

	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("네이버").
		SetPosition("프론트엔드 개발자").
		SetStatus(application.StatusSubmitted).
		SaveX(ctx)

	apps, err := svc.ListApplications(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, apps, 2)

	// First app should have cover letter count = 2 and deadline set
	var samsung *ApplicationDetail
	for _, a := range apps {
		if a.CompanyName == "삼성전자" {
			samsung = &a
			break
		}
	}
	require.NotNil(t, samsung)
	assert.Equal(t, 2, samsung.CoverLetterCount)
	assert.NotNil(t, samsung.Deadline)
	assert.Equal(t, "메모", samsung.Notes)
	assert.Equal(t, "preparing", samsung.Status)
}

func TestApplicationService_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("카카오").
		SetPosition("서버 개발자").
		SetStatus(application.StatusPreparing).
		SaveX(ctx)

	result, err := svc.UpdateStatus(ctx, userID, app.ID, "submitted")
	require.NoError(t, err)
	assert.Equal(t, "submitted", result.Status)
	assert.Equal(t, "카카오", result.CompanyName)
}

func TestApplicationService_UpdateStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	_, err := svc.UpdateStatus(ctx, userID, uuid.New(), "submitted")
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestApplicationService_UpdateStatus_Forbidden(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("토스").
		SetPosition("백엔드").
		SaveX(ctx)

	otherUser := uuid.New()
	_, err := svc.UpdateStatus(ctx, otherUser, app.ID, "submitted")
	assert.ErrorIs(t, err, ErrApplicationForbidden)
}

func TestApplicationService_GetStats(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	// Create applications with different statuses
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성").
		SetPosition("BE").
		SetStatus(application.StatusPreparing).
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("네이버").
		SetPosition("FE").
		SetStatus(application.StatusPreparing).
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("카카오").
		SetPosition("BE").
		SetStatus(application.StatusSubmitted).
		SaveX(ctx)

	// App with upcoming deadline (2 days from now)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("토스").
		SetPosition("서버").
		SetStatus(application.StatusPreparing).
		SetDeadline(time.Now().Add(48 * time.Hour)).
		SaveX(ctx)

	stats, err := svc.GetStats(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 4, stats.Total)
	assert.Equal(t, 3, stats.ByStatus["preparing"])
	assert.Equal(t, 1, stats.ByStatus["submitted"])
	assert.Len(t, stats.UpcomingDeadlines, 1)
	assert.Equal(t, "토스", stats.UpcomingDeadlines[0].CompanyName)
}
