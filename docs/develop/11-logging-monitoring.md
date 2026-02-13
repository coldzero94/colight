# 로깅 & 모니터링 전략

> Go 구조화 로깅, AI 비용 추적, Sentry 에러 추적, 성능 기준 및 헬스체크

---

## 1. Go 백엔드 구조화 로깅

### log/slog 기반 로거 설정

Go 1.21+ 표준 라이브러리인 `log/slog`를 사용한다. 별도 서드파티 로깅 라이브러리를 추가하지 않는다.

```go
// internal/logger/logger.go
package logger

import (
    "log/slog"
    "os"
)

func New(env string) *slog.Logger {
    var handler slog.Handler

    switch env {
    case "production", "staging":
        // JSON 포맷 — Koyeb 로그 수집기와 호환
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    default:
        // Text 포맷 — 로컬 개발 시 가독성
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }

    return slog.New(handler)
}
```

### 로그 레벨 가이드

| 레벨 | 용도 | 예시 |
|------|------|------|
| `DEBUG` | 개발 디버깅 (production에서 비활성화) | SQL 쿼리 내용, AI 프롬프트 전문 |
| `INFO` | 정상 동작 기록 | 요청 처리 완료, 마이그레이션 적용 |
| `WARN` | 주의가 필요한 상황 | 외부 API 재시도, 캐시 미스, Rate limit 근접 |
| `ERROR` | 처리 실패, 즉시 확인 필요 | AI API 실패, DB 연결 오류, 인증 실패 |

### 요청별 trace_id 주입 (Gin 미들웨어)

모든 HTTP 요청에 고유 `trace_id`를 부여하여 로그 추적을 용이하게 한다.

```go
// internal/middleware/request_logger.go
package middleware

import (
    "log/slog"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // trace_id 생성 및 context 주입
        traceID := uuid.New().String()[:8] // 짧은 ID (8자)
        c.Set("trace_id", traceID)
        c.Header("X-Trace-ID", traceID)

        start := time.Now()

        // 요청 처리
        c.Next()

        // 응답 로그
        duration := time.Since(start)
        status := c.Writer.Status()

        attrs := []slog.Attr{
            slog.String("trace_id", traceID),
            slog.String("method", c.Request.Method),
            slog.String("path", c.Request.URL.Path),
            slog.Int("status", status),
            slog.Duration("duration", duration),
            slog.String("ip", c.ClientIP()),
        }

        // 사용자 ID (인증된 요청인 경우)
        if userID, exists := c.Get("user_id"); exists {
            attrs = append(attrs, slog.String("user_id", userID.(string)))
        }

        level := slog.LevelInfo
        if status >= 500 {
            level = slog.LevelError
        } else if status >= 400 {
            level = slog.LevelWarn
        }

        logger.LogAttrs(c.Request.Context(), level, "request completed", attrs...)
    }
}
```

### 로그 출력 예시 (JSON)

```json
{
  "time": "2026-03-15T14:30:22.123Z",
  "level": "INFO",
  "msg": "request completed",
  "trace_id": "a1b2c3d4",
  "method": "POST",
  "path": "/v1/analyses",
  "status": 200,
  "duration": "2.345s",
  "user_id": "usr_abc123",
  "ip": "203.0.113.42"
}
```

### 민감 정보 마스킹

경험 내용, 자소서 본문 등 사용자 개인 데이터는 로그에 기록하지 않는다.

```go
// internal/logger/sanitize.go
package logger

import "strings"

// SanitizeForLog truncates and masks sensitive content for logging
func SanitizeForLog(content string, maxLen int) string {
    if len(content) == 0 {
        return "[empty]"
    }
    // 유니코드 안전한 truncation
    runes := []rune(content)
    if len(runes) > maxLen {
        return string(runes[:maxLen]) + "...[truncated]"
    }
    return content
}

// MaskEmail masks email addresses in log output
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    if len(parts[0]) <= 2 {
        return "**@" + parts[1]
    }
    return parts[0][:2] + "***@" + parts[1]
}
```

