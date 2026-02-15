package controller

import (
	"bytes"
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

func setupInterviewTestRouter(t *testing.T, mockResp string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mockAI := &ai.MockStreamingProvider{
		Response: ai.LLMResponse{Content: mockResp},
	}

	interviewSvc := service.NewInterviewService(mockAI)
	interviewCtrl := NewInterviewController(interviewSvc)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/interview/question", interviewCtrl.PostQuestion)

	return router
}

func TestPostQuestion_Unauthorized(t *testing.T) {
	router := setupInterviewTestRouter(t, `{"question":"test"}`)

	body := `{"stage":"warmup","messages":[]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/interview/question", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostQuestion_InvalidBody(t *testing.T) {
	router := setupInterviewTestRouter(t, `{"question":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/interview/question", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", uuid.New().String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostQuestion_Success(t *testing.T) {
	router := setupInterviewTestRouter(t, `{"question":"안녕하세요! 최근 어떤 활동을 하셨나요?"}`)

	body := `{"stage":"warmup","messages":[]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/interview/question", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", uuid.New().String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp service.GenerateQuestionResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "안녕하세요! 최근 어떤 활동을 하셨나요?", resp.Question)
	assert.Equal(t, service.StageWarmup, resp.Stage)
	assert.Equal(t, service.StageMemory, resp.NextStage)
	assert.False(t, resp.IsComplete)
}

func TestPostQuestion_OutcomeIsComplete(t *testing.T) {
	router := setupInterviewTestRouter(t, `{"question":"배운 점이 무엇인가요?"}`)

	body := `{"stage":"outcome","messages":[{"role":"assistant","content":"q"},{"role":"user","content":"a"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/interview/question", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", uuid.New().String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp service.GenerateQuestionResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.IsComplete)
	assert.Equal(t, service.InterviewStage(""), resp.NextStage)
}
