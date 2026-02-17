package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/companyanalysiscache"
	"github.com/coby/colight/apps/backend/ent/talentprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
)

// CompanyAnalysisService analyzes companies using AI
type CompanyAnalysisService struct {
	entClient          *ent.Client
	aiProvider         *ai.AIProvider
	companyDataService *CompanyDataService
}

// NewCompanyAnalysisService creates a new company analysis service
func NewCompanyAnalysisService(entClient *ent.Client, aiProvider *ai.AIProvider, companyDataService *CompanyDataService) *CompanyAnalysisService {
	return &CompanyAnalysisService{
		entClient:          entClient,
		aiProvider:         aiProvider,
		companyDataService: companyDataService,
	}
}

// CompanyAnalysis represents analyzed company data
type CompanyAnalysis struct {
	CompanyName        string                   `json:"company_name"`
	CoreValues         []CoreValue              `json:"core_values"`
	TalentTraits       []TalentTrait            `json:"talent_traits"`
	RecentTrends       []Trend                  `json:"recent_trends"`
	StrategyKeywords   []string                 `json:"strategy_keywords"`
	AvoidExpressions   []string                 `json:"avoid_expressions"`
	Source             string                   `json:"source"` // "talent_profiles", "cache", "ai_generated"
}

type CoreValue struct {
	Keyword     string `json:"keyword"`
	Description string `json:"description"`
}

type TalentTrait struct {
	Trait       string `json:"trait"`
	Description string `json:"description"`
	Evidence    string `json:"evidence,omitempty"`
}

type Trend struct {
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Relevance string `json:"relevance,omitempty"`
}

// AnalyzeCompany performs full company analysis with 3-tier lookup
func (s *CompanyAnalysisService) AnalyzeCompany(ctx context.Context, companyName string) (*CompanyAnalysis, error) {
	// 1. Check talent_profiles table (pre-seeded major companies)
	talentProfile, err := s.entClient.TalentProfile.Query().
		Where(talentprofile.CompanyNameEQ(companyName)).
		First(ctx)

	if err == nil {
		// Found in talent_profiles - return verified data
		return s.buildAnalysisFromTalentProfile(talentProfile), nil
	}

	// 2. Check company_analysis_cache (365-day TTL)
	cacheKey := s.generateCacheKey(companyName)
	cache, err := s.entClient.CompanyAnalysisCache.Query().
		Where(
			companyanalysiscache.CacheKeyEQ(cacheKey),
			companyanalysiscache.ExpiresAtGT(time.Now()),
		).
		First(ctx)

	if err == nil {
		// Cache hit - increment view_count and return cached data
		_ = s.entClient.CompanyAnalysisCache.UpdateOneID(cache.ID).
			SetViewCount(cache.ViewCount + 1).
			Exec(ctx)

		var analysis CompanyAnalysis
		// cache.Data is map[string]interface{}, convert to JSON first
		dataJSON, err := json.Marshal(cache.Data)
		if err == nil {
			if err := json.Unmarshal(dataJSON, &analysis); err == nil {
				analysis.Source = "cache"
				return &analysis, nil
			}
		}
	}

	// 3. Cache miss - generate new analysis with AI
	if s.aiProvider == nil || s.aiProvider.Claude() == nil {
		return nil, fmt.Errorf("AI provider not available for company analysis")
	}

	// Fetch company data (Naver search + News)
	companyData, err := s.companyDataService.GetCompanyData(ctx, companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch company data: %w", err)
	}

	// Generate analysis with Claude
	analysis, err := s.analyzeWithClaude(ctx, companyName, companyData)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze with AI: %w", err)
	}

	// Save to cache (365 days TTL)
	analysisJSON, _ := json.Marshal(analysis)
	var dataMap map[string]interface{}
	_ = json.Unmarshal(analysisJSON, &dataMap)

	_, _ = s.entClient.CompanyAnalysisCache.Create().
		SetCacheKey(cacheKey).
		SetCacheType("company_analysis").
		SetCompanyName(companyName).
		SetData(dataMap).
		SetExpiresAt(time.Now().AddDate(1, 0, 0)). // 365 days
		SetViewCount(1).
		Save(ctx)

	analysis.Source = "ai_generated"
	return analysis, nil
}

