package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createCoverLetterForTest(t *testing.T, ctx context.Context) (*EditorService, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := testutil.NewTestClient(t)
	svc := NewEditorService(client)

	email := fmt.Sprintf("editor-%s@example.com", uuid.New().String()[:8])
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	cl := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetQuestionText("테스트 문항입니다").
		SetCurrentContent("초기 내용").
		SaveX(ctx)

	return svc, user.ID, cl.ID
}

func TestGetCoverLetter_Success(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	cl, err := svc.GetCoverLetter(ctx, userID, clID)
	require.NoError(t, err)
	assert.Equal(t, clID, cl.ID)
	assert.Equal(t, "초기 내용", cl.CurrentContent)
}

func TestGetCoverLetter_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, userID, _ := createCoverLetterForTest(t, ctx)

	_, err := svc.GetCoverLetter(ctx, userID, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCoverLetter_Forbidden(t *testing.T) {
	ctx := context.Background()
	svc, _, clID := createCoverLetterForTest(t, ctx)

	_, err := svc.GetCoverLetter(ctx, uuid.New(), clID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateContent_Success(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	err := svc.UpdateContent(ctx, userID, clID, "수정된 내용입니다")
	require.NoError(t, err)

	// Verify content updated
	cl, err := svc.GetCoverLetter(ctx, userID, clID)
	require.NoError(t, err)
	assert.Equal(t, "수정된 내용입니다", cl.CurrentContent)
}

func TestUpdateContent_Forbidden(t *testing.T) {
	ctx := context.Background()
	svc, _, clID := createCoverLetterForTest(t, ctx)

	err := svc.UpdateContent(ctx, uuid.New(), clID, "bad")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCreateVersion_Success(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	v, err := svc.CreateVersion(ctx, userID, clID, "버전 1 내용", "")
	require.NoError(t, err)
	assert.Equal(t, 1, v.VersionNumber)
	assert.Equal(t, "버전 1 내용", v.Content)
	assert.Equal(t, 7, *v.CharCount) // len([]rune("버전 1 내용")) = 7
}

func TestCreateVersion_IncrementsVersionNumber(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	v1, err := svc.CreateVersion(ctx, userID, clID, "v1", "")
	require.NoError(t, err)
	assert.Equal(t, 1, v1.VersionNumber)

	v2, err := svc.CreateVersion(ctx, userID, clID, "v2", "")
	require.NoError(t, err)
	assert.Equal(t, 2, v2.VersionNumber)
}

func TestCreateVersion_KoreanCharCount(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	v, err := svc.CreateVersion(ctx, userID, clID, "한글 테스트입니다", "")
	require.NoError(t, err)
	expected := len([]rune("한글 테스트입니다"))
	assert.Equal(t, expected, *v.CharCount)
}

func TestCreateVersion_WithChangeSummary(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	v, err := svc.CreateVersion(ctx, userID, clID, "복원된 내용", "v1에서 복원")
	require.NoError(t, err)
	assert.Equal(t, 1, v.VersionNumber)
	assert.Equal(t, "복원된 내용", v.Content)
	assert.Equal(t, "v1에서 복원", v.ChangeSummary)
}

func TestGetVersions_Ordered(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	_, _ = svc.CreateVersion(ctx, userID, clID, "v1", "")
	_, _ = svc.CreateVersion(ctx, userID, clID, "v2", "")
	_, _ = svc.CreateVersion(ctx, userID, clID, "v3", "")

	versions, err := svc.GetVersions(ctx, userID, clID)
	require.NoError(t, err)
	assert.Len(t, versions, 3)
	// Ordered by version_number DESC
	assert.Equal(t, 3, versions[0].VersionNumber)
	assert.Equal(t, 1, versions[2].VersionNumber)
}

func TestGetVersions_Forbidden(t *testing.T) {
	ctx := context.Background()
	svc, _, clID := createCoverLetterForTest(t, ctx)

	_, err := svc.GetVersions(ctx, uuid.New(), clID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestGetVersions_Empty(t *testing.T) {
	ctx := context.Background()
	svc, userID, clID := createCoverLetterForTest(t, ctx)

	versions, err := svc.GetVersions(ctx, userID, clID)
	require.NoError(t, err)
	assert.Empty(t, versions)
}
