# Colight - 알림 시스템 설계

> 작성일: 2026-02-12
> 인앱 알림, 이메일 알림, 사용자 설정, 발송 아키텍처

---

## 1. 알림 채널 정의

| 채널 | 구현 Phase | 설명 | 기술 |
|------|-----------|------|------|
| **인앱 알림** | Phase 6.1 | 서비스 내 벨 아이콘 + 드롭다운 | DB 기반 (notifications 테이블) |
| **이메일 알림** | Phase 6.2 | 서비스 알림 + 마케팅 | SendGrid API (Go 백엔드) |
| **푸시 알림** | Phase 10+ | 모바일 PWA 푸시 | Web Push API (향후 검토) |

> **Phase 10 이전까지는 인앱 + 이메일만 구현**한다. 푸시 알림은 PWA 도입 시 검토.

### 채널 선택 기준

```
[알림 이벤트 발생]
    │
    ├── 즉시 확인 필요 (분석 완료, 마감 D-1)
    │     → 인앱 + 이메일
    │
    ├── 참고 수준 (한도 80% 도달, 새 기능 안내)
    │     → 인앱만
    │
    └── 마케팅/시즌 안내
          → 이메일만 (사용자 수신 동의 시)
```

---

## 2. 알림 유형별 정리

### 2.1 시스템 알림

| 알림 ID | 이벤트 | 메시지 예시 | 채널 | 발생 시점 |
|---------|--------|-----------|------|----------|
| `SYS_001` | 서비스 점검 예정 | "내일 오전 2시~4시 서비스 점검이 예정되어 있습니다" | 인앱 + 이메일 | 점검 24시간 전 |
| `SYS_002` | 서비스 점검 완료 | "서비스 점검이 완료되었습니다" | 인앱 | 점검 종료 시 |
| `SYS_003` | 새 기능 출시 | "AI 인터뷰 기능이 추가되었습니다!" | 인앱 | 배포 직후 |
| `SYS_004` | 약관 변경 안내 | "서비스 이용약관이 변경되었습니다" | 인앱 + 이메일 | 변경 7일 전 |

### 2.2 분석 완료 알림

> 06-error-scenarios-ux.md의 대기 시간 UX와 연계

| 알림 ID | 이벤트 | 메시지 예시 | 채널 | 발생 시점 |
|---------|--------|-----------|------|----------|
| `ANALYSIS_001` | 기업 분석 완료 | "삼성전자 마케팅 직무 분석이 완료되었습니다" | 인앱 | 분석 완료 즉시 |
| `ANALYSIS_002` | 백그라운드 분석 완료 | "요청하신 분석이 완료되었습니다. 결과를 확인해 보세요" | 인앱 + 이메일 | River job 완료 시 |
| `ANALYSIS_003` | 초안 코칭 완료 | "자소서 초안이 준비되었습니다" | 인앱 | 코칭 완료 즉시 |

### 2.3 한도 경고 알림

> 10-fair-use-policy.md의 한도 초과 UX 정책과 연계

| 알림 ID | 이벤트 | 메시지 예시 | 채널 | 트리거 조건 |
|---------|--------|-----------|------|-----------|
| `LIMIT_001` | 총 한도 80% 도달 | "기업 분석 잔여 횟수가 2회 남았습니다" | 인앱 | `total_used / total_limit >= 0.8` |
| `LIMIT_002` | 총 한도 소진 | "기업 분석 횟수를 모두 사용했습니다" | 인앱 | `total_used >= total_limit` |
| `LIMIT_003` | 일일 한도 80% 도달 | "오늘 분석 한도 4/5회를 사용했습니다" | 인앱 | `daily_used / daily_limit >= 0.8` |
| `LIMIT_004` | 일일 한도 소진 | "오늘의 사용 한도를 모두 사용했습니다" | 인앱 | `daily_used >= daily_limit` |

### 2.4 만료 안내 알림

> 10-fair-use-policy.md의 만료 안내 타임라인과 연계

