package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFromHTML_JSONLD(t *testing.T) {
	html := `<html><head>
		<script type="application/ld+json">
		{
			"@type": "JobPosting",
			"title": "[NAVER Cloud] DevOps 엔지니어 (경력)",
			"hiringOrganization": {"name": "NAVER Cloud"},
			"jobLocation": {"address": {"streetAddress": "경기도 성남시 분당구 판교"}},
			"employmentType": "FULL_TIME",
			"experienceRequirements": "경력"
		}
		</script>
	</head><body></body></html>`

	result := ExtractFromHTML(html, "https://recruit.navercorp.com/rcrt/view.do?annoId=30004542")
	require.NotNil(t, result)
	assert.Equal(t, "NAVER Cloud", result.Raw.CompanyName)
	assert.Equal(t, "[NAVER Cloud] DevOps 엔지니어 (경력)", result.Raw.Position)
	assert.Equal(t, "경기도 성남시 분당구 판교", result.Raw.Location)
	assert.Equal(t, "경력", result.Raw.Career)
	assert.Equal(t, "FULL_TIME", result.Raw.JobType)
}

func TestExtractFromHTML_JSONLD_ArrayEmploymentType(t *testing.T) {
	// Naver uses ["FULL_TIME"] (array) instead of "FULL_TIME" (string)
	html := `<html><head>
		<script type="application/ld+json">
		{
			"@type": "JobPosting",
			"title": "DevOps 엔지니어",
			"hiringOrganization": {"name": "NAVER Cloud"},
			"employmentType": ["FULL_TIME"]
		}
		</script>
	</head><body></body></html>`

	result := ExtractFromHTML(html, "https://recruit.navercorp.com")
	require.NotNil(t, result)
	assert.Equal(t, "NAVER Cloud", result.Raw.CompanyName)
	assert.Equal(t, "DevOps 엔지니어", result.Raw.Position)
	assert.Equal(t, "FULL_TIME", result.Raw.JobType)
}

func TestExtractFromHTML_JSONLD_NonJobPosting(t *testing.T) {
	// Non-JobPosting JSON-LD should be ignored
	html := `<html><head>
		<script type="application/ld+json">
		{"@type": "Organization", "name": "SomeCorp"}
		</script>
	</head><body></body></html>`

	result := ExtractFromHTML(html, "https://example.com")
	assert.Equal(t, "", result.Raw.CompanyName)
}

func TestExtractFromHTML_MetaTags(t *testing.T) {
	html := `<html>
		<head>
			<meta property="og:title" content="백엔드 개발자 - 네이버">
			<meta property="og:site_name" content="네이버 채용">
		</head>
		<body><div>content</div></body>
	</html>`

	result := ExtractFromHTML(html, "https://example.com/job/123")
	require.NotNil(t, result)
	// og:title typically contains position and company
	assert.NotEmpty(t, result.Raw.Position)
}

func TestExtractFromHTML_TitleParsing(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantPos  string
		wantComp string
	}{
		{
			name:     "position - company | site",
			html:     `<html><head><title>백엔드 개발자 - 네이버 | 잡코리아</title></head><body></body></html>`,
			wantPos:  "백엔드 개발자",
			wantComp: "네이버",
		},
		{
			name:     "[company] position",
			html:     `<html><head><title>[카카오] 프론트엔드 개발자</title></head><body></body></html>`,
			wantPos:  "프론트엔드 개발자",
			wantComp: "카카오",
		},
		{
			name:     "company 채용: position",
			html:     `<html><head><title>LG전자 채용: AI 엔지니어</title></head><body></body></html>`,
			wantPos:  "AI 엔지니어",
			wantComp: "LG전자",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFromHTML(tt.html, "https://example.com/job")
			require.NotNil(t, result)

			if tt.wantPos != "" {
				assert.Contains(t, result.Raw.Position, tt.wantPos, "position mismatch")
			}
			if tt.wantComp != "" {
				assert.Contains(t, result.Raw.CompanyName, tt.wantComp, "company mismatch")
			}
		})
	}
}

func TestExtractFromHTML_KoreanSections(t *testing.T) {
	html := `<html><body>
		<h3>담당업무</h3>
		<ul>
			<li>백엔드 API 설계 및 개발</li>
			<li>MSA 아키텍처 설계</li>
			<li>코드 리뷰 및 기술 멘토링</li>
		</ul>
		<h3>자격요건</h3>
		<ul>
			<li>Go 또는 Java 5년 이상 경력</li>
			<li>대규모 트래픽 경험</li>
		</ul>
		<h3>우대사항</h3>
		<ul>
			<li>Kubernetes 운영 경험</li>
			<li>오픈소스 기여 경험</li>
		</ul>
	</body></html>`

	result := ExtractFromHTML(html, "https://example.com/job")
	require.NotNil(t, result)

	assert.Contains(t, result.SectionTexts["main_tasks"], "백엔드 API 설계")
	assert.Contains(t, result.SectionTexts["main_tasks"], "MSA 아키텍처")
	assert.Contains(t, result.SectionTexts["requirements"], "Go 또는 Java")
	assert.Contains(t, result.SectionTexts["preferred"], "Kubernetes 운영")
}

