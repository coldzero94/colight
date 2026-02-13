# Colight - Fair Use Policy (사용량 제한 정책)

> 플랜별 사용 한도, 비용 보호 메커니즘, 한도 초과 UX, 환불 및 만료 정책

---

## 1. 플랜별 사용 한도

### 기본 한도 테이블

| 항목 | 무료 | 스타터 10회권 | 프로 30회권 | 시즌패스 |
|------|------|-------------|------------|---------|
| **가격** | 0원 | 4,900원 | 12,900원 | 24,900원/3개월 |
| **경험 등록** | 3개 | 무제한 | 무제한 | 무제한 |
| **기업 분석** | 1회 | 10회 | 30회 | 무제한* |
| **AI 코칭** (초안+첨삭) | 1회 | 10회 | 30회 | 무제한* |
| **일일 분석 한도** | 1회 | 5회 | 10회 | 20회 |
| **일일 코칭 한도** | 1회 | 5회 | 10회 | 20회 |
| **AI 인터뷰** | 1회 | 5회 | 15회 | 무제한* |
| **첨삭 코칭 횟수** (문항당) | 1회 | 3회 | 5회 | 10회 |

> *시즌패스 "무제한"은 일일 한도(20회) 적용. 일일 한도를 초과할 수 없음.

### 한도 카운팅 기준

| 행동 | 차감 단위 | 비고 |
|------|----------|------|
| 채용공고 URL 제출 → 기업 분석 | 기업 분석 1회 | 캐시 히트 시 차감 없음 (7일 TTL) |
| 자소서 문항 초안 생성 | AI 코칭 1회 | 문항 단위 |
| 자소서 첨삭 요청 | AI 코칭 1회 | 문항당 첨삭 횟수 별도 관리 |
| AI 경험 인터뷰 (5턴 완료) | AI 인터뷰 1회 | 중도 포기 시 차감 없음 |
| 경험 무기 자동 태깅 | 차감 없음 | 경량 모델 (Gemini/Groq), 무료 제공 |
| 임베딩 기반 매칭 | 차감 없음 | 분석 시 자동 수행 |

### 캐시와 한도의 관계

```
사용자 A: 삼성전자 공고 분석 요청
    │
    ├── 캐시 미스 → AI 분석 실행 → 기업 분석 1회 차감
    │
사용자 A: 동일 삼성전자 공고 재조회 (7일 이내)
    │
    └── 캐시 히트 → 캐시 결과 반환 → 차감 없음

사용자 B: 동일 삼성전자 공고 분석 요청
    │
    └── 캐시 히트 (다른 사용자 캐시 활용) → 사용자 B 기업 분석 1회 차감
        (분석 결과 "조회"에는 차감, 실제 AI 호출은 없으므로 비용 절약)
```

> **설계 결정**: 캐시 히트 시에도 다른 사용자에게는 1회 차감한다. 이유: 사용자 입장에서 "기업 분석 결과를 받았다"는 가치는 동일하며, 무과금 시 무한 조회 어뷰징이 가능하다. 단, 동일 사용자의 재조회는 차감하지 않는다.

---

## 2. 비용 보호 메커니즘

### 사용자 단위 비용 보호

