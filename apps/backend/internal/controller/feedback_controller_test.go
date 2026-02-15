package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFeedbackTestRouter(t *testing.T) (*gin.Engine, *service.FeedbackService) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	feedbackSvc := service.NewFeedbackService(client)
	feedbackCtrl := NewFeedbackController(feedbackSvc)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/feedback", feedbackCtrl.SubmitFeedback)

	return router, feedbackSvc
}

func TestSubmitFeedback_Unauthorized(t *testing.T) {
	router, _ := setupFeedbackTestRouter(t)

	body := `{"category":"bug","content":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubmitFeedback_InvalidBody(t *testing.T) {
	router, _ := setupFeedbackTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", uuid.New().String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitFeedback_Success(t *testing.T) {
	router, feedbackSvc := setupFeedbackTestRouter(t)
	ctx := context.Background()

	// Access the db via the service's unexported field using a helper
	client := testutil.NewTestClient(t)
	user := client.UserProfile.Create().
		SetEmail("fb-ctrl@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	_ = feedbackSvc // used to set up router

	body := `{"category":"improvement","content":"Please add dark mode","page_url":"/settings"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.CreatedAt)
}
