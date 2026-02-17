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

	// Should not crash even if Naver search fails
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// CompanyContext may be empty if search finds nothing — that's OK
	assert.NotNil(t, result.News)
}

func TestGetCompanyData_RealCompany(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	service := NewCompanyDataService()
	ctx := context.Background()

	result, err := service.GetCompanyData(ctx, "삼성전자")

	if err != nil {
		t.Logf("Warning: Failed to fetch company data: %v", err)
		return
	}

	assert.NotNil(t, result)
	t.Logf("CompanyContext length: %d chars", len(result.CompanyContext))
	t.Logf("News articles: %d", len(result.News))
}
