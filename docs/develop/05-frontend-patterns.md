# 프론트엔드 패턴

> 작성일: 2026-02-11

---

## 1. 라우트 구조

```
src/app/
├── layout.tsx                          # Root Layout
├── page.tsx                            # 랜딩 페이지 (/)
│
├── (auth)/                             # 인증 그룹 (비로그인 전용)
│   ├── layout.tsx                      # Auth 레이아웃 (심플, 로고만)
│   ├── login/page.tsx                  # /login
│   └── signup/page.tsx                 # /signup
│
├── (main)/                             # 메인 그룹 (로그인 필수)
│   ├── layout.tsx                      # Main 레이아웃 (사이드바 + 헤더)
│   ├── dashboard/
│   │   ├── page.tsx                    # /dashboard (메인 대시보드)
│   │   └── loading.tsx
│   ├── experiences/
│   │   ├── page.tsx                    # /experiences (경험 목록)
│   │   ├── new/page.tsx               # /experiences/new (경험 등록)
│   │   ├── [id]/page.tsx              # /experiences/:id (경험 상세)
│   │   ├── interview/page.tsx         # /experiences/interview (AI 인터뷰)
│   │   ├── loading.tsx
│   │   └── error.tsx
│   ├── analysis/
│   │   ├── page.tsx                    # /analysis (기업 분석)
│   │   ├── [id]/page.tsx              # /analysis/:id (분석 결과)
│   │   ├── loading.tsx
│   │   └── error.tsx
│   ├── coaching/
│   │   ├── page.tsx                    # /coaching (코칭 메인)
│   │   ├── [id]/page.tsx              # /coaching/:id (코칭 세션)
│   │   ├── loading.tsx
│   │   └── error.tsx
│   └── settings/page.tsx              # /settings (설정)
│
└── (Note: API endpoints are now in Go backend at /v1/*)
```

### 라우트 그룹 규칙

| 그룹 | 접근 조건 | 레이아웃 특성 |
|------|----------|-------------|
| `(auth)` | 비로그인 사용자만 (로그인 시 dashboard 리다이렉트) | 심플 레이아웃, 로고 + 폼만 |
| `(main)` | 로그인 필수 (미로그인 시 login 리다이렉트) | 사이드바 네비게이션 + 헤더 |

---

## 2. 상태 관리 규칙

### 원칙: 상태 유형별 도구 분리

```
┌────────────────────────────────────────────────────────────┐
│                    상태 관리 전략                            │
│                                                            │
│  서버 상태 (Server State)          → React Query           │
│  ├── DB 데이터 (경험, 분석, 코칭)                           │
│  ├── API 응답 캐싱                                         │
│  └── 낙관적 업데이트                                        │
│                                                            │
│  클라이언트 상태 (Client State)     → Zustand              │
│  ├── UI 상태 (사이드바 열림/닫힘)                            │
│  ├── 모달/토스트 상태                                       │
│  └── 임시 필터/정렬 설정                                    │
│                                                            │
│  폼 상태 (Form State)              → React Hook Form + Zod│
│  ├── 경험 등록 폼                                          │
│  ├── 회원가입/로그인 폼                                     │
│  └── 자소서 에디터 메타데이터                                │
│                                                            │
│  URL 상태 (URL State)              → searchParams          │
│  ├── 필터, 정렬, 페이지네이션                                │
│  └── 탭, 뷰 모드                                           │
└────────────────────────────────────────────────────────────┘
```

---

## 3. React Query 키 팩토리 패턴

