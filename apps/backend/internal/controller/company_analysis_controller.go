package controller

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CompanyAnalysisController struct {
	analysisService *service.CompanyAnalysisService
}

func NewCompanyAnalysisController(analysisService *service.CompanyAnalysisService) *CompanyAnalysisController {
	return &CompanyAnalysisController{analysisService: analysisService}
}

// AnalyzeCompany handles POST /v1/analyze-company
func (c *CompanyAnalysisController) AnalyzeCompany(ctx *gin.Context) {
	var req struct {
		CompanyName    string  `json:"company_name" binding:"required"`
		JobPostingURL  *string `json:"job_posting_url"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "회사명이 필요합니다.", "code": "VALID_001"},
		})
		return
	}

	// Default to company_name for cache key if URL not provided
	cacheSource := req.CompanyName
	if req.JobPostingURL != nil && *req.JobPostingURL != "" {
		cacheSource = *req.JobPostingURL
	}

	result, fromCache, err := c.analysisService.AnalyzeCompany(ctx.Request.Context(), req.CompanyName, cacheSource)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "기업 분석 실패: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	// Save user-specific analysis record
	if userIDRaw, exists := ctx.Get("user_id"); exists {
		if userID, ok := userIDRaw.(uuid.UUID); ok {
			_, saveErr := c.analysisService.SaveUserAnalysis(
				ctx.Request.Context(),
				userID,
				req.CompanyName,
				req.JobPostingURL,
				result,
			)
			if saveErr != nil {
				slog.Error("failed to save user analysis", "error", saveErr, "user_id", userID)
			}
		}
	}

	// Track usage only if new AI analysis was performed (not cache hit)
	if !fromCache {
		TrackUsageAfterSuccess(ctx, map[string]interface{}{
			"company_name": req.CompanyName,
			"url":          cacheSource,
		})
	}

	ctx.JSON(http.StatusOK, result)
}

// GetRecentAnalyses handles GET /v1/analysis/recent
func (c *CompanyAnalysisController) GetRecentAnalyses(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10
	}

	analyses, err := c.analysisService.GetRecentAnalyses(ctx.Request.Context(), userID, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "분석 기록 조회 실패: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	result := make([]gin.H, 0, len(analyses))
	for _, a := range analyses {
		result = append(result, gin.H{
			"id":           a.ID.String(),
			"company_name": a.CompanyName,
			"job_url":      a.JobURL,
			"created_at":   a.CreatedAt.Format(time.RFC3339),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"analyses": result,
		"total":    len(result),
	})
}
