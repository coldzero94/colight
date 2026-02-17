package service

import (
	"context"
	"strings"
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

func newTestAIProviderForQuestion(mockAI *MockLLMForQuestion) *ai.AIProvider {
	return ai.NewAIProviderForTest(mockAI, mockAI)
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
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, &appID, "", "팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.", 800)
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
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, &appID, "", "문항", 800)
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
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, &appID, "", "문항", 800)
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
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, &appID, "", "문항", 800)
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
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	result, err := svc.AnalyzeQuestion(ctx, userID, &appID, "", "문항", 800)
	require.NoError(t, err)

	// Verify weapon IDs match existing categories
	assert.Equal(t, "W01", result.RequiredWeapons.Primary.WeaponID)
	assert.Equal(t, "문제해결", result.RequiredWeapons.Primary.WeaponName)
}

func TestRecommendExperiences_Top3Sorted(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)
	w02, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W02")).Only(ctx)

	// Create 3 experiences with different weapon matches
	exp1 := client.Experience.Create().
		SetUserID(userID).SetTitle("경험1 - 높은 점수").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(exp1).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	exp2 := client.Experience.Create().
		SetUserID(userID).SetTitle("경험2 - 중간 점수").SetContent("내용").SetCategory("인턴").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(exp2).SetWeaponCode(w02.Code).SetConfidence(0.7).SetIsPrimary(false).SaveX(ctx)

	exp3 := client.Experience.Create().
		SetUserID(userID).SetTitle("경험3 - 낮은 점수").SetContent("내용").SetCategory("동아리").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(exp3).SetWeaponCode(w02.Code).SetConfidence(0.5).SetIsPrimary(false).SaveX(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{{WeaponID: w02.Code, WeaponName: "협업"}},
		},
		Limit: 3,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	assert.Len(t, recommendations, 3)

	// Verify sorted by score (descending)
	for i := 0; i < len(recommendations)-1; i++ {
		assert.GreaterOrEqual(t, recommendations[i].MatchScore, recommendations[i+1].MatchScore)
	}
}

func TestRecommendExperiences_PrimaryWeaponHigherWeight(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)
	w02, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W02")).Only(ctx)

	// Experience with primary weapon (0.9 confidence) -> 35 * 0.9 = 31.5 weapon pts
	expPrimary := client.Experience.Create().
		SetUserID(userID).SetTitle("주 무기 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expPrimary).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	// Experience with secondary weapon (1.0 confidence) -> 15 * 1.0 = 15 weapon pts
	expSecondary := client.Experience.Create().
		SetUserID(userID).SetTitle("부 무기 경험").SetContent("내용").SetCategory("인턴").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expSecondary).SetWeaponCode(w02.Code).SetConfidence(1.0).SetIsPrimary(false).SaveX(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{{WeaponID: w02.Code, WeaponName: "협업"}},
		},
		Limit: 3,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)

	// Primary should rank higher
	assert.Equal(t, "주 무기 경험", recommendations[0].Title)
	assert.Greater(t, recommendations[0].MatchScore, recommendations[1].MatchScore)
}

func TestRecommendExperiences_NoExperiences(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{},
		},
		Limit: 3,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	assert.Empty(t, recommendations)
}

func TestRecommendExperiences_KeywordOverlap(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)

	// Experience with matching keywords
	expWithKW := client.Experience.Create().
		SetUserID(userID).SetTitle("키워드 매칭 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SetKeywords([]string{"데이터", "분석", "성과"}).
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expWithKW).SetWeaponCode(w01.Code).SetConfidence(0.8).SetIsPrimary(true).SaveX(ctx)

	// Experience without matching keywords
	expNoKW := client.Experience.Create().
		SetUserID(userID).SetTitle("키워드 없는 경험").SetContent("내용").SetCategory("인턴").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SetKeywords([]string{"봉사", "헌신"}).
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expNoKW).SetWeaponCode(w01.Code).SetConfidence(0.8).SetIsPrimary(true).SaveX(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{},
		},
		KeyKeywords: []string{"데이터", "분석", "성과", "리더십"},
		Limit:       5,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(recommendations), 2)

	// Experience with keyword matches should score higher
	assert.Equal(t, "키워드 매칭 경험", recommendations[0].Title)
	assert.Greater(t, recommendations[0].MatchScore, recommendations[1].MatchScore)
	assert.Equal(t, []string{"데이터", "분석", "성과"}, recommendations[0].KeywordMatches)
}

