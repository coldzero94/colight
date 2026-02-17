package crawler

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// SaraminParser parses Saramin AJAX job posting HTML.
type SaraminParser struct{}

// NewSaraminParser creates a new Saramin parser.
func NewSaraminParser() *SaraminParser {
	return &SaraminParser{}
}

// ParseHTML parses Saramin AJAX HTML and extracts job posting data.
// Uses selectionToMarkdown for content fields to preserve list structure.
func (p *SaraminParser) ParseHTML(url, html string) (*RawJobPosting, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	result := &RawJobPosting{
		Source:    "saramin",
		SourceURL: url,
	}

	// Company name — multiple possible selectors
	result.CompanyName = strings.TrimSpace(doc.Find(".company_name, .corp_name a").First().Text())

	// Position — job title
	result.Position = strings.TrimSpace(doc.Find(".job_tit, .tit_job").First().Text())

	// Metadata from DL or summary section
	doc.Find(".jv_cont.jv_summary dl dt").Each(func(i int, dt *goquery.Selection) {
		label := strings.TrimSpace(dt.Text())
		dd := dt.Next()
		if !dd.Is("dd") {
			return
		}
		value := strings.TrimSpace(dd.Text())
		if value == "" {
			return
		}

		switch {
		case containsAny(label, "경력"):
			result.Career = value
		case containsAny(label, "학력"):
			result.Education = value
		case containsAny(label, "고용형태", "근무형태"):
			result.JobType = value
		case containsAny(label, "근무지역", "근무지"):
			result.Location = value
		case containsAny(label, "급여", "연봉"):
			result.Salary = value
		case containsAny(label, "마감일", "마감"):
			result.Deadline = value
		}
	})

	// Job detail sections — use universal section extraction within the detail area
	detailArea := doc.Find(".jv_cont.jv_detail .cont, .jv_detail")
	if detailArea.Length() > 0 {
		extractSaraminDetailSections(detailArea, result)
	}

	// Skills — from skill/stack tags
	result.Skills = []string{}
	doc.Find(".jv_cont.jv_skill .stack, .skill_wrap span, .skill-tag").Each(func(i int, s *goquery.Selection) {
		skill := strings.TrimSpace(s.Text())
		if skill != "" {
			result.Skills = append(result.Skills, skill)
		}
	})

	return result, nil
}

// extractSaraminDetailSections extracts content sections from Saramin's detail area.
// Looks for h3/h4/strong headers matching Korean section names and extracts content after them.
func extractSaraminDetailSections(area *goquery.Selection, result *RawJobPosting) {
	area.Find("h3, h4, strong").Each(func(i int, header *goquery.Selection) {
		headerText := strings.TrimSpace(header.Text())
		if headerText == "" {
			return
		}

		switch {
		case containsAny(headerText, "담당업무", "주요업무", "업무내용", "직무내용"):
			if result.MainTasks == "" {
				result.MainTasks = extractContentAfterHeader(header)
			}
		case containsAny(headerText, "자격요건", "지원자격", "필수자격"):
			if result.Requirements == "" {
				result.Requirements = extractContentAfterHeader(header)
			}
		case containsAny(headerText, "우대사항", "우대조건", "우대요건"):
			if result.Preferred == "" {
				result.Preferred = extractContentAfterHeader(header)
			}
		}
	})
}


// containsAny returns true if s contains any of the substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