| 알림 ID | 이벤트 | 메시지 예시 | 채널 | 트리거 조건 |
|---------|--------|-----------|------|-----------|
| `EXPIRY_001` | 회차권 만료 D-30 | "회차권이 30일 후 만료됩니다" | 이메일 | `plan_expires_at - NOW() <= 30일` |
| `EXPIRY_002` | 회차권 만료 D-7 | "7일 후 잔여 N회가 소멸됩니다" | 인앱 + 이메일 | `plan_expires_at - NOW() <= 7일` |
| `EXPIRY_003` | 회차권 만료 D-1 | "내일 잔여 N회가 소멸됩니다" | 인앱 + 이메일 | `plan_expires_at - NOW() <= 1일` |
| `EXPIRY_004` | 회차권 만료 완료 | "회차권이 만료되었습니다. 재구매하기" | 인앱 + 이메일 | `plan_expires_at <= NOW()` |
| `EXPIRY_005` | 마감일 D-7 | "삼성전자 마감일이 7일 남았습니다" | 인앱 + 이메일 | `deadline - NOW() <= 7일` |
| `EXPIRY_006` | 마감일 D-1 | "내일이 마감일입니다!" | 인앱 | `deadline - NOW() <= 1일` |

### 2.5 활동 리마인더 알림

> 07-onboarding-flow.md의 재방문 유저 플로우 및 알림 정책과 연계

| 알림 ID | 이벤트 | 메시지 예시 | 채널 | 트리거 조건 |
|---------|--------|-----------|------|-----------|
| `REMIND_001` | 작성 중 자소서 리마인더 | "작성 중인 자소서가 기다리고 있어요" | 이메일 | 자소서 draft 상태 + 3일 미접속 |
| `REMIND_002` | 프로필 미완성 | "프로필을 완성하면 맞춤 추천 품질이 높아져요!" | 인앱 | 프로필 미완성 + 로그인 시 |
| `REMIND_003` | 7일 미접속 | "지금이 공채 시즌! 준비 시작해볼까요?" | 이메일 | 마지막 접속 >= 7일 |

### 2.6 마케팅 알림

| 알림 ID | 이벤트 | 메시지 예시 | 채널 | 빈도 |
|---------|--------|-----------|------|------|
| `MARKETING_001` | 공채 시즌 안내 | "상반기 공채 시즌이 시작되었습니다" | 이메일 | 시즌당 1회 |
| `MARKETING_002` | 프로모션/할인 | "스타터 회원 30% 할인 이벤트!" | 이메일 | 이벤트 시 |
| `MARKETING_003` | 기능 업데이트 | "새로운 AI 인터뷰 기능을 만나보세요" | 이메일 | 월 1회 이내 |

> **08-legal-privacy.md 참조**: 마케팅 알림은 가입 시 선택적 동의 필요. 모든 이메일 하단에 수신거부 링크 필수.

---

## 3. 알림 빈도 제한 정책 (Frequency Capping)

### 사용자당 일일 알림 상한

| 채널 | 일일 상한 | 비고 |
|------|----------|------|
| 인앱 | 20건/일 | 초과 시 최신 알림으로 교체 (오래된 알림 자동 읽음 처리) |
| 이메일 | 3건/일 | 긴급 알림(서비스 점검, 보안) 제외 |

### 중복 알림 방지

```
[알림 발송 전 중복 체크]
    │
    ├── 동일 알림 ID + 동일 대상 (예: EXPIRY_002 + 삼성전자)
    │     → 24시간 이내 이미 발송 → 스킵
    │
    ├── 동일 카테고리 알림 (예: LIMIT 계열)
    │     → 1시간 이내 동일 카테고리 발송 → 스킵
    │
    └── 마케팅 알림
          → 주 1회 이내 발송
          → 최근 7일 내 마케팅 이메일 발송 이력 확인
```

### Quiet Hours (방해 금지 시간)

| 항목 | 설정 |
|------|------|
| **기본 방해 금지** | 22:00 ~ 08:00 KST |
| **적용 대상** | 이메일만 (인앱은 로그인 시에만 표시되므로 제외) |
| **긴급 알림 예외** | 서비스 점검 안내, 보안 관련 알림은 방해 금지 무시 |
| **사용자 커스텀** | Phase 10+에서 사용자별 설정 가능하도록 확장 |

```
[이메일 발송 요청]
    │
    ├── 긴급 알림 → 즉시 발송
    │
    └── 일반 알림 → Quiet Hours 확인
          ├── 방해 금지 시간 아님 → 즉시 발송
          └── 방해 금지 시간 → 큐에 저장, 08:00 KST에 일괄 발송
```

