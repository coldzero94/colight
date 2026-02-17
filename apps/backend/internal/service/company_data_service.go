package service

import (
	"context"
	"fmt"

	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
)

// CompanyDataService aggregates company information from multiple sources
type CompanyDataService struct {
	searchCrawler *crawler.NaverSearchCrawler
	newsCrawler   *crawler.NaverNewsCrawler
}

// NewCompanyDataService creates a new company data service
func NewCompanyDataService() *CompanyDataService {
	return &CompanyDataService{
		searchCrawler: crawler.NewNaverSearchCrawler(),
		newsCrawler:   crawler.NewNaverNewsCrawler(),
	}
}

// CompanyData represents aggregated company information
type CompanyData struct {
	CompanyContext string                `json:"company_context"`
	News           []crawler.NewsArticle `json:"news"`
}

// GetCompanyData fetches and aggregates all company data
func (s *CompanyDataService) GetCompanyData(ctx context.Context, companyName string) (*CompanyData, error) {
	result := &CompanyData{}

	// 1. Search Naver for company info (graceful: returns "" on failure)
	companyContext, _ := s.searchCrawler.SearchCompanyInfo(companyName)
	result.CompanyContext = companyContext

	// 2. Get recent news from Naver
	news, err := s.newsCrawler.SearchCompanyNews(companyName, 5)
	if err != nil {
		news = []crawler.NewsArticle{}
	}
	result.News = news

	return result, nil
}

// GetCompanyByJobPostingURL extracts company name from job posting and fetches data
func (s *CompanyDataService) GetCompanyByJobPostingURL(ctx context.Context, jobPostingURL string, companyName string) (*CompanyData, error) {
	if companyName == "" {
		return nil, fmt.Errorf("company name is required")
	}

	return s.GetCompanyData(ctx, companyName)
}
