package controller

import (
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/service"
)

func toUserInfo(u *ent.UserProfile) map[string]any {
	info := map[string]any{
		"id":                   u.ID.String(),
		"auth_provider":        string(u.AuthProvider),
		"role":                 string(u.Role),
		"onboarding_completed": u.OnboardingCompleted,
		"suspended":            u.Suspended,
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

func toExperienceResponse(exp *ent.Experience) map[string]any {
	resp := map[string]any{
		"id":             exp.ID.String(),
		"user_id":        exp.UserID.String(),
		"title":          exp.Title,
		"category":       exp.Category,
		"role":           exp.Role,
		"content":        exp.Content,
		"result":         exp.Result,
		"star_situation":  exp.StarSituation,
		"star_task":       exp.StarTask,
		"star_action":     exp.StarAction,
		"star_result":     exp.StarResult,
		"keywords":        exp.Keywords,
		"source":          exp.Source,
		"is_archived":     exp.IsArchived,
		"created_at":      exp.CreatedAt,
		"updated_at":      exp.UpdatedAt,
	}
	if exp.PeriodStart != nil {
		resp["period_start"] = exp.PeriodStart.Format("2006-01-02")
	}
	if exp.PeriodEnd != nil {
		resp["period_end"] = exp.PeriodEnd.Format("2006-01-02")
	}

	// Include weapons if loaded
	if weapons, err := exp.Edges.WeaponsOrErr(); err == nil {
		weaponList := make([]map[string]any, len(weapons))
		for i, w := range weapons {
			weaponList[i] = map[string]any{
				"id":             w.ID.String(),
				"weapon_code":    w.WeaponCode,
				"confidence":     w.Confidence,
				"is_primary":     w.IsPrimary,
				"reasoning":      w.Reasoning,
				"user_confirmed": w.UserConfirmed,
				"user_modified":  w.UserModified,
			}
		}
		resp["weapons"] = weaponList
	}

	return resp
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
