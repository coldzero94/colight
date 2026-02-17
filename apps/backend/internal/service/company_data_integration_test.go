package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Full integration test - crawls real websites
func TestGetCompanyData_FullPipeline_Real(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	service := NewCompanyDataService()
	ctx := context.Background()

	testCases := []struct {
		companyName string
		expectNews  bool
	}{
		{"삼성전자", true},
		{"네이버", true},
		{"카카오", true},
		{"토스", true},
		{"당근마켓", true},
	}

	for _, tc := range testCases {
		t.Run(tc.companyName, func(t *testing.T) {
			t.Logf("\n=== Testing company: %s ===", tc.companyName)

			result, err := service.GetCompanyData(ctx, tc.companyName)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Log Naver search results
			if result.CompanyContext != "" {
				t.Logf("✅ Company context: %d chars", len(result.CompanyContext))
				if len(result.CompanyContext) > 200 {
					t.Logf("   Preview: %s...", result.CompanyContext[:200])
				}
			} else {
				t.Logf("⚠️  No company context found")
			}

			// Log News results
			t.Logf("📰 News: Found %d articles", len(result.News))
			for i, article := range result.News {
				if i >= 3 {
					break
				}
				t.Logf("  [%d] %s", i+1, article.Title)
			}

			if tc.expectNews {
				assert.NotEmpty(t, result.News, "Should find news for major company: "+tc.companyName)
			}
		})
	}
}

func TestNaverNewsCrawler_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	service := NewCompanyDataService()

	t.Run("Major company news", func(t *testing.T) {
		articles, err := service.newsCrawler.SearchCompanyNews("삼성전자", 5)

		assert.NoError(t, err)
		assert.NotNil(t, articles)

		t.Logf("Found %d articles", len(articles))

		if len(articles) > 0 {
			t.Log("Sample articles:")
			for i, article := range articles {
				if i >= 3 {
					break
				}
				t.Logf("  %d. %s", i+1, article.Title)
				t.Logf("     Link: %s", article.Link)
			}

			assert.NotEmpty(t, articles[0].Title)
			assert.NotEmpty(t, articles[0].Link)
			assert.Contains(t, articles[0].Link, "news.naver.com")
		}
	})

	t.Run("Startup company news", func(t *testing.T) {
		articles, err := service.newsCrawler.SearchCompanyNews("토스", 3)

		assert.NoError(t, err)
		t.Logf("Startup news: Found %d articles", len(articles))
	})
}

// End-to-end test: Job posting URL → Company data
func TestCompanyData_EndToEnd(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	service := NewCompanyDataService()
	ctx := context.Background()

	jobURL := "https://www.jobkorea.co.kr/Recruit/GI_Read/45942867"
	companyName := "테스트회사"

	result, err := service.GetCompanyByJobPostingURL(ctx, jobURL, companyName)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	t.Logf("End-to-end test complete")
	t.Logf("CompanyContext length: %d chars", len(result.CompanyContext))
	t.Logf("News count: %d", len(result.News))
}