---

## 4. 인앱 알림 UI

### 4.1 벨 아이콘 & 드롭다운

```
┌─────────────────────────────────────────────────────┐
│  Colight         [경험] [분석] [코칭]    🔔(3)  👤  │
│                                          │         │
│                                          ▼         │
│                    ┌──────────────────────────────┐ │
│                    │ 알림                  모두 읽음│ │
│                    ├──────────────────────────────┤ │
│                    │ ● 삼성전자 마케팅 직무         │ │
│                    │   분석이 완료되었습니다         │ │
│                    │   2분 전                      │ │
│                    ├──────────────────────────────┤ │
│                    │ ● 기업 분석 잔여 횟수가        │ │
│                    │   2회 남았습니다               │ │
│                    │   1시간 전                    │ │
│                    ├──────────────────────────────┤ │
│                    │ ○ 프로필을 완성하면 맞춤       │ │
│                    │   추천 품질이 높아져요!        │ │
│                    │   어제                        │ │
│                    ├──────────────────────────────┤ │
│                    │      [전체 알림 보기 →]       │ │
│                    └──────────────────────────────┘ │
└─────────────────────────────────────────────────────┘

● = 안읽음 (파란 점)
○ = 읽음
```

### 4.2 읽음/안읽음 상태 관리

| 동작 | 상태 변경 |
|------|----------|
| 드롭다운 열기 | 변경 없음 (열기만으로는 읽음 처리 안함) |
| 알림 항목 클릭 | 해당 알림 읽음 처리 + 관련 페이지 이동 |
| "모두 읽음" 클릭 | 모든 미읽음 알림 → 읽음 처리 |
| 알림 개별 삭제 | soft delete (is_deleted = true) |

### 4.3 알림 목록 페이지 (/notifications)

```
┌───────────────────────────────────────────────────────┐
│ 알림                                     [모두 읽음]   │
│                                                       │
│ 필터: [전체 ▾] [안읽음만 □]                             │
│                                                       │
│ ─── 오늘 ───────────────────────────────────────────  │
│ ● 삼성전자 마케팅 직무 분석이 완료되었습니다      2분 전  │
│ ● 기업 분석 잔여 횟수가 2회 남았습니다          1시간 전  │
│                                                       │
│ ─── 어제 ───────────────────────────────────────────  │
│ ○ 프로필을 완성하면 맞춤 추천 품질이 높아져요!    어제    │
│ ○ 작성 중인 자소서가 기다리고 있어요            어제    │
│                                                       │
│ ─── 이번 주 ────────────────────────────────────────  │
│ ○ 7일 후 잔여 5회가 소멸됩니다                 2/5     │
│                                                       │
│ [더 보기]                                             │
└───────────────────────────────────────────────────────┘
```

### 4.4 프론트엔드 구현 패턴

```typescript
// apps/web/src/hooks/use-notifications.ts

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

export function useNotifications() {
  return useQuery({
    queryKey: ['notifications'],
    queryFn: () => api.GET('/v1/notifications', { params: { limit: 20 } }),
    refetchInterval: 30_000, // 30초마다 폴링
  });
}

export function useUnreadCount() {
  return useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () => api.GET('/v1/notifications/unread-count'),
    refetchInterval: 30_000,
  });
}

export function useMarkAsRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.PATCH(`/v1/notifications/${id}/read`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}
```

> **폴링 vs WebSocket**: 초기(Phase 6.1)에는 30초 간격 폴링으로 구현. 사용자 수 증가 시 SSE 또는 WebSocket 도입 검토 (Phase 10+).

---

## 5. 이메일 알림

### 5.1 이메일 발송 인프라

| 항목 | 선택 | 이유 |
|------|------|------|
| **서비스** | SendGrid | 무료 100건/일, Go SDK 지원, 템플릿 관리 |
| **발송 방식** | River background job | 비동기 발송 (API 응답 지연 방지) |
| **발신자** | `noreply@colight.kr` | 커스텀 도메인 |
| **Reply-to** | `support@colight.kr` | 고객 문의 대응 |

### 5.2 이메일 템플릿

