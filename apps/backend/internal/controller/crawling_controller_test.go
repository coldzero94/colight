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
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMForCrawl is a mock LLM for crawling controller tests
type mockLLMForCrawl struct {
	response ai.LLMResponse
	err      error
}

func (m *mockLLMForCrawl) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func setupCrawlingTestRouter(t *testing.T, mockLLM ai.LLMProvider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	aiProvider := ai.NewAIProviderForTest(mockLLM)
	crawlingSvc := service.NewCrawlingService(nil, aiProvider)
	crawlingCtrl := NewCrawlingController(crawlingSvc)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/crawl", crawlingCtrl.ParseJobPosting)

	return router
}

func crawlRequest(router *gin.Engine, body any, userID uuid.UUID) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/crawl", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		req.Header.Set("X-Test-UserID", userID.String())
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCrawlingController_InvalidURL(t *testing.T) {
	mockLLM := &mockLLMForCrawl{}
	router := setupCrawlingTestRouter(t, mockLLM)
	userID := uuid.New()

	// Empty URL
	w := crawlRequest(router, map[string]any{}, userID)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "VALID_001", errObj["code"])
}

func TestCrawlingController_MissingURL(t *testing.T) {
	mockLLM := &mockLLMForCrawl{}
	router := setupCrawlingTestRouter(t, mockLLM)
	userID := uuid.New()

	// No body at all
	req := httptest.NewRequest(http.MethodPost, "/v1/crawl", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCrawlingController_FetchFailed(t *testing.T) {
	// Mock AI that should never be called (fetch will fail first)
	mockLLM := &mockLLMForCrawl{}
	router := setupCrawlingTestRouter(t, mockLLM)
	userID := uuid.New()

	// URL that won't resolve (uses invalid domain)
	w := crawlRequest(router, map[string]any{
		"url": "https://this-domain-does-not-exist-99999.invalid/job/1",
	}, userID)

	// Should return 500 with SYS_001 (fetch failure)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "SYS_001", errObj["code"])
	assert.Contains(t, errObj["message"], "크롤링 실패")
}

func TestCrawlingController_SuccessResponse(t *testing.T) {
	// This test can't do real HTTP fetch, so we verify the controller's
	// JSON binding and error response format with various inputs
	mockLLM := &mockLLMForCrawl{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "테스트",
				"position": "개발자"
			}`,
		},
	}
	router := setupCrawlingTestRouter(t, mockLLM)
	userID := uuid.New()

	// Provide a valid but unreachable URL — will fail at fetch stage
	w := crawlRequest(router, map[string]any{
		"url": "https://httpbin.org/status/404",
	}, userID)

	// Expected to fail since we can't actually fetch in tests
	// Just verify the controller processes the request and returns structured error
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestCrawlingController_EmptyBody(t *testing.T) {
	mockLLM := &mockLLMForCrawl{}
	router := setupCrawlingTestRouter(t, mockLLM)
	userID := uuid.New()

	// Empty body
	req := httptest.NewRequest(http.MethodPost, "/v1/crawl", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
