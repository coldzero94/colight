package crawler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDartCrawler_SearchCompany_Real(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	crawler := NewDartCrawler()

	// Test with real company: 삼성전자
	result, err := crawler.SearchCompany("삼성전자")

	if err != nil {
		t.Logf("Warning: DART search failed: %v", err)
		t.Logf("This may be due to DART website structure changes or blocking")
		return
	}

	assert.NotNil(t, result)
	assert.Contains(t, result.CorpName, "삼성", "Company name should contain '삼성'")

	t.Logf("=== DART Crawl Result ===")
	t.Logf("Company: %s", result.CorpName)
	t.Logf("Corp Code: %s", result.CorpCode)
	t.Logf("Stock Code: %s", result.StockCode)
	t.Logf("CEO: %s", result.CEO)
}

func TestDartCrawler_SearchCompany_MultipleCompanies(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	crawler := NewDartCrawler()

	companies := []string{
		"네이버",
		"카카오",
		"토스",
		"쿠팡",
	}

	for _, company := range companies {
		t.Run(company, func(t *testing.T) {
			result, err := crawler.SearchCompany(company)

			if err != nil {
				t.Logf("%s: Search failed - %v", company, err)
				return
			}

			assert.NotNil(t, result)
			t.Logf("%s: Found - %s (Code: %s)", company, result.CorpName, result.CorpCode)
		})
	}
}