| 템플릿 ID | 용도 | 제목 패턴 | 변수 |
|-----------|------|----------|------|
| `tpl_analysis_complete` | 백그라운드 분석 완료 | "[Colight] {{company_name}} 분석이 완료되었습니다" | `company_name`, `analysis_url` |
| `tpl_expiry_warning` | 만료 경고 | "[Colight] 회차권 만료 {{days_left}}일 전 안내" | `days_left`, `remaining_count`, `upgrade_url` |
| `tpl_deadline_reminder` | 마감일 리마인더 | "[Colight] {{company_name}} 마감일이 {{days_left}}일 남았습니다" | `company_name`, `days_left`, `application_url` |
| `tpl_inactivity_remind` | 미접속 리마인더 | "[Colight] 작성 중인 자소서가 기다리고 있어요" | `cover_letter_title`, `resume_url` |
| `tpl_marketing_season` | 공채 시즌 안내 | "[Colight] {{season}} 공채 시즌이 시작되었습니다" | `season`, `cta_url` |

### 5.3 이메일 공통 구조

```
┌───────────────────────────────────────────────┐
│  Colight 로고                                  │
├───────────────────────────────────────────────┤
│                                               │
│  {{제목}}                                      │
│                                               │
│  {{본문 내용}}                                  │
│                                               │
│  ┌─────────────────────────────────────┐      │
│  │        [CTA 버튼: 확인하러 가기]       │      │
│  └─────────────────────────────────────┘      │
│                                               │
├───────────────────────────────────────────────┤
│  Colight — AI 자소서 코칭 플랫폼                │
│  [수신거부] | [알림 설정 변경]                    │
│  이 메일은 {{email}}으로 발송되었습니다.          │
└───────────────────────────────────────────────┘
```

### 5.4 수신 거부 (Opt-out)

- **이메일 하단**: 원클릭 수신거부 링크 (`/unsubscribe?token={{unsubscribe_token}}`)
- **List-Unsubscribe 헤더**: RFC 8058 준수 (이메일 클라이언트 기본 수신거부 지원)
- **수신거부 처리**: 토큰 검증 → `notification_settings` 업데이트 → 확인 페이지 표시
- **08-legal-privacy.md 참조**: 마케팅 수신 동의는 가입 시 선택적, 수신거부 시 즉시 반영

---

## 6. DB 스키마

### 6.1 notifications 테이블

```go
// apps/backend/ent/schema/notification.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type Notification struct {
    ent.Schema
}

func (Notification) Mixin() []ent.Mixin {
    return []ent.Mixin{
        BaseMixin{},
    }
}

func (Notification) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("user_id", uuid.UUID{}).
            Comment("References auth.users(id)"),
        field.String("notification_type").
            NotEmpty().
            MaxLen(30).
            Comment("Notification type: system, analysis, limit, expiry, remind, marketing"),
        field.String("notification_id").
            NotEmpty().
            MaxLen(30).
            Comment("Notification ID: SYS_001, ANALYSIS_001, etc."),
        field.String("title").
            NotEmpty().
            MaxLen(200).
            Comment("Notification title"),
        field.Text("body").
            Optional().
            Comment("Notification body text"),
        field.String("link").
            Optional().
            MaxLen(500).
            Comment("Deep link URL path (e.g., /analysis/uuid)"),
        field.JSON("metadata", map[string]interface{}{}).
            Optional().
            Comment("Additional data: company_name, deadline, etc."),
        field.Bool("is_read").
            Default(false).
            Comment("Whether notification has been read"),
        field.Bool("is_deleted").
            Default(false).
            Comment("Soft delete flag"),
        field.Enum("channel").
            Values("in_app", "email", "both").
            Default("in_app").
            Comment("Notification channel"),
        field.Bool("email_sent").
            Default(false).
            Comment("Whether email has been sent"),
        field.Time("email_sent_at").
            Optional().
            Nillable().
            Comment("Email sent timestamp"),
    }
}

func (Notification) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", UserProfile.Type).
            Ref("notifications").
            Unique().
            Required().
            Field("user_id").
            Annotations(entsql.OnDelete(entsql.Cascade)),
    }
}

func (Notification) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id", "is_read", "is_deleted"),
        index.Fields("user_id", "created_at"),
        index.Fields("user_id", "notification_type"),
        index.Fields("notification_id"),
    }
}
```

### 6.2 notification_settings 테이블

