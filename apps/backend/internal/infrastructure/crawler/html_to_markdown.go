package crawler

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// selectionToMarkdown converts a goquery Selection's inner HTML to clean markdown.
// Preserves bullet lists, numbered lists, bold/emphasis, line breaks, and paragraph structure.
// Falls back to plain .Text() if HTML-to-markdown conversion fails.
func selectionToMarkdown(sel *goquery.Selection) string {
	if sel.Length() == 0 {
		return ""
	}

	innerHTML, err := sel.Html()
	if err != nil || strings.TrimSpace(innerHTML) == "" {
		return strings.TrimSpace(sel.Text())
	}

	md, err := htmltomarkdown.ConvertString(innerHTML)
	if err != nil {
		return strings.TrimSpace(sel.Text())
	}

	return strings.TrimSpace(md)
}
