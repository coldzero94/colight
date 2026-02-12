# Phase 1: 인증 & 레이아웃

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 이메일/Google 인증 플로우 완성, 공통 사이드바/헤더 레이아웃 구축, 빈 앱 셸 완성 |
| **선행 조건** | Phase 0 완료 |
| **스프린트** | Sprint 0 (Day 4-5) |
| **관련 기능** | F01 (경험 등록 - 인증 필요), F23 (프롬프트 DB - 레이아웃) |
| **예상 공수** | 2일 |
| **산출물** | 로그인/회원가입 페이지, 보호된 라우트, 사이드바+헤더 레이아웃, 공통 UI 컴포넌트, TypeScript 타입 정의 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 1.1 | 인증 페이지 (로그인/회원가입) | ⬜ |
| 1.2 | 인증 미들웨어 | ⬜ |
| 1.3 | 공통 레이아웃 (사이드바/헤더) | ⬜ |
| 1.4 | 공통 컴포넌트 | ⬜ |
| 1.5 | TypeScript 타입 정의 | ⬜ |

---

## Step 1.1: 인증 페이지 (로그인/회원가입)

### 목표
React Hook Form + Zod 기반의 로그인/회원가입 폼, Google OAuth, 비밀번호 재설정, Auth 콜백 라우트를 구현한다.

### 체크리스트

- [ ] 로그인 페이지 구현
  - [ ] `src/app/(auth)/login/page.tsx` 구현
  - [ ] React Hook Form + Zod 유효성 검증 스키마 작성
    - [ ] 이메일: 필수, 이메일 형식 검증
    - [ ] 비밀번호: 필수, 최소 8자
  - [ ] 이메일/비밀번호 로그인 폼 UI
    - [ ] shadcn Input, Label, Button 컴포넌트 사용
    - [ ] 로딩 상태 표시 (Spinner)
    - [ ] 에러 메시지 인라인 표시 (잘못된 이메일/비밀번호)
    - [ ] "비밀번호를 잊으셨나요?" 링크
    - [ ] "계정이 없으신가요? 회원가입" 링크
  - [ ] Google OAuth 로그인 버튼
    - [ ] `supabase.auth.signInWithOAuth({ provider: 'google' })` 호출
    - [ ] Google 아이콘 + "Google로 로그인" 버튼
    - [ ] 구분선 ("또는") 표시
  - [ ] 로그인 성공 시 `/(main)/experiences`로 리디렉션
  - [ ] 로그인 실패 시 에러 toast 표시 (sonner)
- [ ] 회원가입 페이지 구현
  - [ ] `src/app/(auth)/signup/page.tsx` 구현
  - [ ] React Hook Form + Zod 유효성 검증 스키마 작성
    - [ ] 이메일: 필수, 이메일 형식 검증
    - [ ] 비밀번호: 필수, 최소 8자, 영문+숫자 조합
    - [ ] 비밀번호 확인: 비밀번호와 일치 검증
  - [ ] 회원가입 폼 UI
    - [ ] shadcn Input, Label, Button 컴포넌트 사용
    - [ ] 비밀번호 강도 인디케이터 (선택)
    - [ ] "이미 계정이 있으신가요? 로그인" 링크
  - [ ] Google OAuth 회원가입 버튼
  - [ ] 회원가입 성공 시 "이메일 인증 메일을 확인해주세요" 안내 페이지 표시
  - [ ] `supabase.auth.signUp()` 호출
- [ ] 비밀번호 재설정 플로우
  - [ ] `src/app/(auth)/reset-password/page.tsx` 구현
  - [ ] 이메일 입력 → `supabase.auth.resetPasswordForEmail()` 호출
  - [ ] "재설정 이메일이 발송되었습니다" 안내 표시
- [ ] Auth 콜백 라우트
  - [ ] `src/app/auth/callback/route.ts` 구현
  - [ ] OAuth 콜백 처리 (`exchangeCodeForSession`)
  - [ ] 성공 시 `/(main)/experiences`로 리디렉션
  - [ ] 실패 시 `/login?error=auth_error`로 리디렉션
