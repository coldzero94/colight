# Phase 1.7: 어드민 관찰성 대시보드

> Sprint 9-10 | 예상 공수: 6일 | 관련 기능: 어드민 대시보드, AI 에러 추적, 모니터링 통합

## 개요

| 항목 | 내용 |
| ---- | ---- |
| **목표** | 어드민 대시보드를 운영 허브로 재구축. AI 에러 분류 체계 수립, Rate Limit 이벤트 영속화, 사용량·모델 페이지 통합, recharts 기반 시각화 및 에러 인라인 상세 UX 구현 |
| **선행 조건** | Phase 1.6 (어드민 시스템 확장) |
| **주요 산출물** | `ai_call_errors` 테이블(신설), `quota_hit_events` 테이블(신설), `usage_logs.error_message` 컬럼 제거, 대시보드 통합 API, 에러 목록 API, 리빌드된 대시보드 UI, 사용량+모델 통합 페이지 |
| **기획 문서** | 본 문서 |

---

## 진행 상태

| Step | 이름 | 상태 |
| ---- | ---- | ---- |
| 1.7.1 | DB 스키마 보강 (ai_call_errors + quota_hit_events) | ⬜ 대기 |
| 1.7.2 | AI 에러 분류 & 영속화 | ⬜ 대기 |
| 1.7.3 | 대시보드 통합 API (`GET /v1/admin/dashboard`) | ⬜ 대기 |
| 1.7.4 | 에러 목록 API (`GET /v1/admin/usage/errors`) | ⬜ 대기 |
| 1.7.5 | TypeSpec 업데이트 | ⬜ 대기 |
| 1.7.6 | 사용량 + 모델 페이지 통합 (`/admin/usage` 탭 구조) | ⬜ 대기 |
| 1.7.7 | 대시보드 페이지 리빌드 | ⬜ 대기 |

---

## DB 설계

### 정규화 원칙

현재 `usage_logs`에 `error_message` nullable 컬럼이 있어 성공 호출(전체의 ~97%)에도 항상 NULL 컬럼이 존재한다. migration 제약이 없으므로 에러 데이터를 별도 테이블로 분리한다.

| 테이블 | 역할 | 이유 |
| ------ | ---- | ---- |
| `usage_logs` | AI 호출 1건 = 1행, 집계 기준 | nullable error 컬럼 제거로 스키마 순수화 |
| `ai_call_errors` | 실패한 호출의 에러 상세 (1:0..1) | 성공 호출에 불필요한 컬럼 격리 |
| `quota_hit_events` | Rate limit 이벤트 전용 스트림 | 인메모리→DB 영속화, 조회 패턴 분리 |

### 관계도

```text
user_profiles (1)
    └── (N) usage_logs                    status: "success" | "error"
                └── (0..1) ai_call_errors  error_type + error_message
                               │
                               │ usage_log_id (nullable FK)
                               ▼
                    quota_hit_events       rate_limit 이벤트 전용
                    (usage_log_id → usage_logs.id 경유)
```

**3NF 검증**:

- `usage_logs`: 에러 컬럼 없음, 모든 속성이 PK에만 종속 ✓
- `ai_call_errors`: `error_type`과 `error_message`는 서로 독립적으로 PK에 종속 (같은 타입이라도 메시지가 다름) ✓
- `quota_hit_events`: provider·model 간 이행 종속 없음 (같은 provider에 여러 model 가능) ✓

---

### 1. `usage_logs` 테이블 — `error_message` 제거

기존 nullable `error_message` 컬럼을 `ai_call_errors` 테이블로 이전 후 삭제.
`status` 컬럼은 유지 (에러 count 집계 시 JOIN 없이 빠르게 조회 가능).

```sql
-- Step 1: 신규 테이블 생성 (아래 §2 참고)
-- Step 2: 기존 에러 데이터 마이그레이션 (error_type은 'unknown'으로 일괄 초기화)
INSERT INTO ai_call_errors (usage_log_id, error_type, error_message)
SELECT id, 'unknown', error_message
FROM usage_logs
WHERE status = 'error' AND error_message IS NOT NULL;

-- Step 3: 컬럼 제거
ALTER TABLE usage_logs DROP COLUMN error_message;
```

**Ent 스키마 변경 (`ent/schema/usagelog.go`)**:

- `field.Text("error_message")` 블록 삭제
- 에러 상세 접근: `usageLog.QueryErrorDetail()` (새 엣지)

---

### 2. `ai_call_errors` 테이블 — 신설

에러가 발생한 usage_log에 1:1로 연결되는 확장 테이블.
`usage_log_id`를 UNIQUE FK로 사용 (Ent 패턴상 UUID PK 별도 유지).

```sql
CREATE TABLE ai_call_errors (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usage_log_id UUID NOT NULL UNIQUE REFERENCES usage_logs(id) ON DELETE CASCADE,
    error_type   VARCHAR(30) NOT NULL,
    error_message TEXT       NOT NULL,
    created_at   TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_call_errors_type       ON ai_call_errors (error_type, created_at DESC);
CREATE INDEX idx_ai_call_errors_created_at ON ai_call_errors (created_at DESC);
```

**error_type 값 정의**:

| 값 | 설명 | 예시 |
| -- | ---- | ---- |
| `rate_limit` | 분당/일당 요청 한도 초과 (HTTP 429) | Gemini RPM 초과, Groq RPD 초과 |
| `timeout` | 응답 시간 초과 | 30초 이상 무응답 |
| `provider_error` | 프로바이더 내부 오류 (HTTP 5xx) | Gemini 503, Groq 500 |
| `invalid_request` | 잘못된 요청 (HTTP 4xx, non-429) | 프롬프트 형식 오류 |
| `context_exceeded` | 입력 토큰이 모델 컨텍스트 한도 초과 | 긴 경험 기술서 처리 시 |
| `unknown` | 마이그레이션된 기존 에러 (미분류) | 레거시 데이터 |

**Ent 스키마 (`ent/schema/aicallerror.go`)**:

```go
type AICallError struct{ ent.Schema }

func (AICallError) Mixin() []ent.Mixin {
    return []ent.Mixin{TimestampMixin{}}
}

func (AICallError) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("usage_log_id", uuid.UUID{}).
            Unique().
            Comment("1:1 FK to usage_logs — CASCADE on delete"),
        field.Enum("error_type").
            Values("rate_limit", "timeout", "provider_error",
                "invalid_request", "context_exceeded", "unknown").
            Comment("Classified AI error category"),
        field.Text("error_message").
            Comment("Raw error message from provider"),
    }
}

func (AICallError) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("usage_log", UsageLog.Type).
            Ref("error_detail").
            Field("usage_log_id").
            Unique().
            Required().
            Annotations(entsql.OnDelete(entsql.Cascade)),
    }
}

func (AICallError) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("error_type", "created_at"),
        index.Fields("created_at"),
    }
}
```

**`ent/schema/usagelog.go` 엣지 추가**:

```go
// Edges 메서드에 추가
edge.To("error_detail", AICallError.Type).Unique(),
```

---

### 3. `quota_hit_events` 테이블 — 신설

Rate limit 이벤트 전용 스트림. 현재 인메모리 200개 링버퍼를 DB로 영속화.

- `usage_log_id`는 nullable: 호출 전 사전 throttle(미래 기능) 케이스 대비
- `error_message`는 `ai_call_errors`와 중복이나 의도적 비정규화 — quota_hit_events는 usage_log가 없는 경우도 있고 독립적 집계 패턴으로 사용됨

```sql
CREATE TABLE quota_hit_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider      VARCHAR(20)  NOT NULL,
    model         VARCHAR(100) NOT NULL,
    feature       VARCHAR(30)  NULL,
    error_message TEXT         NOT NULL,
    usage_log_id  UUID         NULL REFERENCES usage_logs(id) ON DELETE SET NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quota_hits_provider_model ON quota_hit_events (provider, model, created_at DESC);
CREATE INDEX idx_quota_hits_created_at     ON quota_hit_events (created_at DESC);
```

**Ent 스키마 (`ent/schema/quotahitevent.go`)**:

```go
type QuotaHitEvent struct{ ent.Schema }

func (QuotaHitEvent) Mixin() []ent.Mixin {
    return []ent.Mixin{TimestampMixin{}}
}

func (QuotaHitEvent) Fields() []ent.Field {
    return []ent.Field{
        field.String("provider").MaxLen(20),
        field.String("model").MaxLen(100),
        field.String("feature").MaxLen(30).Optional().Nillable(),
        field.Text("error_message"),
        field.UUID("usage_log_id", uuid.UUID{}).Optional().Nillable().
            Comment("FK to usage_logs, nullable for pre-throttle events"),
    }
}

func (QuotaHitEvent) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("usage_log", UsageLog.Type).
            Field("usage_log_id").Unique().Optional(),
    }
}

func (QuotaHitEvent) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("provider", "model", "created_at"),
        index.Fields("created_at"),
    }
}
```

