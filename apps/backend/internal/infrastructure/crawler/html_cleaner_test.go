package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- CleanHTML tests ---

func TestCleanHTML_RemovesNavHeaderFooter(t *testing.T) {
	html := `<html><body>
		<nav><a href="/">Home</a><a href="/jobs">Jobs</a></nav>
		<header><div class="logo">Corp Logo</div></header>
		<main><h1>Job Posting</h1><p>Content here</p></main>
		<footer><p>Copyright 2026</p></footer>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "Home")
	assert.NotContains(t, cleaned, "Corp Logo")
	assert.NotContains(t, cleaned, "Copyright 2026")
	assert.Contains(t, cleaned, "Job Posting")
	assert.Contains(t, cleaned, "Content here")
}

func TestCleanHTML_RemovesScriptsAndStyles(t *testing.T) {
	html := `<html><head>
		<style>.hidden { display: none; }</style>
		<link rel="stylesheet" href="/style.css">
	</head><body>
		<script>var tracking = true;</script>
		<noscript><p>Enable JavaScript</p></noscript>
		<iframe src="https://ad.example.com"></iframe>
		<div><h1>Real Content</h1></div>
		<script src="/analytics.js"></script>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "tracking")
	assert.NotContains(t, cleaned, "Enable JavaScript")
	assert.NotContains(t, cleaned, "ad.example.com")
	assert.NotContains(t, cleaned, "analytics.js")
	assert.Contains(t, cleaned, "Real Content")
}

func TestCleanHTML_RemovesAdsAndModals(t *testing.T) {
	html := `<html><body>
		<div class="ad">Buy Premium</div>
		<div class="ads">Sponsored</div>
		<div class="advertisement">Click here</div>
		<div class="modal">Login Required</div>
		<div class="popup">Subscribe Now</div>
		<div role="dialog">Alert</div>
		<article><h1>Job Description</h1><p>We are hiring.</p></article>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "Buy Premium")
	assert.NotContains(t, cleaned, "Sponsored")
	assert.NotContains(t, cleaned, "Click here")
	assert.NotContains(t, cleaned, "Login Required")
	assert.NotContains(t, cleaned, "Subscribe Now")
	assert.NotContains(t, cleaned, "Alert")
	assert.Contains(t, cleaned, "Job Description")
	assert.Contains(t, cleaned, "We are hiring.")
}

func TestCleanHTML_RemovesCookieBanners(t *testing.T) {
	html := `<html><body>
		<div class="cookie-banner">We use cookies</div>
		<div id="cookie-consent">Accept all cookies</div>
		<div class="gdpr-notice">GDPR Info</div>
		<div class="consent-popup">Privacy settings</div>
		<main><p>Important job details here</p></main>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "We use cookies")
	assert.NotContains(t, cleaned, "Accept all cookies")
	assert.NotContains(t, cleaned, "GDPR Info")
	assert.NotContains(t, cleaned, "Privacy settings")
	assert.Contains(t, cleaned, "Important job details")
}