- [ ] (auth) 레이아웃
  - [ ] `src/app/(auth)/layout.tsx` 구현
  - [ ] 센터 정렬 카드 레이아웃
  - [ ] Colight 로고 + 브랜드 텍스트
  - [ ] 이미 로그인된 사용자는 `/(main)`으로 리디렉션

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| LoginPage | `src/app/(auth)/login/page.tsx` | - | 로그인 페이지 (이메일+Google) |
| SignupPage | `src/app/(auth)/signup/page.tsx` | - | 회원가입 페이지 (이메일+Google) |
| ResetPasswordPage | `src/app/(auth)/reset-password/page.tsx` | - | 비밀번호 재설정 페이지 |
| AuthCallback | `src/app/auth/callback/route.ts` | - | OAuth 콜백 처리 라우트 핸들러 |
| AuthLayout | `src/app/(auth)/layout.tsx` | children | 인증 페이지 공통 레이아웃 (센터 카드) |

### 검증 방법
- 이메일 회원가입 → 인증 메일 수신 → 이메일 확인 → 로그인 성공
- Google OAuth 로그인 → 콜백 처리 → 메인 페이지 리디렉션
- 잘못된 이메일/비밀번호 → 에러 메시지 표시
- Zod 유효성 검증: 빈 필드, 형식 오류 시 인라인 에러 표시
- 비밀번호 재설정 이메일 발송 정상

### 산출물
- 로그인 페이지 (`/login`)
- 회원가입 페이지 (`/signup`)
- 비밀번호 재설정 페이지 (`/reset-password`)
- Auth 콜백 라우트 (`/auth/callback`)
- (auth) 레이아웃

---

## Step 1.2: 인증 미들웨어

### 목표
`(main)/` 경로를 보호하여 비인증 사용자는 로그인 페이지로 리디렉션하고, 인증된 사용자의 세션을 자동 갱신한다.

### 체크리스트

- [ ] Next.js 미들웨어 구현
  - [ ] `src/middleware.ts` 구현
  - [ ] Supabase 미들웨어 클라이언트 사용 (`src/lib/supabase/middleware.ts`)
  - [ ] 세션 갱신 로직: 모든 요청에서 `supabase.auth.getUser()` 호출하여 쿠키 갱신
- [ ] 보호 라우트 설정
  - [ ] `/(main)/*` 경로: 인증 필수
    - [ ] 미인증 → `/login`으로 리디렉션
    - [ ] 리디렉션 시 원래 URL을 `?redirect=` 쿼리로 전달
  - [ ] `/(auth)/*` 경로: 비인증 전용
    - [ ] 이미 인증됨 → `/(main)/experiences`로 리디렉션
  - [ ] `/api/*` 경로: 세션 갱신만 (리디렉션 없음)
  - [ ] 정적 파일 (`_next/`, `favicon.ico` 등): 미들웨어 스킵
- [ ] matcher 설정
  - [ ] `export const config = { matcher: ['/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)'] }`
- [ ] 로그인 후 원래 페이지로 복귀
  - [ ] 로그인 성공 시 `redirect` 쿼리가 있으면 해당 URL로 이동
  - [ ] 없으면 기본 `/(main)/experiences`로 이동

### 검증 방법
- 비로그인 상태에서 `/experiences` 접속 → `/login` 리디렉션
- 비로그인 상태에서 `/login` 접속 → 정상 표시
- 로그인 상태에서 `/login` 접속 → `/experiences` 리디렉션
- 로그인 상태에서 `/experiences` 접속 → 정상 표시
- 세션 만료 후 페이지 이동 → `/login` 리디렉션 (세션 갱신 실패 시)

### 산출물
- `src/middleware.ts` (라우트 보호 + 세션 갱신 미들웨어)

---

## Step 1.3: 공통 레이아웃 (사이드바/헤더)

### 목표
인증된 사용자를 위한 메인 레이아웃을 구현한다. 데스크톱에서는 사이드바, 모바일에서는 햄버거 메뉴를 제공한다.

