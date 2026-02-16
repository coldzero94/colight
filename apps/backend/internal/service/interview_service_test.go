package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInterviewService(mockAI *MockLLMProvider) *InterviewService {
	return NewInterviewService(ai.NewAIProviderForTest(mockAI, nil), nil, nil)
}

func TestGenerateQuestion_WarmupStage(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{"question": "안녕하세요! 요즘 어떤 활동을 하고 계신가요?"}`,
		},
	}
	svc := newTestInterviewService(mockAI)

	result, err := svc.GenerateQuestion(context.Background(), uuid.New(), GenerateQuestionInput{
		Stage:    StageWarmup,
		Messages: []ChatMessage{},
	})

	require.NoError(t, err)
	assert.Equal(t, "안녕하세요! 요즘 어떤 활동을 하고 계신가요?", result.Question)
	assert.Equal(t, StageWarmup, result.Stage)
	assert.Equal(t, StageMemory, result.NextStage)
	assert.False(t, result.IsComplete)
}

func TestGenerateQuestion_OutcomeComplete(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{"question": "그 경험에서 가장 크게 배운 점은 무엇인가요?"}`,
		},
	}
	svc := newTestInterviewService(mockAI)

	result, err := svc.GenerateQuestion(context.Background(), uuid.New(), GenerateQuestionInput{
		Stage: StageOutcome,
		Messages: []ChatMessage{
			{Role: "assistant", Content: "이전 질문"},
			{Role: "user", Content: "이전 답변"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, StageOutcome, result.Stage)
	assert.Equal(t, InterviewStage(""), result.NextStage)
	assert.True(t, result.IsComplete)
}

func TestGenerateQuestion_InvalidStage(t *testing.T) {
	mockAI := &MockLLMProvider{}
	svc := newTestInterviewService(mockAI)

	_, err := svc.GenerateQuestion(context.Background(), uuid.New(), GenerateQuestionInput{
		Stage:    InterviewStage("invalid"),
		Messages: []ChatMessage{},
	})

	assert.ErrorIs(t, err, ErrInvalidInterviewStage)
}

func TestGenerateQuestion_AIError(t *testing.T) {
	mockAI := &MockLLMProvider{
		err: assert.AnError,
	}
	svc := newTestInterviewService(mockAI)

	_, err := svc.GenerateQuestion(context.Background(), uuid.New(), GenerateQuestionInput{
		Stage:    StageWarmup,
		Messages: []ChatMessage{},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI call failed")
}

func TestGenerateQuestion_WithConversationHistory(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{"question": "그 프로젝트에서 구체적으로 어떤 역할을 맡으셨나요?"}`,
		},
	}
	svc := newTestInterviewService(mockAI)

	result, err := svc.GenerateQuestion(context.Background(), uuid.New(), GenerateQuestionInput{
		Stage: StageMemory,
		Messages: []ChatMessage{
			{Role: "assistant", Content: "요즘 어떤 활동을 하고 계신가요?"},
			{Role: "user", Content: "최근 대학교 팀 프로젝트를 진행했습니다."},
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.Question)
	assert.Equal(t, StageMemory, result.Stage)
	assert.Equal(t, StageChallenge, result.NextStage)
}

func TestExtractSTAR_Success(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{"title":"팀 프로젝트 리더","category":"project","content":"대학교 팀 프로젝트를 이끌었습니다.","result":"A+ 학점을 받았습니다.","star_situation":"4학년 캡스톤 프로젝트에서 5명의 팀을 이끌었습니다.","star_task":"3개월 내 프로토타입을 완성해야 했습니다.","star_action":"매주 스프린트 회의를 진행하고 역할을 분배했습니다.","star_result":"기한 내에 프로토타입을 완성하여 A+ 학점을 받았습니다.","keywords":["리더십","팀워크","프로젝트관리"]}`,
		},
	}
	svc := newTestInterviewService(mockAI)

	messages := []ChatMessage{
		{Role: "assistant", Content: "어떤 경험이 있으셨나요?"},
		{Role: "user", Content: "대학교에서 팀 프로젝트를 이끌었습니다."},
	}

	result, err := svc.ExtractSTAR(context.Background(), messages)
	require.NoError(t, err)
	assert.Equal(t, "팀 프로젝트 리더", result.Title)
	assert.Equal(t, "project", result.Category)
	assert.NotEmpty(t, result.StarSituation)
	assert.NotEmpty(t, result.StarTask)
	assert.NotEmpty(t, result.StarAction)
	assert.NotEmpty(t, result.StarResult)
	assert.Len(t, result.Keywords, 3)
}

func TestExtractSTAR_EmptyField(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{"title":"test","category":"project","content":"c","result":"r","star_situation":"s","star_task":"","star_action":"a","star_result":"r","keywords":[]}`,
		},
	}
	svc := newTestInterviewService(mockAI)

	_, err := svc.ExtractSTAR(context.Background(), []ChatMessage{
		{Role: "user", Content: "test"},
	})
	assert.ErrorIs(t, err, ErrEmptySTARField)
}

func TestParseStage(t *testing.T) {
	tests := []struct {
		name     string
		messages []ChatMessage
		expected InterviewStage
	}{
		{"empty messages", []ChatMessage{}, StageWarmup},
		{"1 user message", []ChatMessage{
			{Role: "assistant", Content: "q"},
			{Role: "user", Content: "a"},
		}, StageWarmup},
		{"2 user messages", []ChatMessage{
			{Role: "assistant", Content: "q1"},
			{Role: "user", Content: "a1"},
			{Role: "assistant", Content: "q2"},
			{Role: "user", Content: "a2"},
		}, StageMemory},
		{"4 user messages", []ChatMessage{
			{Role: "assistant", Content: "q1"}, {Role: "user", Content: "a1"},
			{Role: "assistant", Content: "q2"}, {Role: "user", Content: "a2"},
			{Role: "assistant", Content: "q3"}, {Role: "user", Content: "a3"},
			{Role: "assistant", Content: "q4"}, {Role: "user", Content: "a4"},
		}, StageChallenge},
		{"6 user messages", []ChatMessage{
			{Role: "assistant", Content: "q1"}, {Role: "user", Content: "a1"},
			{Role: "assistant", Content: "q2"}, {Role: "user", Content: "a2"},
			{Role: "assistant", Content: "q3"}, {Role: "user", Content: "a3"},
			{Role: "assistant", Content: "q4"}, {Role: "user", Content: "a4"},
			{Role: "assistant", Content: "q5"}, {Role: "user", Content: "a5"},
			{Role: "assistant", Content: "q6"}, {Role: "user", Content: "a6"},
		}, StageSolution},
		{"8+ user messages", []ChatMessage{
			{Role: "assistant", Content: "q1"}, {Role: "user", Content: "a1"},
			{Role: "assistant", Content: "q2"}, {Role: "user", Content: "a2"},
			{Role: "assistant", Content: "q3"}, {Role: "user", Content: "a3"},
			{Role: "assistant", Content: "q4"}, {Role: "user", Content: "a4"},
			{Role: "assistant", Content: "q5"}, {Role: "user", Content: "a5"},
			{Role: "assistant", Content: "q6"}, {Role: "user", Content: "a6"},
			{Role: "assistant", Content: "q7"}, {Role: "user", Content: "a7"},
			{Role: "assistant", Content: "q8"}, {Role: "user", Content: "a8"},
		}, StageOutcome},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := ParseStage(tt.messages)
			assert.Equal(t, tt.expected, stage)
		})
	}
}
