package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAnalysisTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)

	// Seed talent profile for testing
	client.TalentProfile.Create().
		SetCompanyName("분석테스트회사").
		SetIndustry("IT").
		SetCoreValues([]map[string]string{{"keyword": "혁신", "description": "기술 혁신"}}).
		SetTalentTraits([]map[string]string{{"trait": "도전정신", "description": "도전"}}).
		SetCultureKeywords([]string{"혁신"}).
		SetVerified(true).
		SaveX(context.Background())

	mock := &mockLLMForMatch{
		response: ai.LLMResponse{
			Content: `{"core_values":[],"talent_traits":[],"recent_trends":[],"strategy_keywords":[],"avoid_expressions":[]}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mock, mock)
	companyDataSvc := service.NewCompanyDataService()
	analysisService := service.NewCompanyAnalysisService(client, aiProvider, companyDataSvc)
	ctrl := NewCompanyAnalysisController(analysisService)

	router := gin.New()
	router.POST("/v1/analyze-company", ctrl.AnalyzeCompany)

	return router
}

func TestAnalyzeCompany_Success(t *testing.T) {
	router := setupAnalysisTestRouter(t)

	body, _ := json.Marshal(map[string]string{"company_name": "분석테스트회사"})
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze-company", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "분석테스트회사", resp["company_name"])
	assert.Equal(t, "talent_profiles", resp["source"])
}

func TestAnalyzeCompany_MissingCompanyName(t *testing.T) {
	router := setupAnalysisTestRouter(t)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze-company", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "VALID_001", errObj["code"])
}

func TestAnalyzeCompany_EmptyCompanyName(t *testing.T) {
	router := setupAnalysisTestRouter(t)

	body, _ := json.Marshal(map[string]string{"company_name": ""})
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze-company", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeCompany_ReturnsCoreValues(t *testing.T) {
	router := setupAnalysisTestRouter(t)

	body, _ := json.Marshal(map[string]string{"company_name": "분석테스트회사"})
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze-company", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	coreValues := resp["core_values"].([]any)
	assert.Len(t, coreValues, 1)
	cv := coreValues[0].(map[string]any)
	assert.Equal(t, "혁신", cv["keyword"])
}

func TestAnalyzeCompany_InvalidJSON(t *testing.T) {
	router := setupAnalysisTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/analyze-company", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
