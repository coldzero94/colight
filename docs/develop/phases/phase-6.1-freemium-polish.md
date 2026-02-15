# Phase 6.1: 프리미엄 & 마무리

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(API Routes, Supabase 직접 쿼리, RLS 정책)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, hooks)는 그대로 참고하세요.
>
> - `createClient` from `@/lib/supabase/server` → Go 백엔드 API 호출 (생성된 SDK 사용)
> - `src/app/api/usage/route.ts` → Go 백엔드 `internal/controller/usage_controller.go`
> - Supabase RLS 정책 → Go 미들웨어 JWT 검증 + 서비스 레이어 권한 체크
> - Supabase 직접 쿼리 → Ent ORM 쿼리 (`internal/service/`)
>
> **⚠️ 마이그레이션 참고**: `usage_tracking` 테이블은 Ent 스키마로 정의하고 Atlas로 마이그레이션합니다 (Supabase SQL 마이그레이션이 아닌).

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 프리미엄 사용량 제한 시스템을 구축하고, 전체 앱의 에러 처리/빈 상태/반응형/성능을 점검하여 베타 출시 품질을 확보한다 |
| **선행 조건** | Phase 6 (첨삭 코칭) 완료, 핵심 코칭 파이프라인(문항 분석 → 초안 → 에디터 → 첨삭) 전체 동작 |
| **스프린트** | Sprint 5 |
| **관련 기능** | Freemium 모델 (04-business-roadmap.md), 전체 UX 품질 |
| **예상 공수** | 2일 (Day 3-4) |
| **산출물** | 사용량 추적, 페이월 UI, 에러 처리 강화, 빈 상태 UI, 반응형, 성능 최적화 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 6.1.1 | 사용량 추적 | ⬜ 대기 |
| 6.1.2 | 페이월 UI | ⬜ 대기 |
| 6.1.3 | 에러 처리 점검 | ⬜ 대기 |
| 6.1.4 | 빈 상태 점검 | ⬜ 대기 |
| 6.1.5 | 모바일 반응형 | ⬜ 대기 |
| 6.1.6 | 성능 최적화 | ⬜ 대기 |

---

## Step 6.1.1: 사용량 추적

### 목표

무료 사용자의 기능 사용량을 추적하고, 제한에 도달하면 기능을 차단하는 시스템을 구현한다. 비즈니스 로드맵의 Freemium 전략에 맞춰 무료 제한을 설정한다.

### 무료 사용량 제한

| 기능 | 무료 제한 | 유료 제한 | 추적 단위 |
|------|----------|----------|----------|
| 경험 등록 | 3개 (총) | 무제한 | 총 누적 |
| 기업 분석 | 1회/일 | 무제한 | 일별 |
| 문항 분석 | 1회/일 | 무제한 | 일별 |
| 초안 코칭 | 1회/일 | 무제한 | 일별 |
| 첨삭 코칭 | 1회/일 (무료 한도 내) | 5회/자소서 | 일별 |

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `TestUsageService_CheckLimit_FreeUser` | `internal/service/usage_service_test.go` | 무료 사용자 경험 4번째 등록 시 제한 초과 반환 |
| `TestUsageService_CheckLimit_PaidUser` | `internal/service/usage_service_test.go` | 유료 사용자 제한 없이 허용 |
| `TestUsageService_DailyReset` | `internal/service/usage_service_test.go` | 일별 제한이 자정 이후 리셋 확인 |
| `TestUsageController_GetUsage` | `internal/controller/usage_controller_test.go` | 사용량 조회 API 정상 응답, 본인 사용량만 반환 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `internal/service/usage_service_test.go` 작성
  - [x] `internal/controller/usage_controller_test.go` 작성
- [x] 구현 (GREEN)
  - [x] `user_profiles` 테이블에 `plan` 컬럼 추가 (`free` / `starter` / `pro` / `season`)
  - [x] `usage_tracking` 테이블 생성 (또는 `user_profiles`에 JSONB 컬럼)
  - [x] 사용량 추적 미들웨어 구현 (`src/lib/usage/tracker.ts`)
  - [x] 각 API 엔드포인트에 사용량 체크 로직 추가
  - [x] 사용량 초과 시 403 + 업그레이드 유도 메시지 반환
  - [x] 일별 사용량 자정 리셋 로직 (Supabase Edge Function 또는 앱 내 체크)
  - [x] 사용량 현황 조회 API
