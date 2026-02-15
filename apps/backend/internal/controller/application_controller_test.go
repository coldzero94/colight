package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupApplicationTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	appService := service.NewApplicationService(client)
	appCtrl := NewApplicationController(appService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.GET("/applications", appCtrl.ListApplications)
	v1.PATCH("/applications/:id/status", appCtrl.UpdateStatus)
	v1.GET("/applications/stats", appCtrl.GetStats)

	return router, client
}

func TestApplicationController_UpdateStatus_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("appctrl@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("삼성전자").
		SetPosition("백엔드 개발자").
		SetStatus(application.StatusPreparing).
		SaveX(ctx)

	body, _ := json.Marshal(map[string]string{"status": "submitted"})
	req := httptest.NewRequest(http.MethodPatch, "/v1/applications/"+app.ID.String()+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp service.ApplicationDetail
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "submitted", resp.Status)
	assert.Equal(t, "삼성전자", resp.CompanyName)
}

func TestApplicationController_UpdateStatus_Unauthorized(t *testing.T) {
	router, _ := setupApplicationTestRouter(t)

	body, _ := json.Marshal(map[string]string{"status": "submitted"})
	req := httptest.NewRequest(http.MethodPatch, "/v1/applications/"+uuid.New().String()+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestApplicationController_ListApplications_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("applist@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	deadline := time.Now().Add(48 * time.Hour)
	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("네이버").
		SetPosition("FE 개발자").
		SetStatus(application.StatusPreparing).
		SetDeadline(deadline).
		SaveX(ctx)

	client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplicationID(app.ID).
		SetQuestionText("문항").
		SetCurrentContent("내용").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/applications", nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Applications []service.ApplicationDetail `json:"applications"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Applications, 1)
	assert.Equal(t, "네이버", resp.Applications[0].CompanyName)
	assert.Equal(t, 1, resp.Applications[0].CoverLetterCount)
	assert.NotNil(t, resp.Applications[0].Deadline)
}

func TestApplicationController_GetStats_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("appstats@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("삼성").
		SetPosition("BE").
		SetStatus(application.StatusPreparing).
		SaveX(ctx)
	client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("LG").
		SetPosition("AI").
		SetStatus(application.StatusInterview).
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/applications/stats", nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp service.ApplicationStats
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 1, resp.ByStatus["preparing"])
	assert.Equal(t, 1, resp.ByStatus["interview"])
}