```typescript
// src/hooks/query-keys.ts

export const queryKeys = {
  // ── 경험 ──
  experiences: {
    all: ['experiences'] as const,
    lists: () => [...queryKeys.experiences.all, 'list'] as const,
    list: (filters: ExperienceFilters) =>
      [...queryKeys.experiences.lists(), filters] as const,
    details: () => [...queryKeys.experiences.all, 'detail'] as const,
    detail: (id: string) =>
      [...queryKeys.experiences.details(), id] as const,
    weapons: (id: string) =>
      [...queryKeys.experiences.detail(id), 'weapons'] as const,
  },

  // ── 기업 분석 ──
  analyses: {
    all: ['analyses'] as const,
    lists: () => [...queryKeys.analyses.all, 'list'] as const,
    detail: (id: string) =>
      [...queryKeys.analyses.all, 'detail', id] as const,
    companyData: (companyName: string) =>
      [...queryKeys.analyses.all, 'company', companyName] as const,
  },

  // ── 코칭 ──
  coaching: {
    all: ['coaching'] as const,
    sessions: (coverLetterId: string) =>
      [...queryKeys.coaching.all, 'sessions', coverLetterId] as const,
    questionAnalysis: (questionId: string) =>
      [...queryKeys.coaching.all, 'question', questionId] as const,
  },

  // ── 지원 ──
  applications: {
    all: ['applications'] as const,
    lists: () => [...queryKeys.applications.all, 'list'] as const,
    list: (filters: ApplicationFilters) =>
      [...queryKeys.applications.lists(), filters] as const,
    detail: (id: string) =>
      [...queryKeys.applications.all, 'detail', id] as const,
  },

  // ── 마스터 데이터 ──
  weapons: {
    all: ['weapons'] as const,
    categories: () => [...queryKeys.weapons.all, 'categories'] as const,
  },
} as const;
```

### 사용 예시

```typescript
// src/hooks/use-experiences.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from './query-keys';

// 목록 조회
export function useExperiences(filters: ExperienceFilters) {
  return useQuery({
    queryKey: queryKeys.experiences.list(filters),
    queryFn: () => fetchExperiences(filters),
  });
}

// 상세 조회
export function useExperience(id: string) {
  return useQuery({
    queryKey: queryKeys.experiences.detail(id),
    queryFn: () => fetchExperience(id),
    enabled: !!id,
  });
}

// 등록 (낙관적 업데이트)
export function useCreateExperience() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createExperience,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.experiences.lists(),
      });
    },
  });
}
```

---

## 4. 컴포넌트 라이브러리

| 라이브러리 | 용도 | 사용처 |
|-----------|------|--------|
| **shadcn/ui** | 기본 UI 컴포넌트 (Button, Card, Dialog, Input, Select, Tabs, Toast 등) | 전체 앱 |
| **Tiptap** | 리치 텍스트 에디터 | 자소서 작성/편집 (`coaching/[id]`) |
| **@hello-pangea/dnd** | 드래그 앤 드롭 칸반보드 | 지원 현황 대시보드 (`dashboard`) |
| **Recharts** | 차트/그래프 | 무기 레이더 차트, 통계 (`dashboard`) |
| **Tailwind CSS** | 유틸리티 CSS | 전체 스타일링 |

### shadcn/ui 컴포넌트 목록

```
src/components/ui/
├── button.tsx
├── card.tsx
├── dialog.tsx
├── dropdown-menu.tsx
├── input.tsx
├── label.tsx
├── select.tsx
├── skeleton.tsx
├── tabs.tsx
├── textarea.tsx
├── toast.tsx
├── toaster.tsx
├── badge.tsx
├── progress.tsx
├── separator.tsx
├── sheet.tsx          # 모바일 사이드바용
├── tooltip.tsx
└── avatar.tsx
```

---

## 5. Server vs Client Component 분리 규칙

### Server Component (기본값, `'use client'` 없음)

```
사용 조건:
- 데이터 fetch가 필요한 페이지 컴포넌트
- SEO가 중요한 콘텐츠
- Go 백엔드 API 호출 (데이터 조회)
- 민감한 정보 접근 (환경변수 등)
- 무거운 의존성이 브라우저에 불필요한 경우

예시:
- page.tsx (모든 페이지 진입점)
- layout.tsx (레이아웃)
- 데이터 표시 전용 컴포넌트
```

### Client Component (`'use client'` 명시)

```
사용 조건:
- 사용자 인터랙션 (onClick, onChange 등)
- useState, useEffect 등 React Hook 사용
- 브라우저 API 접근 (localStorage, window 등)
- React Query, Zustand 등 클라이언트 상태
- 실시간 업데이트 (스트리밍, WebSocket)
- 서드파티 라이브러리 (Tiptap, Recharts, DnD)

예시:
- 폼 컴포넌트 (ExperienceForm)
- 인터랙티브 UI (칸반보드, 에디터)
- 스트리밍 AI 응답 표시
- 네비게이션 (active state 필요)
```

