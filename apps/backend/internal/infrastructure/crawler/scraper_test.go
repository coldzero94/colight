package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrapeToMarkdown_BasicHTML(t *testing.T) {
	html := `<html><head><title>Test</title></head><body>
		<article><h1>Software Engineer</h1><p>We are looking for a talented engineer.</p></article>
	</body></html>`

	md, err := ScrapeToMarkdown(html, "https://example.com/jobs/1")
	require.NoError(t, err)
	assert.Contains(t, md, "Software Engineer")
	assert.Contains(t, md, "talented engineer")
}

func TestScrapeToMarkdown_JobPostingHTML(t *testing.T) {
	html := `<html><head><title>채용공고</title></head><body>
		<div class="content">
			<h1>백엔드 개발자</h1>
			<h2>담당업무</h2>
			<ul>
				<li>API 서버 개발</li>
				<li>데이터베이스 설계</li>
			</ul>
			<h2>자격요건</h2>
			<ul>
				<li>Go 또는 Java 경험 3년 이상</li>
				<li>REST API 설계 경험</li>
			</ul>
			<h2>우대사항</h2>
			<ul>
				<li>Kubernetes 운영 경험</li>
			</ul>
		</div>
	</body></html>`

	md, err := ScrapeToMarkdown(html, "https://example.com/jobs/2")
	require.NoError(t, err)
	assert.Contains(t, md, "백엔드 개발자")
	assert.Contains(t, md, "API 서버 개발")
	assert.Contains(t, md, "Go 또는 Java")
	assert.Contains(t, md, "Kubernetes")
}

func TestScrapeToMarkdown_EmptyHTML(t *testing.T) {
	md, err := ScrapeToMarkdown("", "https://example.com")
	require.NoError(t, err)
	assert.Empty(t, md)
}

func TestScrapeToMarkdown_MinimalContent(t *testing.T) {
	html := `<html><body><p>Hi</p></body></html>`

	md, err := ScrapeToMarkdown(html, "https://example.com")
	require.NoError(t, err)
	// Should return something (possibly short)
	assert.NotNil(t, md)
}

func TestScrapeToMarkdown_RichHTML(t *testing.T) {
	html := `<html><head><title>Rich Content</title></head><body>
		<article>
			<h1>Senior Developer Position</h1>
			<table>
				<tr><th>Field</th><th>Value</th></tr>
				<tr><td>Location</td><td>Seoul</td></tr>
				<tr><td>Experience</td><td>5+ years</td></tr>
			</table>
			<h2>Requirements</h2>
			<ol>
				<li>Strong programming skills</li>
				<li>Team collaboration</li>
			</ol>
		</article>
	</body></html>`

	md, err := ScrapeToMarkdown(html, "https://example.com/jobs/3")
	require.NoError(t, err)
	assert.Contains(t, md, "Senior Developer Position")
	assert.Contains(t, md, "Seoul")
	assert.Contains(t, md, "Strong programming skills")
}

func TestScrapeToMarkdown_NavigationStripped(t *testing.T) {
	html := `<html><body>
		<nav><a href="/">Home</a><a href="/about">About</a></nav>
		<header><div class="logo">Company</div></header>
		<main>
			<h1>Job Opening</h1>
			<p>This is the actual job content that matters.</p>
		</main>
		<footer>Copyright 2024</footer>
	</body></html>`

	md, err := ScrapeToMarkdown(html, "https://example.com/jobs/4")
	require.NoError(t, err)
	assert.Contains(t, md, "Job Opening")
	assert.Contains(t, md, "actual job content")
}