---

## API 설계

### 신규 엔드포인트 2개

#### `GET /v1/admin/dashboard`

대시보드 전용 집계 API. 기존 `/v1/admin/stats`, `/v1/admin/usage/summary` 등을 개별 호출하는 대신 단일 응답으로 통합. 클라이언트 N+1 문제 해소.

**응답 구조**:

```text
DashboardData
├── user_stats
│   ├── total_users
│   ├── new_users_today
│   └── active_users_today
├── ai_metrics_today
│   ├── calls
│   ├── calls_delta_pct   // 어제 대비 증감율 (%)
│   ├── error_rate
│   ├── error_rate_delta  // 어제 대비 변화 (pp)
│   └── cost_krw
├── daily_metrics[]       // 최근 7일 차트 데이터
│   ├── date
│   ├── calls
│   ├── error_count
│   └── cost_krw
├── quota_alerts[]        // RPD 80% 이상 모델 경보
│   ├── model
│   ├── provider
│   ├── used_pct
│   └── remaining
├── recent_errors[]       // 최근 에러 5건 (인라인 상세 포함)
│   ├── id
│   ├── error_type
│   ├── error_message
│   ├── model
│   ├── provider
│   ├── feature
│   ├── user_id
│   ├── input_tokens
│   ├── output_tokens
│   └── created_at
├── recent_feedbacks[]    // 미처리 피드백 최근 3건
│   ├── id
│   ├── category
│   ├── content_preview   // 최대 100자
│   └── created_at
└── pending_feedback_count
```

---

#### `GET /v1/admin/usage/errors`

에러 목록 전용 API. 사용량 페이지 "에러" 탭에서 페이지네이션+필터 지원.

**쿼리 파라미터**:

| 파라미터 | 타입 | 기본값 | 설명 |
| -------- | ---- | ------ | ---- |
| `error_type` | string | (전체) | rate_limit \| timeout \| provider_error \| invalid_request \| context_exceeded |
| `provider` | string | (전체) | gemini \| groq |
| `feature` | string | (전체) | experience \| analysis \| draft \| review 등 |
| `days` | int | 7 | 조회 기간 |
| `limit` | int | 20 | 페이지 크기 |
| `offset` | int | 0 | 페이지 오프셋 |

**응답**: `{ data: AIErrorItem[], count: int32 }`

---

### 기존 엔드포인트 유지 (변경 없음)

- `GET /v1/admin/usage/summary` — 사용량 탭 요약 카드용
- `GET /v1/admin/usage/daily` — 사용량 탭 일별 테이블용
- `GET /v1/admin/usage/costs` — 사용량 탭 프로바이더 비용용
- `GET /v1/admin/usage/top-users` — 사용량 탭 상위 사용자용
- `GET /v1/admin/models/stats` — 모델 탭용
- `GET /v1/admin/models/rate-limit-hits` — 모델 탭 quota 알림용 (in-memory + DB 병행)

---

## TypeSpec 업데이트 (`packages/protocol/src/admin/admin.tsp`)

기존 파일에 아래 모델 및 엔드포인트를 추가한다. 기존 선언은 수정하지 않는다.

