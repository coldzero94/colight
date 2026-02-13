package service

import (
	"context"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestExperienceService(t *testing.T) (*ExperienceService, *AuthService) {
	t.Helper()
	db := testutil.NewTestClient(t)
	cfg := testutil.NewTestConfig()
	ts := NewTokenService(testutil.TestJWTSecret, testutil.TestAccessTokenTTL, testutil.TestRefreshTokenTTL)
	as := NewAuthService(cfg, db, ts)
	es := NewExperienceService(db)
	return es, as
}

func createTestUser(t *testing.T, as *AuthService) uuid.UUID {
	t.Helper()
	result, err := as.Signup(context.Background(), "exp-"+uuid.New().String()[:8]+"@test.com", "password123", "tester")
	require.NoError(t, err)
	return result.User.ID
}

// --- CreateExperience ---

func TestCreateExperience_Success(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	input := CreateExperienceInput{
		Title:         "인턴 경험",
		Category:      "인턴",
		Role:          "백엔드 개발",
		StarSituation: "스타트업에서 인턴",
		StarTask:      "API 개발",
		StarAction:    "Go로 REST API 구현",
		StarResult:    "3개월 내 5개 API 완성",
		Content:       "인턴 경험 내용",
	}

	exp, err := es.CreateExperience(ctx, userID, input)
	require.NoError(t, err)
	require.NotNil(t, exp)

	assert.Equal(t, userID, exp.UserID)
	assert.Equal(t, "인턴 경험", exp.Title)
	assert.Equal(t, "인턴", exp.Category)
	assert.Equal(t, "백엔드 개발", exp.Role)
	assert.Equal(t, "스타트업에서 인턴", exp.StarSituation)
	assert.Equal(t, "API 개발", exp.StarTask)
	assert.Equal(t, "Go로 REST API 구현", exp.StarAction)
	assert.Equal(t, "3개월 내 5개 API 완성", exp.StarResult)
	assert.Equal(t, "manual", exp.Source)
	assert.False(t, exp.IsArchived)
	assert.NotEqual(t, uuid.Nil, exp.ID)
}

func TestCreateExperience_InvalidTitle(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	// Empty title
	input := CreateExperienceInput{
		Title:   "",
		Content: "some content",
	}
	_, err := es.CreateExperience(ctx, userID, input)
	assert.ErrorIs(t, err, ErrInvalidTitle)

	// Title exceeding 200 chars
	longTitle := make([]byte, 201)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	input2 := CreateExperienceInput{
		Title:   string(longTitle),
		Content: "some content",
	}
	_, err = es.CreateExperience(ctx, userID, input2)
	assert.ErrorIs(t, err, ErrInvalidTitle)
}

func TestCreateExperience_WithPeriod(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

	input := CreateExperienceInput{
		Title:       "기간 경험",
		Content:     "content",
		PeriodStart: &start,
		PeriodEnd:   &end,
	}

	exp, err := es.CreateExperience(ctx, userID, input)
	require.NoError(t, err)
	assert.NotNil(t, exp.PeriodStart)
	assert.NotNil(t, exp.PeriodEnd)
}

func TestCreateExperience_InvalidPeriod(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	start := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	input := CreateExperienceInput{
		Title:       "역전 기간",
		Content:     "content",
		PeriodStart: &start,
		PeriodEnd:   &end,
	}

	_, err := es.CreateExperience(ctx, userID, input)
	assert.ErrorIs(t, err, ErrInvalidPeriod)
}

// --- GetExperiences ---

func TestGetExperiences_ReturnsOnlyOwnerData(t *testing.T) {
	es, as := newTestExperienceService(t)
	ctx := context.Background()

	user1 := createTestUser(t, as)
	user2 := createTestUser(t, as)

	// Create experiences for user1
	_, err := es.CreateExperience(ctx, user1, CreateExperienceInput{Title: "User1 Exp1", Content: "c1"})
	require.NoError(t, err)
	_, err = es.CreateExperience(ctx, user1, CreateExperienceInput{Title: "User1 Exp2", Content: "c2"})
	require.NoError(t, err)

	// Create experience for user2
	_, err = es.CreateExperience(ctx, user2, CreateExperienceInput{Title: "User2 Exp1", Content: "c3"})
	require.NoError(t, err)

	// Query user1's experiences
	exps, err := es.GetExperiences(ctx, user1, ExperienceListParams{})
	require.NoError(t, err)
	assert.Len(t, exps, 2)
	for _, exp := range exps {
		assert.Equal(t, user1, exp.UserID)
	}

	// Query user2's experiences
	exps2, err := es.GetExperiences(ctx, user2, ExperienceListParams{})
	require.NoError(t, err)
	assert.Len(t, exps2, 1)
	assert.Equal(t, "User2 Exp1", exps2[0].Title)
}

// --- GetExperience ---

func TestGetExperience_NotFound(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	_, err := es.GetExperience(ctx, userID, uuid.New())
	assert.ErrorIs(t, err, ErrExperienceNotFound)
}

func TestGetExperience_ForbiddenAccess(t *testing.T) {
	es, as := newTestExperienceService(t)
	ctx := context.Background()

	owner := createTestUser(t, as)
	other := createTestUser(t, as)

	exp, err := es.CreateExperience(ctx, owner, CreateExperienceInput{Title: "Owner Only", Content: "content"})
	require.NoError(t, err)

	_, err = es.GetExperience(ctx, other, exp.ID)
	assert.ErrorIs(t, err, ErrExperienceForbidden)
}

func TestGetExperience_Success(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	created, err := es.CreateExperience(ctx, userID, CreateExperienceInput{
		Title:   "Get Test",
		Content: "content",
	})
	require.NoError(t, err)

	got, err := es.GetExperience(ctx, userID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Get Test", got.Title)
}

// --- UpdateExperience ---

func TestUpdateExperience_Success(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	created, err := es.CreateExperience(ctx, userID, CreateExperienceInput{
		Title:   "Before",
		Content: "old content",
	})
	require.NoError(t, err)

	newTitle := "After"
	updated, err := es.UpdateExperience(ctx, userID, created.ID, UpdateExperienceInput{
		Title: &newTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, "After", updated.Title)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))
}

func TestUpdateExperience_ForbiddenAccess(t *testing.T) {
	es, as := newTestExperienceService(t)
	ctx := context.Background()

	owner := createTestUser(t, as)
	other := createTestUser(t, as)

	exp, err := es.CreateExperience(ctx, owner, CreateExperienceInput{Title: "Owner", Content: "c"})
	require.NoError(t, err)

	newTitle := "Hacked"
	_, err = es.UpdateExperience(ctx, other, exp.ID, UpdateExperienceInput{Title: &newTitle})
	assert.ErrorIs(t, err, ErrExperienceForbidden)
}

// --- DeleteExperience ---

func TestDeleteExperience_CascadeDelete(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	exp, err := es.CreateExperience(ctx, userID, CreateExperienceInput{
		Title:   "To Delete",
		Content: "content",
	})
	require.NoError(t, err)

	// Add a weapon tag to test cascade
	es.db.ExperienceWeapon.Create().
		SetExperience(exp).
		SetWeaponCode("W01").
		SetConfidence(0.9).
		SetIsPrimary(true).
		SaveX(ctx)

	err = es.DeleteExperience(ctx, userID, exp.ID)
	require.NoError(t, err)

	// Verify experience is gone
	_, err = es.GetExperience(ctx, userID, exp.ID)
	assert.ErrorIs(t, err, ErrExperienceNotFound)
}

func TestDeleteExperience_ForbiddenAccess(t *testing.T) {
	es, as := newTestExperienceService(t)
	ctx := context.Background()

	owner := createTestUser(t, as)
	other := createTestUser(t, as)

	exp, err := es.CreateExperience(ctx, owner, CreateExperienceInput{Title: "No Delete", Content: "c"})
	require.NoError(t, err)

	err = es.DeleteExperience(ctx, other, exp.ID)
	assert.ErrorIs(t, err, ErrExperienceForbidden)
}

// --- Sorting ---

func TestGetExperiences_SortByOldest(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	// Create users without shared email prefix
	db := testutil.NewTestClient(t)
	cfg := testutil.NewTestConfig()
	ts := NewTokenService(testutil.TestJWTSecret, testutil.TestAccessTokenTTL, testutil.TestRefreshTokenTTL)
	as2 := NewAuthService(cfg, db, ts)
	es2 := NewExperienceService(db)
	userID2 := createTestUser(t, as2)
	_ = userID2
	_ = es2

	_, err := es.CreateExperience(ctx, userID, CreateExperienceInput{Title: "First", Content: "c"})
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond) // Ensure ordering
	_, err = es.CreateExperience(ctx, userID, CreateExperienceInput{Title: "Second", Content: "c"})
	require.NoError(t, err)

	exps, err := es.GetExperiences(ctx, userID, ExperienceListParams{Sort: "oldest"})
	require.NoError(t, err)
	require.Len(t, exps, 2)
	assert.Equal(t, "First", exps[0].Title)
	assert.Equal(t, "Second", exps[1].Title)
}

// --- Category filter ---

func TestGetExperiences_FilterByCategory(t *testing.T) {
	es, as := newTestExperienceService(t)
	userID := createTestUser(t, as)
	ctx := context.Background()

	_, err := es.CreateExperience(ctx, userID, CreateExperienceInput{Title: "인턴 경험", Category: "인턴", Content: "c"})
	require.NoError(t, err)
	_, err = es.CreateExperience(ctx, userID, CreateExperienceInput{Title: "프로젝트 경험", Category: "프로젝트", Content: "c"})
	require.NoError(t, err)

	exps, err := es.GetExperiences(ctx, userID, ExperienceListParams{Category: "인턴"})
	require.NoError(t, err)
	assert.Len(t, exps, 1)
	assert.Equal(t, "인턴 경험", exps[0].Title)
}

// --- User who is not auth provider email (direct DB user) ---
func TestCreateExperience_DirectDBUser(t *testing.T) {
	db := testutil.NewTestClient(t)
	es := NewExperienceService(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("direct@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	exp, err := es.CreateExperience(ctx, user.ID, CreateExperienceInput{
		Title:   "Direct User Exp",
		Content: "content",
	})
	require.NoError(t, err)
	assert.Equal(t, user.ID, exp.UserID)
}
