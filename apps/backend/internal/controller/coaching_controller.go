package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CoachingController struct {
	coachingService *service.CoachingService
}

func NewCoachingController(coachingService *service.CoachingService) *CoachingController {
	return &CoachingController{
		coachingService: coachingService,
	}
}

// PostDraft handles POST /v1/coaching/draft
func (c *CoachingController) PostDraft(ctx *gin.Context) {
	// Get user ID from auth middleware
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	var req struct {
		ApplicationID string   `json:"application_id" binding:"required,uuid"`
		ExperienceIDs []string `json:"experience_ids" binding:"required,min=1,max=3,dive,uuid"`
		QuestionText  string   `json:"question_text" binding:"required,min=10"`
		CharLimit     int      `json:"char_limit" binding:"required,min=200,max=2000"`
		AnalysisResult any     `json:"analysis_result"` // Optional
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "입력 값이 올바르지 않습니다: " + err.Error(),
		})
		return
	}

	// Parse UUIDs
	appID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "유효하지 않은 application_id입니다",
		})
		return
	}

	expIDs := make([]uuid.UUID, len(req.ExperienceIDs))
	for i, id := range req.ExperienceIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "유효하지 않은 experience_id입니다",
			})
			return
		}
		expIDs[i] = parsed
	}

	// Generate draft
	draft, err := c.coachingService.GenerateDraft(
		ctx.Request.Context(),
		userID.(uuid.UUID),
		appID,
		expIDs,
		req.QuestionText,
		req.CharLimit,
		req.AnalysisResult,
	)

	if err != nil {
		// Check for specific errors
		if err == service.ErrApplicationNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "지원 정보를 찾을 수 없습니다",
			})
			return
		}
		if err == service.ErrApplicationForbidden || err == service.ErrExperienceForbidden {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "접근 권한이 없습니다",
			})
			return
		}

		// AI failure or other errors
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI 초안 생성 중 오류가 발생했습니다: " + err.Error(),
		})
		return
	}

	// Return draft as JSON (SSE streaming will be added later)
	ctx.JSON(http.StatusOK, gin.H{
		"draft": draft,
	})
}