func TestExtractFromHTML_SectionContentRich(t *testing.T) {
	html := `<html><body>
		<h3>담당업무</h3>
		<ul>
			<li>백엔드 API 설계 및 개발</li>
			<li>MSA 아키텍처 설계</li>
		</ul>
	</body></html>`

	result := ExtractFromHTML(html, "https://example.com/job")
	require.NotNil(t, result)

	// Verify the section preserves list structure (markdown bullets)
	section := result.SectionTexts["main_tasks"]
	assert.Contains(t, section, "백엔드 API 설계 및 개발")
	assert.Contains(t, section, "MSA 아키텍처 설계")
}

func TestExtractFromHTML_MetadataTable(t *testing.T) {
	html := `<html><body>
		<table>
			<tr><th>고용형태</th><td>정규직</td></tr>
			<tr><th>경력</th><td>경력 3년 이상</td></tr>
			<tr><th>학력</th><td>학사 이상</td></tr>
			<tr><th>근무지</th><td>서울 강남구</td></tr>
			<tr><th>마감일</th><td>2026-06-30</td></tr>
		</table>
	</body></html>`

	result := ExtractFromHTML(html, "https://example.com/job")
	require.NotNil(t, result)

	assert.Equal(t, "정규직", result.Raw.JobType)
	assert.Equal(t, "경력 3년 이상", result.Raw.Career)
	assert.Equal(t, "학사 이상", result.Raw.Education)
	assert.Equal(t, "서울 강남구", result.Raw.Location)
	assert.Equal(t, "2026-06-30", result.Raw.Deadline)
}

func TestExtractFromHTML_MetadataDL(t *testing.T) {
	html := `<html><body>
		<dl>
			<dt>고용형태</dt><dd>계약직</dd>
			<dt>근무지역</dt><dd>판교</dd>
			<dt>급여</dt><dd>협의 후 결정</dd>
		</dl>
	</body></html>`

	result := ExtractFromHTML(html, "https://example.com/job")
	require.NotNil(t, result)

	assert.Equal(t, "계약직", result.Raw.JobType)
	assert.Equal(t, "판교", result.Raw.Location)
	assert.Equal(t, "협의 후 결정", result.Raw.Salary)
}

func TestExtractFromHTML_SkillTags(t *testing.T) {
	html := `<html><body>
		<div class="skill-tag">Go</div>
		<div class="skill-tag">PostgreSQL</div>
		<div class="tech-stack"><span>Docker</span><span>Kubernetes</span></div>
	</body></html>`

	result := ExtractFromHTML(html, "https://example.com/job")
	require.NotNil(t, result)

	assert.Contains(t, result.Raw.Skills, "Go")
	assert.Contains(t, result.Raw.Skills, "PostgreSQL")
	assert.Contains(t, result.Raw.Skills, "Docker")
	assert.Contains(t, result.Raw.Skills, "Kubernetes")
}

func TestExtractFromHTML_NoSections(t *testing.T) {
	html := `<html><body><p>Some random content without any sections.</p></body></html>`

	result := ExtractFromHTML(html, "https://example.com/job")
	require.NotNil(t, result)

	assert.Empty(t, result.SectionTexts)
	assert.Equal(t, 0, result.FilledFields)
}

func TestExtractFromMarkdown_Sections(t *testing.T) {
	md := `## 담당업무
- 백엔드 API 설계 및 개발
- MSA 아키텍처 설계

## 자격요건
- Go 또는 Java 5년 이상 경력
- 대규모 트래픽 경험

## 우대사항
- Kubernetes 운영 경험
`

	result := ExtractFromMarkdown(md)
	require.NotNil(t, result)

	assert.Contains(t, result.SectionTexts["main_tasks"], "백엔드 API 설계")
	assert.Contains(t, result.SectionTexts["requirements"], "Go 또는 Java")
	assert.Contains(t, result.SectionTexts["preferred"], "Kubernetes 운영")
}

func TestExtractFromMarkdown_BulletItems(t *testing.T) {
	md := `### 자격요건
- Go 언어 경험
- PostgreSQL 경험
- Docker 사용 경험

### 우대사항
* Kubernetes 운영 경험
* gRPC 경험
`

	result := ExtractFromMarkdown(md)
	require.NotNil(t, result)

	reqText := result.SectionTexts["requirements"]
	assert.Contains(t, reqText, "Go 언어 경험")
	assert.Contains(t, reqText, "PostgreSQL 경험")
	assert.Contains(t, reqText, "Docker 사용 경험")

	prefText := result.SectionTexts["preferred"]
	assert.Contains(t, prefText, "Kubernetes 운영 경험")
}

