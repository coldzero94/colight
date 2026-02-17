package crawler

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "crawling", name)
}

func TestParseJobKorea_CompanyName(t *testing.T) {
	html := `
		<html>
			<div class="coName">테스트 회사</div>
			<h1 class="artReadJobTitle">백엔드 개발자</h1>
		</html>
	`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/12345", html)

	require.NoError(t, err)
	assert.Equal(t, "테스트 회사", result.CompanyName)
	assert.Equal(t, "백엔드 개발자", result.Position)
	assert.Equal(t, "jobkorea", result.Source)
}

func TestParseJobKorea_MissingSections(t *testing.T) {
	html := `
		<html>
			<div class="coName">테스트 회사</div>
		</html>
	`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/12345", html)

	require.NoError(t, err)
	assert.Equal(t, "테스트 회사", result.CompanyName)
	assert.Equal(t, "", result.Position) // Missing - should be empty
}

func TestParseJobKorea_Skills(t *testing.T) {
	html := `
		<html>
			<div class="coName">테스트 회사</div>
			<div class="skillWrap">
				<span class="skill">Go</span>
				<span class="skill">PostgreSQL</span>
				<span class="skill">Docker</span>
			</div>
		</html>
	`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/12345", html)

	require.NoError(t, err)
	assert.Len(t, result.Skills, 3)
	assert.Contains(t, result.Skills, "Go")
	assert.Contains(t, result.Skills, "PostgreSQL")
	assert.Contains(t, result.Skills, "Docker")
}

func TestParseJobKorea_FullPosting(t *testing.T) {
	htmlBytes, err := os.ReadFile(testdataPath("jobkorea_full.html"))
	require.NoError(t, err)

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/99999", string(htmlBytes))

	require.NoError(t, err)
	assert.Equal(t, "jobkorea", result.Source)
	assert.Equal(t, "https://www.jobkorea.co.kr/Recruit/GI_Read/99999", result.SourceURL)

	// Company & position
	assert.Equal(t, "네이버", result.CompanyName)
	assert.Equal(t, "시니어 백엔드 개발자", result.Position)

	// Table row fields
	assert.Equal(t, "서비스개발팀", result.Department)
	assert.Equal(t, "경력 5년 이상", result.Career)
	assert.Equal(t, "학사 이상", result.Education)
	assert.Equal(t, "정규직", result.JobType)
	assert.Equal(t, "회사 내규에 따름", result.Salary)
	assert.Equal(t, "경기 성남시 분당구", result.Location)
	assert.Equal(t, "2026-06-30", result.Deadline)

	// Job description sections
	assert.Contains(t, result.MainTasks, "백엔드 API 설계")
	assert.Contains(t, result.Requirements, "Go 또는 Java")
	assert.Contains(t, result.Preferred, "Kubernetes 운영 경험")

	// Skills
	assert.Len(t, result.Skills, 5)
	assert.Contains(t, result.Skills, "Go")
	assert.Contains(t, result.Skills, "gRPC")
}

func TestParseJobKorea_SelectorMismatch(t *testing.T) {
	// HTML with completely different structure — no matching selectors
	html := `
		<html>
			<div class="unknown-structure">
				<p id="corp">Some Company</p>
				<h2 id="job-title">Some Job</h2>
			</div>
		</html>
	`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/12345", html)

	// Should NOT error — gracefully returns empty values
	require.NoError(t, err)
	assert.Equal(t, "", result.CompanyName)
	assert.Equal(t, "", result.Position)
	assert.Equal(t, "", result.Department)
	assert.Equal(t, "", result.MainTasks)
	assert.Empty(t, result.Skills)
	assert.Equal(t, "jobkorea", result.Source)
}

func TestParseJobKorea_JSONLD(t *testing.T) {
	html := `<html><head>
		<script type="application/ld+json">
		{
			"@type": "JobPosting",
			"title": "[신입/경력] 백엔드 개발자",
			"hiringOrganization": {"name": "㈜신지모루"},
			"jobLocation": {"address": {"addressLocality": "서울 마포구"}},
			"baseSalary": {"value": "연봉 3,700만원 이상"},
			"experienceRequirements": "신입"
		}
		</script>
	</head><body></body></html>`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/48607050", html)

	require.NoError(t, err)
	assert.Equal(t, "㈜신지모루", result.CompanyName)
	assert.Equal(t, "[신입/경력] 백엔드 개발자", result.Position)
	assert.Equal(t, "서울 마포구", result.Location)
	assert.Equal(t, "신입", result.Career)
	assert.Equal(t, "연봉 3,700만원 이상", result.Salary)
}

