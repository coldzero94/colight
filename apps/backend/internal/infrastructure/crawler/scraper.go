package crawler

import (
	"regexp"
	"strings"

	readability "github.com/go-shiori/go-readability"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// MinMarkdownLength is the minimum length of markdown content to be considered usable
const MinMarkdownLength = 100

// ScrapeToMarkdown extracts the main content from raw HTML and converts it to markdown.
// Pipeline: pre-clean → resolve URLs → readability → fallback extraction → markdown → post-processing.
// Returns empty string if the HTML is empty or cannot be parsed.
func ScrapeToMarkdown(rawHTML string, sourceURL string) (string, error) {
	if strings.TrimSpace(rawHTML) == "" {
		return "", nil
	}

	// Step 1: Pre-clean HTML (remove noise elements like nav, header, footer, ads)
	cleaned, _ := CleanHTML(rawHTML)

	// Step 2: Resolve relative URLs to absolute
	cleaned = ResolveURLs(cleaned, sourceURL)

	// Step 3: Try readability on cleaned HTML
	var readabilityHTML string
	article, err := readability.FromReader(strings.NewReader(cleaned), nil)
	if err == nil {
		readabilityHTML = strings.TrimSpace(article.Content)
	}

	// Step 4: Try fallback main content selectors
	fallbackHTML, _ := ExtractMainContent(cleaned)

	// Step 5: Use whichever extraction gave more content
	contentHTML := readabilityHTML
	if len(fallbackHTML) > len(readabilityHTML) {
		contentHTML = fallbackHTML
	}

	if strings.TrimSpace(contentHTML) == "" {
		return "", nil
	}

	// Step 5: Convert HTML to markdown
	md, err := htmltomarkdown.ConvertString(contentHTML)
	if err != nil {
		return "", nil
	}

	// Step 6: Post-process markdown
	md = PostProcessMarkdown(md)

	return md, nil
}

// Markdown post-processing patterns
var (
	reMultipleBlankLines = regexp.MustCompile(`\n{3,}`)
	reEmptyLinks         = regexp.MustCompile(`\[\s*\]\([^)]*\)`)
	reEmptyImages        = regexp.MustCompile(`!\[\s*\]\([^)]*\)`)
	reNavBreadcrumb      = regexp.MustCompile(`(?m)^[\w가-힣\s]+(?:\s*[>»|]\s*[\w가-힣\s]+){2,}\s*$`)
	reBareURL            = regexp.MustCompile(`(?m)^https?://\S+\s*$`)
	reMultipleSpaces     = regexp.MustCompile(`[ \t]{2,}`)
)

// PostProcessMarkdown cleans markdown output by removing noise artifacts.
func PostProcessMarkdown(md string) string {
	// Remove empty links and images
	md = reEmptyLinks.ReplaceAllString(md, "")
	md = reEmptyImages.ReplaceAllString(md, "")

	// Remove breadcrumb navigation lines
	md = reNavBreadcrumb.ReplaceAllString(md, "")

	// Remove bare URLs on their own lines
	md = reBareURL.ReplaceAllString(md, "")

	// Collapse multiple spaces within lines
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		lines[i] = reMultipleSpaces.ReplaceAllString(line, " ")
	}
	md = strings.Join(lines, "\n")

	// Collapse multiple blank lines
	md = reMultipleBlankLines.ReplaceAllString(md, "\n\n")

	return strings.TrimSpace(md)
}
