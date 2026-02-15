package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	caller := db.UserProfile.Create().
		SetEmail("role-caller@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("role-ctrl@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/role",
		map[string]string{"role": "admin"},
		map[string]string{"X-Test-UserID": caller.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "admin", resp["role"])
}

func TestAdminController_UpdateUserRole_InvalidRole(t *testing.T) {
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	caller := db.UserProfile.Create().
		SetEmail("badrole-caller@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("badrole-ctrl@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/role",
		map[string]string{"role": "superuser"},
		map[string]string{"X-Test-UserID": caller.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- UpdateUserRole (role hierarchy validation) ---

func setupRoleValidationRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	r := gin.New()
	// Simulate auth middleware: set user_id + role from headers
	r.Use(func(c *gin.Context) {
		if id := c.GetHeader("X-Test-UserID"); id != "" {
			uid, _ := uuid.Parse(id)
			c.Set("user_id", uid)
		}
		if role := c.GetHeader("X-Test-Role"); role != "" {
			c.Set("role", role)
		}
		c.Next()
	})
	r.PUT("/v1/admin/users/:id/role", ctrl.UpdateUserRole)

	return r, db
}

func TestUpdateUserRole_AdminCanPromoteToManager(t *testing.T) {
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/role",
		map[string]string{"role": "manager"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "manager", resp["role"])
}

func TestUpdateUserRole_AdminCannotPromoteToAdmin(t *testing.T) {
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("admin2@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("target2@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/role",
		map[string]string{"role": "admin"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "admin"},
	)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUserRole_SuperAdminCanPromoteToAdmin(t *testing.T) {
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	superAdmin := db.UserProfile.Create().
		SetEmail("super@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("target3@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/role",
		map[string]string{"role": "admin"},
		map[string]string{"X-Test-UserID": superAdmin.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "admin", resp["role"])
}

func TestUpdateUserRole_CannotChangeSelf(t *testing.T) {
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("self@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+admin.ID.String()+"/role",
		map[string]string{"role": "user"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "admin"},
	)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "자기 자신")
}

func TestUpdateUserRole_LastSuperAdminProtected(t *testing.T) {
	r, db := setupRoleValidationRouter(t)
	ctx := t.Context()

	// Create exactly 1 super_admin (the target) — they are the last one
	target := db.UserProfile.Create().
		SetEmail("lastsuperadmin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	// Caller is a different super_admin — create and then we'll have 2, but demote caller via DB
	// so target is truly the last. But caller's header says super_admin for permission.
	// Actually: to properly test, we need 2 super_admins. Caller (super_admin) tries to demote target,
	// but target would be the last super_admin after demotion. The check counts current super_admins.
	// If target is the only one, count=1 → deny.
	// So: don't create another super_admin. Use a fake caller with super_admin header.
	caller := db.UserProfile.Create().
		SetEmail("caller-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/role",
		map[string]string{"role": "admin"},
		map[string]string{"X-Test-UserID": caller.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "마지막")
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
		SetModel("gemini-2.0-flash").SaveX(t.Context())

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
