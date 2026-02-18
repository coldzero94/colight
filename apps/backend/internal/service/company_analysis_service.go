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
	"github.com/coby/colight/apps/backend/ent/companyanalysis"
	"github.com/coby/colight/apps/backend/ent/companyanalysiscache"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/talentprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
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
	CachedAt           *string                  `json:"cached_at,omitempty"`
	ViewCount          *int                     `json:"view_count,omitempty"`
	SourceNews         []NewsArticle            `json:"source_news,omitempty"`
}

type NewsArticle struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description,omitempty"`
	PubDate     string `json:"pub_date,omitempty"`
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

// AnalyzeCompany performs full company analysis with 3-tier lookup.
// cacheSource is used for cache key (job_posting_url or company_name).
// Returns (analysis, fromCache, error) where fromCache indicates if result came from cache.
func (s *CompanyAnalysisService) AnalyzeCompany(ctx context.Context, companyName, cacheSource string) (*CompanyAnalysis, bool, error) {
	// 1. Check talent_profiles table (pre-seeded major companies)
	talentProfile, err := s.entClient.TalentProfile.Query().
		Where(talentprofile.CompanyNameEQ(companyName)).
		First(ctx)

	if err == nil {
		// Found in talent_profiles - return verified data (not counted as cache)
		return s.buildAnalysisFromTalentProfile(talentProfile), false, nil
	}

	// 2. Check company_analysis_cache (365-day TTL, keyed by URL or company name)
	cacheKey := s.generateCacheKey(cacheSource)
	cache, err := s.entClient.CompanyAnalysisCache.Query().
		Where(
			companyanalysiscache.CacheKeyEQ(cacheKey),
			companyanalysiscache.ExpiresAtGT(time.Now()),
		).
		First(ctx)

	if err == nil {
		// Cache hit - increment view_count and return cached data with metadata
		newViewCount := cache.ViewCount + 1
		_ = s.entClient.CompanyAnalysisCache.UpdateOneID(cache.ID).
			SetViewCount(newViewCount).
			Exec(ctx)

		var analysis CompanyAnalysis
		// cache.Data is map[string]interface{}, convert to JSON first
		dataJSON, err := json.Marshal(cache.Data)
		if err == nil {
			if err := json.Unmarshal(dataJSON, &analysis); err == nil {
				analysis.Source = "cache"

				// Add cache metadata
				cachedAt := cache.CreatedAt.Format(time.RFC3339)
				analysis.CachedAt = &cachedAt
				analysis.ViewCount = &newViewCount

				slog.Info("analysis_cache_hit", "company", companyName, "view_count", newViewCount, "cached_at", cachedAt)
				return &analysis, true, nil // fromCache = true
			}
		}
	}

	// 3. Cache miss - generate new analysis with AI
	if s.aiProvider == nil {
		return nil, false, fmt.Errorf("AI provider not available for company analysis")
	}

	// Fetch company data (Naver search + News)
	companyData, err := s.companyDataService.GetCompanyData(ctx, companyName)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch company data: %w", err)
	}

	// Generate analysis with AI (model from prompt_templates)
	analysis, err := s.analyzeWithAI(ctx, companyName, companyData)
	if err != nil {
		return nil, false, fmt.Errorf("failed to analyze with AI: %w", err)
	}

	// Save to cache (365 days TTL)
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		slog.Error("failed to marshal analysis for cache", "error", err)
		analysis.Source = "ai_generated"
		return analysis, false, nil // Return analysis even if cache save fails
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(analysisJSON, &dataMap); err != nil {
		slog.Error("failed to unmarshal analysis to map", "error", err)
		analysis.Source = "ai_generated"
		return analysis, false, nil
	}

	// Try to save to cache (upsert if key exists)
	existing, err := s.entClient.CompanyAnalysisCache.Query().
		Where(companyanalysiscache.CacheKeyEQ(cacheKey)).
		First(ctx)

	if err == nil {
		// Update existing cache
		if _, err := s.entClient.CompanyAnalysisCache.UpdateOneID(existing.ID).
			SetData(dataMap).
			SetExpiresAt(time.Now().AddDate(1, 0, 0)).
			SetViewCount(existing.ViewCount + 1).
			Save(ctx); err != nil {
			slog.Error("failed to update analysis cache", "company", companyName, "error", err)
		} else {
			slog.Info("analysis_cache_updated", "company", companyName, "cache_key", cacheKey)
		}
	} else {
		// Create new cache
		if _, err := s.entClient.CompanyAnalysisCache.Create().
			SetCacheKey(cacheKey).
			SetCacheType("company_analysis").
			SetCompanyName(companyName).
			SetData(dataMap).
			SetExpiresAt(time.Now().AddDate(1, 0, 0)).
			SetViewCount(1).
			Save(ctx); err != nil {
			slog.Error("failed to create analysis cache", "company", companyName, "error", err)
		} else {
			slog.Info("analysis_cached", "company", companyName, "cache_key", cacheKey)
		}
	}

	analysis.Source = "ai_generated"

	// Add metadata for new analysis
	now := time.Now().Format(time.RFC3339)
	viewCount := 1
	analysis.CachedAt = &now
	analysis.ViewCount = &viewCount

	slog.Info("analysis_new", "company", companyName, "cached_at", now)
	return analysis, false, nil // fromCache = false (new AI analysis)
}

