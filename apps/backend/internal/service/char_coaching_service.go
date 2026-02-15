package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/coverletter"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// CharCoachingService handles AI-powered character count coaching.
type CharCoachingService struct {
	entClient  *ent.Client
	aiProvider ai.LLMProvider
}

// NewCharCoachingService creates a new char coaching service.
func NewCharCoachingService(entClient *ent.Client, aiProvider ai.LLMProvider) *CharCoachingService {
	return &CharCoachingService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// CharCoachingSuggestion represents a single trim/expand suggestion.
type CharCoachingSuggestion struct {
	Type      string `json:"type"`      // "trim" | "expand"
	Section   string `json:"section"`   // which part of the text
	Original  string `json:"original"`  // current text snippet
	Suggested string `json:"suggested"` // rewritten text
	Reason    string `json:"reason"`
	CharDiff  int    `json:"char_diff"` // chars saved (negative) or added (positive)
}

// CharCoachingResult is the full char coaching response from AI.
type CharCoachingResult struct {
	Status       string                   `json:"status"`        // "over" | "under" | "good"
	CurrentCount int                      `json:"current_count"`
	CharLimit    int                      `json:"char_limit"`
	Diff         int                      `json:"diff"` // positive = over, negative = under
	Suggestions  []CharCoachingSuggestion `json:"suggestions"`
	Summary      string                   `json:"summary"`
}

// ParseCharCoachingResponse parses AI JSON output into CharCoachingResult.
func ParseCharCoachingResponse(content string) (*CharCoachingResult, error) {
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

	var result CharCoachingResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse char coaching JSON: %w", err)
	}

	// Validate status
	switch result.Status {
	case "over", "under", "good":
		// valid
	default:
		return nil, fmt.Errorf("invalid char coaching status: %q (must be over/under/good)", result.Status)
	}

	return &result, nil
}

// CoachCharCount performs AI char-count coaching on a cover letter.
func (s *CharCoachingService) CoachCharCount(
	ctx context.Context,
	userID uuid.UUID,
	coverLetterID uuid.UUID,
	content string,
) (*CharCoachingResult, error) {
	// 1. Load cover letter and verify ownership
	cl, err := s.entClient.CoverLetter.Query().
		Where(coverletter.IDEQ(coverLetterID)).
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

	// 2. Get char limit
	charLimit := 0
	if cl.CharLimit != nil {
		charLimit = *cl.CharLimit
	}
	if charLimit == 0 {
		// No char limit set — coaching not applicable
		return &CharCoachingResult{
			Status:       "good",
			CurrentCount: len([]rune(content)),
			CharLimit:    0,
			Diff:         0,
			Summary:      "글자수 제한이 설정되지 않았습니다.",
		}, nil
	}

	currentCount := len([]rune(content))
	diff := currentCount - charLimit

	// 3. Load prompt template
	prompt, err := s.entClient.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("char_coaching"),
			prompttemplate.IsActiveEQ(true),
		).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load char coaching prompt template: %w", err)
	}

	// 4. Build user prompt from template
	userPrompt := strings.ReplaceAll(prompt.UserPromptTemplate, "{{content}}", content)
	userPrompt = strings.ReplaceAll(userPrompt, "{{current_count}}", fmt.Sprintf("%d", currentCount))
	userPrompt = strings.ReplaceAll(userPrompt, "{{char_limit}}", fmt.Sprintf("%d", charLimit))
	userPrompt = strings.ReplaceAll(userPrompt, "{{diff}}", fmt.Sprintf("%d", diff))

	status := "good"
	if diff > 0 {
		status = "over"
	} else if diff < -50 {
		status = "under"
	}
	userPrompt = strings.ReplaceAll(userPrompt, "{{status}}", status)

	// 5. Call AI
	llmReq := ai.LLMRequest{
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  prompt.Temperature,
		MaxTokens:    prompt.MaxTokens,
	}

	resp, err := s.aiProvider.Call(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("AI char coaching call failed: %w", err)
	}

	// 6. Parse response
	result, err := ParseCharCoachingResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI char coaching response: %w", err)
	}

	// Override computed fields with server-side values
	result.CurrentCount = currentCount
	result.CharLimit = charLimit
	result.Diff = diff

	// 7. Record coaching session
	_, err = s.recordCharCoachingSession(ctx, userID, coverLetterID, llmReq, resp, result)
	if err != nil {
		return nil, fmt.Errorf("failed to record char coaching session: %w", err)
	}

	return result, nil
}

func (s *CharCoachingService) recordCharCoachingSession(
	ctx context.Context,
	userID uuid.UUID,
	coverLetterID uuid.UUID,
	req ai.LLMRequest,
	resp ai.LLMResponse,
	result *CharCoachingResult,
) (*ent.CoachingSession, error) {
	inputData := map[string]interface{}{
		"system": req.SystemPrompt,
		"user":   req.UserPrompt,
	}

	outputData := map[string]interface{}{
		"content": resp.Content,
		"status":  result.Status,
		"diff":    result.Diff,
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
		SetSessionType("char_coaching").
		SetInputData(inputData).
		SetOutputData(outputData).
		SetInputTokens(resp.InputTokens).
		SetOutputTokens(resp.OutputTokens).
		Save(ctx)
}
