package service

import (
	"context"
	"os"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
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

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	service := NewCrawlingService(aiProvider)

	testURL := "https://www.jobkorea.co.kr/Recruit/GI_Read/45942867"

	ctx := context.Background()
	result, err := service.CrawlJobPosting(ctx, testURL)

	if err != nil {
		t.Logf("Warning: Crawl failed (URL may be expired): %v", err)
		return
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result.CompanyName, "Should extract company name from real page")
	assert.NotEmpty(t, result.Position, "Should extract position from real page")

	t.Logf("Successfully crawled: %s - %s", result.CompanyName, result.Position)
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

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	service := NewCrawlingService(aiProvider)

	testURL := "https://www.catch.co.kr/NCS/RecruitInfoDetail/321177"

	ctx := context.Background()
	result, err := service.CrawlJobPosting(ctx, testURL)

	if err != nil {
		t.Logf("Warning: Crawl failed (URL may be expired): %v", err)
		return
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result.CompanyName, "Should extract company name from Catch")

	t.Logf("Successfully crawled Catch: %s - %s", result.CompanyName, result.Position)
}

// Manual test function for debugging actual crawling
func TestCrawlJobPosting_Manual(t *testing.T) {
	t.Skip("Manual test - run individually with: go test -run TestCrawlJobPosting_Manual -v")

	mockLLM := &MockLLMForCrawling{}
	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(aiProvider)

	urls := []string{
		"https://www.jobkorea.co.kr/Recruit/GI_Read/45942867",
		"https://www.catch.co.kr/NCS/RecruitInfoDetail/321177",
	}

	for _, url := range urls {
		t.Logf("\n=== Testing URL: %s ===", url)

		html, err := svc.htmlFetcher.FetchHTML(url)
		if err != nil {
			t.Logf("Fetch error: %v", err)
			continue
		}

		t.Logf("HTML length: %d bytes", len(html))

		var rawPosting interface{}
		if containsDomain(url, "jobkorea") {
			rawPosting, err = svc.jobkoreaParser.ParseHTML(url, html)
		} else if containsDomain(url, "catch") {
			rawPosting, err = svc.catchParser.ParseHTML(url, html)
		}

		if err != nil {
			t.Logf("Parse error: %v", err)
		} else {
			t.Logf("Parsed successfully: %+v", rawPosting)
		}
	}
}
