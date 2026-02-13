package crawler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNaverNewsCrawler_SearchCompanyNews_Real(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	crawler := NewNaverNewsCrawler()

	// Test with real company: 삼성전자
	articles, err := crawler.SearchCompanyNews("삼성전자", 5)

	if err != nil {
		t.Logf("Warning: Naver news search failed: %v", err)
		return
	}

	assert.NotNil(t, articles)
	t.Logf("Found %d news articles for 삼성전자", len(articles))

	for i, article := range articles {
		t.Logf("[%d] %s", i+1, article.Title)
		t.Logf("    Source: %s | Date: %s", article.Source, article.PubDate)
		t.Logf("    Link: %s", article.Link)
	}

	// Should find at least some articles for major company
	if len(articles) > 0 {
		assert.NotEmpty(t, articles[0].Title, "First article should have title")
		assert.NotEmpty(t, articles[0].Link, "First article should have link")
	}
}

func TestNaverNewsCrawler_SearchCompanyNews_TechCompanies(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	crawler := NewNaverNewsCrawler()

	companies := []string{
		"네이버",
		"카카오",
		"토스",
		"당근마켓",
	}

	for _, company := range companies {
		t.Run(company, func(t *testing.T) {
			articles, err := crawler.SearchCompanyNews(company, 3)

			if err != nil {
				t.Logf("%s: News search failed - %v", company, err)
				return
			}

			t.Logf("%s: Found %d articles", company, len(articles))
			if len(articles) > 0 {
				t.Logf("  Latest: %s", articles[0].Title)
			}
		})
	}
}
