package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLLMForReview struct {
	response ai.LLMResponse
	err      error
}

func (m *mockLLMForReview) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

const testReviewJSON = `{
	"scores": {"specificity": 75, "job_fit": 80, "company_fit": 65, "authenticity": 85},
	"overall": 76,
	"per_dimension_feedback": [
		{"dimension": "specificity", "score": 75, "good": ["좋음"], "improve": ["개선"]}
	],
	"specific_suggestions": [
		{"original": "원문", "suggested": "수정", "reason": "이유"}
	]
}`

func setupReviewTestRouter(t *testing.T, mockAI *mockLLMForReview) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	reviewService := service.NewReviewService(client, ai.NewAIProviderForTest(mockAI, mockAI))
	reviewCtrl := NewReviewController(reviewService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/coaching/review", reviewCtrl.PostReview)

	return router, client
}

func ensureReviewPromptCtrl(t *testing.T, client *ent.Client) {
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
			SetName("Test Review").
			SetSystemPrompt("Review system prompt").
			SetUserPromptTemplate("Review {{content}} {{question_text}} {{char_limit}} {{company_context}}").
			SetModel("claude-sonnet-4.5").
			SetTemperature(0.3).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func createReviewTestData(t *testing.T, client *ent.Client) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	email := fmt.Sprintf("review-%s@example.com", uuid.New().String()[:8])
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("테스트회사").
		SetPosition("백엔드 개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	cl := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplicationID(app.ID).
		SetQuestionText("팀 프로젝트 경험을 서술하세요").
		SetCharLimit(800).
		SetCurrentContent("긴 테스트 자소서 내용입니다. 팀 프로젝트에서 백엔드 리드 역할을 맡아 API 설계를 주도하고 성과를 도출했습니다. 구체적으로 마이크로서비스 아키텍처를 도입하여 시스템 안정성을 크게 향상시켰습니다.").
		SaveX(ctx)

	client.CoverLetterVersion.Create().
		SetCoverLetter(cl).
		SetVersionNumber(1).
		SetContent("긴 테스트 자소서 내용입니다. 팀 프로젝트에서 백엔드 리드 역할을 맡아 API 설계를 주도하고 성과를 도출했습니다. 구체적으로 마이크로서비스 아키텍처를 도입하여 시스템 안정성을 크게 향상시켰습니다.").
		SetCharCount(40).
		SaveX(ctx)

	return user.ID, cl.ID
}

func reviewRequest(router *gin.Engine, body any, userID uuid.UUID) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/coaching/review", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		req.Header.Set("X-Test-UserID", userID.String())
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestPostReview_Unauthorized(t *testing.T) {
	mockAI := &mockLLMForReview{}
	router, _ := setupReviewTestRouter(t, mockAI)

	w := reviewRequest(router, map[string]any{
		"cover_letter_id": uuid.New().String(),
		"content":         "테스트 자소서 내용입니다. 팀 프로젝트에서 백엔드 리드 역할을 맡아 API 설계를 주도하고 성과를 도출했습니다. 구체적으로 마이크로서비스 아키텍처를 도입하여 시스템 안정성을 크게 향상시켰습니다.",
	}, uuid.Nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostReview_BadRequest(t *testing.T) {
	mockAI := &mockLLMForReview{}
	router, _ := setupReviewTestRouter(t, mockAI)

	// Missing content
	w := reviewRequest(router, map[string]any{
		"cover_letter_id": uuid.New().String(),
	}, uuid.New())

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Content too short
	w = reviewRequest(router, map[string]any{
		"cover_letter_id": uuid.New().String(),
		"content":         "짧음",
	}, uuid.New())

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostReview_NotFound(t *testing.T) {
	mockAI := &mockLLMForReview{}
	router, client := setupReviewTestRouter(t, mockAI)
	ensureReviewPromptCtrl(t, client)

	w := reviewRequest(router, map[string]any{
		"cover_letter_id": uuid.New().String(),
		"content":         "테스트 자소서 내용입니다. 팀 프로젝트에서 백엔드 리드 역할을 맡아 API 설계를 주도하고 성과를 도출했습니다. 구체적으로 마이크로서비스 아키텍처를 도입하여 시스템 안정성을 크게 향상시켰습니다.",
	}, uuid.New())

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPostReview_Success(t *testing.T) {
	mockAI := &mockLLMForReview{
		response: ai.LLMResponse{
			Content:      testReviewJSON,
			InputTokens:  500,
			OutputTokens: 800,
		},
	}

	router, client := setupReviewTestRouter(t, mockAI)
	ensureReviewPromptCtrl(t, client)
	userID, clID := createReviewTestData(t, client)

	w := reviewRequest(router, map[string]any{
		"cover_letter_id": clID.String(),
		"content":         "테스트 자소서 내용입니다. 팀 프로젝트에서 백엔드 리드 역할을 맡아 API 설계를 주도하고 성과를 도출했습니다. 구체적으로 마이크로서비스 아키텍처를 도입하여 시스템 안정성을 크게 향상시켰습니다.",
	}, userID)

	require.Equal(t, http.StatusOK, w.Code)

	var result service.ReviewResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, 75, result.Scores.Specificity)
	assert.Equal(t, 80, result.Scores.JobFit)
	assert.Equal(t, 65, result.Scores.CompanyFit)
	assert.Equal(t, 85, result.Scores.Authenticity)
	assert.Equal(t, 76, result.Overall)
	assert.Len(t, result.PerDimensionFeedback, 1)
	assert.Len(t, result.SpecificSuggestions, 1)
}

func TestPostReview_Forbidden(t *testing.T) {
	mockAI := &mockLLMForReview{
		response: ai.LLMResponse{
			Content: testReviewJSON,
		},
	}

	router, client := setupReviewTestRouter(t, mockAI)
	ensureReviewPromptCtrl(t, client)
	_, clID := createReviewTestData(t, client)

	// Different user
	otherUser := uuid.New()
	w := reviewRequest(router, map[string]any{
		"cover_letter_id": clID.String(),
		"content":         "테스트 자소서 내용입니다. 팀 프로젝트에서 백엔드 리드 역할을 맡아 API 설계를 주도하고 성과를 도출했습니다. 구체적으로 마이크로서비스 아키텍처를 도입하여 시스템 안정성을 크게 향상시켰습니다.",
	}, otherUser)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
