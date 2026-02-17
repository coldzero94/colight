package controller

import (
	"log/slog"
	"net/http"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ProfileController struct {
	db *ent.Client
}

func NewProfileController(db *ent.Client) *ProfileController {
	return &ProfileController{db: db}
}

// UpdateProfile updates the current user's profile (nickname).
// PUT /v1/profile
func (ctrl *ProfileController) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(uuid.UUID)

	var req struct {
		Nickname string `json:"nickname" binding:"required,min=1,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "닉네임을 입력해주세요 (최대 50자).", "code": "VALID_001"},
		})
		return
	}

	user, err := ctrl.db.UserProfile.UpdateOneID(uid).
		SetNickname(req.Nickname).
		Save(c.Request.Context())
	if err != nil {
		slog.Error("update profile failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "프로필 수정에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, toUserInfo(user))
}

// ChangePassword changes the current user's password.
// POST /v1/profile/password
func (ctrl *ProfileController) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(uuid.UUID)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "비밀번호는 최소 8자 이상이어야 합니다.", "code": "VALID_001"},
		})
		return
	}

	ctx := c.Request.Context()
	user, err := ctrl.db.UserProfile.Get(ctx, uid)
	if err != nil {
		slog.Error("get user for password change failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "비밀번호 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	// Check current password
	if user.PasswordHash == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "이메일 계정만 비밀번호를 변경할 수 있습니다.", "code": "AUTH_007"},
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "현재 비밀번호가 올바르지 않습니다.", "code": "AUTH_004"},
		})
		return
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		slog.Error("hash password failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "비밀번호 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	_, err = ctrl.db.UserProfile.UpdateOneID(uid).
		SetPasswordHash(string(hash)).
		Save(ctx)
	if err != nil {
		slog.Error("update password failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "비밀번호 변경에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "비밀번호가 변경되었습니다."})
}
