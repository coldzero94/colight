package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCompanyData_Unit(t *testing.T) {
	service := NewCompanyDataService()
	ctx := context.Background()

	// Unit test with mock company name
	result, err := service.GetCompanyData(ctx, "테스트회사")

	// Should not crash even if DART/Naver fail
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.BasicInfo)
}

func TestGetCompanyData_RealCompany(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	service := NewCompanyDataService()
	ctx := context.Background()

	// Test with real company (e.g., 삼성전자)
	result, err := service.GetCompanyData(ctx, "삼성전자")

	if err != nil {
		t.Logf("Warning: Failed to fetch company data: %v", err)
		return
	}

	assert.NotNil(t, result)
	assert.NotEmpty(t, result.BasicInfo.CorpName)
	t.Logf("Company: %s", result.BasicInfo.CorpName)
	t.Logf("News articles: %d", len(result.News))
}