```
┌──────────────────────────────────────────────────────┐
│              사용자당 일일 비용 캡                        │
├──────────────────────────────────────────────────────┤
│                                                      │
│  일일 최대 AI 호출: 20회 (시즌패스 기준)                  │
│  건당 최대 비용: ~200원 (분석 + 코칭 풀 파이프라인)         │
│  사용자당 일일 최대 비용: 20회 × 200원 = 4,000원/일       │
│                                                      │
│  시즌패스 월간 최대 비용:                                 │
│  30일 × 4,000원 = 120,000원/월                         │
│  → 시즌패스 가격 24,900원/3개월 대비 최대 손실 가능         │
│  → 실제 헤비 유저 비율 5% 미만으로 추정                    │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 서비스 전체 비용 보호

| 보호 레벨 | 임계값 | 동작 |
|----------|--------|------|
| **경고** | 일일 AI 비용 3만원 초과 | 관리자 Slack 알림 |
| **주의** | 일일 AI 비용 5만원 초과 | 관리자 이메일 + Slack |
| **차단** | 일일 AI 비용 `MAX_DAILY_AI_COST` 초과 | 신규 AI 요청 일시 차단 |

```bash
# 환경변수로 전체 서비스 일일 AI 비용 상한 설정
MAX_DAILY_AI_COST=100000  # KRW (기본값: 10만원)
```

### 어뷰징 방지

| 어뷰징 유형 | 탐지 방법 | 대응 |
|-----------|----------|------|
| 다중 계정 무료 체험 남용 | 동일 IP에서 5개 이상 가입 | 경고 → 차단 |
| API 직접 호출 (우회) | Rate Limit + JWT 검증 | 429 응답 |
| 봇/자동화 | 요청 패턴 분석 (짧은 간격 반복) | CAPTCHA 또는 차단 |
| 크레딧 공유 | 동시 세션 제한 (3개) | 이전 세션 만료 |

---

## 3. 한도 초과 UX

### 남은 한도 실시간 표시

사이드바 또는 헤더에 현재 사용량을 표시한다.

```
┌─────────────────────────────────┐
│  🎯 분석 남은 횟수: 7/10          │
│  ✏️ 코칭 남은 횟수: 8/10          │
│  ──────────────────────────────  │
│  오늘 사용: 분석 3회 / 코칭 2회     │
│  (일일 한도: 5회)                  │
└─────────────────────────────────┘
```

### 한도 80% 도달 시 경고 배너

```typescript
// apps/web/src/components/UsageBanner.tsx (참고용 로직)

// 총 한도의 80% 사용 시
// "기업 분석 잔여 횟수가 2회 남았습니다. 프로 30회권으로 업그레이드하면 더 많은 분석이 가능해요."

