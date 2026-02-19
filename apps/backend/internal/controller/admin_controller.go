package controller

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"context"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/adminauditlog"
	"github.com/coby/colight/apps/backend/ent/deletionrequest"
	"github.com/coby/colight/apps/backend/ent/feedback"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/systemconfig"
	"github.com/coby/colight/apps/backend/ent/aicallerror"
	"github.com/coby/colight/apps/backend/ent/predicate"
	"github.com/coby/colight/apps/backend/ent/usagelog"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crypto"
	"github.com/coby/colight/apps/backend/internal/infrastructure/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminController struct {
	db         *ent.Client
	aiProvider *ai.AIProvider
}

func NewAdminController(db *ent.Client, aiProvider *ai.AIProvider) *AdminController {
	return &AdminController{db: db, aiProvider: aiProvider}
}

// ListUsers returns paginated user list with optional role/search filter.
// GET /v1/admin/users
func (ctrl *AdminController) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	roleFilter := c.Query("role")
	search := c.Query("search")

	query := ctrl.db.UserProfile.Query()

	if roleFilter != "" {
		query = query.Where(userprofile.RoleEQ(userprofile.Role(roleFilter)))
	}
	if search != "" {
		query = query.Where(
			userprofile.Or(
				userprofile.NicknameContains(search),
				userprofile.EmailContains(search),
			),
		)
	}

	total, err := query.Clone().Count(c.Request.Context())
	if err != nil {
		slog.Error("list users count failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용자 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	users, err := query.
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(userprofile.FieldCreatedAt)).
		All(c.Request.Context())
	if err != nil {
		slog.Error("list users failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용자 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(users))
	for _, u := range users {
		item := map[string]any{
			"id":            u.ID.String(),
			"auth_provider": string(u.AuthProvider),
			"role":          string(u.Role),
			"plan":          string(u.Plan),
			"suspended":     u.Suspended,
			"created_at":    u.CreatedAt,
		}
		if u.Email != nil {
			item["email"] = *u.Email
		}
		if u.Nickname != "" {
			item["nickname"] = u.Nickname
		}
		if u.LastLoginAt != nil {
			item["last_login_at"] = *u.LastLoginAt
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
	})
}

// GetUser returns a single user by ID.
// GET /v1/admin/users/:id
func (ctrl *AdminController) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	user, err := ctrl.db.UserProfile.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("get user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용자 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, toUserInfo(user))
}

// UpdateUserRole changes a user's role with hierarchy validation.
// PUT /v1/admin/users/:id/role
func (ctrl *AdminController) UpdateUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=user manager admin super_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "올바른 역할을 지정해주세요. (user, manager, admin, super_admin)", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	// Self-change protection
	callerID, _ := c.Get("user_id")
	if callerID != nil {
		if uid, ok := callerID.(uuid.UUID); ok && uid == id {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{"message": "자기 자신의 역할은 변경할 수 없습니다.", "code": "AUTH_003"},
			})
			return
		}
	}

	// Check caller's role level vs target role level
	callerRole, _ := c.Get("role")
	callerRoleStr, _ := callerRole.(string)
	callerLevel := middleware.RoleLevel(callerRoleStr)
	targetLevel := middleware.RoleLevel(req.Role)

	// Caller can only assign roles strictly below their own level
	if targetLevel >= callerLevel {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"message": "자신의 권한 이상으로 역할을 변경할 수 없습니다.", "code": "AUTH_003"},
		})
		return
	}

	// Check the current role of the target user
	target, err := ctrl.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("get target user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "역할 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Cannot modify users at or above caller's level (except super_admin can modify other super_admins)
	targetCurrentLevel := middleware.RoleLevel(string(target.Role))
	if targetCurrentLevel >= callerLevel && !(callerRoleStr == "super_admin" && string(target.Role) == "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"message": "자신의 권한 이상의 사용자를 변경할 수 없습니다.", "code": "AUTH_003"},
		})
		return
	}

	// Last super_admin protection
	if target.Role == userprofile.RoleSuperAdmin {
		count, err := ctrl.db.UserProfile.Query().
			Where(userprofile.RoleEQ(userprofile.RoleSuperAdmin)).
			Count(ctx)
		if err == nil && count <= 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{"message": "마지막 최고 관리자는 강등할 수 없습니다.", "code": "AUTH_003"},
			})
			return
		}
	}

	oldRole := string(target.Role)
	user, err := ctrl.db.UserProfile.UpdateOneID(id).
		SetRole(userprofile.Role(req.Role)).
		Save(ctx)
	if err != nil {
		slog.Error("update user role failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "역할 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Audit log
	if uid, ok := callerID.(uuid.UUID); ok {
		ctrl.recordAuditLog(ctx, uid, "role_change", "user", id.String(), &oldRole, &req.Role)
	}

	c.JSON(http.StatusOK, toUserInfo(user))
}

// GetStats returns system-level statistics.
// GET /v1/admin/stats
func (ctrl *AdminController) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	totalUsers, _ := ctrl.db.UserProfile.Query().Count(ctx)
	emailUsers, _ := ctrl.db.UserProfile.Query().Where(userprofile.AuthProviderEQ(userprofile.AuthProviderEmail)).Count(ctx)
	naverUsers, _ := ctrl.db.UserProfile.Query().Where(userprofile.AuthProviderEQ(userprofile.AuthProviderNaver)).Count(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	activeToday, _ := ctrl.db.UserProfile.Query().Where(userprofile.LastLoginAtGTE(today)).Count(ctx)

	totalExperiences, _ := ctrl.db.Experience.Query().Count(ctx)

	c.JSON(http.StatusOK, gin.H{
		"total_users":        totalUsers,
		"email_auth_users":   emailUsers,
		"naver_auth_users":   naverUsers,
		"active_users_today": activeToday,
		"total_experiences":  totalExperiences,
	})
}

// ListPrompts returns prompt templates with optional category filter.
// GET /v1/admin/prompts
func (ctrl *AdminController) ListPrompts(c *gin.Context) {
	query := ctrl.db.PromptTemplate.Query()

	if cat := c.Query("category"); cat != "" {
		query = query.Where(prompttemplate.CategoryEQ(cat))
	}

	prompts, err := query.
		Order(ent.Asc(prompttemplate.FieldCategory, prompttemplate.FieldSubCategory)).
		All(c.Request.Context())
	if err != nil {
		slog.Error("list prompts failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "프롬프트 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(prompts))
	for _, p := range prompts {
		items = append(items, map[string]any{
			"id":                   p.ID.String(),
			"category":             p.Category,
			"sub_category":         p.SubCategory,
			"name":                 p.Name,
			"system_prompt":        p.SystemPrompt,
			"user_prompt_template": p.UserPromptTemplate,
			"model_name":           p.Model,
			"temperature":          p.Temperature,
			"max_tokens":           p.MaxTokens,
			"version":              p.Version,
			"is_active":            p.IsActive,
			"usage_count":          p.UsageCount,
			"avg_latency_ms":       p.AvgLatencyMs,
			"avg_quality_score":    p.AvgQualityScore,
			"output_schema":        p.OutputSchema,
			"created_at":           p.CreatedAt,
			"updated_at":           p.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListConfigs returns system configuration entries with optional category filter.
// Secret values are masked. GET /v1/admin/configs
func (ctrl *AdminController) ListConfigs(c *gin.Context) {
	query := ctrl.db.SystemConfig.Query()

	if cat := c.Query("category"); cat != "" {
		query = query.Where(systemconfig.CategoryEQ(cat))
	}

	configs, err := query.
		Order(ent.Asc(systemconfig.FieldCategory, systemconfig.FieldConfigKey)).
		All(c.Request.Context())
	if err != nil {
		slog.Error("list configs failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "설정 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(configs))
	for _, cfg := range configs {
		value := cfg.ConfigValue
		if cfg.IsSecret {
			value = crypto.Mask(value)
		}
		item := map[string]any{
			"id":           cfg.ID.String(),
			"config_key":   cfg.ConfigKey,
			"config_value": value,
			"category":     cfg.Category,
			"is_secret":    cfg.IsSecret,
		}
		if cfg.Description != "" {
			item["description"] = cfg.Description
		}
		if cfg.UpdatedBy != nil {
			item["updated_by"] = cfg.UpdatedBy.String()
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// UpdateConfig updates a system configuration value by key.
// PUT /v1/admin/configs/:key
func (ctrl *AdminController) UpdateConfig(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "값을 입력해주세요.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	// Find config by key
	cfg, err := ctrl.db.SystemConfig.Query().
		Where(systemconfig.ConfigKeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "설정을 찾을 수 없습니다.", "code": "SYS_002"},
			})
			return
		}
		slog.Error("get config failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "설정 수정에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	oldValue := cfg.ConfigValue
	update := ctrl.db.SystemConfig.UpdateOne(cfg).
		SetConfigValue(req.Value)

	// Track who updated
	callerID, _ := c.Get("user_id")
	if uid, ok := callerID.(uuid.UUID); ok {
		update = update.SetUpdatedBy(uid)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		slog.Error("update config failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "설정 수정에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Audit log
	if uid, ok := callerID.(uuid.UUID); ok {
		ctrl.recordAuditLog(ctx, uid, "config_update", "config", key, &oldValue, &req.Value)
	}

	value := updated.ConfigValue
	if updated.IsSecret {
		value = crypto.Mask(value)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           updated.ID.String(),
		"config_key":   updated.ConfigKey,
		"config_value": value,
		"category":     updated.Category,
		"is_secret":    updated.IsSecret,
	})
}

// UpdatePrompt updates a prompt template's fields.
// PUT /v1/admin/prompts/:id
func (ctrl *AdminController) UpdatePrompt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 프롬프트 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		SystemPrompt       *string  `json:"system_prompt"`
		UserPromptTemplate *string  `json:"user_prompt_template"`
		ModelName          *string  `json:"model_name"`
		Temperature        *float64 `json:"temperature"`
		MaxTokens          *int     `json:"max_tokens"`
		IsActive           *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "입력값이 올바르지 않습니다.", "code": "VALID_001"},
		})
		return
	}

	update := ctrl.db.PromptTemplate.UpdateOneID(id)
	if req.SystemPrompt != nil {
		update = update.SetSystemPrompt(*req.SystemPrompt)
	}
	if req.UserPromptTemplate != nil {
		update = update.SetUserPromptTemplate(*req.UserPromptTemplate)
	}
	if req.ModelName != nil {
		if !ai.IsSelectableModel(*req.ModelName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "지원하지 않는 모델입니다. GET /v1/admin/models/available에서 사용 가능한 모델을 확인해주세요.",
					"code":    "VALID_001",
				},
			})
			return
		}
		update = update.SetModel(*req.ModelName)
	}
	if req.Temperature != nil {
		update = update.SetTemperature(*req.Temperature)
	}
	if req.MaxTokens != nil {
		update = update.SetMaxTokens(*req.MaxTokens)
	}
	if req.IsActive != nil {
		update = update.SetIsActive(*req.IsActive)
	}

	prompt, err := update.Save(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "프롬프트를 찾을 수 없습니다.", "code": "SYS_002"},
			})
			return
		}
		slog.Error("update prompt failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "프롬프트 수정에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                   prompt.ID.String(),
		"category":             prompt.Category,
		"sub_category":         prompt.SubCategory,
		"name":                 prompt.Name,
		"system_prompt":        prompt.SystemPrompt,
		"user_prompt_template": prompt.UserPromptTemplate,
		"model_name":           prompt.Model,
		"temperature":          prompt.Temperature,
		"max_tokens":           prompt.MaxTokens,
		"version":              prompt.Version,
		"is_active":            prompt.IsActive,
		"usage_count":          prompt.UsageCount,
		"avg_latency_ms":       prompt.AvgLatencyMs,
		"avg_quality_score":    prompt.AvgQualityScore,
		"output_schema":        prompt.OutputSchema,
		"created_at":           prompt.CreatedAt,
		"updated_at":           prompt.UpdatedAt,
	})
}

