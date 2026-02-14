package controller

import (
	"encoding/json"
	"fmt"
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

// PostDraft handles POST /v1/coaching/draft with SSE streaming
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
		ApplicationID  string   `json:"application_id" binding:"required,uuid"`
		ExperienceIDs  []string `json:"experience_ids" binding:"required,min=1,max=3,dive,uuid"`
		QuestionText   string   `json:"question_text" binding:"required,min=10"`
		CharLimit      int      `json:"char_limit" binding:"required,min=200,max=2000"`
		AnalysisResult any      `json:"analysis_result"` // Optional
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

	// Set SSE headers
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.Status(http.StatusOK)

	uid := userID.(uuid.UUID)

	// Stream draft via SSE
	resp, err := c.coachingService.GenerateDraftStream(
		ctx.Request.Context(),
		uid,
		appID,
		expIDs,
		req.QuestionText,
		req.CharLimit,
		req.AnalysisResult,
		func(chunk string) {
			data, _ := json.Marshal(gin.H{"type": "text", "content": chunk})
			fmt.Fprintf(ctx.Writer, "event: text\ndata: %s\n\n", data)
			ctx.Writer.Flush()
		},
	)

	if err != nil {
		// Send error as SSE event (headers already sent, can't change status code)
		errMsg := "AI 초안 생성 중 오류가 발생했습니다"
		if err == service.ErrApplicationNotFound {
			errMsg = "지원 정보를 찾을 수 없습니다"
		} else if err == service.ErrApplicationForbidden || err == service.ErrExperienceForbidden {
			errMsg = "접근 권한이 없습니다"
		}
		data, _ := json.Marshal(gin.H{"type": "error", "message": errMsg})
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", data)
		ctx.Writer.Flush()
		return
	}

	// Save results after successful streaming
	result, err := c.coachingService.SaveDraftResult(
		ctx.Request.Context(),
		uid,
		appID,
		req.QuestionText,
		req.CharLimit,
		resp.Content,
		resp.InputTokens,
		resp.OutputTokens,
	)

	if err != nil {
		data, _ := json.Marshal(gin.H{"type": "error", "message": "결과 저장 중 오류가 발생했습니다"})
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", data)
		ctx.Writer.Flush()
		return
	}

	// Send done event with IDs
	data, _ := json.Marshal(gin.H{
		"type":            "done",
		"cover_letter_id": result.CoverLetterID.String(),
		"session_id":      result.SessionID.String(),
	})
	fmt.Fprintf(ctx.Writer, "event: done\ndata: %s\n\n", data)
	ctx.Writer.Flush()
}

// GetSessions handles GET /v1/coaching/sessions
func (c *CoachingController) GetSessions(ctx *gin.Context) {
	// Get user ID from auth middleware
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	coverLetterID := ctx.Query("cover_letter_id")
	if coverLetterID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "cover_letter_id가 필요합니다",
		})
		return
	}

	clID, err := uuid.Parse(coverLetterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "유효하지 않은 cover_letter_id입니다",
		})
		return
	}

	// Get sessions (only user's own sessions)
	sessions, err := c.coachingService.GetSessions(ctx.Request.Context(), userID.(uuid.UUID), clID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "세션 조회 중 오류가 발생했습니다",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}