func TestExtractFromMarkdown_NoHeaders(t *testing.T) {
	md := `Some plain content without any section headers.
Just a paragraph of text about the job.`

	result := ExtractFromMarkdown(md)
	require.NotNil(t, result)
	assert.Empty(t, result.SectionTexts)
}

func TestIsEnoughForNormalize(t *testing.T) {
	tests := []struct {
		name   string
		result ExtractionResult
		want   bool
	}{
		{
			name: "company + position + main_tasks → true",
			result: ExtractionResult{
				Raw: RawJobPosting{CompanyName: "네이버", Position: "개발자", MainTasks: "API 개발"},
			},
			want: true,
		},
		{
			name: "company + requirements → true",
			result: ExtractionResult{
				Raw: RawJobPosting{CompanyName: "카카오", Requirements: "Go 3년"},
			},
			want: true,
		},
		{
			name: "position + main_tasks → true",
			result: ExtractionResult{
				Raw: RawJobPosting{Position: "백엔드 개발자", MainTasks: "서비스 개발"},
			},
			want: true,
		},
		{
			name: "only company → false",
			result: ExtractionResult{
				Raw: RawJobPosting{CompanyName: "네이버"},
			},
			want: false,
		},
		{
			name: "empty → false",
			result: ExtractionResult{
				Raw: RawJobPosting{},
			},
			want: false,
		},
		{
			name: "only main_tasks → false",
			result: ExtractionResult{
				Raw: RawJobPosting{MainTasks: "API 개발"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.IsEnoughForNormalize())
		})
	}
}

func TestBuildNormalizeInput(t *testing.T) {
	result := &ExtractionResult{
		Raw: RawJobPosting{
			CompanyName: "네이버",
			Position:    "백엔드 개발자",
			Career:      "경력 3년",
			Location:    "판교",
		},
		SectionTexts: map[string]string{
			"main_tasks":   "- API 설계 및 개발\n- MSA 아키텍처",
			"requirements": "- Go 5년 이상\n- PostgreSQL 경험",
			"preferred":    "- Kubernetes 운영 경험",
		},
	}

	input := BuildNormalizeInput(result)

	assert.Contains(t, input, "네이버")
	assert.Contains(t, input, "백엔드 개발자")
	assert.Contains(t, input, "API 설계 및 개발")
	assert.Contains(t, input, "Go 5년 이상")
	assert.Contains(t, input, "Kubernetes 운영 경험")
}

func TestSmartTrimForLLM_SectionPriority(t *testing.T) {
	extracted := &ExtractionResult{
		SectionTexts: map[string]string{
			"main_tasks":   "- API 개발\n- 시스템 설계",
			"requirements": "- Go 경험\n- DB 경험",
			"preferred":    "- K8s 경험",
			"benefits":     "- 자율 출퇴근\n- 식비 지원",
		},
	}

	md := "some long markdown content that would normally be truncated..."
	result := SmartTrimForLLM(md, extracted, 500)

	// High-priority sections should be included
	assert.Contains(t, result, "API 개발")
	assert.Contains(t, result, "Go 경험")
	assert.Contains(t, result, "K8s 경험")
}

func TestSmartTrimForLLM_Fallback(t *testing.T) {
	// No sections extracted — falls back to truncation
	extracted := &ExtractionResult{
		SectionTexts: map[string]string{},
	}

	md := "a short markdown"
	result := SmartTrimForLLM(md, extracted, 500)

	assert.Equal(t, md, result)
}

func TestMergeExtractions(t *testing.T) {
	mdResult := &ExtractionResult{
		Raw: RawJobPosting{
			CompanyName: "네이버",
			Position:    "개발자",
			MainTasks:   "- API 개발",
		},
		SectionTexts: map[string]string{
			"main_tasks": "- API 개발",
		},
	}

	htmlResult := &ExtractionResult{
		Raw: RawJobPosting{
			CompanyName:  "네이버",
			Position:     "백엔드 개발자", // longer/richer
			MainTasks:    "- API 설계 및 개발\n- MSA 아키텍처 설계", // richer
			Requirements: "- Go 5년 이상",
		},
		SectionTexts: map[string]string{
			"main_tasks":   "- API 설계 및 개발\n- MSA 아키텍처 설계",
			"requirements": "- Go 5년 이상",
		},
	}

	merged := MergeExtractions(mdResult, htmlResult)
	require.NotNil(t, merged)

	// Should pick richer values
	assert.Equal(t, "백엔드 개발자", merged.Raw.Position)
	assert.Contains(t, merged.Raw.MainTasks, "MSA 아키텍처")
	assert.Equal(t, "- Go 5년 이상", merged.Raw.Requirements)

	// SectionTexts should be merged
	assert.Contains(t, merged.SectionTexts["main_tasks"], "MSA 아키텍처")
	assert.Contains(t, merged.SectionTexts["requirements"], "Go 5년 이상")
}
