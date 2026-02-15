package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExperienceController struct {
	experienceService *service.ExperienceService
}

func NewExperienceController(experienceService *service.ExperienceService) *ExperienceController {
	return &ExperienceController{experienceService: experienceService}
}

// Create handles POST /v1/experiences
func (ctrl *ExperienceController) Create(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	var req struct {
		Title         string   `json:"title" binding:"required"`
		Category      string   `json:"category"`
		PeriodStart   *string  `json:"period_start"`
		PeriodEnd     *string  `json:"period_end"`
		Role          string   `json:"role"`
		Content       string   `json:"content"`
		Result        string   `json:"result"`
		StarSituation string   `json:"star_situation"`
		StarTask      string   `json:"star_task"`
		StarAction    string   `json:"star_action"`
		StarResult    string   `json:"star_result"`
		Keywords      []string `json:"keywords"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "입력값이 올바르지 않습니다.", "code": "VALID_001"},
		})
		return
	}

	input := service.CreateExperienceInput{
		Title:         req.Title,
		Category:      req.Category,
		Role:          req.Role,
		Content:       req.Content,
		Result:        req.Result,
		StarSituation: req.StarSituation,
		StarTask:      req.StarTask,
		StarAction:    req.StarAction,
		StarResult:    req.StarResult,
		Keywords:      req.Keywords,
	}

	if req.PeriodStart != nil {
		t, err := parseDate(*req.PeriodStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": "시작일 형식이 올바르지 않습니다.", "code": "VALID_002"},
			})
			return
		}
		input.PeriodStart = &t
	}
	if req.PeriodEnd != nil {
		t, err := parseDate(*req.PeriodEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": "종료일 형식이 올바르지 않습니다.", "code": "VALID_002"},
			})
			return
		}
		input.PeriodEnd = &t
	}

	exp, err := ctrl.experienceService.CreateExperience(c.Request.Context(), userID, input)
	if err != nil {
		handleExperienceError(c, err)
		return
	}

	TrackUsageAfterSuccess(c, map[string]interface{}{"experience_id": exp.ID.String()})
	c.JSON(http.StatusCreated, gin.H{"id": exp.ID.String()})
}

// List handles GET /v1/experiences
func (ctrl *ExperienceController) List(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	params := service.ExperienceListParams{
		Sort:       c.Query("sort"),
		Category:   c.Query("category"),
		WeaponCode: c.Query("weapon"),
	}

	exps, err := ctrl.experienceService.GetExperiences(c.Request.Context(), userID, params)
	if err != nil {
		slog.Error("list experiences failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "경험 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, len(exps))
	for i, exp := range exps {
		items[i] = toExperienceResponse(exp)
	}

	c.JSON(http.StatusOK, gin.H{"experiences": items})
}

// Get handles GET /v1/experiences/:id
func (ctrl *ExperienceController) Get(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "유효하지 않은 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	exp, err := ctrl.experienceService.GetExperience(c.Request.Context(), userID, id)
	if err != nil {
		handleExperienceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toExperienceResponse(exp))
}

// Update handles PATCH /v1/experiences/:id
func (ctrl *ExperienceController) Update(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "유효하지 않은 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		Title         *string  `json:"title"`
		Category      *string  `json:"category"`
		PeriodStart   *string  `json:"period_start"`
		PeriodEnd     *string  `json:"period_end"`
		Role          *string  `json:"role"`
		Content       *string  `json:"content"`
		Result        *string  `json:"result"`
		StarSituation *string  `json:"star_situation"`
		StarTask      *string  `json:"star_task"`
		StarAction    *string  `json:"star_action"`
		StarResult    *string  `json:"star_result"`
		Keywords      []string `json:"keywords"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "입력값이 올바르지 않습니다.", "code": "VALID_001"},
		})
		return
	}

	input := service.UpdateExperienceInput{
		Title:         req.Title,
		Category:      req.Category,
		Role:          req.Role,
		Content:       req.Content,
		Result:        req.Result,
		StarSituation: req.StarSituation,
		StarTask:      req.StarTask,
		StarAction:    req.StarAction,
		StarResult:    req.StarResult,
		Keywords:      req.Keywords,
	}

	if req.PeriodStart != nil {
		t, err := parseDate(*req.PeriodStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": "시작일 형식이 올바르지 않습니다.", "code": "VALID_002"},
			})
			return
		}
		input.PeriodStart = &t
	}
	if req.PeriodEnd != nil {
		t, err := parseDate(*req.PeriodEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": "종료일 형식이 올바르지 않습니다.", "code": "VALID_002"},
			})
			return
		}
		input.PeriodEnd = &t
	}

	exp, err := ctrl.experienceService.UpdateExperience(c.Request.Context(), userID, id, input)
	if err != nil {
		handleExperienceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": exp.ID.String()})
}

// Delete handles DELETE /v1/experiences/:id
func (ctrl *ExperienceController) Delete(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "유효하지 않은 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	if err := ctrl.experienceService.DeleteExperience(c.Request.Context(), userID, id); err != nil {
		handleExperienceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return uid
}

func handleExperienceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrExperienceNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"message": "경험을 찾을 수 없습니다.", "code": "EXP_001"},
		})
	case errors.Is(err, service.ErrExperienceForbidden):
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"message": "접근 권한이 없습니다.", "code": "EXP_002"},
		})
	case errors.Is(err, service.ErrInvalidTitle):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "제목은 1~200자 이내로 입력해주세요.", "code": "VALID_001"},
		})
	case errors.Is(err, service.ErrInvalidPeriod):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "종료일은 시작일 이후여야 합니다.", "code": "VALID_002"},
		})
	default:
		slog.Error("experience operation failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "처리 중 오류가 발생했습니다.", "code": "SYS_001"},
		})
	}
}
