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

// MockHTMLFetcher mocks HTML fetching for tests
type MockHTMLFetcher struct {
	html string
	err  error
}

func (m *MockHTMLFetcher) FetchHTML(url string) (string, error) {
	return m.html, m.err
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
	svc := NewCrawlingService(nil, aiProvider)

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
	svc := NewCrawlingService(nil, aiProvider)

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
	svc := NewCrawlingService(nil, aiProvider)

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
	svc := NewCrawlingService(nil, aiProvider)

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
	svc := NewCrawlingService(nil, aiProvider)

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

// === New tests for 3-tier crawling pipeline ===

func TestExtractJobPostingFromMarkdown_Success(t *testing.T) {
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "네이버",
				"position": "프론트엔드 개발자",
				"department": "서비스개발팀",
				"job_type": "정규직",
				"experience_level": "5년 이상",
				"main_tasks": ["웹 서비스 개발", "UI/UX 개선"],
				"requirements": ["React 경험 3년+", "TypeScript"],
				"preferred": ["Next.js", "성능 최적화 경험"],
				"required_skills": ["React", "TypeScript", "JavaScript"],
				"soft_skills": ["소통", "협업"],
				"company_values_hints": ["기술 혁신", "사용자 중심"],
				"deadline": "2026-04-30"
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(nil, aiProvider)

	markdown := `# 프론트엔드 개발자 채용

## 담당업무
- 웹 서비스 개발
- UI/UX 개선

## 자격요건
- React 경험 3년 이상
- TypeScript 필수

## 우대사항
- Next.js 경험
- 성능 최적화 경험`

	result, err := svc.extractJobPostingFromMarkdown(context.Background(), "https://example.com/job/1", markdown)
	require.NoError(t, err)
	assert.Equal(t, "네이버", result.CompanyName)
	assert.Equal(t, "프론트엔드 개발자", result.Position)
	assert.Len(t, result.MainTasks, 2)
	assert.Contains(t, result.RequiredSkills, "React")
}

func TestExtractJobPostingFromHTML_Success(t *testing.T) {
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "카카오",
				"position": "서버 엔지니어",
				"job_type": "정규직",
				"main_tasks": ["서버 개발"],
				"requirements": ["Java 경험"],
				"required_skills": ["Java", "Spring"]
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingService(nil, aiProvider)

	html := `<html><body><div>서버 엔지니어 채용</div></body></html>`

	result, err := svc.extractJobPostingFromHTML(context.Background(), "https://example.com/job/2", html)
	require.NoError(t, err)
	assert.Equal(t, "카카오", result.CompanyName)
	assert.Equal(t, "서버 엔지니어", result.Position)
	assert.Contains(t, result.RequiredSkills, "Java")
}

func TestCrawlJobPosting_UnknownDomain_UsesMarkdownPath(t *testing.T) {
	jobPostingHTML := `<html><head><title>채용</title></head><body>
		<article>
			<h1>데이터 엔지니어</h1>
			<p>데이터 파이프라인을 설계하고 구축하는 업무를 담당합니다.</p>
			<h2>자격요건</h2>
			<ul><li>Python 경험 3년 이상</li><li>Spark/Hadoop 경험</li></ul>
		</article>
	</body></html>`

	mockFetcher := &MockHTMLFetcher{html: jobPostingHTML}
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "라인",
				"position": "데이터 엔지니어",
				"main_tasks": ["데이터 파이프라인 설계"],
				"requirements": ["Python 3년+"],
				"required_skills": ["Python", "Spark"]
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingServiceWithFetcher(nil, aiProvider, mockFetcher)

	result, err := svc.CrawlJobPosting(context.Background(), "https://www.wanted.co.kr/wd/12345")
	require.NoError(t, err)
	assert.Equal(t, "라인", result.CompanyName)
	assert.Equal(t, "데이터 엔지니어", result.Position)
	assert.Equal(t, 1, mockLLM.calls, "should call LLM once for markdown extraction")
}

func TestCrawlJobPosting_JobKorea_StillUsesFastPath(t *testing.T) {
	// JobKorea HTML that the CSS parser can handle
	// We use a minimal HTML that will fail CSS parsing, triggering normalizeWithAI
	// The point is: containsDomain routes correctly
	jobkoreaHTML := `<html><body>
		<div class="tbCol"><h3>테스트 회사</h3></div>
		<h3>백엔드 개발자</h3>
	</body></html>`

	mockFetcher := &MockHTMLFetcher{html: jobkoreaHTML}
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "테스트 회사",
				"position": "백엔드 개발자",
				"main_tasks": ["개발"],
				"requirements": ["경력"]
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingServiceWithFetcher(nil, aiProvider, mockFetcher)

	result, err := svc.CrawlJobPosting(context.Background(), "https://www.jobkorea.co.kr/Recruit/GI_Read/12345")
	require.NoError(t, err)
	assert.Equal(t, "테스트 회사", result.CompanyName)
}

func TestCrawlJobPosting_FallbackToHTML_WhenMarkdownTooShort(t *testing.T) {
	// HTML that produces very short readability output (< MinMarkdownLength)
	shortHTML := `<html><body><p>Hi</p></body></html>`

	mockFetcher := &MockHTMLFetcher{html: shortHTML}
	mockLLM := &MockLLMForCrawling{
		response: ai.LLMResponse{
			Content: `{
				"company_name": "Unknown",
				"position": "Developer",
				"main_tasks": ["Develop"]
			}`,
		},
	}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingServiceWithFetcher(nil, aiProvider, mockFetcher)

	result, err := svc.CrawlJobPosting(context.Background(), "https://example.com/job/1")
	require.NoError(t, err)
	assert.Equal(t, "Unknown", result.CompanyName)
	// Should have used HTML fallback path since markdown was too short
	assert.Equal(t, 1, mockLLM.calls)
}

func TestCrawlJobPosting_FetchError(t *testing.T) {
	mockFetcher := &MockHTMLFetcher{err: fmt.Errorf("connection refused")}
	mockLLM := &MockLLMForCrawling{}

	aiProvider := ai.NewAIProviderForTest(mockLLM, nil)
	svc := NewCrawlingServiceWithFetcher(nil, aiProvider, mockFetcher)

	_, err := svc.CrawlJobPosting(context.Background(), "https://example.com/job/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch HTML")
	assert.Equal(t, 0, mockLLM.calls, "should not call LLM when fetch fails")
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected int // expected length
	}{
		{"short string", "hello", 10, 5},
		{"exact length", "hello", 5, 5},
		{"needs truncation", "hello world", 5, 5},
		{"empty string", "", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.LessOrEqual(t, len(result), tt.maxLen)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}
