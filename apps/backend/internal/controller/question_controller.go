package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuestionController struct {
	questionService *service.QuestionService
}

func NewQuestionController(questionService *service.QuestionService) *QuestionController {
	return &QuestionController{
		questionService: questionService,
	}
}

// GetApplications handles GET /v1/applications
func (c *QuestionController) GetApplications(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	apps, err := c.questionService.ListApplications(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "지원 목록 조회 실패"})
		return
	}

	type appResponse struct {
		ID          string `json:"id"`
		CompanyName string `json:"company_name"`
		Position    string `json:"position"`
		Status      string `json:"status"`
		CreatedAt   string `json:"created_at"`
	}

	result := make([]appResponse, len(apps))
	for i, app := range apps {
		result[i] = appResponse{
			ID:          app.ID.String(),
			CompanyName: app.CompanyName,
			Position:    app.Position,
			Status:      string(app.Status),
			CreatedAt:   app.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"applications": result})
}

// PostQuestionAnalysis handles POST /v1/coaching/question-analysis
func (c *QuestionController) PostQuestionAnalysis(ctx *gin.Context) {
	// Get user ID from auth middleware
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "인증이 필요합니다.",
		})
		return
	}

	var req struct {
		ApplicationID string `json:"application_id" binding:"required,uuid"`
		QuestionText  string `json:"question_text" binding:"required,min=10,max=500"`
		CharLimit     int    `json:"char_limit" binding:"required,min=200,max=2000"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "입력 값이 올바르지 않습니다: " + err.Error(),
		})
		return
	}

	// Parse application ID
	appID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "유효하지 않은 application_id입니다",
		})
		return
	}

	// Call question analysis service
	result, err := c.questionService.AnalyzeQuestion(
		ctx.Request.Context(),
		userID.(uuid.UUID),
		appID,
		req.QuestionText,
		req.CharLimit,
	)

	if err != nil {
		// Check for specific errors
		if err == service.ErrApplicationNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "지원 정보를 찾을 수 없습니다",
			})
			return
		}
		if err == service.ErrApplicationForbidden {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "접근 권한이 없습니다",
			})
			return
		}

		// AI failure or other errors
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI 분석 중 오류가 발생했습니다: " + err.Error(),
		})
		return
	}

	TrackUsageAfterSuccess(ctx, map[string]interface{}{"application_id": req.ApplicationID})
	ctx.JSON(http.StatusOK, result)
}

// PostRecommendExperiences handles POST /v1/coaching/recommend-experiences
func (c *QuestionController) PostRecommendExperiences(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	var req service.RecommendInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "입력 값이 올바르지 않습니다: " + err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	recommendations, err := c.questionService.RecommendExperiences(
		ctx.Request.Context(),
		userID.(uuid.UUID),
		req,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "경험 추천 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}
