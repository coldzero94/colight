package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsageMiddlewareTestRouter(t *testing.T) (*gin.Engine, *service.UsageService, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	testutil.CleanAllTables(client)

	// Create test user
	ctx := context.Background()
	user := client.UserProfile.Create().
		SetEmail("middleware-test@test.com").
		SetAuthProvider("email").
		SetPlan(userprofile.PlanFree).
		SaveX(ctx)

	usageService := service.NewUsageService(client)

	router := gin.New()
	return router, usageService, user.ID
}

func TestUsageLimitMiddleware_NoAuth(t *testing.T) {
	router, usageService, _ := setupUsageMiddlewareTestRouter(t)

	// Setup route with middleware but no auth
	router.GET("/test", UsageLimitMiddleware(usageService, "experience"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "인증이 필요합니다.", resp["error"])
}

func TestUsageLimitMiddleware_UnderLimit(t *testing.T) {
	router, usageService, userID := setupUsageMiddlewareTestRouter(t)

	// Setup route with middleware and auth
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/test", UsageLimitMiddleware(usageService, "experience"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "success", resp["message"])
}

func TestUsageLimitMiddleware_OverLimit(t *testing.T) {
	router, usageService, userID := setupUsageMiddlewareTestRouter(t)

	// Track usage to hit limit (free tier experience limit is 3)
	ctx := context.Background()
	for range 3 {
		err := usageService.TrackUsage(ctx, userID, "experience", nil)
		require.NoError(t, err)
	}

	// Setup route with middleware and auth
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/test", UsageLimitMiddleware(usageService, "experience"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	errorObj, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "무료 사용 횟수를 초과했습니다.", errorObj["message"])
	assert.Equal(t, "USAGE_001", errorObj["code"])
	assert.Equal(t, float64(3), errorObj["used"])
	assert.Equal(t, float64(3), errorObj["limit"])
	assert.Equal(t, "/pricing", errorObj["upgrade_url"])
}

func TestUsageLimitMiddleware_InvalidFeature(t *testing.T) {
	router, usageService, userID := setupUsageMiddlewareTestRouter(t)

	// Setup route with middleware and invalid feature
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/test", UsageLimitMiddleware(usageService, "invalid_feature"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "잘못된 기능입니다", resp["error"])
}

func TestUsageLimitMiddleware_SetsContext(t *testing.T) {
	router, usageService, userID := setupUsageMiddlewareTestRouter(t)

	var capturedUsageService *service.UsageService
	var capturedFeature string

	// Setup route with middleware and capture context values
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/test", UsageLimitMiddleware(usageService, "experience"), func(c *gin.Context) {
		svc, exists := c.Get("usage_service")
		if exists {
			capturedUsageService = svc.(*service.UsageService)
		}
		feature, exists := c.Get("usage_feature")
		if exists {
			capturedFeature = feature.(string)
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, capturedUsageService)
	assert.Equal(t, "experience", capturedFeature)
}

func TestTrackUsageAfterSuccess_Records(t *testing.T) {
	router, usageService, userID := setupUsageMiddlewareTestRouter(t)

	// Setup route with middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.POST("/test", UsageLimitMiddleware(usageService, "draft"), func(c *gin.Context) {
		// Simulate successful operation
		metadata := map[string]any{
			"cover_letter_id": "test-cl-123",
		}
		TrackUsageAfterSuccess(c, metadata)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify usage was tracked
	ctx := context.Background()
	status, err := usageService.CheckLimit(ctx, userID, "draft")
	require.NoError(t, err)
	assert.Equal(t, 1, status.Used)
}

func TestTrackUsageAfterSuccess_NoContext(t *testing.T) {
	router, _, userID := setupUsageMiddlewareTestRouter(t)

	// Setup route WITHOUT middleware (no context values set)
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.POST("/test", func(c *gin.Context) {
		// Call TrackUsageAfterSuccess without middleware context
		metadata := map[string]any{
			"cover_letter_id": "test-cl-123",
		}
		// This should not panic, just silently return
		TrackUsageAfterSuccess(c, metadata)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed without panic
	assert.Equal(t, http.StatusOK, w.Code)
}
