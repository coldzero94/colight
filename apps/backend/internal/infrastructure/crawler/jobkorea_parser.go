package crawler

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// JobKoreaParser parses JobKorea job posting HTML
type JobKoreaParser struct{}

// NewJobKoreaParser creates a new JobKorea parser
func NewJobKoreaParser() *JobKoreaParser {
	return &JobKoreaParser{}
}

// ParseHTML parses JobKorea HTML and extracts job posting data
func (p *JobKoreaParser) ParseHTML(url, html string) (*RawJobPosting, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	result := &RawJobPosting{
		Source:    "jobkorea",
		SourceURL: url,
	}

	// Extract company name
	result.CompanyName = strings.TrimSpace(doc.Find(".coName, .company-name a").First().Text())

	// Extract position
	result.Position = strings.TrimSpace(doc.Find(".artReadJobTitle, .title-wrap h1").First().Text())

	// Extract department
	result.Department = strings.TrimSpace(doc.Find(".tbRow .dept").First().Text())

	// Extract career
	result.Career = strings.TrimSpace(doc.Find(".tbRow .career").First().Text())

	// Extract education
	result.Education = strings.TrimSpace(doc.Find(".tbRow .education").First().Text())

	// Extract job type
	result.JobType = strings.TrimSpace(doc.Find(".tbRow .jobtype").First().Text())

	// Extract salary
	result.Salary = strings.TrimSpace(doc.Find(".tbRow .salary").First().Text())

	// Extract location
	result.Location = strings.TrimSpace(doc.Find(".tbRow .location").First().Text())

	// Extract deadline
	result.Deadline = strings.TrimSpace(doc.Find(".date .tahoma").First().Text())

	// Extract main tasks
	result.MainTasks = strings.TrimSpace(doc.Find(".artReadJobSecCont").Eq(0).Text())

	// Extract requirements
	result.Requirements = strings.TrimSpace(doc.Find(".artReadJobSecCont").Eq(1).Text())

	// Extract preferred
	result.Preferred = strings.TrimSpace(doc.Find(".artReadJobSecCont").Eq(2).Text())

	// Extract skills
	result.Skills = []string{}
	doc.Find(".skillWrap .skill").Each(func(i int, s *goquery.Selection) {
		skill := strings.TrimSpace(s.Text())
		if skill != "" {
			result.Skills = append(result.Skills, skill)
		}
	})

	return result, nil
}
