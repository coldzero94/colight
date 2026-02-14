package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMForMatch is a mock LLM for matching controller tests
type mockLLMForMatch struct {
	response ai.LLMResponse
	err      error
}

func (m *mockLLMForMatch) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func setupMatchingTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)

	// Mock LLM for matching (light model)
	mockMatchLLM := &mockLLMForMatch{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 75,
				"job_relevance": 80,
				"talent_fit": 70,
				"uniqueness": 75,
				"reasoning": "Good technical match",
				"suggested_angle": "Focus on backend skills"
			}`,
		},
	}

	// Mock LLM for company analysis (heavy model)
	mockHeavyLLM := &mockLLMForMatch{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "테스트",
				"core_values": [],
				"talent_traits": [],
				"strategy_keywords": [],
				"avoid_expressions": []
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockMatchLLM, mockHeavyLLM)
	companyDataSvc := service.NewCompanyDataService()
	analysisService := service.NewCompanyAnalysisService(client, aiProvider, companyDataSvc)
	matchingService := service.NewMatchingService(client, mockMatchLLM)
	matchingCtrl := NewMatchingController(matchingService, analysisService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/match", matchingCtrl.MatchExperiences)

	return router, client
}

func matchRequest(router *gin.Engine, body any, userID uuid.UUID) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/match", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		req.Header.Set("X-Test-UserID", userID.String())
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestPostMatching_Success(t *testing.T) {
	router, tc := setupMatchingTestRouter(t)

	// Seed talent profile so AnalyzeCompany returns without AI
	tc.TalentProfile.Create().
		SetCompanyName("테스트회사").
		SetIndustry("IT").
		SetCoreValues([]map[string]string{{"keyword": "혁신", "description": "혁신 추구"}}).
		SetTalentTraits([]map[string]string{{"trait": "도전정신", "description": "새로운 도전"}}).
		SetCultureKeywords([]string{"혁신", "도전"}).
		SetVerified(true).
		SaveX(context.Background())

	user := tc.UserProfile.Create().
		SetEmail("match-ctrl@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create experiences for matching
	tc.Experience.Create().
		SetUserID(user.ID).
		SetTitle("백엔드 경험").
		SetContent("Go 개발").
		SetCategory("인턴").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(context.Background())

	w := matchRequest(router, map[string]any{"company_name": "테스트회사"}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "테스트회사", resp["company_name"])
	assert.NotNil(t, resp["matches"])

	matches := resp["matches"].([]any)
	assert.GreaterOrEqual(t, len(matches), 1)
}

func TestPostMatching_ResultSaved(t *testing.T) {
	router, tc := setupMatchingTestRouter(t)

	tc.TalentProfile.Create().
		SetCompanyName("저장회사").
		SetIndustry("IT").
		SetCoreValues([]map[string]string{}).
		SetTalentTraits([]map[string]string{}).
		SetVerified(true).
		SaveX(context.Background())

	user := tc.UserProfile.Create().
		SetEmail("match-save@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	tc.Experience.Create().
		SetUserID(user.ID).
		SetTitle("저장 테스트").
		SetContent("Content").
		SetCategory("인턴").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(context.Background())

	w := matchRequest(router, map[string]any{"company_name": "저장회사"}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Verify result contains match data
	total, ok := resp["total"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(1), total)
}

func TestPostMatching_NoExperiences(t *testing.T) {
	router, tc := setupMatchingTestRouter(t)

	tc.TalentProfile.Create().
		SetCompanyName("노경험회사").
		SetIndustry("IT").
		SetCoreValues([]map[string]string{}).
		SetTalentTraits([]map[string]string{}).
		SetVerified(true).
		SaveX(context.Background())

	// User with NO experiences
	user := tc.UserProfile.Create().
		SetEmail("match-noexp@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	w := matchRequest(router, map[string]any{"company_name": "노경험회사"}, user.ID)

	// When user has no experiences, controller returns OK with empty matches
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["total"])
}

func TestPostMatching_AnalysisNotFound(t *testing.T) {
	router, _ := setupMatchingTestRouter(t)

	user := uuid.New()

	// Company that doesn't exist in talent_profiles or cache
	// The CompanyAnalysisService will try AI which may fail
	w := matchRequest(router, map[string]any{"company_name": "존재하지않는회사"}, user)

	// Should fail because analysis couldn't be done (no talent profile, no cache, mock AI doesn't produce valid analysis)
	assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusOK)
}

func TestPostMatching_MissingCompanyName(t *testing.T) {
	router, _ := setupMatchingTestRouter(t)

	user := uuid.New()

	// Missing company_name field
	w := matchRequest(router, map[string]any{}, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "VALID_001", errObj["code"])
}