// GetUsageSummary returns aggregated usage statistics for a given period.
// GET /v1/admin/usage/summary
func (ctrl *AdminController) GetUsageSummary(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	ctx := c.Request.Context()
	since := time.Now().AddDate(0, 0, -days)

	logs, err := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		slog.Error("get usage summary failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용량 요약 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	totalCalls := len(logs)
	var totalTokens int
	var totalCost float64
	var errorCount int

	for _, l := range logs {
		totalTokens += l.TotalTokens
		if l.EstimatedCostKrw != nil {
			totalCost += *l.EstimatedCostKrw
		}
		if l.Status == "error" {
			errorCount++
		}
	}

	var errorRate float64
	if totalCalls > 0 {
		errorRate = float64(errorCount) / float64(totalCalls) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total_calls":   totalCalls,
		"total_tokens":  totalTokens,
		"total_cost_krw": totalCost,
		"error_rate":    errorRate,
		"error_count":   errorCount,
		"days":          days,
	})
}

// GetUsageDaily returns daily usage breakdown.
// GET /v1/admin/usage/daily
func (ctrl *AdminController) GetUsageDaily(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 {
		days = 7
	}

	ctx := c.Request.Context()
	since := time.Now().AddDate(0, 0, -days)

	logs, err := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		slog.Error("get usage daily failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "일별 사용량 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Group by date
	dailyMap := make(map[string]map[string]any)
	for _, l := range logs {
		date := l.CreatedAt.Format("2006-01-02")
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = map[string]any{
				"date":        date,
				"total_calls": 0,
				"total_tokens": 0,
				"error_count": 0,
			}
		}
		d := dailyMap[date]
		d["total_calls"] = d["total_calls"].(int) + 1
		d["total_tokens"] = d["total_tokens"].(int) + l.TotalTokens
		if l.Status == "error" {
			d["error_count"] = d["error_count"].(int) + 1
		}
	}

	items := make([]map[string]any, 0, len(dailyMap))
	for _, v := range dailyMap {
		items = append(items, v)
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// SuspendUser suspends a user account.
// POST /v1/admin/users/:id/suspend
func (ctrl *AdminController) SuspendUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()

	user, err := ctrl.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("get user for suspend failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "계정 정지에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	if user.Suspended {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"message": "이미 정지된 계정입니다.", "code": "VALID_001"},
		})
		return
	}

	now := time.Now()
	update := ctrl.db.UserProfile.UpdateOne(user).
		SetSuspended(true).
		SetSuspendedAt(now)
	if req.Reason != "" {
		update = update.SetSuspendedReason(req.Reason)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		slog.Error("suspend user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "계정 정지에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, toUserInfo(updated))
}

