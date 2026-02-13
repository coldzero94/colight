# 크롤러 모니터링

> 크롤링 대상 사이트 모니터링, HTML 구조 변경 감지, 실패 대응, 관리자 대시보드
>
> 관련 문서: `04-ai-pipeline.md` (크롤링 파이프라인), `phases/phase-3-crawling-parsing.md` (크롤링 구현), `11-logging-monitoring.md` (로깅 패턴), `09-error-handling.md` (에러 코드)

---

## 1. 크롤러 대상 사이트 목록

| 사이트 | 도메인 | 파싱 방식 | 구현 Phase | 상태 |
|--------|--------|----------|-----------|------|
| **잡코리아** | `jobkorea.co.kr` | goquery (정적 HTML) | Phase 3 | 메인 |
| **캐치** | `catch.co.kr` | goquery (정적 HTML) | Phase 3 | 메인 |
| **사람인** | `saramin.co.kr` | 공식 API (oAPI) | Phase 3 | API 우선 |
| **원티드** | `wanted.co.kr` | Playwright + River (SPA) | Phase 10 | 동적 |
| **기타** | - | AI fallback (경량 모델 Gemini/Groq) | Phase 3 | 폴백 |

### 사이트별 URL 패턴

```go
// apps/backend/internal/infrastructure/crawler/domains.go
package crawler

var DomainPatterns = map[string]SiteConfig{
    "jobkorea.co.kr": {
        Name:       "잡코리아",
        Method:     MethodGoquery,
        URLPattern: `^https?://(www\.)?jobkorea\.co\.kr/Recruit/GI_Read/\d+`,
        Selectors:  JobKoreaSelectors,
    },
    "catch.co.kr": {
        Name:       "캐치",
        Method:     MethodGoquery,
        URLPattern: `^https?://(www\.)?catch\.co\.kr/Comp/RecruitInfo/`,
        Selectors:  CatchSelectors,
    },
    "saramin.co.kr": {
        Name:       "사람인",
        Method:     MethodAPI,
        URLPattern: `^https?://(www\.)?saramin\.co\.kr/zf_user/jobs/relay/view`,
    },
    "wanted.co.kr": {
        Name:       "원티드",
        Method:     MethodPlaywright,
        URLPattern: `^https?://(www\.)?wanted\.co\.kr/wd/\d+`,
    },
}
```

---

## 2. 모니터링 지표

### 2.1 핵심 지표 (Metrics)

| 지표 | 설명 | 경고 임계값 | 위험 임계값 |
|------|------|-----------|-----------|
| **성공률** (success_rate) | 파싱 성공 / 전체 요청 | < 90% | < 70% |
| **응답 시간** (latency_ms) | HTML fetch + 파싱 소요 시간 | > 5초 (P95) | > 10초 (P95) |
| **필수 필드 존재율** (field_coverage) | 필수 필드(회사명, 포지션) 추출 성공률 | < 95% | < 80% |
| **AI 폴백 비율** (ai_fallback_rate) | AI fallback 사용 비율 | > 20% | > 50% |
| **HTTP 에러율** (http_error_rate) | 403, 404, 5xx 응답 비율 | > 10% | > 30% |

### 2.2 로깅 패턴

> `11-logging-monitoring.md`의 slog 구조화 로깅 패턴을 따른다.

```go
// apps/backend/internal/infrastructure/crawler/logger.go
package crawler

import (
    "context"
    "log/slog"
    "time"
)

type CrawlResult struct {
    TraceID        string        `json:"trace_id"`
    Domain         string        `json:"domain"`
    URL            string        `json:"url"`
    Method         string        `json:"method"`   // goquery, playwright, api, ai_fallback
    HTTPStatus     int           `json:"http_status"`
    Success        bool          `json:"success"`
    Duration       time.Duration `json:"duration"`
    FieldsCaptured int           `json:"fields_captured"` // Number of non-empty fields
    FieldsExpected int           `json:"fields_expected"` // Total expected fields
    AIFallbackUsed bool          `json:"ai_fallback_used"`
    ErrorCode      string        `json:"error_code,omitempty"` // CRAWL_001, etc.
    ErrorMessage   string        `json:"error_message,omitempty"`
}

