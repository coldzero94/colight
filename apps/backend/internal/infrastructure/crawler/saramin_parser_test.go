package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaraminParser_FullPosting(t *testing.T) {
	html := `<html><body>
		<div class="wrap_jv_cont">
			<div class="wrap_jv_header">
				<h1 class="job_tit">시니어 백엔드 개발자</h1>
				<a class="company_name">네이버</a>
			</div>
			<div class="jv_cont jv_summary">
				<div class="cont">
					<dl>
						<dt>경력</dt><dd>경력 5년 이상</dd>
						<dt>학력</dt><dd>학사 이상</dd>
						<dt>고용형태</dt><dd>정규직</dd>
						<dt>근무지역</dt><dd>경기 성남시 분당구</dd>
						<dt>마감일</dt><dd>2026-06-30</dd>
					</dl>
				</div>
			</div>
			<div class="jv_cont jv_detail">
				<div class="cont">
					<h3>담당업무</h3>
					<ul>
						<li>백엔드 API 설계 및 개발</li>
						<li>MSA 아키텍처 설계</li>
						<li>코드 리뷰 및 기술 멘토링</li>
					</ul>
					<h3>자격요건</h3>
					<ul>
						<li>Go 또는 Java 5년 이상 경력</li>
						<li>대규모 트래픽 서비스 경험</li>
					</ul>
					<h3>우대사항</h3>
					<ul>
						<li>Kubernetes 운영 경험</li>
						<li>gRPC 사용 경험</li>
					</ul>
				</div>
			</div>
			<div class="jv_cont jv_skill">
				<div class="cont">
					<span class="stack">Go</span>
					<span class="stack">PostgreSQL</span>
					<span class="stack">Docker</span>
				</div>
			</div>
		</div>
	</body></html>`

	parser := NewSaraminParser()
	result, err := parser.ParseHTML("https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=12345", html)

	require.NoError(t, err)
	assert.Equal(t, "saramin", result.Source)
	assert.Equal(t, "네이버", result.CompanyName)
	assert.Equal(t, "시니어 백엔드 개발자", result.Position)
	assert.Equal(t, "경력 5년 이상", result.Career)
	assert.Equal(t, "학사 이상", result.Education)
	assert.Equal(t, "정규직", result.JobType)
	assert.Equal(t, "경기 성남시 분당구", result.Location)
	assert.Equal(t, "2026-06-30", result.Deadline)
}

func TestSaraminParser_RichContent(t *testing.T) {
	html := `<html><body>
		<div class="wrap_jv_cont">
			<h1 class="job_tit">개발자</h1>
			<a class="company_name">테스트</a>
			<div class="jv_cont jv_detail">
				<div class="cont">
					<h3>담당업무</h3>
					<ul>
						<li>API 설계 및 개발</li>
						<li>아키텍처 설계</li>
						<li>기술 멘토링</li>
					</ul>
				</div>
			</div>
		</div>
	</body></html>`

	parser := NewSaraminParser()
	result, err := parser.ParseHTML("https://www.saramin.co.kr/test", html)

	require.NoError(t, err)
	// MainTasks should preserve list structure (markdown format)
	assert.Contains(t, result.MainTasks, "API 설계 및 개발")
	assert.Contains(t, result.MainTasks, "아키텍처 설계")
	assert.Contains(t, result.MainTasks, "기술 멘토링")
}

func TestSaraminParser_Skills(t *testing.T) {
	html := `<html><body>
		<div class="wrap_jv_cont">
			<h1 class="job_tit">개발자</h1>
			<a class="company_name">테스트</a>
			<div class="jv_cont jv_skill">
				<div class="cont">
					<span class="stack">Go</span>
					<span class="stack">PostgreSQL</span>
					<span class="stack">Docker</span>
				</div>
			</div>
		</div>
	</body></html>`

	parser := NewSaraminParser()
	result, err := parser.ParseHTML("https://www.saramin.co.kr/test", html)

	require.NoError(t, err)
	assert.Contains(t, result.Skills, "Go")
	assert.Contains(t, result.Skills, "PostgreSQL")
	assert.Contains(t, result.Skills, "Docker")
}

func TestSaraminParser_SelectorMismatch(t *testing.T) {
	html := `<html><body><div class="unknown"><p>No matching selectors</p></div></body></html>`

	parser := NewSaraminParser()
	result, err := parser.ParseHTML("https://www.saramin.co.kr/test", html)

	require.NoError(t, err)
	assert.Equal(t, "", result.CompanyName)
	assert.Equal(t, "", result.Position)
	assert.Equal(t, "saramin", result.Source)
}
