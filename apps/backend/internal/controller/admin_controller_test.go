package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/feedback"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
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

	ctrl := NewAdminController(db, nil)

	r := gin.New()
	v1 := r.Group("/v1/admin")
	v1.GET("/users", ctrl.ListUsers)
	v1.GET("/users/:id", ctrl.GetUser)
	v1.PUT("/users/:id/role", ctrl.UpdateUserRole)
	v1.GET("/stats", ctrl.GetStats)
	v1.POST("/users", ctrl.CreateUser)
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
	ctrl := NewAdminController(db, nil)

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
	ctrl := NewAdminController(db, nil)

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
	ctrl := NewAdminController(db, nil)

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
	ctrl := NewAdminController(db, nil)

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
	ctrl := NewAdminController(db, nil)

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
	ctrl := NewAdminController(db, nil)

	// Create test prompt
	db.PromptTemplate.Create().
		SetCategory("coaching").
		SetSubCategory("draft").
		SetName("Draft Prompt").
		SetSystemPrompt("System prompt").
		SetUserPromptTemplate("User {{variable}}").
		SetModel("groq/compound").
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
	ctrl := NewAdminController(db, nil)

	db.PromptTemplate.Create().
		SetCategory("coaching").SetSubCategory("draft").
		SetName("P1").SetSystemPrompt("sys").SetUserPromptTemplate("usr").
		SetModel("groq/compound").SaveX(t.Context())
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
	ctrl := NewAdminController(db, nil)

	prompt := db.PromptTemplate.Create().
		SetCategory("coaching").SetSubCategory("draft").
		SetName("P1").SetSystemPrompt("old system").SetUserPromptTemplate("old user").
		SetModel("groq/compound").SaveX(t.Context())

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
	ctrl := NewAdminController(db, nil)

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
	ctrl := NewAdminController(db, nil)

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
		SetConfigKey("gemini_api_key").
		SetConfigValue("encrypted-value").
		SetCategory("api_key").
		SetIsSecret(true).
		SetUpdatedBy(admin.ID).
		SaveX(ctx)
	db.SystemConfig.Create().
		SetConfigKey("heavy_model").
		SetConfigValue("groq/compound").
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
	assert.Equal(t, "gemini_api_key", item["config_key"])
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
		SetConfigValue("groq/compound").
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
	ctrl := NewAdminController(db, nil)

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
	r.GET("/v1/admin/usage/costs", ctrl.GetUsageCosts)
	r.GET("/v1/admin/usage/top-users", ctrl.GetUsageTopUsers)

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
		SetProvider("groq").
		SetModel("groq/compound").
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
		SetProvider("groq").
		SetModel("groq/compound").
		SetInputTokens(600).
		SetOutputTokens(0).
		SetTotalTokens(600).
		SetStatus("error").
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
		SetProvider("groq").
		SetModel("groq/compound").
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
	ctrl := NewAdminController(db, nil)

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
	r.POST("/v1/admin/users/:id/force-logout", ctrl.ForceLogout)
	r.PUT("/v1/admin/users/:id/plan", ctrl.UpdateUserPlan)
	r.POST("/v1/admin/users/:id/export", ctrl.ExportUserData)

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
	ctrl := NewAdminController(db, nil)

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
		SetOldValue("groq/compound").
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

// --- Health Check ---

func TestAdminController_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	ctrl := NewAdminController(db, nil)

	r := gin.New()
	r.GET("/v1/admin/health", ctrl.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "healthy", resp["database"])
}

// --- Feedback Management ---

func TestAdminController_ListFeedbacks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db, nil)

	r := gin.New()
	r.GET("/v1/admin/feedbacks", ctrl.ListFeedbacks)

	ctx := t.Context()
	user := db.UserProfile.Create().
		SetEmail("fb-user@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)
	db.Feedback.Create().
		SetUserID(user.ID).
		SetCategory(feedback.CategoryBug).
		SetContent("Something broke").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/feedbacks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	fb := data[0].(map[string]any)
	assert.Equal(t, "bug", fb["category"])
}

// --- Usage Costs ---