func LogCrawlResult(ctx context.Context, logger *slog.Logger, result CrawlResult) {
    level := slog.LevelInfo
    if !result.Success {
        level = slog.LevelWarn
    }

    logger.LogAttrs(ctx, level, "crawl_result",
        slog.String("trace_id", result.TraceID),
        slog.String("domain", result.Domain),
        slog.String("method", result.Method),
        slog.Int("http_status", result.HTTPStatus),
        slog.Bool("success", result.Success),
        slog.Duration("duration", result.Duration),
        slog.Int("fields_captured", result.FieldsCaptured),
        slog.Int("fields_expected", result.FieldsExpected),
        slog.Bool("ai_fallback", result.AIFallbackUsed),
        slog.String("error_code", result.ErrorCode),
    )
}
```

---

## 3. HTML 구조 변경 감지

> 채용 사이트는 UI 리뉴얼, A/B 테스트 등으로 HTML 구조가 자주 변경된다. 구조 변경을 빠르게 감지하여 파서를 업데이트해야 사용자 경험 중단을 방지할 수 있다.

### 3.1 CSS Selector 검증

각 사이트의 필수 CSS selector가 여전히 유효한지 주기적으로 검증한다.

```go
// apps/backend/internal/infrastructure/crawler/health_checker.go
package crawler

import (
    "context"
    "log/slog"
    "time"
)

type SelectorCheck struct {
    Selector string
    Field    string
    Required bool
}

// Site-specific expected selectors
var JobKoreaSelectors = []SelectorCheck{
    {Selector: ".company-name a, .coName", Field: "company_name", Required: true},
    {Selector: ".artReadJobTitle, .title-wrap h1", Field: "position", Required: true},
    {Selector: ".tbRow .career", Field: "career", Required: false},
    {Selector: ".tbRow .education", Field: "education", Required: false},
    {Selector: ".date .tahoma", Field: "deadline", Required: false},
    {Selector: ".artReadJobSecCont", Field: "job_description", Required: true},
    {Selector: ".skillWrap .skill", Field: "skills", Required: false},
}

var CatchSelectors = []SelectorCheck{
    {Selector: ".company-name, .comp-name", Field: "company_name", Required: true},
    {Selector: ".recruit-title, .job-title", Field: "position", Required: true},
    {Selector: ".recruit-info .career", Field: "career", Required: false},
    {Selector: ".job-description .task-section", Field: "main_tasks", Required: false},
    {Selector: ".recruit-info .deadline, .dday", Field: "deadline", Required: false},
}

type HealthCheckResult struct {
    Domain          string
    URL             string
    CheckedAt       time.Time
    TotalSelectors  int
    FoundSelectors  int
    MissingRequired []string  // Required selectors that were not found
    MissingOptional []string  // Optional selectors not found
    HTTPStatus      int
    LatencyMS       int64
    Healthy         bool
    Score           float64   // 0.0 ~ 1.0 (found / total)
}

func (c *Crawler) CheckSiteHealth(ctx context.Context, domain string, sampleURL string) (*HealthCheckResult, error) {
    start := time.Now()

    // 1. Fetch sample page
    doc, httpStatus, err := c.fetchHTML(ctx, sampleURL)
    if err != nil {
        return &HealthCheckResult{
            Domain:     domain,
            URL:        sampleURL,
            CheckedAt:  time.Now(),
            HTTPStatus: httpStatus,
            Healthy:    false,
            Score:      0,
        }, nil
    }

    // 2. Check each selector
    selectors := c.getSelectors(domain)
    result := &HealthCheckResult{
        Domain:         domain,
        URL:            sampleURL,
        CheckedAt:      time.Now(),
        TotalSelectors: len(selectors),
        HTTPStatus:     httpStatus,
        LatencyMS:      time.Since(start).Milliseconds(),
    }

    for _, sel := range selectors {
        found := doc.Find(sel.Selector).Length() > 0
        if found {
            result.FoundSelectors++
        } else if sel.Required {
            result.MissingRequired = append(result.MissingRequired, sel.Field)
        } else {
            result.MissingOptional = append(result.MissingOptional, sel.Field)
        }
    }

    result.Score = float64(result.FoundSelectors) / float64(result.TotalSelectors)
    result.Healthy = len(result.MissingRequired) == 0 && result.Score >= 0.5

    return result, nil
}
```

### 3.2 파싱 결과 품질 점수

실제 파싱 결과의 필수 필드 존재 여부로 품질 점수를 계산한다.

```go
// apps/backend/internal/infrastructure/crawler/quality.go
package crawler