**마스킹 원칙:**
- 경험 본문 (STAR 내용): 로그에 기록하지 않음 (ID만 기록)
- 자소서 본문: 로그에 기록하지 않음
- 이메일: 앞 2글자만 노출 (`co***@gmail.com`)
- AI 프롬프트: DEBUG 레벨에서만 기록 (production에서 비활성화)

---

## 2. AI API 비용 모니터링

### 호출별 로깅

모든 AI API 호출에 대해 토큰 수, 비용, 모델, 응답 시간을 기록한다.

```go
// internal/infrastructure/ai/logger.go
package ai

import (
    "context"
    "log/slog"
    "time"
)

type UsageLog struct {
    TraceID       string        `json:"trace_id"`
    UserID        string        `json:"user_id"`
    Model         string        `json:"model"`         // "claude-sonnet-4-5", "gemini-2.0-flash"
    Purpose       string        `json:"purpose"`       // "company_analysis", "draft_coaching", etc.
    InputTokens   int           `json:"input_tokens"`
    OutputTokens  int           `json:"output_tokens"`
    TotalTokens   int           `json:"total_tokens"`
    EstimatedCost float64       `json:"estimated_cost"` // KRW
    Duration      time.Duration `json:"duration"`
    Success       bool          `json:"success"`
}

func LogAIUsage(ctx context.Context, logger *slog.Logger, usage UsageLog) {
    logger.LogAttrs(ctx, slog.LevelInfo, "ai_api_call",
        slog.String("trace_id", usage.TraceID),
        slog.String("user_id", usage.UserID),
        slog.String("model", usage.Model),
        slog.String("purpose", usage.Purpose),
        slog.Int("input_tokens", usage.InputTokens),
        slog.Int("output_tokens", usage.OutputTokens),
        slog.Int("total_tokens", usage.TotalTokens),
        slog.Float64("estimated_cost_krw", usage.EstimatedCost),
        slog.Duration("duration", usage.Duration),
        slog.Bool("success", usage.Success),
    )
}
```

### 비용 계산 기준

```go
// internal/infrastructure/ai/cost.go
package ai

// EstimateCost calculates the estimated KRW cost for an AI API call
func EstimateCost(model string, inputTokens, outputTokens int) float64 {
    // 1 USD ≈ 1,350 KRW 기준
    const usdToKrw = 1350.0

    var inputPricePerMToken, outputPricePerMToken float64

    switch model {
    case "claude-sonnet-4-5":
        inputPricePerMToken = 3.0   // $3/1M input tokens
        outputPricePerMToken = 15.0 // $15/1M output tokens
    case "gemini-2.0-flash":
        inputPricePerMToken = 0.1   // $0.1/1M input tokens
        outputPricePerMToken = 0.4  // $0.4/1M output tokens
    case "text-embedding-3-small":
        inputPricePerMToken = 0.02  // $0.02/1M tokens
        outputPricePerMToken = 0.0
    default:
        return 0
    }

    inputCostUSD := float64(inputTokens) / 1_000_000 * inputPricePerMToken
    outputCostUSD := float64(outputTokens) / 1_000_000 * outputPricePerMToken
    totalKRW := (inputCostUSD + outputCostUSD) * usdToKrw

    return totalKRW
}
```

### ai_usage_logs 테이블

DB 기반으로 AI 사용량을 추적하여 비용 집계 및 사용량 제한에 활용한다.

```sql
-- Ent 스키마로 정의 (참고용 SQL)
CREATE TABLE ai_usage_logs (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES user_profiles(id),
    trace_id     VARCHAR(16) NOT NULL,
    model        VARCHAR(50) NOT NULL,
    purpose      VARCHAR(50) NOT NULL,
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    estimated_cost_krw DECIMAL(10, 2) NOT NULL DEFAULT 0,
    duration_ms  INT NOT NULL DEFAULT 0,
    success      BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_usage_user_date ON ai_usage_logs(user_id, created_at);
CREATE INDEX idx_ai_usage_date ON ai_usage_logs(created_at);
```

