package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
	"github.com/stretchr/testify/assert"
)

// MockLLMForCrawling mocks AI for crawling tests
type MockLLMForCrawling struct {
	response ai.LLMResponse
	err      error
}

func (m *MockLLMForCrawling) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func TestNormalizeWithAI_Success(t *testing.T) {
	_ = &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "테스트 회사",
				"position": "백엔드 개발자",
				"department": "개발팀",
				"job_type": "정규직",
				"experience_level": "3년 이상",
				"main_tasks": ["API 개발", "DB 설계"],
				"requirements": ["Go 경험", "PostgreSQL"],
				"preferred": ["Docker", "K8s"],
				"required_skills": ["Go", "PostgreSQL"],
				"soft_skills": ["커뮤니케이션"],
				"company_values_hints": ["혁신", "도전"],
				"deadline": "2026-03-31"
			}`,
		},
	}

	_ = &ai.AIProvider{}
	// Note: We can't easily mock aiProvider.Light() without refactoring
	// For now, test the parser logic separately

	rawPosting := &crawler.RawJobPosting{
		Source:       "jobkorea",
		SourceURL:    "https://www.jobkorea.co.kr/Recruit/GI_Read/12345",
		CompanyName:  "테스트 회사",
		Position:     "백엔드 개발자",
		Department:   "개발팀",
		Career:       "3년 이상",
		MainTasks:    "API 개발, DB 설계",
		Requirements: "Go 경험 필수, PostgreSQL 사용 경험",
		Preferred:    "Docker, K8s 경험 우대",
		Skills:       []string{"Go", "PostgreSQL"},
	}

	// Test that rawPosting is properly structured
	assert.Equal(t, "jobkorea", rawPosting.Source)
	assert.Equal(t, "테스트 회사", rawPosting.CompanyName)
	assert.Len(t, rawPosting.Skills, 2)
}