type QualityScore struct {
    Total            float64 // 0.0 ~ 1.0
    RequiredPresent  int
    RequiredTotal    int
    OptionalPresent  int
    OptionalTotal    int
    MissingFields    []string
}

func CalculateQuality(posting *RawJobPosting) QualityScore {
    required := map[string]string{
        "company_name": posting.CompanyName,
        "position":     posting.Position,
    }

    optional := map[string]string{
        "department":   posting.Department,
        "career":       posting.Career,
        "education":    posting.Education,
        "job_type":     posting.JobType,
        "location":     posting.Location,
        "deadline":     posting.Deadline,
        "main_tasks":   posting.MainTasks,
        "requirements": posting.Requirements,
        "preferred":    posting.Preferred,
    }

    score := QualityScore{
        RequiredTotal: len(required),
        OptionalTotal: len(optional),
    }

    for field, value := range required {
        if value != "" {
            score.RequiredPresent++
        } else {
            score.MissingFields = append(score.MissingFields, field)
        }
    }

    for field, value := range optional {
        if value != "" {
            score.OptionalPresent++
        } else {
            score.MissingFields = append(score.MissingFields, field)
        }
    }

    // Required fields are weighted 3x
    totalWeight := float64(score.RequiredTotal*3 + score.OptionalTotal)
    achievedWeight := float64(score.RequiredPresent*3 + score.OptionalPresent)
    score.Total = achievedWeight / totalWeight

    return score
}
```

### 3.3 변경 감지 시 알림

```
[헬스체크 결과 분석]
    │
    ├── 모든 필수 selector 존재 + Score >= 0.7
    │     → 정상 (INFO 로그)
    │
    ├── 필수 selector 1개 이상 누락 또는 Score < 0.7
    │     → 경고 (WARN 로그 + Slack 알림)
    │     → 해당 사이트 AI fallback 모드 자동 전환
    │
    └── HTTP 에러 (403, 503) 또는 Score < 0.3
          → 위험 (ERROR 로그 + Slack 알림)
          → 해당 사이트 파싱 일시 중단 + 수동 입력 폼 유도
```

```go
// apps/backend/internal/infrastructure/crawler/alerter.go
package crawler

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type SlackAlerter struct {
    webhookURL string
}

func (a *SlackAlerter) SendAlert(result *HealthCheckResult) error {
    if a.webhookURL == "" {
        return nil // Slack webhook not configured
    }

    emoji := ":warning:"
    color := "#FF9900" // warning yellow
    if len(result.MissingRequired) > 0 {
        emoji = ":rotating_light:"
        color = "#FF0000" // danger red
    }

    payload := map[string]interface{}{
        "text": emoji + " Crawler Health Alert",
        "attachments": []map[string]interface{}{
            {
                "color": color,
                "fields": []map[string]string{
                    {"title": "Site", "value": result.Domain, "short": "true"},
                    {"title": "Score", "value": fmt.Sprintf("%.0f%%", result.Score*100), "short": "true"},
                    {"title": "HTTP Status", "value": fmt.Sprintf("%d", result.HTTPStatus), "short": "true"},
                    {"title": "Latency", "value": fmt.Sprintf("%dms", result.LatencyMS), "short": "true"},
                    {"title": "Missing Required", "value": strings.Join(result.MissingRequired, ", ")},
                    {"title": "Missing Optional", "value": strings.Join(result.MissingOptional, ", ")},
                },
            },
        },
    }

    body, _ := json.Marshal(payload)
    _, err := http.Post(a.webhookURL, "application/json", bytes.NewReader(body))
    return err
}
```

---

## 4. 크롤링 실패 대응 플로우

### 4.1 전체 대응 흐름

> `06-error-scenarios-ux.md`의 URL 파싱 처리 흐름과 연계

```
[크롤링 요청]
    │
    ▼