func TestCleanHTML_RemovesKoreanSiteNoise(t *testing.T) {
	html := `<html><body>
		<div class="GnbWrap"><ul><li>채용</li><li>기업정보</li></ul></div>
		<div class="SideBarWrap"><p>인기 공고</p></div>
		<div class="BtnApply"><button>지원하기</button></div>
		<div class="BtnScrap"><button>스크랩</button></div>
		<div class="BtnShare"><button>공유</button></div>
		<div class="recruitLnb"><ul><li>상세</li><li>기업</li></ul></div>
		<div id="content_top_banner"><img src="banner.jpg"></div>
		<div id="recruit_aside"><p>채용 담당자</p></div>
		<div class="wrap_tit_recruit_company"><h2>회사 로고</h2></div>
		<div class="job-detail"><h1>백엔드 개발자</h1><p>Go 경험 3년</p></div>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "인기 공고")
	assert.NotContains(t, cleaned, "지원하기")
	assert.NotContains(t, cleaned, "스크랩")
	assert.NotContains(t, cleaned, "채용 담당자")
	assert.NotContains(t, cleaned, "회사 로고")
	assert.Contains(t, cleaned, "백엔드 개발자")
	assert.Contains(t, cleaned, "Go 경험 3년")
}

func TestCleanHTML_PreservesMainContent(t *testing.T) {
	html := `<html><body>
		<article>
			<h1>시니어 백엔드 개발자</h1>
			<h2>담당업무</h2>
			<ul><li>API 서버 개발</li><li>시스템 설계</li></ul>
			<h2>자격요건</h2>
			<ul><li>Go 5년 이상</li><li>PostgreSQL 경험</li></ul>
			<h2>우대사항</h2>
			<ul><li>Kubernetes 운영 경험</li></ul>
			<table>
				<tr><td>경력</td><td>5년 이상</td></tr>
				<tr><td>학력</td><td>학사 이상</td></tr>
			</table>
		</article>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.Contains(t, cleaned, "시니어 백엔드 개발자")
	assert.Contains(t, cleaned, "API 서버 개발")
	assert.Contains(t, cleaned, "Go 5년 이상")
	assert.Contains(t, cleaned, "Kubernetes")
	assert.Contains(t, cleaned, "5년 이상")
}

func TestCleanHTML_EmptyAndInvalidHTML(t *testing.T) {
	// Empty
	cleaned, err := CleanHTML("")
	require.NoError(t, err)
	assert.NotEmpty(t, cleaned) // goquery produces minimal HTML skeleton

	// Minimal
	cleaned, err = CleanHTML("<html><body></body></html>")
	require.NoError(t, err)
	assert.NotEmpty(t, cleaned)
}

func TestCleanHTML_RemovesHiddenElements(t *testing.T) {
	html := `<html><body>
		<div hidden>Hidden content</div>
		<div aria-hidden="true">Screen reader hidden</div>
		<div style="display: none">Display none</div>
		<div style="visibility:hidden">Visibility hidden</div>
		<p>Visible content</p>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "Hidden content")
	assert.NotContains(t, cleaned, "Screen reader hidden")
	assert.NotContains(t, cleaned, "Display none")
	assert.NotContains(t, cleaned, "Visibility hidden")
	assert.Contains(t, cleaned, "Visible content")
}

func TestCleanHTML_RemovesSidebarAndRelated(t *testing.T) {
	html := `<html><body>
		<div class="sidebar"><p>Side navigation</p></div>
		<div class="related-jobs"><p>Similar jobs</p></div>
		<div class="recommend-area"><p>Recommended for you</p></div>
		<div><h1>Main Job Content</h1></div>
	</body></html>`

	cleaned, err := CleanHTML(html)
	require.NoError(t, err)
	assert.NotContains(t, cleaned, "Side navigation")
	assert.NotContains(t, cleaned, "Similar jobs")
	assert.NotContains(t, cleaned, "Recommended for you")
	assert.Contains(t, cleaned, "Main Job Content")
}

// --- ExtractMainContent tests ---

func TestExtractMainContent_FindsArticle(t *testing.T) {
	html := `<html><body>
		<div class="noise">Side noise that should be ignored</div>
		<article><h1>Real Content</h1><p>Details about the job posting here</p></article>
	</body></html>`

	content, err := ExtractMainContent(html)
	require.NoError(t, err)
	assert.Contains(t, content, "Real Content")
	assert.Contains(t, content, "Details about the job posting")
}

func TestExtractMainContent_FindsMainElement(t *testing.T) {
	html := `<html><body>
		<div>Wrapper</div>
		<main><h1>Developer Position</h1><p>Hiring for senior role with Go experience</p></main>
	</body></html>`

	content, err := ExtractMainContent(html)
	require.NoError(t, err)
	assert.Contains(t, content, "Developer Position")
	assert.Contains(t, content, "senior role")
}

func TestExtractMainContent_FindsContentByID(t *testing.T) {
	html := `<html><body>
		<div id="content">
			<h1>데이터 엔지니어 채용</h1>
			<p>Python 3년 이상 경험 필수. Spark/Kafka 경험 우대.</p>
		</div>
	</body></html>`

	content, err := ExtractMainContent(html)
	require.NoError(t, err)
	assert.Contains(t, content, "데이터 엔지니어")
	assert.Contains(t, content, "Python")
}

func TestExtractMainContent_FindsKoreanSiteContainer(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		expected string
	}{
		{"Saramin wrap_jv_cont", "wrap_jv_cont", "사람인 채용 상세"},
		{"JobKorea artReadJobWrap", "artReadJobWrap", "잡코리아 채용 상세"},
		{"Wanted job_description", "job_description", "원티드 채용 상세"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body>
				<div class="` + tt.selector + `">
					<h1>` + tt.expected + `</h1>
					<p>상세한 채용 정보가 여기에 있습니다.</p>
				</div>
			</body></html>`

			content, err := ExtractMainContent(html)
			require.NoError(t, err)
			assert.Contains(t, content, tt.expected)
		})
	}
}