- [x] 테스트 통과 확인

### DB 스키마 변경

```sql
-- user_profiles에 plan 컬럼 추가
ALTER TABLE public.user_profiles
  ADD COLUMN plan text DEFAULT 'free'
    CHECK (plan IN ('free', 'starter', 'pro', 'season'));

-- 사용량 추적 테이블
CREATE TABLE public.usage_tracking (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  feature text NOT NULL,      -- 'experience', 'analysis', 'question_analysis', 'draft', 'review'
  used_at timestamptz DEFAULT now(),
  metadata jsonb DEFAULT '{}'
);

CREATE INDEX idx_usage_user_feature_date
  ON public.usage_tracking(user_id, feature, used_at);

-- RLS
ALTER TABLE public.usage_tracking ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can view own usage"
  ON public.usage_tracking FOR SELECT
  USING (auth.uid() = user_id);
```

### 사용량 추적 유틸리티

```typescript
// src/lib/usage/tracker.ts
import { createClient } from '@/lib/supabase/server';

const FREE_LIMITS = {
  experience: { type: 'total', limit: 3 },
  analysis: { type: 'daily', limit: 1 },
  question_analysis: { type: 'daily', limit: 1 },
  draft: { type: 'daily', limit: 1 },
  review: { type: 'daily', limit: 1 },
} as const;

type Feature = keyof typeof FREE_LIMITS;

export async function checkUsageLimit(
  userId: string,
  feature: Feature
): Promise<{ allowed: boolean; used: number; limit: number; remaining: number }> {
  const supabase = await createClient();

  // 사용자 플랜 확인
  const { data: profile } = await supabase
    .from('user_profiles')
    .select('plan')
    .eq('id', userId)
    .single();

  // 유료 사용자는 무제한
  if (profile?.plan !== 'free') {
    return { allowed: true, used: 0, limit: Infinity, remaining: Infinity };
  }

  const config = FREE_LIMITS[feature];
  let query = supabase
    .from('usage_tracking')
    .select('id', { count: 'exact' })
    .eq('user_id', userId)
    .eq('feature', feature);

  // 일별 제한의 경우 오늘 날짜만 필터
  if (config.type === 'daily') {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    query = query.gte('used_at', today.toISOString());
  }

  const { count } = await query;
  const used = count || 0;
  const remaining = Math.max(0, config.limit - used);

  return {
    allowed: used < config.limit,
    used,
    limit: config.limit,
    remaining,
  };
}

export async function trackUsage(userId: string, feature: Feature, metadata?: object) {
  const supabase = await createClient();
  await supabase.from('usage_tracking').insert({
    user_id: userId,
    feature,
    metadata: metadata || {},
  });
}
```

### API 엔드포인트에 적용 예시

```typescript
// 각 API route에서 사용량 체크
const usage = await checkUsageLimit(user.id, 'draft');
if (!usage.allowed) {
  return Response.json({
    error: 'USAGE_LIMIT_EXCEEDED',
    message: '오늘의 무료 사용 횟수를 초과했습니다',
    used: usage.used,
    limit: usage.limit,
    upgrade_url: '/pricing',
  }, { status: 403 });
}

// 정상 처리 후 사용량 기록
await trackUsage(user.id, 'draft', { cover_letter_id: '...' });
```

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `GET` | `/api/usage` | - | `{ plan, features: { [key]: { used, limit, remaining } } }` |

### 산출물

- `supabase/migrations/YYYYMMDD_add_usage_tracking.sql`
- `src/lib/usage/tracker.ts`
- `src/lib/usage/limits.ts` (제한 설정 상수)
- `src/app/api/usage/route.ts`
- 기존 API routes에 사용량 체크 로직 추가

---

## Step 6.1.2: 페이월 UI

### 목표