```typespec
// ─── 신규 모델 ───

model UserStatsSummary {
  @encodedName("application/json", "total_users")
  totalUsers: int32;

  @encodedName("application/json", "new_users_today")
  newUsersToday: int32;

  @encodedName("application/json", "active_users_today")
  activeUsersToday: int32;
}

model AIMetricsToday {
  calls: int32;

  @encodedName("application/json", "calls_delta_pct")
  callsDeltaPct: float32;

  @encodedName("application/json", "error_rate")
  errorRate: float32;

  @encodedName("application/json", "error_rate_delta")
  errorRateDelta: float32;

  @encodedName("application/json", "cost_krw")
  costKrw: float32;
}

model DailyMetric {
  date: string;
  calls: int32;

  @encodedName("application/json", "error_count")
  errorCount: int32;

  @encodedName("application/json", "cost_krw")
  costKrw: float32;
}

model QuotaAlert {
  model: string;
  provider: string;

  @encodedName("application/json", "used_pct")
  usedPct: float32;

  remaining: int32;
}

model AIErrorItem {
  id: string;

  @encodedName("application/json", "error_type")
  errorType: "rate_limit" | "timeout" | "provider_error" | "invalid_request" | "context_exceeded";

  @encodedName("application/json", "error_message")
  errorMessage: string;

  model: string;
  provider: string;
  feature: string;

  @encodedName("application/json", "user_id")
  userId: string;

  @encodedName("application/json", "input_tokens")
  inputTokens: int32;

  @encodedName("application/json", "output_tokens")
  outputTokens: int32;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model FeedbackPreview {
  id: string;
  category: "bug" | "improvement" | "other";

  @encodedName("application/json", "content_preview")
  contentPreview: string;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model DashboardData {
  @encodedName("application/json", "user_stats")
  userStats: UserStatsSummary;

  @encodedName("application/json", "ai_metrics_today")
  aiMetricsToday: AIMetricsToday;

  @encodedName("application/json", "daily_metrics")
  dailyMetrics: DailyMetric[];

  @encodedName("application/json", "quota_alerts")
  quotaAlerts: QuotaAlert[];

  @encodedName("application/json", "recent_errors")
  recentErrors: AIErrorItem[];

  @encodedName("application/json", "recent_feedbacks")
  recentFeedbacks: FeedbackPreview[];

  @encodedName("application/json", "pending_feedback_count")
  pendingFeedbackCount: int32;
}

// ─── 기존 AdminAPI interface에 추가 ───

/** Get dashboard aggregated data */
@get
@route("/dashboard")
getDashboard(): {
  @statusCode statusCode: 200;
  @body body: { data: DashboardData };
} | Colight.Common.ErrorResponse;

/** List AI errors with filters */
@get
@route("/usage/errors")
listErrors(
  @query error_type?: "rate_limit" | "timeout" | "provider_error" | "invalid_request" | "context_exceeded",
  @query provider?: string,
  @query feature?: string,
  @query days?: int32 = 7,
  @query limit?: int32 = 20,
  @query offset?: int32 = 0,
): {
  @statusCode statusCode: 200;
  @body body: { data: AIErrorItem[]; count: int32 };
} | Colight.Common.ErrorResponse;
```

---

## 구현 단계

### 1.7.1 DB 스키마 보강

**목표**: `usage_logs.error_message` 분리 → `ai_call_errors` 신설, `quota_hit_events` 신설

**테스트 명세**:

| 테스트 | 파일 | 설명 |
| ------ | ---- | ---- |
| `TestAICallError_Create` | `internal/repository/ai_call_error_repo_test.go` | ai_call_errors 레코드 생성 및 조회 |
| `TestAICallError_UsageLogEdge` | `internal/repository/ai_call_error_repo_test.go` | usage_log_id FK 연결 및 QueryUsageLog() 검증 |
| `TestAICallError_CascadeOnDelete` | `internal/repository/ai_call_error_repo_test.go` | usage_log 삭제 시 ai_call_error CASCADE 삭제 |
| `TestUsageLog_NoErrorMessageField` | `internal/repository/usage_log_repo_test.go` | usage_logs에 error_message 없음 확인 |
| `TestQuotaHitEvent_Create` | `internal/repository/quota_hit_repo_test.go` | quota_hit_events 레코드 생성 |
| `TestQuotaHitEvent_NullableUsageLogId` | `internal/repository/quota_hit_repo_test.go` | usage_log_id null 허용 검증 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `ent/schema/usagelog.go` — `error_message` 필드 삭제, `error_detail` 엣지 추가
  - [ ] **DB**: `ent/schema/aicallerror.go` 신설 (위 Ent 스키마 참고)
  - [ ] **DB**: `ent/schema/quotahitevent.go` 신설
  - [ ] **DB**: `moon run backend:generate-ent`
  - [ ] **DB**: `moon run backend:migrate-diff -- name=extract_ai_call_errors_quota_hit_events`
  - [ ] **DB**: 마이그레이션 파일에 데이터 이전 SQL 수동 추가 (기존 error_message → ai_call_errors)
- [ ] 테스트 통과 확인

**산출물**:

