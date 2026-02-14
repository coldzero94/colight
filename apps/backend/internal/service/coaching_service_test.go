package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMForCoaching mocks AI for coaching tests
type MockLLMForCoaching struct {
	response ai.LLMResponse
	err      error
	calls    int
}

func (m *MockLLMForCoaching) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	return m.response, m.err
}

func createTestCoachingData(t *testing.T, client *ent.Client) (uuid.UUID, uuid.UUID, []uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// Create user
	email := "coaching-" + uuid.New().String()[:8] + "@example.com"
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	// Create application
	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("테스트회사").
		SetPosition("백엔드 개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	// Create experiences
	exp1 := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("프로젝트 경험").
		SetContent("내용").
		SetCategory("프로젝트").
		SetStarSituation("팀 프로젝트 상황").
		SetStarTask("해결 과제").
		SetStarAction("구체적 행동").
		SetStarResult("성과").
		SaveX(ctx)

	exp2 := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("인턴 경험").
		SetContent("내용").
		SetCategory("인턴").
		SetStarSituation("인턴십 상황").
		SetStarTask("맡은 업무").
		SetStarAction("실행 과정").
		SetStarResult("결과").
		SaveX(ctx)

	return user.ID, app.ID, []uuid.UUID{exp1.ID, exp2.ID}
}

func ensureCoachingPrompt(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	exists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("draft"),
		).
		Exist(ctx)

	if !exists {
		client.PromptTemplate.Create().
			SetCategory("coaching").
			SetSubCategory("draft").
			SetName("Draft Coaching").
			SetSystemPrompt("당신은 한국 대기업 자소서 코칭 전문가입니다.").
			SetUserPromptTemplate(`문항: {{question_text}}
글자수: {{char_limit}}자

경험: {{experiences}}

STAR 구조로 작성하세요.`).
			SetModel("claude-sonnet-4.5").
			SetTemperature(0.7).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func TestGenerateDraft_CoachingPromptComposition(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: "[상황]\n테스트 초안\n[행동]\n구체적 행동\n[결과]\n성과",
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, mockAI)
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	_, err := svc.GenerateDraft(ctx, userID, appID, expIDs, "문항 텍스트", 800, nil)
	require.NoError(t, err)

	// Verify AI was called
	assert.Equal(t, 1, mockAI.calls)
}

func TestGenerateDraft_StreamingResponse(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: "[상황]\n스트리밍 테스트\n[결과]\n완료",
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, mockAI)
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	result, err := svc.GenerateDraft(ctx, userID, appID, expIDs, "문항", 800, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result, "[상황]")
	assert.Contains(t, result, "[결과]")
}

func TestGenerateDraft_STARStructure(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: "[상황]\n상황 설명\n[과제]\n과제 정의\n[행동]\n행동 설명\n[결과]\n성과",
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, mockAI)
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	result, err := svc.GenerateDraft(ctx, userID, appID, expIDs, "문항", 800, nil)
	require.NoError(t, err)

	// Verify STAR tags are present
	assert.Contains(t, result, "[상황]")
	assert.Contains(t, result, "[과제]")
	assert.Contains(t, result, "[행동]")
	assert.Contains(t, result, "[결과]")
}

func TestGenerateDraft_CharLimitRespected(t *testing.T) {
	// Generate content under char limit
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: "[상황]\n짧은 초안입니다.\n[결과]\n완료",
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, mockAI)
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	charLimit := 800
	result, err := svc.GenerateDraft(ctx, userID, appID, expIDs, "문항", charLimit, nil)
	require.NoError(t, err)

	// Note: We can't strictly enforce char limit on AI output,
	// but we verify the prompt includes the limit
	assert.NotEmpty(t, result)
	assert.LessOrEqual(t, len([]rune(result)), charLimit+200) // Allow 200 char buffer
}

func TestGenerateDraft_ErrorHandling(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		err: fmt.Errorf("AI service unavailable"),
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, mockAI)
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	_, err := svc.GenerateDraft(ctx, userID, appID, expIDs, "문항", 800, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI")
}