func TestRecommendExperiences_UsageDedup(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, appID := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)

	// Two experiences with identical weapon scores
	expUsed := client.Experience.Create().
		SetUserID(userID).SetTitle("이미 사용된 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expUsed).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	expFresh := client.Experience.Create().
		SetUserID(userID).SetTitle("새로운 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expFresh).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	// Mark expUsed as already used for this application
	app, _ := client.Application.Get(ctx, appID)
	client.ExperienceUsage.Create().
		SetExperience(expUsed).
		SetUserID(userID).
		SetApplication(app).
		SetQuestionText("이전 문항").
		SaveX(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{},
		},
		ApplicationID: &appID,
		Limit:         5,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(recommendations), 2)

	// Fresh experience should rank higher (no penalty + freshness bonus)
	assert.Equal(t, "새로운 경험", recommendations[0].Title)
	assert.Greater(t, recommendations[0].MatchScore, recommendations[1].MatchScore)

	// Used experience should have is_used flag
	var usedRec *ExperienceRecommendation
	for i := range recommendations {
		if recommendations[i].Title == "이미 사용된 경험" {
			usedRec = &recommendations[i]
			break
		}
	}
	require.NotNil(t, usedRec)
	assert.True(t, usedRec.IsUsed)
}

func TestRecommendExperiences_FreshnessBonus(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)

	// Two experiences with identical weapon scores
	expNeverUsed := client.Experience.Create().
		SetUserID(userID).SetTitle("처음 사용 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expNeverUsed).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	expUsedElsewhere := client.Experience.Create().
		SetUserID(userID).SetTitle("다른곳 사용 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(expUsedElsewhere).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	// Mark expUsedElsewhere as used globally (different application)
	otherApp := client.Application.Create().
		SetUserID(userID).
		SetCompanyName("다른회사").
		SetPosition("개발자").
		SetJobURL("https://example.com/other").
		SaveX(ctx)
	client.ExperienceUsage.Create().
		SetExperience(expUsedElsewhere).
		SetUserID(userID).
		SetApplication(otherApp).
		SetQuestionText("다른 문항").
		SaveX(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{},
		},
		Limit: 5,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(recommendations), 2)

	// Never-used experience should rank higher (freshness bonus)
	assert.Equal(t, "처음 사용 경험", recommendations[0].Title)
	assert.Greater(t, recommendations[0].MatchScore, recommendations[1].MatchScore)
}

func TestRecommendExperiences_MatchReasons(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)
	w02, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W02")).Only(ctx)

	exp := client.Experience.Create().
		SetUserID(userID).SetTitle("종합 경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SetKeywords([]string{"데이터", "분석"}).
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(exp).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(exp).SetWeaponCode(w02.Code).SetConfidence(0.7).SetIsPrimary(false).SaveX(ctx)

	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{{WeaponID: w02.Code, WeaponName: "협업"}},
		},
		KeyKeywords: []string{"데이터", "분석", "성과"},
		Limit:       5,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(recommendations), 1)

	rec := recommendations[0]
	assert.Equal(t, "종합 경험", rec.Title)
	assert.NotEmpty(t, rec.MatchReasons)

	// Should have weapon match reasons
	hasWeaponReason := false
	hasKeywordReason := false
	hasFreshnessReason := false
	for _, reason := range rec.MatchReasons {
		if strings.Contains(reason, "주 무기") {
			hasWeaponReason = true
		}
		if strings.Contains(reason, "키워드") {
			hasKeywordReason = true
		}
		if strings.Contains(reason, "사용되지 않은") {
			hasFreshnessReason = true
		}
	}
	assert.True(t, hasWeaponReason, "should have weapon match reason")
	assert.True(t, hasKeywordReason, "should have keyword match reason")
	assert.True(t, hasFreshnessReason, "should have freshness reason")

	// Should have keyword matches
	assert.Equal(t, []string{"데이터", "분석"}, rec.KeywordMatches)
}

