package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validCharCoachingJSON = `{
	"status": "over",
	"current_count": 850,
	"char_limit": 800,
	"diff": 50,
	"suggestions": [
		{
			"type": "trim",
			"section": "도입부",
			"original": "저는 대학교 시절부터 다양한 프로젝트를 수행하면서 많은 경험을 쌓아왔습니다.",
			"suggested": "대학 시절 다양한 프로젝트 경험을 쌓았습니다.",
			"reason": "불필요한 수식어를 제거하여 간결하게",
			"char_diff": -20
		}
	],
	"summary": "50자 초과입니다. 도입부의 불필요한 수식어를 정리하면 글자수 내로 조절할 수 있습니다."
}`

const validCharCoachingUnderJSON = `{
	"status": "under",
	"current_count": 500,
	"char_limit": 800,
	"diff": -300,
	"suggestions": [
		{
			"type": "expand",
			"section": "결과 부분",
			"original": "프로젝트가 성공했습니다.",
			"suggested": "프로젝트 완료 후 사용자 만족도가 30% 향상되었고, 팀 내 코드 리뷰 문화 정착에도 기여했습니다.",
			"reason": "구체적 수치와 후속 영향을 추가하여 설득력 강화",
			"char_diff": 40
		}
	],
	"summary": "300자 부족합니다. 결과 부분에 구체적 수치와 영향을 추가해보세요."
}`

func TestParseCharCoachingResponse(t *testing.T) {
	result, err := ParseCharCoachingResponse(validCharCoachingJSON)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "over", result.Status)
	assert.Equal(t, 850, result.CurrentCount)
	assert.Equal(t, 800, result.CharLimit)
	assert.Equal(t, 50, result.Diff)
	require.Len(t, result.Suggestions, 1)
	assert.Equal(t, "trim", result.Suggestions[0].Type)
	assert.Equal(t, -20, result.Suggestions[0].CharDiff)
	assert.NotEmpty(t, result.Summary)
}

func TestParseCharCoachingResponse_WithCodeFence(t *testing.T) {
	fenced := "```json\n" + validCharCoachingJSON + "\n```"
	result, err := ParseCharCoachingResponse(fenced)
	require.NoError(t, err)
	assert.Equal(t, "over", result.Status)
}

func TestParseCharCoachingResponse_InvalidJSON(t *testing.T) {
	_, err := ParseCharCoachingResponse("not json at all")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse char coaching JSON")
}

