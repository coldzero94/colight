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

// --- System Configs ---

func setupConfigTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	r := gin.New()
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
	r.GET("/v1/admin/configs", ctrl.ListConfigs)
	r.PUT("/v1/admin/configs/:key", ctrl.UpdateConfig)

	return r, db
}

func TestAdminController_ListConfigs_ByCategory(t *testing.T) {
	r, db := setupConfigTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("config-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)

	db.SystemConfig.Create().
		SetConfigKey("anthropic_api_key").
		SetConfigValue("encrypted-value").
		SetCategory("api_key").
		SetIsSecret(true).
		SetUpdatedBy(admin.ID).
		SaveX(ctx)
	db.SystemConfig.Create().
		SetConfigKey("heavy_model").
		SetConfigValue("claude-sonnet-4-5").
		SetCategory("model").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/configs?category=api_key", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "super_admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	item := data[0].(map[string]any)
	assert.Equal(t, "anthropic_api_key", item["config_key"])
	// Secret values should be masked
	assert.NotEqual(t, "encrypted-value", item["config_value"])
}

func TestAdminController_UpdateConfig_Success(t *testing.T) {
	r, db := setupConfigTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("config-upd@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)

	db.SystemConfig.Create().
		SetConfigKey("heavy_model").
		SetConfigValue("claude-sonnet-4-5").
		SetCategory("model").
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/configs/heavy_model",
		map[string]string{"value": "gpt-4o"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "gpt-4o", resp["config_value"])
}

func TestAdminController_UpdateConfig_NotFound(t *testing.T) {
	r, db := setupConfigTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("config-nf@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/configs/nonexistent",
		map[string]string{"value": "something"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Usage Monitoring ---

func setupAdminUsageTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	r := gin.New()
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
	r.GET("/v1/admin/usage/summary", ctrl.GetUsageSummary)
	r.GET("/v1/admin/usage/daily", ctrl.GetUsageDaily)

	return r, db
}

func TestAdminController_GetUsageSummary(t *testing.T) {
	r, db := setupAdminUsageTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("usage-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	user := db.UserProfile.Create().
		SetEmail("usage-user@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// Create usage logs with AI metadata
	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetProvider("anthropic").
		SetModel("claude-sonnet-4-5").
		SetInputTokens(500).
		SetOutputTokens(200).
		SetTotalTokens(700).
		SetEstimatedCostKrw(14.0).
		SetLatencyMs(1500).
		SetStatus("success").
		SaveX(ctx)
	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("analysis").
		SetProvider("gemini").
		SetModel("gemini-2.0-flash").
		SetInputTokens(300).
		SetOutputTokens(100).
		SetTotalTokens(400).
		SetEstimatedCostKrw(2.0).
		SetLatencyMs(800).
		SetStatus("success").
		SaveX(ctx)
	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetProvider("anthropic").
		SetModel("claude-sonnet-4-5").
		SetInputTokens(600).
		SetOutputTokens(0).
		SetTotalTokens(600).
		SetStatus("error").
		SetErrorMessage("rate limit exceeded").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/summary?days=30", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, float64(3), resp["total_calls"])
	assert.Equal(t, float64(1700), resp["total_tokens"])
	assert.InDelta(t, 16.0, resp["total_cost_krw"], 0.01)
	assert.InDelta(t, 33.33, resp["error_rate"], 1.0) // 1/3 errors
}

func TestAdminController_GetUsageDaily(t *testing.T) {
	r, db := setupAdminUsageTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("daily-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	user := db.UserProfile.Create().
		SetEmail("daily-user@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// Create today's usage
	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetProvider("anthropic").
		SetModel("claude-sonnet-4-5").
		SetInputTokens(500).
		SetOutputTokens(200).
		SetTotalTokens(700).
		SetStatus("success").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/daily?days=7", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.NotEmpty(t, data)
}

// --- Suspend / Unsuspend ---

func setupSuspendTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	r := gin.New()
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
	r.POST("/v1/admin/users/:id/suspend", ctrl.SuspendUser)
	r.DELETE("/v1/admin/users/:id/suspend", ctrl.UnsuspendUser)
	r.GET("/v1/admin/users/:id/detail", ctrl.GetUserDetail)

	return r, db
}

func TestAdminController_SuspendUser_Success(t *testing.T) {
	r, db := setupSuspendTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("suspend-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("suspend-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPost, "/v1/admin/users/"+target.ID.String()+"/suspend",
		map[string]string{"reason": "규정 위반"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, true, resp["suspended"])
}

func TestAdminController_UnsuspendUser(t *testing.T) {
	r, db := setupSuspendTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("unsuspend-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("unsuspend-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SetSuspended(true).
		SetSuspendedReason("test").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/users/"+target.ID.String()+"/suspend", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, false, resp["suspended"])
}

func TestAdminController_GetUserDetail(t *testing.T) {
	r, db := setupSuspendTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("detail-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("detail-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// Add some experiences for the user
	db.Experience.Create().
		SetUserID(target.ID).
		SetTitle("Test Experience").
		SetContent("Test content").
		SetStarSituation("Situation").
		SetStarTask("Task").
		SetStarAction("Action").
		SetStarResult("Result").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/users/"+target.ID.String()+"/detail", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "detail-target@test.com", resp["email"])
	assert.Equal(t, float64(1), resp["experience_count"])
}

// --- Audit Logs ---

func setupAuditTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db)

	r := gin.New()
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
	r.GET("/v1/admin/audit-logs", ctrl.ListAuditLogs)

	return r, db
}

func TestAdminController_ListAuditLogs(t *testing.T) {
	r, db := setupAuditTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("audit-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	// Create audit log entries
	db.AdminAuditLog.Create().
		SetAdminID(admin.ID).
		SetAction("role_change").
		SetTargetType("user").
		SetTargetID("some-user-id").
		SetOldValue("user").
		SetNewValue("manager").
		SaveX(ctx)
	db.AdminAuditLog.Create().
		SetAdminID(admin.ID).
		SetAction("config_update").
		SetTargetType("config").
		SetTargetID("heavy_model").
		SetOldValue("claude-sonnet-4-5").
		SetNewValue("gpt-4o").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit-logs", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 2)
}

func TestAdminController_ListAuditLogs_FilterByAction(t *testing.T) {
	r, db := setupAuditTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("audit-filter@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	db.AdminAuditLog.Create().
		SetAdminID(admin.ID).
		SetAction("role_change").
		SetTargetType("user").
		SaveX(ctx)
	db.AdminAuditLog.Create().
		SetAdminID(admin.ID).
		SetAction("config_update").
		SetTargetType("config").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit-logs?action=role_change", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
}
