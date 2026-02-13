package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type CompanyDataController struct {
	companyDataService *service.CompanyDataService
}

func NewCompanyDataController(companyDataService *service.CompanyDataService) *CompanyDataController {
	return &CompanyDataController{companyDataService: companyDataService}
}

// GetCompanyData handles GET /v1/company-data?name=회사명
func (c *CompanyDataController) GetCompanyData(ctx *gin.Context) {
	companyName := ctx.Query("name")
	if companyName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "회사명이 필요합니다.", "code": "VALID_001"},
		})
		return
	}

	result, err := c.companyDataService.GetCompanyData(ctx.Request.Context(), companyName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "기업 정보 조회 실패: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