사용량 제한에 도달한 사용자에게 업그레이드를 유도하는 페이월 모달과 블러 처리된 프리미엄 미리보기를 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('PaywallModal')` | `src/components/paywall/__tests__/paywall-modal.test.tsx` | 사용량 초과 시 모달 표시, 현재 사용량/제한 렌더링, "나중에 할게요" 클릭 시 닫힘 |
| `describe('UsageBadge')` | `src/components/paywall/__tests__/usage-badge.test.tsx` | 사용량 배지 렌더링, 남은 횟수 표시 |
| `describe('PricingCard')` | `src/components/paywall/__tests__/pricing-card.test.tsx` | 4개 플랜 카드 렌더링, CTA 버튼 동작 |
| `describe('BlurredPreview')` | `src/components/paywall/__tests__/blurred-preview.test.tsx` | 블러 처리 래퍼 렌더링, 업그레이드 링크 표시 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/paywall/__tests__/paywall-modal.test.tsx` 작성
  - [x] `src/components/paywall/__tests__/usage-badge.test.tsx` 작성
  - [x] `src/components/paywall/__tests__/pricing-card.test.tsx` 작성
  - [x] `src/components/paywall/__tests__/blurred-preview.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 페이월 모달 컴포넌트 구현 (`PaywallModal`)
  - [x] 블러 처리된 프리미엄 미리보기 (분석 결과 일부 흐릿하게)
  - [x] 업그레이드 CTA 버튼 (가격표 페이지 또는 결제 모달로 이동)
  - [x] 사용량 현황 표시 ("오늘 1/1 사용 완료")
  - [x] 토스트 기반 알림 (사용량 80% 도달 시 경고)
  - [x] 가격표 페이지 구현 (`/pricing`)
  - [x] 결제 기능은 Phase 9에서 구현, 현재는 CTA만
- [x] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `PaywallModal` | `src/components/paywall/paywall-modal.tsx` | `feature: string, usage: Usage, onClose: fn` | 페이월 모달 |
| `BlurredPreview` | `src/components/paywall/blurred-preview.tsx` | `children: ReactNode` | 프리미엄 콘텐츠 블러 처리 래퍼 |
| `UsageBadge` | `src/components/paywall/usage-badge.tsx` | `used: number, limit: number, feature: string` | 사용량 배지 (헤더 또는 사이드바) |
| `PricingPage` | `src/app/(main)/pricing/page.tsx` | - | 가격표 페이지 |
| `PricingCard` | `src/components/paywall/pricing-card.tsx` | `plan: Plan` | 가격 카드 (무료/스타터/프로/시즌패스) |

### 페이월 모달 레이아웃

```text
┌──────────────────────────────────────────────┐
│                    ×                          │
│                                              │
│  🔒 오늘의 무료 분석을 모두 사용했어요           │
│                                              │
│  오늘 사용: 1/1회                             │
│  (매일 자정에 초기화됩니다)                     │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │  스타터 10회권 - 4,900원                │  │
│  │  • 기업 분석 10회                       │  │
│  │  • AI 코칭 10회                        │  │
│  │  [ 시작하기 ]                           │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │  프로 30회권 - 12,900원  ⭐ 인기         │  │
│  │  • 기업 분석 30회                       │  │
│  │  • AI 코칭 30회                        │  │
│  │  • 경험 무제한                          │  │
│  │  [ 시작하기 ]                           │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  나중에 할게요                                 │
└──────────────────────────────────────────────┘
```

### 블러 처리 미리보기

```typescript
// src/components/paywall/blurred-preview.tsx
'use client';