### 체크리스트

- [ ] 메인 레이아웃 구현
  - [ ] `src/app/(main)/layout.tsx` 구현
  - [ ] 사이드바 + 메인 콘텐츠 영역 2단 구성
  - [ ] 서버 컴포넌트에서 사용자 정보 로드
- [ ] 사이드바 구현
  - [ ] `src/components/layout/sidebar.tsx` 구현
  - [ ] 상단: Colight 로고 + 브랜드명
  - [ ] 네비게이션 메뉴 항목:
    - [ ] 경험 관리 (`/experiences`) - 아이콘: BookOpen
    - [ ] 기업 분석 (`/analysis`) - 아이콘: Building2
    - [ ] 자소서 코칭 (`/coaching`) - 아이콘: PenTool
    - [ ] 대시보드 (`/dashboard`) - 아이콘: LayoutDashboard
  - [ ] 현재 경로 활성 상태 표시 (active state)
  - [ ] 하단: 사용자 아바타 + 이름 + 설정 드롭다운
  - [ ] 데스크톱: 고정 사이드바 (w-64)
  - [ ] 반응형: lg 이하에서 사이드바 숨김
- [ ] 헤더 구현
  - [ ] `src/components/layout/header.tsx` 구현
  - [ ] 모바일: 햄버거 메뉴 버튼 (Sheet 트리거)
  - [ ] 페이지 제목 표시 (현재 라우트 기반)
  - [ ] 우측: 사용자 드롭다운 메뉴
    - [ ] 프로필 보기
    - [ ] 설정
    - [ ] 로그아웃 (`supabase.auth.signOut()`)
- [ ] 모바일 네비게이션
  - [ ] shadcn Sheet 컴포넌트 사용
  - [ ] 햄버거 버튼 클릭 → 좌측에서 슬라이드
  - [ ] 사이드바와 동일한 메뉴 구성
  - [ ] 메뉴 항목 클릭 시 자동 닫힘
- [ ] 로그아웃 기능
  - [ ] `supabase.auth.signOut()` 호출
  - [ ] 로그아웃 성공 → `/login`으로 리디렉션
  - [ ] 쿠키 정리

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| MainLayout | `src/app/(main)/layout.tsx` | children | 메인 앱 레이아웃 (사이드바+콘텐츠) |
| Sidebar | `src/components/layout/sidebar.tsx` | user, pathname | 데스크톱 사이드바 네비게이션 |
| Header | `src/components/layout/header.tsx` | user, title | 헤더 (모바일 햄버거+유저 드롭다운) |
| MobileNav | `src/components/layout/mobile-nav.tsx` | pathname, onClose | 모바일 Sheet 네비게이션 |
| UserDropdown | `src/components/layout/user-dropdown.tsx` | user | 사용자 드롭다운 메뉴 (프로필/설정/로그아웃) |

### 검증 방법
- 데스크톱 (1280px+): 사이드바 고정 표시, 4개 메뉴 항목 정상
- 태블릿/모바일 (1024px 이하): 사이드바 숨김, 햄버거 메뉴 표시
- 햄버거 클릭 → Sheet 사이드바 슬라이드
- 메뉴 클릭 → 해당 페이지 이동 + 활성 상태 변경
- 사용자 드롭다운 → 로그아웃 클릭 → `/login` 이동

### 산출물
- `src/app/(main)/layout.tsx` (메인 레이아웃)
- `src/components/layout/sidebar.tsx` (사이드바)
- `src/components/layout/header.tsx` (헤더)
- `src/components/layout/mobile-nav.tsx` (모바일 네비게이션)
- `src/components/layout/user-dropdown.tsx` (사용자 드롭다운)

---

## Step 1.4: 공통 컴포넌트

### 목표
앱 전체에서 재사용되는 공통 UI 컴포넌트(로딩, 빈 상태, 에러)와 각 라우트 그룹의 기본 로딩/에러 페이지를 생성한다.

### 체크리스트

