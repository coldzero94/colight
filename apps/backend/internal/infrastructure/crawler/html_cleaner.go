package crawler

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// noiseSelectors are HTML elements/selectors stripped before readability extraction.
// Inspired by Firecrawl's transform_html() noise removal, adapted for Korean job sites.
var noiseSelectors = []string{
	// Structural navigation
	"header",
	"footer",
	"nav",
	"aside",

	// Scripts, styles, metadata
	"script",
	"style",
	"noscript",
	"iframe",
	"object",
	"embed",
	"link[rel='stylesheet']",

	// Advertising
	".ad",
	".ads",
	".advertisement",
	"[class*='advert']",
	"[class*='banner']",
	"[id*='advert']",
	"[id*='banner']",

	// UI chrome — modals, popups, overlays
	".modal",
	".popup",
	".overlay",
	".dialog",
	"[class*='modal']",
	"[class*='popup']",
	"[role='dialog']",
	"[role='alertdialog']",

	// Cookie / consent banners
	"[class*='cookie']",
	"[class*='consent']",
	"[id*='cookie']",
	"[id*='consent']",
	"[class*='gdpr']",

	// Navigation / sidebar / menu
	".sidebar",
	".sidenav",
	".menu",
	".breadcrumb",
	".pagination",
	".paging",
	"[class*='sidebar']",
	"[class*='sidenav']",
	"[class*='breadcrumb']",
	"[class*='pagination']",
	"[id*='sidebar']",
	"[role='navigation']",
	"[role='banner']",
	"[role='contentinfo']",

	// Social / sharing buttons
	".share-buttons",
	".social-links",
	"[class*='social']",
	"[class*='sns']",

	// Comments
	"[class*='comment']",
	"[id*='comment']",

	// Related / recommended content
	"[class*='related']",
	"[class*='recommend']",
	"[class*='similar']",

	// Hidden elements
	"[hidden]",
	"[aria-hidden='true']",
	"[style*='display: none']",
	"[style*='display:none']",
	"[style*='visibility: hidden']",
	"[style*='visibility:hidden']",

	// Korean job site specific noise
	"[class*='GnbWrap']",
	"[class*='SideBarWrap']",
	"[class*='BtnApply']",
	"[class*='BtnScrap']",
	"[class*='BtnShare']",
	"[class*='recruitLnb']",
	"#content_top_banner",
	"#recruit_aside",
	".wrap_tit_recruit_company",
}

// mainContentSelectors are tried in order when readability fails.
// Returns the inner HTML of the first match with 50+ characters.
var mainContentSelectors = []string{
	// Semantic HTML5
	"article",
	"main",
	"[role='main']",

	// Common content IDs
	"#content",
	"#main-content",
	"#job-content",
	"#recruit_content",
	"#container",

	// Common content classes
	".content",
	".main-content",
	".job-content",
	".job-posting",
	".recruit-content",

	// Korean job site specific containers
	".artReadJobWrap",      // JobKorea
	".wrap_jv_cont",        // Saramin
	".job_description",     // Wanted
	".section_detail_group", // Programmers
}

// CleanHTML removes noise elements from HTML before readability extraction.
// Uses goquery to strip elements matching noiseSelectors.
// Returns the original HTML on parse errors (graceful fallback).
func CleanHTML(rawHTML string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML, err
	}

	for _, selector := range noiseSelectors {
		doc.Find(selector).Remove()
	}

	result, err := doc.Html()
	if err != nil {
		return rawHTML, err
	}

	return result, nil
}

// ExtractMainContent tries to extract main content using fallback selectors
// when readability gives too little output. Returns the inner HTML of the
// first matching selector with 50+ characters, or empty string if none match.
func ExtractMainContent(rawHTML string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return "", err
	}

	for _, selector := range mainContentSelectors {
		sel := doc.Find(selector).First()
		if sel.Length() > 0 {
			html, err := sel.Html()
			if err != nil {
				continue
			}
			if len(strings.TrimSpace(html)) > 50 {
				return html, nil
			}
		}
	}

	return "", nil
}

// ResolveURLs converts relative URLs (href, src) to absolute URLs.
// Non-HTTP schemes (data:, mailto:, tel:, javascript:, blob:) are left unchanged.
func ResolveURLs(htmlContent string, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return htmlContent
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	doc.Find("[href]").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists && href != "" {
			if resolved := resolveURL(base, href); resolved != "" {
				s.SetAttr("href", resolved)
			}
		}
	})

	doc.Find("[src]").Each(func(_ int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists && src != "" {
			if resolved := resolveURL(base, src); resolved != "" {
				s.SetAttr("src", resolved)
			}
		}
	})

	result, err := doc.Html()
	if err != nil {
		return htmlContent
	}

	return result
}

// resolveURL resolves a potentially relative URL against a base URL.
func resolveURL(base *url.URL, rawURL string) string {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	for _, prefix := range []string{"mailto:", "tel:", "javascript:", "data:", "blob:"} {
		if strings.HasPrefix(rawURL, prefix) {
			return rawURL
		}
	}
	ref, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return base.ResolveReference(ref).String()
}
