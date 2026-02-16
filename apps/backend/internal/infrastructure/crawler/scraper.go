package crawler

import (
	"strings"

	readability "github.com/go-shiori/go-readability"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// MinMarkdownLength is the minimum length of markdown content to be considered usable
const MinMarkdownLength = 100

// ScrapeToMarkdown extracts the main content from raw HTML and converts it to markdown.
// It uses Mozilla's Readability algorithm to extract the main content, then converts to markdown.
// Returns empty string if the HTML is empty or cannot be parsed.
func ScrapeToMarkdown(rawHTML string, sourceURL string) (string, error) {
	if strings.TrimSpace(rawHTML) == "" {
		return "", nil
	}

	// Step 1: Extract main content with Readability
	article, err := readability.FromReader(strings.NewReader(rawHTML), nil)
	if err != nil {
		// If readability fails, return empty — caller should fall back to raw HTML
		return "", nil
	}

	if strings.TrimSpace(article.Content) == "" {
		return "", nil
	}

	// Step 2: Convert extracted HTML to markdown
	md, err := htmltomarkdown.ConvertString(article.Content)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(md), nil
}
