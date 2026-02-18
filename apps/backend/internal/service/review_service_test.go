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

// MockLLMForReview mocks AI for review tests (implements LLMProvider).
type MockLLMForReview struct {
	response ai.LLMResponse
	err      error
	calls    int
	lastReq  ai.LLMRequest
}

func (m *MockLLMForReview) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	m.lastReq = req
	return m.response, m.err
}

func newTestAIProviderForReview(mockAI *MockLLMForReview) *ai.AIProvider {
	return ai.NewAIProviderForTest(mockAI)
}

const validReviewJSON = `{
	"scores": {
		"specificity": 75,
		"job_fit": 80,
		"company_fit": 65,
		"authenticity": 85
	},
	"overall": 76,
	"per_dimension_feedback": [
		{
			"dimension": "specificity",
			"score": 75,
			"good": ["구체적인 수치 활용이 좋습니다"],
			"improve": ["행동의 구체적 과정을 더 서술하세요"]
		}
	],
	"specific_suggestions": [
		{
			"original": "팀 프로젝트를 했습니다",
			"suggested": "5명의 팀에서 백엔드 리드를 맡아 API 설계를 주도했습니다",
			"reason": "구체적 역할과 규모를 명시하면 신뢰도가 높아집니다"
		}
	]
}`

func TestParseReviewResponse(t *testing.T) {
	result, err := ParseReviewResponse(validReviewJSON)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 75, result.Scores.Specificity)
	assert.Equal(t, 80, result.Scores.JobFit)
	assert.Equal(t, 65, result.Scores.CompanyFit)
	assert.Equal(t, 85, result.Scores.Authenticity)
	assert.Equal(t, 76, result.Overall)

	require.Len(t, result.PerDimensionFeedback, 1)
	assert.Equal(t, "specificity", result.PerDimensionFeedback[0].Dimension)
	assert.Len(t, result.PerDimensionFeedback[0].Good, 1)
	assert.Len(t, result.PerDimensionFeedback[0].Improve, 1)

	require.Len(t, result.SpecificSuggestions, 1)
	assert.Equal(t, "팀 프로젝트를 했습니다", result.SpecificSuggestions[0].Original)
}

func TestParseReviewResponse_WithCodeFence(t *testing.T) {
	fenced := "```json\n" + validReviewJSON + "\n```"
	result, err := ParseReviewResponse(fenced)
	require.NoError(t, err)
	assert.Equal(t, 76, result.Overall)
}

func TestParseReviewResponse_InvalidJSON(t *testing.T) {
	_, err := ParseReviewResponse("not json at all")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse review JSON")
}

func TestParseReviewResponse_ScoreOutOfRange(t *testing.T) {
	badJSON := `{
		"scores": {"specificity": 150, "job_fit": 80, "company_fit": 65, "authenticity": 85},
		"overall": 76,
		"per_dimension_feedback": [],
		"specific_suggestions": []
	}`
	_, err := ParseReviewResponse(badJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "specificity")
}

