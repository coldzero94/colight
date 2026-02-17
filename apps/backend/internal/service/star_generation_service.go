package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
)

// StarGenerationService generates STAR fields from free-form content using AI
type StarGenerationService struct {
	entClient  *ent.Client
	aiProvider *ai.AIProvider
}

// NewStarGenerationService creates a new StarGenerationService
func NewStarGenerationService(entClient *ent.Client, aiProvider *ai.AIProvider) *StarGenerationService {
	return &StarGenerationService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// StarGenerationRequest is the input for STAR generation
type StarGenerationRequest struct {
	Content string `json:"content"`
	Title   string `json:"title"`
}

// StarGenerationResult is the AI-generated STAR fields
type StarGenerationResult struct {
	Situation string `json:"star_situation"`
	Task      string `json:"star_task"`
	Action    string `json:"star_action"`
	Result    string `json:"star_result"`
}

// GenerateSTAR generates STAR fields from free-form content
func (s *StarGenerationService) GenerateSTAR(ctx context.Context, req StarGenerationRequest) (*StarGenerationResult, error) {
	if len(req.Content) < 30 {
		return nil, fmt.Errorf("content too short (minimum 30 characters)")
	}

	pt, err := s.loadStarPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	userPrompt := pt.UserPromptTemplate
	for k, v := range map[string]string{
		"title":   req.Title,
		"content": req.Content,
	} {
		userPrompt = strings.ReplaceAll(userPrompt, "{{"+k+"}}", v)
	}

	startTime := time.Now()
	aiResp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, pt.Model, ai.LLMRequest{
		SystemPrompt: pt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  pt.Temperature,
		MaxTokens:    pt.MaxTokens,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())
	if err != nil {
		return nil, fmt.Errorf("AI STAR generation failed: %w", err)
	}

	if s.entClient != nil && pt.ID.String() != "00000000-0000-0000-0000-000000000000" {
		latencyMs := int(time.Since(startTime).Milliseconds())
		_ = s.entClient.PromptTemplate.UpdateOneID(pt.ID).
			SetUsageCount(pt.UsageCount + 1).
			SetAvgLatencyMs((pt.AvgLatencyMs*pt.UsageCount + latencyMs) / (pt.UsageCount + 1)).
			Exec(ctx)
	}

	// Try standard format
	var result StarGenerationResult
	if err := ai.ExtractJSON(aiResp.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Also handle Korean keys as fallback
	if result.Situation == "" && result.Task == "" {
		var koreanResult struct {
			Situation string `json:"상황"`
			Task      string `json:"과제"`
			Action    string `json:"행동"`
			Result    string `json:"결과"`
		}
		if err := ai.ExtractJSON(aiResp.Content, &koreanResult); err == nil && koreanResult.Situation != "" {
			result.Situation = koreanResult.Situation
			result.Task = koreanResult.Task
			result.Action = koreanResult.Action
			result.Result = koreanResult.Result
		}
	}

	if result.Situation == "" && result.Task == "" && result.Action == "" && result.Result == "" {
		return nil, fmt.Errorf("AI returned empty STAR fields")
	}

	return &result, nil
}

// loadStarPrompt loads the STAR generation prompt template from DB.
// Falls back to hardcoded defaults when DB is unavailable.
func (s *StarGenerationService) loadStarPrompt(ctx context.Context) (*ent.PromptTemplate, error) {
	if s.entClient != nil {
		pt, err := s.entClient.PromptTemplate.Query().
			Where(
				prompttemplate.CategoryEQ("star_generation"),
				prompttemplate.SubCategoryEQ("generate"),
				prompttemplate.IsActiveEQ(true),
			).
			Order(prompttemplate.ByVersion(sql.OrderDesc())).
			First(ctx)
		if err == nil {
			return pt, nil
		}
	}
	return &ent.PromptTemplate{
		Model: "gemini-2.0-flash",
		SystemPrompt: "당신은 취업 준비생의 자유 형식 경험 텍스트를 STAR 기법으로 구조화하는 AI입니다.\n오직 JSON만 출력하세요.\n\n응답 형식:\n{\"star_situation\": \"상황\", \"star_task\": \"과제\", \"star_action\": \"행동\", \"star_result\": \"결과\"}\n\n규칙:\n- 원문의 핵심 내용을 보존하되 STAR 구조로 재배치\n- 각 필드는 2~4문장\n- 원문에 없는 내용을 지어내지 말 것\n- 한국어로 작성",
		UserPromptTemplate: "제목: {{title}}\n\n내용:\n{{content}}\n\nJSON만 출력하세요.",
		Temperature:        0.3,
		MaxTokens:          1500,
	}, nil
}
