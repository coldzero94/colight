package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EditorController struct {
	editorService *service.EditorService
}

func NewEditorController(editorService *service.EditorService) *EditorController {
	return &EditorController{
		editorService: editorService,
	}
}

// GetCoverLetter handles GET /v1/coaching/cover-letters/:id
func (c *EditorController) GetCoverLetter(ctx *gin.Context) {
	// Get user ID from auth middleware
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다",
		})
		return
	}

	coverLetterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "유효하지 않은 ID입니다",
		})
		return
	}

	coverLetter, err := c.editorService.GetCoverLetter(ctx.Request.Context(), userID.(uuid.UUID), coverLetterID)
	if err != nil {
		if err.Error() == "cover letter not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "자소서를 찾을 수 없습니다",
			})
			return
		}
		if err.Error() == "forbidden: not the owner" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "접근 권한이 없습니다",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "조회 중 오류가 발생했습니다",
		})
		return
	}

	ctx.JSON(http.StatusOK, coverLetter)
}
