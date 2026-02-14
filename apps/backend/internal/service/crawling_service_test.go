package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMForCrawling mocks AI for crawling tests
type MockLLMForCrawling struct {
	response ai.LLMResponse
	err      error
	calls    int // track number of calls
}

func (m *MockLLMForCrawling) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	return m.response, m.err
}

func TestNormalizeWithAI_Success(t *testing.T) {
	mockLLM := &MockLLMForCrawling{
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

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(aiProvider)

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

	result, err := svc.normalizeWithAI(context.Background(), rawPosting)
	require.NoError(t, err)
	assert.Equal(t, "테스트 회사", result.CompanyName)
	assert.Equal(t, "백엔드 개발자", result.Position)
	assert.Equal(t, "개발팀", result.Department)
	assert.Equal(t, "정규직", result.JobType)
	assert.Len(t, result.MainTasks, 2)
	assert.Contains(t, result.RequiredSkills, "Go")
}

func TestNormalizeJobPosting_RequiredFieldsMissing(t *testing.T) {
	// AI returns JSON missing required fields (empty company_name and position)
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "",
				"position": "",
				"main_tasks": [],
				"requirements": []
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(aiProvider)

	rawPosting := &crawler.RawJobPosting{
		Source:    "unknown",
		SourceURL: "https://example.com/job/1",
		RawHTML:   "<html><body>some job posting</body></html>",
	}

	result, err := svc.normalizeWithAI(context.Background(), rawPosting)
	// Current implementation does NOT validate required fields —
	// it just unmarshals whatever AI returns. Verify this behavior.
	require.NoError(t, err)
	assert.Equal(t, "", result.CompanyName)
	assert.Equal(t, "", result.Position)
}

func TestNormalizeJobPosting_InvalidJSON(t *testing.T) {
	// AI returns invalid JSON
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `not valid json at all`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(aiProvider)

	rawPosting := &crawler.RawJobPosting{
		Source:      "jobkorea",
		SourceURL:   "https://www.jobkorea.co.kr/Recruit/GI_Read/12345",
		CompanyName: "테스트",
		Position:    "개발자",
	}

	_, err := svc.normalizeWithAI(context.Background(), rawPosting)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse AI response")
}

func TestNormalizeJobPosting_AICallFailure(t *testing.T) {
	mockLLM := &MockLLMForCrawling{
		err: fmt.Errorf("rate limit exceeded"),
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(aiProvider)

	rawPosting := &crawler.RawJobPosting{
		Source:      "jobkorea",
		SourceURL:   "https://www.jobkorea.co.kr/Recruit/GI_Read/12345",
		CompanyName: "테스트",
		Position:    "개발자",
	}

	_, err := svc.normalizeWithAI(context.Background(), rawPosting)
	require.Error(t, err)
}

func TestDomainRouter_SelectsCorrectParser(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool // true = matches domain
		domain   string
	}{
		{"jobkorea URL", "https://www.jobkorea.co.kr/Recruit/GI_Read/12345", true, "jobkorea.co.kr"},
		{"catch URL", "https://www.catch.co.kr/NCS/RecruitInfoDetail/12345", true, "catch.co.kr"},
		{"unknown URL", "https://www.example.com/jobs/12345", false, "jobkorea.co.kr"},
		{"unknown URL for catch", "https://www.example.com/jobs/12345", false, "catch.co.kr"},
		{"empty URL", "", false, "jobkorea.co.kr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDomain(tt.url, tt.domain)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCrawlJobPosting_DomainRouting(t *testing.T) {
	// Verify that CrawlJobPosting routes to correct parser based on URL domain
	normalizedJSON := `{
		"company_name": "테스트",
		"position": "개발자",
		"main_tasks": ["개발"],
		"requirements": ["경력"]
	}`

	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{Content: normalizedJSON},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(aiProvider)

	// We can't test full CrawlJobPosting without HTTP calls,
	// but we can verify the normalization output structure
	rawPosting := &crawler.RawJobPosting{
		Source:      "jobkorea",
		SourceURL:   "https://www.jobkorea.co.kr/Recruit/GI_Read/12345",
		CompanyName: "테스트",
		Position:    "개발자",
	}

	result, err := svc.normalizeWithAI(context.Background(), rawPosting)
	require.NoError(t, err)

	// Verify output is a valid JobPosting
	data, err := json.Marshal(result)
	require.NoError(t, err)

	var parsed crawler.JobPosting
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "테스트", parsed.CompanyName)
}
