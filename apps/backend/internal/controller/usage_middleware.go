package controller

import (
	"errors"
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UsageLimitMiddleware returns a gin middleware that checks usage limits for a feature.
// It blocks the request with 403 if the free-tier limit is exceeded.
func UsageLimitMiddleware(usageService *service.UsageService, feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
			c.Abort()
			return
		}

		status, err := usageService.CheckLimit(c.Request.Context(), userID.(uuid.UUID), feature)
		if err != nil {
			if errors.Is(err, service.ErrInvalidFeature) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "잘못된 기능입니다"})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "사용량 확인 중 오류가 발생했습니다"})
			c.Abort()
			return
		}

		if !status.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message":     "무료 사용 횟수를 초과했습니다.",
					"code":        "USAGE_001",
					"used":        status.Used,
					"limit":       status.Limit,
					"upgrade_url": "/pricing",
				},
			})
			c.Abort()
			return
		}

		// Store usage service and feature in context for post-handler tracking
		c.Set("usage_service", usageService)
		c.Set("usage_feature", feature)
		c.Next()
	}
}

// TrackUsageAfterSuccess records usage after a successful handler.
// Call this at the end of a handler after confirming success.
func TrackUsageAfterSuccess(c *gin.Context, metadata map[string]interface{}) {
	usageSvc, exists := c.Get("usage_service")
	if !exists {
		return
	}
	feature, exists := c.Get("usage_feature")
	if !exists {
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		return
	}

	_ = usageSvc.(*service.UsageService).TrackUsage(
		c.Request.Context(),
		userID.(uuid.UUID),
		feature.(string),
		metadata,
	)
}
