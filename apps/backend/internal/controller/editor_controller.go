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

// PatchCoverLetter handles PATCH /v1/coaching/cover-letters/:id
func (c *EditorController) PatchCoverLetter(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	coverLetterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content가 필요합니다"})
		return
	}

	err = c.editorService.UpdateContent(ctx.Request.Context(), userID.(uuid.UUID), coverLetterID, req.Content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "저장 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"updated_at": "now"})
}

// PostVersion handles POST /v1/coaching/cover-letters/:id/versions
func (c *EditorController) PostVersion(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	coverLetterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다"})
		return
	}

	var req struct {
		Content       string `json:"content" binding:"required"`
		ChangeSummary string `json:"change_summary"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content가 필요합니다"})
		return
	}

	version, err := c.editorService.CreateVersion(ctx.Request.Context(), userID.(uuid.UUID), coverLetterID, req.Content, req.ChangeSummary)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "버전 생성 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"version": version})
}

// GetVersions handles GET /v1/coaching/cover-letters/:id/versions
func (c *EditorController) GetVersions(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	coverLetterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다"})
		return
	}

	versions, err := c.editorService.GetVersions(ctx.Request.Context(), userID.(uuid.UUID), coverLetterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "버전 조회 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"versions": versions})
}
