package controller

import (
	"bytes"
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

func setupExperienceTestRouter(t *testing.T) (*gin.Engine, *service.ExperienceService, *service.AuthService) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestClient(t)
	cfg := testutil.NewTestConfig()
	ts := service.NewTokenService(testutil.TestJWTSecret, testutil.TestAccessTokenTTL, testutil.TestRefreshTokenTTL)
	as := service.NewAuthService(cfg, db, ts)
	es := service.NewExperienceService(db)
	ctrl := NewExperienceController(es)

	r := gin.New()
	v1 := r.Group("/v1/experiences")
	v1.Use(func(c *gin.Context) {
		// Simulate auth middleware: parse user_id from header
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	{
		v1.POST("", ctrl.Create)
		v1.GET("", ctrl.List)
		v1.GET("/:id", ctrl.Get)
		v1.PATCH("/:id", ctrl.Update)
		v1.DELETE("/:id", ctrl.Delete)
	}

	return r, es, as
}

func createTestUserViaService(t *testing.T, as *service.AuthService) uuid.UUID {
	t.Helper()
	result, err := as.Signup(t.Context(), "ctrl-exp-"+uuid.New().String()[:8]+"@test.com", "password123", "tester")
	require.NoError(t, err)
	return result.User.ID
}

func expRequest(router *gin.Engine, method, path string, body any, userID uuid.UUID) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		req.Header.Set("X-Test-UserID", userID.String())
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// --- Create ---

func TestExperienceController_Create(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	w := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title":          "인턴 경험",
		"category":       "인턴",
		"content":        "내용",
		"star_situation":  "상황",
		"star_task":       "과제",
		"star_action":     "행동",
		"star_result":     "결과",
	}, userID)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseJSON(t, w)
	assert.NotEmpty(t, resp["id"])
}

func TestExperienceController_Create_InvalidTitle(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	w := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"content": "내용",
	}, userID)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- List ---

func TestExperienceController_List(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	// Create 2 experiences
	expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "경험1", "content": "c1",
	}, userID)
	expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "경험2", "content": "c2",
	}, userID)

	w := expRequest(r, http.MethodGet, "/v1/experiences", nil, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	exps := resp["experiences"].([]any)
	assert.Len(t, exps, 2)
}

// --- Unauthorized ---

func TestExperienceController_Unauthorized(t *testing.T) {
	r, _, _ := setupExperienceTestRouter(t)

	// No user ID header
	w := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "test", "content": "c",
	}, uuid.Nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseJSON(t, w)
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "AUTH_001", errObj["code"])
}

// --- Get ---

func TestExperienceController_Get(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	createW := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "상세 경험", "content": "내용", "category": "프로젝트",
	}, userID)
	createResp := parseJSON(t, createW)
	id := createResp["id"].(string)

	w := expRequest(r, http.MethodGet, "/v1/experiences/"+id, nil, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "상세 경험", resp["title"])
	assert.Equal(t, "프로젝트", resp["category"])
}

func TestExperienceController_Get_NotFound(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	w := expRequest(r, http.MethodGet, "/v1/experiences/"+uuid.New().String(), nil, userID)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseJSON(t, w)
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "EXP_001", errObj["code"])
}

func TestExperienceController_Get_Forbidden(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	owner := createTestUserViaService(t, as)
	other := createTestUserViaService(t, as)

	createW := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "Owner Exp", "content": "c",
	}, owner)
	createResp := parseJSON(t, createW)
	id := createResp["id"].(string)

	w := expRequest(r, http.MethodGet, "/v1/experiences/"+id, nil, other)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Update ---

func TestExperienceController_Update(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	createW := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "Before", "content": "c",
	}, userID)
	createResp := parseJSON(t, createW)
	id := createResp["id"].(string)

	w := expRequest(r, http.MethodPatch, "/v1/experiences/"+id, map[string]any{
		"title": "After",
	}, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, id, resp["id"])

	// Verify the update took effect
	getW := expRequest(r, http.MethodGet, "/v1/experiences/"+id, nil, userID)
	getResp := parseJSON(t, getW)
	assert.Equal(t, "After", getResp["title"])
}

// --- Delete ---

func TestExperienceController_Delete(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	createW := expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "To Delete", "content": "c",
	}, userID)
	createResp := parseJSON(t, createW)
	id := createResp["id"].(string)

	w := expRequest(r, http.MethodDelete, "/v1/experiences/"+id, nil, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, true, resp["success"])

	// Verify it's gone
	getW := expRequest(r, http.MethodGet, "/v1/experiences/"+id, nil, userID)
	assert.Equal(t, http.StatusNotFound, getW.Code)
}

// --- Sorting ---

func TestExperienceController_List_WithSort(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "B title", "content": "c",
	}, userID)
	expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "A title", "content": "c",
	}, userID)

	w := expRequest(r, http.MethodGet, "/v1/experiences?sort=title", nil, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	exps := resp["experiences"].([]any)
	require.Len(t, exps, 2)
	assert.Equal(t, "A title", exps[0].(map[string]any)["title"])
	assert.Equal(t, "B title", exps[1].(map[string]any)["title"])
}

// --- Category filter ---

func TestExperienceController_List_WithCategory(t *testing.T) {
	r, _, as := setupExperienceTestRouter(t)
	userID := createTestUserViaService(t, as)

	expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "인턴", "content": "c", "category": "인턴",
	}, userID)
	expRequest(r, http.MethodPost, "/v1/experiences", map[string]any{
		"title": "프로젝트", "content": "c", "category": "프로젝트",
	}, userID)

	w := expRequest(r, http.MethodGet, "/v1/experiences?category=인턴", nil, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	exps := resp["experiences"].([]any)
	assert.Len(t, exps, 1)
	assert.Equal(t, "인턴", exps[0].(map[string]any)["title"])
}