- `ent/schema/usagelog.go` (수정)
- `ent/schema/aicallerror.go` (신설)
- `ent/schema/quotahitevent.go` (신설)
- `migrations/` 새 마이그레이션 파일 (데이터 이전 SQL 포함)

---

### 1.7.2 AI 에러 분류 & 영속화

**목표**: AI 호출 실패 시 error_type을 분류하여 `ai_call_errors`에 저장, rate_limit은 `quota_hit_events`에도 기록

**분류 로직** (`internal/infrastructure/ai/`):

```text
에러 분류 기준:
- HTTP 429 or "quota" 포함          → rate_limit
- context.DeadlineExceeded or "timeout" 포함 → timeout
- HTTP 503/500 계열                  → provider_error
- "context length" or "too many tokens" 포함 → context_exceeded
- 기타 4xx                           → invalid_request
```

**테스트 명세**:

| 테스트 | 파일 | 설명 |
| ------ | ---- | ---- |
| `TestClassifyAIError_RateLimit` | `internal/infrastructure/ai/error_classifier_test.go` | 429 에러 → rate_limit 분류 |
| `TestClassifyAIError_Timeout` | `internal/infrastructure/ai/error_classifier_test.go` | deadline exceeded → timeout 분류 |
| `TestClassifyAIError_ProviderError` | `internal/infrastructure/ai/error_classifier_test.go` | 503 → provider_error 분류 |
| `TestClassifyAIError_ContextExceeded` | `internal/infrastructure/ai/error_classifier_test.go` | "context length" 포함 → context_exceeded 분류 |
| `TestClassifyAIError_Fallback` | `internal/infrastructure/ai/error_classifier_test.go` | 미분류 4xx → invalid_request |
| `TestLogUsage_CreatesAICallError` | `internal/infrastructure/ai/provider_test.go` | AI 호출 실패 시 ai_call_errors 레코드 생성됨 |
| `TestLogUsage_UsageLogLinkedToError` | `internal/infrastructure/ai/provider_test.go` | ai_call_errors.usage_log_id가 usage_logs.id를 참조 |
| `TestRateLimitHit_PersistsToQuotaHitEvents` | `internal/infrastructure/ai/provider_test.go` | rate_limit 에러 시 quota_hit_events에도 기록됨 |
| `TestRateLimitHit_UsageLogIdLinked` | `internal/infrastructure/ai/provider_test.go` | quota_hit_events.usage_log_id가 usage_logs.id를 참조 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **Backend**: `internal/infrastructure/ai/error_classifier.go` 신설
    - `ClassifyError(err error) string` — 에러 → error_type 문자열 반환
  - [ ] **Backend**: `internal/infrastructure/ai/provider.go` (또는 `gemini.go`, `groq.go`)
    - AI 호출 실패 경로에서 `ClassifyError()` 호출 → `usage_logs` (status=error) + `ai_call_errors` 동시 생성
  - [ ] **Backend**: `internal/repository/ai_call_error_repo.go` 신설
    - `Create(ctx, usageLogID, errorType, errorMessage)` 메서드
  - [ ] **Backend**: `internal/infrastructure/ai/provider.go`
    - `errorType == "rate_limit"` 시 `QuotaHitEventRepo.Create()` 추가 호출
  - [ ] **Backend**: `internal/repository/quota_hit_event_repo.go` 신설
    - `Create(ctx, provider, model, feature, errMsg, usageLogID)` 메서드
    - `ListRecent(ctx, hours int, limit int)` 메서드
- [ ] 테스트 통과 확인

**산출물**:

- `internal/infrastructure/ai/error_classifier.go` (신설)
- `internal/repository/ai_call_error_repo.go` (신설)
- `internal/repository/quota_hit_event_repo.go` (신설)
- AI provider 파일 수정

---

### 1.7.3 대시보드 통합 API

**목표**: `GET /v1/admin/dashboard` — 대시보드에 필요한 모든 데이터를 단일 응답으로 집계

**성능 고려**: 내부적으로 6개 DB 쿼리를 `errgroup` 병렬 실행. P95 < 200ms 목표.

```text
병렬 쿼리 구성:
1. 사용자 통계 (total, today new, today active)
2. 오늘 AI 지표 + 어제 AI 지표 (delta 계산용)
3. 최근 7일 일별 집계 (calls, errors, cost)
4. quota_hit_events + 현재 quota 상태 → 80% 이상 모델 필터
5. 최근 에러 5건 (ai_call_errors JOIN usage_logs ORDER BY created_at DESC LIMIT 5)
6. 미처리 피드백 (pending status, 최근 3건 + 전체 count)
```