### 분리 패턴: Container/Presentational

> **규칙**: Server Component에서 Supabase를 직접 쿼리하지 않습니다. 모든 데이터는 Go 백엔드 API를 통해 조회합니다.

```typescript
// page.tsx (Server Component) — Go API 호출 (SDK 사용)
import { getExperiences } from '@/api/generated';

export default async function ExperiencesPage() {
  const { data: experiences } = await getExperiences();
  return <ExperienceList initialData={experiences} />;
}

// ExperienceList.tsx (Client Component) — 인터랙션
'use client';

export function ExperienceList({ initialData }: Props) {
  const { data: experiences } = useExperiences({
    initialData, // SSR 초기 데이터 활용
  });

  return (
    <div>
      {experiences?.map((exp) => (
        <ExperienceCard key={exp.id} experience={exp} />
      ))}
    </div>
  );
}
```

---

## 6. 로딩 상태: loading.tsx 스켈레톤 패턴

```typescript
// src/app/(main)/experiences/loading.tsx
import { Skeleton } from '@/components/ui/skeleton';
import { Card } from '@/components/ui/card';

export default function ExperiencesLoading() {
  return (
    <div className="space-y-4">
      {/* 헤더 스켈레톤 */}
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-10 w-32" />
      </div>

      {/* 필터 바 스켈레톤 */}
      <div className="flex gap-2">
        <Skeleton className="h-10 w-24" />
        <Skeleton className="h-10 w-24" />
        <Skeleton className="h-10 w-24" />
      </div>

      {/* 경험 카드 스켈레톤 (3개) */}
      {Array.from({ length: 3 }).map((_, i) => (
        <Card key={i} className="p-6">
          <div className="space-y-3">
            <Skeleton className="h-6 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
            <div className="flex gap-2">
              <Skeleton className="h-6 w-16 rounded-full" />
              <Skeleton className="h-6 w-16 rounded-full" />
              <Skeleton className="h-6 w-16 rounded-full" />
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
          </div>
        </Card>
      ))}
    </div>
  );
}
```

### 각 페이지별 스켈레톤 구조

| 페이지 | 스켈레톤 구성 |
|--------|-------------|
| `experiences/` | 카드 리스트 (제목 + 무기 뱃지 + 내용 미리보기) |
| `experiences/[id]` | 상세 카드 (STAR 섹션 + 무기 태그 + 사용 이력) |
| `analysis/` | URL 입력 폼 + 최근 분석 목록 |
| `analysis/[id]` | 분석 리포트 (기업 정보 + 인재상 + 매칭 결과) |
| `coaching/[id]` | 에디터 + 코칭 패널 (좌우 분할) |
| `dashboard/` | 칸반 컬럼 + 통계 차트 |

---

## 7. 에러 바운더리: error.tsx 패턴

```typescript
// src/app/(main)/experiences/error.tsx
'use client';

import { useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

export default function ExperiencesError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // 에러 로깅 (Sentry 등 연동 가능)
    console.error('Experiences Error:', error);
  }, [error]);

  return (
    <div className="flex items-center justify-center min-h-[50vh]">
      <Card className="p-8 text-center max-w-md">
        <h2 className="text-xl font-semibold mb-2">
          문제가 발생했습니다
        </h2>
        <p className="text-muted-foreground mb-4">
          {error.message || '경험 데이터를 불러오는 중 오류가 발생했습니다.'}
        </p>
        <div className="flex gap-2 justify-center">
          <Button onClick={reset} variant="default">
            다시 시도
          </Button>
          <Button onClick={() => window.location.href = '/dashboard'} variant="outline">
            대시보드로 이동
          </Button>
        </div>
      </Card>
    </div>
  );
}
```

### 에러 유형별 처리

