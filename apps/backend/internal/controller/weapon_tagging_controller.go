package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WeaponTaggingController struct {
	service *service.WeaponTaggingService
}

func NewWeaponTaggingController(service *service.WeaponTaggingService) *WeaponTaggingController {
	return &WeaponTaggingController{service: service}
}

// Tag handles POST /v1/experiences/:id/tag
func (c *WeaponTaggingController) Tag(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	// Parse experience ID
	experienceIDStr := ctx.Param("id")
	experienceID, err := uuid.Parse(experienceIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "유효하지 않은 경험 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	// Call service
	result, err := c.service.TagExperience(ctx.Request.Context(), experienceID, userID.(uuid.UUID))
	if err != nil {
		handleWeaponTaggingError(ctx, err)
		return
	}

	// Return result
	ctx.JSON(http.StatusOK, gin.H{
		"primary_weapon": gin.H{
			"code":       result.PrimaryWeapon.Code,
			"confidence": result.PrimaryWeapon.Confidence,
			"reasoning":  result.PrimaryWeapon.Reasoning,
		},
		"secondary_weapons": result.SecondaryWeapons,
	})
}

func handleWeaponTaggingError(ctx *gin.Context, err error) {
	switch err {
	case service.ErrExperienceNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"message": "경험을 찾을 수 없습니다.", "code": "EXP_001"},
		})
	case service.ErrExperienceForbidden:
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"message": "이 경험에 접근할 권한이 없습니다.", "code": "EXP_002"},
		})
	default:
		if err.Error() == "experience text too short (minimum 50 characters)" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": "경험 내용이 너무 짧습니다 (최소 50자).", "code": "VALID_003"},
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": "AI 태깅 중 오류가 발생했습니다: " + err.Error(), "code": "SYS_001"},
			})
		}
	}
}
