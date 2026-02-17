package crawler

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// JobKoreaParser parses JobKorea job posting HTML.
// Handles both legacy CSS format and new Next.js App Router format.
type JobKoreaParser struct{}

// NewJobKoreaParser creates a new JobKorea parser
func NewJobKoreaParser() *JobKoreaParser {
	return &JobKoreaParser{}
}

// jobKoreaJSONLD represents schema.org/JobPosting JSON-LD from JobKorea pages.
type jobKoreaJSONLD struct {
	Type               string `json:"@type"`
	Title              string `json:"title"`
	HiringOrganization struct {
		Name string `json:"name"`
	} `json:"hiringOrganization"`
	JobLocation struct {
		Address struct {
			AddressLocality string `json:"addressLocality"`
		} `json:"address"`
	} `json:"jobLocation"`
	BaseSalary struct {
		Value interface{} `json:"value"` // can be string or object
	} `json:"baseSalary"`
	ExperienceRequirements string `json:"experienceRequirements"`
}

// descriptionURLPattern matches S3-hosted description URLs in JobKorea pages.
var descriptionURLPattern = regexp.MustCompile(`https://job-hub-files[^"'\s]+_DESCRIPTION\.html[^"'\s]*`)

// ParseHTML parses JobKorea HTML. Handles both legacy and new Next.js formats.
func (p *JobKoreaParser) ParseHTML(url, html string) (*RawJobPosting, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	result := &RawJobPosting{
		Source:    "jobkorea",
		SourceURL: url,
	}

	// Strategy 1: JSON-LD extraction (new Next.js format, most reliable)
	p.extractFromJSONLD(doc, result)

	// Strategy 2: data-sentry-component selectors (new format)
	if result.CompanyName == "" {
		p.extractFromSentryComponents(doc, result)
	}

	// Strategy 3: Legacy CSS selectors (old format)
	if result.CompanyName == "" && result.Position == "" {
		p.extractFromLegacyCSS(doc, result)
	}

	return result, nil
}

// DescriptionURL extracts the S3 description URL from JobKorea HTML.
// Returns empty string if not found (legacy pages don't have this).
func (p *JobKoreaParser) DescriptionURL(html string) string {
	return descriptionURLPattern.FindString(html)
}

// ParseDescriptionHTML parses the separately-fetched S3 description HTML.
// Fills in MainTasks, Requirements, Preferred from the description content.
func (p *JobKoreaParser) ParseDescriptionHTML(result *RawJobPosting, descHTML string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(descHTML))
	if err != nil {
		return
	}

	// Try section headers (h3/strong with Korean labels)
	doc.Find("h3, h4, strong, .hd_3 .tit").Each(func(i int, header *goquery.Selection) {
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

	// Fallback: try ordered section containers (old S3 format)
	if result.MainTasks == "" {
		secConts := doc.Find(".artReadJobSecCont, .detailType_1 .cont, .artTplDetail .cont")
		if secConts.Length() > 0 {
			result.MainTasks = selectionToMarkdown(secConts.Eq(0))
		}
		if secConts.Length() > 1 && result.Requirements == "" {
			result.Requirements = selectionToMarkdown(secConts.Eq(1))
		}
		if secConts.Length() > 2 && result.Preferred == "" {
			result.Preferred = selectionToMarkdown(secConts.Eq(2))
		}
	}

	// Last resort: get full body text as MainTasks
	if result.MainTasks == "" {
		allText := selectionToMarkdown(doc.Find("body"))
		if allText != "" {
			result.MainTasks = allText
		}
	}
}

func (p *JobKoreaParser) extractFromJSONLD(doc *goquery.Document, result *RawJobPosting) {
	doc.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
		var ld jobKoreaJSONLD
		if err := json.Unmarshal([]byte(s.Text()), &ld); err != nil {
			return
		}
		if ld.Type != "JobPosting" {
			return
		}

		if result.CompanyName == "" {
			result.CompanyName = ld.HiringOrganization.Name
		}
		if result.Position == "" {
			result.Position = ld.Title
		}
		if result.Location == "" {
			result.Location = ld.JobLocation.Address.AddressLocality
		}
		if result.Career == "" {
			result.Career = ld.ExperienceRequirements
		}
		if result.Salary == "" {
			switch v := ld.BaseSalary.Value.(type) {
			case string:
				result.Salary = v
			}
		}
	})
}

func (p *JobKoreaParser) extractFromSentryComponents(doc *goquery.Document, result *RawJobPosting) {
	if cn := strings.TrimSpace(doc.Find(`[data-sentry-component="CompanyName"] h2`).First().Text()); cn != "" {
		result.CompanyName = cn
	}
	if pos := strings.TrimSpace(doc.Find(`[data-sentry-component="TitleContent"] h1`).First().Text()); pos != "" {
		result.Position = pos
	}
}

func (p *JobKoreaParser) extractFromLegacyCSS(doc *goquery.Document, result *RawJobPosting) {
	result.CompanyName = strings.TrimSpace(doc.Find(".coName, .company-name a").First().Text())
	result.Position = strings.TrimSpace(doc.Find(".artReadJobTitle, .title-wrap h1").First().Text())
	result.Department = strings.TrimSpace(doc.Find(".tbRow .dept").First().Text())
	result.Career = strings.TrimSpace(doc.Find(".tbRow .career").First().Text())
	result.Education = strings.TrimSpace(doc.Find(".tbRow .education").First().Text())
	result.JobType = strings.TrimSpace(doc.Find(".tbRow .jobtype").First().Text())
	result.Salary = strings.TrimSpace(doc.Find(".tbRow .salary").First().Text())
	result.Location = strings.TrimSpace(doc.Find(".tbRow .location").First().Text())
	result.Deadline = strings.TrimSpace(doc.Find(".date .tahoma").First().Text())

	result.MainTasks = selectionToMarkdown(doc.Find(".artReadJobSecCont").Eq(0))
	result.Requirements = selectionToMarkdown(doc.Find(".artReadJobSecCont").Eq(1))
	result.Preferred = selectionToMarkdown(doc.Find(".artReadJobSecCont").Eq(2))

	result.Skills = []string{}
	doc.Find(".skillWrap .skill").Each(func(i int, s *goquery.Selection) {
		skill := strings.TrimSpace(s.Text())
		if skill != "" {
			result.Skills = append(result.Skills, skill)
		}
	})
}
