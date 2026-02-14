package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/weaponcategory"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMForQuestion mocks AI for question analysis tests
type MockLLMForQuestion struct {
	response ai.LLMResponse
	err      error
	calls    int
}

func (m *MockLLMForQuestion) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	return m.response, m.err
}

func createTestUserAndApplication(t *testing.T, client *ent.Client) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// Use unique email for each test
	email := "question-" + uuid.New().String()[:8] + "@example.com"

	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	app := client.Application.Create().
		SetUserID(user.ID).
		SetCompanyName("테스트회사").
		SetPosition("백엔드 개발자").
		SetJobURL("https://example.com").
		SaveX(ctx)

	return user.ID, app.ID
}

func ensureQuestionAnalysisPrompt(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	exists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("question_analysis"),
		).
		Exist(ctx)

	if !exists {
		client.PromptTemplate.Create().
			SetCategory("coaching").
			SetSubCategory("question_analysis").
			SetName("Question Analysis").
			SetSystemPrompt("당신은 한국 대기업 자소서 전문가입니다.").
			SetUserPromptTemplate(`기업명: {{company_name}}
직무: {{position}}

자소서 문항: "{{question_text}}"
글자수 제한: {{char_limit}}자

위 문항을 분석하여 JSON 형식으로 응답하세요.`).
			SetModel("claude-sonnet-4.5").
			SetTemperature(0.3).
			SetMaxTokens(3000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func ensureWeaponCategoriesForQuestion(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	weapons := []struct {
		code string
		name string
		icon string
	}{
		{"W01", "문제해결", "🔧"},
		{"W02", "협업", "🤝"},
		{"W03", "리더십", "👑"},
	}

	for _, w := range weapons {
		exists, _ := client.WeaponCategory.Query().
			Where(weaponcategory.CodeEQ(w.code)).
			Exist(ctx)

		if !exists {
			client.WeaponCategory.Create().
				SetCode(w.code).
				SetParentCode(w.code).
				SetName(w.name).
				SetKeywords([]string{"test"}).
				SetIcon(w.icon).
				SaveX(ctx)
		}
	}
}

func TestAnalyzeQuestion_ReturnsStructuredResult(t *testing.T) {
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "팀 프로젝트 경험 서술",
				"real_intents": [
					{"intent": "갈등 해결 능력 확인", "why": "팀워크 중시"},
					{"intent": "주도적 문제 인식", "why": "능동성 평가"},
					{"intent": "성과 측정 역량", "why": "정량적 사고"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "핵심 역량"},
					"secondary": [
						{"weapon_id": "W02", "weapon_name": "협업", "reason": "팀 경험"}
					]
				},
				"writing_structure": {
					"total_chars": 800,
					"sections": [
						{"name": "상황", "char_ratio": 0.2, "char_count": 160, "guide": "상황 설정"},
						{"name": "과제", "char_ratio": 0.15, "char_count": 120, "guide": "과제 정의"},
						{"name": "행동", "char_ratio": 0.4, "char_count": 320, "guide": "실행 과정"},
						{"name": "결과", "char_ratio": 0.25, "char_count": 200, "guide": "성과"}
					]
				},
				"key_keywords": ["데이터 기반", "개선율", "주도적", "소통", "성과"],
				"avoid_list": ["열심히", "최선", "노력"],
				"good_structure_example": "상황: 팀 프로젝트에서...\n과제: 해결해야 할 문제는...\n행동: 구체적으로...\n결과: 그 결과..."
			}`,
			InputTokens:  500,
			OutputTokens: 400,
			Model:        "claude-sonnet-4.5",
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, mockAI)
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, appID, "팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.", 800)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify structured result
	assert.NotEmpty(t, result.SurfaceQuestion)
	assert.Len(t, result.RealIntents, 3)
	assert.NotEmpty(t, result.RequiredWeapons.Primary.WeaponID)
	assert.NotEmpty(t, result.WritingStructure.Sections)
	assert.NotEmpty(t, result.KeyKeywords)
	assert.NotEmpty(t, result.AvoidList)
}

func TestAnalyzeQuestion_RealIntentsAlwaysThree(t *testing.T) {
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "경험 서술",
				"real_intents": [
					{"intent": "의도1", "why": "이유1"},
					{"intent": "의도2", "why": "이유2"},
					{"intent": "의도3", "why": "이유3"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "핵심"},
					"secondary": []
				},
				"writing_structure": {
					"total_chars": 800,
					"sections": [
						{"name": "상황", "char_ratio": 1.0, "char_count": 800, "guide": "가이드"}
					]
				},
				"key_keywords": ["키워드"],
				"avoid_list": ["피할말"],
				"good_structure_example": "예시"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, mockAI)
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, appID, "문항", 800)
	require.NoError(t, err)
	assert.Len(t, result.RealIntents, 3, "real_intents should always have exactly 3 items")
}

func TestAnalyzeQuestion_PrimaryWeaponExists(t *testing.T) {
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "경험",
				"real_intents": [
					{"intent": "1", "why": "a"},
					{"intent": "2", "why": "b"},
					{"intent": "3", "why": "c"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "핵심 역량"},
					"secondary": []
				},
				"writing_structure": {
					"total_chars": 800,
					"sections": [{"name": "S", "char_ratio": 1.0, "char_count": 800, "guide": "G"}]
				},
				"key_keywords": ["k"],
				"avoid_list": ["a"],
				"good_structure_example": "ex"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, mockAI)
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, appID, "문항", 800)
	require.NoError(t, err)
	assert.NotEmpty(t, result.RequiredWeapons.Primary.WeaponID, "primary weapon must exist")
	assert.NotEmpty(t, result.RequiredWeapons.Primary.WeaponName)
}

func TestAnalyzeQuestion_CharCountSumsCorrectly(t *testing.T) {
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "질문",
				"real_intents": [
					{"intent": "1", "why": "a"},
					{"intent": "2", "why": "b"},
					{"intent": "3", "why": "c"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "이유"},
					"secondary": []
				},
				"writing_structure": {
					"total_chars": 800,
					"sections": [
						{"name": "상황", "char_ratio": 0.2, "char_count": 160, "guide": "G1"},
						{"name": "과제", "char_ratio": 0.15, "char_count": 120, "guide": "G2"},
						{"name": "행동", "char_ratio": 0.4, "char_count": 320, "guide": "G3"},
						{"name": "결과", "char_ratio": 0.25, "char_count": 200, "guide": "G4"}
					]
				},
				"key_keywords": ["k"],
				"avoid_list": ["a"],
				"good_structure_example": "ex"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, mockAI)
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, appID, "문항", 800)
	require.NoError(t, err)

	// Calculate sum of char_count
	sum := 0
	for _, section := range result.WritingStructure.Sections {
		sum += section.CharCount
	}

	assert.Equal(t, result.WritingStructure.TotalChars, sum, "sum of char_count should equal total_chars")
}

func TestAnalyzeQuestion_CategoryMatching(t *testing.T) {
	// This test ensures weapon category matching works
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "문제 해결 경험",
				"real_intents": [
					{"intent": "1", "why": "a"},
					{"intent": "2", "why": "b"},
					{"intent": "3", "why": "c"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "핵심"},
					"secondary": [
						{"weapon_id": "W02", "weapon_name": "협업", "reason": "보조"}
					]
				},
				"writing_structure": {
					"total_chars": 800,
					"sections": [{"name": "전체", "char_ratio": 1.0, "char_count": 800, "guide": "가이드"}]
				},
				"key_keywords": ["키워드"],
				"avoid_list": ["피할말"],
				"good_structure_example": "예시"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, mockAI)
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, appID, "문항", 800)
	require.NoError(t, err)

	// Verify weapon IDs match existing categories
	assert.Equal(t, "W01", result.RequiredWeapons.Primary.WeaponID)
	assert.Equal(t, "문제해결", result.RequiredWeapons.Primary.WeaponName)
}