- [ ] 로딩 스피너 컴포넌트
  - [ ] `src/components/common/loading-spinner.tsx` 구현
  - [ ] Props: `size?: 'sm' | 'md' | 'lg'`, `text?: string`
  - [ ] Tailwind animate-spin 사용
  - [ ] 센터 정렬 (flex items-center justify-center)
- [ ] 빈 상태 컴포넌트
  - [ ] `src/components/common/empty-state.tsx` 구현
  - [ ] Props: `icon?: ReactNode`, `title: string`, `description?: string`, `action?: { label: string, onClick: () => void }`
  - [ ] 아이콘 + 제목 + 설명 + 액션 버튼 구성
- [ ] 라우트별 loading.tsx 생성
  - [ ] `src/app/(main)/experiences/loading.tsx` → LoadingSpinner 표시
  - [ ] `src/app/(main)/analysis/loading.tsx` → LoadingSpinner 표시
  - [ ] `src/app/(main)/coaching/loading.tsx` → LoadingSpinner 표시
  - [ ] `src/app/(main)/dashboard/loading.tsx` → LoadingSpinner 표시
- [ ] 라우트별 error.tsx 생성
  - [ ] `src/app/(main)/experiences/error.tsx` → 에러 메시지 + 재시도 버튼 ("use client")
  - [ ] `src/app/(main)/analysis/error.tsx` → 에러 메시지 + 재시도 버튼
  - [ ] `src/app/(main)/coaching/error.tsx` → 에러 메시지 + 재시도 버튼
  - [ ] `src/app/(main)/dashboard/error.tsx` → 에러 메시지 + 재시도 버튼
- [ ] 글로벌 not-found 페이지
  - [ ] `src/app/not-found.tsx` → 404 페이지 UI
- [ ] 각 메인 페이지에 빈 상태 표시
  - [ ] `/experiences` → "아직 등록된 경험이 없습니다. 첫 번째 경험을 등록해보세요!" + "경험 등록" 버튼
  - [ ] `/analysis` → "분석된 기업이 없습니다. 채용공고 URL을 입력해보세요!"
  - [ ] `/coaching` → "작성 중인 자소서가 없습니다."
  - [ ] `/dashboard` → "지원 현황이 없습니다."

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| LoadingSpinner | `src/components/common/loading-spinner.tsx` | size, text | 재사용 로딩 스피너 |
| EmptyState | `src/components/common/empty-state.tsx` | icon, title, description, action | 빈 상태 안내 |
| ErrorPage | `src/app/(main)/*/error.tsx` | error, reset | 에러 페이지 (재시도 버튼) |
| LoadingPage | `src/app/(main)/*/loading.tsx` | - | 로딩 페이지 (스피너) |
| NotFound | `src/app/not-found.tsx` | - | 404 페이지 |

### 검증 방법
- 각 라우트 접속 시 loading.tsx → 실제 페이지 순서로 렌더링
- 서버 에러 발생 시 error.tsx 표시 + "다시 시도" 버튼 동작
- 존재하지 않는 URL → not-found.tsx 표시
- 빈 상태에서 각 페이지 정상 표시 (EmptyState 컴포넌트)

### 산출물
- `src/components/common/loading-spinner.tsx`
- `src/components/common/empty-state.tsx`
- 4개 라우트별 `loading.tsx`, `error.tsx`
- `src/app/not-found.tsx`

---

## Step 1.5: TypeScript 타입 정의

### 목표
프로젝트 전체에서 사용할 도메인 타입을 정의한다. Supabase 자동 생성 타입을 기반으로 비즈니스 로직에 맞는 타입을 확장한다.

### 체크리스트

- [ ] 경험 관련 타입
  - [ ] `src/types/experience.ts` 작성
  - [ ] `Experience` 인터페이스 (DB 컬럼 + 관계)
  - [ ] `ExperienceFormData` 타입 (폼 입력 데이터)
  - [ ] `ExperienceStar` 타입 (STAR 구조)
  - [ ] `ExperienceWithWeapons` 타입 (경험 + 무기 태깅 결과)
  - [ ] `ExperienceListItem` 타입 (목록 표시용 요약)
