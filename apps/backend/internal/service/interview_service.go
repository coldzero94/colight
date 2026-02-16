package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// InterviewStage represents a stage of the interview.
type InterviewStage string

const (
	StageWarmup    InterviewStage = "warmup"
	StageMemory    InterviewStage = "memory"
	StageChallenge InterviewStage = "challenge"
	StageSolution  InterviewStage = "solution"
	StageOutcome   InterviewStage = "outcome"
)

// InterviewStages defines the ordered progression.
var InterviewStages = []InterviewStage{StageWarmup, StageMemory, StageChallenge, StageSolution, StageOutcome}

var stageKorean = map[InterviewStage]string{
	StageWarmup:    "가볍게",
	StageMemory:    "기억에 남는 순간",
	StageChallenge: "어려웠던 점",
	StageSolution:  "해결법",
	StageOutcome:   "결과/배운 점",
}

var stageInstruction = map[InterviewStage]string{
	StageWarmup:    "최근 활동이나 관심사에 대해 가볍게 물어보세요. 긴장을 풀어주세요.",
	StageMemory:    "가장 기억에 남는 구체적인 경험을 발굴하세요. 어떤 상황이었는지 자세히 물어보세요.",
	StageChallenge: "그 경험에서 어려웠던 점이나 도전 과제를 파악하세요.",
	StageSolution:  "어려움을 어떻게 해결했는지, 실제로 어떤 행동을 취했는지 물어보세요.",
	StageOutcome:   "최종 결과와 그 경험에서 배운 점을 정리하세요.",
}

func validStage(s InterviewStage) bool {
	for _, stage := range InterviewStages {
		if stage == s {
			return true
		}
	}
	return false
}

// ChatMessage represents a single message in the interview conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// GenerateQuestionInput is the input for generating the next interview question.
type GenerateQuestionInput struct {
	Stage    InterviewStage `json:"stage"`
	Messages []ChatMessage  `json:"messages"`
}

// GenerateQuestionResult is the AI's response with the next question.
type GenerateQuestionResult struct {
	Question   string         `json:"question"`
	Stage      InterviewStage `json:"stage"`
	NextStage  InterviewStage `json:"next_stage,omitempty"`
	IsComplete bool           `json:"is_complete"`
}

// ExtractSTARResult holds the STAR-structured extraction from an interview.
type ExtractSTARResult struct {
	Title         string   `json:"title"`
	Category      string   `json:"category"`
	Content       string   `json:"content"`
	Result        string   `json:"result"`
	StarSituation string   `json:"star_situation"`
	StarTask      string   `json:"star_task"`
	StarAction    string   `json:"star_action"`
	StarResult    string   `json:"star_result"`
	Keywords      []string `json:"keywords"`
}

// SaveExperienceInput is the request body for saving an interview experience.
type SaveExperienceInput struct {
	Title         string   `json:"title"`
	Category      string   `json:"category"`
	Content       string   `json:"content"`
	Result        string   `json:"result"`
	StarSituation string   `json:"star_situation"`
	StarTask      string   `json:"star_task"`
	StarAction    string   `json:"star_action"`
	StarResult    string   `json:"star_result"`
	Keywords      []string `json:"keywords"`
}

// SaveExperienceResult is the response for saving an interview experience.
type SaveExperienceResult struct {
	ExperienceID string `json:"experience_id"`
	Tagged       bool   `json:"tagged"`
}

// InterviewService handles AI-powered experience interviews.
type InterviewService struct {
	aiProvider           *ai.AIProvider
	db                   *ent.Client
	weaponTaggingService *WeaponTaggingService
}

// NewInterviewService creates a new interview service.
func NewInterviewService(aiProvider *ai.AIProvider, db *ent.Client, weaponTaggingService *WeaponTaggingService) *InterviewService {
	return &InterviewService{aiProvider: aiProvider, db: db, weaponTaggingService: weaponTaggingService}
}

const interviewSystemPrompt = `당신은 취업 준비생의 경험을 발굴하는 친절한 AI 인터뷰어입니다.
한국어로 대화하며, 자연스럽고 편안한 톤으로 질문합니다.
한 번에 하나의 질문만 합니다. 질문은 간결하게 2-3문장 이내로 합니다.

현재 인터뷰 단계: {{stage_name}}
단계 지시사항: {{stage_instruction}}

반드시 아래 JSON 형식으로만 응답하세요:
{"question": "질문 내용"}`

