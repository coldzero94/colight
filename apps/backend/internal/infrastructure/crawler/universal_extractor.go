package crawler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractionResult holds structured data extracted from HTML/markdown without LLM.
type ExtractionResult struct {
	Raw          RawJobPosting
	SectionTexts map[string]string // key → full section content (markdown)
	FilledFields int               // number of non-empty fields populated
}

// IsEnoughForNormalize returns true if enough structured data was extracted
// to use the lightweight normalize LLM prompt instead of the heavy extract prompt.
// Requires: (CompanyName OR Position) AND (MainTasks OR Requirements non-empty).
func (r *ExtractionResult) IsEnoughForNormalize() bool {
	hasIdentity := r.Raw.CompanyName != "" || r.Raw.Position != ""
	hasContent := r.Raw.MainTasks != "" || r.Raw.Requirements != ""
	return hasIdentity && hasContent
}

// sectionAliases maps canonical section keys to Korean/English header aliases.
var sectionAliases = map[string][]string{
	"main_tasks":    {"담당업무", "주요업무", "업무내용", "직무내용", "직무소개", "주요 업무", "업무 내용", "What you'll do", "Job Description"},
	"requirements":  {"자격요건", "지원자격", "필수자격", "자격조건", "필수요건", "지원 자격", "자격 요건", "What we're looking for", "Requirements"},
	"preferred":     {"우대사항", "우대조건", "우대요건", "우대역량", "우대 사항", "우대 조건", "Plus if you have", "Preferred"},
	"skills":        {"기술스택", "사용기술", "필요역량", "Tech Stack", "기술 스택", "개발환경"},
	"benefits":      {"복리후생", "혜택", "Benefits", "복지", "처우"},
	"company_intro": {"회사소개", "기업소개", "About us", "회사 소개"},
}

// metadataLabels maps RawJobPosting fields to Korean label aliases found in tables/DLs.
var metadataLabels = map[string][]string{
	"job_type":  {"고용형태", "근무형태", "계약형태"},
	"career":    {"경력", "경력조건"},
	"education": {"학력", "학력조건"},
	"location":  {"근무지", "근무지역", "근무위치"},
	"salary":    {"급여", "연봉", "보상"},
	"deadline":  {"마감일", "마감일자", "접수마감", "모집기한"},
}

// skillSelectors are CSS selectors commonly used for skill/tech-stack tags across Korean job sites.
var skillSelectors = []string{
	".skill-tag",
	".tech-stack span", ".tech-stack li",
	".skillWrap .skill",
	"[class*='skill'] span", "[class*='skill'] li",
	"[class*='stack'] span", "[class*='tech'] span",
	".tag_area a", ".tag_area span",
}

// headingSelectors are CSS selectors for elements that act as section headers.
var headingSelectors = "h1, h2, h3, h4, h5, dt, th, strong, .tit"

// ExtractFromHTML extracts structured job posting data from raw HTML without LLM.
// Uses meta tags, Korean section headers, table/DL metadata, and skill tag selectors.
func ExtractFromHTML(html, sourceURL string) *ExtractionResult {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return &ExtractionResult{SectionTexts: map[string]string{}}
	}

	result := &ExtractionResult{
		Raw:          RawJobPosting{SourceURL: sourceURL},
		SectionTexts: map[string]string{},
	}

	// A0. JSON-LD structured data (most reliable when available)
	extractJSONLD(doc, result)

	// A. Meta tag extraction
	extractMetaTags(doc, result)

	// B. Korean section header detection + full content extraction
	extractSections(doc, result)

	// C. Table/DL metadata extraction
	extractMetadata(doc, result)

	// D. Skill tag extraction
	extractSkills(doc, result)

	// Count filled fields
	result.FilledFields = countFilledFields(&result.Raw)

	return result
}

// universalJSONLD represents a schema.org/JobPosting JSON-LD object.
type universalJSONLD struct {
	Type               string `json:"@type"`
	Title              string `json:"title"`
	HiringOrganization struct {
		Name string `json:"name"`
	} `json:"hiringOrganization"`
	JobLocation struct {
		Address struct {
			AddressLocality string `json:"addressLocality"`
			StreetAddress   string `json:"streetAddress"`
		} `json:"address"`
	} `json:"jobLocation"`
	EmploymentType         json.RawMessage `json:"employmentType"` // can be string or []string
	ExperienceRequirements string          `json:"experienceRequirements"`
}

