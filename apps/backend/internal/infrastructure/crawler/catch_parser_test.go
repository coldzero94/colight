package crawler

import (
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
