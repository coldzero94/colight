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

// NaverNewsCrawler crawls Naver news search for company news
type NaverNewsCrawler struct {
	httpClient *http.Client
}

// NewNaverNewsCrawler creates a new Naver news crawler
func NewNaverNewsCrawler() *NaverNewsCrawler {
	return &NaverNewsCrawler{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewsArticle represents a news article
type NewsArticle struct {
	Title       string
	Link        string
	Description string
	PubDate     string
	Source      string
}

// SearchCompanyNews searches for recent news about a company
func (n *NaverNewsCrawler) SearchCompanyNews(companyName string, maxResults int) ([]NewsArticle, error) {
	if maxResults == 0 {
		maxResults = 5
	}

	// Naver news search URL
	searchURL := fmt.Sprintf("https://search.naver.com/search.naver?where=news&query=%s&sort=1",
		url.QueryEscape(companyName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.naver.com")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var articles []NewsArticle

	// Parse news articles from search results
	// Naver news search results have specific selectors
	doc.Find(".news_area, .list_news li").Each(func(i int, s *goquery.Selection) {
		if i >= maxResults {
			return
		}

		article := NewsArticle{}

		// Extract title and link
		titleLink := s.Find("a.news_tit, a.link_txt")
		article.Title = strings.TrimSpace(titleLink.Text())
		article.Link, _ = titleLink.Attr("href")

		// Extract description
		article.Description = strings.TrimSpace(s.Find(".news_dsc, .dsc_txt_wrap").First().Text())

		// Extract source and date
		infoArea := s.Find(".info_group, .news_info")
		article.Source = strings.TrimSpace(infoArea.Find(".press, .info.press").First().Text())
		article.PubDate = strings.TrimSpace(infoArea.Find(".info").Last().Text())

		if article.Title != "" {
			articles = append(articles, article)
		}
	})

	return articles, nil
}
