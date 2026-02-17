package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/companyanalysiscache"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
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
	if s.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available for company analysis")
	}

	// Fetch company data (Naver search + News)
	companyData, err := s.companyDataService.GetCompanyData(ctx, companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch company data: %w", err)
	}

	// Generate analysis with AI (model from prompt_templates)
	analysis, err := s.analyzeWithAI(ctx, companyName, companyData)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze with AI: %w", err)
	}

	// Save to cache (365 days TTL)
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		slog.Error("failed to marshal analysis for cache", "error", err)
		analysis.Source = "ai_generated"
		return analysis, nil // Return analysis even if cache save fails
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(analysisJSON, &dataMap); err != nil {
		slog.Error("failed to unmarshal analysis to map", "error", err)
		analysis.Source = "ai_generated"
		return analysis, nil
	}

	if _, err := s.entClient.CompanyAnalysisCache.Create().
		SetCacheKey(cacheKey).
		SetCacheType("company_analysis").
		SetCompanyName(companyName).
		SetData(dataMap).
		SetExpiresAt(time.Now().AddDate(1, 0, 0)). // 365 days
		SetViewCount(1).
		Save(ctx); err != nil {
		slog.Error("failed to save analysis to cache", "company", companyName, "error", err)
	} else {
		slog.Info("analysis_cached", "company", companyName, "cache_key", cacheKey)
	}

	analysis.Source = "ai_generated"
	return analysis, nil
}

// analyzeWithAI uses prompt_templates to analyze company (model configurable via admin)
func (s *CompanyAnalysisService) analyzeWithAI(ctx context.Context, companyName string, data *CompanyData) (*CompanyAnalysis, error) {
	// Load prompt template from DB, fall back to defaults
	pt, err := s.loadAnalysisPrompt(ctx)
	if err != nil {
		return nil, err
	}

	// Build context variables
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

	// Substitute variables in user prompt template
	userPrompt := pt.UserPromptTemplate
	for k, v := range map[string]string{
		"company_name":    companyName,
		"company_context": companyContext,
		"news_text":       newsText,
	} {
		userPrompt = strings.ReplaceAll(userPrompt, "{{"+k+"}}", v)
	}

	startTime := time.Now()
	resp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, pt.Model, ai.LLMRequest{
		SystemPrompt: pt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  pt.Temperature,
		MaxTokens:    pt.MaxTokens,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", pt.Model, err)
	}

	// Update usage stats
	if pt.ID.String() != "00000000-0000-0000-0000-000000000000" {
		latencyMs := int(time.Since(startTime).Milliseconds())
		_ = s.entClient.PromptTemplate.UpdateOneID(pt.ID).
			SetUsageCount(pt.UsageCount + 1).
			SetAvgLatencyMs((pt.AvgLatencyMs*pt.UsageCount + latencyMs) / (pt.UsageCount + 1)).
			Exec(ctx)
	}

	// Parse AI response
	var analysis CompanyAnalysis
	if err := ai.ExtractJSON(resp.Content, &analysis); err != nil {
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

// loadAnalysisPrompt loads the company analysis prompt template from DB.
// Falls back to hardcoded defaults when template is not seeded.
func (s *CompanyAnalysisService) loadAnalysisPrompt(ctx context.Context) (*ent.PromptTemplate, error) {
	if s.entClient != nil {
		pt, err := s.entClient.PromptTemplate.Query().
			Where(
				prompttemplate.CategoryEQ("company_analysis"),
				prompttemplate.SubCategoryEQ("analyze"),
				prompttemplate.IsActiveEQ(true),
			).
			Order(prompttemplate.ByVersion(sql.OrderDesc())).
			First(ctx)
		if err == nil {
			return pt, nil
		}
		slog.Warn("analysis_prompt_fallback", "error", err)
	}
	return &ent.PromptTemplate{
		Model:              "groq/compound",
		SystemPrompt:       "당신은 한국 기업을 분석하는 AI 전문가입니다. 기업의 핵심가치, 인재상, 최근 트렌드를 분석해주세요. JSON으로 응답하세요.",
		UserPromptTemplate: "기업명: {{company_name}}\n\n기업 정보:\n{{company_context}}\n\n최근 뉴스:\n{{news_text}}\n\n위 정보를 기반으로 기업을 분석하세요.",
		Temperature:        0.3,
		MaxTokens:          2000,
	}, nil
}

// generateCacheKey creates a unique cache key for company
func (s *CompanyAnalysisService) generateCacheKey(companyName string) string {
	hash := sha256.Sum256([]byte(companyName))
	return "company_" + hex.EncodeToString(hash[:16])
}
