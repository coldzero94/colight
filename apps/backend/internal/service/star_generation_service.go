package service

import (
	"context"
	"fmt"
	"log"

	"github.com/coby/colight/apps/backend/ent"
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

	systemPrompt := `당신은 취업 준비생의 자유 형식 경험 텍스트를 STAR 기법으로 구조화하는 AI입니다.
오직 JSON만 출력하세요. 설명이나 해설 없이 JSON만 반환하세요.

응답 형식:
{
  "star_situation": "상황 설명 (배경, 맥락, 시기, 조직)",
  "star_task": "과제/목표 설명 (해결해야 할 문제, 기대 성과)",
  "star_action": "구체적 행동 (전략, 실행한 것, 수치 포함)",
  "star_result": "결과 (정량적 성과, 질적 변화, 배운 점)"
}

규칙:
- 원문의 핵심 내용을 보존하되 STAR 구조로 재배치
- 각 필드는 2~4문장으로 구성
- 원문에 없는 내용을 지어내지 말 것
- 한국어로 작성`

	userPrompt := req.Content
	if req.Title != "" {
		userPrompt = fmt.Sprintf("제목: %s\n\n내용:\n%s", req.Title, req.Content)
	}

	aiResp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, "groq/compound", ai.LLMRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt + "\n\nJSON만 출력하세요.",
		Temperature:  0.3,
		MaxTokens:    1500,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())
	if err != nil {
		return nil, fmt.Errorf("AI STAR generation failed: %w", err)
	}

	log.Printf("[star-gen] raw AI response (model=%s): %s", aiResp.Model, aiResp.Content)

	// Try standard format
	var result StarGenerationResult
	if err := ai.ExtractJSON(aiResp.Content, &result); err != nil {
		log.Printf("[star-gen] ExtractJSON failed: %v", err)
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
