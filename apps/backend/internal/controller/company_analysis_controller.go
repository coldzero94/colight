package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
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
		CompanyName string `json:"company_name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "회사명이 필요합니다.", "code": "VALID_001"},
		})
		return
	}

	result, err := c.analysisService.AnalyzeCompany(ctx.Request.Context(), req.CompanyName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "기업 분석 실패: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