// extractJSONLD parses schema.org/JobPosting JSON-LD from <script type="application/ld+json">.
func extractJSONLD(doc *goquery.Document, result *ExtractionResult) {
	doc.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
		var ld universalJSONLD
		if err := json.Unmarshal([]byte(s.Text()), &ld); err != nil {
			return
		}
		if ld.Type != "JobPosting" {
			return
		}

		if result.Raw.CompanyName == "" && ld.HiringOrganization.Name != "" {
			result.Raw.CompanyName = ld.HiringOrganization.Name
		}
		if result.Raw.Position == "" && ld.Title != "" {
			result.Raw.Position = ld.Title
		}
		if result.Raw.Location == "" {
			loc := ld.JobLocation.Address.StreetAddress
			if loc == "" {
				loc = ld.JobLocation.Address.AddressLocality
			}
			if loc != "" {
				result.Raw.Location = loc
			}
		}
		if result.Raw.Career == "" && ld.ExperienceRequirements != "" {
			result.Raw.Career = ld.ExperienceRequirements
		}
		if result.Raw.JobType == "" && len(ld.EmploymentType) > 0 {
			var s string
			if json.Unmarshal(ld.EmploymentType, &s) == nil {
				result.Raw.JobType = s
			} else {
				var arr []string
				if json.Unmarshal(ld.EmploymentType, &arr) == nil && len(arr) > 0 {
					result.Raw.JobType = arr[0]
				}
			}
		}
	})
}

// extractMetaTags extracts company/position hints from meta tags and <title>.
func extractMetaTags(doc *goquery.Document, result *ExtractionResult) {
	// og:title → position hint
	ogTitle, _ := doc.Find(`meta[property="og:title"]`).Attr("content")
	if ogTitle != "" {
		parsePositionFromTitle(ogTitle, result)
	}

	// og:site_name → site name (not company)
	// Just use for context, don't overwrite company

	// <title> parsing as fallback
	title := doc.Find("title").First().Text()
	if title != "" && result.Raw.Position == "" {
		parsePositionFromTitle(title, result)
	}
}

// Title parsing patterns
var (
	// "포지션 - 회사 | 사이트" or "포지션 - 회사"
	reTitleDashPipe = regexp.MustCompile(`^(.+?)\s*[-–—]\s*(.+?)(?:\s*[|｜]\s*.+)?$`)
	// "[회사] 포지션"
	reTitleBracket = regexp.MustCompile(`^\[(.+?)\]\s*(.+)$`)
	// "회사 채용: 포지션"
	reTitleRecruit = regexp.MustCompile(`^(.+?)\s*채용[:\s]+(.+)$`)
)

func parsePositionFromTitle(title string, result *ExtractionResult) {
	title = strings.TrimSpace(title)

	if m := reTitleBracket.FindStringSubmatch(title); m != nil {
		if result.Raw.CompanyName == "" {
			result.Raw.CompanyName = strings.TrimSpace(m[1])
		}
		if result.Raw.Position == "" {
			result.Raw.Position = strings.TrimSpace(m[2])
		}
		return
	}

	if m := reTitleRecruit.FindStringSubmatch(title); m != nil {
		if result.Raw.CompanyName == "" {
			result.Raw.CompanyName = strings.TrimSpace(m[1])
		}
		if result.Raw.Position == "" {
			result.Raw.Position = strings.TrimSpace(m[2])
		}
		return
	}

	if m := reTitleDashPipe.FindStringSubmatch(title); m != nil {
		if result.Raw.Position == "" {
			result.Raw.Position = strings.TrimSpace(m[1])
		}
		if result.Raw.CompanyName == "" {
			result.Raw.CompanyName = strings.TrimSpace(m[2])
		}
		return
	}
}

