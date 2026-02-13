package crawler

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CatchParser parses Catch job posting HTML
type CatchParser struct{}

// NewCatchParser creates a new Catch parser
func NewCatchParser() *CatchParser {
	return &CatchParser{}
}

// ParseHTML parses Catch HTML and extracts job posting data
func (p *CatchParser) ParseHTML(url, html string) (*RawJobPosting, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	result := &RawJobPosting{
		Source:    "catch",
		SourceURL: url,
	}

	// Catch selectors (example - actual selectors may vary)
	result.CompanyName = strings.TrimSpace(doc.Find(".company-name, .recruit-company").First().Text())
	result.Position = strings.TrimSpace(doc.Find(".job-title, h1.title").First().Text())
	result.Department = strings.TrimSpace(doc.Find(".department").First().Text())
	result.Career = strings.TrimSpace(doc.Find(".career").First().Text())
	result.Location = strings.TrimSpace(doc.Find(".location, .work-place").First().Text())
	result.Deadline = strings.TrimSpace(doc.Find(".deadline, .dday").First().Text())

	// Extract job details
	result.MainTasks = strings.TrimSpace(doc.Find(".job-description, .main-tasks").First().Text())
	result.Requirements = strings.TrimSpace(doc.Find(".requirements, .qualifications").First().Text())
	result.Preferred = strings.TrimSpace(doc.Find(".preferred, .preferred-qualifications").First().Text())

	// Extract skills
	result.Skills = []string{}
	doc.Find(".skill-tag, .tech-stack span").Each(func(i int, s *goquery.Selection) {
		skill := strings.TrimSpace(s.Text())
		if skill != "" {
			result.Skills = append(result.Skills, skill)
		}
	})

	return result, nil
}