[1차: 사이트 전용 파서 (goquery/API)]
    │
    ├── 성공 → 품질 점수 계산
    │     ├── Score >= 0.5 → 정상 반환
    │     └── Score < 0.5  → 2차 AI 폴백 시도
    │
    └── 실패
          │
          ▼
[2차: 자동 재시도 (exponential backoff)]
    │
    ├── 재시도 정책:
    │     ├── 1차: 1초 대기 후 재시도
    │     ├── 2차: 3초 대기 후 재시도
    │     └── 3차: 없음 (폴백 전환)
    │
    ├── 재시도 대상:
    │     ├── 5xx 에러 (서버 일시 장애)
    │     ├── 타임아웃 (10초 초과)
    │     └── 일시적 네트워크 오류
    │
    └── 재시도 비대상:
          ├── 403 Forbidden (차단)
          ├── 404 Not Found (페이지 없음)
          └── 파싱 성공했지만 필수 필드 없음 (구조 변경)
          │
          ▼
[3차: AI Fallback (경량 모델 Gemini/Groq)]
    │
    ├── raw HTML을 경량 모델 (Gemini/Groq)에 전달
    │     ├── HTML body만 추출 (max 8,000자)
    │     └── JobPosting 스키마로 구조화 요청
    │
    ├── 성공 → 결과에 "AI 자동 분석 결과입니다" 배너 표시
    │     → 비용: ~15원/건 (일반 파싱 대비 3배)
    │
    └── 실패
          │
          ▼
[4차: 수동 입력 안내]
    │
    └── 사용자에게 수동 입력 폼 제공
          └── "공고 내용을 직접 붙여넣어 주세요"
```

### 4.2 Go 구현: 재시도 + 폴백

```go
// apps/backend/internal/infrastructure/crawler/crawler.go
package crawler

import (
    "context"
    "log/slog"
    "time"
)

type Crawler struct {
    logger  *slog.Logger
    ai      *ai.AIProvider
    alerter *SlackAlerter
}

func (c *Crawler) Parse(ctx context.Context, url string) (*JobPosting, error) {
    domain := detectDomain(url)

    // 1. Try site-specific parser with retry
    var rawPosting *RawJobPosting
    var parseErr error

    for attempt := 0; attempt < 3; attempt++ {
        if attempt > 0 {
            delay := time.Duration(1<<uint(attempt-1)) * time.Second // 1s, 2s
            time.Sleep(delay)
            c.logger.Info("crawl retry",
                slog.String("url", url),
                slog.Int("attempt", attempt+1),
            )
        }

        rawPosting, parseErr = c.parseSite(ctx, domain, url)
        if parseErr == nil {
            break
        }

        // Don't retry non-retryable errors
        if isNonRetryable(parseErr) {
            break
        }
    }

    // 2. Check quality if parsing succeeded
    if parseErr == nil && rawPosting != nil {
        quality := CalculateQuality(rawPosting)
        if quality.Total >= 0.5 {
            // Good enough, normalize with AI
            return c.normalizeWithAI(ctx, rawPosting)
        }
        c.logger.Warn("low quality parse result, falling back to AI",
            slog.Float64("quality_score", quality.Total),
            slog.String("url", url),
        )
    }

    // 3. AI Fallback
    c.logger.Info("using AI fallback parser",
        slog.String("url", url),
        slog.String("reason", errorReason(parseErr)),
    )

    posting, err := c.aiProvider.ParseHTMLFallback(ctx, url)
    if err != nil {
        // 4. Manual input required
        return nil, &CrawlError{
            Code:    "CRAWL_001",
            Message: "해당 URL을 분석할 수 없습니다. URL을 확인해 주세요",
            URL:     url,
            Cause:   err,
        }
    }

    posting.AIFallbackUsed = true
    return posting, nil
}

