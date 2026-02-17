package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type StarGenerationController struct {
	service *service.StarGenerationService
}

func NewStarGenerationController(service *service.StarGenerationService) *StarGenerationController {
	return &StarGenerationController{service: service}
}

// GenerateSTAR handles POST /v1/experiences/generate-star
func (c *StarGenerationController) GenerateSTAR(ctx *gin.Context) {
	var req service.StarGenerationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 요청입니다.", "code": "VALID_001"},
		})
		return
	}

	if len(req.Content) < 30 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "내용이 너무 짧습니다 (최소 30자).", "code": "VALID_003"},
		})
		return
	}

	result, err := c.service.GenerateSTAR(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "AI STAR 생성 중 오류가 발생했습니다: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