**테스트 명세**:

| 테스트 | 파일 | 설명 |
| ------ | ---- | ---- |
| `TestGetDashboard_Returns200` | `internal/controller/admin_controller_test.go` | 정상 응답 200 및 구조 검증 |
| `TestGetDashboard_AIMetricsDelta` | `internal/controller/admin_controller_test.go` | calls_delta_pct: 어제 10건, 오늘 15건 → +50% |
| `TestGetDashboard_QuotaAlerts_AboveThreshold` | `internal/controller/admin_controller_test.go` | 80% 이상 모델만 quota_alerts에 포함 |
| `TestGetDashboard_RecentErrors_Limit5` | `internal/controller/admin_controller_test.go` | 에러 6건 있어도 5건만 반환 |
| `TestGetDashboard_EmptyState` | `internal/controller/admin_controller_test.go` | 데이터 없을 때 빈 배열 반환 (null 아님) |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **Backend**: `internal/service/admin_service.go` — `GetDashboard()` 메서드 (errgroup 병렬)
  - [ ] **Backend**: `internal/controller/admin_controller.go` — `GetDashboard()` 핸들러 추가
  - [ ] **Backend**: `cmd/api/main.go` — `GET /v1/admin/dashboard` 라우트 등록
- [ ] 테스트 통과 확인

**산출물**:

- `internal/service/admin_service.go` 수정
- `internal/controller/admin_controller.go` 수정

---

### 1.7.4 에러 목록 API

**목표**: `GET /v1/admin/usage/errors` — 사용량 페이지 에러 탭용 페이지네이션 API

**테스트 명세**:

| 테스트 | 파일 | 설명 |
| ------ | ---- | ---- |
| `TestListErrors_FilterByErrorType` | `internal/controller/admin_controller_test.go` | error_type 필터 동작 검증 |
| `TestListErrors_FilterByProvider` | `internal/controller/admin_controller_test.go` | provider 필터 동작 검증 |
| `TestListErrors_FilterByFeature` | `internal/controller/admin_controller_test.go` | feature 필터 동작 검증 |
| `TestListErrors_Pagination` | `internal/controller/admin_controller_test.go` | offset/limit 페이지네이션 |
| `TestListErrors_DaysFilter` | `internal/controller/admin_controller_test.go` | days 파라미터로 기간 필터 |
| `TestListErrors_OnlyFromAICallErrorsTable` | `internal/controller/admin_controller_test.go` | ai_call_errors 테이블 기반 — success 행 없음 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **Backend**: `internal/repository/ai_call_error_repo.go` — `List(ctx, filter)` 메서드 추가
  - [ ] **Backend**: `internal/service/admin_service.go` — `ListErrors()` 메서드
  - [ ] **Backend**: `internal/controller/admin_controller.go` — `ListErrors()` 핸들러
  - [ ] **Backend**: `cmd/api/main.go` — `GET /v1/admin/usage/errors` 라우트 등록
- [ ] 테스트 통과 확인

**산출물**:

- `internal/repository/ai_call_error_repo.go` 수정
- `internal/service/admin_service.go` 수정
- `internal/controller/admin_controller.go` 수정

---

### 1.7.5 TypeSpec 업데이트

**목표**: `packages/protocol/src/admin/admin.tsp`에 신규 모델 및 엔드포인트 추가 후 codegen 재실행

**구현 체크리스트**:

- [ ] `packages/protocol/src/admin/admin.tsp` 수정 (위 TypeSpec 섹션 참고)
- [ ] `moon run protocol:build` — OpenAPI 재생성
- [ ] `moon run backend:generate-oapi` — Go 서버 코드 재생성
- [ ] `moon run web:generate-client` — TS 클라이언트 재생성
- [ ] 생성 코드 빌드 에러 없는지 확인 (`moon run :build`)

---

### 1.7.6 사용량 + 모델 페이지 통합

**목표**: `/admin/models` 제거, `/admin/usage`를 3탭 구조로 재편, 사이드바 정리

**UX 설계**:

