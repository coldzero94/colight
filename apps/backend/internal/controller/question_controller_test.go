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

// mockLLMForQuestion is a mock LLM for question controller tests
type mockLLMForQuestion struct {
	response ai.LLMResponse
	err      error
}

func (m *mockLLMForQuestion) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func setupQuestionTestRouter(t *testing.T, mockAI *mockLLMForQuestion) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	questionService := service.NewQuestionService(client, ai.NewAIProviderForTest(mockAI))
	questionCtrl := NewQuestionController(questionService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/coaching/question-analysis", questionCtrl.PostQuestionAnalysis)

	return router, client
}

func questionAnalysisRequest(router *gin.Engine, body any, userID uuid.UUID) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/coaching/question-analysis", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		req.Header.Set("X-Test-UserID", userID.String())
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func ensureTestPrompt(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	exists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("question_analysis"),
		).
		Exist(ctx)

	if !exists {
		client.PromptTemplate.Create().
			SetCategory("coaching").
			SetSubCategory("question_analysis").
			SetName("Test Question Analysis").
			SetSystemPrompt("Test system").
			SetUserPromptTemplate("Test {{question_text}}").
			SetModel("groq/compound").
			SetTemperature(0.3).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func TestPostQuestionAnalysis_Unauthorized(t *testing.T) {
	mockAI := &mockLLMForQuestion{
		response: ai.LLMResponse{Content: `{}`},
	}

	router, _ := setupQuestionTestRouter(t, mockAI)

	w := questionAnalysisRequest(router, map[string]any{
		"application_id": uuid.New().String(),
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, uuid.Nil) // No user ID

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostQuestionAnalysis_InvalidApplicationId(t *testing.T) {
	mockAI := &mockLLMForQuestion{
		response: ai.LLMResponse{Content: `{}`},
	}

	router, client := setupQuestionTestRouter(t, mockAI)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("question-invalid@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	// Use non-existent application ID
	nonExistentAppID := uuid.New()

	w := questionAnalysisRequest(router, map[string]any{
		"application_id": nonExistentAppID.String(),
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPostQuestionAnalysis_AIFailure(t *testing.T) {
	mockAI := &mockLLMForQuestion{
		err: fmt.Errorf("AI service unavailable"),
	}

	router, client := setupQuestionTestRouter(t, mockAI)
	ctx := context.Background()
	ensureTestPrompt(t, client)

	user := client.UserProfile.Create().
		SetEmail("question-aifail@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("테스트회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	w := questionAnalysisRequest(router, map[string]any{
		"application_id": app.ID.String(),
		"question_text":  "팀 프로젝트에서 어려움을 극복한 경험",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"].(string), "AI")
}

func TestPostQuestionAnalysis_KeywordExtraction(t *testing.T) {
	mockAI := &mockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "표면 질문",
				"real_intents": [
					{"intent": "1", "why": "a"},
					{"intent": "2", "why": "b"},
					{"intent": "3", "why": "c"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "핵심"},
					"secondary": []
				},
				"writing_structure": {
					"total_chars": 800,
					"sections": [{"name": "전체", "char_ratio": 1.0, "char_count": 800, "guide": "가이드"}]
				},
				"key_keywords": ["데이터 기반", "개선", "성과"],
				"avoid_list": ["열심히"],
				"good_structure_example": "예시"
			}`,
		},
	}

	router, client := setupQuestionTestRouter(t, mockAI)
	ctx := context.Background()
	ensureTestPrompt(t, client)

	user := client.UserProfile.Create().
		SetEmail("question-keyword@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("테스트회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	w := questionAnalysisRequest(router, map[string]any{
		"application_id": app.ID.String(),
		"question_text":  "팀 프로젝트에서 어려움을 극복한 경험을 기술하세요.",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Verify keywords are extracted
	keywords := resp["key_keywords"].([]any)
	assert.GreaterOrEqual(t, len(keywords), 1)
	assert.Contains(t, keywords, "데이터 기반")
}