```go
// apps/backend/ent/schema/notificationsetting.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type NotificationSetting struct {
    ent.Schema
}

func (NotificationSetting) Mixin() []ent.Mixin {
    return []ent.Mixin{
        BaseMixin{},
    }
}

func (NotificationSetting) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("user_id", uuid.UUID{}).
            Unique().
            Comment("References auth.users(id)"),
        // Category toggles
        field.Bool("system_in_app").
            Default(true).
            Comment("System notifications: in-app"),
        field.Bool("system_email").
            Default(true).
            Comment("System notifications: email"),
        field.Bool("analysis_in_app").
            Default(true).
            Comment("Analysis complete: in-app"),
        field.Bool("analysis_email").
            Default(true).
            Comment("Analysis complete: email"),
        field.Bool("limit_in_app").
            Default(true).
            Comment("Usage limit warnings: in-app"),
        field.Bool("limit_email").
            Default(false).
            Comment("Usage limit warnings: email"),
        field.Bool("expiry_in_app").
            Default(true).
            Comment("Expiry notifications: in-app"),
        field.Bool("expiry_email").
            Default(true).
            Comment("Expiry notifications: email"),
        field.Bool("remind_in_app").
            Default(true).
            Comment("Activity reminders: in-app"),
        field.Bool("remind_email").
            Default(true).
            Comment("Activity reminders: email"),
        field.Bool("marketing_email").
            Default(false).
            Comment("Marketing emails (opt-in required)"),
        // Quiet hours
        field.Int("quiet_start_hour").
            Default(22).
            Comment("Quiet hours start (0-23, KST)"),
        field.Int("quiet_end_hour").
            Default(8).
            Comment("Quiet hours end (0-23, KST)"),
    }
}

func (NotificationSetting) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", UserProfile.Type).
            Ref("notification_settings").
            Unique().
            Required().
            Field("user_id").
            Annotations(entsql.OnDelete(entsql.Cascade)),
    }
}

func (NotificationSetting) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id"),
    }
}
```

---

## 7. 알림 발송 아키텍처

### 7.1 전체 흐름

```
[이벤트 발생]
    │ (분석 완료, 한도 도달, 스케줄 트리거 등)
    │
    ▼
[NotificationService.Send()]
    │
    ├── 1. 사용자 알림 설정 조회 (notification_settings)
    │     └── 해당 유형 + 채널이 비활성화 → 스킵
    │
    ├── 2. 빈도 제한 확인
    │     ├── 동일 알림 24시간 이내 발송 이력 → 스킵
    │     └── 일일 상한 초과 → 스킵
    │
    ├── 3. 인앱 알림 생성
    │     └── notifications 테이블 INSERT
    │
    └── 4. 이메일 발송 (필요 시)
          │
          ├── Quiet Hours 확인
          │     ├── 방해 금지 → River delayed job (08:00 KST 발송)
          │     └── 방해 금지 아님 → 즉시 발송
          │
          └── River job 생성: SendEmailArgs{...}
                │
                └── [River Worker] → SendGrid API 호출
                      ├── 성공 → notifications.email_sent = true
                      └── 실패 → 재시도 (최대 3회, 지수 백오프)
```

### 7.2 Go 서비스 구현 패턴

