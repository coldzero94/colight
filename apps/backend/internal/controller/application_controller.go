package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ApplicationController handles dashboard application endpoints.
type ApplicationController struct {
	applicationService *service.ApplicationService
}

// NewApplicationController creates a new application controller.
func NewApplicationController(applicationService *service.ApplicationService) *ApplicationController {
	return &ApplicationController{applicationService: applicationService}
}

// ListApplications handles GET /v1/applications
func (c *ApplicationController) ListApplications(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	apps, err := c.applicationService.ListApplications(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "지원 목록 조회 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"applications": apps})
}

// UpdateStatus handles PATCH /v1/applications/:id/status
func (c *ApplicationController) UpdateStatus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	appID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다."})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status 필드가 필요합니다."})
		return
	}

	result, err := c.applicationService.UpdateStatus(ctx.Request.Context(), userID.(uuid.UUID), appID, req.Status)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "지원 정보를 찾을 수 없습니다."})
			return
		}
		if err == service.ErrApplicationForbidden {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다."})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "상태 변경 실패"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetStats handles GET /v1/applications/stats
func (c *ApplicationController) GetStats(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	stats, err := c.applicationService.GetStats(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "통계 조회 실패"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// CreateApplication handles POST /v1/applications
func (c *ApplicationController) CreateApplication(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	var input service.CreateApplicationInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "요청 형식이 올바르지 않습니다."})
		return
	}

	result, err := c.applicationService.CreateApplication(ctx.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

// UpdateApplication handles PATCH /v1/applications/:id
func (c *ApplicationController) UpdateApplication(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	appID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다."})
		return
	}

	var input service.UpdateApplicationInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "요청 형식이 올바르지 않습니다."})
		return
	}

	result, err := c.applicationService.UpdateApplication(ctx.Request.Context(), userID.(uuid.UUID), appID, input)
	if err != nil {
		switch err {
		case service.ErrApplicationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "지원 정보를 찾을 수 없습니다."})
		case service.ErrApplicationForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다."})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "수정 실패"})
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// DeleteApplication handles DELETE /v1/applications/:id
func (c *ApplicationController) DeleteApplication(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	appID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다."})
		return
	}

	err = c.applicationService.DeleteApplication(ctx.Request.Context(), userID.(uuid.UUID), appID)
	if err != nil {
		switch err {
		case service.ErrApplicationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "지원 정보를 찾을 수 없습니다."})
		case service.ErrApplicationForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다."})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "삭제 실패"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

// LinkAnalysis handles POST /v1/applications/:id/link-analysis
func (c *ApplicationController) LinkAnalysis(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	appID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 ID입니다."})
		return
	}

	var req struct {
		AnalysisID string `json:"analysis_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "analysis_id 필드가 필요합니다."})
		return
	}

	analysisID, err := uuid.Parse(req.AnalysisID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 analysis_id입니다."})
		return
	}

	result, err := c.applicationService.LinkAnalysis(ctx.Request.Context(), userID.(uuid.UUID), appID, analysisID)
	if err != nil {
		switch err {
		case service.ErrApplicationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "지원 정보를 찾을 수 없습니다."})
		case service.ErrApplicationForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다."})
		case service.ErrAnalysisNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "기업 분석을 찾을 수 없습니다."})
		case service.ErrAnalysisForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "기업 분석 접근 권한이 없습니다."})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "연결 실패"})
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// SearchApplications handles GET /v1/applications/search
func (c *ApplicationController) SearchApplications(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다."})
		return
	}

	input := service.SearchApplicationsInput{
		Query: ctx.Query("q"),
	}

	if tagsParam := ctx.Query("tags"); tagsParam != "" {
		input.Tags = strings.Split(tagsParam, ",")
	}

	if from := ctx.Query("deadline_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			input.DeadlineFrom = &t
		}
	}
	if to := ctx.Query("deadline_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			input.DeadlineTo = &t
		}
	}

	results, err := c.applicationService.SearchApplications(ctx.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "검색 실패"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"applications": results})
}