export function BlurredPreview({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <div className="blur-sm pointer-events-none select-none">
        {children}
      </div>
      <div className="absolute inset-0 flex items-center justify-center bg-white/50">
        <div className="text-center p-4">
          <p className="text-lg font-semibold">프리미엄 기능입니다</p>
          <p className="text-sm text-gray-500 mt-1">업그레이드하면 전체 결과를 확인할 수 있어요</p>
          <Button className="mt-3" asChild>
            <Link href="/pricing">업그레이드</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
```

### 산출물

- `src/components/paywall/paywall-modal.tsx`
- `src/components/paywall/blurred-preview.tsx`
- `src/components/paywall/usage-badge.tsx`
- `src/components/paywall/pricing-card.tsx`
- `src/app/(main)/pricing/page.tsx`

---

## Step 6.1.3: 에러 처리 점검

### 목표

모든 API 라우트와 페이지에 일관된 에러 처리가 적용되어 있는지 점검하고, 누락된 부분을 보완한다. 사용자에게 친화적인 에러 메시지를 표시한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('ErrorBoundary pages')` | `src/app/__tests__/error-pages.test.tsx` | 404/500 커스텀 에러 페이지 렌더링 확인 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/app/__tests__/error-pages.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 모든 API 라우트에 try-catch + 일관된 에러 응답 형식 적용
  - [x] 모든 `(main)/*/` 라우트에 `error.tsx` 존재 확인
  - [x] 모든 `(main)/*/` 라우트에 `loading.tsx` 존재 확인
  - [x] 토스트 기반 에러 알림 (sonner)
  - [x] AI API 실패 시 재시도 안내 (retry 버튼)
  - [x] 네트워크 에러 감지 + 오프라인 배너
  - [x] Supabase 연결 실패 시 메시지
  - [x] 404 페이지 커스터마이징 (`src/app/not-found.tsx`)
  - [x] 500 에러 페이지 커스터마이징 (`src/app/global-error.tsx`)
- [x] 테스트 통과 확인

### 에러 응답 형식 (API)

```typescript
// src/lib/errors.ts
export class AppError extends Error {
  constructor(
    message: string,
    public code: string,
    public status: number = 500,
    public details?: unknown
  ) {
    super(message);
  }
}

// 일관된 에러 응답
export function errorResponse(error: unknown) {
  if (error instanceof AppError) {
    return Response.json(
      { error: error.code, message: error.message, details: error.details },
      { status: error.status }
    );
  }

  console.error('Unexpected error:', error);
  return Response.json(
    { error: 'INTERNAL_ERROR', message: '서버에 문제가 발생했습니다' },
    { status: 500 }
  );
}
```

### 에러 페이지 목록

| 경로 | 파일 | 설명 |
|------|------|------|
| `src/app/not-found.tsx` | 글로벌 404 | "페이지를 찾을 수 없습니다" |
| `src/app/global-error.tsx` | 글로벌 500 | "서버에 문제가 발생했습니다" |
| `src/app/(main)/dashboard/error.tsx` | 대시보드 에러 | 대시보드 로드 실패 |
| `src/app/(main)/experiences/error.tsx` | 경험 에러 | 경험 관련 에러 |
| `src/app/(main)/analysis/error.tsx` | 분석 에러 | 기업 분석 에러 |
| `src/app/(main)/coaching/error.tsx` | 코칭 에러 | 코칭 관련 에러 |
| `src/app/(main)/coaching/[id]/edit/error.tsx` | 에디터 에러 | 에디터 로드 실패 |

### 산출물

- `src/lib/errors.ts`
- `src/app/not-found.tsx`
- `src/app/global-error.tsx`
- 각 라우트별 `error.tsx`, `loading.tsx` 누락분 보완

---

## Step 6.1.4: 빈 상태 점검

### 목표

데이터가 없는 모든 목록 페이지에 적절한 빈 상태 UI를 제공하여, 사용자가 다음 행동을 할 수 있도록 안내한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('EmptyState')` | `src/components/ui/__tests__/empty-state.test.tsx` | 빈 상태 컴포넌트 렌더링, CTA 버튼 링크 확인 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/ui/__tests__/empty-state.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 경험 목록 (`/experiences`): "아직 등록한 경험이 없어요" + 경험 등록 CTA
  - [x] 기업 분석 목록 (`/analysis`): "아직 분석한 기업이 없어요" + URL 입력 CTA
  - [x] 코칭 메인 (`/coaching`): "아직 코칭 이력이 없어요" + 코칭 시작 CTA
  - [x] 대시보드 (`/dashboard`): 첫 사용자 온보딩 가이드
  - [x] 경험 추천 (코칭 내): "매칭되는 경험이 없습니다" + 경험 등록 유도
  - [x] 첨삭 이력: "아직 첨삭 이력이 없어요"
- [x] 테스트 통과 확인

### 빈 상태 UI 패턴

```typescript
// src/components/ui/empty-state.tsx
interface EmptyStateProps {
  icon: React.ReactNode;      // 일러스트 또는 아이콘
  title: string;              // "아직 등록한 경험이 없어요"
  description: string;        // "경험을 등록하면 AI가 자동으로 역량을 분류해드려요"
  action?: {
    label: string;            // "경험 등록하기"
    href: string;             // "/experiences/new"
  };
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="text-gray-300 mb-4">{icon}</div>
      <h3 className="text-lg font-medium text-gray-900">{title}</h3>
      <p className="text-sm text-gray-500 mt-1 max-w-sm">{description}</p>
      {action && (
        <Button asChild className="mt-4">
          <Link href={action.href}>{action.label}</Link>
        </Button>
      )}
    </div>
  );
}
```

### 산출물

- `src/components/ui/empty-state.tsx`
- 각 페이지에 빈 상태 적용

---

## Step 6.1.5: 모바일 반응형

### 목표

모든 페이지를 375px (iPhone SE), 768px (iPad) 기준으로 반응형 테스트하고, 주요 레이아웃 이슈를 수정한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('Sidebar responsive')` | `src/components/layout/__tests__/sidebar.test.tsx` | 모바일에서 햄버거 메뉴 표시, 데스크탑에서 사이드바 표시 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/layout/__tests__/sidebar.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 사이드바 → 햄버거 메뉴 (768px 이하)
  - [x] 테이블 → 카드 레이아웃 (768px 이하)
  - [x] 에디터 + 사이드 패널 → 단일 컬럼 + 토글/바텀시트 (768px 이하)
  - [x] 레이더 차트 크기 조정 (모바일에서 가독성)
  - [x] 폼 입력 영역 모바일 최적화 (터치 타겟 48px)
  - [x] 모달 크기 모바일 대응 (전체 화면 또는 바텀시트)
  - [x] 글자 크기 가독성 확인 (최소 14px)
  - [x] 가로 스크롤 발생하지 않도록 확인
- [x] 테스트 통과 확인

### 반응형 브레이크포인트

| 브레이크포인트 | Tailwind | 디바이스 | 레이아웃 변경 |
|--------------|----------|---------|-------------|
| < 640px | `sm:` 미만 | 모바일 (375px) | 단일 컬럼, 햄버거, 카드 |
| 640~767px | `sm:` ~ `md:` | 태블릿 세로 | 일부 2컬럼 |
| 768~1023px | `md:` ~ `lg:` | 태블릿 가로 | 사이드바 + 메인 |
| ≥ 1024px | `lg:` | 데스크탑 | 전체 레이아웃 |

### 주요 수정 대상

```text
1. Main Layout (사이드바)
   Desktop: 사이드바(240px) + 콘텐츠
   Mobile: 햄버거 → Drawer 사이드바

2. 분석 결과 페이지
   Desktop: 결과 카드 2~3컬럼 그리드
   Mobile: 단일 컬럼 스택

3. 에디터 페이지
   Desktop: 에디터(70%) + 사이드패널(30%)
   Mobile: 에디터 전체폭 + 사이드패널은 바텀시트(Sheet)

4. 첨삭 결과 (레이더 차트)
   Desktop: 차트(50%) + 점수(50%) 나란히
   Mobile: 차트 → 점수 수직 스택

5. 경험 목록
   Desktop: 카드 그리드 (3컬럼)
   Tablet: 카드 그리드 (2컬럼)
   Mobile: 카드 리스트 (1컬럼)
```

### 사이드바 → 햄버거 전환

```typescript
// src/components/layout/sidebar.tsx
'use client';

import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Menu } from 'lucide-react';

export function Sidebar() {
  return (
    <>
      {/* 데스크탑 사이드바 */}
      <aside className="hidden md:flex w-60 flex-col border-r">
        <SidebarContent />
      </aside>

      {/* 모바일 햄버거 */}
      <div className="md:hidden">
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon">
              <Menu className="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-60 p-0">
            <SidebarContent />
          </SheetContent>
        </Sheet>
      </div>
    </>
  );
}
```

### 산출물

- `src/components/layout/sidebar.tsx` 수정 (반응형)
- 각 페이지 반응형 스타일 수정
- 테이블 → 카드 전환 컴포넌트

---

## Step 6.1.6: 성능 최적화

### 목표

Lighthouse Performance 점수 70점 이상을 달성하고, 주요 번들 크기를 최적화한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('Lazy loaded components')` | `src/components/__tests__/lazy-load.test.tsx` | Tiptap/Recharts 컴포넌트 lazy load 시 로딩 스켈레톤 표시 확인 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/__tests__/lazy-load.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] Lighthouse 측정 (현재 점수 기록)
  - [x] Tiptap 에디터 lazy load (`next/dynamic`, ssr: false)
  - [x] Recharts 레이더 차트 lazy load (`next/dynamic`, ssr: false)
  - [x] 이미지 최적화 (`next/image` 사용, WebP 포맷)
  - [x] 번들 분석 (`@next/bundle-analyzer`)
  - [x] 불필요한 클라이언트 컴포넌트 서버 컴포넌트로 전환
  - [x] React Query 캐싱 설정 최적화 (staleTime, gcTime)
  - [x] Suspense 경계 적용 (병렬 데이터 로딩)
  - [x] 폰트 최적화 (`next/font`)
- [x] 테스트 통과 확인

### Lazy Load 대상

```typescript
// 무거운 컴포넌트 lazy load
import dynamic from 'next/dynamic';

// Tiptap 에디터 (~150KB)
const TiptapEditor = dynamic(
  () => import('@/components/coaching/tiptap-editor').then(m => m.TiptapEditor),
  { ssr: false, loading: () => <EditorSkeleton /> }
);

// Recharts 레이더 차트 (~120KB)
const ScoreRadarChart = dynamic(
  () => import('@/components/coaching/score-radar-chart').then(m => m.ScoreRadarChart),
  { ssr: false, loading: () => <ChartSkeleton /> }
);
```

### 번들 분석

```bash
# package.json에 추가
npm install --save-dev @next/bundle-analyzer

# next.config.ts에 추가
const withBundleAnalyzer = require('@next/bundle-analyzer')({
  enabled: process.env.ANALYZE === 'true',
});

# 분석 실행
ANALYZE=true npm run build
```

### 성능 목표

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| Lighthouse Performance | ≥ 70 | Chrome DevTools Lighthouse |
| First Contentful Paint | ≤ 2.0s | Lighthouse |
| Largest Contentful Paint | ≤ 3.0s | Lighthouse |
| Time to Interactive | ≤ 4.0s | Lighthouse |
| 메인 번들 크기 | ≤ 200KB (gzip) | Bundle Analyzer |

### React Query 캐싱 설정

```typescript
// src/lib/query-client.ts
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,     // 5분간 fresh
      gcTime: 30 * 60 * 1000,        // 30분 후 GC
      refetchOnWindowFocus: false,    // 포커스 시 재요청 X
      retry: 1,                      // 1회 재시도
    },
  },
});
```

### 산출물

- 각 페이지 lazy load 적용
- `next.config.ts` 번들 분석 설정
- React Query 캐싱 최적화
- Lighthouse 리포트 스크린샷

---

## Phase 완료 체크리스트

- [x] 무료 사용자: 경험 3개 / 분석 1회/일 / 코칭 1회/일 제한 동작
- [x] 사용량 초과 시 페이월 모달 표시 + 업그레이드 CTA
- [x] 블러 처리된 프리미엄 미리보기 동작
- [x] 모든 API 라우트에 일관된 에러 처리 적용
- [x] 모든 페이지에 `loading.tsx` + `error.tsx` 존재
- [x] 모든 목록 페이지에 빈 상태 UI 적용
- [x] 404/500 커스텀 에러 페이지 동작
- [x] 모바일 375px: 사이드바→햄버거, 테이블→카드, 레이아웃 정상
- [x] 태블릿 768px: 레이아웃 정상
- [x] Lighthouse Performance ≥ 70
- [x] Tiptap/Recharts lazy load 적용
- [x] 토스트 에러 알림 동작
- [x] `moon run backend:test` → 전체 통과
- [x] `moon run web:test` → 전체 통과
- [x] `moon run :lint` → 경고 0건
- [x] `moon run web:build` → 빌드 성공

---

## 다음 Phase

**[Phase 6.2: 랜딩 & 베타](./phase-6.2-landing-beta.md)** — 랜딩 페이지, 법적 페이지, Vercel 프로덕션 배포, 베타 사용자 모집
