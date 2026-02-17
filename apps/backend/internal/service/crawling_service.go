package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
)

const (
	maxMarkdownLen = 12000 // increased for rich content preservation
	maxHTMLLen     = 8000
)

// HTMLFetcher abstracts HTML fetching for testability
type HTMLFetcher interface {
	FetchHTML(url string) (string, error)
}

// HeadlessRenderer abstracts headless browser rendering for testability
type HeadlessRenderer interface {
	FetchRenderedHTML(ctx context.Context, targetURL string) (string, error)
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
	entClient         *ent.Client
	aiProvider        *ai.AIProvider
	jobkoreaParser    *crawler.JobKoreaParser
	catchParser       *crawler.CatchParser
	saraminParser     *crawler.SaraminParser
	wantedExtractor   *crawler.WantedExtractor
	htmlFetcher       HTMLFetcher
	headlessRenderer  HeadlessRenderer
}

// NewCrawlingService creates a new CrawlingService with default HTTP fetcher
func NewCrawlingService(entClient *ent.Client, aiProvider *ai.AIProvider) *CrawlingService {
	return &CrawlingService{
		entClient:        entClient,
		aiProvider:       aiProvider,
		jobkoreaParser:   crawler.NewJobKoreaParser(),
		catchParser:      crawler.NewCatchParser(),
		saraminParser:    crawler.NewSaraminParser(),
		wantedExtractor:  crawler.NewWantedExtractor(),
		headlessRenderer: crawler.NewHeadlessFetcher(),
		htmlFetcher: &defaultHTMLFetcher{
			client: &http.Client{Timeout: 30 * time.Second},
		},
	}
}

// NewCrawlingServiceWithFetcher creates a new CrawlingService with a custom HTMLFetcher (for testing)
func NewCrawlingServiceWithFetcher(entClient *ent.Client, aiProvider *ai.AIProvider, fetcher HTMLFetcher) *CrawlingService {
	return &CrawlingService{
		entClient:        entClient,
		aiProvider:       aiProvider,
		jobkoreaParser:   crawler.NewJobKoreaParser(),
		catchParser:      crawler.NewCatchParser(),
		saraminParser:    crawler.NewSaraminParser(),
		wantedExtractor:  crawler.NewWantedExtractor(),
		headlessRenderer: crawler.NewHeadlessFetcher(),
		htmlFetcher:      fetcher,
	}
}

// NewCrawlingServiceForTest creates a CrawlingService with all dependencies injectable (for testing)
func NewCrawlingServiceForTest(aiProvider *ai.AIProvider, fetcher HTMLFetcher, headless HeadlessRenderer) *CrawlingService {
	return &CrawlingService{
		aiProvider:       aiProvider,
		jobkoreaParser:   crawler.NewJobKoreaParser(),
		catchParser:      crawler.NewCatchParser(),
		saraminParser:    crawler.NewSaraminParser(),
		wantedExtractor:  crawler.NewWantedExtractor(),
		headlessRenderer: headless,
		htmlFetcher:      fetcher,
	}
}