### 비용 집계 쿼리

```sql
-- 일별 비용 집계
SELECT
    DATE(created_at AT TIME ZONE 'Asia/Seoul') AS date,
    model,
    COUNT(*) AS call_count,
    SUM(input_tokens) AS total_input_tokens,
    SUM(output_tokens) AS total_output_tokens,
    SUM(estimated_cost_krw) AS total_cost_krw
FROM ai_usage_logs
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at AT TIME ZONE 'Asia/Seoul'), model
ORDER BY date DESC;

-- 사용자별 일일 비용
SELECT
    user_id,
    SUM(estimated_cost_krw) AS daily_cost_krw,
    COUNT(*) AS call_count
FROM ai_usage_logs
WHERE created_at >= DATE_TRUNC('day', NOW() AT TIME ZONE 'Asia/Seoul')
GROUP BY user_id
ORDER BY daily_cost_krw DESC;
```

### 비용 임계값 알림

```go
// internal/service/cost_monitor.go
package service

import (
    "context"
    "log/slog"
)

const (
    DailyCostWarningKRW = 30000  // 일 3만원 경고
    DailyCostLimitKRW   = 50000  // 일 5만원 차단
)

type CostMonitor struct {
    logger *slog.Logger
    repo   AIUsageRepository
}

func (m *CostMonitor) CheckDailyCost(ctx context.Context) error {
    totalCost, err := m.repo.GetTodayTotalCost(ctx)
    if err != nil {
        return err
    }

    if totalCost >= DailyCostLimitKRW {
        m.logger.Error("daily AI cost limit exceeded",
            slog.Float64("total_cost_krw", totalCost),
            slog.Float64("limit_krw", DailyCostLimitKRW),
        )
        // TODO: Slack/Discord 웹훅 알림 발송
        return ErrDailyCostLimitExceeded
    }

    if totalCost >= DailyCostWarningKRW {
        m.logger.Warn("daily AI cost approaching limit",
            slog.Float64("total_cost_krw", totalCost),
            slog.Float64("warning_krw", DailyCostWarningKRW),
        )
    }

    return nil
}
```

**환경변수 기반 상한 설정:**

```bash
# 전체 서비스 일일 AI 비용 상한
MAX_DAILY_AI_COST=50000  # KRW
```

---

## 3. 에러 추적 (Sentry)

### Go 백엔드: sentry-go

```go
// cmd/api/main.go
package main

import (
    "log"
    "time"

    "github.com/getsentry/sentry-go"
    sentrygin "github.com/getsentry/sentry-go/gin"
    "github.com/gin-gonic/gin"
)

func initSentry(dsn, env string) {
    err := sentry.Init(sentry.ClientOptions{
        Dsn:              dsn,
        Environment:      env,                   // "development", "staging", "production"
        Release:          "colight-api@" + version, // Git tag 또는 commit hash
        TracesSampleRate: 0.2,                   // 20% 트랜잭션 샘플링
        EnableTracing:    true,
    })
    if err != nil {
        log.Fatalf("sentry.Init: %s", err)
    }
}

func setupRouter(r *gin.Engine) {
    // Sentry 미들웨어 (panic recovery + 에러 캡처)
    r.Use(sentrygin.New(sentrygin.Options{
        Repanic: true,
    }))

    // 요청 완료 후 Sentry에 전송
    r.Use(func(c *gin.Context) {
        c.Next()

        // 500 에러 발생 시 Sentry에 보고
        if len(c.Errors) > 0 {
            hub := sentrygin.GetHubFromContext(c)
            if hub != nil {
                for _, err := range c.Errors {
                    hub.CaptureException(err.Err)
                }
            }
        }
    })
}

func main() {
    initSentry(os.Getenv("SENTRY_DSN"), os.Getenv("APP_ENV"))
    defer sentry.Flush(2 * time.Second)

    // ...
}
```

