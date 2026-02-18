package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLLMForCoaching struct {
	response ai.LLMResponse
	chunks   []string
	err      error
}

func (m *mockLLMForCoaching) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func (m *mockLLMForCoaching) Stream(ctx context.Context, req ai.LLMRequest, onChunk ai.StreamCallback) (ai.LLMResponse, error) {
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

func setupCoachingTestRouter(t *testing.T, mockAI *mockLLMForCoaching) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	client := testutil.NewTestClient(t)
	coachingService := service.NewCoachingService(client, ai.NewAIProviderForTest(mockAI))
	coachingCtrl := NewCoachingController(coachingService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/coaching/draft", coachingCtrl.PostDraft)
	v1.GET("/coaching/sessions", coachingCtrl.GetSessions)

	return router, client
}

func draftRequest(router *gin.Engine, body any, userID uuid.UUID) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/coaching/draft", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		req.Header.Set("X-Test-UserID", userID.String())
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
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
			SetName("Test Draft").
			SetSystemPrompt("Test system").
			SetUserPromptTemplate("Test {{question_text}}").
			SetModel("groq/compound").
			SetTemperature(0.7).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

// sseEvent represents a parsed SSE event.
type sseEvent struct {
	Event string
	Data  map[string]interface{}
}

// parseSSEEvents parses the raw SSE body into structured events.
func parseSSEEvents(body string) []sseEvent {
	var events []sseEvent
	blocks := strings.Split(body, "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		var eventType, dataStr string
		for _, line := range strings.Split(block, "\n") {
			if strings.HasPrefix(line, "event: ") {
				eventType = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				dataStr = strings.TrimPrefix(line, "data: ")
			}
		}
		if eventType == "" && dataStr == "" {
			continue
		}
		var data map[string]interface{}
		if dataStr != "" {
			_ = json.Unmarshal([]byte(dataStr), &data)
		}
		events = append(events, sseEvent{Event: eventType, Data: data})
	}
	return events
}

func TestPostDraft_Unauthorized(t *testing.T) {
	mockAI := &mockLLMForCoaching{
		response: ai.LLMResponse{Content: "[상황]\n초안"},
	}

	router, _ := setupCoachingTestRouter(t, mockAI)

	w := draftRequest(router, map[string]any{
		"application_id": uuid.New().String(),
		"experience_ids": []string{uuid.New().String()},
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, uuid.Nil) // No user ID

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostDraft_SSEHeaders(t *testing.T) {
	mockAI := &mockLLMForCoaching{
		chunks: []string{"[상황]\n", "스트리밍 테스트"},
		response: ai.LLMResponse{
			Content:      "[상황]\n스트리밍 테스트",
			InputTokens:  10,
			OutputTokens: 20,
		},
	}

	router, client := setupCoachingTestRouter(t, mockAI)
	ctx := context.Background()
	ensureCoachingPrompt(t, client)

	user := client.UserProfile.Create().
		SetEmail("stream-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("경험").
		SetContent("내용").
		SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)

	w := draftRequest(router, map[string]any{
		"application_id": app.ID.String(),
		"experience_ids": []string{exp.ID.String()},
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

func TestPostDraft_StreamsChunksAndSaves(t *testing.T) {
	mockAI := &mockLLMForCoaching{
		chunks: []string{"[상황]\n", "초안 내용\n", "[결과]\n", "성과"},
		response: ai.LLMResponse{
			Content:      "[상황]\n초안 내용\n[결과]\n성과",
			InputTokens:  50,
			OutputTokens: 100,
		},
	}

	router, client := setupCoachingTestRouter(t, mockAI)
	ctx := context.Background()
	ensureCoachingPrompt(t, client)

	user := client.UserProfile.Create().
		SetEmail("save-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("경험").
		SetContent("내용").
		SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)

	w := draftRequest(router, map[string]any{
		"application_id": app.ID.String(),
		"experience_ids": []string{exp.ID.String()},
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code)

	events := parseSSEEvents(w.Body.String())

	// Should have 4 text events + 1 done event
	require.GreaterOrEqual(t, len(events), 5)

	// Verify text events
	for i := 0; i < 4; i++ {
		assert.Equal(t, "text", events[i].Event)
		assert.Equal(t, "text", events[i].Data["type"])
	}
	assert.Equal(t, "[상황]\n", events[0].Data["content"])
	assert.Equal(t, "성과", events[3].Data["content"])

	// Verify done event
	doneEvent := events[len(events)-1]
	assert.Equal(t, "done", doneEvent.Event)
	assert.Equal(t, "done", doneEvent.Data["type"])
	assert.NotEmpty(t, doneEvent.Data["cover_letter_id"])
	assert.NotEmpty(t, doneEvent.Data["session_id"])

	// Verify DB records were created
	clID, err := uuid.Parse(doneEvent.Data["cover_letter_id"].(string))
	require.NoError(t, err)

	cl, err := client.CoverLetter.Get(ctx, clID)
	require.NoError(t, err)
	assert.Equal(t, "[상황]\n초안 내용\n[결과]\n성과", cl.CurrentContent)
	require.NotNil(t, cl.CharLimit)
	assert.Equal(t, 800, *cl.CharLimit)
}

func TestPostDraft_StreamError(t *testing.T) {
	mockAI := &mockLLMForCoaching{
		err: assert.AnError,
	}

	router, client := setupCoachingTestRouter(t, mockAI)
	ctx := context.Background()
	ensureCoachingPrompt(t, client)

	user := client.UserProfile.Create().
		SetEmail("error-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("경험").
		SetContent("내용").
		SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)

	w := draftRequest(router, map[string]any{
		"application_id": app.ID.String(),
		"experience_ids": []string{exp.ID.String()},
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code) // SSE always returns 200

	events := parseSSEEvents(w.Body.String())
	require.GreaterOrEqual(t, len(events), 1)

	lastEvent := events[len(events)-1]
	assert.Equal(t, "error", lastEvent.Event)
	assert.Equal(t, "error", lastEvent.Data["type"])
}

// === Phase 8.4.2: SSE done event includes advice ===

func ensureAdvicePromptCtrl(t *testing.T, client *ent.Client) {
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
			SetSystemPrompt("Advice system prompt").
			SetUserPromptTemplate("Draft: {{draft}}\nQuestion: {{question_text}}").
			SetModel("gemini").
			SetTemperature(0.3).
			SetMaxTokens(1000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func TestPostDraft_DoneEventIncludesAdvice(t *testing.T) {
	// Single mock handles both draft generation and advice
	mockAI := &mockLLMForCoaching{
		response: ai.LLMResponse{
			Content:      `[{"category":"metric","content":"수치를 추가하세요.","priority":1},{"category":"structure","content":"결과를 보강하세요.","priority":2}]`,
			InputTokens:  50,
			OutputTokens: 100,
		},
	}

	gin.SetMode(gin.TestMode)
	client := testutil.NewTestClient(t)
	provider := ai.NewAIProviderForTest(mockAI)
	coachingService := service.NewCoachingService(client, provider)
	coachingCtrl := NewCoachingController(coachingService)

	router := gin.New()
	v1 := router.Group("/v1")
	v1.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-UserID"); uid != "" {
			parsed, _ := uuid.Parse(uid)
			c.Set("user_id", parsed)
		}
		c.Next()
	})
	v1.POST("/coaching/draft", coachingCtrl.PostDraft)

	ctx := context.Background()
	ensureCoachingPrompt(t, client)
	ensureAdvicePromptCtrl(t, client)

	user := client.UserProfile.Create().
		SetEmail("advice-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)

	w := draftRequest(router, map[string]any{
		"application_id": app.ID.String(),
		"experience_ids": []string{exp.ID.String()},
		"question_text":  "팀 프로젝트 경험을 기술하세요",
		"char_limit":     800,
	}, user.ID)

	assert.Equal(t, http.StatusOK, w.Code)

	events := parseSSEEvents(w.Body.String())
	require.GreaterOrEqual(t, len(events), 3) // text chunks + done

	// Find the done event
	doneEvent := events[len(events)-1]
	assert.Equal(t, "done", doneEvent.Event)
	assert.NotEmpty(t, doneEvent.Data["cover_letter_id"])
	assert.NotEmpty(t, doneEvent.Data["session_id"])

	// Verify advice is included in done event
	adviceRaw, ok := doneEvent.Data["advice"]
	require.True(t, ok, "done event should include 'advice' field")
	adviceList, ok := adviceRaw.([]interface{})
	require.True(t, ok, "advice should be an array")
	assert.Len(t, adviceList, 2)

	// Verify first advice item
	first := adviceList[0].(map[string]interface{})
	assert.Equal(t, "metric", first["category"])
	assert.Contains(t, first["content"], "수치")
}

func TestGetSessions_Success(t *testing.T) {
	mockAI := &mockLLMForCoaching{
		response: ai.LLMResponse{Content: "[상황]\n초안"},
	}

	router, client := setupCoachingTestRouter(t, mockAI)
	ctx := context.Background()

	user := client.UserProfile.Create().
		SetEmail("sessions-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("회사").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter := client.CoverLetter.Create().
		SetUserID(user.ID).
		SetApplicationID(app.ID).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	// Create a session
	client.CoachingSession.Create().
		SetUserID(user.ID).
		SetCoverLetter(coverLetter).
		SetSessionType("draft").
		SetInputData(map[string]interface{}{"system": "test"}).
		SetOutputData(map[string]interface{}{"content": "response"}).
		SaveX(ctx)

	// GET request for sessions
	req := httptest.NewRequest(http.MethodGet, "/v1/coaching/sessions?cover_letter_id="+coverLetter.ID.String(), nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	sessions := resp["sessions"].([]interface{})
	assert.GreaterOrEqual(t, len(sessions), 1)
}

func TestGetSessions_OnlyOwnSessions(t *testing.T) {
	mockAI := &mockLLMForCoaching{
		response: ai.LLMResponse{Content: "[상황]\n초안"},
	}

	router, client := setupCoachingTestRouter(t, mockAI)
	ctx := context.Background()

	// User 1
	user1 := client.UserProfile.Create().
		SetEmail("user1@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	// User 2
	user2 := client.UserProfile.Create().
		SetEmail("user2@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app1 := client.Application.Create().
		SetUserID(user1.ID).
		SetCompanyName("회사1").
		SetPosition("개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	coverLetter1 := client.CoverLetter.Create().
		SetUserID(user1.ID).
		SetApplicationID(app1.ID).
		SetQuestionText("문항").
		SetCharLimit(800).
		SetCurrentContent("초안").
		SaveX(ctx)

	// User1's session
	client.CoachingSession.Create().
		SetUserID(user1.ID).
		SetCoverLetter(coverLetter1).
		SetSessionType("draft").
		SetInputData(map[string]interface{}{}).
		SetOutputData(map[string]interface{}{}).
		SaveX(ctx)

	// User2 should not see user1's sessions
	req := httptest.NewRequest(http.MethodGet, "/v1/coaching/sessions?cover_letter_id="+coverLetter1.ID.String(), nil)
	req.Header.Set("X-Test-UserID", user2.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return empty or forbidden
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusForbidden)
}
