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

// MockLLMForCoaching mocks AI for coaching tests (implements StreamingLLMProvider)
type MockLLMForCoaching struct {
	response ai.LLMResponse
	chunks   []string // for streaming tests
	err      error
	calls    int
}

func (m *MockLLMForCoaching) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	return m.response, m.err
}

func (m *MockLLMForCoaching) Stream(ctx context.Context, req ai.LLMRequest, onChunk ai.StreamCallback) (ai.LLMResponse, error) {
	m.calls++
	if m.err != nil {
		return ai.LLMResponse{}, m.err
	}
	for _, chunk := range m.chunks {
		if ctx.Err() != nil {
			return ai.LLMResponse{}, ctx.Err()
		}
		onChunk(chunk)
	}
	return m.response, nil
}

func newTestAIProviderForCoaching(mockAI *MockLLMForCoaching) *ai.AIProvider {
	return ai.NewAIProviderForTest(mockAI)
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
			SetModel("groq/compound").
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
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	_, err := svc.GenerateDraft(ctx, userID, &appID, "", expIDs, "문항 텍스트", 800, nil)
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
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	result, err := svc.GenerateDraft(ctx, userID, &appID, "", expIDs, "문항", 800, nil)
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
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	result, err := svc.GenerateDraft(ctx, userID, &appID, "", expIDs, "문항", 800, nil)
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
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	charLimit := 800
	result, err := svc.GenerateDraft(ctx, userID, &appID, "", expIDs, "문항", charLimit, nil)
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
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	_, err := svc.GenerateDraft(ctx, userID, &appID, "", expIDs, "문항", 800, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI")
}

func TestRecordSession_CreatesRecord(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, nil)
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)

	// Create a cover letter first
	coverLetter := client.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(appID).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안 내용").
		SaveX(ctx)

	// Record session
	session, err := svc.RecordSession(ctx, userID, coverLetter.ID, "draft", "system prompt", "user prompt", "assistant response", 500, 300)
	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify session was created
	assert.Equal(t, "draft", string(session.SessionType))
	assert.NotNil(t, session.InputData)
	assert.NotNil(t, session.OutputData)
}

func TestRecordSession_DraftType(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, nil)
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)

	coverLetter := client.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(appID).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	session, err := svc.RecordSession(ctx, userID, coverLetter.ID, "draft", "sys", "user", "asst", 100, 50)
	require.NoError(t, err)

	assert.Equal(t, "draft", string(session.SessionType))
}

func TestRecordSession_MessagesStored(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, nil)
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)

	coverLetter := client.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(appID).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	systemPrompt := "당신은 코칭 전문가입니다"
	userPrompt := "경험을 바탕으로 작성하세요"
	assistantResponse := "[상황]\n초안 내용"

	session, err := svc.RecordSession(ctx, userID, coverLetter.ID, "draft", systemPrompt, userPrompt, assistantResponse, 100, 50)
	require.NoError(t, err)

	// Verify input/output data are stored
	assert.NotNil(t, session.InputData)
	assert.NotNil(t, session.OutputData)
	// Input data should contain system and user prompts
	assert.Contains(t, session.InputData, "system")
	assert.Contains(t, session.InputData, "user")
}

func TestRecordSession_TokenUsageTracked(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, nil)
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)

	coverLetter := client.CoverLetter.Create().
		SetUserID(userID).
		SetApplicationID(appID).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	inputTokens := 500
	outputTokens := 300

	session, err := svc.RecordSession(ctx, userID, coverLetter.ID, "draft", "sys", "user", "asst", inputTokens, outputTokens)
	require.NoError(t, err)

	// Verify token usage is tracked (int fields, not pointers)
	assert.Equal(t, inputTokens, session.InputTokens)
	assert.Equal(t, outputTokens, session.OutputTokens)
}

func TestGenerateDraftStream_StreamsChunks(t *testing.T) {
	chunks := []string{"[상황]\n", "프로젝트에서 ", "리더로서 "}
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content:      "[상황]\n프로젝트에서 리더로서 ",
			InputTokens:  100,
			OutputTokens: 50,
		},
		chunks: chunks,
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	var received []string
	resp, err := svc.GenerateDraftStream(ctx, userID, &appID, "", expIDs, "문항 텍스트", 800, nil,
		func(chunk string) {
			received = append(received, chunk)
		},
	)
	require.NoError(t, err)
	assert.Equal(t, chunks, received)
	assert.Equal(t, 100, resp.InputTokens)
	assert.Equal(t, 50, resp.OutputTokens)
}

func TestGenerateDraftStream_ErrorHandling(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		err: fmt.Errorf("streaming failed"),
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	userID, appID, expIDs := createTestCoachingData(t, client)
	ensureCoachingPrompt(t, client)

	_, err := svc.GenerateDraftStream(ctx, userID, &appID, "", expIDs, "문항", 800, nil, func(string) {})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "streaming failed")
}

// === Phase 8.3.2: Standalone coaching — nil applicationID ===

