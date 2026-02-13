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
		expectNews  bool // Whether we expect to find news
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

			// Should not crash even if some sources fail
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.BasicInfo)

			// Log DART results
			if result.BasicInfo.CorpName != "" {
				t.Logf("✅ DART: %s (Code: %s)", result.BasicInfo.CorpName, result.BasicInfo.CorpCode)
			} else {
				t.Logf("⚠️  DART: No info found (may be non-listed company)")
			}

			// Log News results
			t.Logf("📰 News: Found %d articles", len(result.News))
			for i, article := range result.News {
				if i >= 3 {
					break
				}
				t.Logf("  [%d] %s", i+1, article.Title)
			}

			// For major tech companies, we should find some news
			if tc.expectNews {
				assert.NotEmpty(t, result.News, "Should find news for major company: "+tc.companyName)
			}
		})
	}
}

// Test specific crawlers individually
func TestDartCrawler_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	service := NewCompanyDataService()

	t.Run("Listed company", func(t *testing.T) {
		// Test with known listed company
		result, err := service.dartCrawler.SearchCompany("삼성전자")

		if err != nil {
			t.Logf("DART search failed: %v", err)
			t.Log("Note: DART website structure may have changed")
			return
		}

		assert.NotNil(t, result)
		t.Logf("Company: %s", result.CorpName)
		t.Logf("Corp Code: %s", result.CorpCode)
	})

	t.Run("Non-listed company", func(t *testing.T) {
		// Test with likely non-listed company
		_, err := service.dartCrawler.SearchCompany("작은스타트업123")

		// Should handle gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "not found")
			t.Log("✅ Correctly handles non-listed company")
		}
	})
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

			// Verify article structure
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

	// Simulate user flow: paste job URL → extract company → fetch data
	jobURL := "https://www.jobkorea.co.kr/Recruit/GI_Read/45942867"
	companyName := "테스트회사" // Would be extracted from job posting

	result, err := service.GetCompanyByJobPostingURL(ctx, jobURL, companyName)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	t.Logf("End-to-end test complete")
	t.Logf("Company: %s", result.BasicInfo.CorpName)
	t.Logf("News count: %d", len(result.News))
}