func TestParseJobKorea_SentryComponents(t *testing.T) {
	html := `<html><body>
		<div data-sentry-component="CompanyName"><h2>카카오엔터프라이즈</h2></div>
		<div data-sentry-component="TitleContent"><h1>시니어 Go 개발자</h1></div>
	</body></html>`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/12345", html)

	require.NoError(t, err)
	assert.Equal(t, "카카오엔터프라이즈", result.CompanyName)
	assert.Equal(t, "시니어 Go 개발자", result.Position)
}

func TestParseJobKorea_DescriptionURL(t *testing.T) {
	html := `<html><body>
		<div>some content</div>
		<script>var descUrl = "https://job-hub-files-prd-f1720431.s3.ap-northeast-2.amazonaws.com/jobhub-descriptions/abc123_DESCRIPTION.html?sig=xxx"</script>
	</body></html>`

	parser := NewJobKoreaParser()
	url := parser.DescriptionURL(html)
	assert.Contains(t, url, "job-hub-files")
	assert.Contains(t, url, "_DESCRIPTION.html")
}

func TestParseJobKorea_DescriptionURL_NotFound(t *testing.T) {
	html := `<html><body><div>no description url here</div></body></html>`

	parser := NewJobKoreaParser()
	url := parser.DescriptionURL(html)
	assert.Equal(t, "", url)
}

func TestParseJobKorea_ParseDescriptionHTML_Sections(t *testing.T) {
	descHTML := `<html><body>
		<h3>담당업무</h3>
		<ul>
			<li>백엔드 API 설계 및 개발</li>
			<li>MSA 아키텍처 설계</li>
		</ul>
		<h3>자격요건</h3>
		<ul>
			<li>Go 3년 이상</li>
			<li>PostgreSQL 경험</li>
		</ul>
		<h3>우대사항</h3>
		<ul>
			<li>Kubernetes 운영 경험</li>
		</ul>
	</body></html>`

	parser := NewJobKoreaParser()
	result := &RawJobPosting{Source: "jobkorea"}
	parser.ParseDescriptionHTML(result, descHTML)

	assert.Contains(t, result.MainTasks, "백엔드 API 설계")
	assert.Contains(t, result.MainTasks, "MSA 아키텍처")
	assert.Contains(t, result.Requirements, "Go 3년 이상")
	assert.Contains(t, result.Preferred, "Kubernetes 운영")
}

func TestParseJobKorea_ParseDescriptionHTML_FallbackSections(t *testing.T) {
	descHTML := `<html><body>
		<div class="artReadJobSecCont">개발 업무 수행</div>
		<div class="artReadJobSecCont">Java 경험 필수</div>
		<div class="artReadJobSecCont">AWS 경험 우대</div>
	</body></html>`

	parser := NewJobKoreaParser()
	result := &RawJobPosting{Source: "jobkorea"}
	parser.ParseDescriptionHTML(result, descHTML)

	assert.Contains(t, result.MainTasks, "개발 업무 수행")
	assert.Contains(t, result.Requirements, "Java 경험 필수")
	assert.Contains(t, result.Preferred, "AWS 경험 우대")
}

func TestParseJobKorea_JSONLDPrioritizedOverLegacy(t *testing.T) {
	// When JSON-LD is present, it should be used even if legacy selectors are also present
	html := `<html><head>
		<script type="application/ld+json">
		{
			"@type": "JobPosting",
			"title": "JSON-LD Position",
			"hiringOrganization": {"name": "JSON-LD Company"}
		}
		</script>
	</head><body>
		<div class="coName">Legacy Company</div>
		<h1 class="artReadJobTitle">Legacy Position</h1>
	</body></html>`

	parser := NewJobKoreaParser()
	result, err := parser.ParseHTML("https://www.jobkorea.co.kr/Recruit/GI_Read/12345", html)

	require.NoError(t, err)
	assert.Equal(t, "JSON-LD Company", result.CompanyName)
	assert.Equal(t, "JSON-LD Position", result.Position)
}