func TestExtractMainContent_ReturnsEmptyWhenNoMatch(t *testing.T) {
	html := `<html><body>
		<div class="random"><p>Short</p></div>
	</body></html>`

	content, err := ExtractMainContent(html)
	require.NoError(t, err)
	assert.Empty(t, content)
}

func TestExtractMainContent_SkipsShortContent(t *testing.T) {
	html := `<html><body>
		<article><p>Hi</p></article>
		<div id="content"><h1>Full job posting with enough content to pass the 50 char threshold</h1></div>
	</body></html>`

	content, err := ExtractMainContent(html)
	require.NoError(t, err)
	// article has < 50 chars, should skip to #content
	assert.Contains(t, content, "Full job posting")
}

// --- ResolveURLs tests ---

func TestResolveURLs_RelativeHrefs(t *testing.T) {
	html := `<html><body><a href="/jobs/123">Link</a></body></html>`
	resolved := ResolveURLs(html, "https://www.saramin.co.kr")
	assert.Contains(t, resolved, "https://www.saramin.co.kr/jobs/123")
}

func TestResolveURLs_RelativeSrc(t *testing.T) {
	html := `<html><body><img src="/images/logo.png"></body></html>`
	resolved := ResolveURLs(html, "https://www.jobkorea.co.kr/Recruit/GI_Read/123")
	assert.Contains(t, resolved, "https://www.jobkorea.co.kr/images/logo.png")
}

func TestResolveURLs_AbsoluteURLsUnchanged(t *testing.T) {
	html := `<html><body><a href="https://example.com/page">Link</a></body></html>`
	resolved := ResolveURLs(html, "https://www.saramin.co.kr")
	assert.Contains(t, resolved, "https://example.com/page")
}

func TestResolveURLs_SpecialSchemes(t *testing.T) {
	html := `<html><body>
		<a href="mailto:hr@company.com">Email</a>
		<a href="tel:+82-2-1234-5678">Call</a>
		<a href="javascript:void(0)">Click</a>
	</body></html>`

	resolved := ResolveURLs(html, "https://www.example.com")
	assert.Contains(t, resolved, "mailto:hr@company.com")
	assert.Contains(t, resolved, "tel:+82-2-1234-5678")
	assert.Contains(t, resolved, "javascript:void(0)")
}

func TestResolveURLs_InvalidBaseURL(t *testing.T) {
	html := `<html><body><a href="/path">Link</a></body></html>`
	resolved := ResolveURLs(html, "://invalid")
	// Should return original HTML when base URL is invalid
	assert.Contains(t, resolved, "/path")
}

func TestResolveURLs_ProtocolRelative(t *testing.T) {
	html := `<html><body><img src="//cdn.example.com/img.jpg"></body></html>`
	resolved := ResolveURLs(html, "https://www.saramin.co.kr")
	assert.Contains(t, resolved, "https://cdn.example.com/img.jpg")
}