| 에러 | 표시 | 액션 |
|------|------|------|
| 네트워크 오류 | "서버에 연결할 수 없습니다" | 다시 시도 버튼 |
| 인증 만료 | "세션이 만료되었습니다" | 로그인 페이지 이동 |
| 권한 없음 | "접근 권한이 없습니다" | 대시보드 이동 |
| 리소스 없음 | "페이지를 찾을 수 없습니다" | 목록 페이지 이동 |
| AI API 오류 | "AI 서비스에 일시적인 문제가 있습니다" | 다시 시도 버튼 |
| 사용량 초과 | "이번 달 무료 사용 한도를 초과했습니다" | 업그레이드 안내 |

---

## 8. 스트리밍 UI 패턴

AI 코칭 응답을 실시간으로 표시하는 패턴입니다.

```typescript
// src/components/coaching/streaming-response.tsx
'use client';

import { useChat } from '@ai-sdk/react';

interface StreamingResponseProps {
  coverLetterId: string;
  experienceIds: string[];
  questionAnalysis: QuestionAnalysis;
}

export function StreamingResponse({
  coverLetterId,
  experienceIds,
  questionAnalysis,
}: StreamingResponseProps) {
  const { messages, isLoading, error, reload } = useChat({
    api: '/api/coaching/draft',
    body: { coverLetterId, experienceIds, questionAnalysis },
  });

  const lastMessage = messages[messages.length - 1];

  return (
    <div className="space-y-4">
      {/* 로딩 인디케이터 */}
      {isLoading && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <div className="animate-pulse h-2 w-2 rounded-full bg-primary" />
          AI 코치가 분석 중입니다...
        </div>
      )}

      {/* 스트리밍 텍스트 */}
      {lastMessage?.role === 'assistant' && (
        <div className="prose prose-sm max-w-none">
          {lastMessage.content}
          {isLoading && (
            <span className="inline-block w-1 h-4 bg-primary animate-pulse ml-0.5" />
          )}
        </div>
      )}

      {/* 에러 처리 */}
      {error && (
        <div className="text-sm text-destructive">
          오류가 발생했습니다.
          <button onClick={reload} className="underline ml-1">
            다시 시도
          </button>
        </div>
      )}
    </div>
  );
}
```

### Tiptap 에디터와 스트리밍 연동

```typescript
// 코칭 결과를 에디터에 삽입
function insertCoachingResult(editor: Editor, content: string) {
  editor
    .chain()
    .focus()
    .insertContent(content)
    .run();
}
```

---

## 9. 반응형 브레이크포인트 (Mobile-First)

### Tailwind CSS 브레이크포인트

| 브레이크포인트 | 크기 | 대상 디바이스 | 레이아웃 |
|--------------|------|-------------|---------|
| 기본 (모바일) | < 640px | 스마트폰 | 단일 컬럼, 하단 네비게이션 |
| `sm` | >= 640px | 대형 스마트폰 | 단일 컬럼, 여백 증가 |
| `md` | >= 768px | 태블릿 | 2컬럼 가능, 사이드바 축소 |
| `lg` | >= 1024px | 노트북 | 사이드바 + 메인 콘텐츠 |
| `xl` | >= 1280px | 데스크탑 | 사이드바 + 메인 + 사이드 패널 |

### 주요 레이아웃 반응형 처리

```typescript
// Main Layout — 사이드바 반응형
export function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      {/* 데스크탑: 고정 사이드바 */}
      <aside className="hidden lg:flex lg:w-64 lg:flex-col border-r">
        <Sidebar />
      </aside>

      {/* 모바일: Sheet (슬라이드 메뉴) */}
      <Sheet>
        <SheetTrigger asChild className="lg:hidden">
          <Button variant="ghost" size="icon">
            <MenuIcon />
          </Button>
        </SheetTrigger>
        <SheetContent side="left">
          <Sidebar />
        </SheetContent>
      </Sheet>

      <main className="flex-1 p-4 md:p-6 lg:p-8">
        {children}
      </main>
    </div>
  );
}
```

```typescript
// 코칭 페이지 — 에디터/코칭 패널 반응형
export function CoachingLayout() {
  return (
    <div className="flex flex-col lg:flex-row gap-4 h-[calc(100vh-4rem)]">
      {/* 에디터 (모바일: 전체 폭, 데스크탑: 좌측 60%) */}
      <div className="w-full lg:w-3/5 min-h-[50vh] lg:min-h-0">
        <TiptapEditor />
      </div>

      {/* 코칭 패널 (모바일: 전체 폭, 데스크탑: 우측 40%) */}
      <div className="w-full lg:w-2/5 border-t lg:border-t-0 lg:border-l">
        <CoachingPanel />
      </div>
    </div>
  );
}
```