// CrawlJobPosting crawls and normalizes a job posting from URL using a multi-tier pipeline:
//   - SPA detection: If page is a JS shell → headless Chrome render → replace HTML
//   - Tier 1: Known domains (jobkorea/catch) → CSS parser → normalizeWithAI
//   - Tier 1.5: Saramin relay → AJAX fetch for real content
//   - Tier 2: Universal → readability + markdown → LLM extraction
//   - Tier 3: Fallback → truncated raw HTML → LLM extraction
func (s *CrawlingService) CrawlJobPosting(ctx context.Context, url string) (*crawler.JobPosting, error) {
	// 1. Fetch HTML
	html, err := s.htmlFetcher.FetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}

	// 2. Detect SPA pages — attempt headless browser rendering
	if isSPAPage(html) {
		slog.Info("crawl_spa_detected", "url", url)
		renderedHTML, headlessErr := s.headlessRenderer.FetchRenderedHTML(ctx, url)
		if headlessErr != nil {
			slog.Warn("crawl_headless_failed", "url", url, "error", headlessErr)
			return nil, fmt.Errorf("이 사이트는 JavaScript로 렌더링됩니다. 헤드리스 브라우저 렌더링도 실패했습니다: %w", headlessErr)
		}
		slog.Info("crawl_headless_success", "url", url, "html_len", len(renderedHTML))
		html = renderedHTML // replace SPA shell with rendered content
	}

	// Tier 1: Known domain CSS parsers (fast, rich content via selectionToMarkdown)
	if containsDomain(url, "jobkorea.co.kr") {
		rawPosting, parseErr := s.jobkoreaParser.ParseHTML(url, html)
		if parseErr == nil && rawPosting.CompanyName != "" {
			// Fetch S3 description for rich content (new Next.js format)
			if descURL := s.jobkoreaParser.DescriptionURL(html); descURL != "" {
				if descHTML, fetchErr := s.htmlFetcher.FetchHTML(descURL); fetchErr == nil {
					s.jobkoreaParser.ParseDescriptionHTML(rawPosting, descHTML)
				}
			}
			slog.Info("crawl_tier1_jobkorea", "url", url, "has_main_tasks", rawPosting.MainTasks != "")
			return s.normalizeWithAI(ctx, rawPosting)
		}
	} else if containsDomain(url, "catch.co.kr") {
		rawPosting, parseErr := s.catchParser.ParseHTML(url, html)
		if parseErr == nil {
			slog.Info("crawl_tier1_catch", "url", url)
			return s.normalizeWithAI(ctx, rawPosting)
		}
	}

	// Tier 1.5: Saramin relay pages — AJAX fetch + CSS parser for rich content
	if containsDomain(url, "saramin.co.kr") {
		if ajaxHTML, ajaxErr := s.fetchSaraminAjax(url); ajaxErr == nil && len(ajaxHTML) > 1000 {
			slog.Info("crawl_saramin_ajax", "url", url, "ajax_len", len(ajaxHTML))
			if rawPosting, parseErr := s.saraminParser.ParseHTML(url, ajaxHTML); parseErr == nil && rawPosting.CompanyName != "" {
				slog.Info("crawl_tier1.5_saramin_css", "url", url)
				return s.normalizeWithAI(ctx, rawPosting)
			}
			html = ajaxHTML // CSS parser failed — use AJAX HTML for downstream tiers
		}
	}

	// Tier 1.6: Wanted — extract from __NEXT_DATA__ JSON (no LLM for extraction)
	if containsDomain(url, "wanted.co.kr") {
		if rawPosting, err := s.wantedExtractor.ExtractFromNextData(html); err == nil && rawPosting != nil && rawPosting.CompanyName != "" {
			slog.Info("crawl_tier1.6_wanted_nextdata", "url", url)
			return s.normalizeWithAI(ctx, rawPosting)
		}
	}

	// Tier 2: Readability + Markdown
	markdown, _ := crawler.ScrapeToMarkdown(html, url)
	if len(markdown) >= crawler.MinMarkdownLength {
		// Tier 1.7: Universal extraction — attempt to extract sections without LLM
		mdExtracted := crawler.ExtractFromMarkdown(markdown)
		htmlExtracted := crawler.ExtractFromHTML(html, url)
		merged := crawler.MergeExtractions(mdExtracted, htmlExtracted)

		if merged != nil && merged.IsEnoughForNormalize() {
			slog.Info("crawl_tier1.7_universal", "url", url, "filled", merged.FilledFields)
			return s.normalizeWithAI(ctx, &merged.Raw)
		}

		// Smart trimming for LLM — use section-based priority instead of blind truncation
		trimmed := crawler.SmartTrimForLLM(markdown, merged, maxMarkdownLen)
		slog.Info("crawl_tier2_markdown", "url", url, "original_len", len(markdown), "trimmed_len", len(trimmed))
		return s.extractJobPostingFromMarkdown(ctx, url, trimmed)
	}

	// Tier 3: Raw HTML + LLM (last resort)
	slog.Info("crawl_tier3_fallback", "url", url, "markdown_len", len(markdown))
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
		slog.Warn("crawling_prompt_fallback", "sub_category", subCategory, "error", err)
	}
	// Fallback defaults for tests or missing seed data
	return crawlingDefaultPrompt(subCategory), nil
}

// crawlingDefaultPrompt returns hardcoded defaults for crawling prompts
func crawlingDefaultPrompt(subCategory string) *ent.PromptTemplate {
	defaults := map[string]*ent.PromptTemplate{
		"extract_markdown": {
			Model:       "groq/compound",
			SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from the provided content. Always respond in valid JSON.",
			UserPromptTemplate: "URL: {{source_url}}\n\n{{content}}\n\nExtract job posting fields as JSON.",
			Temperature: 0.1,
			MaxTokens:   2000,
		},
		"extract_html": {
			Model:       "groq/compound",
			SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from raw HTML. Always respond in valid JSON.",
			UserPromptTemplate: "URL: {{source_url}}\n\n{{content}}\n\nExtract job posting fields as JSON.",
			Temperature: 0.1,
			MaxTokens:   2000,
		},
		"normalize": {
			Model:       "groq/compound",
			SystemPrompt: "You are a job posting data extractor. Extract structured information from raw text.",
			UserPromptTemplate: "Company: {{company_name}}\nPosition: {{position}}\nDepartment: {{department}}\nCareer: {{career}}\nLocation: {{location}}\nMain Tasks: {{main_tasks}}\nRequirements: {{requirements}}\nPreferred: {{preferred}}\nSkills: {{skills}}\n\nReturn structured JSON.",
			Temperature: 0.2,
			MaxTokens:   2000,
		},
	}
	if pt, ok := defaults[subCategory]; ok {
		return pt
	}
	return &ent.PromptTemplate{Model: "groq/compound", Temperature: 0.1, MaxTokens: 2000}
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

// fetchSaraminAjax fetches real job content from Saramin's AJAX endpoint.
// Saramin relay view pages load job details via XHR, so the initial HTML is a shell.
// This method extracts rec_idx from the URL and calls the AJAX endpoint directly.
func (s *CrawlingService) fetchSaraminAjax(sourceURL string) (string, error) {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return "", err
	}
	recIdx := parsed.Query().Get("rec_idx")
	if recIdx == "" {
		return "", fmt.Errorf("no rec_idx in Saramin URL")
	}

	ajaxURL := "https://www.saramin.co.kr/zf_user/jobs/relay/view-ajax?rec_idx=" + recIdx

	req, err := http.NewRequest("GET", ajaxURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", sourceURL)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Saramin AJAX HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
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

// spaBodyPattern matches SPA shell pages where <body> contains only empty root divs and scripts.
var spaBodyPattern = regexp.MustCompile(`(?is)<body[^>]*>\s*(<div\s+id="(root|app|__next|__nuxt)"[^>]*>\s*</div>\s*)+`)

// isSPAPage detects JavaScript-rendered SPA pages that have no server-side content.
// These pages return an HTML shell with an empty root div and JS bundles.
func isSPAPage(html string) bool {
	if len(html) > 5000 {
		return false // real content pages are typically much larger
	}
	return spaBodyPattern.MatchString(html)
}