func isNonRetryable(err error) bool {
    var httpErr *HTTPError
    if errors.As(err, &httpErr) {
        return httpErr.StatusCode == 403 || httpErr.StatusCode == 404
    }
    return false
}
```

---

## 5. 크롤링 상태 대시보드 (관리자용)

### 5.1 대시보드 화면 구성

```
┌───────────────────────────────────────────────────────────────┐
│ 크롤러 모니터링 대시보드                     최근 갱신: 10분 전 │
│                                                               │
│ ┌──────────────────────────────────────────────────────────┐  │
│ │ 사이트별 상태                                             │  │
│ │                                                          │  │
│ │  잡코리아   ● 정상  성공률 96%  응답 2.1s  마지막 체크 10분 전│  │
│ │  캐치      ● 정상  성공률 94%  응답 1.8s  마지막 체크 10분 전│  │
│ │  사람인    ● 정상  성공률 99%  응답 0.5s  마지막 체크 10분 전│  │
│ │  원티드    ○ 미지원 (Phase 10)                            │  │
│ │  AI 폴백   ● 정상  성공률 88%  응답 4.2s                  │  │
│ └──────────────────────────────────────────────────────────┘  │
│                                                               │
│ ┌──────────────────────────────────────────────────────────┐  │
│ │ 성공률 시계열 (최근 7일)                                   │  │
│ │                                                          │  │
│ │ 100%|  ___    ____    _____                              │  │
│ │  90%| /   \__/    \__/     \___                          │  │
│ │  80%|                          \__  ← 잡코리아            │  │
│ │  70%|                                                    │  │
│ │     |──────────────────────────────────                  │  │
│ │      2/5  2/6  2/7  2/8  2/9  2/10  2/11  2/12         │  │
│ └──────────────────────────────────────────────────────────┘  │
│                                                               │
│ ┌──────────────────────────────────────────────────────────┐  │
│ │ 최근 실패 로그 (24시간)                                    │  │
│ │                                                          │  │
│ │ 시간           사이트    URL              에러             │  │
│ │ 14:23:15     잡코리아  /Recruit/GI..   403 Forbidden     │  │
│ │ 13:45:02     캐치     /Comp/Recru..   Selector 미스매치  │  │
│ │ 12:10:33     기타     example.com     AI 폴백 실패       │  │
│ │ ...                                                      │  │
│ └──────────────────────────────────────────────────────────┘  │
│                                                               │
│ ┌──────────────────────────────────────────────────────────┐  │
│ │ 파서 버전 관리                                            │  │
│ │                                                          │  │
│ │ 사이트      현재 버전   마지막 업데이트   Selector 수       │  │
│ │ 잡코리아    v3         2026-02-01      7개              │  │
│ │ 캐치       v2         2026-01-15      5개              │  │
│ │ 사람인     v1 (API)   2026-01-01      -               │  │
│ └──────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

### 5.2 관리자 API 엔드포인트

| Method | Path | 설명 |
|--------|------|------|
| `GET` | `/admin/crawler/status` | 사이트별 현재 상태 |
| `GET` | `/admin/crawler/stats?period=7d` | 기간별 통계 (성공률, 응답 시간) |
| `GET` | `/admin/crawler/failures?limit=50` | 최근 실패 로그 |
| `GET` | `/admin/crawler/health-checks` | 헬스체크 이력 |
| `POST` | `/admin/crawler/health-check/:domain` | 수동 헬스체크 실행 |

> 관리자 API는 별도 인증 미들웨어 (admin role 확인)로 보호한다.

---

## 6. Rate Limiting & 예의 바른 크롤링

> `08-legal-privacy.md`의 크롤링 3원칙 준수

### 6.1 요청 간격 제한

| 사이트 | 최소 요청 간격 | 동시 요청 수 | 비고 |
|--------|-------------|------------|------|
| 잡코리아 | 1초 | 1 | 동일 도메인 직렬 처리 |
| 캐치 | 1초 | 1 | 동일 도메인 직렬 처리 |
| 사람인 | API Rate Limit 준수 | 1 | oAPI 한도 내 |
| 원티드 | 2초 | 1 | SPA → 리소스 많음 |

