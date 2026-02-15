package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FeedbackController handles feedback endpoints.
type FeedbackController struct {
	feedbackService *service.FeedbackService
}

// NewFeedbackController creates a new feedback controller.
func NewFeedbackController(feedbackService *service.FeedbackService) *FeedbackController {
	return &FeedbackController{feedbackService: feedbackService}
}

// SubmitFeedback handles POST /v1/feedback
func (ctrl *FeedbackController) SubmitFeedback(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	var input service.FeedbackInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "잘못된 요청입니다.",
		})
		return
	}

	fb, err := ctrl.feedbackService.Submit(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		if err == service.ErrEmptyFeedbackContent {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "피드백 내용을 입력해주세요.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "피드백 제출 중 오류가 발생했습니다.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         fb.ID,
		"created_at": fb.CreatedAt,
	})
}
