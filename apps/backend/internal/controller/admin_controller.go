package controller

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"context"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/adminauditlog"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/systemconfig"
	"github.com/coby/colight/apps/backend/ent/usagelog"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crypto"
	"github.com/coby/colight/apps/backend/internal/infrastructure/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminController struct {
	db *ent.Client
}

func NewAdminController(db *ent.Client) *AdminController {
	return &AdminController{db: db}
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
			"model_name":          p.Model,
			"temperature":          p.Temperature,
			"max_tokens":           p.MaxTokens,
			"version":              p.Version,
			"is_active":            p.IsActive,
			"usage_count":          p.UsageCount,
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
		"model_name":          prompt.Model,
		"temperature":          prompt.Temperature,
		"max_tokens":           prompt.MaxTokens,
		"version":              prompt.Version,
		"is_active":            prompt.IsActive,
		"usage_count":          prompt.UsageCount,
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
