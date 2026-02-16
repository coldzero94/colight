package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
	aiProvider     *ai.AIProvider
	jobkoreaParser *crawler.JobKoreaParser
	catchParser    *crawler.CatchParser
	htmlFetcher    HTMLFetcher
}

// NewCrawlingService creates a new CrawlingService with default HTTP fetcher
func NewCrawlingService(aiProvider *ai.AIProvider) *CrawlingService {
	return &CrawlingService{
		aiProvider:     aiProvider,
		jobkoreaParser: crawler.NewJobKoreaParser(),
		catchParser:    crawler.NewCatchParser(),
		htmlFetcher: &defaultHTMLFetcher{
			client: &http.Client{Timeout: 30 * time.Second},
		},
	}
}

// NewCrawlingServiceWithFetcher creates a new CrawlingService with a custom HTMLFetcher (for testing)
func NewCrawlingServiceWithFetcher(aiProvider *ai.AIProvider, fetcher HTMLFetcher) *CrawlingService {
	return &CrawlingService{
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
	content := truncateString(markdown, maxMarkdownLen)

	prompt := fmt.Sprintf(`다음은 채용공고 페이지에서 추출된 마크다운 콘텐츠입니다. 구조화된 JSON으로 변환해주세요.

URL: %s

--- 콘텐츠 ---
%s
--- 끝 ---

다음 필드를 포함한 JSON을 반환하세요:
- company_name: string (회사명)
- position: string (포지션)
- department: string (부서/팀)
- job_type: string (고용형태: 정규직/계약직/인턴 등)
- experience_level: string (경력 요건)
- main_tasks: string[] (주요 업무)
- requirements: string[] (자격요건)
- preferred: string[] (우대사항)
- required_skills: string[] (필수 기술 스택)
- soft_skills: string[] (소프트 스킬)
- company_values_hints: string[] (회사 가치/문화 힌트)
- deadline: string (마감일, 없으면 빈 문자열)`, sourceURL, content)

	resp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, "groq", ai.LLMRequest{
		SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from the provided content. Always respond in valid JSON.",
		UserPrompt:   prompt,
		Temperature:  0.1,
		MaxTokens:    2000,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())

	if err != nil {
		return nil, fmt.Errorf("AI extraction from markdown failed: %w", err)
	}

	var result crawler.JobPosting
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// extractJobPostingFromHTML extracts structured job posting data from raw HTML (fallback)
func (s *CrawlingService) extractJobPostingFromHTML(ctx context.Context, sourceURL string, html string) (*crawler.JobPosting, error) {
	content := truncateString(html, maxHTMLLen)

	prompt := fmt.Sprintf(`다음은 채용공고 페이지의 HTML입니다. 구조화된 JSON으로 변환해주세요.

URL: %s

--- HTML ---
%s
--- 끝 ---

다음 필드를 포함한 JSON을 반환하세요:
- company_name: string (회사명)
- position: string (포지션)
- department: string (부서/팀)
- job_type: string (고용형태)
- experience_level: string (경력 요건)
- main_tasks: string[] (주요 업무)
- requirements: string[] (자격요건)
- preferred: string[] (우대사항)
- required_skills: string[] (필수 기술 스택)
- soft_skills: string[] (소프트 스킬)
- company_values_hints: string[] (회사 가치/문화 힌트)
- deadline: string (마감일, 없으면 빈 문자열)`, sourceURL, content)

	resp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, "groq", ai.LLMRequest{
		SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from raw HTML. Always respond in valid JSON.",
		UserPrompt:   prompt,
		Temperature:  0.1,
		MaxTokens:    2000,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())

	if err != nil {
		return nil, fmt.Errorf("AI extraction from HTML failed: %w", err)
	}

	var result crawler.JobPosting
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// normalizeWithAI uses LLM to normalize RawJobPosting into structured JobPosting
func (s *CrawlingService) normalizeWithAI(ctx context.Context, raw *crawler.RawJobPosting) (*crawler.JobPosting, error) {
	prompt := fmt.Sprintf(`Extract and structure this job posting data into JSON format.

Company: %s
Position: %s
Department: %s
Career: %s
Location: %s
Main Tasks: %s
Requirements: %s
Preferred: %s
Skills: %v

Return JSON with these fields:
- company_name: string
- position: string
- department: string
- job_type: string
- experience_level: string
- main_tasks: string[]
- requirements: string[]
- preferred: string[]
- required_skills: string[]
- soft_skills: string[]
- company_values_hints: string[]
- deadline: string`,
		raw.CompanyName, raw.Position, raw.Department, raw.Career, raw.Location,
		raw.MainTasks, raw.Requirements, raw.Preferred, raw.Skills)

	resp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, "groq", ai.LLMRequest{
		SystemPrompt: "You are a job posting data extractor. Extract structured information from raw text.",
		UserPrompt:   prompt,
		Temperature:  0.2,
		MaxTokens:    2000,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())

	if err != nil {
		return nil, err
	}

	var result crawler.JobPosting
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
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
