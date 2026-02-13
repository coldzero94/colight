package controller

import (
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/service"
)

func toUserInfo(u *ent.UserProfile) map[string]any {
	info := map[string]any{
		"id":                   u.ID.String(),
		"auth_provider":        string(u.AuthProvider),
		"role":                 string(u.Role),
		"onboarding_completed": u.OnboardingCompleted,
		"created_at":           u.CreatedAt,
	}
	if u.Email != nil {
		info["email"] = *u.Email
	}
	if u.Nickname != "" {
		info["nickname"] = u.Nickname
	}
	return info
}

func toAuthTokens(t *service.TokenPair) map[string]any {
	return map[string]any{
		"access_token":  t.AccessToken,
		"refresh_token": t.RefreshToken,
		"expires_in":    t.ExpiresIn,
	}
}
