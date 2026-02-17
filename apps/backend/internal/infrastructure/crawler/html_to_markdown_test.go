package crawler

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func selFromHTML(t *testing.T, html string) *goquery.Selection {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	return doc.Find("body").First()
}

func TestSelectionToMarkdown_UnorderedList(t *testing.T) {
	html := `<body><ul><li>API 설계 및 개발</li><li>코드 리뷰</li><li>기술 멘토링</li></ul></body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Contains(t, result, "- API 설계 및 개발")
	assert.Contains(t, result, "- 코드 리뷰")
	assert.Contains(t, result, "- 기술 멘토링")
}

func TestSelectionToMarkdown_OrderedList(t *testing.T) {
	html := `<body><ol><li>1차 서류 심사</li><li>2차 기술 면접</li><li>3차 컬처 면접</li></ol></body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Contains(t, result, "1차 서류 심사")
	assert.Contains(t, result, "2차 기술 면접")
	assert.Contains(t, result, "3차 컬처 면접")
}

func TestSelectionToMarkdown_Bold(t *testing.T) {
	html := `<body><p><strong>필수</strong>: Go 언어 경험 3년 이상</p></body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Contains(t, result, "필수")
	assert.Contains(t, result, "Go 언어 경험 3년 이상")
}

func TestSelectionToMarkdown_LineBreaks(t *testing.T) {
	html := `<body>첫 번째 줄<br>두 번째 줄<br>세 번째 줄</body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	lines := strings.Split(strings.TrimSpace(result), "\n")
	// Should have at least 2 lines (br creates newlines)
	assert.GreaterOrEqual(t, len(lines), 2, "br tags should create line breaks")
}

func TestSelectionToMarkdown_Mixed(t *testing.T) {
	html := `<body>
		<p><strong>담당업무</strong></p>
		<ul>
			<li>백엔드 API 설계 및 개발</li>
			<li>MSA 아키텍처 설계</li>
		</ul>
		<p>추가 사항: 코드 리뷰 참여</p>
	</body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Contains(t, result, "담당업무")
	assert.Contains(t, result, "- 백엔드 API 설계 및 개발")
	assert.Contains(t, result, "- MSA 아키텍처 설계")
	assert.Contains(t, result, "코드 리뷰 참여")
}

func TestSelectionToMarkdown_PlainText(t *testing.T) {
	html := `<body>Go 또는 Java 5년 이상 경력, 대규모 트래픽 경험</body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Contains(t, result, "Go 또는 Java 5년 이상 경력")
	assert.Contains(t, result, "대규모 트래픽 경험")
}

func TestSelectionToMarkdown_EmptySelection(t *testing.T) {
	html := `<body></body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Equal(t, "", result)
}

func TestSelectionToMarkdown_NestedLists(t *testing.T) {
	html := `<body>
		<ul>
			<li>프론트엔드
				<ul>
					<li>React</li>
					<li>TypeScript</li>
				</ul>
			</li>
			<li>백엔드
				<ul>
					<li>Go</li>
					<li>PostgreSQL</li>
				</ul>
			</li>
		</ul>
	</body>`
	result := selectionToMarkdown(selFromHTML(t, html))

	assert.Contains(t, result, "React")
	assert.Contains(t, result, "TypeScript")
	assert.Contains(t, result, "Go")
	assert.Contains(t, result, "PostgreSQL")
}
