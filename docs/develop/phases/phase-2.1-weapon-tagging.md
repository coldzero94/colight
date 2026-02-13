# Phase 2.1: AI 무기 자동 태깅

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 경험 등록/수정 시 AI가 자동으로 무기(역량 카테고리)를 태깅하고, 사용자가 수동으로 편집할 수 있도록 구현 |
| **선행 조건** | Phase 2 완료 |
| **스프린트** | Sprint 1 (Day 4-5) |
| **관련 기능** | F03 (역량 태그 자동 분류), F01 (경험 등록 - 태깅 연동) |
| **예상 공수** | 2일 |
| **산출물** | AI 태깅 API 엔드포인트, 자동 태깅 트리거, 무기 배지 UI, 무기 셀렉터 모달, 무기별 필터 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 2.1.1 | AI 태깅 API | ⬜ |
| 2.1.2 | 자동 트리거 | ⬜ |
| 2.1.3 | 무기 태그 UI | ⬜ |
| 2.1.4 | 무기별 필터 | ⬜ |

---

## Step 2.1.1: AI 태깅 API

### 목표
경험 텍스트를 경량 모델 (Gemini/Groq)로 분석하여 무기 카테고리를 자동 분류하는 Go backend API를 구현한다. 프롬프트는 DB에서 동적으로 로드한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

**Backend (Go) -- MockAIClient 패턴 + testdata/ fixtures 사용**