// UnsuspendUser lifts suspension on a user account.
// DELETE /v1/admin/users/:id/suspend
func (ctrl *AdminController) UnsuspendUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	updated, err := ctrl.db.UserProfile.UpdateOneID(id).
		SetSuspended(false).
		ClearSuspendedAt().
		ClearSuspendedReason().
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("unsuspend user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "정지 해제에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, toUserInfo(updated))
}

// GetUserDetail returns detailed information about a user including counts.
// GET /v1/admin/users/:id/detail
func (ctrl *AdminController) GetUserDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	user, err := ctrl.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("get user detail failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용자 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	expCount, _ := user.QueryExperiences().Count(ctx)
	coachCount, _ := user.QueryCoachingSessions().Count(ctx)
	usageCount, _ := user.QueryUsageLogs().Count(ctx)

	info := toUserInfo(user)
	info["experience_count"] = expCount
	info["coaching_count"] = coachCount
	info["usage_count"] = usageCount
	info["suspended"] = user.Suspended
	if user.SuspendedAt != nil {
		info["suspended_at"] = *user.SuspendedAt
	}
	if user.SuspendedReason != nil {
		info["suspended_reason"] = *user.SuspendedReason
	}

	c.JSON(http.StatusOK, info)
}

// recordAuditLog saves an admin action to the audit log.
func (ctrl *AdminController) recordAuditLog(ctx context.Context, adminID uuid.UUID, action, targetType, targetID string, oldValue, newValue *string) {
	create := ctrl.db.AdminAuditLog.Create().
		SetAdminID(adminID).
		SetAction(action).
		SetTargetType(targetType).
		SetTargetID(targetID)
	if oldValue != nil {
		create = create.SetOldValue(*oldValue)
	}
	if newValue != nil {
		create = create.SetNewValue(*newValue)
	}
	if err := create.Exec(ctx); err != nil {
		slog.Error("record audit log failed", "error", err, "action", action)
	}
}

