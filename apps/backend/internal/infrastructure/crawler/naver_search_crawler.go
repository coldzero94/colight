package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const maxCompanyContextLen = 4000

// NaverSearchCrawler searches Naver web for company information
type NaverSearchCrawler struct {
	httpClient *http.Client
}

// NewNaverSearchCrawler creates a new Naver search crawler
func NewNaverSearchCrawler() *NaverSearchCrawler {
	return &NaverSearchCrawler{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SearchCompanyInfo searches Naver for company information and returns
// aggregated markdown content from top search results.
// Returns empty string (not error) if nothing is found.
func (n *NaverSearchCrawler) SearchCompanyInfo(companyName string) (string, error) {
	if strings.TrimSpace(companyName) == "" {
		return "", nil
	}

	// 1. Search Naver for company info
	searchURL := fmt.Sprintf("https://search.naver.com/search.naver?query=%s",
		url.QueryEscape(companyName+" 기업정보"))

	searchHTML, err := n.fetchHTML(searchURL)
	if err != nil {
		return "", nil // graceful: search failure is not an error
	}

	// 2. Extract top result links (exclude news links)
	links := parseSearchResultLinks(strings.NewReader(searchHTML), 2)
	if len(links) == 0 {
		return "", nil
	}

	// 3. Scrape each link and collect markdown
	var parts []string
	for _, link := range links {
		html, fetchErr := n.fetchHTML(link)
		if fetchErr != nil {
			continue
		}

		md, scrapeErr := ScrapeToMarkdown(html, link)
		if scrapeErr != nil || len(md) < MinMarkdownLength {
			continue
		}

		parts = append(parts, md)
	}

	if len(parts) == 0 {
		return "", nil
	}

	// 4. Combine and truncate
	combined := strings.Join(parts, "\n\n---\n\n")
	if len(combined) > maxCompanyContextLen {
		combined = combined[:maxCompanyContextLen]
	}

	return combined, nil
}

// fetchHTML fetches HTML from URL with browser-like headers
func (n *NaverSearchCrawler) fetchHTML(targetURL string) (string, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseSearchResultLinks extracts relevant URLs from Naver search results HTML.
// It filters out news links, Naver internal links, and returns up to maxLinks external URLs.
func parseSearchResultLinks(reader io.Reader, maxLinks int) []string {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil
	}

	var links []string
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if len(links) >= maxLinks {
			return
		}

		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Must be absolute HTTP(S) URL
		if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
			return
		}

		// Skip Naver internal pages and news
		if isExcludedDomain(href) {
			return
		}

		// Skip already-seen links
		if seen[href] {
			return
		}

		// Only include links with meaningful anchor text
		text := strings.TrimSpace(s.Text())
		if len(text) < 5 {
			return
		}

		seen[href] = true
		links = append(links, href)
	})

	return links
}

// isExcludedDomain returns true if the URL should be excluded from search results
func isExcludedDomain(href string) bool {
	excludedPatterns := []string{
		"news.naver.com",
		"n.news.naver.com",
		"search.naver.com",
		"naver.com/search",
		"login.naver.com",
		"help.naver.com",
		"ad.naver.com",
		"section.blog.naver.com",
	}
	for _, pattern := range excludedPatterns {
		if strings.Contains(href, pattern) {
			return true
		}
	}
	return false
}