func TestParseReviewResponse_NegativeScore(t *testing.T) {
	badJSON := `{
		"scores": {"specificity": 75, "job_fit": -10, "company_fit": 65, "authenticity": 85},
		"overall": 76,
		"per_dimension_feedback": [],
		"specific_suggestions": []
	}`
	_, err := ParseReviewResponse(badJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job_fit")
}

func ensureReviewPrompt(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	exists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("review"),
		).
		Exist(ctx)

	if !exists {
		client.PromptTemplate.Create().
			SetCategory("coaching").
			SetSubCategory("review").
			SetName("Cover Letter Review").
			SetSystemPrompt(`당신은 자기소개서 첨삭 전문가입니다.
4가지 차원으로 평가하세요: specificity, job_fit, company_fit, authenticity.
반드시 JSON으로만 응답하세요.`).
			SetUserPromptTemplate(`자소서 내용:
{{content}}

문항: {{question_text}}
글자수 제한: {{char_limit}}자
{{company_context}}

위 자소서를 평가하고 JSON으로 응답하세요.`).
			SetModel("groq/compound").
			SetTemperature(0.3).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func createTestCoverLetter(t *testing.T, client *ent.Client, userID, appID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	cl := client.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(appID).
		SetQuestionText("팀 프로젝트에서 어려움을 극복한 경험을 서술하세요").
		SetCharLimit(800).
		SetCurrentContent("테스트 자소서 내용입니다.").
		SaveX(ctx)

	// Create version 1
	client.CoverLetterVersion.Create().
		SetCoverLetter(cl).
		SetVersionNumber(1).
		SetContent("테스트 자소서 내용입니다.").
		SetCharCount(13).
		SaveX(ctx)

	return cl.ID
}

func TestReviewCoverLetter_Success(t *testing.T) {
	mockAI := &MockLLMForReview{
		response: ai.LLMResponse{
			Content:      validReviewJSON,
			InputTokens:  500,
			OutputTokens: 800,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewReviewService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureReviewPrompt(t, client)
	coverLetterID := createTestCoverLetter(t, client, userID, appID)

	result, err := svc.ReviewCoverLetter(ctx, userID, coverLetterID, "테스트 자소서 내용입니다.")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 75, result.Scores.Specificity)
	assert.Equal(t, 76, result.Overall)
	assert.Equal(t, 1, mockAI.calls)

	// Verify feedback was saved to version
	versions, err := client.CoverLetterVersion.Query().All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, versions)

	latestVersion := versions[len(versions)-1]
	assert.NotNil(t, latestVersion.Feedback)
	assert.NotNil(t, latestVersion.Scores)
}

func TestReviewCoverLetter_NotFound(t *testing.T) {
	mockAI := &MockLLMForReview{}
	client := testutil.NewTestClient(t)
	svc := NewReviewService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, _, _ := createTestCoachingData(t, client)

	_, err := svc.ReviewCoverLetter(ctx, userID, uuid.New(), "content")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCoverLetterNotFound)
}

func TestReviewCoverLetter_Forbidden(t *testing.T) {
	mockAI := &MockLLMForReview{}
	client := testutil.NewTestClient(t)
	svc := NewReviewService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureReviewPrompt(t, client)
	coverLetterID := createTestCoverLetter(t, client, userID, appID)

	otherUserID := uuid.New()
	_, err := svc.ReviewCoverLetter(ctx, otherUserID, coverLetterID, "content")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCoverLetterForbidden)
}

func TestReviewCoverLetter_AIError(t *testing.T) {
	mockAI := &MockLLMForReview{
		err: fmt.Errorf("AI unavailable"),
	}

	client := testutil.NewTestClient(t)
	svc := NewReviewService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureReviewPrompt(t, client)
	coverLetterID := createTestCoverLetter(t, client, userID, appID)

	_, err := svc.ReviewCoverLetter(ctx, userID, coverLetterID, "content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI review call failed")
}

func TestReviewCoverLetter_RecordsSession(t *testing.T) {
	mockAI := &MockLLMForReview{
		response: ai.LLMResponse{
			Content:      validReviewJSON,
			InputTokens:  500,
			OutputTokens: 800,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewReviewService(client, newTestAIProviderForReview(mockAI))
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)
	ensureReviewPrompt(t, client)
	coverLetterID := createTestCoverLetter(t, client, userID, appID)

	_, err := svc.ReviewCoverLetter(ctx, userID, coverLetterID, "테스트 자소서 내용입니다.")
	require.NoError(t, err)

	// Verify coaching session was recorded for this cover letter
	sessions, err := client.CoverLetter.Query().
		QueryCoachingSessions().
		All(ctx)
	require.NoError(t, err)

	// Find sessions belonging to our cover letter
	var found bool
	for _, s := range sessions {
		if s.UserID == userID {
			assert.Equal(t, "review", string(s.SessionType))
			assert.Equal(t, 500, s.InputTokens)
			assert.Equal(t, 800, s.OutputTokens)
			found = true
			break
		}
	}
	assert.True(t, found, "expected to find a review session for user")
}
