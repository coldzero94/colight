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

// === Phase 8.2: Manual CRUD Tests ===

func TestApplicationService_CreateApplication(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	deadline := time.Now().Add(7 * 24 * time.Hour)
	input := CreateApplicationInput{
		CompanyName: "카카오",
		Position:    "백엔드 개발자",
		JobURL:      "https://careers.kakao.com/123",
		Deadline:    &deadline,
		Notes:       "지원 준비 중",
		Tags:        []string{"관심", "백엔드"},
	}

	result, err := svc.CreateApplication(ctx, userID, input)
	require.NoError(t, err)
	assert.Equal(t, "카카오", result.CompanyName)
	assert.Equal(t, "백엔드 개발자", result.Position)
	assert.Equal(t, "preparing", result.Status)
	assert.Equal(t, "지원 준비 중", result.Notes)
	assert.Equal(t, []string{"관심", "백엔드"}, result.Tags)
	assert.NotNil(t, result.Deadline)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, 0, result.CoverLetterCount)
}

func TestApplicationService_CreateApplication_MinimalFields(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	input := CreateApplicationInput{
		CompanyName: "네이버",
	}

	result, err := svc.CreateApplication(ctx, userID, input)
	require.NoError(t, err)
	assert.Equal(t, "네이버", result.CompanyName)
	assert.Equal(t, "", result.Position)
	assert.Equal(t, "preparing", result.Status)
	assert.Nil(t, result.Deadline)
}

func TestApplicationService_CreateApplication_EmptyCompanyName(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	input := CreateApplicationInput{
		CompanyName: "",
	}

	_, err := svc.CreateApplication(ctx, userID, input)
	assert.Error(t, err)
}

func TestApplicationService_UpdateApplication(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("토스").
		SetPosition("서버 개발자").
		SaveX(ctx)

	newPosition := "시니어 서버 개발자"
	newNotes := "면접 준비"
	tags := []string{"우선"}
	input := UpdateApplicationInput{
		Position: &newPosition,
		Notes:    &newNotes,
		Tags:     &tags,
	}

	result, err := svc.UpdateApplication(ctx, userID, app.ID, input)
	require.NoError(t, err)
	assert.Equal(t, "토스", result.CompanyName) // unchanged
	assert.Equal(t, "시니어 서버 개발자", result.Position)
	assert.Equal(t, "면접 준비", result.Notes)
	assert.Equal(t, []string{"우선"}, result.Tags)
}

func TestApplicationService_UpdateApplication_Forbidden(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("토스").
		SetPosition("서버").
		SaveX(ctx)

	newName := "changed"
	_, err := svc.UpdateApplication(ctx, uuid.New(), app.ID, UpdateApplicationInput{CompanyName: &newName})
	assert.ErrorIs(t, err, ErrApplicationForbidden)
}

func TestApplicationService_UpdateApplication_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	newName := "changed"
	_, err := svc.UpdateApplication(ctx, userID, uuid.New(), UpdateApplicationInput{CompanyName: &newName})
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestApplicationService_DeleteApplication(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성").
		SetPosition("BE").
		SaveX(ctx)

	err := svc.DeleteApplication(ctx, userID, app.ID)
	require.NoError(t, err)

	// Verify deleted
	apps, _ := svc.ListApplications(ctx, userID)
	assert.Len(t, apps, 0)
}

func TestApplicationService_DeleteApplication_Forbidden(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성").
		SetPosition("BE").
		SaveX(ctx)

	err := svc.DeleteApplication(ctx, uuid.New(), app.ID)
	assert.ErrorIs(t, err, ErrApplicationForbidden)
}

func TestApplicationService_DeleteApplication_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	err := svc.DeleteApplication(ctx, userID, uuid.New())
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

// === Phase 8.2.3: Search + LinkAnalysis Tests ===

