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

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *service.AuthService, *service.TokenService) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestClient(t)
	cfg := testutil.NewTestConfig()
	ts := service.NewTokenService(testutil.TestJWTSecret, testutil.TestAccessTokenTTL, testutil.TestRefreshTokenTTL)
	as := service.NewAuthService(cfg, db, ts)
	ctrl := NewAuthController(as, cfg)

	r := gin.New()
	v1 := r.Group("/v1/auth")
	v1.POST("/signup", ctrl.Signup)
	v1.POST("/login", ctrl.Login)
	v1.POST("/refresh", ctrl.Refresh)
	v1.POST("/logout", ctrl.Logout)
	v1.GET("/me", func(c *gin.Context) {
		// Simulate auth middleware: parse user_id from header
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		ctrl.Me(c)
	})

	return r, as, ts
}

func postJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	return sendJSON(router, http.MethodPost, path, body)
}

func putJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	return sendJSON(router, http.MethodPut, path, body)
}

func sendJSON(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func parseJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// --- Signup ---

func TestAuthController_Signup_Success(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/signup", map[string]string{
		"email":    "ctrl-test@example.com",
		"password": "password123",
	})

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseJSON(t, w)
	assert.NotNil(t, resp["user"])
	assert.NotNil(t, resp["tokens"])

	tokens := resp["tokens"].(map[string]any)
	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
}

func TestAuthController_Signup_WithNickname(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/signup", map[string]any{
		"email":    "ctrl-nick@example.com",
		"password": "password123",
		"nickname": "tester",
	})

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseJSON(t, w)
	user := resp["user"].(map[string]any)
	assert.Equal(t, "tester", user["nickname"])
}

func TestAuthController_Signup_DuplicateEmail(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "ctrl-dup@example.com", "password": "password123",
	})

	w := postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "ctrl-dup@example.com", "password": "password456",
	})

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := parseJSON(t, w)
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "AUTH_004", errObj["code"])
}

func TestAuthController_Signup_InvalidEmail(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "not-an-email", "password": "password123",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseJSON(t, w)
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "VALID_001", errObj["code"])
}

func TestAuthController_Signup_ShortPassword(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "short@example.com", "password": "short",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Signup_MissingFields(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/signup", map[string]string{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Login ---

func TestAuthController_Login_Success(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "ctrl-login@example.com", "password": "password123",
	})

	w := postJSON(r, "/v1/auth/login", map[string]string{
		"email": "ctrl-login@example.com", "password": "password123",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.NotNil(t, resp["user"])
	assert.NotNil(t, resp["tokens"])
}

func TestAuthController_Login_WrongPassword(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "ctrl-wrong@example.com", "password": "password123",
	})

	w := postJSON(r, "/v1/auth/login", map[string]string{
		"email": "ctrl-wrong@example.com", "password": "wrongpassword",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseJSON(t, w)
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "AUTH_005", errObj["code"])
}

func TestAuthController_Login_NonExistentUser(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/login", map[string]string{
		"email": "ctrl-noone@example.com", "password": "password123",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_Login_InvalidInput(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/login", map[string]string{
		"email": "bad-email",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Refresh ---

func TestAuthController_Refresh_Success(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	signupW := postJSON(r, "/v1/auth/signup", map[string]string{
		"email": "ctrl-refresh@example.com", "password": "password123",
	})
	signupResp := parseJSON(t, signupW)
	tokens := signupResp["tokens"].(map[string]any)

	w := postJSON(r, "/v1/auth/refresh", map[string]string{
		"refresh_token": tokens["refresh_token"].(string),
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	newTokens := resp["tokens"].(map[string]any)
	assert.NotEmpty(t, newTokens["access_token"])
	assert.NotEmpty(t, newTokens["refresh_token"])
}

func TestAuthController_Refresh_InvalidToken(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/refresh", map[string]string{
		"refresh_token": "invalid-token",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseJSON(t, w)
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "AUTH_002", errObj["code"])
}

func TestAuthController_Refresh_MissingToken(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	w := postJSON(r, "/v1/auth/refresh", map[string]string{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Me ---

func TestAuthController_Me_Success(t *testing.T) {
	r, as, _ := setupAuthTestRouter(t)

	result, err := as.Signup(t.Context(), "ctrl-me@example.com", "password123", "myuser")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("X-Test-UserID", result.User.ID.String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "ctrl-me@example.com", resp["email"])
	assert.Equal(t, "myuser", resp["nickname"])
}

func TestAuthController_Me_Unauthenticated(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Logout ---

func TestAuthController_Logout(t *testing.T) {
	r, _, _ := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Contains(t, resp["message"], "로그아웃")
}
