package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/experienceweapon"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// WeaponTaggingService handles AI-powered weapon classification for experiences
type WeaponTaggingService struct {
	entClient      *ent.Client
	aiProvider     *ai.AIProvider
	promptCache    *ent.PromptTemplate
	promptCachedAt time.Time
}

// NewWeaponTaggingService creates a new WeaponTaggingService
func NewWeaponTaggingService(entClient *ent.Client, aiProvider *ai.AIProvider) *WeaponTaggingService {
	return &WeaponTaggingService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// WeaponTagResult represents the result of weapon tagging
type WeaponTagResult struct {
	PrimaryWeapon struct {
		Code       string
		Confidence float64
		Reasoning  string
	}
	SecondaryWeapons []struct {
		Code       string
		Confidence float64
		Reasoning  string
	}
}

// aiWeaponResponse matches the expected AI response format
type aiWeaponResponse struct {
	PrimaryWeapon struct {
		Code       string  `json:"code"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	} `json:"primary_weapon"`
	SecondaryWeapons []struct {
		Code       string  `json:"code"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	} `json:"secondary_weapons"`
}

// weaponCodePattern matches W01~W07 and sub-codes like W05-D
var weaponCodePattern = regexp.MustCompile(`W0[1-7](?:-[A-D])?`)

// extractWeaponCodesFromJSON parses any JSON map and extracts weapon codes (W01~W07, W01-A~W07-D)
// by walking all values. The first code found is primary, the rest are secondary.
// This handles groq/compound returning arbitrary key names.
func extractWeaponCodesFromJSON(content string) *aiWeaponResponse {
	var raw map[string]any
	if err := ai.ExtractJSON(content, &raw); err != nil {
		return nil
	}

	var codes []string
	seen := map[string]bool{}

	// Walk all values and collect weapon codes in order
	var walk func(v any)
	walk = func(v any) {
		switch val := v.(type) {
		case string:
			for _, m := range weaponCodePattern.FindAllString(val, -1) {
				if !seen[m] {
					seen[m] = true
					codes = append(codes, m)
				}
			}
		case []any:
			for _, item := range val {
				walk(item)
			}
		case map[string]any:
			for _, item := range val {
				walk(item)
			}
		}
	}
	walk(raw)

	if len(codes) == 0 {
		return nil
	}

	resp := &aiWeaponResponse{}
	resp.PrimaryWeapon.Code = codes[0]
	resp.PrimaryWeapon.Confidence = 0.8
	for _, code := range codes[1:] {
		resp.SecondaryWeapons = append(resp.SecondaryWeapons, struct {
			Code       string  `json:"code"`
			Confidence float64 `json:"confidence"`
			Reasoning  string  `json:"reasoning"`
		}{
			Code:       code,
			Confidence: 0.6,
		})
	}
	return resp
}

// TagExperience analyzes an experience and tags it with weapon categories.
// If content hasn't changed since last tagging, returns existing tags (skips AI call).
func (s *WeaponTaggingService) TagExperience(ctx context.Context, experienceID uuid.UUID, userID uuid.UUID) (*WeaponTagResult, error) {
	// 1. Verify experience exists and user owns it
	exp, err := s.entClient.Experience.Query().
		Where(experience.IDEQ(experienceID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrExperienceNotFound
		}
		return nil, err
	}

	if exp.UserID != userID {
		return nil, ErrExperienceForbidden
	}

	// 2. Validate experience content is long enough
	contentLength := len(exp.Title) + len(exp.StarSituation) + len(exp.StarTask) + len(exp.StarAction) + len(exp.StarResult) + len(exp.Content)
	if contentLength < 50 {
		// Return existing tags if available, even if content is too short for re-tagging
		if existing := s.loadExistingTags(ctx, exp.ID); existing != nil {
			return existing, nil
		}
		return nil, fmt.Errorf("experience text too short (minimum 50 characters)")
	}

	// 3. Load existing weapons
	existingWeapons, err := s.entClient.ExperienceWeapon.Query().
		Where(experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 4. If content hasn't changed since last tag, return existing tags (skip AI)
	if len(existingWeapons) > 0 {
		latestTag := existingWeapons[0].CreatedAt
		for _, w := range existingWeapons[1:] {
			if w.CreatedAt.After(latestTag) {
				latestTag = w.CreatedAt
			}
		}
		if !exp.UpdatedAt.After(latestTag) {
			return s.weaponsToResult(existingWeapons), nil
		}

		// If user confirmed, preserve their choice
		for _, w := range existingWeapons {
			if w.UserConfirmed {
				return s.weaponsToResult(existingWeapons), nil
			}
		}
	}

	// 5. Build experience text and call AI
	experienceText := s.buildExperienceText(exp)

	weapons, err := s.entClient.WeaponCategory.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load weapon categories: %w", err)
	}

	weaponList := s.formatWeaponList(weapons)

	// Load prompt template (with 5min cache)
	var promptTemplate *ent.PromptTemplate
	now := time.Now()
	if s.promptCache != nil && now.Sub(s.promptCachedAt) < 5*time.Minute {
		promptTemplate = s.promptCache
	} else {
		pt, err := s.entClient.PromptTemplate.Query().
			Where(
				prompttemplate.CategoryEQ("experience_classify"),
				prompttemplate.SubCategoryEQ("weapon_tagging"),
				prompttemplate.IsActiveEQ(true),
			).
			Order(prompttemplate.ByVersion(sql.OrderDesc())).
			First(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load prompt template: %w", err)
		}
		promptTemplate = pt
		s.promptCache = pt
		s.promptCachedAt = now
	}

	userPrompt := s.substituteVariables(promptTemplate.UserPromptTemplate, map[string]string{
		"experience_text":   experienceText,
		"weapon_categories": weaponList,
	})

	// 6. Call AI
	startTime := time.Now()
	aiResp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, promptTemplate.Model, ai.LLMRequest{
		SystemPrompt: promptTemplate.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  promptTemplate.Temperature,
		MaxTokens:    promptTemplate.MaxTokens,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())
	latencyMs := int(time.Since(startTime).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("AI tagging failed: %w", err)
	}

	_ = s.entClient.PromptTemplate.UpdateOneID(promptTemplate.ID).
		SetUsageCount(promptTemplate.UsageCount + 1).
		SetAvgLatencyMs((promptTemplate.AvgLatencyMs*promptTemplate.UsageCount + latencyMs) / (promptTemplate.UsageCount + 1)).
		Exec(ctx)

	// 7. Parse AI response
	var aiResult aiWeaponResponse
	if err := ai.ExtractJSON(aiResp.Content, &aiResult); err != nil || aiResult.PrimaryWeapon.Code == "" {
		if fallback := extractWeaponCodesFromJSON(aiResp.Content); fallback != nil {
			aiResult = *fallback
		} else {
			if existing := s.weaponsToResult(existingWeapons); existing != nil && existing.PrimaryWeapon.Code != "" {
				return existing, nil
			}
			return nil, fmt.Errorf("failed to parse AI response: no weapon codes found")
		}
	}

	// 8. Validate weapon codes
	validCodes := make(map[string]bool)
	for _, w := range weapons {
		validCodes[w.Code] = true
	}

	if !validCodes[aiResult.PrimaryWeapon.Code] {
		if existing := s.weaponsToResult(existingWeapons); existing != nil && existing.PrimaryWeapon.Code != "" {
			return existing, nil
		}
		return nil, fmt.Errorf("invalid weapon code from AI: %s", aiResult.PrimaryWeapon.Code)
	}

	// Filter out invalid secondary weapon codes
	var validSecondary []struct {
		Code       string  `json:"code"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}
	for _, sw := range aiResult.SecondaryWeapons {
		if validCodes[sw.Code] {
			validSecondary = append(validSecondary, sw)
		}
	}
	aiResult.SecondaryWeapons = validSecondary

	// 9. Save to DB (transaction)
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, err
	}

	_, _ = tx.ExperienceWeapon.Delete().
		Where(
			experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID)),
			experienceweapon.UserModifiedEQ(false),
		).
		Exec(ctx)

	_, err = tx.ExperienceWeapon.Create().
		SetExperienceID(exp.ID).
		SetWeaponCode(aiResult.PrimaryWeapon.Code).
		SetConfidence(aiResult.PrimaryWeapon.Confidence).
		SetIsPrimary(true).
		SetReasoning(aiResult.PrimaryWeapon.Reasoning).
		SetUserConfirmed(false).
		SetUserModified(false).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to save primary weapon: %w", err)
	}

	for _, sw := range aiResult.SecondaryWeapons {
		_, err = tx.ExperienceWeapon.Create().
			SetExperienceID(exp.ID).
			SetWeaponCode(sw.Code).
			SetConfidence(sw.Confidence).
			SetIsPrimary(false).
			SetReasoning(sw.Reasoning).
			SetUserConfirmed(false).
			SetUserModified(false).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to save secondary weapon: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	// 10. Build result
	result := &WeaponTagResult{}
	result.PrimaryWeapon.Code = aiResult.PrimaryWeapon.Code
	result.PrimaryWeapon.Confidence = aiResult.PrimaryWeapon.Confidence
	result.PrimaryWeapon.Reasoning = aiResult.PrimaryWeapon.Reasoning
	for _, sw := range aiResult.SecondaryWeapons {
		result.SecondaryWeapons = append(result.SecondaryWeapons, struct {
			Code       string
			Confidence float64
			Reasoning  string
		}{Code: sw.Code, Confidence: sw.Confidence, Reasoning: sw.Reasoning})
	}
	return result, nil
}

// loadExistingTags loads existing weapon tags for an experience, returns nil if none.
func (s *WeaponTaggingService) loadExistingTags(ctx context.Context, experienceID uuid.UUID) *WeaponTagResult {
	weapons, err := s.entClient.ExperienceWeapon.Query().
		Where(experienceweapon.HasExperienceWith(experience.IDEQ(experienceID))).
		All(ctx)
	if err != nil || len(weapons) == 0 {
		return nil
	}
	return s.weaponsToResult(weapons)
}

// weaponsToResult converts DB weapon records to WeaponTagResult.
func (s *WeaponTaggingService) weaponsToResult(weapons []*ent.ExperienceWeapon) *WeaponTagResult {
	if len(weapons) == 0 {
		return nil
	}
	result := &WeaponTagResult{}
	for _, w := range weapons {
		if w.IsPrimary {
			result.PrimaryWeapon.Code = w.WeaponCode
			result.PrimaryWeapon.Confidence = w.Confidence
			result.PrimaryWeapon.Reasoning = w.Reasoning
		} else {
			result.SecondaryWeapons = append(result.SecondaryWeapons, struct {
				Code       string
				Confidence float64
				Reasoning  string
			}{Code: w.WeaponCode, Confidence: w.Confidence, Reasoning: w.Reasoning})
		}
	}
	return result
}

// substituteVariables replaces {{variable}} placeholders in template
func (s *WeaponTaggingService) substituteVariables(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// buildExperienceText combines all STAR fields into a single text for AI analysis
func (s *WeaponTaggingService) buildExperienceText(exp *ent.Experience) string {
	var parts []string
	parts = append(parts, "제목: "+exp.Title)
	if exp.StarSituation != "" {
		parts = append(parts, "상황(Situation): "+exp.StarSituation)
	}
	if exp.StarTask != "" {
		parts = append(parts, "과제(Task): "+exp.StarTask)
	}
	if exp.StarAction != "" {
		parts = append(parts, "행동(Action): "+exp.StarAction)
	}
	if exp.StarResult != "" {
		parts = append(parts, "결과(Result): "+exp.StarResult)
	}
	if exp.Content != "" {
		parts = append(parts, "추가 내용: "+exp.Content)
	}
	return strings.Join(parts, "\n\n")
}

// formatWeaponList formats weapon categories for the AI prompt
func (s *WeaponTaggingService) formatWeaponList(weapons []*ent.WeaponCategory) string {
	var lines []string
	for _, w := range weapons {
		line := fmt.Sprintf("- %s: %s", w.Code, w.Name)
		if w.Description != "" {
			line += " - " + w.Description
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
