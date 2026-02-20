package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crypto"
	"github.com/coby/colight/apps/backend/internal/infrastructure/middleware"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminController struct {
	db           *ent.Client // retained for HealthCheck DB ping
	aiProvider   *ai.AIProvider
	userSvc      *service.AdminUserService
	analyticsSvc *service.AdminAnalyticsService
	configSvc    *service.AdminConfigService
	opsSvc       *service.AdminOperationsService
}

func NewAdminController(db *ent.Client, aiProvider *ai.AIProvider) *AdminController {
	return &AdminController{
		db:           db,
		aiProvider:   aiProvider,
		userSvc:      service.NewAdminUserService(db),
		analyticsSvc: service.NewAdminAnalyticsService(db),
		configSvc:    service.NewAdminConfigService(db),
		opsSvc:       service.NewAdminOperationsService(db),
	}
}

// adminErr writes a JSON error response.
func adminErr(c *gin.Context, status int, msg, code string) {
	c.JSON(status, gin.H{"error": gin.H{"message": msg, "code": code}})
}

// ── User Management ──────────────────────────────────────────────────────────

// ListUsers returns paginated user list with optional role/search filter.
// GET /v1/admin/users
func (ctrl *AdminController) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := ctrl.userSvc.ListUsers(c.Request.Context(), service.AdminListUsersParams{
		Limit:      limit,
		Offset:     offset,
		RoleFilter: c.Query("role"),
		Search:     c.Query("search"),
	})
	if err != nil {
		slog.Error("list users failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "사용자 목록 조회에 실패했습니다.", "SYS_001")
		return
	}

	items := make([]map[string]any, 0, len(result.Users))
	for _, u := range result.Users {
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

	c.JSON(http.StatusOK, gin.H{"data": items, "total": result.Total})
}

// GetUser returns a single user by ID.
// GET /v1/admin/users/:id
func (ctrl *AdminController) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	user, err := ctrl.userSvc.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("get user failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "사용자 조회에 실패했습니다.", "SYS_001")
		return
	}

	c.JSON(http.StatusOK, toUserInfo(user))
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
		adminErr(c, http.StatusBadRequest, "입력값이 올바르지 않습니다.", "VALID_001")
		return
	}

	user, err := ctrl.userSvc.CreateUser(c.Request.Context(), service.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
		Role:     req.Role,
		Plan:     req.Plan,
	})
	if err != nil {
		if errors.Is(err, service.ErrAdminEmailExists) {
			adminErr(c, http.StatusConflict, "이미 존재하는 이메일입니다.", "AUTH_002")
			return
		}
		slog.Error("create user failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "사용자 생성에 실패했습니다.", "SYS_001")
		return
	}

	c.JSON(http.StatusCreated, toUserInfo(user))
}

// UpdateUserRole changes a user's role with hierarchy validation.
// PUT /v1/admin/users/:id/role
func (ctrl *AdminController) UpdateUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=user manager admin super_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminErr(c, http.StatusBadRequest, "올바른 역할을 지정해주세요. (user, manager, admin, super_admin)", "VALID_001")
		return
	}

	callerID, _ := c.Get("user_id")
	uid, _ := callerID.(uuid.UUID)
	callerRole, _ := c.Get("role")
	callerRoleStr, _ := callerRole.(string)

	user, oldRole, err := ctrl.userSvc.UpdateUserRole(c.Request.Context(), service.UpdateRoleInput{
		TargetID:        id,
		CallerID:        uid,
		CallerRoleLevel: middleware.RoleLevel(callerRoleStr),
		CallerRoleStr:   callerRoleStr,
		NewRole:         req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminSelfRoleChange):
			adminErr(c, http.StatusForbidden, "자기 자신의 역할은 변경할 수 없습니다.", "AUTH_003")
		case errors.Is(err, service.ErrAdminInsufficientLevel):
			adminErr(c, http.StatusForbidden, "자신의 권한 이상으로 역할을 변경할 수 없습니다.", "AUTH_003")
		case errors.Is(err, service.ErrAdminUserNotFound):
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
		case errors.Is(err, service.ErrAdminTargetTooHigh):
			adminErr(c, http.StatusForbidden, "자신의 권한 이상의 사용자를 변경할 수 없습니다.", "AUTH_003")
		case errors.Is(err, service.ErrAdminLastSuperAdmin):
			adminErr(c, http.StatusForbidden, "마지막 최고 관리자는 강등할 수 없습니다.", "AUTH_003")
		default:
			slog.Error("update user role failed", "error", err)
			adminErr(c, http.StatusInternalServerError, "역할 변경에 실패했습니다.", "SYS_001")
		}
		return
	}

	// Audit log
	ctrl.opsSvc.RecordAuditLog(c.Request.Context(), uid, "role_change", "user", id.String(), &oldRole, &req.Role)

	c.JSON(http.StatusOK, toUserInfo(user))
}