### 6.2 Go 구현: Rate Limiter

```go
// apps/backend/internal/infrastructure/crawler/rate_limiter.go
package crawler

import (
    "sync"
    "time"
)

type DomainRateLimiter struct {
    mu       sync.Mutex
    lastReq  map[string]time.Time
    interval map[string]time.Duration
}

func NewDomainRateLimiter() *DomainRateLimiter {
    return &DomainRateLimiter{
        lastReq: make(map[string]time.Time),
        interval: map[string]time.Duration{
            "jobkorea.co.kr": 1 * time.Second,
            "catch.co.kr":    1 * time.Second,
            "saramin.co.kr":  1 * time.Second,
            "wanted.co.kr":   2 * time.Second,
        },
    }
}

func (r *DomainRateLimiter) Wait(domain string) {
    r.mu.Lock()
    defer r.mu.Unlock()

    interval, ok := r.interval[domain]
    if !ok {
        interval = 1 * time.Second // default
    }

    if last, exists := r.lastReq[domain]; exists {
        elapsed := time.Since(last)
        if elapsed < interval {
            time.Sleep(interval - elapsed)
        }
    }

    r.lastReq[domain] = time.Now()
}
```

### 6.3 robots.txt 준수

```go
// apps/backend/internal/infrastructure/crawler/robots.go
package crawler

import (
    "sync"
    "time"

    "github.com/temoto/robotstxt"
)

type RobotsChecker struct {
    mu    sync.RWMutex
    cache map[string]*robotsCacheEntry
}

type robotsCacheEntry struct {
    robots    *robotstxt.RobotsData
    fetchedAt time.Time
}

const robotsCacheTTL = 24 * time.Hour

func (r *RobotsChecker) IsAllowed(domain, path, userAgent string) (bool, error) {
    r.mu.RLock()
    entry, exists := r.cache[domain]
    r.mu.RUnlock()

    // Refresh if cache miss or expired
    if !exists || time.Since(entry.fetchedAt) > robotsCacheTTL {
        robots, err := r.fetchRobotsTxt(domain)
        if err != nil {
            // If robots.txt is unavailable, allow (conservative approach)
            return true, nil
        }
        r.mu.Lock()
        r.cache[domain] = &robotsCacheEntry{robots: robots, fetchedAt: time.Now()}
        r.mu.Unlock()
        entry = r.cache[domain]
    }

    group := entry.robots.FindGroup(userAgent)
    return group.Test(path), nil
}
```

### 6.4 User-Agent 설정

```go
const CrawlerUserAgent = "Mozilla/5.0 (compatible; Colight/1.0; +https://colight.kr/bot)"
```

> **08-legal-privacy.md 참조**: 봇 식별 가능한 User-Agent를 사용하며, 일반 브라우저 User-Agent를 사칭하지 않는다.

---

## 7. DB 스키마: crawler_health_checks

```go
// apps/backend/ent/schema/crawlerhealthcheck.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type CrawlerHealthCheck struct {
    ent.Schema
}

func (CrawlerHealthCheck) Mixin() []ent.Mixin {
    return []ent.Mixin{
        TimestampMixin{},
    }
}

func (CrawlerHealthCheck) Fields() []ent.Field {
    return []ent.Field{
        field.String("domain").
            NotEmpty().
            MaxLen(100).
            Comment("Target domain: jobkorea.co.kr, catch.co.kr, etc."),
        field.String("sample_url").
            Optional().
            MaxLen(500).
            Comment("URL used for health check"),
        field.Int("http_status").
            Default(0).
            Comment("HTTP response status code"),
        field.Int("latency_ms").
            Default(0).
            Comment("Request + parse latency in ms"),
        field.Int("total_selectors").
            Default(0).
            Comment("Total CSS selectors checked"),
        field.Int("found_selectors").
            Default(0).
            Comment("Successfully matched selectors"),
        field.JSON("missing_required", []string{}).
            Optional().
            Comment("Required selectors that were not found"),
        field.JSON("missing_optional", []string{}).
            Optional().
            Comment("Optional selectors that were not found"),
        field.Float("score").
            Default(0).
            Comment("Health score: 0.0 ~ 1.0"),
        field.Bool("healthy").
            Default(true).
            Comment("Overall health status"),
        field.Bool("alert_sent").
            Default(false).
            Comment("Whether alert was sent for this check"),
    }
}

func (CrawlerHealthCheck) Edges() []ent.Edge {
    return nil
}

func (CrawlerHealthCheck) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("domain", "created_at"),
        index.Fields("healthy"),
        index.Fields("domain", "healthy"),
    }
}
```

