package service

import (
	"context"
	"fmt"

	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
)

// CompanyDataService aggregates company information from multiple sources
type CompanyDataService struct {
	dartCrawler  *crawler.DartCrawler
	newsCrawler  *crawler.NaverNewsCrawler
}

// NewCompanyDataService creates a new company data service
func NewCompanyDataService() *CompanyDataService {
	return &CompanyDataService{
		dartCrawler: crawler.NewDartCrawler(),
		newsCrawler: crawler.NewNaverNewsCrawler(),
	}
}

// CompanyData represents aggregated company information
type CompanyData struct {
	BasicInfo *crawler.CompanyInfo      `json:"basic_info"`
	News      []crawler.NewsArticle     `json:"news"`
}

// GetCompanyData fetches and aggregates all company data
func (s *CompanyDataService) GetCompanyData(ctx context.Context, companyName string) (*CompanyData, error) {
	result := &CompanyData{}

	// 1. Get company basic info from DART (may fail for non-listed companies)
	dartInfo, err := s.dartCrawler.SearchCompany(companyName)
	if err != nil {
		// Non-listed company - gracefully handle
		dartInfo = &crawler.CompanyInfo{
			CorpName: companyName,
		}
	}
	result.BasicInfo = dartInfo

	// 2. Get recent news from Naver
	news, err := s.newsCrawler.SearchCompanyNews(companyName, 5)
	if err != nil {
		// News fetch failed - continue with empty news
		news = []crawler.NewsArticle{}
	}
	result.News = news

	// 3. Future: Add Google search for talent profile hints

	return result, nil
}

// GetCompanyByJobPostingURL extracts company name from job posting and fetches data
func (s *CompanyDataService) GetCompanyByJobPostingURL(ctx context.Context, jobPostingURL string, companyName string) (*CompanyData, error) {
	if companyName == "" {
		return nil, fmt.Errorf("company name is required")
	}

	return s.GetCompanyData(ctx, companyName)
}