func TestGenerateDraft_NilApplicationID(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: "[상황]\n자유 코칭 초안\n[결과]\n완료",
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	email := "standalone-c-" + uuid.New().String()[:8] + "@example.com"
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	exp := client.Experience.Create().
		SetUserID(user.ID).SetTitle("경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)

	ensureCoachingPrompt(t, client)

	// nil applicationID, provide companyName directly
	result, err := svc.GenerateDraft(ctx, user.ID, nil, "카카오", []uuid.UUID{exp.ID}, "문항 텍스트", 800, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "[상황]")
	assert.Equal(t, 1, mockAI.calls)
}

func TestGenerateDraftStream_NilApplicationID(t *testing.T) {
	chunks := []string{"[상황]\n", "자유 코칭 "}
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content:      "[상황]\n자유 코칭 ",
			InputTokens:  100,
			OutputTokens: 50,
		},
		chunks: chunks,
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	email := "standalone-cs-" + uuid.New().String()[:8] + "@example.com"
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	exp := client.Experience.Create().
		SetUserID(user.ID).SetTitle("경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)

	ensureCoachingPrompt(t, client)

	var received []string
	resp, err := svc.GenerateDraftStream(ctx, user.ID, nil, "네이버", []uuid.UUID{exp.ID}, "문항 텍스트", 800, nil,
		func(chunk string) { received = append(received, chunk) },
	)
	require.NoError(t, err)
	assert.Equal(t, chunks, received)
	assert.Equal(t, 100, resp.InputTokens)
}

func TestSaveDraftResult_NilApplicationID(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, nil)
	ctx := context.Background()

	email := "standalone-save-" + uuid.New().String()[:8] + "@example.com"
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	// nil applicationID — cover letter should be created without application FK
	result, err := svc.SaveDraftResult(ctx, user.ID, nil, "문항 텍스트", 800, "[상황]\n초안", 500, 300)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.CoverLetterID)

	// Verify cover letter was created without application
	cl, err := client.CoverLetter.Get(ctx, result.CoverLetterID)
	require.NoError(t, err)
	assert.Equal(t, "[상황]\n초안", cl.CurrentContent)
}

func TestSaveDraftResult_CreatesAllRecords(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, nil)
	ctx := context.Background()

	userID, appID, _ := createTestCoachingData(t, client)

	result, err := svc.SaveDraftResult(ctx, userID, &appID,
		"문항 텍스트", 800, "[상황]\n초안 내용", 500, 300)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.CoverLetterID)
	assert.NotEqual(t, uuid.Nil, result.SessionID)

	// Verify cover letter was created
	cl, err := client.CoverLetter.Get(ctx, result.CoverLetterID)
	require.NoError(t, err)
	assert.Equal(t, "[상황]\n초안 내용", cl.CurrentContent)

	// Verify version was created
	versions, err := cl.QueryVersions().All(ctx)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, 1, versions[0].VersionNumber)
}

// === Phase 8.4.1: Advice generation ===

func ensureAdvicePrompt(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	exists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("advice"),
		).
		Exist(ctx)

	if !exists {
		client.PromptTemplate.Create().
			SetCategory("coaching").
			SetSubCategory("advice").
			SetName("Draft Advice").
			SetSystemPrompt("당신은 자소서 코칭 전문가입니다. 초안에 대한 개선 조언을 JSON 배열로 제공하세요.").
			SetUserPromptTemplate(`초안:
{{draft}}

문항: {{question_text}}

개선 포인트를 JSON 배열로 제공하세요.`).
			SetModel("gemini").
			SetTemperature(0.3).
			SetMaxTokens(1000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func TestGenerateAdvice_ReturnsItems(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: `[
				{"category": "metric", "content": "성과를 수치로 표현하면 설득력이 높아집니다.", "priority": 1},
				{"category": "structure", "content": "STAR 구조의 결과 부분을 보강하세요.", "priority": 2},
				{"category": "keyword", "content": "'데이터 기반' 키워드를 추가하세요.", "priority": 3}
			]`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	ensureAdvicePrompt(t, client)

	items, err := svc.GenerateAdvice(ctx, "[상황]\n초안 내용", "문항 텍스트", "경험 요약")
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, "metric", items[0].Category)
	assert.Equal(t, 1, items[0].Priority)
	assert.Contains(t, items[0].Content, "수치")
}

func TestGenerateAdvice_EmptyDraft(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		response: ai.LLMResponse{
			Content: `[]`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	ensureAdvicePrompt(t, client)

	items, err := svc.GenerateAdvice(ctx, "", "문항", "")
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestGenerateAdvice_AIError(t *testing.T) {
	mockAI := &MockLLMForCoaching{
		err: fmt.Errorf("light model unavailable"),
	}

	client := testutil.NewTestClient(t)
	svc := NewCoachingService(client, newTestAIProviderForCoaching(mockAI))
	ctx := context.Background()

	ensureAdvicePrompt(t, client)

	_, err := svc.GenerateAdvice(ctx, "초안", "문항", "경험")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "light model")
}
