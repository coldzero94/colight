# Phase 1.6: 공통 레이아웃 (사이드바/헤더)

## 목표

인증된 사용자를 위한 메인 레이아웃을 구현한다. 데스크톱에서는 사이드바, 모바일에서는 햄버거 메뉴를 제공한다. 공통 컴포넌트(로딩, 빈 상태, 에러)도 함께 구현한다.

---

## 디렉토리 구조

```text
apps/web/src/
├── app/
│   └── (main)/
│       ├── layout.tsx                # 메인 레이아웃 (AuthGuard + Sidebar + Header)
│       ├── experiences/
│       │   ├── page.tsx              # 경험 목록 (빈 상태)
│       │   ├── loading.tsx           # 로딩 UI
│       │   └── error.tsx             # 에러 UI
│       ├── analysis/
│       │   ├── page.tsx              # 기업 분석 (빈 상태)
│       │   ├── loading.tsx
│       │   └── error.tsx
│       ├── coaching/
│       │   ├── page.tsx              # 코칭 (빈 상태)
│       │   ├── loading.tsx
│       │   └── error.tsx
│       └── dashboard/
│           ├── page.tsx              # 대시보드 (빈 상태)
│           ├── loading.tsx
│           └── error.tsx
├── components/
│   ├── layout/
│   │   ├── sidebar.tsx               # 데스크톱 사이드바
│   │   ├── header.tsx                # 헤더 (모바일 햄버거 + 페이지 제목)
│   │   ├── mobile-nav.tsx            # 모바일 Sheet 네비게이션
│   │   └── user-dropdown.tsx         # 사용자 드롭다운 메뉴
│   └── common/
│       ├── loading-spinner.tsx       # 재사용 로딩 스피너
│       └── empty-state.tsx           # 재사용 빈 상태
└── app/
    └── not-found.tsx                 # 404 페이지
```

---

## A. 메인 레이아웃

### `src/app/(main)/layout.tsx`

```
데스크톱 (lg+):
┌──────────┬───────────────────────────────────┐
│          │  Header (페이지 제목 + 유저 메뉴)    │
│ Sidebar  ├───────────────────────────────────┤
│ (w-64)   │                                   │
│          │  Main Content                     │
│          │                                   │
│          │                                   │
└──────────┴───────────────────────────────────┘

모바일 (<lg):
┌──────────────────────────────────────────────┐
│ ☰ 페이지 제목                          유저 ▾  │
├──────────────────────────────────────────────┤
│                                              │
│  Main Content                                │
│                                              │
└──────────────────────────────────────────────┘
```

```typescript
import { AuthGuard } from "@/components/auth/auth-guard";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <div className="flex h-screen">
        {/* Desktop Sidebar */}
        <div className="hidden lg:block">
          <Sidebar />
        </div>

        <div className="flex flex-1 flex-col overflow-hidden">
          <Header />
          <main className="flex-1 overflow-auto p-6">
            {children}
          </main>
        </div>
      </div>
    </AuthGuard>
  );
}
```

---

## B. 사이드바

### `src/components/layout/sidebar.tsx`

```
┌──────────────────┐
│  ✨ Colight       │  ← 로고 + 브랜드
│                  │
│  📋 경험 관리     │  ← /experiences (active state)
│  🏢 기업 분석     │  ← /analysis
│  ✍️ 자소서 코칭   │  ← /coaching
│  📊 대시보드      │  ← /dashboard
│                  │
│                  │
│                  │
│  ──────────────  │
│  👤 홍길동       │  ← 사용자 아바타 + 이름
│  admin           │  ← role badge (어드민만)
│                  │
└──────────────────┘
```

### 네비게이션 메뉴

| 메뉴 | 경로 | 아이콘 (lucide-react) |
|------|------|----------------------|
| 경험 관리 | `/experiences` | BookOpen |
| 기업 분석 | `/analysis` | Building2 |
| 자소서 코칭 | `/coaching` | PenTool |
| 대시보드 | `/dashboard` | LayoutDashboard |

### 어드민 추가 메뉴 (role === "admin"일 때)

| 메뉴 | 경로 | 아이콘 |
|------|------|--------|
| 구분선 | --- | --- |
| 어드민 | `/admin` | Shield |

---

## C. 헤더

### `src/components/layout/header.tsx`

- 모바일: 햄버거 버튼 (Sheet 트리거) + 페이지 제목
- 데스크톱: 페이지 제목만
- 우측: 사용자 드롭다운 메뉴

---

## D. 사용자 드롭다운

### `src/components/layout/user-dropdown.tsx`

