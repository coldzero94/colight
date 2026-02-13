package controller

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/userprofile"
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

// UpdateUserRole changes a user's role.
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
		Role string `json:"role" binding:"required,oneof=user admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "올바른 역할을 지정해주세요. (user, admin)", "code": "VALID_001"},
		})
		return
	}

	user, err := ctrl.db.UserProfile.UpdateOneID(id).
		SetRole(userprofile.Role(req.Role)).
		Save(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
			})
			return
		}
		slog.Error("update user role failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "역할 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
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
