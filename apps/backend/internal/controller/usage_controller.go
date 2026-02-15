package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UsageController handles usage tracking endpoints.
type UsageController struct {
	usageService *service.UsageService
}

// NewUsageController creates a new usage controller.
func NewUsageController(usageService *service.UsageService) *UsageController {
	return &UsageController{usageService: usageService}
}

// GetUsage handles GET /v1/usage
func (c *UsageController) GetUsage(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	features, plan, err := c.usageService.GetAllUsage(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "사용량 조회 중 오류가 발생했습니다",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"plan":     plan,
		"features": features,
	})
}
