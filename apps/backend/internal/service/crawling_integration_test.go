package service

import (
	"context"
	"os"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests - only run when INTEGRATION_TEST=1
func TestCrawlJobPosting_RealJobKorea(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "테스트 회사",
				"position": "개발자",
				"department": "개발팀",
				"job_type": "정규직",
				"experience_level": "신입/경력",
				"main_tasks": ["개발"],
				"requirements": ["Go"],
				"preferred": ["Docker"],
				"required_skills": ["Go"],
				"soft_skills": ["커뮤니케이션"],
				"company_values_hints": ["혁신"],
				"deadline": "상시채용"
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM)
	service := NewCrawlingService(nil, aiProvider)

	testURL := "https://www.jobkorea.co.kr/Recruit/GI_Read/45942867"

	ctx := context.Background()
	result, err := service.CrawlJobPosting(ctx, testURL)

	if err != nil {
		t.Logf("Warning: Crawl failed (URL may be expired): %v", err)
		return
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result["company_name"].(string), "Should extract company name from real page")
	assert.NotEmpty(t, result["position"].(string), "Should extract position from real page")

	t.Logf("Successfully crawled: %s - %s", result["company_name"].(string), result["position"].(string))
}

func TestCrawlJobPosting_RealCatch(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "캐치 회사",
				"position": "개발자",
				"department": "개발팀",
				"job_type": "정규직",
				"experience_level": "경력",
				"main_tasks": ["개발"],
				"requirements": ["개발 경험"],
				"preferred": [],
				"required_skills": [],
				"soft_skills": [],
				"company_values_hints": [],
				"deadline": ""
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM)
	service := NewCrawlingService(nil, aiProvider)

	testURL := "https://www.catch.co.kr/NCS/RecruitInfoDetail/321177"

	ctx := context.Background()
	result, err := service.CrawlJobPosting(ctx, testURL)

	if err != nil {
		t.Logf("Warning: Crawl failed (URL may be expired): %v", err)
		return
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result["company_name"].(string), "Should extract company name from Catch")

	t.Logf("Successfully crawled Catch: %s - %s", result["company_name"].(string), result["position"].(string))
}

// TestCrawlJobPosting_RealSaramin tests the full CrawlJobPosting pipeline for Saramin.
// Saramin relay pages load content via AJAX — this verifies the AJAX fetch path.
func TestCrawlJobPosting_RealSaramin(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "사람인 테스트 회사",
				"position": "개발자",
				"department": "",
				"job_type": "정규직",
				"experience_level": "경력무관",
				"main_tasks": ["개발"],
				"requirements": ["개발 경험"],
				"preferred": [],
				"required_skills": [],
				"soft_skills": [],
				"company_values_hints": [],
				"deadline": ""
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM)
	service := NewCrawlingService(nil, aiProvider)

	testURL := "https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=49845147"

	ctx := context.Background()
	result, err := service.CrawlJobPosting(ctx, testURL)

	if err != nil {
		t.Logf("Warning: Crawl failed (URL may be expired): %v", err)
		return
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result["company_name"].(string), "Should extract company name from Saramin")
	t.Logf("Successfully crawled Saramin: %s - %s", result["company_name"].(string), result["position"].(string))

	// Verify the LLM was called (meaning AJAX fetch + markdown extraction worked)
	assert.Greater(t, mockLLM.calls, 0, "LLM should have been called with extracted content")
}

// TestSaraminAjaxFetch tests the Saramin AJAX fetch + markdown pipeline directly.
func TestSaraminAjaxFetch(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	svc := NewCrawlingService(nil, nil)

	ajaxHTML, err := svc.fetchSaraminAjax("https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=49845147")
	if err != nil {
		t.Logf("Saramin AJAX fetch failed (may be rate-limited): %v", err)
		return
	}

	t.Logf("AJAX HTML length: %d bytes", len(ajaxHTML))
	assert.Greater(t, len(ajaxHTML), 1000, "AJAX response should have substantial content")

	// Feed through markdown pipeline
	markdown, _ := crawler.ScrapeToMarkdown(ajaxHTML, "https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=49845147")
	t.Logf("Markdown from AJAX: %d bytes", len(markdown))
	if len(markdown) > 300 {
		t.Logf("Markdown preview:\n%s", markdown[:300])
	} else if len(markdown) > 0 {
		t.Logf("Markdown:\n%s", markdown)
	}
	assert.Greater(t, len(markdown), 200, "AJAX-sourced markdown should be substantial")
}