// analyzeWithClaude uses Claude to analyze company
func (s *CompanyAnalysisService) analyzeWithClaude(ctx context.Context, companyName string, data *CompanyData) (*CompanyAnalysis, error) {
	// Build prompt from company data
	newsText := ""
	for i, article := range data.News {
		if i >= 5 {
			break
		}
		newsText += fmt.Sprintf("- %s\n", article.Title)
	}

	companyContext := data.CompanyContext
	if companyContext == "" {
		companyContext = "(기업 정보를 찾을 수 없습니다. 기업명과 뉴스만으로 분석해주세요.)"
	}

	prompt := fmt.Sprintf(`다음 정보를 바탕으로 이 한국 기업을 분석해주세요.

기업명: %s

--- 기업 정보 ---
%s

--- 최근 뉴스 ---
%s

다음 항목을 추출하여 JSON으로 반환해주세요:
1. 핵심 가치 (core_values): 3-5개
2. 인재상 (talent_traits): 3-5개 — 기업이 원하는 인재 특성
3. 최근 동향 (recent_trends): 뉴스 기반
4. 자기소개서 전략 키워드 (strategy_keywords): 이 기업에 지원할 때 효과적인 키워드
5. 피해야 할 표현 (avoid_expressions): 이 기업에 맞지 않는 표현

{
  "core_values": [{"keyword": "", "description": ""}],
  "talent_traits": [{"trait": "", "description": "", "evidence": ""}],
  "recent_trends": [{"title": "", "summary": "", "relevance": ""}],
  "strategy_keywords": [],
  "avoid_expressions": []
}`, companyName, companyContext, newsText)

	resp, err := s.aiProvider.CallByModelName(ctx, "claude-sonnet-4-5", ai.LLMRequest{
		SystemPrompt: "당신은 한국 기업 분석 전문가입니다. 기업 정보와 뉴스를 분석하여 취업 준비생에게 유용한 인사이트를 제공합니다. 반드시 한국어로 응답해주세요.",
		UserPrompt:   prompt,
		Temperature:  0.3,
		MaxTokens:    3000,
		JSONMode:     true,
	})

	if err != nil {
		return nil, err
	}

	// Parse AI response
	var analysis CompanyAnalysis
	if err := json.Unmarshal([]byte(resp.Content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	analysis.CompanyName = companyName
	return &analysis, nil
}

// buildAnalysisFromTalentProfile converts talent_profiles record to CompanyAnalysis
func (s *CompanyAnalysisService) buildAnalysisFromTalentProfile(tp *ent.TalentProfile) *CompanyAnalysis {
	analysis := &CompanyAnalysis{
		CompanyName:      tp.CompanyName,
		CoreValues:       []CoreValue{},
		TalentTraits:     []TalentTrait{},
		RecentTrends:     []Trend{},
		StrategyKeywords: tp.CultureKeywords,
		AvoidExpressions: []string{},
		Source:           "talent_profiles",
	}

	// Convert core_values from DB format ([]map[string]string)
	for _, cv := range tp.CoreValues {
		analysis.CoreValues = append(analysis.CoreValues, CoreValue{
			Keyword:     cv["keyword"],
			Description: cv["description"],
		})
	}

	// Convert talent_traits from DB format ([]map[string]string)
	for _, tt := range tp.TalentTraits {
		analysis.TalentTraits = append(analysis.TalentTraits, TalentTrait{
			Trait:       tt["trait"],
			Description: tt["description"],
		})
	}

	return analysis
}

// generateCacheKey creates a unique cache key for company
func (s *CompanyAnalysisService) generateCacheKey(companyName string) string {
	hash := sha256.Sum256([]byte(companyName))
	return "company_" + hex.EncodeToString(hash[:16])
}