```go
// apps/backend/internal/service/notification.go
package service

import (
    "context"
    "log/slog"
    "time"

    "github.com/google/uuid"
)

type NotificationType string

const (
    NotifSystem    NotificationType = "system"
    NotifAnalysis  NotificationType = "analysis"
    NotifLimit     NotificationType = "limit"
    NotifExpiry    NotificationType = "expiry"
    NotifRemind    NotificationType = "remind"
    NotifMarketing NotificationType = "marketing"
)

type SendNotificationParams struct {
    UserID           uuid.UUID
    NotificationType NotificationType
    NotificationID   string
    Title            string
    Body             string
    Link             string
    Metadata         map[string]interface{}
}

type NotificationService struct {
    logger   *slog.Logger
    db       *ent.Client
    river    *river.Client[pgx.Tx]
}

func (s *NotificationService) Send(ctx context.Context, params SendNotificationParams) error {
    // 1. Load user notification settings
    settings, err := s.db.NotificationSetting.
        Query().
        Where(notificationsetting.UserIDEQ(params.UserID)).
        Only(ctx)
    if err != nil {
        // Settings not found — use defaults
        settings = defaultSettings(params.UserID)
    }

    // 2. Check if notification type is enabled
    inAppEnabled := s.isInAppEnabled(settings, params.NotificationType)
    emailEnabled := s.isEmailEnabled(settings, params.NotificationType)

    if !inAppEnabled && !emailEnabled {
        return nil // User opted out of this notification type
    }

    // 3. Frequency capping check
    if s.isDuplicate(ctx, params.UserID, params.NotificationID) {
        s.logger.Debug("notification skipped (duplicate)",
            slog.String("notification_id", params.NotificationID),
        )
        return nil
    }

    // 4. Create in-app notification
    channel := "in_app"
    if emailEnabled {
        channel = "both"
    }

    notif, err := s.db.Notification.Create().
        SetUserID(params.UserID).
        SetNotificationType(string(params.NotificationType)).
        SetNotificationID(params.NotificationID).
        SetTitle(params.Title).
        SetBody(params.Body).
        SetLink(params.Link).
        SetMetadata(params.Metadata).
        SetChannel(notification.Channel(channel)).
        Save(ctx)
    if err != nil {
        return err
    }

    // 5. Enqueue email job if needed
    if emailEnabled {
        _, err = s.river.Insert(ctx, SendEmailArgs{
            NotificationDBID: notif.ID,
            UserID:           params.UserID,
            TemplateID:       s.resolveTemplate(params.NotificationID),
            Variables:        params.Metadata,
        }, nil)
        if err != nil {
            s.logger.Error("failed to enqueue email job",
                slog.String("error", err.Error()),
            )
        }
    }

    return nil
}
```

### 7.3 River Worker: 이메일 발송

```go
// apps/backend/internal/worker/send_email.go
package worker

import (
    "context"
    "log/slog"
    "time"

    "github.com/riverqueue/river"
    "github.com/sendgrid/sendgrid-go"
)

type SendEmailArgs struct {
    NotificationDBID uuid.UUID              `json:"notification_db_id"`
    UserID           uuid.UUID              `json:"user_id"`
    TemplateID       string                 `json:"template_id"`
    Variables        map[string]interface{} `json:"variables"`
}

func (SendEmailArgs) Kind() string { return "send_email" }

type SendEmailWorker struct {
    river.WorkerDefaults[SendEmailArgs]
    logger   *slog.Logger
    db       *ent.Client
    sgClient *sendgrid.Client
}

func (w *SendEmailWorker) Work(ctx context.Context, job *river.Job[SendEmailArgs]) error {
    // 1. Get user email
    user, err := w.db.UserProfile.
        Query().
        Where(userprofile.IDEQ(job.Args.UserID)).
        Only(ctx)
    if err != nil {
        return err
    }

    // 2. Check quiet hours
    now := time.Now().In(time.FixedZone("KST", 9*3600))
    hour := now.Hour()
    if hour >= 22 || hour < 8 {
        // Re-enqueue for 08:00 KST
        // Return river.JobSnooze to delay
        return river.JobSnooze(nextMorning(now))
    }

    // 3. Send via SendGrid
    err = w.sendEmail(ctx, user.Email, job.Args.TemplateID, job.Args.Variables)
    if err != nil {
        w.logger.Error("email send failed",
            slog.String("user_id", job.Args.UserID.String()),
            slog.String("error", err.Error()),
        )
        return err // River will retry with backoff
    }

    // 4. Update notification record
    _, err = w.db.Notification.
        UpdateOneID(job.Args.NotificationDBID).
        SetEmailSent(true).
        SetEmailSentAt(time.Now()).
        Save(ctx)

    return err
}
```

### 7.4 스케줄 기반 알림 (River Periodic Jobs)

```go
// River scheduler registration (cmd/api/main.go)

periodicJobs := []*river.PeriodicJob{
    // 만료 안내 체크: 매일 09:00 KST (00:00 UTC)
    river.NewPeriodicJob(
        river.PeriodicInterval(24*time.Hour),
        func() (river.JobArgs, *river.InsertOpts) {
            return CheckExpiryArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: false},
    ),

    // 마감일 리마인더 체크: 매일 09:00 KST
    river.NewPeriodicJob(
        river.PeriodicInterval(24*time.Hour),
        func() (river.JobArgs, *river.InsertOpts) {
            return CheckDeadlineArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: false},
    ),

    // 미접속 리마인더: 매일 10:00 KST
    river.NewPeriodicJob(
        river.PeriodicInterval(24*time.Hour),
        func() (river.JobArgs, *river.InsertOpts) {
            return CheckInactivityArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: false},
    ),
}
```

