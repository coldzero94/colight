package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ApplicationController handles dashboard application endpoints.
type ApplicationController struct {
	applicationService *service.ApplicationService
}

// NewApplicationController creates a new application controller.
func NewApplicationController(applicationService *service.ApplicationService) *ApplicationController {
	return &ApplicationController{applicationService: applicationService}
}

// ListApplications handles GET /v1/applications
func (c *ApplicationController) ListApplications(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	apps, err := c.applicationService.ListApplications(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "지원 목록 조회 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"applications": apps})
}

// UpdateStatus handles PATCH /v1/applications/:id/status
func (c *ApplicationController) UpdateStatus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	appID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다."})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status 필드가 필요합니다."})
		return
	}

	result, err := c.applicationService.UpdateStatus(ctx.Request.Context(), userID.(uuid.UUID), appID, req.Status)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "지원 정보를 찾을 수 없습니다."})
			return
		}
		if err == service.ErrApplicationForbidden {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다."})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "상태 변경 실패"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetStats handles GET /v1/applications/stats
func (c *ApplicationController) GetStats(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	stats, err := c.applicationService.GetStats(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "통계 조회 실패"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
