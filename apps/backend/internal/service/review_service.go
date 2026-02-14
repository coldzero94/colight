package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/coverletter"
	"github.com/coby/colight/apps/backend/ent/coverletterversion"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// ReviewService handles AI-powered cover letter review.
type ReviewService struct {
	entClient  *ent.Client
	aiProvider ai.LLMProvider
}

// NewReviewService creates a new review service.
func NewReviewService(entClient *ent.Client, aiProvider ai.LLMProvider) *ReviewService {
	return &ReviewService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// ReviewScores contains 4-dimension scores.
type ReviewScores struct {
	Specificity  int `json:"specificity"`
	JobFit       int `json:"job_fit"`
	CompanyFit   int `json:"company_fit"`
	Authenticity int `json:"authenticity"`
}

// DimensionFeedback contains per-dimension detailed feedback.
type DimensionFeedback struct {
	Dimension string   `json:"dimension"`
	Score     int      `json:"score"`
	Good      []string `json:"good"`
	Improve   []string `json:"improve"`
}

// SpecificSuggestion represents a line-level edit suggestion.
type SpecificSuggestion struct {
	Original  string `json:"original"`
	Suggested string `json:"suggested"`
	Reason    string `json:"reason"`
}

// ReviewResult is the full review response from AI.
type ReviewResult struct {
	Scores               ReviewScores         `json:"scores"`
	Overall              int                  `json:"overall"`
	PerDimensionFeedback []DimensionFeedback  `json:"per_dimension_feedback"`
	SpecificSuggestions  []SpecificSuggestion `json:"specific_suggestions"`
}

// ParseReviewResponse parses AI JSON output into ReviewResult and validates scores.
func ParseReviewResponse(content string) (*ReviewResult, error) {
	// Strip markdown code fences if present
	cleaned := strings.TrimSpace(content)
	if strings.HasPrefix(cleaned, "```") {
		if idx := strings.Index(cleaned, "\n"); idx != -1 {
			cleaned = cleaned[idx+1:]
		}
		if idx := strings.LastIndex(cleaned, "```"); idx != -1 {
			cleaned = cleaned[:idx]
		}
		cleaned = strings.TrimSpace(cleaned)
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse review JSON: %w", err)
	}

	if err := validateScore("specificity", result.Scores.Specificity); err != nil {
		return nil, err
	}
	if err := validateScore("job_fit", result.Scores.JobFit); err != nil {
		return nil, err
	}
	if err := validateScore("company_fit", result.Scores.CompanyFit); err != nil {
		return nil, err
	}
	if err := validateScore("authenticity", result.Scores.Authenticity); err != nil {
		return nil, err
	}
	if err := validateScore("overall", result.Overall); err != nil {
		return nil, err
	}

	return &result, nil
}

func validateScore(name string, score int) error {
	if score < 0 || score > 100 {
		return fmt.Errorf("invalid %s score: %d (must be 0-100)", name, score)
	}
	return nil
}

// ReviewCoverLetter performs AI review on a cover letter and saves feedback.
func (s *ReviewService) ReviewCoverLetter(
	ctx context.Context,
	userID uuid.UUID,
	coverLetterID uuid.UUID,
	content string,
) (*ReviewResult, error) {
	// 1. Load cover letter and verify ownership
	cl, err := s.entClient.CoverLetter.Query().
		Where(coverletter.IDEQ(coverLetterID)).
		WithApplication(func(q *ent.ApplicationQuery) {
			q.WithAnalysis()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCoverLetterNotFound
		}
		return nil, fmt.Errorf("failed to load cover letter: %w", err)
	}

	if cl.UserID != userID {
		return nil, ErrCoverLetterForbidden
	}

	// 2. Load prompt template
	prompt, err := s.entClient.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("review"),
			prompttemplate.IsActiveEQ(true),
		).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load review prompt template: %w", err)
	}

	// 3. Build company context from analysis
	companyContext := ""
	if cl.Edges.Application != nil {
		app := cl.Edges.Application
		companyContext = fmt.Sprintf("기업: %s\n직무: %s", app.CompanyName, app.Position)

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
				if len(keywords) > 0 {
					companyContext += fmt.Sprintf("\n핵심가치: %s", strings.Join(keywords, ", "))
				}
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
				if len(traits) > 0 {
					companyContext += fmt.Sprintf("\n인재상: %s", strings.Join(traits, ", "))
				}
			}
		}
	}

	// 4. Build user prompt from template
	userPrompt := strings.ReplaceAll(prompt.UserPromptTemplate, "{{content}}", content)
	userPrompt = strings.ReplaceAll(userPrompt, "{{question_text}}", cl.QuestionText)
	userPrompt = strings.ReplaceAll(userPrompt, "{{company_context}}", companyContext)
	charLimit := 0
	if cl.CharLimit != nil {
		charLimit = *cl.CharLimit
	}
	userPrompt = strings.ReplaceAll(userPrompt, "{{char_limit}}", fmt.Sprintf("%d", charLimit))

	// 5. Call AI
	llmReq := ai.LLMRequest{
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  prompt.Temperature,
		MaxTokens:    prompt.MaxTokens,
	}

	resp, err := s.aiProvider.Call(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("AI review call failed: %w", err)
	}

	// 6. Parse response
	result, err := ParseReviewResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI review response: %w", err)
	}

	// 7. Save feedback to latest version
	latestVersion, err := s.entClient.CoverLetterVersion.Query().
		Where(coverletterversion.HasCoverLetterWith(coverletter.IDEQ(coverLetterID))).
		Order(ent.Desc(coverletterversion.FieldVersionNumber)).
		First(ctx)
	if err == nil {
		feedbackMap := map[string]interface{}{
			"scores":                 result.Scores,
			"overall":               result.Overall,
			"per_dimension_feedback": result.PerDimensionFeedback,
			"specific_suggestions":   result.SpecificSuggestions,
		}
		scoresMap := map[string]interface{}{
			"specificity":  result.Scores.Specificity,
			"job_fit":      result.Scores.JobFit,
			"company_fit":  result.Scores.CompanyFit,
			"authenticity": result.Scores.Authenticity,
			"overall":      result.Overall,
		}

		_, err = latestVersion.Update().
			SetFeedback(feedbackMap).
			SetScores(scoresMap).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to save review feedback: %w", err)
		}
	}

	// 8. Record coaching session
	_, err = s.recordReviewSession(ctx, userID, coverLetterID, llmReq, resp, result)
	if err != nil {
		return nil, fmt.Errorf("failed to record review session: %w", err)
	}

	return result, nil
}

func (s *ReviewService) recordReviewSession(
	ctx context.Context,
	userID uuid.UUID,
	coverLetterID uuid.UUID,
	req ai.LLMRequest,
	resp ai.LLMResponse,
	result *ReviewResult,
) (*ent.CoachingSession, error) {
	inputData := map[string]interface{}{
		"system": req.SystemPrompt,
		"user":   req.UserPrompt,
	}

	outputData := map[string]interface{}{
		"content": resp.Content,
		"scores":  result.Scores,
		"overall": result.Overall,
		"usage": map[string]interface{}{
			"input_tokens":  resp.InputTokens,
			"output_tokens": resp.OutputTokens,
		},
	}

	cl, err := s.entClient.CoverLetter.Get(ctx, coverLetterID)
	if err != nil {
		return nil, err
	}

	return s.entClient.CoachingSession.Create().
		SetUserID(userID).
		SetCoverLetter(cl).
		SetSessionType("review").
		SetInputData(inputData).
		SetOutputData(outputData).
		SetInputTokens(resp.InputTokens).
		SetOutputTokens(resp.OutputTokens).
		Save(ctx)
}
