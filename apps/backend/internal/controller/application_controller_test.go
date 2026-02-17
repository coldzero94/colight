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
	v1.POST("/applications", appCtrl.CreateApplication)
	v1.PATCH("/applications/:id", appCtrl.UpdateApplication)
	v1.DELETE("/applications/:id", appCtrl.DeleteApplication)
	v1.PATCH("/applications/:id/status", appCtrl.UpdateStatus)
	v1.GET("/applications/stats", appCtrl.GetStats)
	v1.POST("/applications/:id/link-analysis", appCtrl.LinkAnalysis)
	v1.GET("/applications/search", appCtrl.SearchApplications)

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

// === Phase 8.2.5: CRUD Controller Tests ===

func TestApplicationController_CreateApplication_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("create-app@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	body, _ := json.Marshal(map[string]interface{}{
		"company_name": "카카오",
		"position":     "백엔드 개발자",
		"notes":        "지원 준비 중",
		"tags":         []string{"관심"},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/applications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp service.ApplicationDetail
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "카카오", resp.CompanyName)
	assert.Equal(t, "백엔드 개발자", resp.Position)
	assert.Equal(t, "preparing", resp.Status)
	assert.Equal(t, []string{"관심"}, resp.Tags)
}

func TestApplicationController_CreateApplication_MissingCompanyName(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("create-app2@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	body, _ := json.Marshal(map[string]interface{}{"position": "FE"})
	req := httptest.NewRequest(http.MethodPost, "/v1/applications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApplicationController_UpdateApplication_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("update-app@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("토스").
		SetPosition("서버 개발자").
		SaveX(ctx)

	body, _ := json.Marshal(map[string]interface{}{
		"position": "시니어 서버 개발자",
		"notes":    "면접 준비",
	})
	req := httptest.NewRequest(http.MethodPatch, "/v1/applications/"+app.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp service.ApplicationDetail
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "토스", resp.CompanyName)
	assert.Equal(t, "시니어 서버 개발자", resp.Position)
}

func TestApplicationController_DeleteApplication_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("delete-app@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("삼성").
		SetPosition("BE").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodDelete, "/v1/applications/"+app.ID.String(), nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestApplicationController_LinkAnalysis_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("link-app@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	analysis := client.CompanyAnalysis.Create().
		SetUserID(user.ID).
		SetCompanyName("삼성전자").
		SaveX(ctx)

	body, _ := json.Marshal(map[string]string{"analysis_id": analysis.ID.String()})
	req := httptest.NewRequest(http.MethodPost, "/v1/applications/"+app.ID.String()+"/link-analysis", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp service.ApplicationDetail
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.AnalysisID)
	assert.Equal(t, analysis.ID.String(), *resp.AnalysisID)
}

func TestApplicationController_SearchApplications_Success(t *testing.T) {
	router, client := setupApplicationTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("search-app@test.com").
		SetAuthProvider("email").
		SaveX(ctx)

	client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("삼성전자").
		SetPosition("백엔드").
		SaveX(ctx)
	client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("네이버").
		SetPosition("프론트엔드").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/applications/search?q=삼성", nil)
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
	assert.Equal(t, "삼성전자", resp.Applications[0].CompanyName)
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
