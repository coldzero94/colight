package crawler

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WantedExtractor extracts job posting data from Wanted's __NEXT_DATA__ JSON.
type WantedExtractor struct{}

// NewWantedExtractor creates a new WantedExtractor.
func NewWantedExtractor() *WantedExtractor {
	return &WantedExtractor{}
}

// wantedNextData represents the __NEXT_DATA__ JSON structure from Wanted pages.
// Supports both the new structure (initialData, 2026~) and legacy (job.detail).
type wantedNextData struct {
	Props struct {
		PageProps struct {
			// New structure (2026~)
			InitialData *wantedInitialData `json:"initialData"`
			// Legacy structure
			Job struct {
				Detail struct {
					Position     string `json:"position"`
					Company      struct {
						Name string `json:"name"`
					} `json:"company"`
					Intro        string `json:"intro"`
					Requirements string `json:"requirements"`
					Preferred    string `json:"preferred"`
					SkillTags    []struct {
						Title string `json:"title"`
					} `json:"skill_tags"`
				} `json:"detail"`
			} `json:"job"`
		} `json:"pageProps"`
	} `json:"props"`
}

// wantedInitialData represents the new Wanted __NEXT_DATA__ structure.
type wantedInitialData struct {
	Position string `json:"position"`
	Company  struct {
		CompanyName string `json:"company_name"`
	} `json:"company"`
	MainTasks       string `json:"main_tasks"`
	Intro           string `json:"intro"` // company intro (not main tasks)
	Requirements    string `json:"requirements"`
	PreferredPoints string `json:"preferred_points"`
	Benefits        string `json:"benefits"`
	CategoryTag     struct {
		ChildTags []struct {
			Text string `json:"text"`
		} `json:"child_tags"`
	} `json:"category_tag"`
	Address struct {
		FullLocation string `json:"full_location"`
	} `json:"address"`
}

// ExtractFromNextData parses __NEXT_DATA__ JSON from Wanted SSR HTML.
// Supports both new (initialData) and legacy (job.detail) structures.
// Returns nil if __NEXT_DATA__ not found or parsing fails — caller falls through to Tier 2.
func (e *WantedExtractor) ExtractFromNextData(html string) (*RawJobPosting, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, nil
	}

	scriptSel := doc.Find(`script#__NEXT_DATA__`)
	if scriptSel.Length() == 0 {
		return nil, nil
	}

	jsonText := strings.TrimSpace(scriptSel.First().Text())
	if jsonText == "" {
		return nil, nil
	}

	var data wantedNextData
	if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
		return nil, nil // graceful: malformed JSON → fallback to Tier 2
	}

	// Try new structure first (initialData)
	if data.Props.PageProps.InitialData != nil {
		return e.extractFromInitialData(data.Props.PageProps.InitialData)
	}

	// Fallback to legacy structure (job.detail)
	return e.extractFromLegacyDetail(&data)
}

func (e *WantedExtractor) extractFromInitialData(d *wantedInitialData) (*RawJobPosting, error) {
	if d.Position == "" && d.Company.CompanyName == "" {
		return nil, nil
	}

	result := &RawJobPosting{
		Source:       "wanted",
		CompanyName:  d.Company.CompanyName,
		Position:     d.Position,
		MainTasks:    d.MainTasks,
		Requirements: d.Requirements,
		Preferred:    d.PreferredPoints,
		Location:     d.Address.FullLocation,
		Skills:       make([]string, 0, len(d.CategoryTag.ChildTags)),
	}

	for _, tag := range d.CategoryTag.ChildTags {
		if t := strings.TrimSpace(tag.Text); t != "" {
			result.Skills = append(result.Skills, t)
		}
	}

	return result, nil
}

func (e *WantedExtractor) extractFromLegacyDetail(data *wantedNextData) (*RawJobPosting, error) {
	detail := data.Props.PageProps.Job.Detail
	if detail.Position == "" && detail.Company.Name == "" {
		return nil, nil
	}

	result := &RawJobPosting{
		Source:       "wanted",
		CompanyName:  detail.Company.Name,
		Position:     detail.Position,
		MainTasks:    detail.Intro,
		Requirements: detail.Requirements,
		Preferred:    detail.Preferred,
		Skills:       make([]string, 0, len(detail.SkillTags)),
	}

	for _, tag := range detail.SkillTags {
		if t := strings.TrimSpace(tag.Title); t != "" {
			result.Skills = append(result.Skills, t)
		}
	}

	return result, nil
}
