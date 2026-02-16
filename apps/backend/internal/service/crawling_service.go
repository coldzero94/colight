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

// CrawlingService handles job posting crawling and parsing
type CrawlingService struct {
	aiProvider      *ai.AIProvider
	jobkoreaParser  *crawler.JobKoreaParser
	catchParser     *crawler.CatchParser
	httpClient      *http.Client
}

// NewCrawlingService creates a new CrawlingService
func NewCrawlingService(aiProvider *ai.AIProvider) *CrawlingService {
	return &CrawlingService{
		aiProvider:     aiProvider,
		jobkoreaParser: crawler.NewJobKoreaParser(),
		catchParser:    crawler.NewCatchParser(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CrawlJobPosting crawls and normalizes a job posting from URL
func (s *CrawlingService) CrawlJobPosting(ctx context.Context, url string) (*crawler.JobPosting, error) {
	// 1. Fetch HTML
	html, err := s.fetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}

	// 2. Parse based on domain
	var rawPosting *crawler.RawJobPosting

	// Detect domain (simple check)
	if containsDomain(url, "jobkorea.co.kr") {
		rawPosting, err = s.jobkoreaParser.ParseHTML(url, html)
	} else if containsDomain(url, "catch.co.kr") {
		rawPosting, err = s.catchParser.ParseHTML(url, html)
	} else {
		// AI fallback for unknown domains
		rawPosting = &crawler.RawJobPosting{
			Source:    "unknown",
			SourceURL: url,
			RawHTML:   html,
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// 3. Normalize with AI
	normalized, err := s.normalizeWithAI(ctx, rawPosting)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize: %w", err)
	}

	return normalized, nil
}

// fetchHTML fetches HTML from URL with proper User-Agent
func (s *CrawlingService) fetchHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Set User-Agent to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.httpClient.Do(req)
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

// normalizeWithAI uses Gemini to normalize RawJobPosting into structured JobPosting
func (s *CrawlingService) normalizeWithAI(ctx context.Context, raw *crawler.RawJobPosting) (*crawler.JobPosting, error) {
	// Build prompt
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

	// Parse JSON response
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