- [ ] 기업 분석 관련 타입
  - [ ] `src/types/analysis.ts` 작성
  - [ ] `CompanyAnalysis` 인터페이스
  - [ ] `JobPosting` 타입 (파싱된 채용공고)
  - [ ] `CompanyProfile` 타입 (기업 프로필)
  - [ ] `TalentProfile` 타입 (인재상)
  - [ ] `MatchingResult` 타입 (적합도 매칭 결과)
- [ ] 코칭 관련 타입
  - [ ] `src/types/coaching.ts` 작성
  - [ ] `CoverLetter` 인터페이스
  - [ ] `CoverLetterVersion` 타입
  - [ ] `CoachingSession` 타입
  - [ ] `QuestionAnalysis` 타입 (문항 분석 결과)
  - [ ] `CoachingFeedback` 타입 (첨삭 피드백)
- [ ] 무기 관련 타입
  - [ ] `src/types/weapon.ts` 작성
  - [ ] `WeaponCategory` 인터페이스 (대분류)
  - [ ] `WeaponSubCategory` 타입 (소분류)
  - [ ] `WeaponCode` 리터럴 타입 (`'W01' | 'W02' | ... | 'W07'`)
  - [ ] `WeaponTagResult` 타입 (AI 태깅 결과)
  - [ ] `ExperienceWeapon` 타입 (경험-무기 매핑)
  - [ ] `WEAPON_COLORS` 상수 맵 (W01~W07 → 색상)
- [ ] 공통 타입
  - [ ] `src/types/common.ts` 작성
  - [ ] `ApiResponse<T>` 제네릭 타입
  - [ ] `PaginatedResponse<T>` 타입
  - [ ] `SortOption` 타입
  - [ ] `FilterOption` 타입
  - [ ] `User` 타입 (Supabase User 확장)

### 검증 방법
- `pnpm run build` → 타입 에러 0건
- 각 타입 파일에서 Supabase 자동 생성 타입과의 호환성 확인
- IDE 자동완성이 정상 동작하는지 확인

### 산출물
- `src/types/experience.ts`
- `src/types/analysis.ts`
- `src/types/coaching.ts`
- `src/types/weapon.ts`
- `src/types/common.ts`

---

## Phase 완료 체크리스트

### 기능
- [ ] 이메일 회원가입 → 이메일 인증 → 로그인 정상
- [ ] Google OAuth 로그인/회원가입 정상
- [ ] 비밀번호 재설정 이메일 발송 정상
- [ ] 비인증 사용자 → `/login` 리디렉션 정상
- [ ] 인증 사용자 → `/(main)/*` 접근 정상
- [ ] 사이드바 4개 메뉴 네비게이션 정상
- [ ] 모바일 햄버거 메뉴 정상
- [ ] 로그아웃 → 세션 삭제 → `/login` 리디렉션
- [ ] 각 페이지에 빈 상태 UI 표시

### 테스트
- [ ] `pnpm run build` → 빌드 성공
- [ ] 로그인/회원가입 전체 플로우 수동 테스트
- [ ] 데스크톱/모바일 반응형 레이아웃 테스트
- [ ] 미들웨어 보호 라우트 테스트 (인증/비인증)
- [ ] error.tsx 동작 테스트 (의도적 에러 발생)

### 코드 품질
- [ ] TypeScript strict mode 에러 0건
- [ ] ESLint 경고/에러 0건
- [ ] 모든 폼에 Zod 유효성 검증 적용
- [ ] 사용자 입력 데이터 서버 사이드 검증
- [ ] XSS 방지: 사용자 입력 데이터 이스케이프 처리

---

## 다음 Phase

**→ Phase 2: 경험 CRUD** (Sprint 1, Day 1-3)
- 경험 목록/등록/상세/수정/삭제 CRUD 구현
- STAR 구조 입력 폼
- Server Actions 기반 데이터 처리
