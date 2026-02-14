package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// CoachingService handles draft coaching
type CoachingService struct {
	entClient  *ent.Client
	aiProvider ai.LLMProvider
}

// NewCoachingService creates a new coaching service
func NewCoachingService(entClient *ent.Client, aiProvider ai.LLMProvider) *CoachingService {
	return &CoachingService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// GenerateDraft generates a cover letter draft using Claude
func (s *CoachingService) GenerateDraft(
	ctx context.Context,
	userID uuid.UUID,
	applicationID uuid.UUID,
	experienceIDs []uuid.UUID,
	questionText string,
	charLimit int,
	analysisResult any, // Can be nil
) (string, error) {
	// 1. Verify application ownership
	app, err := s.entClient.Application.Query().
		Where(application.IDEQ(applicationID)).
		WithAnalysis().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", ErrApplicationNotFound
		}
		return "", err
	}

	if app.UserID != userID {
		return "", ErrApplicationForbidden
	}

	// 2. Load selected experiences
	experiences, err := s.entClient.Experience.Query().
		Where(experience.IDIn(experienceIDs...)).
		WithWeapons().
		All(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load experiences: %w", err)
	}

	// Verify all experiences belong to user
	for _, exp := range experiences {
		if exp.UserID != userID {
			return "", ErrExperienceForbidden
		}
	}

	// 3. Load prompt template
	prompt, err := s.entClient.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("draft"),
			prompttemplate.IsActiveEQ(true),
		).
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt template: %w", err)
	}

	// 4. Build experiences context
	experiencesContext := ""
	for i, exp := range experiences {
		experiencesContext += fmt.Sprintf(`### 경험 %d: %s
- 상황(S): %s
- 과제(T): %s
- 행동(A): %s
- 결과(R): %s

`, i+1, exp.Title, exp.StarSituation, exp.StarTask, exp.StarAction, exp.StarResult)
	}

	// 5. Build company context
	talentKeywords := ""
	valuesKeywords := ""
	if app.Edges.Analysis != nil && app.Edges.Analysis.AnalysisResult != nil {
		if coreValues, ok := app.Edges.Analysis.AnalysisResult["core_values"].([]any); ok {
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

		if talentTraits, ok := app.Edges.Analysis.AnalysisResult["talent_traits"].([]any); ok {
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

	// 6. Build user prompt
	userPrompt := strings.ReplaceAll(prompt.UserPromptTemplate, "{{company_name}}", app.CompanyName)
	userPrompt = strings.ReplaceAll(userPrompt, "{{position}}", app.Position)
	userPrompt = strings.ReplaceAll(userPrompt, "{{talent_keywords}}", talentKeywords)
	userPrompt = strings.ReplaceAll(userPrompt, "{{values_keywords}}", valuesKeywords)
	userPrompt = strings.ReplaceAll(userPrompt, "{{question_text}}", questionText)
	userPrompt = strings.ReplaceAll(userPrompt, "{{char_limit}}", fmt.Sprintf("%d", charLimit))
	userPrompt = strings.ReplaceAll(userPrompt, "{{experiences}}", experiencesContext)

	// 7. Call Claude
	llmReq := ai.LLMRequest{
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  prompt.Temperature,
		MaxTokens:    prompt.MaxTokens,
	}

	resp, err := s.aiProvider.Call(ctx, llmReq)
	if err != nil {
		return "", fmt.Errorf("AI call failed: %w", err)
	}

	return resp.Content, nil
}
