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
