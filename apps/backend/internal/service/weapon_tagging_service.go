package service

import (
	"context"
	"encoding/json"
	"fmt"
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
	entClient *ent.Client
	aiClient  ai.LLMProvider
}

// NewWeaponTaggingService creates a new WeaponTaggingService
func NewWeaponTaggingService(entClient *ent.Client, aiClient ai.LLMProvider) *WeaponTaggingService {
	return &WeaponTaggingService{
		entClient: entClient,
		aiClient:  aiClient,
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

// TagExperience analyzes an experience and tags it with weapon categories
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

	// 2. Validate experience content is long enough (count actual content only)
	contentLength := len(exp.Title) + len(exp.StarSituation) + len(exp.StarTask) + len(exp.StarAction) + len(exp.StarResult) + len(exp.Content)
	if contentLength < 50 {
		return nil, fmt.Errorf("experience text too short (minimum 50 characters)")
	}

	experienceText := s.buildExperienceText(exp)

	// 3. Load weapon categories from DB
	weapons, err := s.entClient.WeaponCategory.Query().
		Where().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load weapon categories: %w", err)
	}

	weaponList := s.formatWeaponList(weapons)

	// 4. Load prompt template from DB
	promptTemplate, err := s.entClient.PromptTemplate.Query().
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

	// 5. Substitute variables in user prompt
	userPrompt := s.substituteVariables(promptTemplate.UserPromptTemplate, map[string]string{
		"experience_text":   experienceText,
		"weapon_categories": weaponList,
	})

	// 6. Call AI with DB-loaded prompt
	startTime := time.Now()
	aiResp, err := s.aiClient.Call(ctx, ai.LLMRequest{
		SystemPrompt: promptTemplate.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  promptTemplate.Temperature,
		MaxTokens:    promptTemplate.MaxTokens,
		JSONMode:     true,
	})
	latencyMs := int(time.Since(startTime).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("AI tagging failed: %w", err)
	}

	// 7. Update prompt usage stats
	_ = s.entClient.PromptTemplate.UpdateOneID(promptTemplate.ID).
		SetUsageCount(promptTemplate.UsageCount + 1).
		SetAvgLatencyMs((promptTemplate.AvgLatencyMs*promptTemplate.UsageCount + latencyMs) / (promptTemplate.UsageCount + 1)).
		Exec(ctx)

	// 8. Parse AI response
	var aiResult aiWeaponResponse
	if err := json.Unmarshal([]byte(aiResp.Content), &aiResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 6. Validate weapon codes
	validCodes := make(map[string]bool)
	for _, w := range weapons {
		validCodes[w.Code] = true
	}

	if !validCodes[aiResult.PrimaryWeapon.Code] {
		return nil, fmt.Errorf("invalid weapon code from AI: %s", aiResult.PrimaryWeapon.Code)
	}

	for _, sw := range aiResult.SecondaryWeapons {
		if !validCodes[sw.Code] {
			return nil, fmt.Errorf("invalid weapon code from AI: %s", sw.Code)
		}
	}

	// 7. Check if user has confirmed any weapons
	existingWeapons, err := s.entClient.ExperienceWeapon.Query().
		Where(experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// If any weapon is user-confirmed, skip re-tagging (preserve user's choice)
	for _, w := range existingWeapons {
		if w.UserConfirmed {
			// User has confirmed the classification - don't override
			// But still return the AI result for reference
			result := &WeaponTagResult{}
			result.PrimaryWeapon.Code = aiResult.PrimaryWeapon.Code
			result.PrimaryWeapon.Confidence = aiResult.PrimaryWeapon.Confidence
			result.PrimaryWeapon.Reasoning = aiResult.PrimaryWeapon.Reasoning
			for _, sw := range aiResult.SecondaryWeapons {
				result.SecondaryWeapons = append(result.SecondaryWeapons, struct {
					Code       string
					Confidence float64
					Reasoning  string
				}{
					Code:       sw.Code,
					Confidence: sw.Confidence,
					Reasoning:  sw.Reasoning,
				})
			}
			return result, nil
		}
	}

	// 8. Delete existing AI-tagged weapons and insert new ones (transaction)
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, err
	}

	// Delete only AI-generated weapons (user_modified=false)
	_, _ = tx.ExperienceWeapon.Delete().
		Where(
			experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID)),
			experienceweapon.UserModifiedEQ(false),
		).
		Exec(ctx)

	// Insert primary weapon
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

	// Insert secondary weapons
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

	// 8. Build result
	result := &WeaponTagResult{}
	result.PrimaryWeapon.Code = aiResult.PrimaryWeapon.Code
	result.PrimaryWeapon.Confidence = aiResult.PrimaryWeapon.Confidence
	result.PrimaryWeapon.Reasoning = aiResult.PrimaryWeapon.Reasoning

	for _, sw := range aiResult.SecondaryWeapons {
		result.SecondaryWeapons = append(result.SecondaryWeapons, struct {
			Code       string
			Confidence float64
			Reasoning  string
		}{
			Code:       sw.Code,
			Confidence: sw.Confidence,
			Reasoning:  sw.Reasoning,
		})
	}

	return result, nil
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
