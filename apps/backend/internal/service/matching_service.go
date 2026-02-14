package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// MatchingService handles experience-company matching
type MatchingService struct {
	entClient  *ent.Client
	aiProvider ai.LLMProvider
}

// NewMatchingService creates a new matching service
func NewMatchingService(entClient *ent.Client, aiProvider ai.LLMProvider) *MatchingService {
	return &MatchingService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// MatchResult represents matching result for an experience
type MatchResult struct {
	ExperienceID   uuid.UUID `json:"experience_id"`
	OverallFit     int       `json:"overall_fit"`      // 0-100
	JobRelevance   int       `json:"job_relevance"`    // 0-100 (40% weight)
	TalentFit      int       `json:"talent_fit"`       // 0-100 (35% weight)
	Uniqueness     int       `json:"uniqueness"`       // 0-100 (25% weight)
	Reasoning      string    `json:"reasoning"`
	SuggestedAngle string    `json:"suggested_angle"`
}

// aiMatchResponse matches expected AI response format
type aiMatchResponse struct {
	OverallFit     int    `json:"overall_fit"`
	JobRelevance   int    `json:"job_relevance"`
	TalentFit      int    `json:"talent_fit"`
	Uniqueness     int    `json:"uniqueness"`
	Reasoning      string `json:"reasoning"`
	SuggestedAngle string `json:"suggested_angle"`
}

// MatchExperience matches a single experience against company analysis
func (s *MatchingService) MatchExperience(ctx context.Context, userID uuid.UUID, experienceID uuid.UUID, companyAnalysis *CompanyAnalysis) (*MatchResult, error) {
	// 1. Verify experience exists and user owns it
	exp, err := s.entClient.Experience.Query().
		Where(experience.IDEQ(experienceID)).
		WithWeapons().
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

	// 2. Build experience text
	experienceText := fmt.Sprintf(`제목: %s
상황: %s
과제: %s
행동: %s
결과: %s
카테고리: %s`, exp.Title, exp.StarSituation, exp.StarTask, exp.StarAction, exp.StarResult, exp.Category)

	// 3. Build company context
	companyContext := fmt.Sprintf(`회사명: %s
핵심가치: %v
인재상: %v`, companyAnalysis.CompanyName, formatCoreValues(companyAnalysis.CoreValues), formatTalentTraits(companyAnalysis.TalentTraits))

	// 4. Call AI for matching
	prompt := fmt.Sprintf(`Match this experience against the company requirements and rate fit scores.

Experience:
%s

Company Requirements:
%s

Return JSON with scores (0-100):
- overall_fit: weighted average (job_relevance*0.4 + talent_fit*0.35 + uniqueness*0.25)
- job_relevance: how relevant is this experience to the job
- talent_fit: how well does this match the talent profile
- uniqueness: differentiation factor
- reasoning: brief explanation
- suggested_angle: how to position this experience`, experienceText, companyContext)

	resp, err := s.aiProvider.Call(ctx, ai.LLMRequest{
		SystemPrompt: "You are an experience-company matching analyst. Provide objective fit scores.",
		UserPrompt:   prompt,
		Temperature:  0.2,
		MaxTokens:    1000,
		JSONMode:     true,
	})

	if err != nil {
		return nil, fmt.Errorf("AI matching failed: %w", err)
	}

	// 5. Parse AI response
	var aiResult aiMatchResponse
	if err := json.Unmarshal([]byte(resp.Content), &aiResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 6. Validate scores
	if aiResult.OverallFit < 0 || aiResult.OverallFit > 100 {
		return nil, fmt.Errorf("invalid overall_fit score: %d", aiResult.OverallFit)
	}

	// 7. Build result
	result := &MatchResult{
		ExperienceID:   exp.ID,
		OverallFit:     aiResult.OverallFit,
		JobRelevance:   aiResult.JobRelevance,
		TalentFit:      aiResult.TalentFit,
		Uniqueness:     aiResult.Uniqueness,
		Reasoning:      aiResult.Reasoning,
		SuggestedAngle: aiResult.SuggestedAngle,
	}

	return result, nil
}

func formatCoreValues(values []CoreValue) string {
	result := ""
	for _, v := range values {
		result += v.Keyword + ", "
	}
	return result
}

func formatTalentTraits(traits []TalentTrait) string {
	result := ""
	for _, t := range traits {
		result += t.Trait + ", "
	}
	return result
}
