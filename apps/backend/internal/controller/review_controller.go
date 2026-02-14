package controller

import (
	"errors"
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReviewController struct {
	reviewService *service.ReviewService
}

func NewReviewController(reviewService *service.ReviewService) *ReviewController {
	return &ReviewController{
		reviewService: reviewService,
	}
}

// PostReview handles POST /v1/coaching/review
func (c *ReviewController) PostReview(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	var req struct {
		CoverLetterID string `json:"cover_letter_id" binding:"required,uuid"`
		Content       string `json:"content" binding:"required,min=50"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "입력 값이 올바르지 않습니다: " + err.Error(),
		})
		return
	}

	clID, err := uuid.Parse(req.CoverLetterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "유효하지 않은 cover_letter_id입니다",
		})
		return
	}

	result, err := c.reviewService.ReviewCoverLetter(
		ctx.Request.Context(),
		userID.(uuid.UUID),
		clID,
		req.Content,
	)
	if err != nil {
		if errors.Is(err, service.ErrCoverLetterNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "자소서를 찾을 수 없습니다",
			})
			return
		}
		if errors.Is(err, service.ErrCoverLetterForbidden) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "접근 권한이 없습니다",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "첨삭 평가 중 오류가 발생했습니다",
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