// SaveUserAnalysis saves/updates user's analysis record in company_analysis table.
// Upserts: creates new if first time, updates if user already analyzed this company.
func (s *CompanyAnalysisService) SaveUserAnalysis(
	ctx context.Context,
	userID uuid.UUID,
	companyName string,
	jobURL *string,
	analysis *CompanyAnalysis,
) (*ent.CompanyAnalysis, error) {
	// Check if user already analyzed this company
	existing, err := s.entClient.CompanyAnalysis.Query().
		Where(
			companyanalysis.UserIDEQ(userID),
			companyanalysis.CompanyNameEQ(companyName),
		).
		First(ctx)

	// Build analysis map for storage
	analysisMap := map[string]interface{}{
		"company_name":      analysis.CompanyName,
		"core_values":       analysis.CoreValues,
		"talent_traits":     analysis.TalentTraits,
		"recent_trends":     analysis.RecentTrends,
		"strategy_keywords": analysis.StrategyKeywords,
		"avoid_expressions": analysis.AvoidExpressions,
		"source":            analysis.Source,
	}

	// Update existing record
	if err == nil {
		update := s.entClient.CompanyAnalysis.UpdateOneID(existing.ID).
			SetAnalysisResult(analysisMap)

		if jobURL != nil {
			update = update.SetJobURL(*jobURL)
		}

		return update.Save(ctx)
	}

	// Create new record
	builder := s.entClient.CompanyAnalysis.Create().
		SetUserID(userID).
		SetCompanyName(companyName).
		SetAnalysisResult(analysisMap)

	if jobURL != nil {
		builder = builder.SetJobURL(*jobURL)
	}

	return builder.Save(ctx)
}

// GetRecentAnalyses retrieves user's recent analysis history, ordered by created_at DESC.
func (s *CompanyAnalysisService) GetRecentAnalyses(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]*ent.CompanyAnalysis, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return s.entClient.CompanyAnalysis.Query().
		Where(companyanalysis.UserIDEQ(userID)).
		Order(ent.Desc(companyanalysis.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
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
	resp, modelUsed, err := ai.CallWithModelFallback(ctx, s.aiProvider, ai.LLMRequest{
		SystemPrompt: pt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  pt.Temperature,
		MaxTokens:    pt.MaxTokens,
		JSONMode:     true,
	}, ai.DefaultFallbackConfig())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", pt.Model, err)
	}
	slog.Info("ai_model_used", "feature", "company_analysis", "configured", pt.Model, "actual", modelUsed)

	// Update usage stats and auto-switch model on fallback
	if pt.ID.String() != "00000000-0000-0000-0000-000000000000" {
		latencyMs := int(time.Since(startTime).Milliseconds())
		update := s.entClient.PromptTemplate.UpdateOneID(pt.ID).
			SetUsageCount(pt.UsageCount + 1).
			SetAvgLatencyMs((pt.AvgLatencyMs*pt.UsageCount + latencyMs) / (pt.UsageCount + 1))

		// Auto-update model in DB when fallback used a different model
		if modelUsed != "" && modelUsed != pt.Model {
			slog.Warn("ai_model_auto_switch", "feature", "company_analysis", "from", pt.Model, "to", modelUsed)
			update = update.SetModel(modelUsed)
		}
		_ = update.Exec(ctx)
	}

	// Parse AI response (flexible map to preserve all fields)
	resultMap, err := ai.ExtractJSONFlexible(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Normalize fields to prevent null errors
	normalizeCompanyAnalysisFields(resultMap, companyName)

	// Convert to struct for return (preserving extra fields in cache)
	var analysis CompanyAnalysis
	dataJSON, _ := json.Marshal(resultMap)
	if err := json.Unmarshal(dataJSON, &analysis); err != nil {
		return nil, fmt.Errorf("failed to convert to struct: %w", err)
	}

	analysis.CompanyName = companyName

	// Include source news for transparency
	analysis.SourceNews = make([]NewsArticle, 0, len(data.News))
	for _, n := range data.News {
		analysis.SourceNews = append(analysis.SourceNews, NewsArticle{
			Title:       n.Title,
			Link:        n.Link,
			Description: n.Description,
			PubDate:     n.PubDate,
		})
	}

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

// generateCacheKey creates a unique cache key based on URL or company name.
// URL-based caching ensures different job postings get different analyses.
func (s *CompanyAnalysisService) generateCacheKey(source string) string {
	hash := sha256.Sum256([]byte(source))
	return "company_" + hex.EncodeToString(hash[:16])
}

// normalizeCompanyAnalysisFields ensures all expected fields exist with proper defaults.
func normalizeCompanyAnalysisFields(result map[string]interface{}, companyName string) {
	// Required string field
	if _, ok := result["company_name"]; !ok {
		result["company_name"] = companyName
	}

	// Array fields - default to empty array to prevent null errors in frontend
	arrayFields := []string{"core_values", "talent_traits", "recent_trends", "strategy_keywords", "avoid_expressions"}
	for _, field := range arrayFields {
		if val, ok := result[field]; !ok || val == nil {
			result[field] = []interface{}{}
		}
	}

	slog.Info("analysis_normalized", "company", companyName, "field_count", len(result))
}