### Next.js 프론트엔드: @sentry/nextjs

```bash
# apps/web
npx @sentry/wizard@latest -i nextjs
```

```typescript
// apps/web/sentry.client.config.ts
import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  environment: process.env.NODE_ENV,
  tracesSampleRate: 0.1,  // 10% 트랜잭션 샘플링

  // 프론트엔드 에러 필터링
  ignoreErrors: [
    "ResizeObserver loop",  // 브라우저 노이즈
    "Network request failed",
    "Load failed",
  ],
});
```

```typescript
// apps/web/sentry.server.config.ts
import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.SENTRY_DSN,
  environment: process.env.NODE_ENV,
  tracesSampleRate: 0.2,
});
```

### Sentry 환경 분리

| 환경 | DSN | 샘플링 | 알림 |
|------|-----|--------|------|
| development | 없음 (비활성화) | - | - |
| staging | staging DSN | 100% | Slack (저빈도) |
| production | production DSN | 20% | Slack + 이메일 (즉시) |

### 에러 그루핑 및 알림

```
Sentry Alert Rules:
├── P0 (즉시 알림 — Slack + 이메일)
│   ├── 500 에러 5분 내 10회 이상
│   ├── AI API 호출 실패율 > 30%
│   └── DB 연결 에러
│
├── P1 (1시간 내 확인 — Slack)
│   ├── 401/403 비정상 증가
│   ├── 크롤링 실패율 > 50%
│   └── 외부 API (DART, 네이버) 연속 실패
│
└── P2 (일일 리뷰)
    ├── 4xx 에러 트렌드
    └── 성능 저하 트렌드
```

---

## 4. 성능 모니터링 기준

### API 응답 시간 목표

| 카테고리 | P50 목표 | P95 목표 | 비고 |
|----------|---------|---------|------|
| **일반 CRUD** (경험, 프로필) | < 100ms | < 300ms | DB 쿼리 기반 |
| **목록 조회** (검색, 필터) | < 200ms | < 500ms | 페이지네이션 필수 |
| **외부 API** (DART, 네이버) | < 1s | < 3s | 캐시 적용 시 < 100ms |
| **AI 분석** (Claude) | < 10s | < 30s | 스트리밍 응답 |
| **AI 경량** (Gemini/Groq) | < 3s | < 8s | 태깅, 파싱 |
| **임베딩 생성** | < 1s | < 3s | 배치 처리 가능 |

### Koyeb 리소스 모니터링

```
┌─────────────────────────────────┐
│  Koyeb 512MB RAM 사용 가이드       │
├─────────────────────────────────┤
│  Go 바이너리 기본          ~20MB  │
│  Gin 라우터 + 미들웨어      ~10MB  │
│  Ent 클라이언트 + 커넥션 풀  ~30MB  │
│  River 워커               ~20MB  │
│  요청 처리 버퍼            ~100MB  │
│  ─────────────────────────────  │
│  예상 상시 사용            ~180MB  │
│  피크 (동시 요청 처리)      ~350MB  │
│  여유 마진                ~162MB  │
│  소프트 리밋 (GC target)   450MB  │
│  하드 리밋 (Koyeb OOM)    512MB  │
└─────────────────────────────────┘
```

**메모리 모니터링 로그:**

```go
// internal/middleware/memory_monitor.go
package middleware

import (
    "log/slog"
    "runtime"
    "time"

    "github.com/gin-gonic/gin"
)

// MemoryMonitor logs memory stats periodically
func MemoryMonitor(logger *slog.Logger, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)

            logger.Info("memory_stats",
                slog.Uint64("alloc_mb", m.Alloc/1024/1024),
                slog.Uint64("sys_mb", m.Sys/1024/1024),
                slog.Uint64("heap_inuse_mb", m.HeapInuse/1024/1024),
                slog.Uint64("goroutines", uint64(runtime.NumGoroutine())),
            )

            // 400MB 초과 시 경고
            if m.Alloc > 400*1024*1024 {
                logger.Warn("high memory usage",
                    slog.Uint64("alloc_mb", m.Alloc/1024/1024),
                )
            }
        }
    }()
}
```