// extractSections finds Korean section headers and extracts full content until the next header.
func extractSections(doc *goquery.Document, result *ExtractionResult) {
	// Build a reverse map: alias → canonical key
	aliasToKey := make(map[string]string)
	for key, aliases := range sectionAliases {
		for _, alias := range aliases {
			aliasToKey[strings.TrimSpace(alias)] = key
		}
	}

	// Find all heading-like elements
	doc.Find(headingSelectors).Each(func(i int, sel *goquery.Selection) {
		headerText := strings.TrimSpace(sel.Text())
		if headerText == "" {
			return
		}

		// Check if this header matches any section alias
		for alias, key := range aliasToKey {
			if strings.Contains(headerText, alias) {
				// Extract content: collect all siblings until the next heading
				content := extractContentAfterHeader(sel)
				if content != "" {
					// Keep the richer (longer) content if already exists
					if existing, ok := result.SectionTexts[key]; !ok || len(content) > len(existing) {
						result.SectionTexts[key] = content
					}
				}
				break
			}
		}
	})

	// Map section texts to RawJobPosting fields
	if text, ok := result.SectionTexts["main_tasks"]; ok && result.Raw.MainTasks == "" {
		result.Raw.MainTasks = text
	}
	if text, ok := result.SectionTexts["requirements"]; ok && result.Raw.Requirements == "" {
		result.Raw.Requirements = text
	}
	if text, ok := result.SectionTexts["preferred"]; ok && result.Raw.Preferred == "" {
		result.Raw.Preferred = text
	}
}

// extractContentAfterHeader collects content from all siblings after a header element
// until the next heading-like element is encountered.
func extractContentAfterHeader(header *goquery.Selection) string {
	var parts []string

	// First check: if the header's parent contains the content directly (common pattern)
	// e.g., <h3>담당업무</h3> followed by <ul> as sibling
	header.NextAll().EachWithBreak(func(i int, sibling *goquery.Selection) bool {
		// Stop at the next heading
		if sibling.Is(headingSelectors) {
			return false
		}
		md := selectionToMarkdown(sibling)
		if md != "" {
			parts = append(parts, md)
		}
		return true
	})

	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}

	// Fallback: check parent's next sibling
	parent := header.Parent()
	parent.NextAll().EachWithBreak(func(i int, sibling *goquery.Selection) bool {
		// Stop at element containing a heading
		if sibling.Find(headingSelectors).Length() > 0 {
			return false
		}
		md := selectionToMarkdown(sibling)
		if md != "" {
			parts = append(parts, md)
		}
		return true
	})

	return strings.Join(parts, "\n")
}

// extractMetadata extracts structured metadata from table rows and definition lists.
func extractMetadata(doc *goquery.Document, result *ExtractionResult) {
	// Build label → field map
	labelToField := make(map[string]string)
	for field, labels := range metadataLabels {
		for _, label := range labels {
			labelToField[label] = field
		}
	}

	// Extract from <table> th/td pairs
	doc.Find("table tr").Each(func(i int, tr *goquery.Selection) {
		th := strings.TrimSpace(tr.Find("th").First().Text())
		td := strings.TrimSpace(tr.Find("td").First().Text())
		if th == "" || td == "" {
			return
		}
		for label, field := range labelToField {
			if strings.Contains(th, label) {
				setMetadataField(&result.Raw, field, td)
				break
			}
		}
	})

	// Extract from <dl> dt/dd pairs
	doc.Find("dl dt").Each(func(i int, dt *goquery.Selection) {
		dtText := strings.TrimSpace(dt.Text())
		dd := dt.Next()
		if !dd.Is("dd") {
			return
		}
		ddText := strings.TrimSpace(dd.Text())
		if dtText == "" || ddText == "" {
			return
		}
		for label, field := range labelToField {
			if strings.Contains(dtText, label) {
				setMetadataField(&result.Raw, field, ddText)
				break
			}
		}
	})
}

// setMetadataField sets a metadata field on RawJobPosting by field name.
func setMetadataField(raw *RawJobPosting, field, value string) {
	switch field {
	case "job_type":
		if raw.JobType == "" {
			raw.JobType = value
		}
	case "career":
		if raw.Career == "" {
			raw.Career = value
		}
	case "education":
		if raw.Education == "" {
			raw.Education = value
		}
	case "location":
		if raw.Location == "" {
			raw.Location = value
		}
	case "salary":
		if raw.Salary == "" {
			raw.Salary = value
		}
	case "deadline":
		if raw.Deadline == "" {
			raw.Deadline = value
		}
	}
}