---

## 8. 사용자 설정 UI

### 8.1 알림 설정 페이지 (/settings/notifications)

```
┌───────────────────────────────────────────────────────┐
│ 알림 설정                                              │
│                                                       │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 시스템 알림                                      │   │
│ │ 서비스 점검, 약관 변경 등                          │   │
│ │ 인앱: [✓]  이메일: [✓]                           │   │
│ └─────────────────────────────────────────────────┘   │
│                                                       │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 분석 완료                                        │   │
│ │ 기업 분석, 코칭 결과 완료 시                       │   │
│ │ 인앱: [✓]  이메일: [✓]                           │   │
│ └─────────────────────────────────────────────────┘   │
│                                                       │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 사용량 한도                                      │   │
│ │ 잔여 횟수 부족, 일일 한도 도달                     │   │
│ │ 인앱: [✓]  이메일: [ ]                           │   │
│ └─────────────────────────────────────────────────┘   │
│                                                       │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 만료 안내                                        │   │
│ │ 회차권 만료, 마감일 리마인더                       │   │
│ │ 인앱: [✓]  이메일: [✓]                           │   │
│ └─────────────────────────────────────────────────┘   │
│                                                       │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 활동 리마인더                                    │   │
│ │ 미완성 자소서, 미접속 리마인더                     │   │
│ │ 인앱: [✓]  이메일: [✓]                           │   │
│ └─────────────────────────────────────────────────┘   │
│                                                       │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 마케팅                                           │   │
│ │ 공채 시즌 안내, 프로모션, 기능 업데이트             │   │
│ │ 이메일: [ ]                                      │   │
│ └─────────────────────────────────────────────────┘   │
│                                                       │
│                              [저장]                    │
└───────────────────────────────────────────────────────┘
```

### 8.2 API 엔드포인트

| Method | Path | 설명 |
|--------|------|------|
| `GET` | `/v1/notifications` | 알림 목록 조회 (pagination, 필터) |
| `GET` | `/v1/notifications/unread-count` | 미읽음 개수 |
| `PATCH` | `/v1/notifications/:id/read` | 읽음 처리 |
| `POST` | `/v1/notifications/mark-all-read` | 전체 읽음 처리 |
| `DELETE` | `/v1/notifications/:id` | 알림 삭제 (soft delete) |
| `GET` | `/v1/notification-settings` | 알림 설정 조회 |
| `PUT` | `/v1/notification-settings` | 알림 설정 업데이트 |

---

## 9. Phase별 구현 계획

| Phase | 구현 내용 | 알림 유형 |
|-------|----------|----------|
| **Phase 6.1** | 인앱 알림 기본 인프라 (DB, API, 벨 아이콘 UI) | `SYS_001~003` (시스템) |
| **Phase 6.1** | 한도 경고 인앱 알림 | `LIMIT_001~004` |
| **Phase 6.2** | 이메일 알림 인프라 (SendGrid, River worker) | `EXPIRY_001~004` (만료), `MARKETING_001` (시즌) |
| **Phase 6.2** | 수신거부, 알림 설정 페이지 | 사용자 설정 전체 |
| **Phase 7** | AI 인터뷰 관련 알림 | 인터뷰 완료 알림 추가 |
| **Phase 8** | 마감일 리마인더 | `EXPIRY_005~006` |
| **Phase 8.1** | 활동 리마인더 (미접속, 미완성 자소서) | `REMIND_001~003` |
| **Phase 9** | 결제 관련 알림 (결제 완료, 갱신 안내) | 결제 알림 추가 |
| **Phase 10+** | 푸시 알림 (Web Push), Quiet Hours 커스텀 | 전체 채널 확장 |

---

## 10. 환경변수

| 변수명 | 설명 | 필요 Phase |
|--------|------|-----------|
| `SENDGRID_API_KEY` | SendGrid API 키 | Phase 6.2 |
| `SENDGRID_FROM_EMAIL` | 발신 이메일 주소 | Phase 6.2 |
| `SENDGRID_FROM_NAME` | 발신자 표시명 ("Colight") | Phase 6.2 |