### 크롤링 통계 집계 테이블

```go
// apps/backend/ent/schema/crawlerstat.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type CrawlerStat struct {
    ent.Schema
}

func (CrawlerStat) Mixin() []ent.Mixin {
    return []ent.Mixin{
        TimestampMixin{},
    }
}

func (CrawlerStat) Fields() []ent.Field {
    return []ent.Field{
        field.String("domain").
            NotEmpty().
            MaxLen(100).
            Comment("Target domain"),
        field.Time("period_start").
            Comment("Stats period start (hourly bucket)"),
        field.Int("total_requests").
            Default(0).
            Comment("Total crawl requests in period"),
        field.Int("success_count").
            Default(0).
            Comment("Successful parses"),
        field.Int("failure_count").
            Default(0).
            Comment("Failed parses"),
        field.Int("ai_fallback_count").
            Default(0).
            Comment("AI fallback used count"),
        field.Int("avg_latency_ms").
            Default(0).
            Comment("Average latency in period"),
        field.Int("p95_latency_ms").
            Default(0).
            Comment("P95 latency in period"),
        field.Float("avg_quality_score").
            Default(0).
            Comment("Average quality score in period"),
    }
}

func (CrawlerStat) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("domain", "period_start"),
    }
}
```

---

## 8. River 기반 헬스체크 스케줄링

### 8.1 정기 헬스체크 Job

```go
// apps/backend/internal/worker/crawler_health.go
package worker

import (
    "context"
    "log/slog"

    "github.com/riverqueue/river"
)

type CrawlerHealthCheckArgs struct{}

func (CrawlerHealthCheckArgs) Kind() string { return "crawler_health_check" }

type CrawlerHealthCheckWorker struct {
    river.WorkerDefaults[CrawlerHealthCheckArgs]
    logger   *slog.Logger
    crawler  *crawler.Crawler
    alerter  *crawler.SlackAlerter
    db       *ent.Client
}

// SampleURLs — representative URLs for each domain (rotated periodically)
var SampleURLs = map[string][]string{
    "jobkorea.co.kr": {
        "https://www.jobkorea.co.kr/Recruit/GI_Read/44000000",
        "https://www.jobkorea.co.kr/Recruit/GI_Read/44000001",
    },
    "catch.co.kr": {
        "https://www.catch.co.kr/Comp/RecruitInfo/12345",
    },
}

func (w *CrawlerHealthCheckWorker) Work(ctx context.Context, job *river.Job[CrawlerHealthCheckArgs]) error {
    w.logger.Info("starting crawler health check")

    for domain, urls := range SampleURLs {
        // Pick a sample URL (rotate)
        url := urls[time.Now().Unix()%int64(len(urls))]

        result, err := w.crawler.CheckSiteHealth(ctx, domain, url)
        if err != nil {
            w.logger.Error("health check error",
                slog.String("domain", domain),
                slog.String("error", err.Error()),
            )
            continue
        }

        // Save to DB
        _, err = w.db.CrawlerHealthCheck.Create().
            SetDomain(result.Domain).
            SetSampleURL(result.URL).
            SetHTTPStatus(result.HTTPStatus).
            SetLatencyMs(int(result.LatencyMS)).
            SetTotalSelectors(result.TotalSelectors).
            SetFoundSelectors(result.FoundSelectors).
            SetMissingRequired(result.MissingRequired).
            SetMissingOptional(result.MissingOptional).
            SetScore(result.Score).
            SetHealthy(result.Healthy).
            Save(ctx)
        if err != nil {
            w.logger.Error("failed to save health check",
                slog.String("error", err.Error()),
            )
        }

        // Send alert if unhealthy
        if !result.Healthy {
            w.logger.Warn("crawler unhealthy",
                slog.String("domain", domain),
                slog.Float64("score", result.Score),
                slog.Any("missing_required", result.MissingRequired),
            )
            if err := w.alerter.SendAlert(result); err != nil {
                w.logger.Error("failed to send Slack alert",
                    slog.String("error", err.Error()),
                )
            }
        } else {
            w.logger.Info("crawler healthy",
                slog.String("domain", domain),
                slog.Float64("score", result.Score),
            )
        }
    }

    return nil
}
```