func TestParseCharCoachingResponse_InvalidStatus(t *testing.T) {
	badJSON := `{"status": "invalid", "current_count": 100, "char_limit": 800, "diff": -700, "suggestions": [], "summary": "test"}`
	_, err := ParseCharCoachingResponse(badJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid char coaching status")
}

func ensureCharCoachingPrompt(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	exists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("char_coaching"),
		).
		Exist(ctx)

	if !exists {
		client.PromptTemplate.Create().
			SetCategory("coaching").
			SetSubCategory("char_coaching").
			SetName("Character Count Coaching").
			SetSystemPrompt(`당신은 자기소개서 글자수 조절 전문가입니다.
글자수 제한에 맞춰 축약 또는 보강 제안을 해주세요.
반드시 JSON으로만 응답하세요.`).
			SetUserPromptTemplate(`자소서 내용:
{{content}}

현재 글자수: {{current_count}}자
글자수 제한: {{char_limit}}자
차이: {{diff}}자 (양수=초과, 음수=부족)
상태: {{status}}

위 자소서의 글자수를 조절하는 제안을 JSON으로 응답하세요.`).
			SetModel("groq/compound").
			SetTemperature(0.3).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func createTestCoverLetterWithLimit(t *testing.T, client *ent.Client, userID, appID uuid.UUID, charLimit int) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	cl := client.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(appID).
		SetQuestionText("팀 프로젝트에서 어려움을 극복한 경험을 서술하세요").
		SetCharLimit(charLimit).
		SetCurrentContent("테스트 자소서 내용입니다.").
		SaveX(ctx)

	client.CoverLetterVersion.Create().
		SetCoverLetter(cl).
		SetVersionNumber(1).
		SetContent("테스트 자소서 내용입니다.").
		SetCharCount(13).
		SaveX(ctx)

	return cl.ID
}

func TestCharCoaching_OverLimit(t *testing.T) {
	mockAI := &MockLLMForReview{
		response: ai.LLMResponse{
			Content:      validCharCoachingJSON,
			InputTokens:  400,
			OutputTokens: 600,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureCharCoachingPrompt(t, client)
	coverLetterID := createTestCoverLetterWithLimit(t, client, userID, appID, 800)

	// Content that is over the 800 char limit (using repeated Korean text)
	content := ""
	for i := 0; i < 100; i++ {
		content += "가나다라마바사아자차"
	}

	result, err := svc.CoachCharCount(ctx, userID, coverLetterID, content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Server-side computed fields override AI response
	assert.Equal(t, 1000, result.CurrentCount)
	assert.Equal(t, 800, result.CharLimit)
	assert.Equal(t, 200, result.Diff)
	assert.Equal(t, 1, mockAI.calls)
	require.Len(t, result.Suggestions, 1)
	assert.Equal(t, "trim", result.Suggestions[0].Type)
}

func TestCharCoaching_UnderLimit(t *testing.T) {
	mockAI := &MockLLMForReview{
		response: ai.LLMResponse{
			Content:      validCharCoachingUnderJSON,
			InputTokens:  400,
			OutputTokens: 600,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureCharCoachingPrompt(t, client)
	coverLetterID := createTestCoverLetterWithLimit(t, client, userID, appID, 800)

	// Short content under limit
	content := "이것은 짧은 자소서입니다. 팀 프로젝트에서 백엔드 개발을 담당했고 좋은 성과를 거두었습니다. 프로젝트가 성공했습니다."

	result, err := svc.CoachCharCount(ctx, userID, coverLetterID, content)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 800, result.CharLimit)
	assert.True(t, result.Diff < 0, "expected negative diff for under-limit content")
	assert.Equal(t, 1, mockAI.calls)
}

func TestCharCoaching_NoCharLimit(t *testing.T) {
	mockAI := &MockLLMForReview{}

	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)

	// Create cover letter without char limit (0)
	coverLetterID := createTestCoverLetterWithLimit(t, client, userID, appID, 0)

	result, err := svc.CoachCharCount(ctx, userID, coverLetterID, "어떤 내용이든 상관없는 충분히 긴 내용입니다. 이 자소서는 글자수 제한이 설정되어 있지 않습니다.")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "good", result.Status)
	assert.Equal(t, 0, result.CharLimit)
	assert.Equal(t, 0, mockAI.calls, "should not call AI when no char limit")
}

func TestCharCoaching_NotFound(t *testing.T) {
	mockAI := &MockLLMForReview{}
	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, _, _ := createTestCoachingData(t, client)

	_, err := svc.CoachCharCount(ctx, userID, uuid.New(), "content that is long enough to pass validation minimum of fifty characters here")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCoverLetterNotFound)
}

func TestCharCoaching_Forbidden(t *testing.T) {
	mockAI := &MockLLMForReview{}
	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureCharCoachingPrompt(t, client)
	coverLetterID := createTestCoverLetterWithLimit(t, client, userID, appID, 800)

	otherUserID := uuid.New()
	_, err := svc.CoachCharCount(ctx, otherUserID, coverLetterID, "content that is long enough to pass validation minimum of fifty characters here")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCoverLetterForbidden)
}

func TestCharCoaching_AIError(t *testing.T) {
	mockAI := &MockLLMForReview{
		err: fmt.Errorf("AI unavailable"),
	}

	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureCharCoachingPrompt(t, client)
	coverLetterID := createTestCoverLetterWithLimit(t, client, userID, appID, 800)

	_, err := svc.CoachCharCount(ctx, userID, coverLetterID, "이것은 테스트를 위한 충분히 긴 자소서 내용입니다. AI가 실패하는 경우를 테스트합니다.")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI char coaching call failed")
}

func TestCharCoaching_RecordsSession(t *testing.T) {
	mockAI := &MockLLMForReview{
		response: ai.LLMResponse{
			Content:      validCharCoachingJSON,
			InputTokens:  400,
			OutputTokens: 600,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCharCoachingService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureCharCoachingPrompt(t, client)
	coverLetterID := createTestCoverLetterWithLimit(t, client, userID, appID, 800)

	content := ""
	for i := 0; i < 100; i++ {
		content += "가나다라마바사아자차"
	}

	_, err := svc.CoachCharCount(ctx, userID, coverLetterID, content)
	require.NoError(t, err)

	// Verify coaching session was recorded
	sessions, err := client.CoverLetter.Query().
		QueryCoachingSessions().
		All(ctx)
	require.NoError(t, err)

	var found bool
	for _, s := range sessions {
		if s.UserID == userID && string(s.SessionType) == "char_coaching" {
			assert.Equal(t, 400, s.InputTokens)
			assert.Equal(t, 600, s.OutputTokens)
			found = true
			break
		}
	}
	assert.True(t, found, "expected to find a char_coaching session for user")
}