// ListAuditLogs returns paginated audit log entries.
// GET /v1/admin/audit-logs
func (ctrl *AdminController) ListAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	actionFilter := c.Query("action")

	query := ctrl.db.AdminAuditLog.Query()
	if actionFilter != "" {
		query = query.Where(adminauditlog.ActionEQ(actionFilter))
	}

	total, err := query.Clone().Count(c.Request.Context())
	if err != nil {
		slog.Error("list audit logs count failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "감사 로그 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	logs, err := query.
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(adminauditlog.FieldCreatedAt)).
		All(c.Request.Context())
	if err != nil {
		slog.Error("list audit logs failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "감사 로그 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(logs))
	for _, l := range logs {
		item := map[string]any{
			"id":          l.ID.String(),
			"admin_id":    l.AdminID.String(),
			"action":      l.Action,
			"target_type": l.TargetType,
			"target_id":   l.TargetID,
			"created_at":  l.CreatedAt,
		}
		if l.OldValue != nil {
			item["old_value"] = *l.OldValue
		}
		if l.NewValue != nil {
			item["new_value"] = *l.NewValue
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
	})
}

// HealthCheck returns system health status including database connectivity.
// GET /v1/admin/health
func (ctrl *AdminController) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	dbStatus := "healthy"
	if _, err := ctrl.db.UserProfile.Query().Count(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"database": dbStatus,
	})
}