```text
/admin/usage
├── [탭] 사용량         ← 기존 usage/page.tsx 내용
│   ├── 요약 카드 (5개): 총 호출, 총 토큰, 총 비용, 에러율, 에러 수
│   ├── 7일 추이 차트 (AreaChart): calls + error_count 이중축
│   ├── 프로바이더별 비용 테이블
│   ├── 상위 사용자 테이블
│   └── 일별 사용량 테이블
│
├── [탭] 모델           ← 기존 models/page.tsx 내용
│   ├── Quota 상태 카드 (프로바이더별 그룹)
│   ├── Rate Limit 발생 알림 (빨간 박스, 있을 때만)
│   ├── 기능별 설정 모델
│   └── 모델별 사용 통계 테이블
│
└── [탭] 에러           ← 신규
    ├── 필터 바: error_type | provider | feature | 기간
    ├── 에러 목록 테이블 (20개/페이지)
    │   ├── 컬럼: 시각 | 타입 배지 | 모델 | 기능 | 사용자 | 토큰 | ▼
    │   └── 행 클릭 → 인라인 accordion 펼침 (아래 참고)
    └── 페이지네이션
```

**에러 인라인 상세 (accordion)**:

```text
▼ 펼쳐진 행
┌──────────────────────────────────────────────────────┐
│ error_message (전체, 스크롤 가능 텍스트박스 스타일)   │
│                                                      │
│ 모델: gemini-2.0-flash    프로바이더: gemini         │
│ 기능: draft               입력 토큰: 2,341           │
│ 사용자: abc12345...  [사용자 관리에서 보기 →]         │
│                                                      │
│ 발생 시각: 2026-02-19 14:32:11 KST                   │
└──────────────────────────────────────────────────────┘
```

**테스트 명세**:

| 테스트 | 파일 | 설명 |
| ------ | ---- | ---- |
| `UsagePage shows 3 tabs` | `src/app/(admin)/admin/usage/__tests__/page.test.tsx` | 사용량/모델/에러 탭 렌더링 |
| `Error tab renders filter bar` | `src/app/(admin)/admin/usage/__tests__/page.test.tsx` | 에러 탭 필터 렌더링 |
| `Error row click expands detail` | `src/app/(admin)/admin/usage/__tests__/page.test.tsx` | 행 클릭 시 accordion 토글 |
| `AdminSidebar no models menu` | `src/components/admin/__tests__/admin-sidebar.test.tsx` | 모델 통계 메뉴 항목 없음 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **Frontend**: `src/app/(admin)/admin/usage/page.tsx` — 탭 구조 리팩토링
    - `useSearchParams`로 `?tab=usage|models|errors` URL 상태 관리
    - 사용량 탭: 기존 내용 + AreaChart 추가 (recharts)
    - 모델 탭: 기존 `models/page.tsx` 내용 이전
    - 에러 탭: 필터 + 테이블 + 인라인 accordion
  - [ ] **Frontend**: `src/app/(admin)/admin/models/` 디렉토리 삭제
  - [ ] **Frontend**: `src/components/admin/admin-sidebar.tsx` — "모델 통계" 메뉴 제거
- [ ] 테스트 통과 확인

**산출물**:

- `src/app/(admin)/admin/usage/page.tsx` (전면 수정)
- `src/app/(admin)/admin/models/` (삭제)
- `src/components/admin/admin-sidebar.tsx` (수정)

---

### 1.7.7 대시보드 페이지 리빌드

**목표**: `/admin` 페이지를 운영 허브로 재구축

**UX 설계 — 레이아웃**:

```text
/admin 대시보드
┌──────────────────────────────────────────────────────────────┐
│ 대시보드                               [마지막 갱신: 14:32]  │
│                                        [새로고침 버튼]       │
├─────────────┬─────────────┬─────────────┬─────────────┬──────┤
│ 전체 사용자  │  오늘 AI    │  오늘 에러율 │  오늘 비용  │ 미처리│
│    347      │    호출     │             │             │피드백 │
│  ↑ +3 오늘  │   1,204     │    2.3%     │   ₩4,820    │   5  │
│             │  ↑ +18%     │  ↓ -0.4pp   │             │  건  │
└─────────────┴─────────────┴─────────────┴─────────────┴──────┘
│             [Quota 경보 배너 — 80% 이상 모델 있을 때만 표시] │
├──────────────────────────────┬───────────────────────────────┤
│ 7일 AI 호출 추이 (2/3)       │ 시스템 경보 (1/3)             │
│                              │                               │
│ AreaChart: calls (파랑)      │ [빨강] gemini-flash 87% RPD   │
│            errors (빨강)     │ [초록] 모든 DB 연결 정상       │
│                              │                               │
├──────────────────────────────┴───────────────────────────────┤
│ 최근 AI 에러 (2/3)           │ 미처리 피드백 (1/3)           │
│                              │                               │
│ [rate_limit] gemini-flash    │ [버그] 분석 중 오류 발생...   │
│   analysis  14:31  ▼         │ 2시간 전                      │
│                              │                               │
│ [timeout]   groq-llama       │ [개선] 경험 편집이 더 쉬웠    │
│   draft     13:55  ▼         │ 으면 좋겠어요  5시간 전       │
│                              │                               │
│ [에러 전체 보기 →]           │ [피드백 전체 보기 →]          │
└──────────────────────────────┴───────────────────────────────┘
```