// 일일 한도의 80% 사용 시
// "오늘 분석 한도 4/5회를 사용했습니다. 내일 자정에 리셋됩니다."
```

### 한도 소진 시 소프트 페이월

총 한도 소진과 일일 한도 소진을 구분하여 안내한다.

**총 한도 소진:**
```
┌──────────────────────────────────────────┐
│                                          │
│     기업 분석 횟수를 모두 사용했습니다         │
│                                          │
│  현재 플랜: 스타터 10회권 (10/10 사용)       │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  프로 30회권으로 업그레이드           │  │
│  │  12,900원 → 30회 추가 분석 + 코칭    │  │
│  │            [업그레이드하기]           │  │
│  └────────────────────────────────────┘  │
│                                          │
│  또는 시즌패스 24,900원/3개월 (무제한)      │
│                                          │
└──────────────────────────────────────────┘
```

**일일 한도 소진:**
```
┌──────────────────────────────────────────┐
│                                          │
│     오늘의 분석 한도를 모두 사용했습니다      │
│                                          │
│  일일 한도: 5/5회 사용                     │
│  리셋 시간: 오늘 자정 (00:00 KST)          │
│                                          │
│  총 잔여 횟수: 15회 남음                    │
│                                          │
│  서비스 이용량이 많습니다.                    │
│  내일 다시 시도해 주세요.                    │
│                                          │
└──────────────────────────────────────────┘
```

### 일일 한도 리셋 시간

- **리셋 시간**: 매일 00:00 KST (UTC+9)
- **리셋 방식**: River scheduled job (cron: `0 15 * * *` UTC = 00:00 KST)
- **리셋 대상**: `usage_tracking` 테이블의 `daily_analysis_count`, `daily_coaching_count`

---

## 4. 환불 및 만료 정책

### 회차권 (스타터 / 프로)

| 항목 | 정책 |
|------|------|
| **유효 기간** | 구매 후 6개월 |
| **부분 사용 후 환불** | 불가 (1회라도 사용 시) |
| **미사용 환불** | 구매 후 7일 이내, 1회도 미사용 시 전액 환불 |
| **만료 시** | 잔여 횟수 소멸, 별도 안내 (만료 7일 전 이메일) |
| **재구매** | 잔여 횟수에 누적 (스타터 3회 남음 + 프로 구매 = 33회) |

### 시즌패스

| 항목 | 정책 |
|------|------|
| **유효 기간** | 구매 후 90일 (3개월) |
| **미사용 환불** | 구매 후 7일 이내, 1회도 미사용 시 전액 환불 |
| **사용 후 환불** | 불가 |
| **만료 시** | 무료 플랜으로 자동 전환 (데이터 유지) |
| **갱신** | 자동 갱신 없음, 수동 재구매 |

### 크레딧 (이벤트/프로모션)

| 항목 | 정책 |
|------|------|
| **유효 기간** | 이벤트별 별도 설정 (기본 30일) |
| **환불** | 불가 (무료 지급) |
| **만료 시** | 자동 소멸 |
| **적용 순서** | 이벤트 크레딧 → 회차권 순으로 차감 (만료 임박 순) |

> `user_profiles.credits` 필드는 만료 없는 기본 크레딧이다. 이벤트 크레딧은 별도 `credit_events` 테이블에서 만료일과 함께 관리한다.

### 만료 안내 타임라인

```
D-30  이메일: "회차권이 30일 후 만료됩니다"
D-7   이메일 + 앱 내 배너: "7일 후 잔여 N회가 소멸됩니다"
D-1   이메일 + 앱 내 모달: "내일 잔여 N회가 소멸됩니다"
D-Day 소멸 처리 + 이메일: "회차권이 만료되었습니다. 재구매하기"
```

---

## 5. 구현 방법

### usage_tracking 테이블

```sql
-- Ent 스키마로 정의 (참고용 SQL)
CREATE TABLE usage_tracking (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              UUID NOT NULL REFERENCES user_profiles(id),

    -- 총 사용량 (전체 기간)
    total_analysis_count  INT NOT NULL DEFAULT 0,
    total_coaching_count  INT NOT NULL DEFAULT 0,
    total_interview_count INT NOT NULL DEFAULT 0,

    -- 일일 사용량 (매일 리셋)
    daily_analysis_count  INT NOT NULL DEFAULT 0,
    daily_coaching_count  INT NOT NULL DEFAULT 0,
    daily_reset_at        TIMESTAMPTZ NOT NULL DEFAULT DATE_TRUNC('day', NOW() AT TIME ZONE 'Asia/Seoul'),

    -- 플랜 한도
    plan_type             VARCHAR(20) NOT NULL DEFAULT 'free',  -- free, starter, pro, season_pass
    plan_analysis_limit   INT NOT NULL DEFAULT 1,
    plan_coaching_limit   INT NOT NULL DEFAULT 1,
    daily_analysis_limit  INT NOT NULL DEFAULT 1,
    daily_coaching_limit  INT NOT NULL DEFAULT 1,
    plan_expires_at       TIMESTAMPTZ,

    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id)
);

CREATE INDEX idx_usage_tracking_user ON usage_tracking(user_id);
CREATE INDEX idx_usage_tracking_plan_expiry ON usage_tracking(plan_expires_at);
```

### Go 미들웨어: 요청 전 사용량 체크

```go
// internal/middleware/usage_limit.go
package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type UsageType string

const (
    UsageAnalysis  UsageType = "analysis"
    UsageCoaching  UsageType = "coaching"
    UsageInterview UsageType = "interview"
)

