package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// QuestionService handles question analysis
type QuestionService struct {
	entClient  *ent.Client
	aiProvider ai.LLMProvider
}

// NewQuestionService creates a new question service
func NewQuestionService(entClient *ent.Client, aiProvider ai.LLMProvider) *QuestionService {
	return &QuestionService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// QuestionAnalysisResult represents the analysis result for a question
type QuestionAnalysisResult struct {
	SurfaceQuestion       string            `json:"surface_question"`
	RealIntents           []RealIntent      `json:"real_intents"`
	RequiredWeapons       RequiredWeapons   `json:"required_weapons"`
	WritingStructure      WritingStructure  `json:"writing_structure"`
	KeyKeywords           []string          `json:"key_keywords"`
	AvoidList             []string          `json:"avoid_list"`
	GoodStructureExample  string            `json:"good_structure_example"`
}

// RealIntent represents a hidden intent behind the question
type RealIntent struct {
	Intent string `json:"intent"`
	Why    string `json:"why"`
}

// RequiredWeapons represents required competencies
type RequiredWeapons struct {
	Primary   WeaponInfo   `json:"primary"`
	Secondary []WeaponInfo `json:"secondary"`
}

// WeaponInfo represents weapon/competency information
type WeaponInfo struct {
	WeaponID   string `json:"weapon_id"`
	WeaponName string `json:"weapon_name"`
	Reason     string `json:"reason"`
}

// WritingStructure represents recommended writing structure
type WritingStructure struct {
	TotalChars int       `json:"total_chars"`
	Sections   []Section `json:"sections"`
}

// Section represents a writing section
type Section struct {
	Name      string  `json:"name"`
	CharRatio float64 `json:"char_ratio"`
	CharCount int     `json:"char_count"`
	Guide     string  `json:"guide"`
}

// AnalyzeQuestion analyzes a cover letter question using Claude
func (s *QuestionService) AnalyzeQuestion(ctx context.Context, userID uuid.UUID, applicationID uuid.UUID, questionText string, charLimit int) (*QuestionAnalysisResult, error) {
	// 1. Verify application exists and user owns it
	app, err := s.entClient.Application.Query().
		Where(application.IDEQ(applicationID)).
		WithAnalysis(). // Load company analysis if exists
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}

	if app.UserID != userID {
		return nil, ErrApplicationForbidden
	}

	// 2. Load weapon categories for context
	weapons, err := s.entClient.WeaponCategory.Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load weapon categories: %w", err)
	}

	weaponList := make([]string, len(weapons))
	for i, w := range weapons {
		weaponList[i] = fmt.Sprintf("- %s (%s): %s", w.Code, w.Name, strings.Join(w.Keywords, ", "))
	}
	weaponContext := strings.Join(weaponList, "\n")

	// 3. Extract talent keywords and values keywords from analysis if available
	talentKeywords := ""
	valuesKeywords := ""

	if app.Edges.Analysis != nil && app.Edges.Analysis.AnalysisResult != nil {
		// Extract core values from analysis_result JSON
		if coreValues, ok := app.Edges.Analysis.AnalysisResult["core_values"].([]any); ok && len(coreValues) > 0 {
			var keywords []string
			for _, v := range coreValues {
				if vm, ok := v.(map[string]any); ok {
					if keyword, ok := vm["keyword"].(string); ok {
						keywords = append(keywords, keyword)
					}
				}
			}
			valuesKeywords = strings.Join(keywords, ", ")
		}

		// Extract talent traits from analysis_result JSON
		if talentTraits, ok := app.Edges.Analysis.AnalysisResult["talent_traits"].([]any); ok && len(talentTraits) > 0 {
			var traits []string
			for _, t := range talentTraits {
				if tm, ok := t.(map[string]any); ok {
					if trait, ok := tm["trait"].(string); ok {
						traits = append(traits, trait)
					}
				}
			}
			talentKeywords = strings.Join(traits, ", ")
		}
	}

	// 4. Load prompt template
	prompt, err := s.entClient.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("question_analysis"),
			prompttemplate.IsActiveEQ(true),
		).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	// 5. Build prompt
	userPrompt := strings.ReplaceAll(prompt.UserPromptTemplate, "{{company_name}}", app.CompanyName)
	userPrompt = strings.ReplaceAll(userPrompt, "{{position}}", app.Position)
	userPrompt = strings.ReplaceAll(userPrompt, "{{talent_keywords}}", talentKeywords)
	userPrompt = strings.ReplaceAll(userPrompt, "{{values_keywords}}", valuesKeywords)
	userPrompt = strings.ReplaceAll(userPrompt, "{{weapon_categories}}", weaponContext)
	userPrompt = strings.ReplaceAll(userPrompt, "{{question_text}}", questionText)
	userPrompt = strings.ReplaceAll(userPrompt, "{{char_limit}}", fmt.Sprintf("%d", charLimit))

	// 6. Call Claude
	llmReq := ai.LLMRequest{
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  prompt.Temperature,
		MaxTokens:    prompt.MaxTokens,
	}

	resp, err := s.aiProvider.Call(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// 7. Parse response
	var result QuestionAnalysisResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 8. Validate result structure
	if len(result.RealIntents) != 3 {
		return nil, fmt.Errorf("real_intents must have exactly 3 items, got %d", len(result.RealIntents))
	}

	if result.RequiredWeapons.Primary.WeaponID == "" {
		return nil, fmt.Errorf("primary weapon must exist")
	}

	// Verify char_count sums to total_chars
	sum := 0
	for _, section := range result.WritingStructure.Sections {
		sum += section.CharCount
	}
	if sum != result.WritingStructure.TotalChars {
		return nil, fmt.Errorf("char_count sum (%d) does not match total_chars (%d)", sum, result.WritingStructure.TotalChars)
	}

	return &result, nil
}
