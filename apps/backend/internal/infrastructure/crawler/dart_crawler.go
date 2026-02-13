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

// DartCrawler crawls DART e-disclosure system for company information
type DartCrawler struct {
	httpClient *http.Client
}

// NewDartCrawler creates a new DART crawler
func NewDartCrawler() *DartCrawler {
	return &DartCrawler{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CompanyInfo represents basic company information from DART
type CompanyInfo struct {
	CorpName    string // 회사명
	CorpCode    string // 기업 코드
	StockCode   string // 주식 코드
	CEO         string // 대표이사
	CorpCls     string // 법인구분
	JurirNo     string // 법인등록번호
	BizrNo      string // 사업자등록번호
	Address     string // 주소
	HomPage     string // 홈페이지
	IrHomPage   string // IR 홈페이지
	PhoneNumber string // 전화번호
	FaxNumber   string // 팩스번호
	Industry    string // 업종
	EstDate     string // 설립일
	AccMonth    string // 결산월
}

// SearchCompany searches for a company by name on DART website
func (d *DartCrawler) SearchCompany(companyName string) (*CompanyInfo, error) {
	// DART 회사 검색 URL (공개 검색 페이지)
	searchURL := fmt.Sprintf("https://dart.fss.or.kr/dsac001/search.ax?textCrpNm=%s",
		url.QueryEscape(companyName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://dart.fss.or.kr")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DART HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	// Parse company info from search results
	info := &CompanyInfo{}

	// Extract from first result row
	firstRow := doc.Find("table.result-table tbody tr").First()
	if firstRow.Length() == 0 {
		return nil, fmt.Errorf("company not found: %s", companyName)
	}

	info.CorpName = strings.TrimSpace(firstRow.Find("td").Eq(0).Text())
	info.StockCode = strings.TrimSpace(firstRow.Find("td").Eq(1).Text())
	info.CorpCode = strings.TrimSpace(firstRow.Find("td").Eq(2).Text())
	info.CEO = strings.TrimSpace(firstRow.Find("td").Eq(3).Text())

	// For detailed info, we would need to make another request to the detail page
	// For MVP, basic info from search is sufficient

	return info, nil
}

// GetCompanyBasicInfo gets basic info by crawling DART detail page
func (d *DartCrawler) GetCompanyBasicInfo(corpCode string) (*CompanyInfo, error) {
	// DART 기업 개황 페이지
	detailURL := fmt.Sprintf("https://dart.fss.or.kr/dsaf001/main.do?rcpNo=%s", corpCode)

	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := d.httpClient.Do(req)
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

	info := &CompanyInfo{
		CorpCode: corpCode,
	}

	// Parse company details from page
	// Note: Actual selectors depend on DART's HTML structure
	info.CorpName = strings.TrimSpace(doc.Find(".company-name, h1.corp-name").First().Text())
	info.CEO = strings.TrimSpace(doc.Find(".ceo-name, .representative").First().Text())
	info.Address = strings.TrimSpace(doc.Find(".address").First().Text())
	info.Industry = strings.TrimSpace(doc.Find(".industry").First().Text())

	return info, nil
}