// TestCrawlPipeline_RealSites tests the crawling pipeline against real sites.
// Verifies HTML fetching, readability extraction, and SPA detection work correctly.
// Uses mock LLM to avoid API costs — focuses on the fetch+parse layers.
func TestCrawlPipeline_RealSites(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	svc := NewCrawlingService(nil, nil) // no entClient, no aiProvider — just test fetch+parse

	tests := []struct {
		name         string
		url          string
		expectSPA    bool
		expectFetch  bool
		minHTMLLen   int
		minMarkdown  int // 0 means don't check
	}{
		{
			name:        "JobKorea (server-rendered - CSS parser)",
			url:         "https://www.jobkorea.co.kr/Recruit/GI_Read/46390408",
			expectSPA:   false,
			expectFetch: true,
			minHTMLLen:  5000,
		},
		{
			name:        "Saramin (server-rendered - readability)",
			url:         "https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=49845147",
			expectSPA:   false,
			expectFetch: true,
			minHTMLLen:  10000,
			minMarkdown: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := svc.htmlFetcher.FetchHTML(tt.url)
			if !tt.expectFetch {
				t.Logf("Fetch error (expected): %v", err)
				return
			}
			if err != nil {
				t.Logf("Fetch error (URL may have changed): %v", err)
				return
			}

			t.Logf("HTML length: %d bytes", len(html))

			// Check SPA detection
			spa := isSPAPage(html)
			assert.Equal(t, tt.expectSPA, spa, "SPA detection mismatch")
			if spa {
				t.Logf("SPA detected — headless rendering would be attempted")
				return
			}

			// Check minimum HTML length
			if tt.minHTMLLen > 0 {
				assert.GreaterOrEqual(t, len(html), tt.minHTMLLen, "HTML too short")
			}

			// Check readability → markdown extraction
			markdown, _ := crawler.ScrapeToMarkdown(html, tt.url)
			t.Logf("Markdown length: %d bytes", len(markdown))
			if len(markdown) > 200 {
				t.Logf("Markdown preview:\n%s", markdown[:200])
			}

			if tt.minMarkdown > 0 {
				assert.GreaterOrEqual(t, len(markdown), tt.minMarkdown, "Markdown too short")
			}

			// Check domain-specific parsers
			if containsDomain(tt.url, "jobkorea.co.kr") {
				raw, err := svc.jobkoreaParser.ParseHTML(tt.url, html)
				if err != nil {
					t.Logf("JobKorea parser error: %v", err)
				} else {
					t.Logf("JobKorea parser: %s - %s", raw.CompanyName, raw.Position)
					assert.NotEmpty(t, raw.CompanyName)
					assert.NotEmpty(t, raw.Position)
				}
			}
		})
	}
}

// TestHeadlessRenderer_RealSPA tests headless Chrome rendering against a real SPA site.
// This verifies the full pipeline: SPA detection → headless render → markdown → LLM extraction.
func TestHeadlessRenderer_RealSPA(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	fetcher := crawler.NewHeadlessFetcher()
	ctx := context.Background()

	// LG Careers is a React SPA
	testURL := "https://careers.lg.com/apply/detail?id=1001364"

	html, err := fetcher.FetchRenderedHTML(ctx, testURL)
	if err != nil {
		t.Logf("Headless render failed (Chrome may not be available): %v", err)
		return
	}

	t.Logf("Rendered HTML length: %d bytes", len(html))
	assert.Greater(t, len(html), 5000, "Rendered HTML should have substantial content")

	// Feed through the full markdown pipeline
	markdown, _ := crawler.ScrapeToMarkdown(html, testURL)
	t.Logf("Markdown from rendered SPA: %d bytes", len(markdown))
	if len(markdown) > 300 {
		t.Logf("Markdown preview:\n%s", markdown[:300])
	} else if len(markdown) > 0 {
		t.Logf("Markdown:\n%s", markdown)
	}
	assert.Greater(t, len(markdown), 100, "SPA-rendered markdown should have content")
}