func UsageLimit(usageType UsageType, repo UsageRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")

        // 사용량 조회
        usage, err := repo.GetUserUsage(c.Request.Context(), userID)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "사용량을 확인할 수 없습니다", "code": "USAGE_001"},
            })
            return
        }

        // 일일 리셋 확인
        if usage.NeedsDailyReset() {
            if err := repo.ResetDailyCount(c.Request.Context(), userID); err != nil {
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "error": gin.H{"message": "서버 오류가 발생했습니다", "code": "USAGE_002"},
                })
                return
            }
            usage.ResetDaily()
        }

        // 플랜 만료 확인
        if usage.IsPlanExpired() {
            c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
                "error": gin.H{
                    "message": "플랜이 만료되었습니다. 업그레이드해 주세요.",
                    "code":    "PLAN_EXPIRED",
                },
            })
            return
        }

        // 총 한도 확인
        if !usage.HasRemainingTotal(usageType) {
            c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
                "error": gin.H{
                    "message": "사용 한도를 모두 소진했습니다. 업그레이드해 주세요.",
                    "code":    "LIMIT_TOTAL_EXCEEDED",
                },
            })
            return
        }

        // 일일 한도 확인
        if !usage.HasRemainingDaily(usageType) {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": gin.H{
                    "message": "오늘의 사용 한도를 모두 사용했습니다. 내일 다시 시도해 주세요.",
                    "code":    "LIMIT_DAILY_EXCEEDED",
                },
            })
            return
        }

        c.Next()
    }
}
```

### 사용량 증가 (Service 레이어)

```go
// internal/service/usage.go
package service

func (s *UsageService) IncrementUsage(ctx context.Context, userID string, usageType UsageType) error {
    // 트랜잭션으로 사용량 증가 (동시성 안전)
    tx, err := s.client.Tx(ctx)
    if err != nil {
        return err
    }

    _, err = tx.UsageTracking.Update().
        Where(usagetracking.UserID(userID)).
        AddTotalAnalysisCount(1).    // usageType에 따라 분기
        AddDailyAnalysisCount(1).
        Save(ctx)

    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}
```

### 일일 리셋: River Scheduled Job

```go
// internal/worker/daily_reset.go
package worker

import (
    "context"
    "log/slog"
    "time"

    "github.com/riverqueue/river"
)

type DailyResetArgs struct{}

func (DailyResetArgs) Kind() string { return "daily_usage_reset" }

type DailyResetWorker struct {
    river.WorkerDefaults[DailyResetArgs]
    logger *slog.Logger
    repo   UsageRepository
}

func (w *DailyResetWorker) Work(ctx context.Context, job *river.Job[DailyResetArgs]) error {
    w.logger.Info("starting daily usage reset")

    count, err := w.repo.ResetAllDailyCounts(ctx)
    if err != nil {
        w.logger.Error("daily usage reset failed", slog.String("error", err.Error()))
        return err
    }

    w.logger.Info("daily usage reset completed",
        slog.Int("users_reset", count),
    )
    return nil
}
```

```go
// River 스케줄러 등록 (cmd/api/main.go)
// cron: "0 15 * * *" (UTC) = 매일 00:00 KST
periodicJobs := []*river.PeriodicJob{
    river.NewPeriodicJob(
        river.PeriodicInterval(24*time.Hour),
        func() (river.JobArgs, *river.InsertOpts) {
            return DailyResetArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: false},
    ),
}
```

### 사용량 조회 API

```
GET /v1/usage
Authorization: Bearer <jwt>

Response:
{
  "data": {
    "plan_type": "starter",
    "analysis": {
      "used": 3,
      "limit": 10,
      "daily_used": 2,
      "daily_limit": 5
    },
    "coaching": {
      "used": 1,
      "limit": 10,
      "daily_used": 1,
      "daily_limit": 5
    },
    "plan_expires_at": "2026-08-15T00:00:00+09:00",
    "daily_resets_at": "2026-03-16T00:00:00+09:00"
  }
}
```
