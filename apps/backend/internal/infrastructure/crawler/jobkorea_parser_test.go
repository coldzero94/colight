package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