// extractSkills extracts skill tags using common CSS selectors.
func extractSkills(doc *goquery.Document, result *ExtractionResult) {
	seen := make(map[string]bool)
	result.Raw.Skills = []string{}

	for _, selector := range skillSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			skill := strings.TrimSpace(s.Text())
			if skill != "" && !seen[skill] {
				seen[skill] = true
				result.Raw.Skills = append(result.Raw.Skills, skill)
			}
		})
	}
}

// countFilledFields counts non-empty fields in a RawJobPosting.
func countFilledFields(raw *RawJobPosting) int {
	count := 0
	if raw.CompanyName != "" {
		count++
	}
	if raw.Position != "" {
		count++
	}
	if raw.Department != "" {
		count++
	}
	if raw.Career != "" {
		count++
	}
	if raw.Education != "" {
		count++
	}
	if raw.JobType != "" {
		count++
	}
	if raw.Salary != "" {
		count++
	}
	if raw.Location != "" {
		count++
	}
	if raw.Deadline != "" {
		count++
	}
	if raw.MainTasks != "" {
		count++
	}
	if raw.Requirements != "" {
		count++
	}
	if raw.Preferred != "" {
		count++
	}
	if len(raw.Skills) > 0 {
		count++
	}
	return count
}

// Markdown section header pattern: ## 담당업무, ### 자격요건, **우대사항**, etc.
var markdownSectionPattern = regexp.MustCompile(
	`(?m)^(?:#{1,4}\s*|\*{2})?(담당업무|주요업무|업무내용|직무내용|직무소개|자격요건|지원자격|필수자격|자격조건|필수요건|우대사항|우대조건|우대요건|우대역량|기술스택|사용기술|필요역량|복리후생|회사소개|기업소개)(?:\*{2})?[:\s]*$`,
)

// markdownAlias maps Korean section headers to canonical keys.
var markdownAlias = map[string]string{
	"담당업무": "main_tasks", "주요업무": "main_tasks", "업무내용": "main_tasks",
	"직무내용": "main_tasks", "직무소개": "main_tasks",
	"자격요건": "requirements", "지원자격": "requirements", "필수자격": "requirements",
	"자격조건": "requirements", "필수요건": "requirements",
	"우대사항": "preferred", "우대조건": "preferred", "우대요건": "preferred", "우대역량": "preferred",
	"기술스택": "skills", "사용기술": "skills", "필요역량": "skills",
	"복리후생": "benefits",
	"회사소개": "company_intro", "기업소개": "company_intro",
}

// ExtractFromMarkdown extracts structured sections from markdown content using Korean headers.
func ExtractFromMarkdown(markdown string) *ExtractionResult {
	result := &ExtractionResult{
		SectionTexts: map[string]string{},
	}

	lines := strings.Split(markdown, "\n")
	matches := markdownSectionPattern.FindAllStringIndex(markdown, -1)
	if len(matches) == 0 {
		return result
	}

	// Find line indices for each match
	type sectionMatch struct {
		key       string
		lineStart int
	}

	var sections []sectionMatch
	lineOffset := 0
	lineIdx := 0

	for _, m := range matches {
		// Find which line this match is on
		for lineIdx < len(lines) && lineOffset+len(lines[lineIdx])+1 <= m[0] {
			lineOffset += len(lines[lineIdx]) + 1 // +1 for \n
			lineIdx++
		}

		headerText := strings.TrimSpace(markdown[m[0]:m[1]])
		// Strip markdown formatting
		headerText = strings.TrimLeft(headerText, "# ")
		headerText = strings.Trim(headerText, "* :")

		if key, ok := markdownAlias[headerText]; ok {
			sections = append(sections, sectionMatch{key: key, lineStart: lineIdx})
		}
	}

	// Extract content between section headers
	for i, sec := range sections {
		startLine := sec.lineStart + 1 // skip the header line itself
		var endLine int
		if i+1 < len(sections) {
			endLine = sections[i+1].lineStart
		} else {
			endLine = len(lines)
		}

		if startLine >= len(lines) {
			continue
		}
		if endLine > len(lines) {
			endLine = len(lines)
		}

		sectionContent := strings.TrimSpace(strings.Join(lines[startLine:endLine], "\n"))
		if sectionContent != "" {
			if existing, ok := result.SectionTexts[sec.key]; !ok || len(sectionContent) > len(existing) {
				result.SectionTexts[sec.key] = sectionContent
			}
		}
	}

	// Map to RawJobPosting
	if text, ok := result.SectionTexts["main_tasks"]; ok {
		result.Raw.MainTasks = text
	}
	if text, ok := result.SectionTexts["requirements"]; ok {
		result.Raw.Requirements = text
	}
	if text, ok := result.SectionTexts["preferred"]; ok {
		result.Raw.Preferred = text
	}

	result.FilledFields = countFilledFields(&result.Raw)
	return result
}

