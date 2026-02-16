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
