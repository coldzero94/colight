package crawler

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNaverSearchCrawler_SearchCompanyInfo_EmptyName(t *testing.T) {
	c := NewNaverSearchCrawler()
	result, err := c.SearchCompanyInfo("")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestNaverSearchCrawler_ParseSearchResults(t *testing.T) {
	// Test the link extraction logic with a fixture
	html := `<html><body>
		<div class="search_result">
			<a href="https://www.toss.im/about" class="link_tit">토스 기업소개</a>
			<a href="https://news.naver.com/article/123">뉴스 기사</a>
			<a href="https://namu.wiki/w/토스">토스 - 나무위키</a>
			<a href="https://www.jobplanet.co.kr/companies/12345">토스 기업리뷰</a>
		</div>
	</body></html>`

	links := parseSearchResultLinks(strings.NewReader(html), 3)
	// Should exclude news.naver.com links
	for _, link := range links {
		assert.NotContains(t, link, "news.naver.com")
	}
	assert.LessOrEqual(t, len(links), 3)
}

func TestNaverSearchCrawler_ParseSearchResults_NoLinks(t *testing.T) {
	html := `<html><body><div>No relevant links here</div></body></html>`
	links := parseSearchResultLinks(strings.NewReader(html), 3)
	assert.Empty(t, links)
}

func TestNaverSearchCrawler_SearchCompanyInfo_Real(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	c := NewNaverSearchCrawler()

	tests := []struct {
		name        string
		companyName string
	}{
		{"listed company", "삼성전자"},
		{"startup", "토스"},
		{"non-listed", "당근마켓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.SearchCompanyInfo(tt.companyName)
			require.NoError(t, err)
			t.Logf("Company: %s, Context length: %d chars", tt.companyName, len(result))
			if len(result) > 200 {
				t.Logf("Preview: %s...", result[:200])
			} else {
				t.Logf("Preview: %s", result)
			}
		})
	}
}
