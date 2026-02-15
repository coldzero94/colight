package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InterviewController handles interview endpoints.
type InterviewController struct {
	interviewService *service.InterviewService
}

// NewInterviewController creates a new interview controller.
func NewInterviewController(interviewService *service.InterviewService) *InterviewController {
	return &InterviewController{interviewService: interviewService}
}

// PostQuestion handles POST /v1/interview/question
func (ctrl *InterviewController) PostQuestion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	var req service.GenerateQuestionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "잘못된 요청입니다.",
		})
		return
	}

	result, err := ctrl.interviewService.GenerateQuestion(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		if err == service.ErrInvalidInterviewStage {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "유효하지 않은 인터뷰 단계입니다.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "질문 생성 중 오류가 발생했습니다.",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