func TestScrapeToMarkdown_NoisyHTML_StillExtractsContent(t *testing.T) {
	html := `<html><body>
		<nav><a href="/">홈</a><a href="/jobs">채용</a><a href="/my">마이페이지</a></nav>
		<header><div class="logo">사람인</div><div class="search-bar">검색</div></header>
		<aside class="sidebar"><p>인기 공고</p><p>광고 배너</p><p>추천 채용</p></aside>
		<main>
			<h1>시니어 백엔드 개발자</h1>
			<h2>담당업무</h2>
			<ul><li>API 서버 개발 및 운영</li><li>시스템 설계 및 아키텍처</li></ul>
			<h2>자격요건</h2>
			<ul><li>Go 경험 5년 이상</li><li>PostgreSQL 운영 경험</li></ul>
			<h2>우대사항</h2>
			<ul><li>Kubernetes 운영 경험</li><li>대규모 트래픽 처리 경험</li></ul>
		</main>
		<footer><p>Copyright 2026 사람인HR</p><p>이용약관 | 개인정보처리방침</p></footer>
		<script>var ga = 'tracking';</script>
	</body></html>`

	md, err := ScrapeToMarkdown(html, "https://www.saramin.co.kr/zf_user/jobs/view?rec_idx=12345")
	require.NoError(t, err)
	assert.Contains(t, md, "시니어 백엔드 개발자")
	assert.Contains(t, md, "API 서버 개발")
	assert.Contains(t, md, "Go 경험 5년")
	assert.Contains(t, md, "Kubernetes")
	assert.NotContains(t, md, "사람인")
	assert.NotContains(t, md, "인기 공고")
	assert.NotContains(t, md, "Copyright")
	assert.NotContains(t, md, "tracking")
}

func TestScrapeToMarkdown_FallbackToMainContent(t *testing.T) {
	// HTML where readability may fail but #content selector should work
	html := `<html><body>
		<div id="content">
			<h1>데이터 엔지니어 채용</h1>
			<p>Python 3년 이상 경험 필수. Spark 및 Kafka 경험을 보유한 분을 우대합니다.</p>
		</div>
	</body></html>`

	md, err := ScrapeToMarkdown(html, "https://example.com/job")
	require.NoError(t, err)
	assert.Contains(t, md, "데이터 엔지니어")
	assert.Contains(t, md, "Python")
}

func TestPostProcessMarkdown_CleansArtifacts(t *testing.T) {
	md := "# Job Title\n\n" +
		"[](https://example.com)\n\n" +
		"홈 > 채용 > 공고상세\n\n" +
		"Some actual content here.\n\n\n\n\n" +
		"More content with   extra   spaces.\n\n" +
		"https://tracking.example.com/pixel\n\n" +
		"![](https://example.com/spacer.gif)\n\n" +
		"Final content."

	cleaned := PostProcessMarkdown(md)
	assert.NotContains(t, cleaned, "[](https://example.com)")
	assert.NotContains(t, cleaned, "홈 > 채용 > 공고상세")
	assert.NotContains(t, cleaned, "tracking.example.com")
	assert.NotContains(t, cleaned, "![](https://example.com/spacer.gif)")
	assert.Contains(t, cleaned, "Some actual content")
	assert.Contains(t, cleaned, "More content with extra spaces.")
	assert.Contains(t, cleaned, "Final content.")
	// No triple+ newlines
	assert.NotContains(t, cleaned, "\n\n\n")
}

func TestPostProcessMarkdown_PreservesMarkdownHeaders(t *testing.T) {
	md := "# Title\n\n## Section One\n\nContent here.\n\n## Section Two\n\nMore content."
	cleaned := PostProcessMarkdown(md)
	assert.Contains(t, cleaned, "# Title")
	assert.Contains(t, cleaned, "## Section One")
	assert.Contains(t, cleaned, "## Section Two")
}

func TestPostProcessMarkdown_PreservesLinks(t *testing.T) {
	md := "[Apply Now](https://example.com/apply)\n\n[회사 소개](https://example.com/about)"
	cleaned := PostProcessMarkdown(md)
	assert.Contains(t, cleaned, "[Apply Now](https://example.com/apply)")
	assert.Contains(t, cleaned, "[회사 소개](https://example.com/about)")
}