// SuspendUser suspends a user account.
// POST /v1/admin/users/:id/suspend
func (ctrl *AdminController) SuspendUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	user, err := ctrl.userSvc.SuspendUser(c.Request.Context(), id, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminUserNotFound):
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
		case errors.Is(err, service.ErrAdminUserAlreadySuspended):
			adminErr(c, http.StatusConflict, "이미 정지된 계정입니다.", "VALID_001")
		default:
			slog.Error("suspend user failed", "error", err)
			adminErr(c, http.StatusInternalServerError, "계정 정지에 실패했습니다.", "SYS_001")
		}
		return
	}

	c.JSON(http.StatusOK, toUserInfo(user))
}

// UnsuspendUser lifts suspension on a user account.
// DELETE /v1/admin/users/:id/suspend
func (ctrl *AdminController) UnsuspendUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	user, err := ctrl.userSvc.UnsuspendUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("unsuspend user failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "정지 해제에 실패했습니다.", "SYS_001")
		return
	}

	c.JSON(http.StatusOK, toUserInfo(user))
}

// GetUserDetail returns detailed information about a user including counts.
// GET /v1/admin/users/:id/detail
func (ctrl *AdminController) GetUserDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	detail, err := ctrl.userSvc.GetUserDetail(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("get user detail failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "사용자 조회에 실패했습니다.", "SYS_001")
		return
	}

	info := toUserInfo(detail.User)
	info["experience_count"] = detail.ExperienceCount
	info["coaching_count"] = detail.CoachingCount
	info["usage_count"] = detail.UsageCount
	info["suspended"] = detail.User.Suspended
	if detail.User.SuspendedAt != nil {
		info["suspended_at"] = *detail.User.SuspendedAt
	}
	if detail.User.SuspendedReason != nil {
		info["suspended_reason"] = *detail.User.SuspendedReason
	}

	c.JSON(http.StatusOK, info)
}

