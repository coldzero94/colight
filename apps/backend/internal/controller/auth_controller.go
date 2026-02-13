package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authService *service.AuthService
	config      *config.Config
}

func NewAuthController(authService *service.AuthService, cfg *config.Config) *AuthController {
	return &AuthController{
		authService: authService,
		config:      cfg,
	}
}

// NaverLogin redirects to Naver OAuth authorization URL.
// GET /v1/auth/naver/login
func (ctrl *AuthController) NaverLogin(c *gin.Context) {
	authURL, state := ctrl.authService.GetNaverAuthURL()

	// Store state in cookie for CSRF verification
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	c.Redirect(http.StatusFound, authURL)
}

// NaverCallback handles the Naver OAuth callback.
// GET /v1/auth/naver/callback?code=xxx&state=yyy
func (ctrl *AuthController) NaverCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.Redirect(http.StatusFound, ctrl.config.FrontendURL+"/auth/callback?error=missing_params")
		return
	}

	// Verify CSRF state
	savedState, err := c.Cookie("oauth_state")
	if err != nil || savedState != state {
		c.Redirect(http.StatusFound, ctrl.config.FrontendURL+"/auth/callback?error=invalid_state")
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	result, err := ctrl.authService.HandleNaverCallback(c.Request.Context(), code, state)
	if err != nil {
		slog.Error("naver callback failed", "error", err)
		c.Redirect(http.StatusFound, ctrl.config.FrontendURL+"/auth/callback?error=auth_failed")
		return
	}

	// Redirect to frontend with tokens
	redirectURL := ctrl.config.FrontendURL + "/auth/callback" +
		"?access_token=" + result.Tokens.AccessToken +
		"&refresh_token=" + result.Tokens.RefreshToken

	c.Redirect(http.StatusFound, redirectURL)
}

// Signup handles email/password registration.
// POST /v1/auth/signup
func (ctrl *AuthController) Signup(c *gin.Context) {
	var req struct {
		Email    string  `json:"email" binding:"required,email"`
		Password string  `json:"password" binding:"required,min=8"`
		Nickname *string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "입력값이 올바르지 않습니다.", "code": "VALID_001"},
		})
		return
	}

	nickname := ""
	if req.Nickname != nil {
		nickname = *req.Nickname
	}

	result, err := ctrl.authService.Signup(c.Request.Context(), req.Email, req.Password, nickname)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{"message": "이미 사용 중인 이메일입니다.", "code": "AUTH_004"},
			})
			return
		}
		slog.Error("signup failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "회원가입에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":   toUserInfo(result.User),
		"tokens": toAuthTokens(result.Tokens),
	})
}

// Login handles email/password authentication.
// POST /v1/auth/login
func (ctrl *AuthController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "입력값이 올바르지 않습니다.", "code": "VALID_001"},
		})
		return
	}

	result, err := ctrl.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": "이메일 또는 비밀번호가 올바르지 않습니다.", "code": "AUTH_005"},
			})
			return
		}
		slog.Error("login failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "로그인에 실패했습니다.", "code": "SYS_001"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   toUserInfo(result.User),
		"tokens": toAuthTokens(result.Tokens),
	})
}

// Refresh issues new tokens using a valid refresh token.
// POST /v1/auth/refresh
func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "refresh_token이 필요합니다.", "code": "VALID_001"},
		})
		return
	}

	tokens, err := ctrl.authService.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "유효하지 않은 토큰입니다.", "code": "AUTH_002"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": toAuthTokens(tokens),
	})
}

// Me returns the current authenticated user's info.
// GET /v1/auth/me (requires auth middleware)
func (ctrl *AuthController) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
		})
		return
	}

	user, err := ctrl.authService.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"message": "사용자를 찾을 수 없습니다.", "code": "AUTH_006"},
		})
		return
	}

	c.JSON(http.StatusOK, toUserInfo(user))
}

// Logout is a placeholder — client-side token disposal.
// POST /v1/auth/logout (requires auth middleware)
func (ctrl *AuthController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "로그아웃 되었습니다."})
}
