package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
)

const (
	maxMarkdownLen = 6000
	maxHTMLLen     = 8000
)

// HTMLFetcher abstracts HTML fetching for testability
type HTMLFetcher interface {
	FetchHTML(url string) (string, error)
}

// defaultHTMLFetcher implements HTMLFetcher using http.Client
type defaultHTMLFetcher struct {
	client *http.Client
}

func (f *defaultHTMLFetcher) FetchHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// CrawlingService handles job posting crawling and parsing
type CrawlingService struct {
	entClient      *ent.Client
	aiProvider     *ai.AIProvider
	jobkoreaParser *crawler.JobKoreaParser
	catchParser    *crawler.CatchParser
	htmlFetcher    HTMLFetcher
}

// NewCrawlingService creates a new CrawlingService with default HTTP fetcher
func NewCrawlingService(entClient *ent.Client, aiProvider *ai.AIProvider) *CrawlingService {
	return &CrawlingService{
		entClient:      entClient,
		aiProvider:     aiProvider,
		jobkoreaParser: crawler.NewJobKoreaParser(),
		catchParser:    crawler.NewCatchParser(),
		htmlFetcher: &defaultHTMLFetcher{
			client: &http.Client{Timeout: 30 * time.Second},
		},
	}
}

// NewCrawlingServiceWithFetcher creates a new CrawlingService with a custom HTMLFetcher (for testing)
func NewCrawlingServiceWithFetcher(entClient *ent.Client, aiProvider *ai.AIProvider, fetcher HTMLFetcher) *CrawlingService {
	return &CrawlingService{
		entClient:      entClient,
		aiProvider:     aiProvider,
		jobkoreaParser: crawler.NewJobKoreaParser(),
		catchParser:    crawler.NewCatchParser(),
		htmlFetcher:    fetcher,
	}
}

// CrawlJobPosting crawls and normalizes a job posting from URL using a 3-tier pipeline:
//   - Tier 1: Known domains (jobkorea/catch) → CSS parser → normalizeWithAI
//   - Tier 2: Universal → readability + markdown → LLM extraction
//   - Tier 3: Fallback → truncated raw HTML → LLM extraction
func (s *CrawlingService) CrawlJobPosting(ctx context.Context, url string) (*crawler.JobPosting, error) {
	// 1. Fetch HTML
	html, err := s.htmlFetcher.FetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}

	// Tier 1: Known domain CSS parsers (fast, no extra LLM cost for parsing)
	if containsDomain(url, "jobkorea.co.kr") {
		rawPosting, parseErr := s.jobkoreaParser.ParseHTML(url, html)
		if parseErr == nil {
			return s.normalizeWithAI(ctx, rawPosting)
		}
		// If CSS parse fails, fall through to Tier 2
	} else if containsDomain(url, "catch.co.kr") {
		rawPosting, parseErr := s.catchParser.ParseHTML(url, html)
		if parseErr == nil {
			return s.normalizeWithAI(ctx, rawPosting)
		}
		// If CSS parse fails, fall through to Tier 2
	}

	// Tier 2: Readability + Markdown + LLM
	markdown, _ := crawler.ScrapeToMarkdown(html, url)
	if len(markdown) >= crawler.MinMarkdownLength {
		return s.extractJobPostingFromMarkdown(ctx, url, markdown)
	}

	// Tier 3: Raw HTML + LLM (last resort)
	return s.extractJobPostingFromHTML(ctx, url, html)
}

