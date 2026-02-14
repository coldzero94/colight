package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupEditorTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	editorService := service.NewEditorService(client)
	editorCtrl := NewEditorController(editorService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.GET("/coaching/cover-letters/:id", editorCtrl.GetCoverLetter)
	v1.PATCH("/coaching/cover-letters/:id", editorCtrl.PatchCoverLetter)
	v1.POST("/coaching/cover-letters/:id/versions", editorCtrl.PostVersion)
	v1.GET("/coaching/cover-letters/:id/versions", editorCtrl.GetVersions)

	return router, client
}

func TestGetCoverLetter_LoadsLatestVersion(t *testing.T) {
	router, client := setupEditorTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("editor-ctrl-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("테스트회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplication(app).
		SetQuestionText("팀 프로젝트 경험을 기술하세요").
		SetCharLimit(800).
		SetCurrentContent("[상황]\n초안 내용").
		SaveX(ctx)

	// Create version
	client.CoverLetterVersion.Create().
		SetCoverLetter(coverLetter).
		SetVersionNumber(1).
		SetContent("[상황]\n초안 버전 1").
		SetCharCount(10).
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/coaching/cover-letters/"+coverLetter.ID.String(), nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "팀 프로젝트 경험을 기술하세요", resp["question_text"])
}

func TestGetCoverLetter_NotFound(t *testing.T) {
	router, client := setupEditorTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("editor-nf-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	nonExistentID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/coaching/cover-letters/"+nonExistentID.String(), nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCoverLetter_Forbidden(t *testing.T) {
	router, client := setupEditorTestRouter(t)
	ctx := context.Background()

	user1 := client.UserProfile.Create().
		SetEmail("editor-user1-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	user2 := client.UserProfile.Create().
		SetEmail("editor-user2-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user1.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter := client.CoverLetter.Create().
		SetUserID(user1.ID).
		SetApplication(app).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	// User2 tries to access user1's cover letter
	req := httptest.NewRequest(http.MethodGet, "/v1/coaching/cover-letters/"+coverLetter.ID.String(), nil)
	req.Header.Set("X-Test-UserID", user2.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPatchCoverLetter_UpdatesContent(t *testing.T) {
	router, client := setupEditorTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("patch-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplication(app).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	// PATCH request
	body := map[string]interface{}{"content": "수정된 내용"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/v1/coaching/cover-letters/"+coverLetter.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify content was updated
	updated, _ := client.CoverLetter.Get(ctx, coverLetter.ID)
	assert.Contains(t, updated.CurrentContent, "수정된 내용")
}

func TestPostVersion_CreatesNewVersion(t *testing.T) {
	router, client := setupEditorTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("version-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplication(app).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	// POST new version
	body := map[string]interface{}{"content": "버전 1 내용"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/coaching/cover-letters/"+coverLetter.ID.String()+"/versions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify version was created
	versions, _ := client.CoverLetterVersion.Query().All(ctx)
	assert.GreaterOrEqual(t, len(versions), 1)
}

func TestGetVersions_ReturnsList(t *testing.T) {
	router, client := setupEditorTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("verlist-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplication(app).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	// Create versions
	client.CoverLetterVersion.Create().
		SetCoverLetter(coverLetter).
		SetVersionNumber(1).
		SetContent("버전 1").
		SetCharCount(5).
		SaveX(ctx)

	client.CoverLetterVersion.Create().
		SetCoverLetter(coverLetter).
		SetVersionNumber(2).
		SetContent("버전 2").
		SetCharCount(5).
		SaveX(ctx)

	// GET versions
	req := httptest.NewRequest(http.MethodGet, "/v1/coaching/cover-letters/"+coverLetter.ID.String()+"/versions", nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	versions := resp["versions"].([]interface{})
	assert.GreaterOrEqual(t, len(versions), 2)
}
