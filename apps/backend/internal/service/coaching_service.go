package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/ent/coachingsession"
	"github.com/coby/colight/apps/backend/ent/coverletter"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// CoachingService handles draft coaching
type CoachingService struct {
	entClient  *ent.Client
	aiProvider *ai.AIProvider
}

// NewCoachingService creates a new coaching service
func NewCoachingService(entClient *ent.Client, aiProvider *ai.AIProvider) *CoachingService {
	return &CoachingService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// buildDraftRequest builds the LLM request for draft generation (shared by sync and streaming).
// Returns the LLM request and the model name from the prompt template.
func (s *CoachingService) buildDraftRequest(
	ctx context.Context,
	userID uuid.UUID,
	applicationID uuid.UUID,
	experienceIDs []uuid.UUID,
	questionText string,
	charLimit int,
) (ai.LLMRequest, string, error) {
	// 1. Verify application ownership
	app, err := s.entClient.Application.Query().
		Where(application.IDEQ(applicationID)).
		WithAnalysis().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ai.LLMRequest{}, "", ErrApplicationNotFound
		}
		return ai.LLMRequest{}, "", err
	}

	if app.UserID != userID {
		return ai.LLMRequest{}, "", ErrApplicationForbidden
	}

	// 2. Load selected experiences
	experiences, err := s.entClient.Experience.Query().
		Where(experience.IDIn(experienceIDs...)).
		WithWeapons().
		All(ctx)
	if err != nil {
		return ai.LLMRequest{}, "", fmt.Errorf("failed to load experiences: %w", err)
	}

	for _, exp := range experiences {
		if exp.UserID != userID {
			return ai.LLMRequest{}, "", ErrExperienceForbidden
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
		return ai.LLMRequest{}, "", fmt.Errorf("failed to load prompt template: %w", err)
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

	return ai.LLMRequest{
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  prompt.Temperature,
		MaxTokens:    prompt.MaxTokens,
	}, prompt.Model, nil
}

// GenerateDraft generates a cover letter draft synchronously.
func (s *CoachingService) GenerateDraft(
	ctx context.Context,
	userID uuid.UUID,
	applicationID uuid.UUID,
	experienceIDs []uuid.UUID,
	questionText string,
	charLimit int,
	analysisResult any,
) (string, error) {
	llmReq, modelName, err := s.buildDraftRequest(ctx, userID, applicationID, experienceIDs, questionText, charLimit)
	if err != nil {
		return "", err
	}

	resp, err := s.aiProvider.CallByModelName(ctx, modelName, llmReq)
	if err != nil {
		return "", fmt.Errorf("AI call failed: %w", err)
	}

	return resp.Content, nil
}

// GenerateDraftStream generates a cover letter draft with streaming.
// Calls onChunk for each text delta. Returns final LLMResponse with token usage.
// Streaming always uses the heavy model (Claude) as it's the only streaming provider.
func (s *CoachingService) GenerateDraftStream(
	ctx context.Context,
	userID uuid.UUID,
	applicationID uuid.UUID,
	experienceIDs []uuid.UUID,
	questionText string,
	charLimit int,
	analysisResult any,
	onChunk ai.StreamCallback,
) (ai.LLMResponse, error) {
	llmReq, _, err := s.buildDraftRequest(ctx, userID, applicationID, experienceIDs, questionText, charLimit)
	if err != nil {
		return ai.LLMResponse{}, err
	}

	streaming := s.aiProvider.HeavyStreaming()
	if streaming == nil {
		return ai.LLMResponse{}, fmt.Errorf("streaming provider not available")
	}

	resp, err := streaming.Stream(ctx, llmReq, onChunk)
	if err != nil {
		return ai.LLMResponse{}, fmt.Errorf("AI streaming failed: %w", err)
	}

	return resp, nil
}

// DraftResult contains IDs created after a successful draft generation.
type DraftResult struct {
	CoverLetterID uuid.UUID
	SessionID     uuid.UUID
}

// SaveDraftResult creates cover_letter + version + coaching_session after streaming completes.
func (s *CoachingService) SaveDraftResult(
	ctx context.Context,
	userID uuid.UUID,
	applicationID uuid.UUID,
	questionText string,
	charLimit int,
	content string,
	inputTokens int,
	outputTokens int,
) (*DraftResult, error) {
	// 1. Create cover letter
	cl, err := s.entClient.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(applicationID).
		SetQuestionText(questionText).
		SetCharLimit(charLimit).
		SetCurrentContent(content).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cover letter: %w", err)
	}

	// 2. Create version 1
	_, err = s.entClient.CoverLetterVersion.Create().
		SetCoverLetter(cl).
		SetVersionNumber(1).
		SetContent(content).
		SetCharCount(len([]rune(content))).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// 3. Record coaching session
	session, err := s.RecordSession(ctx, userID, cl.ID, "draft", "", "", content, inputTokens, outputTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to record session: %w", err)
	}

	return &DraftResult{
		CoverLetterID: cl.ID,
		SessionID:     session.ID,
	}, nil
}

// RecordSession records a coaching session to coaching_sessions table
func (s *CoachingService) RecordSession(
	ctx context.Context,
	userID uuid.UUID,
	coverLetterID uuid.UUID,
	sessionType string,
	systemPrompt string,
	userPrompt string,
	assistantResponse string,
	inputTokens int,
	outputTokens int,
) (*ent.CoachingSession, error) {
	// Build input/output data
	inputData := map[string]interface{}{
		"system": systemPrompt,
		"user":   userPrompt,
	}

	outputData := map[string]interface{}{
		"content": assistantResponse,
		"usage": map[string]interface{}{
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
		},
	}

	// Load cover letter for edge
	coverLetter, err := s.entClient.CoverLetter.Get(ctx, coverLetterID)
	if err != nil {
		return nil, fmt.Errorf("cover letter not found: %w", err)
	}

	// Create session
	session, err := s.entClient.CoachingSession.Create().
		SetUserID(userID).
		SetCoverLetter(coverLetter).
		SetSessionType(coachingsession.SessionType(sessionType)).
		SetInputData(inputData).
		SetOutputData(outputData).
		SetInputTokens(inputTokens).
		SetOutputTokens(outputTokens).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to record session: %w", err)
	}

	return session, nil
}

// GetSessions retrieves coaching sessions for a cover letter
func (s *CoachingService) GetSessions(ctx context.Context, userID uuid.UUID, coverLetterID uuid.UUID) ([]*ent.CoachingSession, error) {
	sessions, err := s.entClient.CoachingSession.Query().
		Where(
			coachingsession.UserIDEQ(userID),
			coachingsession.HasCoverLetterWith(coverletter.IDEQ(coverLetterID)),
		).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}

	return sessions, nil
}