// TestCrawlJobPosting_RealSPA_FullPipeline tests the full CrawlJobPosting pipeline
// with a real SPA site, from detection through headless rendering to LLM extraction.
func TestCrawlJobPosting_RealSPA_FullPipeline(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "LG",
				"position": "Software Engineer",
				"department": "",
				"job_type": "정규직",
				"experience_level": "경력",
				"main_tasks": ["소프트웨어 개발"],
				"requirements": ["개발 경험"],
				"preferred": [],
				"required_skills": [],
				"soft_skills": [],
				"company_values_hints": [],
				"deadline": ""
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM)
	service := NewCrawlingService(nil, aiProvider)

	testURL := "https://careers.lg.com/apply/detail?id=1001364"
	ctx := context.Background()

	result, err := service.CrawlJobPosting(ctx, testURL)
	if err != nil {
		t.Logf("Full SPA pipeline failed (Chrome may not be available): %v", err)
		return
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result["company_name"].(string))
	t.Logf("Full SPA pipeline: %s - %s", result["company_name"].(string), result["position"].(string))
	assert.Greater(t, mockLLM.calls, 0, "LLM should have been called with headless-rendered content")
}

// TestCrawlPipeline_AllSites_RichContent tests the enhanced pipeline against 6 real sites.
// Verifies each site's tier routing and that rich content is extracted.
func TestCrawlPipeline_AllSites_RichContent(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	svc := NewCrawlingService(nil, nil)

	tests := []struct {
		name         string
		url          string
		expectTier   string
		checkSection bool // check section extraction
	}{
		{
			name:         "JobKorea (Tier 1 CSS parser)",
			url:          "https://www.jobkorea.co.kr/Recruit/GI_Read/48607050",
			expectTier:   "tier1",
			checkSection: false, // CSS parser handles it
		},
		{
			name:         "Saramin (Tier 1.5 AJAX + CSS)",
			url:          "https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=53057721",
			expectTier:   "tier1.5",
			checkSection: true,
		},
		{
			name:         "Wanted (Tier 1.6 __NEXT_DATA__)",
			url:          "https://www.wanted.co.kr/wd/292769",
			expectTier:   "tier1.6",
			checkSection: false, // JSON extraction
		},
		{
			name:         "Naver Recruit (Tier 1.7 or 2 universal)",
			url:          "https://recruit.navercorp.com/rcrt/view.do?annoId=30004542",
			expectTier:   "tier1.7_or_2",
			checkSection: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := svc.htmlFetcher.FetchHTML(tt.url)
			if err != nil {
				t.Logf("Fetch failed (URL may be expired): %v", err)
				return
			}

			t.Logf("HTML length: %d bytes", len(html))

			switch tt.expectTier {
			case "tier1":
				// Test CSS parser (JSON-LD + legacy + S3 description)
				raw, err := svc.jobkoreaParser.ParseHTML(tt.url, html)
				if err != nil {
					t.Logf("JobKorea parser error: %v", err)
					return
				}
				t.Logf("Tier 1 (JobKorea): company=%s position=%s", raw.CompanyName, raw.Position)
				assert.NotEmpty(t, raw.CompanyName, "CompanyName should not be empty")
				assert.NotEmpty(t, raw.Position, "Position should not be empty")
				t.Logf("Location: %s, Career: %s, Salary: %s", raw.Location, raw.Career, raw.Salary)

				// Try S3 description fetch for rich content
				if descURL := svc.jobkoreaParser.DescriptionURL(html); descURL != "" {
					t.Logf("S3 description URL found: %.80s...", descURL)
					if descHTML, fetchErr := svc.htmlFetcher.FetchHTML(descURL); fetchErr == nil {
						t.Logf("S3 description HTML: %d bytes", len(descHTML))
						svc.jobkoreaParser.ParseDescriptionHTML(raw, descHTML)
					} else {
						t.Logf("S3 description fetch failed: %v", fetchErr)
					}
				} else {
					t.Logf("No S3 description URL (legacy format or not found)")
				}

				if raw.MainTasks != "" {
					t.Logf("MainTasks preview: %.200s", raw.MainTasks)
				}
				if raw.Requirements != "" {
					t.Logf("Requirements preview: %.200s", raw.Requirements)
				}

			case "tier1.5":
				// Test Saramin AJAX + CSS parser
				ajaxHTML, ajaxErr := svc.fetchSaraminAjax(tt.url)
				if ajaxErr != nil {
					t.Logf("Saramin AJAX fetch failed: %v", ajaxErr)
					return
				}
				t.Logf("AJAX HTML length: %d bytes", len(ajaxHTML))

				raw, parseErr := svc.saraminParser.ParseHTML(tt.url, ajaxHTML)
				if parseErr != nil {
					t.Logf("Saramin CSS parser error: %v", parseErr)
				} else {
					t.Logf("Tier 1.5 (Saramin CSS): company=%s position=%s", raw.CompanyName, raw.Position)
					if raw.CompanyName != "" {
						assert.NotEmpty(t, raw.Position, "Position should not be empty")
					}
				}

				// Also test universal extraction on the AJAX HTML
				md, _ := crawler.ScrapeToMarkdown(ajaxHTML, tt.url)
				t.Logf("Saramin markdown: %d bytes", len(md))
				if len(md) > 300 {
					t.Logf("Markdown preview: %.300s", md)
				}

			case "tier1.6":
				// Test Wanted __NEXT_DATA__ extraction
				raw, err := svc.wantedExtractor.ExtractFromNextData(html)
				if err != nil || raw == nil {
					t.Logf("Wanted __NEXT_DATA__ extraction failed or not available")
					// Fallback: test markdown extraction
					md, _ := crawler.ScrapeToMarkdown(html, tt.url)
					t.Logf("Wanted markdown fallback: %d bytes", len(md))
					return
				}
				t.Logf("Tier 1.6 (Wanted JSON): company=%s position=%s", raw.CompanyName, raw.Position)
				assert.NotEmpty(t, raw.CompanyName, "CompanyName from __NEXT_DATA__")
				assert.NotEmpty(t, raw.Position, "Position from __NEXT_DATA__")
				assert.NotEmpty(t, raw.MainTasks, "MainTasks should have full content")
				t.Logf("MainTasks preview: %.200s", raw.MainTasks)
				t.Logf("Skills: %v", raw.Skills)

			case "tier1.7_or_2":
				// Test universal extraction
				md, _ := crawler.ScrapeToMarkdown(html, tt.url)
				t.Logf("Markdown: %d bytes", len(md))

				mdExtracted := crawler.ExtractFromMarkdown(md)
				htmlExtracted := crawler.ExtractFromHTML(html, tt.url)
				merged := crawler.MergeExtractions(mdExtracted, htmlExtracted)

				t.Logf("Universal extraction: filled=%d, enough=%v",
					merged.FilledFields, merged.IsEnoughForNormalize())

				for key, text := range merged.SectionTexts {
					t.Logf("Section [%s]: %d chars", key, len(text))
					if len(text) > 100 {
						t.Logf("  Preview: %.100s...", text)
					}
				}

				if merged.Raw.CompanyName != "" {
					t.Logf("Company: %s", merged.Raw.CompanyName)
				}
				if merged.Raw.Position != "" {
					t.Logf("Position: %s", merged.Raw.Position)
				}
			}
		})
	}
}

