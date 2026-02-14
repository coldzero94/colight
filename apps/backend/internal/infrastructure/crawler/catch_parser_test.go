package crawler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCatch_CompanyName(t *testing.T) {
	html := `
		<html>
			<div class="company-name">캐치 테스트 회사</div>
			<h1 class="title">프론트엔드 개발자</h1>
		</html>
	`

	parser := NewCatchParser()
	result, err := parser.ParseHTML("https://www.catch.co.kr/NCS/RecruitInfoDetail/12345", html)

	require.NoError(t, err)
	assert.Equal(t, "캐치 테스트 회사", result.CompanyName)
	assert.Equal(t, "프론트엔드 개발자", result.Position)
	assert.Equal(t, "catch", result.Source)
}

func TestParseCatch_Skills(t *testing.T) {
	html := `
		<html>
			<div class="company-name">테스트</div>
			<div class="tech-stack">
				<span>React</span>
				<span>TypeScript</span>
			</div>
		</html>
	`

	parser := NewCatchParser()
	result, err := parser.ParseHTML("https://www.catch.co.kr/NCS/RecruitInfoDetail/12345", html)

	require.NoError(t, err)
	assert.Len(t, result.Skills, 2)
	assert.Contains(t, result.Skills, "React")
	assert.Contains(t, result.Skills, "TypeScript")
}

func TestParseCatch_FullPosting(t *testing.T) {
	htmlBytes, err := os.ReadFile(testdataPath("catch_full.html"))
	require.NoError(t, err)

	parser := NewCatchParser()
	result, err := parser.ParseHTML("https://www.catch.co.kr/NCS/RecruitInfoDetail/99999", string(htmlBytes))

	require.NoError(t, err)
	assert.Equal(t, "catch", result.Source)
	assert.Equal(t, "https://www.catch.co.kr/NCS/RecruitInfoDetail/99999", result.SourceURL)

	// Company & position
	assert.Equal(t, "카카오", result.CompanyName)
	assert.Equal(t, "프론트엔드 개발자", result.Position)

	// Detail fields
	assert.Equal(t, "FE플랫폼팀", result.Department)
	assert.Equal(t, "경력 3년 이상", result.Career)
	assert.Equal(t, "경기 성남시 판교", result.Location)
	assert.Equal(t, "2026-07-15", result.Deadline)

	// Job description sections
	assert.Contains(t, result.MainTasks, "React/Next.js")
	assert.Contains(t, result.Requirements, "React 3년 이상")
	assert.Contains(t, result.Preferred, "Next.js App Router")

	// Skills
	assert.Len(t, result.Skills, 3)
	assert.Contains(t, result.Skills, "React")
	assert.Contains(t, result.Skills, "TypeScript")
	assert.Contains(t, result.Skills, "Next.js")
}

func TestParseCatch_RawJobPostingStructure(t *testing.T) {
	html := `
		<html>
			<div class="company-name">구조 확인 회사</div>
			<h1 class="title">QA 엔지니어</h1>
			<div class="career">신입</div>
			<div class="location">서울 강남구</div>
		</html>
	`

	parser := NewCatchParser()
	result, err := parser.ParseHTML("https://www.catch.co.kr/NCS/RecruitInfoDetail/12345", html)

	require.NoError(t, err)

	// Verify the returned struct matches RawJobPosting structure
	assert.IsType(t, &RawJobPosting{}, result)
	assert.Equal(t, "catch", result.Source)
	assert.Equal(t, "구조 확인 회사", result.CompanyName)
	assert.Equal(t, "QA 엔지니어", result.Position)
	assert.Equal(t, "신입", result.Career)
	assert.Equal(t, "서울 강남구", result.Location)
	assert.NotNil(t, result.Skills) // Should be initialized (empty slice, not nil)
}

func TestParseCatch_MissingSections(t *testing.T) {
	html := `
		<html>
			<div class="company-name">캐치 회사</div>
		</html>
	`

	parser := NewCatchParser()
	result, err := parser.ParseHTML("https://www.catch.co.kr/NCS/RecruitInfoDetail/12345", html)

	require.NoError(t, err)
	assert.Equal(t, "캐치 회사", result.CompanyName)
	assert.Equal(t, "", result.Position)
	assert.Equal(t, "", result.Department)
	assert.Equal(t, "", result.Career)
	assert.Equal(t, "", result.MainTasks)
	assert.Equal(t, "", result.Requirements)
	assert.Empty(t, result.Skills)
}

func TestParseCatch_SelectorMismatch(t *testing.T) {
	// HTML with completely different structure — no matching selectors
	html := `
		<html>
			<div class="unknown-structure">
				<p id="corp">Some Company</p>
				<h2 id="job-title">Some Job</h2>
			</div>
		</html>
	`

	parser := NewCatchParser()
	result, err := parser.ParseHTML("https://www.catch.co.kr/NCS/RecruitInfoDetail/12345", html)

	// Should NOT error — gracefully returns empty values
	require.NoError(t, err)
	assert.Equal(t, "", result.CompanyName)
	assert.Equal(t, "", result.Position)
	assert.Equal(t, "", result.Department)
	assert.Equal(t, "", result.MainTasks)
	assert.Empty(t, result.Skills)
	assert.Equal(t, "catch", result.Source)
}