// ForceLogout invalidates all tokens for a user by setting force_logout_at.
// POST /v1/admin/users/:id/force-logout
func (ctrl *AdminController) ForceLogout(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	user, err := ctrl.userSvc.ForceLogout(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("force logout failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "강제 로그아웃에 실패했습니다.", "SYS_001")
		return
	}

	info := toUserInfo(user)
	if user.ForceLogoutAt != nil {
		info["force_logout_at"] = *user.ForceLogoutAt
	}

	c.JSON(http.StatusOK, info)
}

// UpdateUserPlan changes a user's subscription plan.
// PUT /v1/admin/users/:id/plan
func (ctrl *AdminController) UpdateUserPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	var req struct {
		Plan string `json:"plan" binding:"required,oneof=free starter pro season"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminErr(c, http.StatusBadRequest, "올바른 플랜을 지정해주세요. (free, starter, pro, season)", "VALID_001")
		return
	}

	user, err := ctrl.userSvc.UpdateUserPlan(c.Request.Context(), id, req.Plan)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("update user plan failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "플랜 변경에 실패했습니다.", "SYS_001")
		return
	}

	info := toUserInfo(user)
	info["plan"] = string(user.Plan)

	c.JSON(http.StatusOK, info)
}

// ExportUserData exports all user data for PIPA compliance.
// POST /v1/admin/users/:id/export
func (ctrl *AdminController) ExportUserData(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	export, err := ctrl.userSvc.ExportUserData(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("export user data failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "데이터 내보내기에 실패했습니다.", "SYS_001")
		return
	}

	expList := make([]map[string]any, 0, len(export.Experiences))
	for _, exp := range export.Experiences {
		expList = append(expList, map[string]any{
			"id":      exp.ID.String(),
			"title":   exp.Title,
			"content": exp.Content,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"profile":     toUserInfo(export.User),
		"experiences": expList,
	})
}

// CreateDeletionRequest creates a data deletion request for a user.
// POST /v1/admin/users/:id/delete-request
func (ctrl *AdminController) CreateDeletionRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	callerID, _ := c.Get("user_id")
	adminID, _ := callerID.(uuid.UUID)

	dr, err := ctrl.userSvc.CreateDeletionRequest(c.Request.Context(), id, adminID, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			adminErr(c, http.StatusNotFound, "사용자를 찾을 수 없습니다.", "AUTH_006")
			return
		}
		slog.Error("create deletion request failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "삭제 요청 생성에 실패했습니다.", "SYS_001")
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
		adminErr(c, http.StatusBadRequest, "잘못된 사용자 ID입니다.", "VALID_001")
		return
	}

	dr, err := ctrl.userSvc.CancelDeletionRequest(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminDeletionNotFound) {
			adminErr(c, http.StatusNotFound, "대기 중인 삭제 요청이 없습니다.", "SYS_002")
			return
		}
		slog.Error("cancel deletion request failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "삭제 요청 취소에 실패했습니다.", "SYS_001")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           dr.ID.String(),
		"user_id":      dr.UserID.String(),
		"status":       string(dr.Status),
		"cancelled_at": dr.CancelledAt,
	})
}

// ListDeletionQueue returns pending deletion requests.
// GET /v1/admin/deletion-queue
func (ctrl *AdminController) ListDeletionQueue(c *gin.Context) {
	requests, err := ctrl.userSvc.ListDeletionQueue(c.Request.Context())
	if err != nil {
		slog.Error("list deletion queue failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "삭제 대기열 조회에 실패했습니다.", "SYS_001")
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

// ── Analytics ────────────────────────────────────────────────────────────────

// GetStats returns system-level statistics.
// GET /v1/admin/stats
func (ctrl *AdminController) GetStats(c *gin.Context) {
	stats, err := ctrl.analyticsSvc.GetStats(c.Request.Context())
	if err != nil {
		slog.Error("get stats failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "통계 조회에 실패했습니다.", "SYS_001")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users":        stats.TotalUsers,
		"email_auth_users":   stats.EmailAuthUsers,
		"naver_auth_users":   stats.NaverAuthUsers,
		"active_users_today": stats.ActiveUsersToday,
		"total_experiences":  stats.TotalExperiences,
	})
}

// GetUsageSummary returns aggregated usage statistics for a given period.
// GET /v1/admin/usage/summary
func (ctrl *AdminController) GetUsageSummary(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	summary, err := ctrl.analyticsSvc.GetUsageSummary(c.Request.Context(), days)
	if err != nil {
		slog.Error("get usage summary failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "사용량 요약 조회에 실패했습니다.", "SYS_001")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_calls":    summary.TotalCalls,
		"total_tokens":   summary.TotalTokens,
		"total_cost_krw": summary.TotalCostKRW,
		"error_rate":     summary.ErrorRate,
		"error_count":    summary.ErrorCount,
		"days":           summary.Days,
	})
}

// GetUsageDaily returns daily usage breakdown.
// GET /v1/admin/usage/daily
func (ctrl *AdminController) GetUsageDaily(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 {
		days = 7
	}

	daily, err := ctrl.analyticsSvc.GetUsageDaily(c.Request.Context(), days)
	if err != nil {
		slog.Error("get usage daily failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "일별 사용량 조회에 실패했습니다.", "SYS_001")
		return
	}

	items := make([]map[string]any, 0, len(daily))
	for _, d := range daily {
		items = append(items, map[string]any{
			"date":         d.Date,
			"total_calls":  d.TotalCalls,
			"total_tokens": d.TotalTokens,
			"error_count":  d.ErrorCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// GetUsageCosts returns provider-level cost breakdown.
// GET /v1/admin/usage/costs
func (ctrl *AdminController) GetUsageCosts(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	costs, err := ctrl.analyticsSvc.GetUsageCosts(c.Request.Context(), days)
	if err != nil {
		slog.Error("get usage costs failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "비용 조회에 실패했습니다.", "SYS_001")
		return
	}

	items := make([]map[string]any, 0, len(costs))
	for _, pc := range costs {
		items = append(items, map[string]any{
			"provider":      pc.Provider,
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, err := ctrl.analyticsSvc.GetUsageTopUsers(c.Request.Context(), days, limit)
	if err != nil {
		slog.Error("get usage top users failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "상위 사용자 조회에 실패했습니다.", "SYS_001")
		return
	}

	items := make([]map[string]any, 0, len(users))
	for _, uu := range users {
		items = append(items, map[string]any{
			"user_id":       uu.UserID.String(),
			"total_tokens":  uu.TotalTokens,
			"total_cost_krw": uu.TotalCost,
			"call_count":    uu.CallCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "days": days})
}

// GetModelStats returns model-level statistics with quota estimates.
// GET /v1/admin/models/stats?days=30
func (ctrl *AdminController) GetModelStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	result, err := ctrl.analyticsSvc.GetModelStats(c.Request.Context(), days)
	if err != nil {
		slog.Error("get model stats failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "모델 통계 조회에 실패했습니다.", "SYS_001")
		return
	}

	items := make([]map[string]any, 0, len(result.Models))
	for _, ms := range result.Models {
		items = append(items, map[string]any{
			"model":          ms.Model,
			"provider":       ms.Provider,
			"call_count":     ms.CallCount,
			"success_count":  ms.SuccessCount,
			"error_count":    ms.ErrorCount,
			"error_rate":     ms.ErrorRate,
			"total_tokens":   ms.TotalTokens,
			"input_tokens":   ms.InputTokens,
			"output_tokens":  ms.OutputTokens,
			"total_cost_krw": ms.TotalCostKRW,
			"avg_cost_krw":   ms.AvgCostKRW,
			"avg_latency_ms": ms.AvgLatencyMs,
			"last_used":      ms.LastUsed,
			"rpm":            ms.RPM,
			"tpm":            ms.TPM,
			"rpd":            ms.RPD,
			"context_size":   ms.ContextSize,
			"note":           ms.Note,
		})
	}

	quotas := make([]map[string]any, 0, len(result.ModelQuotas))
	for _, q := range result.ModelQuotas {
		quotas = append(quotas, map[string]any{
			"model":      q.Model,
			"provider":   q.Provider,
			"limit":      q.Limit,
			"used":       q.Used,
			"remaining":  q.Remaining,
			"percentage": q.Percentage,
			"rpm":        q.RPM,
			"tpm":        q.TPM,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":           items,
		"days":           days,
		"feature_models": result.FeatureModels,
		"model_quotas":   quotas,
	})
}

// GetDashboard returns a single aggregated response for the admin dashboard.
// GET /v1/admin/dashboard
func (ctrl *AdminController) GetDashboard(c *gin.Context) {
	data, err := ctrl.analyticsSvc.GetDashboard(c.Request.Context())
	if err != nil {
		slog.Error("get dashboard failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "대시보드 조회에 실패했습니다.", "SYS_001")
		return
	}

	dailyMetrics := make([]map[string]any, 0, len(data.DailyMetrics))
	for _, d := range data.DailyMetrics {
		dailyMetrics = append(dailyMetrics, map[string]any{
			"date":        d.Date,
			"calls":       d.Calls,
			"error_count": d.ErrorCount,
			"cost_krw":    d.CostKRW,
		})
	}

	quotaAlerts := make([]map[string]any, 0, len(data.QuotaAlerts))
	for _, q := range data.QuotaAlerts {
		quotaAlerts = append(quotaAlerts, map[string]any{
			"model":     q.Model,
			"provider":  q.Provider,
			"used_pct":  q.Percentage,
			"remaining": q.Remaining,
		})
	}

	recentErrors := make([]map[string]any, 0, len(data.RecentErrors))
	for _, er := range data.RecentErrors {
		item := map[string]any{
			"id":            er.ID,
			"error_type":    er.ErrorType,
			"error_message": er.ErrorMessage,
			"created_at":    er.CreatedAt,
		}
		// Usage log fields (AICallErrors always have an associated usage log)
		item["feature"] = er.Feature
		item["user_id"] = er.UserID
		item["input_tokens"] = er.InputTokens
		item["output_tokens"] = er.OutputTokens
		if er.Model != "" {
			item["model"] = er.Model
		}
		if er.Provider != "" {
			item["provider"] = er.Provider
		}
		recentErrors = append(recentErrors, item)
	}

	recentFeedbacks := make([]map[string]any, 0, len(data.RecentFeedbacks))
	for _, f := range data.RecentFeedbacks {
		recentFeedbacks = append(recentFeedbacks, map[string]any{
			"id":              f.ID,
			"category":        f.Category,
			"content_preview": f.ContentPreview,
			"created_at":      f.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]any{
		"user_stats": map[string]any{
			"total_users":        data.UserStats.TotalUsers,
			"new_users_today":    data.UserStats.NewUsersToday,
			"active_users_today": data.UserStats.ActiveUsersToday,
		},
		"ai_metrics_today": map[string]any{
			"calls":            data.AIMetrics.Calls,
			"calls_delta_pct":  data.AIMetrics.CallsDeltaPct,
			"error_rate":       data.AIMetrics.ErrorRate,
			"error_rate_delta": data.AIMetrics.ErrorRateDelta,
			"cost_krw":         data.AIMetrics.CostKRW,
		},
		"daily_metrics":          dailyMetrics,
		"quota_alerts":           quotaAlerts,
		"recent_errors":          recentErrors,
		"recent_feedbacks":       recentFeedbacks,
		"pending_feedback_count": data.PendingFeedbackCount,
	}})
}

// ListErrors returns paginated AI call errors from ai_call_errors table.
// GET /v1/admin/usage/errors?error_type=&provider=&feature=&days=7&limit=20&offset=0
func (ctrl *AdminController) ListErrors(c *gin.Context) {
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

	result, err := ctrl.analyticsSvc.ListErrors(c.Request.Context(), service.ErrorListParams{
		Days:      days,
		Limit:     limit,
		Offset:    offset,
		ErrorType: c.Query("error_type"),
		Provider:  c.Query("provider"),
		Feature:   c.Query("feature"),
	})
	if err != nil {
		slog.Error("ListErrors failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "에러 목록 조회에 실패했습니다.", "SYS_001")
		return
	}

	items := make([]map[string]any, 0, len(result.Errors))
	for _, er := range result.Errors {
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

	c.JSON(http.StatusOK, gin.H{"data": items, "count": result.Total})
}

// ── Configuration ────────────────────────────────────────────────────────────

// ListConfigs returns system configuration entries with optional category filter.
// Secret values are masked. GET /v1/admin/configs
func (ctrl *AdminController) ListConfigs(c *gin.Context) {
	configs, err := ctrl.configSvc.ListConfigs(c.Request.Context(), c.Query("category"))
	if err != nil {
		slog.Error("list configs failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "설정 목록 조회에 실패했습니다.", "SYS_001")
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
		adminErr(c, http.StatusBadRequest, "값을 입력해주세요.", "VALID_001")
		return
	}

	callerID, _ := c.Get("user_id")
	uid, _ := callerID.(uuid.UUID)

	updated, oldValue, err := ctrl.configSvc.UpdateConfig(c.Request.Context(), service.UpdateConfigInput{
		Key:      key,
		Value:    req.Value,
		CallerID: uid,
	})
	if err != nil {
		if errors.Is(err, service.ErrAdminConfigNotFound) {
			adminErr(c, http.StatusNotFound, "설정을 찾을 수 없습니다.", "SYS_002")
			return
		}
		slog.Error("update config failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "설정 수정에 실패했습니다.", "SYS_001")
		return
	}

	// Audit log
	ctrl.opsSvc.RecordAuditLog(c.Request.Context(), uid, "config_update", "config", key, &oldValue, &req.Value)

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

// ListPrompts returns prompt templates with optional category filter.
// GET /v1/admin/prompts
func (ctrl *AdminController) ListPrompts(c *gin.Context) {
	prompts, err := ctrl.configSvc.ListPrompts(c.Request.Context(), c.Query("category"))
	if err != nil {
		slog.Error("list prompts failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "프롬프트 목록 조회에 실패했습니다.", "SYS_001")
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

// UpdatePrompt updates a prompt template's fields.
// PUT /v1/admin/prompts/:id
func (ctrl *AdminController) UpdatePrompt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 프롬프트 ID입니다.", "VALID_001")
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
		adminErr(c, http.StatusBadRequest, "입력값이 올바르지 않습니다.", "VALID_001")
		return
	}

	prompt, err := ctrl.configSvc.UpdatePrompt(c.Request.Context(), service.UpdatePromptInput{
		ID:                 id,
		SystemPrompt:       req.SystemPrompt,
		UserPromptTemplate: req.UserPromptTemplate,
		ModelName:          req.ModelName,
		Temperature:        req.Temperature,
		MaxTokens:          req.MaxTokens,
		IsActive:           req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminPromptNotFound):
			adminErr(c, http.StatusNotFound, "프롬프트를 찾을 수 없습니다.", "SYS_002")
		case errors.Is(err, service.ErrAdminUnsupportedModel):
			adminErr(c, http.StatusBadRequest, "지원하지 않는 모델입니다. GET /v1/admin/models/available에서 사용 가능한 모델을 확인해주세요.", "VALID_001")
		default:
			slog.Error("update prompt failed", "error", err)
			adminErr(c, http.StatusInternalServerError, "프롬프트 수정에 실패했습니다.", "SYS_001")
		}
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

// ── Operations (Audit, Feedback) ─────────────────────────────────────────────

// ListAuditLogs returns paginated audit log entries.
// GET /v1/admin/audit-logs
func (ctrl *AdminController) ListAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, total, err := ctrl.opsSvc.ListAuditLogs(c.Request.Context(), limit, offset, c.Query("action"))
	if err != nil {
		slog.Error("list audit logs failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "감사 로그 조회에 실패했습니다.", "SYS_001")
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

	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

// ListFeedbacks returns paginated feedback entries with optional category filter.
// GET /v1/admin/feedbacks
func (ctrl *AdminController) ListFeedbacks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	feedbacks, total, err := ctrl.opsSvc.ListFeedbacks(c.Request.Context(), limit, offset, c.Query("category"))
	if err != nil {
		slog.Error("list feedbacks failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "피드백 목록 조회에 실패했습니다.", "SYS_001")
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

	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

// UpdateFeedbackStatus updates the admin review status of a feedback entry.
// PUT /v1/admin/feedbacks/:id
func (ctrl *AdminController) UpdateFeedbackStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminErr(c, http.StatusBadRequest, "잘못된 피드백 ID입니다.", "VALID_001")
		return
	}

	var req struct {
		AdminStatus string `json:"admin_status" binding:"required,oneof=pending reviewed resolved dismissed"`
		AdminNote   string `json:"admin_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminErr(c, http.StatusBadRequest, "올바른 상태를 지정해주세요.", "VALID_001")
		return
	}

	callerID, _ := c.Get("user_id")
	uid, _ := callerID.(uuid.UUID)

	fb, err := ctrl.opsSvc.UpdateFeedbackStatus(c.Request.Context(), id, uid, req.AdminStatus, req.AdminNote)
	if err != nil {
		if errors.Is(err, service.ErrAdminFeedbackNotFound) {
			adminErr(c, http.StatusNotFound, "피드백을 찾을 수 없습니다.", "SYS_002")
			return
		}
		slog.Error("update feedback status failed", "error", err)
		adminErr(c, http.StatusInternalServerError, "피드백 상태 변경에 실패했습니다.", "SYS_001")
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

// ── Infrastructure Pass-throughs ─────────────────────────────────────────────

// HealthCheck returns system health status including database connectivity.
// GET /v1/admin/health
func (ctrl *AdminController) HealthCheck(c *gin.Context) {
	dbStatus := "healthy"
	if _, err := ctrl.db.UserProfile.Query().Count(c.Request.Context()); err != nil {
		dbStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{"database": dbStatus})
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
