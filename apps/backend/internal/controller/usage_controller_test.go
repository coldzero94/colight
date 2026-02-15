package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsageTestRouter(t *testing.T) (*gin.Engine, *ent.Client, *service.UsageService) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	usageService := service.NewUsageService(client)
	usageCtrl := NewUsageController(usageService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.GET("/usage", usageCtrl.GetUsage)

	return router, client, usageService
}

func TestGetUsage_Unauthorized(t *testing.T) {
	router, _, _ := setupUsageTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUsage_Success(t *testing.T) {
	router, client, usageSvc := setupUsageTestRouter(t)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("usage-test@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanFree).
		SaveX(ctx)

	// Track some usage
	_ = usageSvc.TrackUsage(ctx, user.ID, "experience", nil)
	_ = usageSvc.TrackUsage(ctx, user.ID, "experience", nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/usage", nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Plan     string                          `json:"plan"`
		Features map[string]*service.UsageStatus `json:"features"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "free", resp.Plan)
	assert.Len(t, resp.Features, 6)
	assert.Equal(t, 2, resp.Features["experience"].Used)
	assert.Equal(t, 3, resp.Features["experience"].Limit)
	assert.Equal(t, 1, resp.Features["experience"].Remaining)
}