func TestApplicationService_SearchApplications_ByQuery(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SetPosition("백엔드 개발자").
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("네이버").
		SetPosition("프론트엔드 개발자").
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("카카오").
		SetPosition("서버 개발자").
		SaveX(ctx)

	// Search by company name
	results, err := svc.SearchApplications(ctx, userID, SearchApplicationsInput{Query: "삼성"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "삼성전자", results[0].CompanyName)

	// Search by position (case-insensitive)
	results, err = svc.SearchApplications(ctx, userID, SearchApplicationsInput{Query: "백엔드"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "삼성전자", results[0].CompanyName)

	// Search matching both company and position
	results, err = svc.SearchApplications(ctx, userID, SearchApplicationsInput{Query: "개발자"})
	require.NoError(t, err)
	assert.Len(t, results, 3) // All have "개발자" in position
}

func TestApplicationService_SearchApplications_ByTags(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SetTags([]string{"관심", "대기업"}).
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("토스").
		SetTags([]string{"관심", "핀테크"}).
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("네이버").
		SetTags([]string{"대기업"}).
		SaveX(ctx)

	// Filter by single tag
	results, err := svc.SearchApplications(ctx, userID, SearchApplicationsInput{Tags: []string{"핀테크"}})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "토스", results[0].CompanyName)

	// Filter by tag matching multiple apps
	results, err = svc.SearchApplications(ctx, userID, SearchApplicationsInput{Tags: []string{"관심"}})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestApplicationService_SearchApplications_ByDeadlineRange(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	now := time.Now()
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SetDeadline(now.Add(24 * time.Hour)). // 1 day from now
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("네이버").
		SetDeadline(now.Add(7 * 24 * time.Hour)). // 7 days from now
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("카카오"). // No deadline
		SaveX(ctx)

	from := now
	to := now.Add(3 * 24 * time.Hour)
	results, err := svc.SearchApplications(ctx, userID, SearchApplicationsInput{
		DeadlineFrom: &from,
		DeadlineTo:   &to,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "삼성전자", results[0].CompanyName)
}

func TestApplicationService_SearchApplications_Combined(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	now := time.Now()
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SetPosition("백엔드 개발자").
		SetTags([]string{"대기업"}).
		SetDeadline(now.Add(48 * time.Hour)).
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성SDS").
		SetPosition("풀스택 개발자").
		SetTags([]string{"대기업"}).
		SetDeadline(now.Add(10 * 24 * time.Hour)).
		SaveX(ctx)

	from := now
	to := now.Add(5 * 24 * time.Hour)
	results, err := svc.SearchApplications(ctx, userID, SearchApplicationsInput{
		Query:        "삼성",
		Tags:         []string{"대기업"},
		DeadlineFrom: &from,
		DeadlineTo:   &to,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "삼성전자", results[0].CompanyName)
}

func TestApplicationService_SearchApplications_EmptyFilters(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)
	svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("네이버").
		SaveX(ctx)

	// Empty filter = return all
	results, err := svc.SearchApplications(ctx, userID, SearchApplicationsInput{})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestApplicationService_LinkAnalysis(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	analysis := svc.db.CompanyAnalysis.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	result, err := svc.LinkAnalysis(ctx, userID, app.ID, analysis.ID)
	require.NoError(t, err)
	assert.NotNil(t, result.AnalysisID)
	assert.Equal(t, analysis.ID.String(), *result.AnalysisID)
}

func TestApplicationService_LinkAnalysis_AppNotFound(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	analysis := svc.db.CompanyAnalysis.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	_, err := svc.LinkAnalysis(ctx, userID, uuid.New(), analysis.ID)
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestApplicationService_LinkAnalysis_AppForbidden(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	analysis := svc.db.CompanyAnalysis.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	_, err := svc.LinkAnalysis(ctx, uuid.New(), app.ID, analysis.ID)
	assert.ErrorIs(t, err, ErrApplicationForbidden)
}

func TestApplicationService_LinkAnalysis_AnalysisNotFound(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	_, err := svc.LinkAnalysis(ctx, userID, app.ID, uuid.New())
	assert.ErrorIs(t, err, ErrAnalysisNotFound)
}

func TestApplicationService_LinkAnalysis_AnalysisForbidden(t *testing.T) {
	ctx := context.Background()
	svc, userID := createApplicationForTest(t, ctx)

	app := svc.db.Application.Create().
		SetUserID(userID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	// Create analysis owned by different user
	otherEmail := fmt.Sprintf("other-%s@example.com", uuid.New().String()[:8])
	otherUser := svc.db.UserProfile.Create().
		SetEmail(otherEmail).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	analysis := svc.db.CompanyAnalysis.Create().
		SetUserID(otherUser.ID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	_, err := svc.LinkAnalysis(ctx, userID, app.ID, analysis.ID)
	assert.ErrorIs(t, err, ErrAnalysisForbidden)
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