### 8.2 River 스케줄러 등록

```go
// cmd/api/main.go — River periodic jobs

periodicJobs := []*river.PeriodicJob{
    // Crawler health check: every 30 minutes
    river.NewPeriodicJob(
        river.PeriodicInterval(30*time.Minute),
        func() (river.JobArgs, *river.InsertOpts) {
            return CrawlerHealthCheckArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: true}, // Check on startup
    ),

    // Crawler stats aggregation: every hour
    river.NewPeriodicJob(
        river.PeriodicInterval(1*time.Hour),
        func() (river.JobArgs, *river.InsertOpts) {
            return AggregateStatsArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: false},
    ),

    // Old health check cleanup: daily (keep 30 days)
    river.NewPeriodicJob(
        river.PeriodicInterval(24*time.Hour),
        func() (river.JobArgs, *river.InsertOpts) {
            return CleanupHealthChecksArgs{RetentionDays: 30}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: false},
    ),
}
```

---

## 9. Phase별 구현 계획

| Phase | 구현 내용 |
|-------|----------|
| **Phase 3** | 잡코리아/캐치 파서 + AI 폴백 + 기본 로깅 |
| **Phase 3** | 크롤링 결과 품질 점수 계산 + 재시도 로직 |
| **Phase 3.2** | 사람인 API 연동, DART/네이버 API 에러 처리 |
| **Phase 6.1** | 크롤러 헬스체크 River job + `crawler_health_checks` 테이블 |
| **Phase 6.1** | Slack webhook 알림 + 자동 AI 폴백 전환 |
| **Phase 8** | 관리자 대시보드 UI (사이트별 성공률, 실패 로그) |
| **Phase 10** | 원티드 Playwright 파서 + 동적 사이트 모니터링 |
| **Phase 10** | 파서 버전 관리 시스템 + 자동 selector 후보 탐색 |

---

## 10. 환경변수

| 변수명 | 설명 | 필요 Phase |
|--------|------|-----------|
| `SLACK_WEBHOOK_URL` | Slack 알림 Webhook URL | Phase 6.1 |
| `CRAWLER_HEALTH_INTERVAL` | 헬스체크 간격 (기본: 30m) | Phase 6.1 |
| `CRAWLER_MAX_RETRIES` | 최대 재시도 횟수 (기본: 3) | Phase 3 |
| `CRAWLER_TIMEOUT` | 요청 타임아웃 (기본: 10s) | Phase 3 |

---

## 11. 에러 코드 추가

> `09-error-handling.md`에 추가할 크롤러 관련 에러 코드

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `CRAWL_001` | URL 파싱 실패 (기존) | 400 | 해당 URL을 분석할 수 없습니다 |
| `CRAWL_002` | 사이트 접근 차단 (403) | 502 | 해당 공고에 접근하기 어렵습니다 |
| `CRAWL_003` | 사이트 구조 변경 감지 | 502 | 자동 분석에 실패했습니다. 수동 입력을 이용해 주세요 |
| `CRAWL_004` | 크롤링 타임아웃 | 504 | 공고 페이지 응답이 느립니다. 다시 시도해 주세요 |
| `CRAWL_005` | AI 폴백 실패 | 502 | 공고 분석에 실패했습니다. 수동 입력을 이용해 주세요 |