| 테스트 파일 | 테스트 함수 | 설명 |
|------------|-----------|------|
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_Success` | MockAIClient로 경량 LLM 응답 모킹, 주 무기 + 부 무기 분류 결과 파싱 확인 |
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_ParseAIResponse` | `testdata/ai/weapon_tagging_response.json` fixture 로드 후 JSON 파싱 + 구조 검증 |
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_ConfidenceSorting` | confidence 기준 내림차순 정렬 확인 (가장 높은 confidence가 primary) |
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_InvalidWeaponCode` | AI가 존재하지 않는 weapon_code 반환 시 에러 처리 |
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_ShortExperience` | 50자 미만 경험 입력 시 400 에러 반환 |
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_AIError` | MockAIClient에서 rate limit 에러 반환 시 적절한 에러 핸들링 |
| `internal/service/weapon_tagging_service_test.go` | `TestTagExperience_JSONParseFailure` | AI가 잘못된 JSON 반환 시 재시도 1회 후 실패 응답 |
| `internal/controller/weapon_tagging_controller_test.go` | `TestWeaponTaggingController_Tag` | POST /v1/experiences/:id/tag 통합 테스트 (200 반환) |
| `internal/controller/weapon_tagging_controller_test.go` | `TestWeaponTaggingController_Unauthorized` | 미인증 요청 시 401 반환 |
| `internal/controller/weapon_tagging_controller_test.go` | `TestWeaponTaggingController_Forbidden` | 타인의 경험 태깅 시도 시 403 반환 |

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/weapon_tagging_service_test.go` 작성 (7 tests)
  - [ ] `internal/controller/weapon_tagging_controller_test.go` 작성 (3 tests)
  - [ ] `testdata/ai/weapon_tagging_response.json` fixture 파일 작성
- [ ] 구현 (GREEN)
  - [ ] Go backend API 구현
    - [ ] `POST /v1/experiences/:id/tag` 엔드포인트 구현
    - [ ] 인증 확인 (미인증 → 401)
    - [ ] 경험 소유자 확인 (타인 → 403)
  - [ ] 프롬프트 DB 로드
    - [ ] `prompt_templates` 테이블에서 `category='experience_classify', sub_category='weapon_tagging'` 조회
    - [ ] `is_active=true` AND 최신 `version` 필터
    - [ ] 프롬프트 캐싱 (인메모리, 5분 TTL)
  - [ ] 무기 카테고리 DB 로드
    - [ ] `weapon_categories` 테이블에서 전체 목록 조회 (대분류 + 소분류)
    - [ ] 프롬프트 변수 `{{weapon_categories}}`에 주입
    - [ ] 포맷: 코드 - 이름 - 설명 - 키워드 (구조화된 텍스트)
  - [ ] 경험 텍스트 준비
    - [ ] experience 조회 (title, situation, task, action, result, raw_content)
    - [ ] STAR 필드를 하나의 텍스트로 조합
    - [ ] raw_content가 있으면 함께 포함
    - [ ] 프롬프트 변수 `{{experience_text}}`에 주입
  - [ ] 경량 모델 (Gemini/Groq) 호출
    - [ ] 경량 LLM 호출 (공통 인터페이스)
    - [ ] temperature: 0.2 (DB에서 로드)
    - [ ] max_tokens: 2000 (DB에서 로드)
    - [ ] JSON mode 활성화 (`response_format: { type: "json_object" }`)
    - [ ] 타임아웃: 30초
  - [ ] AI 응답 파싱
    - [ ] JSON 응답 파싱 (try-catch)
    - [ ] Zod 스키마로 응답 구조 검증
    - [ ] 필수 필드 확인: primary_weapon, secondary_weapons
    - [ ] confidence 범위 검증 (0.0 ~ 1.0)
    - [ ] weapon_code 유효성 검증 (DB에 존재하는 코드인지)
  - [ ] DB 저장
    - [ ] 기존 `experience_weapons` 레코드 삭제 (해당 experience_id)
    - [ ] 새 분류 결과 INSERT
      - [ ] primary_weapon → `is_primary=true`
      - [ ] secondary_weapons → `is_primary=false`
      - [ ] 각각 `confidence`, `reasoning` 저장
      - [ ] `user_confirmed=false`, `user_modified=false`
    - [ ] 트랜잭션 처리 (삭제 + 삽입을 원자적으로)
  - [ ] 프롬프트 사용 통계 업데이트
    - [ ] `prompt_templates.usage_count` 증가
    - [ ] `avg_latency_ms` 업데이트 (응답 시간 측정)
  - [ ] 에러 핸들링
    - [ ] LLM API 에러 → 500 + 에러 메시지
    - [ ] JSON 파싱 실패 → 재시도 1회 후 실패 응답
    - [ ] Rate limit → 429 반환
    - [ ] 경험이 너무 짧은 경우 (50자 미만) → 400 + 안내 메시지
- [ ] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| POST | `/v1/experiences/:id/tag` | `{}` (body 없음, id는 URL param) | `{ primary_weapon: { code, sub_code, name, confidence, reasoning }, secondary_weapons: [{ code, sub_code, name, confidence, reasoning }], matchable_questions: string[], strength_keywords: string[] }` |

### 산출물
- Go backend: `POST /v1/experiences/:id/tag` 엔드포인트 (backend 레포지토리)
- 프론트엔드: `@/api/generated/sdk.gen.ts` (SDK에 태깅 API 포함)
- 프론트엔드: `src/hooks/use-weapon-tagging.ts` (태깅 상태 관리 훅)

---

## Step 2.1.2: 자동 트리거

### 목표
경험 등록/수정 후 자동으로 AI 태깅이 실행되도록 트리거를 구현하고, 사용자에게 로딩 상태를 표시한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md (frontend), docs/develop/12-backend-testing.md (backend)

**Backend (Go)**

| 테스트 파일 | 테스트 함수 | 설명 |
|------------|-----------|------|
| `internal/service/weapon_tagging_service_test.go` | `TestRetagExperience_SkipUserConfirmed` | `user_confirmed=true`인 태그가 있으면 재태깅 스킵 |
| `internal/service/weapon_tagging_service_test.go` | `TestRetagExperience_DeleteOnlyAITags` | 재태깅 시 `user_modified=false`인 태그만 삭제 |

**Frontend**

| 테스트 파일 | 테스트 케이스 | 설명 |
|------------|-------------|------|
| `src/hooks/__tests__/use-weapon-tagging.test.ts` | `describe('useWeaponTagging')` / `it('transitions through idle → tagging → success states')` | 태깅 상태 전이 확인 |
| `src/hooks/__tests__/use-weapon-tagging.test.ts` | `describe('useWeaponTagging')` / `it('handles tagging error and allows retry')` | 태깅 실패 시 에러 상태 + 재시도 |
| `src/components/experiences/__tests__/tagging-status.test.tsx` | `describe('TaggingStatus')` / `it('shows loading skeleton during tagging')` | 태깅 진행 중 스켈레톤 UI 표시 |
| `src/components/experiences/__tests__/tagging-status.test.tsx` | `describe('TaggingStatus')` / `it('shows error message with retry button on failure')` | 태깅 실패 시 에러 메시지 + 재시도 버튼 |

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/weapon_tagging_service_test.go`에 재태깅 테스트 추가 (2 tests)
  - [ ] `src/hooks/__tests__/use-weapon-tagging.test.ts` 작성
  - [ ] `src/components/experiences/__tests__/tagging-status.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 경험 등록 후 자동 태깅 트리거
    - [ ] 프론트엔드: 경험 생성 후 클라이언트에서 태깅 API 호출
    - [ ] SDK를 통해 `POST /v1/experiences/:id/tag` 호출
    - [ ] UI에서 로딩 상태를 표시
  - [ ] 경험 수정 후 재태깅 트리거
    - [ ] 프론트엔드: STAR 필드 변경 시에만 재태깅 (제목/기간만 변경 시 스킵)
    - [ ] Go backend: 기존 태그의 `user_confirmed=true`인 경우 재태깅 스킵
    - [ ] Go backend: 재태깅 시 기존 AI 태그만 삭제 (`user_modified=false`인 것만)
  - [ ] 태깅 상태 관리 (클라이언트)
    - [ ] `src/hooks/use-weapon-tagging.ts` 커스텀 훅 구현
    - [ ] 상태: `idle` | `tagging` | `success` | `error`
    - [ ] `triggerTagging(experienceId)` 함수
    - [ ] 태깅 완료 시 경험 데이터 재조회 (react-query invalidation 또는 router.refresh)
  - [ ] 태깅 로딩 UI
    - [ ] 경험 상세 페이지에서 태깅 진행 중 표시
    - [ ] 무기 배지 영역: Skeleton 로딩 애니메이션
    - [ ] "AI가 경험을 분석하고 있어요..." 텍스트
    - [ ] 예상 소요 시간: "약 3~5초"
    - [ ] 태깅 완료: 배지 표시 + 성공 toast
    - [ ] 태깅 실패: 에러 메시지 + "다시 시도" 버튼
  - [ ] 수동 태깅 재실행 버튼
    - [ ] 경험 상세 페이지에 "AI 재분석" 버튼
    - [ ] 클릭 → 태깅 API 재호출 → 결과 업데이트
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| useWeaponTagging | `src/hooks/use-weapon-tagging.ts` | - | 무기 태깅 상태 관리 훅 |
| TaggingStatus | `src/components/experiences/tagging-status.tsx` | status, onRetry | 태깅 진행/완료/에러 상태 표시 |

### 산출물
- 수정: `src/lib/actions/experience.ts` (태깅 트리거 연동)
- `src/hooks/use-weapon-tagging.ts`
- `src/components/experiences/tagging-status.tsx`

---

## Step 2.1.3: 무기 태그 UI

### 목표
무기 배지를 색상으로 시각화하고, 수동으로 무기를 편집할 수 있는 셀렉터 모달을 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 파일 | 테스트 케이스 | 설명 |
|------------|-------------|------|
| `src/components/experiences/__tests__/weapon-badge.test.tsx` | `describe('WeaponBadge')` / `it('renders correct color for each weapon code W01-W07')` | 무기 코드별 색상 매핑 (7가지) 정확성 |
| `src/components/experiences/__tests__/weapon-badge.test.tsx` | `describe('WeaponBadge')` / `it('renders primary weapon with emphasis styling')` | 주 무기 강조 스타일 (두꺼운 테두리, 큰 크기) |
| `src/components/experiences/__tests__/weapon-badge.test.tsx` | `describe('WeaponBadge')` / `it('renders correct icon for each weapon code')` | 무기 코드별 아이콘 매핑 정확성 |
| `src/components/experiences/__tests__/weapon-selector-modal.test.tsx` | `describe('WeaponSelectorModal')` / `it('opens modal and displays 7 weapon categories')` | 모달 열림 + 7대 무기 대분류 표시 |
| `src/components/experiences/__tests__/weapon-selector-modal.test.tsx` | `describe('WeaponSelectorModal')` / `it('expands subcategories on category click')` | 대분류 클릭 → 소분류 아코디언 펼침 |
| `src/components/experiences/__tests__/weapon-selector-modal.test.tsx` | `describe('WeaponSelectorModal')` / `it('saves selection with user_modified=true')` | 무기 선택 저장 시 user_modified 플래그 설정 |

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/experiences/__tests__/weapon-badge.test.tsx` 작성
  - [ ] `src/components/experiences/__tests__/weapon-selector-modal.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 무기 배지 컴포넌트 개선
    - [ ] `src/components/experiences/weapon-badge.tsx` 구현 (기존 weapon-badges.tsx 확장)
    - [ ] Props: `weapon: { code, name, confidence, is_primary }`, `size: 'sm' | 'md' | 'lg'`
    - [ ] 무기별 색상 매핑 (W01~W07)
      - [ ] W01 위기극복: `bg-red-100 text-red-700 border-red-200` (#EF4444)
      - [ ] W02 리더십: `bg-amber-100 text-amber-700 border-amber-200` (#F59E0B)
      - [ ] W03 팀워크/협업: `bg-emerald-100 text-emerald-700 border-emerald-200` (#10B981)
      - [ ] W04 도전정신: `bg-violet-100 text-violet-700 border-violet-200` (#8B5CF6)
      - [ ] W05 문제해결: `bg-blue-100 text-blue-700 border-blue-200` (#3B82F6)
      - [ ] W06 소통/설득: `bg-pink-100 text-pink-700 border-pink-200` (#EC4899)
      - [ ] W07 성장/학습: `bg-indigo-100 text-indigo-700 border-indigo-200` (#6366F1)
    - [ ] 주 무기: 테두리 두껍게 + "주" 라벨 또는 약간 큰 크기
    - [ ] 부 무기: 기본 크기
    - [ ] confidence 표시 (선택): 배지에 80% 같은 확신도 작게 표시
    - [ ] 아이콘 포함: W01 🔥, W02 👑, W03 🤝, W04 🚀, W05 🧩, W06 💬, W07 📚
  - [ ] 무기 셀렉터 모달 구현
    - [ ] `src/components/experiences/weapon-selector-modal.tsx` 구현
    - [ ] shadcn Dialog 사용
    - [ ] 트리거: 경험 상세 페이지의 "무기 편집" 버튼 (연필 아이콘)
    - [ ] 모달 내용:
      - [ ] 7대 무기 대분류를 카드/타일로 표시
      - [ ] 각 대분류 클릭 → 4개 소분류 펼침 (아코디언)
      - [ ] 체크박스로 선택/해제
      - [ ] 주 무기 선택: 라디오 버튼 (하나만 선택 가능)
      - [ ] 현재 AI 추천 결과 표시 ("AI 추천" 배지)
    - [ ] 저장 로직:
      - [ ] 선택된 무기 → `experience_weapons` 업데이트
      - [ ] `user_modified=true` 설정 (사용자가 수동 편집함을 표시)
      - [ ] `user_confirmed=true` 설정
    - [ ] Server Action: `updateExperienceWeapons(experienceId, weapons)`
  - [ ] 무기 확정 기능
    - [ ] "이 분류가 맞아요" 확인 버튼 → `user_confirmed=true` 업데이트
    - [ ] 확정 후에는 경험 수정 시 재태깅 스킵
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| WeaponBadge | `src/components/experiences/weapon-badge.tsx` | weapon, size | 개별 무기 배지 (색상+아이콘) |
| WeaponSelectorModal | `src/components/experiences/weapon-selector-modal.tsx` | experienceId, currentWeapons, onSave | 무기 수동 편집 모달 |
| WeaponCategoryCard | `src/components/experiences/weapon-category-card.tsx` | category, subCategories, selected, onToggle | 무기 카테고리 선택 카드 |

### 산출물
- `src/components/experiences/weapon-badge.tsx`
- `src/components/experiences/weapon-selector-modal.tsx`
- `src/components/experiences/weapon-category-card.tsx`
- `src/lib/actions/weapon.ts` (무기 업데이트 Server Actions)
- `src/lib/constants/weapon-colors.ts` (무기별 색상/아이콘 상수)

---

## Step 2.1.4: 무기별 필터

### 목표
경험 목록 페이지에 무기별 필터 탭을 추가하여, 특정 무기를 가진 경험만 필터링한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 파일 | 테스트 케이스 | 설명 |
|------------|-------------|------|
| `src/components/experiences/__tests__/weapon-filter-tabs.test.tsx` | `describe('WeaponFilterTabs')` / `it('renders all 8 tabs with correct counts')` | 전체 + 7 무기 탭 + 괄호 안 경험 수 표시 |
| `src/components/experiences/__tests__/weapon-filter-tabs.test.tsx` | `describe('WeaponFilterTabs')` / `it('highlights active weapon tab with correct color')` | 선택된 탭 활성 스타일 + 무기 색상 적용 |
| `src/components/experiences/__tests__/weapon-filter-tabs.test.tsx` | `describe('WeaponFilterTabs')` / `it('updates URL params on tab click')` | 탭 클릭 시 URL searchParams에 weapon 파라미터 반영 |
| `src/components/experiences/__tests__/weapon-filter-tabs.test.tsx` | `describe('WeaponFilterTabs')` / `it('shows empty state for zero-count weapon filter')` | 필터 결과 0건 시 EmptyState 표시 |

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/experiences/__tests__/weapon-filter-tabs.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 무기 필터 탭 컴포넌트
    - [ ] `src/components/experiences/weapon-filter-tabs.tsx` 구현
    - [ ] 경험 목록 상단에 가로 스크롤 탭으로 표시
    - [ ] 탭 항목:
      - [ ] "전체" (기본 선택) - 모든 경험
      - [ ] 🔥 위기극복 (N) - W01 태깅된 경험 수
      - [ ] 👑 리더십 (N) - W02 태깅된 경험 수
      - [ ] 🤝 팀워크 (N) - W03 태깅된 경험 수
      - [ ] 🚀 도전정신 (N) - W04 태깅된 경험 수
      - [ ] 🧩 문제해결 (N) - W05 태깅된 경험 수
      - [ ] 💬 소통/설득 (N) - W06 태깅된 경험 수
      - [ ] 📚 성장/학습 (N) - W07 태깅된 경험 수
    - [ ] 각 탭에 해당 무기 색상 적용
    - [ ] 선택된 탭: 진한 배경색 + 밑줄 또는 active 스타일
    - [ ] 괄호 안 숫자: 해당 무기가 태깅된 경험 수 (주 무기 + 부 무기 포함)
  - [ ] 필터 로직 구현
    - [ ] URL searchParams로 필터 상태 관리: `?weapon=W01`
    - [ ] `getExperiences` Server Action에 weapon 필터 파라미터 추가
    - [ ] Supabase 쿼리: `experience_weapons` 테이블 JOIN → `weapon_code` 필터
    - [ ] 필터 변경 시 URL 업데이트 + 목록 재조회
  - [ ] 무기별 경험 수 집계 쿼리
    - [ ] `getWeaponCounts` Server Action 구현
    - [ ] `SELECT weapon_code, COUNT(DISTINCT experience_id) FROM experience_weapons WHERE experience_id IN (사용자 경험) GROUP BY weapon_code`
    - [ ] 결과를 탭에 표시
  - [ ] 빈 필터 결과 처리
    - [ ] 특정 무기로 필터했는데 경험이 0건인 경우
    - [ ] EmptyState: "아직 {무기명} 역량의 경험이 없습니다"
    - [ ] "경험 등록" 버튼 + "다른 무기 보기" 링크
  - [ ] 복합 필터 (선택)
    - [ ] 정렬과 무기 필터 동시 적용 가능
    - [ ] URL: `?weapon=W01&sort=created_at`
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| WeaponFilterTabs | `src/components/experiences/weapon-filter-tabs.tsx` | counts, activeWeapon, onChange | 무기별 필터 탭 (가로 스크롤) |

### API 엔드포인트 (Server Actions)

| Method | Path | Request | Response |
|--------|------|---------|----------|
| Server Action | `getExperiences(options?)` | `{ sort?, category?, weapon?: WeaponCode }` | `ExperienceWithWeapons[]` |
| Server Action | `getWeaponCounts()` | - | `Record<WeaponCode, number>` |

### 산출물
- `src/components/experiences/weapon-filter-tabs.tsx`
- 수정: `src/lib/actions/experience.ts` (weapon 필터 파라미터 추가)
- `src/lib/actions/weapon.ts`에 `getWeaponCounts` 추가

---

## Phase 완료 체크리스트

### 기능
- [ ] AI 태깅 API: 경험 → 경량 모델 (Gemini/Groq) → 무기 분류 정상 동작
- [ ] 프롬프트 DB에서 동적 로드 (하드코딩 아님)
- [ ] 경험 등록 후 자동 태깅 트리거 → 3~5초 후 결과 표시
- [ ] 경험 수정 후 재태깅 (STAR 변경 시에만)
- [ ] 무기 배지 7가지 색상 정상 표시 (주 무기/부 무기 구분)
- [ ] 무기 셀렉터 모달로 수동 편집 가능
- [ ] "이 분류가 맞아요" 확정 → 재태깅 보호
- [ ] 무기별 필터 탭 정상 동작 (경험 수 표시)
- [ ] 빈 필터 결과 적절한 EmptyState 표시

### 테스트
- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

### 코드 품질
- [ ] TypeScript strict mode 에러 0건
- [ ] ESLint 경고/에러 0건
- [ ] LLM API 에러 핸들링 (타임아웃, rate limit, 파싱 에러)
- [ ] AI 응답 Zod 스키마 검증
- [ ] 프롬프트 인젝션 방지: 사용자 입력 텍스트 이스케이프
- [ ] API 키 환경변수 검증 (없으면 명확한 에러 메시지)

---

## 다음 Phase

**→ Phase 3: 기업 분석** (Sprint 2, Day 1-5)
- 채용공고 크롤링 엔진 구현 (Cheerio + Playwright)
- DART OpenAPI / 네이버 뉴스 / 사람인 API 연동
- AI 기업 종합 분석 파이프라인
- 기업 분석 UI