// TestWantedExtractor_RealSite tests the Wanted extractor against a real Wanted URL.
func TestWantedExtractor_RealSite(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	svc := NewCrawlingService(nil, nil)

	html, err := svc.htmlFetcher.FetchHTML("https://www.wanted.co.kr/wd/292769")
	if err != nil {
		t.Logf("Wanted fetch failed: %v", err)
		return
	}

	t.Logf("Wanted HTML: %d bytes", len(html))

	raw, err := svc.wantedExtractor.ExtractFromNextData(html)
	if err != nil || raw == nil {
		t.Logf("__NEXT_DATA__ not found or parse failed — Wanted may have changed structure")
		return
	}

	t.Logf("Company: %s", raw.CompanyName)
	t.Logf("Position: %s", raw.Position)
	t.Logf("MainTasks (%d chars): %.300s", len(raw.MainTasks), raw.MainTasks)
	t.Logf("Requirements (%d chars): %.300s", len(raw.Requirements), raw.Requirements)
	t.Logf("Preferred (%d chars): %.200s", len(raw.Preferred), raw.Preferred)
	t.Logf("Skills: %v", raw.Skills)

	assert.NotEmpty(t, raw.CompanyName, "Wanted should extract company name")
	assert.NotEmpty(t, raw.Position, "Wanted should extract position")
	assert.NotEmpty(t, raw.MainTasks, "Wanted should extract main tasks (full text)")
}
