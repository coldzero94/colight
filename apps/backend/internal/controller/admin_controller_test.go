package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAdminTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestClient(t)

	// Clean all tables so tests that assert exact counts start fresh
	testutil.CleanAllTables(db)

	ctrl := NewAdminController(db)

	r := gin.New()
	v1 := r.Group("/v1/admin")
	v1.GET("/users", ctrl.ListUsers)
	v1.GET("/users/:id", ctrl.GetUser)
	v1.PUT("/users/:id/role", ctrl.UpdateUserRole)
	v1.GET("/stats", ctrl.GetStats)
	v1.GET("/prompts", ctrl.ListPrompts)
	v1.PUT("/prompts/:id", ctrl.UpdatePrompt)

	// Seed test data
	ctx := context.Background()
	db.UserProfile.Create().
		SetEmail("admin-ctrl@test.com").
		SetNickname("Admin").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	db.UserProfile.Create().
		SetEmail("user1-ctrl@test.com").
		SetNickname("User1").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)
	db.UserProfile.Create().
		SetNaverID("naver-user").
		SetNickname("NaverUser").
		SetAuthProvider(userprofile.AuthProviderNaver).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	return r
}

// --- ListUsers ---

func TestAdminController_ListUsers_Default(t *testing.T) {
	r := setupAdminTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 3)
	assert.Equal(t, float64(3), resp["total"])
}

func TestAdminController_ListUsers_FilterByRole(t *testing.T) {
	r := setupAdminTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users?role=admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, float64(1), resp["total"])
}

func TestAdminController_ListUsers_Search(t *testing.T) {
	r := setupAdminTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users?search=Naver", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	first := data[0].(map[string]any)
	assert.Equal(t, "NaverUser", first["nickname"])
}

func TestAdminController_ListUsers_Pagination(t *testing.T) {
	r := setupAdminTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users?limit=1&offset=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, float64(3), resp["total"])
}

// --- GetUser ---

func TestAdminController_GetUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.GET("/v1/admin/users/:id", ctrl.GetUser)

	user := db.UserProfile.Create().
		SetEmail("get-ctrl@test.com").
		SetNickname("GetUser").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(t.Context())

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users/"+user.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "get-ctrl@test.com", resp["email"])
}

func TestAdminController_GetUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.GET("/v1/admin/users/:id", ctrl.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users/00000000-0000-0000-0000-000000000099", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminController_GetUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.GET("/v1/admin/users/:id", ctrl.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- UpdateUserRole ---

func TestAdminController_UpdateUserRole_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.PUT("/v1/admin/users/:id/role", ctrl.UpdateUserRole)

	user := db.UserProfile.Create().
		SetEmail("role-ctrl@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(t.Context())

	w := putJSON(r, "/v1/admin/users/"+user.ID.String()+"/role", map[string]string{
		"role": "admin",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "admin", resp["role"])
}

func TestAdminController_UpdateUserRole_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.PUT("/v1/admin/users/:id/role", ctrl.UpdateUserRole)

	user := db.UserProfile.Create().
		SetEmail("badrole-ctrl@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(t.Context())

	w := putJSON(r, "/v1/admin/users/"+user.ID.String()+"/role", map[string]string{
		"role": "superuser",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- GetStats ---

func TestAdminController_GetStats(t *testing.T) {
	r := setupAdminTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, float64(3), resp["total_users"])
	assert.Equal(t, float64(2), resp["email_auth_users"])
	assert.Equal(t, float64(1), resp["naver_auth_users"])
	assert.Equal(t, float64(0), resp["total_experiences"])
}

// --- ListPrompts ---

func TestAdminController_ListPrompts_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.GET("/v1/admin/prompts", ctrl.ListPrompts)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/prompts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Empty(t, data)
}

func TestAdminController_ListPrompts_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	// Create test prompt
	db.PromptTemplate.Create().
		SetCategory("coaching").
		SetSubCategory("draft").
		SetName("Draft Prompt").
		SetSystemPrompt("System prompt").
		SetUserPromptTemplate("User {{variable}}").
		SetModel("claude-sonnet-4-5").
		SaveX(t.Context())

	r := gin.New()
	r.GET("/v1/admin/prompts", ctrl.ListPrompts)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/prompts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	require.Len(t, data, 1)
	prompt := data[0].(map[string]any)
	assert.Equal(t, "coaching", prompt["category"])
	assert.Equal(t, "Draft Prompt", prompt["name"])
}

func TestAdminController_ListPrompts_FilterByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	db.PromptTemplate.Create().
		SetCategory("coaching").SetSubCategory("draft").
		SetName("P1").SetSystemPrompt("sys").SetUserPromptTemplate("usr").
		SetModel("claude-sonnet-4-5").SaveX(t.Context())
	db.PromptTemplate.Create().
		SetCategory("analysis").SetSubCategory("company").
		SetName("P2").SetSystemPrompt("sys").SetUserPromptTemplate("usr").
		SetModel("gpt-4.1-mini").SaveX(t.Context())

	r := gin.New()
	r.GET("/v1/admin/prompts", ctrl.ListPrompts)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/prompts?category=coaching", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
}

// --- UpdatePrompt ---

func TestAdminController_UpdatePrompt_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	prompt := db.PromptTemplate.Create().
		SetCategory("coaching").SetSubCategory("draft").
		SetName("P1").SetSystemPrompt("old system").SetUserPromptTemplate("old user").
		SetModel("claude-sonnet-4-5").SaveX(t.Context())

	r := gin.New()
	r.PUT("/v1/admin/prompts/:id", ctrl.UpdatePrompt)

	temp := 0.5
	maxTk := 1000
	w := putJSON(r, "/v1/admin/prompts/"+prompt.ID.String(), map[string]any{
		"system_prompt": "new system prompt",
		"temperature":   temp,
		"max_tokens":    maxTk,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "new system prompt", resp["system_prompt"])
	assert.InDelta(t, 0.5, resp["temperature"], 0.01)
	assert.Equal(t, float64(1000), resp["max_tokens"])
}

func TestAdminController_UpdatePrompt_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db)

	r := gin.New()
	r.PUT("/v1/admin/prompts/:id", ctrl.UpdatePrompt)

	w := putJSON(r, "/v1/admin/prompts/00000000-0000-0000-0000-000000000099", map[string]any{
		"system_prompt": "new",
	})

	assert.Equal(t, http.StatusNotFound, w.Code)
}