```typescript
// 칸반보드 — 수평 스크롤 (모바일)
export function KanbanBoard() {
  return (
    <div className="flex gap-4 overflow-x-auto pb-4 snap-x snap-mandatory
                    md:snap-none md:overflow-x-visible">
      {columns.map((col) => (
        <div
          key={col.id}
          className="min-w-[280px] w-[280px] md:w-auto md:flex-1
                     snap-center flex-shrink-0 md:flex-shrink"
        >
          <KanbanColumn column={col} />
        </div>
      ))}
    </div>
  );
}
```

### 모바일 네비게이션

```typescript
// 모바일 하단 네비게이션 바
export function BottomNav() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 bg-background border-t
                    flex justify-around py-2 lg:hidden z-50">
      <NavItem href="/experiences" icon={<FileTextIcon />} label="경험" />
      <NavItem href="/analysis" icon={<SearchIcon />} label="분석" />
      <NavItem href="/coaching" icon={<PenIcon />} label="코칭" />
      <NavItem href="/dashboard" icon={<LayoutIcon />} label="대시보드" />
    </nav>
  );
}
```

---

## 10. 접근성(Accessibility) 패턴

### 폼 접근성 규칙

모든 폼 입력 요소는 다음 규칙을 따릅니다:

| 규칙 | 설명 | 예시 |
|------|------|------|
| `htmlFor`/`id` 연결 | `<label>`과 `<input>`을 `htmlFor`/`id`로 연결 | `<label htmlFor="title">` + `<input id="title">` |
| `role="alert"` | 유효성 검증 에러 메시지에 `role="alert"` 부여 | `<p role="alert">제목을 입력해주세요</p>` |
| `maxLength` | HTML `maxLength` 속성으로 클라이언트 제한 보강 | `<input maxLength={100}>` |
| `aria-describedby` | 보조 설명(가이드 텍스트)을 입력과 연결 | `<textarea aria-describedby="guide-id">` |
| `aria-live="polite"` | 동적 카운터에 스크린리더 공지 | `<p aria-live="polite">50/1000</p>` |

### 네비게이션 접근성

```typescript
// Back link — aria-label로 명확한 액션 안내
<Link
  href="/experiences"
  aria-label="경험 목록으로 돌아가기"
>
  ← 경험 목록
</Link>

// Tablist — aria-label로 탭 그룹 설명
<div role="tablist" aria-label="무기 역량 필터">
  <button role="tab" aria-selected={isActive}>...</button>
</div>
```

### 터치 타겟

모바일 터치 타겟은 최소 `44px` 이상을 유지합니다:

```
// 필터 탭 버튼 — px-4 py-2 (높이 약 40px+)
className="px-4 py-2 text-sm font-medium"

// 폼 버튼 — px-6 py-2
className="px-6 py-2 text-sm font-medium"
```

### 모달/다이얼로그

```typescript
// DeleteDialog — 접근성 속성 필수
<div role="dialog" aria-modal="true" aria-label="삭제 확인">
  ...
</div>
```

---

## 11. 성능 최적화 (Vercel Best Practices)

Phase 2.1 완료 시 적용된 최적화:

| 규칙 | 적용 | 파일 |
|------|------|------|
| `bundle-defer-third-party` | ReactQueryDevtools를 `next/dynamic`으로 lazy-load | `providers.tsx` |
| `bundle-barrel-imports` | `optimizePackageImports` 설정 | `next.config.ts` |
| Navigation prefetch | 네비게이션에 `router.push` 대신 `<Link>` 사용 | 경험 상세/등록/수정 페이지 |

### optimizePackageImports 설정

```typescript
// next.config.ts
experimental: {
  optimizePackageImports: [
    "@tanstack/react-query",
    "sonner",
    "react-hook-form",
    "@hookform/resolvers",
  ],
},
```
