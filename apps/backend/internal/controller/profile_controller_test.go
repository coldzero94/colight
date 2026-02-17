package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupProfileTestRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)

	hash, _ := bcrypt.GenerateFromPassword([]byte("oldPassword123"), 12)
	user := db.UserProfile.Create().
		SetEmail("test@example.com").
		SetPasswordHash(string(hash)).
		SetNickname("TestUser").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser).
		SaveX(context.Background())

	ctrl := NewProfileController(db)

	r := gin.New()
	// Test middleware to set user_id from header
	r.Use(func(c *gin.Context) {
		if userIDStr := c.GetHeader("X-Test-UserID"); userIDStr != "" {
			if uid, err := uuid.Parse(userIDStr); err == nil {
				c.Set("user_id", uid)
			}
		}
		c.Next()
	})
	v1 := r.Group("/v1")
	v1.PUT("/profile", ctrl.UpdateProfile)
	v1.POST("/profile/password", ctrl.ChangePassword)

	return r, user.ID.String()
}

// --- UpdateProfile ---

func TestProfileController_UpdateProfile_Success(t *testing.T) {
	r, userID := setupProfileTestRouter(t)

	body := `{"nickname":"NewNickname"}`
	req := httptest.NewRequest(http.MethodPut, "/v1/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseJSON(t, w)
	assert.Equal(t, "NewNickname", resp["nickname"])
}

func TestProfileController_UpdateProfile_EmptyNickname(t *testing.T) {
	r, userID := setupProfileTestRouter(t)

	body := `{"nickname":""}`
	req := httptest.NewRequest(http.MethodPut, "/v1/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- ChangePassword ---

func TestProfileController_ChangePassword_Success(t *testing.T) {
	r, userID := setupProfileTestRouter(t)

	body := `{"current_password":"oldPassword123","new_password":"newPassword456"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/profile/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProfileController_ChangePassword_WrongCurrent(t *testing.T) {
	r, userID := setupProfileTestRouter(t)

	body := `{"current_password":"wrongPassword","new_password":"newPassword456"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/profile/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProfileController_ChangePassword_WeakNew(t *testing.T) {
	r, userID := setupProfileTestRouter(t)

	body := `{"current_password":"oldPassword123","new_password":"weak"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/profile/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