// GetUsageCosts returns provider-level cost breakdown.
// GET /v1/admin/usage/costs
func (ctrl *AdminController) GetUsageCosts(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	ctx := c.Request.Context()
	since := time.Now().AddDate(0, 0, -days)

	logs, err := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		slog.Error("get usage costs failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "비용 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	type providerCost struct {
		TotalTokens int     `json:"total_tokens"`
		TotalCost   float64 `json:"total_cost_krw"`
		CallCount   int     `json:"call_count"`
	}
	providerMap := make(map[string]*providerCost)
	for _, l := range logs {
		p := "unknown"
		if l.Provider != nil {
			p = *l.Provider
		}
		if _, ok := providerMap[p]; !ok {
			providerMap[p] = &providerCost{}
		}
		pc := providerMap[p]
		pc.TotalTokens += l.TotalTokens
		if l.EstimatedCostKrw != nil {
			pc.TotalCost += *l.EstimatedCostKrw
		}
		pc.CallCount++
	}

	items := make([]map[string]any, 0, len(providerMap))
	for provider, pc := range providerMap {
		items = append(items, map[string]any{
			"provider":      provider,
			"total_tokens":  pc.TotalTokens,
			"total_cost_krw": pc.TotalCost,
			"call_count":    pc.CallCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "days": days})
}

// GetUsageTopUsers returns top users ranked by token usage.
// GET /v1/admin/usage/top-users
func (ctrl *AdminController) GetUsageTopUsers(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	ctx := c.Request.Context()
	since := time.Now().AddDate(0, 0, -days)

	logs, err := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		slog.Error("get usage top users failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "상위 사용자 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	type userUsage struct {
		UserID      uuid.UUID
		TotalTokens int
		TotalCost   float64
		CallCount   int
	}
	userMap := make(map[uuid.UUID]*userUsage)
	for _, l := range logs {
		if _, ok := userMap[l.UserID]; !ok {
			userMap[l.UserID] = &userUsage{UserID: l.UserID}
		}
		uu := userMap[l.UserID]
		uu.TotalTokens += l.TotalTokens
		if l.EstimatedCostKrw != nil {
			uu.TotalCost += *l.EstimatedCostKrw
		}
		uu.CallCount++
	}

	// Sort by total tokens descending
	sorted := make([]*userUsage, 0, len(userMap))
	for _, uu := range userMap {
		sorted = append(sorted, uu)
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].TotalTokens > sorted[i].TotalTokens {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > len(sorted) {
		limit = len(sorted)
	}

	items := make([]map[string]any, 0, limit)
	for _, uu := range sorted[:limit] {
		items = append(items, map[string]any{
			"user_id":       uu.UserID.String(),
			"total_tokens":  uu.TotalTokens,
			"total_cost_krw": uu.TotalCost,
			"call_count":    uu.CallCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "days": days})
}

// ForceLogout invalidates all tokens for a user by setting force_logout_at.
// POST /v1/admin/users/:id/force-logout
func (ctrl *AdminController) ForceLogout(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()
	now := time.Now()

	updated, err := ctrl.db.UserProfile.UpdateOneID(id).
		SetForceLogoutAt(now).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("force logout failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "강제 로그아웃에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	info := toUserInfo(updated)
	if updated.ForceLogoutAt != nil {
		info["force_logout_at"] = *updated.ForceLogoutAt
	}

	c.JSON(http.StatusOK, info)
}

// UpdateUserPlan changes a user's subscription plan.
// PUT /v1/admin/users/:id/plan
func (ctrl *AdminController) UpdateUserPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		Plan string `json:"plan" binding:"required,oneof=free starter pro season"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "올바른 플랜을 지정해주세요. (free, starter, pro, season)", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	updated, err := ctrl.db.UserProfile.UpdateOneID(id).
		SetPlan(userprofile.Plan(req.Plan)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("update user plan failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "플랜 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	info := toUserInfo(updated)
	info["plan"] = string(updated.Plan)

	c.JSON(http.StatusOK, info)
}

// UpdateFeedbackStatus updates the admin review status of a feedback entry.
// PUT /v1/admin/feedbacks/:id
func (ctrl *AdminController) UpdateFeedbackStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 피드백 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		AdminStatus string `json:"admin_status" binding:"required,oneof=pending reviewed resolved dismissed"`
		AdminNote   string `json:"admin_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "올바른 상태를 지정해주세요.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()
	now := time.Now()

	update := ctrl.db.Feedback.UpdateOneID(id).
		SetAdminStatus(feedback.AdminStatus(req.AdminStatus)).
		SetReviewedAt(now)

	if req.AdminNote != "" {
		update = update.SetAdminNote(req.AdminNote)
	}

	callerID, _ := c.Get("user_id")
	if uid, ok := callerID.(uuid.UUID); ok {
		update = update.SetReviewedBy(uid)
	}

	fb, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "피드백을 찾을 수 없습니다.", "code": "SYS_002"},
			})
			return
		}
		slog.Error("update feedback status failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "피드백 상태 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	resp := map[string]any{
		"id":           fb.ID.String(),
		"user_id":      fb.UserID.String(),
		"category":     string(fb.Category),
		"content":      fb.Content,
		"admin_status": string(fb.AdminStatus),
		"created_at":   fb.CreatedAt,
	}
	if fb.AdminNote != "" {
		resp["admin_note"] = fb.AdminNote
	}
	if fb.ReviewedBy != nil {
		resp["reviewed_by"] = fb.ReviewedBy.String()
	}
	if fb.ReviewedAt != nil {
		resp["reviewed_at"] = *fb.ReviewedAt
	}

	c.JSON(http.StatusOK, resp)
}

// ExportUserData exports all user data for PIPA compliance.
// POST /v1/admin/users/:id/export
func (ctrl *AdminController) ExportUserData(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	user, err := ctrl.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("export user data failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "데이터 내보내기에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	experiences, _ := user.QueryExperiences().All(ctx)
	expList := make([]map[string]any, 0, len(experiences))
	for _, exp := range experiences {
		expList = append(expList, map[string]any{
			"id":      exp.ID.String(),
			"title":   exp.Title,
			"content": exp.Content,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"profile":     toUserInfo(user),
		"experiences": expList,
	})
}

// CreateDeletionRequest creates a data deletion request for a user.
// POST /v1/admin/users/:id/delete-request
func (ctrl *AdminController) CreateDeletionRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()

	// Verify user exists
	if _, err := ctrl.db.UserProfile.Get(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("get user for deletion request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "삭제 요청 생성에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	callerID, _ := c.Get("user_id")
	adminID, _ := callerID.(uuid.UUID)

	scheduledAt := time.Now().AddDate(0, 0, 30)

	create := ctrl.db.DeletionRequest.Create().
		SetUserID(id).
		SetRequestedBy(adminID).
		SetScheduledAt(scheduledAt)
	if req.Reason != "" {
		create = create.SetReason(req.Reason)
	}

	dr, err := create.Save(ctx)
	if err != nil {
		slog.Error("create deletion request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "삭제 요청 생성에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           dr.ID.String(),
		"user_id":      dr.UserID.String(),
		"status":       string(dr.Status),
		"reason":       dr.Reason,
		"scheduled_at": dr.ScheduledAt,
		"requested_by": dr.RequestedBy.String(),
		"created_at":   dr.CreatedAt,
	})
}

// CancelDeletionRequest cancels a pending deletion request for a user.
// DELETE /v1/admin/users/:id/delete-request
func (ctrl *AdminController) CancelDeletionRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "잘못된 사용자 ID입니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	dr, err := ctrl.db.DeletionRequest.Query().
		Where(
			deletionrequest.UserIDEQ(id),
			deletionrequest.StatusEQ(deletionrequest.StatusPending),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "대기 중인 삭제 요청이 없습니다.", "code": "SYS_002"},
			})
			return
		}
		slog.Error("find deletion request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "삭제 요청 취소에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	now := time.Now()
	updated, err := ctrl.db.DeletionRequest.UpdateOne(dr).
		SetStatus(deletionrequest.StatusCancelled).
		SetCancelledAt(now).
		Save(ctx)
	if err != nil {
		slog.Error("cancel deletion request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "삭제 요청 취소에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           updated.ID.String(),
		"user_id":      updated.UserID.String(),
		"status":       string(updated.Status),
		"cancelled_at": updated.CancelledAt,
	})
}

// ListDeletionQueue returns pending deletion requests.
// GET /v1/admin/deletion-queue
func (ctrl *AdminController) ListDeletionQueue(c *gin.Context) {
	ctx := c.Request.Context()

	requests, err := ctrl.db.DeletionRequest.Query().
		Where(deletionrequest.StatusEQ(deletionrequest.StatusPending)).
		Order(ent.Asc(deletionrequest.FieldScheduledAt)).
		All(ctx)
	if err != nil {
		slog.Error("list deletion queue failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "삭제 대기열 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(requests))
	for _, dr := range requests {
		items = append(items, map[string]any{
			"id":           dr.ID.String(),
			"user_id":      dr.UserID.String(),
			"status":       string(dr.Status),
			"reason":       dr.Reason,
			"scheduled_at": dr.ScheduledAt,
			"requested_by": dr.RequestedBy.String(),
			"created_at":   dr.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListFeedbacks returns paginated feedback entries with optional category filter.
// GET /v1/admin/feedbacks
func (ctrl *AdminController) ListFeedbacks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	categoryFilter := c.Query("category")

	query := ctrl.db.Feedback.Query()
	if categoryFilter != "" {
		query = query.Where(feedback.CategoryEQ(feedback.Category(categoryFilter)))
	}

	total, err := query.Clone().Count(c.Request.Context())
	if err != nil {
		slog.Error("list feedbacks count failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "피드백 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	feedbacks, err := query.
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(feedback.FieldCreatedAt)).
		All(c.Request.Context())
	if err != nil {
		slog.Error("list feedbacks failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "피드백 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(feedbacks))
	for _, fb := range feedbacks {
		item := map[string]any{
			"id":         fb.ID.String(),
			"user_id":    fb.UserID.String(),
			"category":   string(fb.Category),
			"content":    fb.Content,
			"created_at": fb.CreatedAt,
		}
		if fb.PageURL != "" {
			item["page_url"] = fb.PageURL
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
	})
}

// CreateUser creates a new user (admin only).
// POST /v1/admin/users
func (ctrl *AdminController) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Nickname string `json:"nickname"`
		Role     string `json:"role" binding:"required,oneof=user manager admin super_admin"`
		Plan     string `json:"plan" binding:"oneof=free starter pro season"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "입력값이 올바르지 않습니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()

	// Check if email already exists
	exists, _ := ctrl.db.UserProfile.Query().
		Where(userprofile.EmailEQ(req.Email)).
		Exist(ctx)
	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"message": "이미 존재하는 이메일입니다.", "code": "AUTH_002"},
		})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		slog.Error("hash password failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용자 생성에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Create user
	builder := ctrl.db.UserProfile.Create().
		SetEmail(req.Email).
		SetPasswordHash(string(hash)).
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.Role(req.Role))
	if req.Nickname != "" {
		builder = builder.SetNickname(req.Nickname)
	}
	if req.Plan != "" {
		builder = builder.SetPlan(userprofile.Plan(req.Plan))
	}

	user, err := builder.Save(ctx)
	if err != nil {
		slog.Error("create user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "사용자 생성에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusCreated, toUserInfo(user))
}

// GetModelStats returns model-level statistics with quota estimates
// GET /v1/admin/models/stats?days=30
func (ctrl *AdminController) GetModelStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	ctx := c.Request.Context()
	since := time.Now().AddDate(0, 0, -days)

	logs, err := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		slog.Error("get model stats failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "모델 통계 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Aggregate by model
	type modelStats struct {
		Model        string
		Provider     string
		CallCount    int
		SuccessCount int
		ErrorCount   int
		TotalTokens  int
		InputTokens  int
		OutputTokens int
		TotalCost    float64
		AvgLatency   float64
		LastUsed     time.Time
	}

	modelMap := make(map[string]*modelStats)
	for _, l := range logs {
		model := "unknown"
		if l.Model != nil {
			model = *l.Model
		}
		provider := "unknown"
		if l.Provider != nil {
			provider = *l.Provider
		}

		if _, ok := modelMap[model]; !ok {
			modelMap[model] = &modelStats{
				Model:    model,
				Provider: provider,
			}
		}
		ms := modelMap[model]
		ms.CallCount++
		if l.Status == "success" {
			ms.SuccessCount++
		} else {
			ms.ErrorCount++
		}
		ms.TotalTokens += l.TotalTokens
		ms.InputTokens += l.InputTokens
		ms.OutputTokens += l.OutputTokens
		if l.EstimatedCostKrw != nil {
			ms.TotalCost += *l.EstimatedCostKrw
		}
		if l.LatencyMs != nil {
			ms.AvgLatency += float64(*l.LatencyMs)
		}
		if l.CreatedAt.After(ms.LastUsed) {
			ms.LastUsed = l.CreatedAt
		}
	}

	// Calculate averages and add limit info
	items := make([]map[string]any, 0, len(modelMap))
	for _, ms := range modelMap {
		if ms.CallCount > 0 {
			ms.AvgLatency /= float64(ms.CallCount)
		}

		// Get rate limits for this model
		limits := ai.GetModelLimits(ms.Model)

		items = append(items, map[string]any{
			"model":          ms.Model,
			"provider":       ms.Provider,
			"call_count":     ms.CallCount,
			"success_count":  ms.SuccessCount,
			"error_count":    ms.ErrorCount,
			"error_rate":     float64(ms.ErrorCount) / float64(ms.CallCount) * 100,
			"total_tokens":   ms.TotalTokens,
			"input_tokens":   ms.InputTokens,
			"output_tokens":  ms.OutputTokens,
			"total_cost_krw": ms.TotalCost,
			"avg_cost_krw":   ms.TotalCost / float64(ms.CallCount),
			"avg_latency_ms": ms.AvgLatency,
			"last_used":      ms.LastUsed,
			// Rate limit info
			"rpm":         limits.RPM,
			"tpm":         limits.TPM,
			"rpd":         limits.RPD,
			"context_size": limits.Context,
			"note":        limits.Note,
		})
	}

	// Get configured models from prompt templates
	prompts, _ := ctrl.db.PromptTemplate.Query().
		Where(prompttemplate.IsActiveEQ(true)).
		All(ctx)

	featureModels := make(map[string]string)
	for _, pt := range prompts {
		featureModels[pt.Category+"/"+pt.SubCategory] = pt.Model
	}

	// Estimate quotas (today's usage) for all providers
	quotas := estimateModelQuotas(logs)

	c.JSON(http.StatusOK, gin.H{
		"data":           items,
		"days":           days,
		"feature_models": featureModels,
		"model_quotas":   quotas,
	})
}

// estimateModelQuotas estimates remaining RPD quota for all models with daily limits.
// Gemini quotas reset at midnight Pacific Time, Groq at midnight UTC.
func estimateModelQuotas(logs []*ent.UsageLog) []map[string]any {
	// Use centralized model limits
	selectableModels := ai.GetSelectableModels()

	// Get today boundaries
	loc, _ := time.LoadLocation("America/Los_Angeles")
	nowPT := time.Now().In(loc)
	geminiStartOfDay := time.Date(nowPT.Year(), nowPT.Month(), nowPT.Day(), 0, 0, 0, 0, loc)
	nowUTC := time.Now().UTC()
	groqStartOfDay := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)

	// Count usage per model (today only)
	geminiUsage := make(map[string]int)
	groqUsage := make(map[string]int)
	for _, l := range logs {
		if l.Model == nil {
			continue
		}
		if l.CreatedAt.After(geminiStartOfDay) {
			geminiUsage[*l.Model]++
		}
		if l.CreatedAt.After(groqStartOfDay) {
			groqUsage[*l.Model]++
		}
	}

	result := []map[string]any{}
	for _, m := range selectableModels {
		if m.RPD == 0 {
			continue // Skip models without daily limits
		}
		used := 0
		if m.Provider == "gemini" {
			used = geminiUsage[m.ID]
		} else if m.Provider == "groq" {
			used = groqUsage[m.ID]
		}
		remaining := m.RPD - used
		if remaining < 0 {
			remaining = 0
		}
		result = append(result, map[string]any{
			"model":      m.ID,
			"provider":   m.Provider,
			"limit":      m.RPD,
			"used":       used,
			"remaining":  remaining,
			"percentage": float64(used) / float64(m.RPD) * 100,
			"rpm":        m.RPM,
			"tpm":        m.TPM,
		})
	}

	return result
}

// ListAvailableModels returns all models that can be assigned to prompt templates.
// GET /v1/admin/models/available
func (ctrl *AdminController) ListAvailableModels(c *gin.Context) {
	models := ai.GetSelectableModels()
	c.JSON(http.StatusOK, gin.H{"data": models})
}

// GetRateLimitHits returns recent rate limit hit events for monitoring.
// GET /v1/admin/models/rate-limit-hits?hours=24
func (ctrl *AdminController) GetRateLimitHits(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours <= 0 {
		hours = 24
	}

	if ctrl.aiProvider == nil || ctrl.aiProvider.Throttler() == nil {
		c.JSON(http.StatusOK, gin.H{"data": []any{}, "hours": hours})
		return
	}

	summary := ctrl.aiProvider.Throttler().GetQuotaHitSummary(time.Duration(hours) * time.Hour)
	c.JSON(http.StatusOK, gin.H{"data": summary, "hours": hours})
}

// GetDashboard returns a single aggregated response for the admin dashboard.
// GET /v1/admin/dashboard
func (ctrl *AdminController) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	sevenDaysAgo := today.AddDate(0, 0, -7)

	// ── User stats ──
	totalUsers, _ := ctrl.db.UserProfile.Query().Count(ctx)
	newUsersToday, _ := ctrl.db.UserProfile.Query().
		Where(userprofile.CreatedAtGTE(today)).Count(ctx)
	activeToday, _ := ctrl.db.UserProfile.Query().
		Where(userprofile.LastLoginAtGTE(today)).Count(ctx)

	// ── AI metrics today & yesterday ──
	todayLogs, _ := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(today)).All(ctx)
	yesterdayLogs, _ := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(yesterday), usagelog.CreatedAtLT(today)).All(ctx)

	callsToday := len(todayLogs)
	callsYesterday := len(yesterdayLogs)
	var errorsToday, errorsYesterday int
	var costToday float64
	for _, l := range todayLogs {
		if l.Status == "error" {
			errorsToday++
		}
		if l.EstimatedCostKrw != nil {
			costToday += *l.EstimatedCostKrw
		}
	}
	for _, l := range yesterdayLogs {
		if l.Status == "error" {
			errorsYesterday++
		}
	}
	var callsDeltaPct, errorRateToday, errorRateDelta float64
	if callsYesterday > 0 {
		callsDeltaPct = float64(callsToday-callsYesterday) / float64(callsYesterday) * 100
	}
	if callsToday > 0 {
		errorRateToday = float64(errorsToday) / float64(callsToday) * 100
	}
	var errorRateYesterday float64
	if callsYesterday > 0 {
		errorRateYesterday = float64(errorsYesterday) / float64(callsYesterday) * 100
	}
	errorRateDelta = errorRateToday - errorRateYesterday

	// ── Daily metrics (last 7 days) ──
	allLogs7, _ := ctrl.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(sevenDaysAgo)).All(ctx)
	dailyMap := map[string]map[string]any{}
	for i := 0; i < 7; i++ {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		dailyMap[d] = map[string]any{"date": d, "calls": 0, "error_count": 0, "cost_krw": float64(0)}
	}
	for _, l := range allLogs7 {
		d := l.CreatedAt.Format("2006-01-02")
		if _, ok := dailyMap[d]; !ok {
			continue
		}
		dailyMap[d]["calls"] = dailyMap[d]["calls"].(int) + 1
		if l.Status == "error" {
			dailyMap[d]["error_count"] = dailyMap[d]["error_count"].(int) + 1
		}
		if l.EstimatedCostKrw != nil {
			dailyMap[d]["cost_krw"] = dailyMap[d]["cost_krw"].(float64) + *l.EstimatedCostKrw
		}
	}
	dailyMetrics := make([]map[string]any, 0, 7)
	for i := 6; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		dailyMetrics = append(dailyMetrics, dailyMap[d])
	}

	// ── Quota alerts (models > 80% RPD) ──
	allQuotas := estimateModelQuotas(allLogs7)
	quotaAlerts := make([]map[string]any, 0)
	for _, q := range allQuotas {
		pct, _ := q["percentage"].(float64)
		if pct >= 80 {
			quotaAlerts = append(quotaAlerts, map[string]any{
				"model":    q["model"],
				"provider": q["provider"],
				"used_pct": pct,
				"remaining": q["remaining"],
			})
		}
	}

	// ── Recent errors (last 5 from ai_call_errors) ──
	recentErrRecords, _ := ctrl.db.AICallError.Query().
		WithUsageLog().
		Order(ent.Desc("created_at")).
		Limit(5).
		All(ctx)
	recentErrors := make([]map[string]any, 0, len(recentErrRecords))
	for _, er := range recentErrRecords {
		item := map[string]any{
			"id":            er.ID,
			"error_type":    string(er.ErrorType),
			"error_message": er.ErrorMessage,
			"created_at":    er.CreatedAt,
		}
		if er.Edges.UsageLog != nil {
			ul := er.Edges.UsageLog
			item["feature"] = ul.Feature
			item["user_id"] = ul.UserID
			item["input_tokens"] = ul.InputTokens
			item["output_tokens"] = ul.OutputTokens
			if ul.Model != nil {
				item["model"] = *ul.Model
			}
			if ul.Provider != nil {
				item["provider"] = *ul.Provider
			}
		}
		recentErrors = append(recentErrors, item)
	}

	// ── Recent pending feedbacks (last 3) ──
	pendingFeedbackCount, _ := ctrl.db.Feedback.Query().
		Where(feedback.AdminStatusEQ(feedback.AdminStatusPending)).Count(ctx)
	recentFeedbackRecords, _ := ctrl.db.Feedback.Query().
		Where(feedback.AdminStatusEQ(feedback.AdminStatusPending)).
		Order(ent.Desc("created_at")).
		Limit(3).
		All(ctx)
	recentFeedbacks := make([]map[string]any, 0, len(recentFeedbackRecords))
	for _, f := range recentFeedbackRecords {
		preview := f.Content
		if len(preview) > 100 {
			preview = preview[:100]
		}
		recentFeedbacks = append(recentFeedbacks, map[string]any{
			"id":              f.ID,
			"category":        string(f.Category),
			"content_preview": preview,
			"created_at":      f.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]any{
		"user_stats": map[string]any{
			"total_users":        totalUsers,
			"new_users_today":    newUsersToday,
			"active_users_today": activeToday,
		},
		"ai_metrics_today": map[string]any{
			"calls":            callsToday,
			"calls_delta_pct":  callsDeltaPct,
			"error_rate":       errorRateToday,
			"error_rate_delta": errorRateDelta,
			"cost_krw":         costToday,
		},
		"daily_metrics":          dailyMetrics,
		"quota_alerts":           quotaAlerts,
		"recent_errors":          recentErrors,
		"recent_feedbacks":       recentFeedbacks,
		"pending_feedback_count": pendingFeedbackCount,
	}})
}

// ListErrors returns paginated AI call errors from ai_call_errors table.
// GET /v1/admin/usage/errors?error_type=&provider=&feature=&days=7&limit=20&offset=0
func (ctrl *AdminController) ListErrors(c *gin.Context) {
	ctx := c.Request.Context()

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 {
		days = 7
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	since := time.Now().AddDate(0, 0, -days)

	q := ctrl.db.AICallError.Query().
		WithUsageLog().
		Where(aicallerror.CreatedAtGTE(since)).
		Order(ent.Desc("created_at"))

	if et := c.Query("error_type"); et != "" {
		q = q.Where(aicallerror.ErrorTypeEQ(aicallerror.ErrorType(et)))
	}

	// Filter by provider/feature via usage_log edge (DB-level, not post-query)
	var usageLogPreds []predicate.UsageLog
	if prov := c.Query("provider"); prov != "" {
		usageLogPreds = append(usageLogPreds, usagelog.ProviderEQ(prov))
	}
	if feat := c.Query("feature"); feat != "" {
		usageLogPreds = append(usageLogPreds, usagelog.FeatureEQ(feat))
	}
	if len(usageLogPreds) > 0 {
		q = q.Where(aicallerror.HasUsageLogWith(usageLogPreds...))
	}

	total, _ := q.Count(ctx)

	records, err := q.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		slog.Error("ListErrors failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "에러 목록 조회에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	items := make([]map[string]any, 0, len(records))
	for _, er := range records {
		item := map[string]any{
			"id":            er.ID,
			"error_type":    string(er.ErrorType),
			"error_message": er.ErrorMessage,
			"created_at":    er.CreatedAt,
		}
		if er.Edges.UsageLog != nil {
			ul := er.Edges.UsageLog
			item["feature"] = ul.Feature
			item["user_id"] = ul.UserID
			item["input_tokens"] = ul.InputTokens
			item["output_tokens"] = ul.OutputTokens
			if ul.Model != nil {
				item["model"] = *ul.Model
			}
			if ul.Provider != nil {
				item["provider"] = *ul.Provider
			}
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "count": total})
}