// GenerateQuestion produces the next AI interviewer question.
func (s *InterviewService) GenerateQuestion(ctx context.Context, _ uuid.UUID, input GenerateQuestionInput) (*GenerateQuestionResult, error) {
	if !validStage(input.Stage) {
		return nil, ErrInvalidInterviewStage
	}

	// Build conversation history
	var historyLines []string
	for _, msg := range input.Messages {
		prefix := "사용자"
		if msg.Role == "assistant" {
			prefix = "AI"
		}
		historyLines = append(historyLines, fmt.Sprintf("%s: %s", prefix, msg.Content))
	}
	history := strings.Join(historyLines, "\n")

	// Build prompt
	systemPrompt := strings.ReplaceAll(interviewSystemPrompt, "{{stage_name}}", stageKorean[input.Stage])
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{stage_instruction}}", stageInstruction[input.Stage])

	userPrompt := "새 인터뷰를 시작합니다."
	if len(input.Messages) > 0 {
		userPrompt = fmt.Sprintf("대화 기록:\n%s\n\n위 대화를 바탕으로 다음 질문을 생성하세요.", history)
	}

	resp, err := s.aiProvider.CallByModelName(ctx, "groq", ai.LLMRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    300,
		JSONMode:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse response
	var aiResp struct {
		Question string `json:"question"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	if aiResp.Question == "" {
		return nil, fmt.Errorf("AI returned empty question")
	}

	// Determine next stage
	stageIdx := stageIndex(input.Stage)
	isComplete := stageIdx >= len(InterviewStages)-1
	var nextStage InterviewStage
	if !isComplete {
		nextStage = InterviewStages[stageIdx+1]
	}

	return &GenerateQuestionResult{
		Question:   aiResp.Question,
		Stage:      input.Stage,
		NextStage:  nextStage,
		IsComplete: isComplete,
	}, nil
}

// ParseStage determines the current interview stage from message count.
func ParseStage(messages []ChatMessage) InterviewStage {
	userMsgCount := 0
	for _, msg := range messages {
		if msg.Role == "user" {
			userMsgCount++
		}
	}

	switch {
	case userMsgCount <= 1:
		return StageWarmup
	case userMsgCount <= 3:
		return StageMemory
	case userMsgCount <= 5:
		return StageChallenge
	case userMsgCount <= 7:
		return StageSolution
	default:
		return StageOutcome
	}
}

func stageIndex(s InterviewStage) int {
	for i, stage := range InterviewStages {
		if stage == s {
			return i
		}
	}
	return 0
}

const extractSTARPrompt = `당신은 인터뷰 대화에서 경험을 STAR 구조로 추출하는 전문가입니다.
아래 대화를 분석하여 핵심 경험을 STAR 구조로 정리하세요.

규칙:
- 모든 필드를 한국어로 작성
- title: 경험을 한 줄로 요약 (20자 이내)
- category: 다음 중 하나 — project, work, activity, competition, education, volunteer, other
- content: 경험의 전체적인 설명 (2-3문장)
- result: 최종 결과 요약 (1-2문장)
- star_situation: 상황 설명
- star_task: 해결해야 할 과제/목표
- star_action: 실제 취한 행동
- star_result: 행동의 결과와 배운 점
- keywords: 핵심 키워드 3-5개 배열

반드시 아래 JSON 형식으로만 응답하세요:
{"title":"","category":"","content":"","result":"","star_situation":"","star_task":"","star_action":"","star_result":"","keywords":[]}`

// ExtractSTAR extracts STAR-structured experience from interview messages.
func (s *InterviewService) ExtractSTAR(ctx context.Context, messages []ChatMessage) (*ExtractSTARResult, error) {
	var historyLines []string
	for _, msg := range messages {
		prefix := "사용자"
		if msg.Role == "assistant" {
			prefix = "AI"
		}
		historyLines = append(historyLines, fmt.Sprintf("%s: %s", prefix, msg.Content))
	}
	history := strings.Join(historyLines, "\n")

	resp, err := s.aiProvider.CallByModelName(ctx, "groq", ai.LLMRequest{
		SystemPrompt: extractSTARPrompt,
		UserPrompt:   fmt.Sprintf("인터뷰 대화:\n%s\n\n위 대화에서 STAR 구조를 추출하세요.", history),
		Temperature:  0.3,
		MaxTokens:    800,
		JSONMode:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	var result ExtractSTARResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse STAR response: %w", err)
	}

	// Validate non-empty STAR fields
	if result.StarSituation == "" || result.StarTask == "" || result.StarAction == "" || result.StarResult == "" {
		return nil, ErrEmptySTARField
	}

	return &result, nil
}

// SaveExperience persists the STAR-extracted experience to the database.
func (s *InterviewService) SaveExperience(ctx context.Context, userID uuid.UUID, input SaveExperienceInput) (*SaveExperienceResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database client not configured")
	}

	builder := s.db.Experience.Create().
		SetUserID(userID).
		SetTitle(input.Title).
		SetContent(input.Content).
		SetSource("interview")

	if input.Category != "" {
		builder = builder.SetCategory(input.Category)
	}
	if input.Result != "" {
		builder = builder.SetResult(input.Result)
	}
	if input.StarSituation != "" {
		builder = builder.SetStarSituation(input.StarSituation)
	}
	if input.StarTask != "" {
		builder = builder.SetStarTask(input.StarTask)
	}
	if input.StarAction != "" {
		builder = builder.SetStarAction(input.StarAction)
	}
	if input.StarResult != "" {
		builder = builder.SetStarResult(input.StarResult)
	}
	if len(input.Keywords) > 0 {
		builder = builder.SetKeywords(input.Keywords)
	}

	exp, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to save experience: %w", err)
	}

	// Auto-tag with weapons if service is available
	tagged := false
	if s.weaponTaggingService != nil {
		if _, tagErr := s.weaponTaggingService.TagExperience(ctx, exp.ID, userID); tagErr == nil {
			tagged = true
		}
	}

	return &SaveExperienceResult{
		ExperienceID: exp.ID.String(),
		Tagged:       tagged,
	}, nil
}