### Vercel 성능 기준

| 지표 | 목표 | 측정 |
|------|------|------|
| TTFB (Time To First Byte) | < 500ms | Vercel Analytics |
| LCP (Largest Contentful Paint) | < 2.5s | Web Vitals |
| FID (First Input Delay) | < 100ms | Web Vitals |
| CLS (Cumulative Layout Shift) | < 0.1 | Web Vitals |

```typescript
// apps/web/src/app/layout.tsx
import { SpeedInsights } from "@vercel/speed-insights/next";
import { Analytics } from "@vercel/analytics/react";

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <SpeedInsights />
        <Analytics />
      </body>
    </html>
  );
}
```

---

## 5. 헬스체크

### 기본 헬스체크: GET /health

Koyeb 헬스체크 및 기본 상태 확인에 사용한다. DB 연결 확인을 포함한다.

```go
// internal/controller/health.go
package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

type HealthController struct {
    db *ent.Client
}

// Health — Koyeb 헬스체크용 (빠른 응답 필수)
func (h *HealthController) Health(c *gin.Context) {
    // DB ping
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()

    if err := h.db.DB().PingContext(ctx); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error":  "database connection failed",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}
```

### 상세 준비 상태: GET /health/ready

외부 의존성 (AI API 등) 연결 확인을 포함한다. 배포 직후 서비스 준비 상태 확인에 사용한다.

```go
// Readiness — 모든 외부 의존성 확인
func (h *HealthController) Ready(c *gin.Context) {
    checks := map[string]string{}
    allHealthy := true

    // 1. DB 연결
    ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
    defer cancel()

    if err := h.db.DB().PingContext(ctx); err != nil {
        checks["database"] = "unhealthy: " + err.Error()
        allHealthy = false
    } else {
        checks["database"] = "healthy"
    }

    // 2. AI API 키 설정 확인 (실제 호출은 하지 않음)
    if os.Getenv("ANTHROPIC_API_KEY") != "" {
        checks["anthropic"] = "configured"
    } else {
        checks["anthropic"] = "not_configured"
        // 키가 없어도 서비스는 동작 가능 (AI 기능만 비활성화)
    }

    if os.Getenv("OPENAI_API_KEY") != "" {
        checks["openai"] = "configured"
    } else {
        checks["openai"] = "not_configured"
    }

    status := http.StatusOK
    if !allHealthy {
        status = http.StatusServiceUnavailable
    }

    c.JSON(status, gin.H{
        "status":    map[bool]string{true: "ready", false: "not_ready"}[allHealthy],
        "checks":    checks,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}
```

### 라우터 등록

```go
// cmd/api/main.go
func setupRoutes(r *gin.Engine, healthCtrl *controller.HealthController) {
    // 헬스체크 (인증 불필요)
    r.GET("/health", healthCtrl.Health)
    r.GET("/health/ready", healthCtrl.Ready)

    // API 라우트 (인증 필요)
    v1 := r.Group("/v1")
    v1.Use(middleware.Auth())
    // ...
}
```

### Koyeb 헬스체크 설정

```
Koyeb Dashboard → Service → Health Checks:
├── Type: HTTP
├── Port: 8000
├── Path: /health
├── Interval: 30s
├── Timeout: 5s
├── Healthy threshold: 2 (연속 2회 성공 시 healthy)
└── Unhealthy threshold: 3 (연속 3회 실패 시 unhealthy → 재시작)
```

### 헬스체크 응답 예시

```json
// GET /health (Koyeb용 — 간결)
{
  "status": "healthy",
  "timestamp": "2026-03-15T14:30:00Z"
}

// GET /health/ready (상세 상태)
{
  "status": "ready",
  "checks": {
    "database": "healthy",
    "anthropic": "configured",
    "openai": "configured"
  },
  "timestamp": "2026-03-15T14:30:00Z"
}
```