// MergeExtractions merges two extraction results, keeping the richer (longer) value for each field.
func MergeExtractions(a, b *ExtractionResult) *ExtractionResult {
	if a == nil && b == nil {
		return &ExtractionResult{SectionTexts: map[string]string{}}
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	merged := &ExtractionResult{
		SectionTexts: map[string]string{},
	}

	// Merge RawJobPosting fields — pick longer value
	merged.Raw.CompanyName = longerString(a.Raw.CompanyName, b.Raw.CompanyName)
	merged.Raw.Position = longerString(a.Raw.Position, b.Raw.Position)
	merged.Raw.Department = longerString(a.Raw.Department, b.Raw.Department)
	merged.Raw.Career = longerString(a.Raw.Career, b.Raw.Career)
	merged.Raw.Education = longerString(a.Raw.Education, b.Raw.Education)
	merged.Raw.JobType = longerString(a.Raw.JobType, b.Raw.JobType)
	merged.Raw.Salary = longerString(a.Raw.Salary, b.Raw.Salary)
	merged.Raw.Location = longerString(a.Raw.Location, b.Raw.Location)
	merged.Raw.Deadline = longerString(a.Raw.Deadline, b.Raw.Deadline)
	merged.Raw.MainTasks = longerString(a.Raw.MainTasks, b.Raw.MainTasks)
	merged.Raw.Requirements = longerString(a.Raw.Requirements, b.Raw.Requirements)
	merged.Raw.Preferred = longerString(a.Raw.Preferred, b.Raw.Preferred)
	merged.Raw.SourceURL = longerString(a.Raw.SourceURL, b.Raw.SourceURL)
	merged.Raw.Source = longerString(a.Raw.Source, b.Raw.Source)

	// Merge skills — union
	skillSet := make(map[string]bool)
	for _, s := range a.Raw.Skills {
		skillSet[s] = true
	}
	for _, s := range b.Raw.Skills {
		skillSet[s] = true
	}
	merged.Raw.Skills = make([]string, 0, len(skillSet))
	for s := range skillSet {
		merged.Raw.Skills = append(merged.Raw.Skills, s)
	}

	// Merge SectionTexts — pick longer for each key
	allKeys := make(map[string]bool)
	for k := range a.SectionTexts {
		allKeys[k] = true
	}
	for k := range b.SectionTexts {
		allKeys[k] = true
	}
	for k := range allKeys {
		merged.SectionTexts[k] = longerString(a.SectionTexts[k], b.SectionTexts[k])
	}

	merged.FilledFields = countFilledFields(&merged.Raw)
	return merged
}

func longerString(a, b string) string {
	if len(a) >= len(b) {
		return a
	}
	return b
}

// BuildNormalizeInput constructs a structured text for the normalize LLM prompt
// from extracted data. Does NOT truncate — passes full section content.
func BuildNormalizeInput(extracted *ExtractionResult) string {
	var b strings.Builder

	// Metadata fields
	if extracted.Raw.CompanyName != "" {
		fmt.Fprintf(&b, "Company: %s\n", extracted.Raw.CompanyName)
	}
	if extracted.Raw.Position != "" {
		fmt.Fprintf(&b, "Position: %s\n", extracted.Raw.Position)
	}
	if extracted.Raw.Department != "" {
		fmt.Fprintf(&b, "Department: %s\n", extracted.Raw.Department)
	}
	if extracted.Raw.Career != "" {
		fmt.Fprintf(&b, "Career: %s\n", extracted.Raw.Career)
	}
	if extracted.Raw.Education != "" {
		fmt.Fprintf(&b, "Education: %s\n", extracted.Raw.Education)
	}
	if extracted.Raw.JobType != "" {
		fmt.Fprintf(&b, "JobType: %s\n", extracted.Raw.JobType)
	}
	if extracted.Raw.Location != "" {
		fmt.Fprintf(&b, "Location: %s\n", extracted.Raw.Location)
	}
	if extracted.Raw.Salary != "" {
		fmt.Fprintf(&b, "Salary: %s\n", extracted.Raw.Salary)
	}
	if extracted.Raw.Deadline != "" {
		fmt.Fprintf(&b, "Deadline: %s\n", extracted.Raw.Deadline)
	}

	// Section texts (full content, not truncated)
	sectionOrder := []struct {
		key   string
		label string
	}{
		{"main_tasks", "담당업무"},
		{"requirements", "자격요건"},
		{"preferred", "우대사항"},
		{"skills", "기술스택"},
		{"benefits", "복리후생"},
		{"company_intro", "회사소개"},
	}

	for _, sec := range sectionOrder {
		text := extracted.SectionTexts[sec.key]
		// Also check RawJobPosting fields as fallback
		if text == "" {
			switch sec.key {
			case "main_tasks":
				text = extracted.Raw.MainTasks
			case "requirements":
				text = extracted.Raw.Requirements
			case "preferred":
				text = extracted.Raw.Preferred
			}
		}
		if text != "" {
			fmt.Fprintf(&b, "\n--- %s ---\n%s\n", sec.label, text)
		}
	}

	// Skills as comma-separated list if not already in sections
	if _, hasSec := extracted.SectionTexts["skills"]; !hasSec && len(extracted.Raw.Skills) > 0 {
		fmt.Fprintf(&b, "\n--- 기술스택 ---\n%s\n", strings.Join(extracted.Raw.Skills, ", "))
	}

	return strings.TrimSpace(b.String())
}

// SmartTrimForLLM trims markdown content intelligently for LLM input.
// If section texts are available, concatenates them by priority instead of blind truncation.
// Falls back to simple truncation if no sections were detected.
func SmartTrimForLLM(markdown string, extracted *ExtractionResult, maxLen int) string {
	if extracted == nil || len(extracted.SectionTexts) == 0 {
		// No sections — fall back to simple truncation
		if len(markdown) <= maxLen {
			return markdown
		}
		return markdown[:maxLen]
	}

	// Build content from sections in priority order
	var b strings.Builder

	// Add metadata header
	if extracted.Raw.CompanyName != "" || extracted.Raw.Position != "" {
		if extracted.Raw.CompanyName != "" {
			fmt.Fprintf(&b, "Company: %s\n", extracted.Raw.CompanyName)
		}
		if extracted.Raw.Position != "" {
			fmt.Fprintf(&b, "Position: %s\n", extracted.Raw.Position)
		}
		b.WriteString("\n")
	}

	// Priority order for sections
	prioritySections := []string{"main_tasks", "requirements", "preferred", "skills", "benefits", "company_intro"}

	for _, key := range prioritySections {
		text, ok := extracted.SectionTexts[key]
		if !ok || text == "" {
			continue
		}

		// Check if adding this section would exceed the limit
		sectionHeader := sectionKeyToLabel(key)
		sectionBlock := fmt.Sprintf("--- %s ---\n%s\n\n", sectionHeader, text)

		if b.Len()+len(sectionBlock) > maxLen {
			// Truncate this section to fit remaining space
			remaining := maxLen - b.Len() - len(sectionHeader) - 10 // overhead for "--- xxx ---\n"
			if remaining > 50 {
				fmt.Fprintf(&b, "--- %s ---\n%s\n\n", sectionHeader, text[:min(remaining, len(text))])
			}
			break
		}

		b.WriteString(sectionBlock)
	}

	result := strings.TrimSpace(b.String())
	if result == "" {
		// Fallback if section assembly produced nothing
		if len(markdown) <= maxLen {
			return markdown
		}
		return markdown[:maxLen]
	}

	return result
}

func sectionKeyToLabel(key string) string {
	labels := map[string]string{
		"main_tasks":    "담당업무",
		"requirements":  "자격요건",
		"preferred":     "우대사항",
		"skills":        "기술스택",
		"benefits":      "복리후생",
		"company_intro": "회사소개",
	}
	if label, ok := labels[key]; ok {
		return label
	}
	return key
}
