package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MatchingController struct {
	matchingService *service.MatchingService
	analysisService *service.CompanyAnalysisService
}

func NewMatchingController(matchingService *service.MatchingService, analysisService *service.CompanyAnalysisService) *MatchingController {
	return &MatchingController{
		matchingService: matchingService,
		analysisService: analysisService,
	}
}

// MatchExperiences handles POST /v1/match
func (c *MatchingController) MatchExperiences(ctx *gin.Context) {
	// Get user ID from auth middleware
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	var req struct {
		CompanyName string `json:"company_name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "회사명이 필요합니다.", "code": "VALID_001"},
		})
		return
	}

	// 1. Get company analysis
	companyAnalysis, err := c.analysisService.AnalyzeCompany(ctx.Request.Context(), req.CompanyName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "기업 분석 실패: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	// 2. Get user's experiences
	experienceService := service.NewExperienceService(c.matchingService.GetClient())
	experiences, err := experienceService.GetExperiences(ctx.Request.Context(), userID.(uuid.UUID), service.ExperienceListParams{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "경험 조회 실패", "code": "SYS_001"},
		})
		return
	}

	// 3. Match all experiences
	var matches []service.MatchResult
	for _, exp := range experiences {
		result, err := c.matchingService.MatchExperience(ctx.Request.Context(), userID.(uuid.UUID), exp.ID, companyAnalysis)
		if err != nil {
			// Log error but continue with other experiences
			continue
		}
		matches = append(matches, *result)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"company_name": req.CompanyName,
		"matches":      matches,
		"total":        len(matches),
	})
}