// === Phase 8.3.1: Standalone coaching — nil applicationID ===

func TestAnalyzeQuestion_NilApplicationID_UsesCompanyName(t *testing.T) {
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "자유 코칭 분석",
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
					"sections": [{"name": "전체", "char_ratio": 1.0, "char_count": 800, "guide": "가이드"}]
				},
				"key_keywords": ["키워드"],
				"avoid_list": ["피할말"],
				"good_structure_example": "예시"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	email := "standalone-q-" + uuid.New().String()[:8] + "@example.com"
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	// Call with nil applicationID and a company name
	result, err := svc.AnalyzeQuestion(ctx, user.ID, nil, "카카오", "팀 프로젝트 경험을 기술하세요.", 800)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.RealIntents, 3)
	assert.Equal(t, 1, mockAI.calls)
}

func TestAnalyzeQuestion_NilApplicationID_EmptyCompanyName(t *testing.T) {
	mockAI := &MockLLMForQuestion{
		response: ai.LLMResponse{
			Content: `{
				"surface_question": "범용 코칭",
				"real_intents": [
					{"intent": "1", "why": "a"},
					{"intent": "2", "why": "b"},
					{"intent": "3", "why": "c"}
				],
				"required_weapons": {
					"primary": {"weapon_id": "W01", "weapon_name": "문제해결", "reason": "핵심"},
					"secondary": []
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
	svc := NewQuestionService(client, newTestAIProviderForQuestion(mockAI))
	ctx := context.Background()

	email := "standalone-q2-" + uuid.New().String()[:8] + "@example.com"
	user := client.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(ctx)

	ensureQuestionAnalysisPrompt(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	// Call with nil applicationID and empty company name
	result, err := svc.AnalyzeQuestion(ctx, user.ID, nil, "", "리더십을 발휘한 경험을 기술하세요.", 800)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, mockAI.calls)
}

func TestRecommendExperiences_NilApplicationID(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewQuestionService(client, nil)
	ctx := context.Background()

	userID, _ := createTestUserAndApplication(t, client)
	ensureWeaponCategoriesForQuestion(t, client)

	w01, _ := client.WeaponCategory.Query().Where(weaponcategory.CodeEQ("W01")).Only(ctx)

	exp := client.Experience.Create().
		SetUserID(userID).SetTitle("경험").SetContent("내용").SetCategory("프로젝트").
		SetStarSituation("S").SetStarTask("T").SetStarAction("A").SetStarResult("R").
		SaveX(ctx)
	client.ExperienceWeapon.Create().SetExperience(exp).SetWeaponCode(w01.Code).SetConfidence(0.9).SetIsPrimary(true).SaveX(ctx)

	// ApplicationID is nil pointer — should still work (no usage dedup)
	input := RecommendInput{
		RequiredWeapons: RequiredWeapons{
			Primary:   WeaponInfo{WeaponID: w01.Code, WeaponName: "문제해결"},
			Secondary: []WeaponInfo{},
		},
		ApplicationID: nil,
		Limit:         5,
	}

	recommendations, err := svc.RecommendExperiences(ctx, userID, input)
	require.NoError(t, err)
	require.Len(t, recommendations, 1)
	assert.Equal(t, "경험", recommendations[0].Title)
}