func TestAdminController_GetUsageCosts(t *testing.T) {
	r, db := setupAdminUsageTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("costs-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	user := db.UserProfile.Create().
		SetEmail("costs-user@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	db.UsageLog.Create().
		SetUserID(user.ID).SetFeature("draft").
		SetProvider("groq").SetModel("groq/compound").
		SetInputTokens(500).SetOutputTokens(200).SetTotalTokens(700).
		SetEstimatedCostKrw(14.0).SetStatus("success").SaveX(ctx)
	db.UsageLog.Create().
		SetUserID(user.ID).SetFeature("analysis").
		SetProvider("gemini").SetModel("gemini-2.0-flash").
		SetInputTokens(300).SetOutputTokens(100).SetTotalTokens(400).
		SetEstimatedCostKrw(2.0).SetStatus("success").SaveX(ctx)
	db.UsageLog.Create().
		SetUserID(user.ID).SetFeature("review").
		SetProvider("groq").SetModel("groq/compound").
		SetInputTokens(600).SetOutputTokens(300).SetTotalTokens(900).
		SetEstimatedCostKrw(18.0).SetStatus("success").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/costs?days=30", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.NotEmpty(t, data)
	// Should have 2 providers: groq and gemini
	assert.Len(t, data, 2)
}

// --- Usage Top Users ---

func TestAdminController_GetUsageTopUsers(t *testing.T) {
	r, db := setupAdminUsageTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("top-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)

	user1 := db.UserProfile.Create().
		SetEmail("top-user1@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)
	user2 := db.UserProfile.Create().
		SetEmail("top-user2@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// user1: 1000 tokens
	db.UsageLog.Create().
		SetUserID(user1.ID).SetFeature("draft").
		SetProvider("groq").SetModel("groq/compound").
		SetInputTokens(600).SetOutputTokens(400).SetTotalTokens(1000).
		SetEstimatedCostKrw(20.0).SetStatus("success").SaveX(ctx)
	// user2: 500 tokens
	db.UsageLog.Create().
		SetUserID(user2.ID).SetFeature("draft").
		SetProvider("gemini").SetModel("gemini-2.0-flash").
		SetInputTokens(300).SetOutputTokens(200).SetTotalTokens(500).
		SetEstimatedCostKrw(5.0).SetStatus("success").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/top-users?days=30", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	require.Len(t, data, 2)
	// First user should have more tokens
	first := data[0].(map[string]any)
	assert.Equal(t, float64(1000), first["total_tokens"])
}

// --- Force Logout ---

func TestAdminController_ForceLogout(t *testing.T) {
	r, db := setupSuspendTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("logout-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("logout-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/"+target.ID.String()+"/force-logout", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.NotNil(t, resp["force_logout_at"])

	// Verify DB field was set
	updated, _ := db.UserProfile.Get(ctx, target.ID)
	assert.NotNil(t, updated.ForceLogoutAt)
}

// --- Update User Plan ---

func TestAdminController_UpdateUserPlan(t *testing.T) {
	r, db := setupSuspendTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("plan-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("plan-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/users/"+target.ID.String()+"/plan",
		map[string]string{"plan": "pro"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "pro", resp["plan"])

	// Verify DB
	updated, _ := db.UserProfile.Get(ctx, target.ID)
	assert.Equal(t, "pro", string(updated.Plan))
}

// --- Update Feedback Status ---

func TestAdminController_UpdateFeedbackStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db, nil)

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
	r.PUT("/v1/admin/feedbacks/:id", ctrl.UpdateFeedbackStatus)

	ctx := t.Context()
	admin := db.UserProfile.Create().
		SetEmail("fb-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	user := db.UserProfile.Create().
		SetEmail("fb-user2@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)
	fb := db.Feedback.Create().
		SetUserID(user.ID).
		SetCategory(feedback.CategoryBug).
		SetContent("Something broke").
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPut, "/v1/admin/feedbacks/"+fb.ID.String(),
		map[string]any{"admin_status": "resolved", "admin_note": "Fixed in v2.1"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "resolved", resp["admin_status"])
	assert.Equal(t, "Fixed in v2.1", resp["admin_note"])
}

// --- Export User Data ---

func TestAdminController_ExportUserData(t *testing.T) {
	r, db := setupSuspendTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("export-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("export-target@test.com").
		SetNickname("ExportUser").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// Add experience
	db.Experience.Create().
		SetUserID(target.ID).
		SetTitle("Test Experience").
		SetContent("Test content").
		SetStarSituation("Situation").
		SetStarTask("Task").
		SetStarAction("Action").
		SetStarResult("Result").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/"+target.ID.String()+"/export", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.NotNil(t, resp["profile"])
	assert.NotNil(t, resp["experiences"])
	exps := resp["experiences"].([]any)
	assert.Len(t, exps, 1)
}

// --- Deletion Request ---

func setupDeletionTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db, nil)

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
	r.POST("/v1/admin/users/:id/delete-request", ctrl.CreateDeletionRequest)
	r.DELETE("/v1/admin/users/:id/delete-request", ctrl.CancelDeletionRequest)
	r.GET("/v1/admin/deletion-queue", ctrl.ListDeletionQueue)

	return r, db
}

func TestAdminController_CreateDeletionRequest(t *testing.T) {
	r, db := setupDeletionTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("del-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("del-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	w := sendJSONWithHeaders(r, http.MethodPost, "/v1/admin/users/"+target.ID.String()+"/delete-request",
		map[string]string{"reason": "PIPA 요청"},
		map[string]string{"X-Test-UserID": admin.ID.String(), "X-Test-Role": "super_admin"},
	)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "pending", resp["status"])
	assert.NotNil(t, resp["scheduled_at"])
}

func TestAdminController_CancelDeletionRequest(t *testing.T) {
	r, db := setupDeletionTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("cancel-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("cancel-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// Create a pending deletion request
	db.DeletionRequest.Create().
		SetUserID(target.ID).
		SetRequestedBy(admin.ID).
		SetReason("test").
		SetScheduledAt(time.Now().AddDate(0, 0, 30)).
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/users/"+target.ID.String()+"/delete-request", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "super_admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "cancelled", resp["status"])
}

func TestAdminController_ListDeletionQueue(t *testing.T) {
	r, db := setupDeletionTestRouter(t)
	ctx := t.Context()

	admin := db.UserProfile.Create().
		SetEmail("queue-admin@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleSuperAdmin).
		SaveX(ctx)
	target := db.UserProfile.Create().
		SetEmail("queue-target@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	db.DeletionRequest.Create().
		SetUserID(target.ID).
		SetRequestedBy(admin.ID).
		SetReason("PIPA").
		SetScheduledAt(time.Now().AddDate(0, 0, 30)).
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/deletion-queue", nil)
	req.Header.Set("X-Test-UserID", admin.ID.String())
	req.Header.Set("X-Test-Role", "super_admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	item := data[0].(map[string]any)
	assert.Equal(t, "pending", item["status"])
}

// --- CreateUser ---

func TestAdminController_CreateUser_Success(t *testing.T) {
	r := setupAdminTestRouter(t)

	body := `{"email":"newuser@test.com","password":"securePass123!","nickname":"NewUser","role":"user","plan":"free"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "newuser@test.com", resp["email"])
	assert.Equal(t, "NewUser", resp["nickname"])
	assert.Equal(t, "user", resp["role"])
}

func TestAdminController_CreateUser_InvalidEmail(t *testing.T) {
	r := setupAdminTestRouter(t)

	body := `{"email":"notanemail","password":"securePass123!","nickname":"Test","role":"user"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminController_CreateUser_WeakPassword(t *testing.T) {
	r := setupAdminTestRouter(t)

	body := `{"email":"test@test.com","password":"weak","nickname":"Test","role":"user"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminController_CreateUser_DuplicateEmail(t *testing.T) {
	r := setupAdminTestRouter(t)

	body := `{"email":"admin-ctrl@test.com","password":"securePass123!","nickname":"Dup","role":"user","plan":"free"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- Model Stats ---

func setupModelStatsTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db, nil)

	r := gin.New()
	r.GET("/v1/admin/models/stats", ctrl.GetModelStats)
	r.GET("/v1/admin/models/available", ctrl.ListAvailableModels)
	r.GET("/v1/admin/models/rate-limit-hits", ctrl.GetRateLimitHits)

	return r, db
}

func TestAdminController_GetModelStats_Empty(t *testing.T) {
	r, _ := setupModelStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models/stats?days=30", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Empty(t, data)
	assert.NotNil(t, resp["feature_models"])
	assert.NotNil(t, resp["model_quotas"])
}

func TestAdminController_GetModelStats_WithData(t *testing.T) {
	r, db := setupModelStatsTestRouter(t)
	ctx := t.Context()

	user := db.UserProfile.Create().
		SetEmail("model-stat-user@test.com").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	// Create usage logs for different models
	cost1 := 10.0
	latency1 := 1200
	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetProvider("gemini").
		SetModel("gemini-2.0-flash").
		SetInputTokens(500).
		SetOutputTokens(200).
		SetTotalTokens(700).
		SetEstimatedCostKrw(cost1).
		SetLatencyMs(latency1).
		SetStatus("success").
		SaveX(ctx)

	cost2 := 5.0
	latency2 := 800
	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("analysis").
		SetProvider("gemini").
		SetModel("gemini-2.0-flash").
		SetInputTokens(300).
		SetOutputTokens(100).
		SetTotalTokens(400).
		SetEstimatedCostKrw(cost2).
		SetLatencyMs(latency2).
		SetStatus("success").
		SaveX(ctx)

	db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("coaching").
		SetProvider("groq").
		SetModel("llama-3.3-70b-versatile").
		SetInputTokens(600).
		SetOutputTokens(0).
		SetTotalTokens(600).
		SetStatus("error").
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models/stats?days=30", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Len(t, data, 2) // gemini-2.0-flash and llama-3.3-70b-versatile

	// Find each model in the results
	for _, item := range data {
		m := item.(map[string]any)
		switch m["model"] {
		case "gemini-2.0-flash":
			assert.Equal(t, float64(2), m["call_count"])
			assert.Equal(t, float64(2), m["success_count"])
			assert.Equal(t, float64(0), m["error_count"])
			assert.Equal(t, float64(1100), m["total_tokens"])
			assert.InDelta(t, 15.0, m["total_cost_krw"], 0.01)
			assert.Equal(t, "gemini", m["provider"])
		case "llama-3.3-70b-versatile":
			assert.Equal(t, float64(1), m["call_count"])
			assert.Equal(t, float64(0), m["success_count"])
			assert.Equal(t, float64(1), m["error_count"])
			assert.Equal(t, "groq", m["provider"])
		}
	}
}

func TestAdminController_GetModelStats_FeatureModels(t *testing.T) {
	r, db := setupModelStatsTestRouter(t)
	ctx := t.Context()

	// Create active prompt templates
	db.PromptTemplate.Create().
		SetCategory("coaching").
		SetSubCategory("draft").
		SetName("Draft Prompt").
		SetSystemPrompt("sys").
		SetUserPromptTemplate("usr").
		SetModel("gemini-2.0-flash").
		SetIsActive(true).
		SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models/stats?days=7", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	featureModels := resp["feature_models"].(map[string]any)
	assert.Equal(t, "gemini-2.0-flash", featureModels["coaching/draft"])
}

func TestAdminController_ListAvailableModels(t *testing.T) {
	r, _ := setupModelStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models/available", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Greater(t, len(data), 10) // We have 16+ models

	// Verify each model has required fields
	for _, item := range data {
		m := item.(map[string]any)
		assert.NotEmpty(t, m["id"])
		assert.NotEmpty(t, m["provider"])
		provider := m["provider"].(string)
		assert.True(t, provider == "gemini" || provider == "groq",
			"provider should be gemini or groq, got: %s", provider)
	}
}

func TestAdminController_GetRateLimitHits_NoProvider(t *testing.T) {
	// Controller with nil aiProvider
	r, _ := setupModelStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models/rate-limit-hits?hours=24", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Empty(t, data)
	assert.Equal(t, float64(24), resp["hours"])
}

func TestAdminController_GetRateLimitHits_WithHits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)

	// Create AIProvider with real throttler
	mock := &ai.MockStreamingProvider{}
	aiProvider := ai.NewAIProviderForTest(mock)

	// Record some quota hits
	aiProvider.Throttler().RecordQuotaHit("gemini-3-pro", fmt.Errorf("429 rate limit"))
	aiProvider.Throttler().RecordQuotaHit("gemini-3-pro", fmt.Errorf("429 rate limit again"))
	aiProvider.Throttler().RecordQuotaHit("llama-3.3-70b-versatile", fmt.Errorf("quota exceeded"))

	ctrl := NewAdminController(db, aiProvider)
	r := gin.New()
	r.GET("/v1/admin/models/rate-limit-hits", ctrl.GetRateLimitHits)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models/rate-limit-hits?hours=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	require.Len(t, data, 2) // 2 distinct models

	for _, item := range data {
		h := item.(map[string]any)
		switch h["model"] {
		case "gemini-3-pro":
			assert.Equal(t, float64(2), h["hit_count"])
			assert.Equal(t, "gemini", h["provider"])
		case "llama-3.3-70b-versatile":
			assert.Equal(t, float64(1), h["hit_count"])
			assert.Equal(t, "groq", h["provider"])
		}
	}
}


// ─── Phase 1.7.3: Dashboard API ───

func setupDashboardTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctrl := NewAdminController(db, nil)
	r := gin.New()
	v1 := r.Group("/v1/admin")
	v1.GET("/dashboard", ctrl.GetDashboard)
	v1.GET("/usage/errors", ctrl.ListErrors)
	return r, db
}

func TestGetDashboard_Returns200(t *testing.T) {
	r, _ := setupDashboardTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	data := resp["data"].(map[string]any)
	assert.Contains(t, data, "user_stats")
	assert.Contains(t, data, "ai_metrics_today")
	assert.Contains(t, data, "daily_metrics")
	assert.Contains(t, data, "quota_alerts")
	assert.Contains(t, data, "recent_errors")
	assert.Contains(t, data, "recent_feedbacks")
	assert.Contains(t, data, "pending_feedback_count")
}

func TestGetDashboard_UserStats(t *testing.T) {
	r, db := setupDashboardTestRouter(t)
	ctx := context.Background()

	db.UserProfile.Create().
		SetEmail("u1@test.com").SetNickname("U1").
		SetAuthProvider(userprofile.AuthProviderEmail).SaveX(ctx)
	db.UserProfile.Create().
		SetEmail("u2@test.com").SetNickname("U2").
		SetAuthProvider(userprofile.AuthProviderEmail).SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := parseJSON(t, w)
	data := resp["data"].(map[string]any)
	stats := data["user_stats"].(map[string]any)
	assert.Equal(t, float64(2), stats["total_users"])
}

func TestGetDashboard_RecentErrors_Limit5(t *testing.T) {
	r, db := setupDashboardTestRouter(t)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("err-user@test.com").SetNickname("ErrUser").
		SetAuthProvider(userprofile.AuthProviderEmail).SaveX(ctx)

	// Create 6 error usage logs
	for i := 0; i < 6; i++ {
		ul := db.UsageLog.Create().
			SetUserID(user.ID).SetFeature("draft").SetStatus("error").SaveX(ctx)
		db.AICallError.Create().
			SetUsageLogID(ul.ID).
			SetErrorType("rate_limit").
			SetErrorMessage(fmt.Sprintf("error %d", i)).
			SaveX(ctx)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := parseJSON(t, w)
	data := resp["data"].(map[string]any)
	errors := data["recent_errors"].([]any)
	assert.LessOrEqual(t, len(errors), 5)
}

func TestGetDashboard_EmptyState(t *testing.T) {
	r, _ := setupDashboardTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := parseJSON(t, w)
	data := resp["data"].(map[string]any)
	// Empty arrays, not null
	assert.IsType(t, []any{}, data["recent_errors"])
	assert.IsType(t, []any{}, data["recent_feedbacks"])
	assert.IsType(t, []any{}, data["quota_alerts"])
}

// ─── Phase 1.7.4: Error List API ───

func TestListErrors_Returns200(t *testing.T) {
	r, _ := setupDashboardTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/errors", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Contains(t, resp, "data")
	assert.Contains(t, resp, "count")
}

func TestListErrors_FilterByErrorType(t *testing.T) {
	r, db := setupDashboardTestRouter(t)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("filter-test@test.com").SetNickname("FT").
		SetAuthProvider(userprofile.AuthProviderEmail).SaveX(ctx)

	ulRL := db.UsageLog.Create().SetUserID(user.ID).SetFeature("draft").SetStatus("error").SaveX(ctx)
	db.AICallError.Create().SetUsageLogID(ulRL.ID).SetErrorType("rate_limit").SetErrorMessage("429").SaveX(ctx)

	ulTO := db.UsageLog.Create().SetUserID(user.ID).SetFeature("analysis").SetStatus("error").SaveX(ctx)
	db.AICallError.Create().SetUsageLogID(ulTO.ID).SetErrorType("timeout").SetErrorMessage("deadline").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/errors?error_type=rate_limit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Equal(t, 1, len(data))
	assert.Equal(t, "rate_limit", data[0].(map[string]any)["error_type"])
}

func TestListErrors_Pagination(t *testing.T) {
	r, db := setupDashboardTestRouter(t)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("page-test@test.com").SetNickname("PT").
		SetAuthProvider(userprofile.AuthProviderEmail).SaveX(ctx)

	for i := 0; i < 5; i++ {
		ul := db.UsageLog.Create().SetUserID(user.ID).SetFeature("draft").SetStatus("error").SaveX(ctx)
		db.AICallError.Create().SetUsageLogID(ul.ID).SetErrorType("timeout").SetErrorMessage("to").SaveX(ctx)
	}

	// limit=2
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage/errors?limit=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := parseJSON(t, w)
	data := resp["data"].([]any)
	assert.Equal(t, 2, len(data))
	assert.Equal(t, float64(5), resp["count"])
}