// extractJobPostingFromMarkdown extracts structured job posting data from clean markdown content
func (s *CrawlingService) extractJobPostingFromMarkdown(ctx context.Context, sourceURL string, markdown string) (*crawler.JobPosting, error) {
	pt, err := s.loadCrawlingPrompt(ctx, "extract_markdown")
	if err != nil {
		return nil, err
	}

	content := truncateString(markdown, maxMarkdownLen)
	userPrompt := pt.UserPromptTemplate
	for k, v := range map[string]string{
		"source_url": sourceURL,
		"content":    content,
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
		return nil, fmt.Errorf("AI extraction from markdown failed: %w", err)
	}

	s.updatePromptStats(ctx, pt, time.Since(startTime))

	var result crawler.JobPosting
	if err := ai.ExtractJSON(resp.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// extractJobPostingFromHTML extracts structured job posting data from raw HTML (fallback)
func (s *CrawlingService) extractJobPostingFromHTML(ctx context.Context, sourceURL string, html string) (*crawler.JobPosting, error) {
	pt, err := s.loadCrawlingPrompt(ctx, "extract_html")
	if err != nil {
		return nil, err
	}

	content := truncateString(html, maxHTMLLen)
	userPrompt := pt.UserPromptTemplate
	for k, v := range map[string]string{
		"source_url": sourceURL,
		"content":    content,
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
		return nil, fmt.Errorf("AI extraction from HTML failed: %w", err)
	}

	s.updatePromptStats(ctx, pt, time.Since(startTime))

	var result crawler.JobPosting
	if err := ai.ExtractJSON(resp.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// normalizeWithAI uses LLM to normalize RawJobPosting into structured JobPosting
func (s *CrawlingService) normalizeWithAI(ctx context.Context, raw *crawler.RawJobPosting) (*crawler.JobPosting, error) {
	pt, err := s.loadCrawlingPrompt(ctx, "normalize")
	if err != nil {
		return nil, err
	}

	userPrompt := pt.UserPromptTemplate
	for k, v := range map[string]string{
		"company_name": raw.CompanyName,
		"position":     raw.Position,
		"department":   raw.Department,
		"career":       raw.Career,
		"location":     raw.Location,
		"main_tasks":   raw.MainTasks,
		"requirements": raw.Requirements,
		"preferred":    raw.Preferred,
		"skills":       fmt.Sprintf("%v", raw.Skills),
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
		return nil, err
	}

	s.updatePromptStats(ctx, pt, time.Since(startTime))

	var result crawler.JobPosting
	if err := ai.ExtractJSON(resp.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// loadCrawlingPrompt loads a prompt template for crawling by sub_category.
// Falls back to hardcoded defaults when DB is unavailable (e.g. tests).
func (s *CrawlingService) loadCrawlingPrompt(ctx context.Context, subCategory string) (*ent.PromptTemplate, error) {
	if s.entClient != nil {
		pt, err := s.entClient.PromptTemplate.Query().
			Where(
				prompttemplate.CategoryEQ("crawling"),
				prompttemplate.SubCategoryEQ(subCategory),
				prompttemplate.IsActiveEQ(true),
			).
			Order(prompttemplate.ByVersion(sql.OrderDesc())).
			First(ctx)
		if err == nil {
			return pt, nil
		}
	}
	// Fallback defaults for tests or missing seed data
	return crawlingDefaultPrompt(subCategory), nil
}

// crawlingDefaultPrompt returns hardcoded defaults for crawling prompts
func crawlingDefaultPrompt(subCategory string) *ent.PromptTemplate {
	defaults := map[string]*ent.PromptTemplate{
		"extract_markdown": {
			Model:       "gemini-2.0-flash",
			SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from the provided content. Always respond in valid JSON.",
			UserPromptTemplate: "URL: {{source_url}}\n\n{{content}}\n\nExtract job posting fields as JSON.",
			Temperature: 0.1,
			MaxTokens:   2000,
		},
		"extract_html": {
			Model:       "gemini-2.0-flash",
			SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from raw HTML. Always respond in valid JSON.",
			UserPromptTemplate: "URL: {{source_url}}\n\n{{content}}\n\nExtract job posting fields as JSON.",
			Temperature: 0.1,
			MaxTokens:   2000,
		},
		"normalize": {
			Model:       "gemini-2.0-flash",
			SystemPrompt: "You are a job posting data extractor. Extract structured information from raw text.",
			UserPromptTemplate: "Company: {{company_name}}\nPosition: {{position}}\nDepartment: {{department}}\nCareer: {{career}}\nLocation: {{location}}\nMain Tasks: {{main_tasks}}\nRequirements: {{requirements}}\nPreferred: {{preferred}}\nSkills: {{skills}}\n\nReturn structured JSON.",
			Temperature: 0.2,
			MaxTokens:   2000,
		},
	}
	if pt, ok := defaults[subCategory]; ok {
		return pt
	}
	return &ent.PromptTemplate{Model: "gemini-2.0-flash", Temperature: 0.1, MaxTokens: 2000}
}

// updatePromptStats updates usage count and avg latency for a prompt template
func (s *CrawlingService) updatePromptStats(ctx context.Context, pt *ent.PromptTemplate, latency time.Duration) {
	if s.entClient == nil || pt.ID.String() == "00000000-0000-0000-0000-000000000000" {
		return
	}
	latencyMs := int(latency.Milliseconds())
	_ = s.entClient.PromptTemplate.UpdateOneID(pt.ID).
		SetUsageCount(pt.UsageCount + 1).
		SetAvgLatencyMs((pt.AvgLatencyMs*pt.UsageCount + latencyMs) / (pt.UsageCount + 1)).
		Exec(ctx)
}

func containsDomain(url, domain string) bool {
	return len(url) > 0 && len(domain) > 0 && (url[0:1] != "" && domain[0:1] != "") &&
		(url == domain || strings.Contains(url, domain))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
