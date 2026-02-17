package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHeavyLLM struct {
	response ai.LLMResponse
	err      error
	calls    int
}

func (m *mockHeavyLLM) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	if m.err != nil {
		return ai.LLMResponse{}, m.err
	}
	return m.response, nil
}

func TestAnalyzeCompany_TalentProfileTier(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewTestClient(t)

	client.TalentProfile.Create().
		SetCompanyName("삼성전자").
		SetIndustry("전자").
		SetCoreValues([]map[string]string{{"keyword": "혁신", "description": "기술 혁신"}}).
		SetTalentTraits([]map[string]string{{"trait": "도전정신", "description": "끊임없는 도전"}}).
		SetCultureKeywords([]string{"혁신", "도전", "열정"}).
		SetVerified(true).
		SaveX(ctx)

	svc := NewCompanyAnalysisService(client, nil, nil)

	result, _, err := svc.AnalyzeCompany(ctx, "삼성전자", "삼성전자")
	require.NoError(t, err)
	assert.Equal(t, "삼성전자", result.CompanyName)
	assert.Equal(t, "talent_profiles", result.Source)
	assert.Len(t, result.CoreValues, 1)
	assert.Equal(t, "혁신", result.CoreValues[0].Keyword)
	assert.Len(t, result.TalentTraits, 1)
	assert.Equal(t, "도전정신", result.TalentTraits[0].Trait)
	assert.Equal(t, []string{"혁신", "도전", "열정"}, result.StrategyKeywords)
}

func TestAnalyzeCompany_CacheTier(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewTestClient(t)

	mock := &mockHeavyLLM{
		response: ai.LLMResponse{
			Content: `{
				"core_values": [{"keyword": "성장", "description": "함께 성장"}],
				"talent_traits": [{"trait": "협업", "description": "팀워크"}],
				"recent_trends": [],
				"strategy_keywords": ["성장"],
				"avoid_expressions": []
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mock, mock)
	companyDataSvc := NewCompanyDataService()
	svc := NewCompanyAnalysisService(client, aiProvider, companyDataSvc)

	// First call generates and caches
	result1, _, err := svc.AnalyzeCompany(ctx, "캐시테스트회사", "캐시테스트회사")
	require.NoError(t, err)
	assert.Equal(t, "ai_generated", result1.Source)
	firstCallCount := mock.calls

	// Second call should hit cache
	result2, _, err := svc.AnalyzeCompany(ctx, "캐시테스트회사", "캐시테스트회사")
	require.NoError(t, err)
	assert.Equal(t, "cache", result2.Source)
	assert.Equal(t, firstCallCount, mock.calls) // No additional AI call
}

func TestAnalyzeCompany_AITier(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewTestClient(t)

	mock := &mockHeavyLLM{
		response: ai.LLMResponse{
			Content: `{
				"core_values": [{"keyword": "도전", "description": "도전 정신"}],
				"talent_traits": [{"trait": "창의성", "description": "창의적 사고"}],
				"recent_trends": [{"title": "AI 투자", "summary": "AI 분야 확대"}],
				"strategy_keywords": ["AI", "혁신"],
				"avoid_expressions": ["진부한"]
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mock, mock)
	companyDataSvc := NewCompanyDataService()
	svc := NewCompanyAnalysisService(client, aiProvider, companyDataSvc)

	result, _, err := svc.AnalyzeCompany(ctx, "AI분석회사", "AI분석회사")
	require.NoError(t, err)
	assert.Equal(t, "AI분석회사", result.CompanyName)
	assert.Equal(t, "ai_generated", result.Source)
	assert.Len(t, result.CoreValues, 1)
	assert.Equal(t, "도전", result.CoreValues[0].Keyword)
	assert.Contains(t, result.StrategyKeywords, "AI")
	assert.Contains(t, result.AvoidExpressions, "진부한")
}

func TestAnalyzeCompany_NoAIProvider(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewTestClient(t)

	svc := NewCompanyAnalysisService(client, nil, nil)

	_, _, err := svc.AnalyzeCompany(ctx, "프로바이더없음", "프로바이더없음")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI provider not available")
}

func TestAnalyzeCompany_AIFailure(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewTestClient(t)

	mock := &mockHeavyLLM{
		err: assert.AnError,
	}

	aiProvider := ai.NewAIProviderForTest(mock, mock)
	companyDataSvc := NewCompanyDataService()
	svc := NewCompanyAnalysisService(client, aiProvider, companyDataSvc)

	_, _, err := svc.AnalyzeCompany(ctx, "실패회사", "실패회사")
	assert.Error(t, err)
}

func TestGenerateCacheKey_Deterministic(t *testing.T) {
	client := testutil.NewTestClient(t)
	svc := NewCompanyAnalysisService(client, nil, nil)

	key1 := svc.generateCacheKey("삼성전자")
	key2 := svc.generateCacheKey("삼성전자")
	key3 := svc.generateCacheKey("LG전자")

	assert.Equal(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.Contains(t, key1, "company_")
}

func TestBuildAnalysisFromTalentProfile(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewTestClient(t)
	svc := NewCompanyAnalysisService(client, nil, nil)

	tp := client.TalentProfile.Create().
		SetCompanyName("빌드테스트").
		SetIndustry("IT").
		SetCoreValues([]map[string]string{
			{"keyword": "혁신", "description": "기술 혁신"},
			{"keyword": "도전", "description": "새로운 도전"},
		}).
		SetTalentTraits([]map[string]string{
			{"trait": "리더십", "description": "팀을 이끄는"},
		}).
		SetCultureKeywords([]string{"혁신", "도전"}).
		SetVerified(true).
		SaveX(ctx)

	result := svc.buildAnalysisFromTalentProfile(tp)
	assert.Equal(t, "빌드테스트", result.CompanyName)
	assert.Equal(t, "talent_profiles", result.Source)
	assert.Len(t, result.CoreValues, 2)
	assert.Len(t, result.TalentTraits, 1)
	assert.Equal(t, []string{"혁신", "도전"}, result.StrategyKeywords)
	assert.Empty(t, result.RecentTrends)
	assert.Empty(t, result.AvoidExpressions)
}