```
┌──────────────┐
│ 👤 홍길동     │
│ a@test.com   │
├──────────────┤
│ 프로필 설정   │
│ 어드민 ▸     │  ← 어드민만 표시
├──────────────┤
│ 로그아웃      │
└──────────────┘
```

- 프로필 설정: `/settings` (Phase 6.1에서 구현, 지금은 빈 페이지)
- 어드민: `/admin` (role === "admin"일 때만 표시)
- 로그아웃: `useAuthStore().logout()` → `/login` 리디렉션

---

## E. 모바일 네비게이션

### `src/components/layout/mobile-nav.tsx`

- shadcn `Sheet` 컴포넌트 사용
- 좌측에서 슬라이드
- 사이드바와 동일한 메뉴 구성
- 메뉴 클릭 시 Sheet 자동 닫힘

---

## F. 공통 컴포넌트

### LoadingSpinner

```typescript
// src/components/common/loading-spinner.tsx
interface LoadingSpinnerProps {
  size?: "sm" | "md" | "lg";
  text?: string;
}
// Tailwind animate-spin, 센터 정렬
```

### EmptyState

```typescript
// src/components/common/empty-state.tsx
interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}
```

### 각 라우트 빈 상태 메시지

| 라우트 | 제목 | 설명 | 액션 |
|--------|------|------|------|
| `/experiences` | 아직 등록된 경험이 없습니다 | 첫 번째 경험을 등록해보세요! | "경험 등록" 버튼 |
| `/analysis` | 분석된 기업이 없습니다 | 채용공고 URL을 입력해보세요! | "기업 분석" 버튼 |
| `/coaching` | 작성 중인 자소서가 없습니다 | 기업 분석 후 코칭을 시작하세요 | - |
| `/dashboard` | 지원 현황이 없습니다 | 첫 지원을 등록해보세요! | - |

---

## G. 에러 / 404 페이지

### `src/app/(main)/*/error.tsx`

```typescript
"use client";
export default function Error({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20">
      <h2>문제가 발생했습니다</h2>
      <p className="text-muted-foreground">{error.message}</p>
      <Button onClick={reset}>다시 시도</Button>
    </div>
  );
}
```

### `src/app/not-found.tsx`

- 404 페이지 UI
- "홈으로 돌아가기" 링크

---

## 체크리스트

### 레이아웃

- [x] `(main)/layout.tsx` — AuthGuard + Sidebar + Header 조합
- [x] `components/layout/sidebar.tsx` — 로고, 4개 메뉴, 활성 상태, 어드민 링크
- [x] `components/layout/header.tsx` — 모바일 햄버거 + 페이지 제목 + 유저 드롭다운
- [x] `components/layout/mobile-nav.tsx` — Sheet 모바일 메뉴
- [x] `components/layout/user-dropdown.tsx` — 프로필/로그아웃/어드민 메뉴

### 공통 컴포넌트

- [x] `components/common/loading-spinner.tsx`
- [x] `components/common/empty-state.tsx`

### 페이지 UI

- [x] `(main)/experiences/page.tsx` — 빈 상태 UI
- [x] `(main)/analysis/page.tsx` — 빈 상태 UI
- [x] `(main)/coaching/page.tsx` — 빈 상태 UI
- [x] `(main)/dashboard/page.tsx` — 빈 상태 UI
- [x] 각 라우트 `loading.tsx` — LoadingSpinner
- [x] 각 라우트 `error.tsx` — 에러 + 재시도
- [x] `not-found.tsx` — 404 페이지

---

## 검증 방법

1. 데스크톱 (1280px+): 사이드바 고정 표시, 4개 메뉴 네비게이션 정상
2. 모바일 (1024px 이하): 사이드바 숨김, 햄버거 메뉴 → Sheet 표시
3. 메뉴 클릭 → 해당 페이지 이동 + 활성 상태 변경
4. 사용자 드롭다운 → 로그아웃 클릭 → `/login` 이동
5. 어드민 사용자: 사이드바에 "어드민" 메뉴 표시
6. 일반 사용자: "어드민" 메뉴 숨김
7. 각 페이지 빈 상태 UI 정상 표시
8. 존재하지 않는 URL → 404 페이지

---

## 산출물

- `src/app/(main)/layout.tsx`
- `src/components/layout/sidebar.tsx`
- `src/components/layout/header.tsx`
- `src/components/layout/mobile-nav.tsx`
- `src/components/layout/user-dropdown.tsx`
- `src/components/common/loading-spinner.tsx`
- `src/components/common/empty-state.tsx`
- 4개 라우트 `page.tsx` (빈 상태) + `loading.tsx` + `error.tsx`
- `src/app/not-found.tsx`