**KPI 카드 UX 세부사항**:

- `calls_delta_pct > 0` → 초록 화살표 ↑, `< 0` → 빨강 화살표 ↓, `== 0` → 회색 →
- 에러율: `< 2%` 초록, `2~5%` 노랑, `> 5%` 빨강 (텍스트 색상)
- Quota 경보 배너: quota_alerts 배열이 비어있으면 렌더링 안 함

**에러 위젯 인라인 상세**:

- 에러 행 클릭 → 해당 행 아래에 accordion 슬라이드 인
- 펼쳐진 내용: error_message 전문 + model/provider/feature/user_id/tokens/timestamp
- user_id 클릭 → `/admin/users?search={userId}` 링크
- "에러 전체 보기 →" → `/admin/usage?tab=errors`

**테스트 명세**:

| 테스트 | 파일 | 설명 |
| ------ | ---- | ---- |
| `DashboardPage renders KPI cards` | `src/app/(admin)/admin/__tests__/page.test.tsx` | 5개 KPI 카드 렌더링 |
| `DashboardPage shows delta indicator` | `src/app/(admin)/admin/__tests__/page.test.tsx` | calls_delta_pct > 0 시 ↑ 표시 |
| `DashboardPage hides quota banner when empty` | `src/app/(admin)/admin/__tests__/page.test.tsx` | quota_alerts=[] 시 배너 없음 |
| `ErrorWidget row click toggles accordion` | `src/app/(admin)/admin/__tests__/page.test.tsx` | 에러 행 클릭 → 상세 토글 |
| `ErrorWidget second click collapses` | `src/app/(admin)/admin/__tests__/page.test.tsx` | 같은 행 다시 클릭 시 닫힘 |
| `ErrorWidget only one open at a time` | `src/app/(admin)/admin/__tests__/page.test.tsx` | 다른 행 클릭 시 이전 행 닫힘 |
| `DashboardPage links to usage errors tab` | `src/app/(admin)/admin/__tests__/page.test.tsx` | "에러 전체 보기" 링크 href 검증 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **Frontend**: `src/app/(admin)/admin/page.tsx` 전면 재작성
    - `GET /v1/admin/dashboard` 단일 호출로 모든 데이터 로드
    - recharts `AreaChart` — `calls` (fill: primary/30), `error_count` (fill: red/30)
    - 에러 위젯: `expandedErrorId` state로 accordion 관리
    - 자동 새로고침 없음 (수동 버튼, 운영자가 제어 가능하도록)
- [ ] 테스트 통과 확인

**산출물**:

- `src/app/(admin)/admin/page.tsx` (전면 재작성)

---

## 완료 게이트

```bash
moon run backend:lint && \
moon run backend:test && \
moon run web:lint && \
moon run web:typecheck && \
moon run web:test && \
moon run web:build
```

---

## 사이드바 최종 메뉴 구조

Phase 1.7 완료 후:

| 메뉴 | 경로 | minRole |
| ---- | ---- | ------- |
| 대시보드 | `/admin` | admin |
| 사용자 관리 | `/admin/users` | admin |
| 프롬프트 관리 | `/admin/prompts` | admin |
| ~~모델 통계~~ | ~~`/admin/models`~~ | ~~삭제~~ |
| 시스템 설정 | `/admin/settings` | super_admin |
| 사용량 모니터링 | `/admin/usage` | admin |
| 감사 로그 | `/admin/logs` | admin |
| 피드백 관리 | `/admin/feedbacks` | admin |
