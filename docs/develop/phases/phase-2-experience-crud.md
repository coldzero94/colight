# Phase 2: 경험 CRUD

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 경험 목록/등록/상세/수정/삭제 CRUD 전체 구현, STAR 구조 입력 폼 완성 |
| **선행 조건** | Phase 1 완료 |
| **스프린트** | Sprint 1 (Day 1-3) |
| **관련 기능** | F01 (경험 등록), F03 (무기 태깅 - 준비), F04 (활용 이력 - 준비) |
| **예상 공수** | 3일 |
| **산출물** | 경험 목록 페이지, 경험 등록/수정 폼, 경험 상세 페이지, Go Backend API CRUD, 경험 카드 컴포넌트 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 2.1 | 경험 목록 페이지 | ⬜ |
| 2.2 | 경험 등록 폼 | ⬜ |
| 2.3 | 경험 CRUD API (Server Actions) | ⬜ |
| 2.4 | 경험 상세 페이지 | ⬜ |
| 2.5 | 경험 카드 컴포넌트 | ⬜ |

---

## Step 2.1: 경험 목록 페이지

### 목표
사용자의 경험 목록을 그리드/리스트 뷰로 표시하고, 정렬/필터 기능을 제공한다.

### 체크리스트

- [ ] 경험 목록 페이지 구현
  - [ ] `src/app/(main)/experiences/page.tsx` 구현 (서버 컴포넌트)
  - [ ] Go API를 통해 현재 사용자의 경험 목록 조회
  - [ ] `experiences` 테이블 + `experience_weapons` JOIN 쿼리
  - [ ] 최신순 정렬 (기본)
- [ ] 뷰 토글 (그리드/리스트)
  - [ ] `src/components/experiences/view-toggle.tsx` 구현
  - [ ] 그리드 뷰: 3열 카드 레이아웃 (lg:grid-cols-3 md:grid-cols-2 grid-cols-1)
  - [ ] 리스트 뷰: 1열 상세 목록 레이아웃
  - [ ] 뷰 상태 localStorage 저장 (사용자 선호 유지)
- [ ] 정렬 기능
  - [ ] 최신순 (기본) / 오래된순 / 제목순
  - [ ] shadcn Select 컴포넌트 사용
  - [ ] URL searchParams로 정렬 상태 관리
- [ ] 빈 상태 처리
  - [ ] 경험이 0건일 때 EmptyState 표시
  - [ ] "첫 번째 경험을 등록해보세요!" 메시지
  - [ ] "경험 등록" 버튼 → `/experiences/new`로 이동
- [ ] 경험 등록 CTA 버튼
  - [ ] 우측 하단 FAB (Floating Action Button) 또는 상단 "경험 추가" 버튼
  - [ ] `+ 경험 추가` → `/experiences/new`로 이동
- [ ] 검색 기능 (선택)
  - [ ] 제목 기반 검색 input
  - [ ] 디바운스 적용 (300ms)

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| ExperiencesPage | `src/app/(main)/experiences/page.tsx` | searchParams | 경험 목록 서버 페이지 |
| ExperienceList | `src/components/experiences/experience-list.tsx` | experiences, viewMode | 경험 목록 래퍼 (그리드/리스트) |
| ViewToggle | `src/components/experiences/view-toggle.tsx` | mode, onChange | 그리드/리스트 뷰 토글 |
| SortSelect | `src/components/experiences/sort-select.tsx` | value, onChange | 정렬 선택 드롭다운 |

### 검증 방법
- 경험이 없을 때 빈 상태 UI 표시
- 경험 등록 후 목록에 새 카드 표시
- 그리드/리스트 뷰 전환 정상
- 정렬 변경 시 목록 재정렬
- 반응형: 모바일 1열, 태블릿 2열, 데스크톱 3열

### 산출물
- `src/app/(main)/experiences/page.tsx`
- `src/components/experiences/experience-list.tsx`
- `src/components/experiences/view-toggle.tsx`
- `src/components/experiences/sort-select.tsx`

---

## Step 2.2: 경험 등록 폼

### 목표
STAR 구조 기반의 경험 등록 폼을 구현한다. React Hook Form + Zod로 유효성 검증하고, 입력 가이드를 제공한다.

### 체크리스트

- [ ] 경험 등록 페이지 구현
  - [ ] `src/app/(main)/experiences/new/page.tsx` 구현
  - [ ] React Hook Form + Zod 스키마 설정
  - [ ] 폼 제출 → Go Backend API 호출 (생성된 SDK 사용)
- [ ] Zod 유효성 검증 스키마
  - [ ] `src/lib/validations/experience.ts` 작성
  - [ ] title: 필수, 2~100자
  - [ ] period_start: 선택, Date 형식
  - [ ] period_end: 선택, Date 형식, period_start 이후
  - [ ] role: 선택, 최대 50자
  - [ ] category: 선택 (동아리, 인턴, 프로젝트, 대외활동, 아르바이트, 봉사활동, 기타)
  - [ ] situation: 선택, 최대 1000자
  - [ ] task: 선택, 최대 1000자
  - [ ] action: 선택, 최대 2000자
  - [ ] result: 선택, 최대 1000자
  - [ ] raw_content: 선택, 최대 5000자
- [ ] 폼 UI 구현
  - [ ] 기본 정보 섹션
    - [ ] 제목 (Input) - placeholder: "동아리 축제 부스 운영"
    - [ ] 활동 기간 (DatePicker × 2) - 시작일, 종료일
    - [ ] 역할 (Input) - placeholder: "부스 운영 총괄"
    - [ ] 카테고리 (Select) - 7개 항목 드롭다운
  - [ ] STAR 구조 섹션
    - [ ] Situation (Textarea) - placeholder: "어떤 상황이었나요? (배경, 맥락)"
      - [ ] 글자수 카운터 표시 (현재/최대)
      - [ ] 가이드 텍스트: "언제, 어디서, 어떤 조직에서, 규모는?"
    - [ ] Task (Textarea) - placeholder: "무엇을 해야 했나요? (과제, 목표)"
      - [ ] 글자수 카운터 표시
      - [ ] 가이드 텍스트: "구체적 목표, 기대 성과, 주어진 제약 조건은?"
    - [ ] Action (Textarea) - placeholder: "어떻게 행동했나요? (구체적 행동)"
      - [ ] 글자수 카운터 표시
      - [ ] 가이드 텍스트: "어떤 전략을 세웠고, 구체적으로 무엇을 했나요? 수치를 포함하면 좋아요."
    - [ ] Result (Textarea) - placeholder: "결과는 어땠나요? (성과, 교훈)"
      - [ ] 글자수 카운터 표시
      - [ ] 가이드 텍스트: "정량적 성과 (%, 건수, 금액), 질적 변화, 배운 점은?"
  - [ ] 자유 입력 섹션 (선택)
    - [ ] raw_content (Textarea) - STAR 형식 없이 자유롭게 작성
    - [ ] "STAR 구조가 어렵다면 자유롭게 작성해도 AI가 정리해드려요" 안내
  - [ ] 폼 하단
    - [ ] "등록" 버튼 (primary) - 제출
    - [ ] "취소" 버튼 (ghost) - `/experiences`로 이동
    - [ ] 로딩 상태: 버튼 disabled + 스피너
- [ ] 경험 수정 페이지 재사용
  - [ ] `src/app/(main)/experiences/[id]/edit/page.tsx` 구현
  - [ ] 동일한 폼 컴포넌트를 재사용 (mode: 'create' | 'edit')
  - [ ] 기존 데이터 pre-fill
  - [ ] "수정" 버튼으로 변경

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| NewExperiencePage | `src/app/(main)/experiences/new/page.tsx` | - | 경험 등록 페이지 |
| EditExperiencePage | `src/app/(main)/experiences/[id]/edit/page.tsx` | params.id | 경험 수정 페이지 |
| ExperienceForm | `src/components/experiences/experience-form.tsx` | mode, defaultValues, onSubmit | 경험 입력 폼 (등록/수정 공용) |
| StarField | `src/components/experiences/star-field.tsx` | label, placeholder, guide, maxLength, field | STAR 개별 Textarea (글자수 카운터 포함) |
| CategorySelect | `src/components/experiences/category-select.tsx` | value, onChange | 카테고리 선택 드롭다운 |

### 검증 방법
- 필수 필드(제목) 미입력 시 Zod 에러 메시지 표시
- 모든 STAR 필드 입력 → 글자수 카운터 실시간 반영
- 등록 버튼 클릭 → 로딩 상태 → 성공 시 상세 페이지 이동
- 수정 페이지: 기존 데이터 pre-fill → 수정 → 저장
- placeholder 가이드 텍스트가 입력 힌트 제공

### 산출물
- `src/app/(main)/experiences/new/page.tsx`
- `src/app/(main)/experiences/[id]/edit/page.tsx`
- `src/components/experiences/experience-form.tsx`
- `src/components/experiences/star-field.tsx`
- `src/components/experiences/category-select.tsx`
- `src/lib/validations/experience.ts`

---

## Step 2.3: 경험 CRUD API (Go Backend)

### 목표
Go backend API를 사용하여 경험의 생성/조회/수정/삭제 로직을 구현한다. 프론트엔드는 생성된 SDK를 통해 API를 호출한다.

### 체크리스트

- [ ] Go backend API 엔드포인트 구현 (backend 레포지토리)
  - [ ] `POST /v1/experiences` - 경험 생성
  - [ ] `GET /v1/experiences` - 경험 목록 조회
  - [ ] `GET /v1/experiences/:id` - 경험 상세 조회
  - [ ] `PATCH /v1/experiences/:id` - 경험 수정
  - [ ] `DELETE /v1/experiences/:id` - 경험 삭제
- [ ] SDK 생성 (프론트엔드)
  - [ ] OpenAPI 스펙에서 TypeScript SDK 자동 생성
  - [ ] `@/api/generated/sdk.gen.ts`에 저장
- [ ] 프론트엔드 API 호출 구현
  - [ ] `src/lib/api/experiences.ts` 작성
  - [ ] 생성된 SDK를 사용하여 API 호출
  - [ ] React Query 훅으로 감싸기 (캐싱, 낙관적 업데이트)
- [ ] createExperience 구현
  - [ ] 프론트엔드: Zod 스키마로 입력 데이터 검증
  - [ ] SDK를 통해 `POST /v1/experiences` 호출
  - [ ] React Query cache 무효화
  - [ ] 생성된 경험 ID 반환
  - [ ] 에러 핸들링 (네트워크 에러, 인증 에러)
- [ ] getExperiences 구현
  - [ ] SDK를 통해 `GET /v1/experiences` 호출
  - [ ] 쿼리 파라미터: sort, category, weapon 필터 지원
  - [ ] React Query로 캐싱 및 자동 재조회
- [ ] getExperience 구현
  - [ ] SDK를 통해 `GET /v1/experiences/:id` 호출
  - [ ] 경험 상세 + 무기 태그 + 사용 이력 포함
  - [ ] 본인 경험이 아닌 경우 에러 처리 (Go backend에서 검증)
- [ ] updateExperience 구현
  - [ ] SDK를 통해 `PATCH /v1/experiences/:id` 호출
  - [ ] React Query cache 무효화
  - [ ] 수정된 경험 반환
- [ ] deleteExperience 구현
  - [ ] SDK를 통해 `DELETE /v1/experiences/:id` 호출
  - [ ] React Query cache 무효화
  - [ ] 삭제 전 확인 다이얼로그 (프론트엔드에서 처리)

### API 엔드포인트 (Go Backend)

| Method | Path | Request | Response |
|--------|------|---------|----------|
| POST | `/v1/experiences` | `ExperienceCreateRequest` | `{ id: string }` |
| GET | `/v1/experiences` | Query: `sort?, category?, weapon?` | `ExperienceWithWeapons[]` |
| GET | `/v1/experiences/:id` | - | `ExperienceWithWeapons \| null` |
| PATCH | `/v1/experiences/:id` | `ExperienceUpdateRequest` | `{ id: string }` |
| DELETE | `/v1/experiences/:id` | - | `{ success: boolean }` |

### 검증 방법
- 경험 등록 → DB에 레코드 생성 확인 (Supabase Dashboard)
- 목록 조회 → 현재 사용자 경험만 반환 (다른 사용자 데이터 미포함)
- 수정 → DB 레코드 업데이트 + `updated_at` 갱신 확인
- 삭제 → DB에서 레코드 삭제 + 관련 데이터 CASCADE 삭제 확인
- 에러 케이스: 존재하지 않는 ID, 다른 사용자의 경험 접근 시도

### 산출물
- Go backend: `/v1/experiences` 엔드포인트 구현 (backend 레포지토리)
- 프론트엔드: `@/api/generated/sdk.gen.ts` (자동 생성된 SDK)
- 프론트엔드: `src/lib/api/experiences.ts` (API 호출 래퍼)
- 프론트엔드: `src/hooks/use-experiences.ts` (React Query 훅)

---

## Step 2.4: 경험 상세 페이지

### 목표
경험 상세 정보를 STAR 구조로 표시하고, 무기 배지, 수정/삭제 기능을 제공한다.

### 체크리스트

- [ ] 경험 상세 페이지 구현
  - [ ] `src/app/(main)/experiences/[id]/page.tsx` 구현 (서버 컴포넌트)
  - [ ] `getExperience(id)` Go Backend API 호출 (생성된 SDK 사용)
  - [ ] 존재하지 않는 경험 → `notFound()` 호출
- [ ] 상세 정보 표시
  - [ ] 상단: 제목 + 카테고리 배지 + 기간
  - [ ] STAR 구조 섹션 (Situation → Task → Action → Result)
    - [ ] 각 섹션을 Card 컴포넌트로 구분
    - [ ] 비어있는 섹션은 "아직 작성되지 않았습니다" 표시
    - [ ] 아이콘: S(🔍), T(🎯), A(⚡), R(📊)
  - [ ] 무기 태그 표시 영역
    - [ ] 주 무기: 큰 배지 (예: 🔥 위기극복)
    - [ ] 부 무기: 작은 배지들
    - [ ] 무기가 없으면 "AI 분석 대기 중..." 또는 "무기 태깅 실행" 버튼
  - [ ] 자유 입력 내용 (raw_content가 있는 경우)
- [ ] 액션 버튼
  - [ ] "수정" 버튼 → `/experiences/[id]/edit` 이동
  - [ ] "삭제" 버튼 → 확인 다이얼로그 → deleteExperience 호출
  - [ ] 삭제 확인: shadcn AlertDialog 사용 ("정말 삭제하시겠습니까?")
  - [ ] 삭제 성공 → `/experiences`로 리디렉션 + toast 알림
- [ ] 메타 정보 표시
  - [ ] 생성일
  - [ ] 최종 수정일
  - [ ] 사용 이력 카운트 ("3개 자소서에 사용됨")
- [ ] 뒤로 가기
  - [ ] 상단에 "← 경험 목록" 브레드크럼 또는 뒤로 가기 버튼

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| ExperienceDetailPage | `src/app/(main)/experiences/[id]/page.tsx` | params.id | 경험 상세 서버 페이지 |
| StarDisplay | `src/components/experiences/star-display.tsx` | situation, task, action, result | STAR 구조 읽기 전용 표시 |
| WeaponBadges | `src/components/experiences/weapon-badges.tsx` | weapons, size | 무기 태그 배지 표시 |
| ExperienceActions | `src/components/experiences/experience-actions.tsx` | experienceId | 수정/삭제 액션 버튼 |
| DeleteDialog | `src/components/experiences/delete-dialog.tsx` | experienceId, title, onConfirm | 삭제 확인 다이얼로그 |

### 검증 방법
- `/experiences/[id]` 접속 → 해당 경험 상세 정보 표시
- STAR 각 섹션 내용 정상 렌더링
- 무기 배지 색상 + 이름 정상 표시
- "수정" 클릭 → 수정 페이지 이동 (기존 데이터 pre-fill)
- "삭제" 클릭 → AlertDialog → "삭제" 확인 → 목록 페이지 이동
- 존재하지 않는 ID → 404 페이지

### 산출물
- `src/app/(main)/experiences/[id]/page.tsx`
- `src/components/experiences/star-display.tsx`
- `src/components/experiences/weapon-badges.tsx`
- `src/components/experiences/experience-actions.tsx`
- `src/components/experiences/delete-dialog.tsx`

---

## Step 2.5: 경험 카드 컴포넌트

### 목표
경험 목록에서 재사용되는 경험 카드 컴포넌트를 구현한다. 그리드 뷰와 리스트 뷰 양쪽에서 사용 가능하도록 한다.

### 체크리스트

- [ ] 경험 카드 컴포넌트 구현
  - [ ] `src/components/experiences/experience-card.tsx` 구현
  - [ ] Props: experience (ExperienceWithWeapons), viewMode ('grid' | 'list')
  - [ ] 클릭 → `/experiences/[id]` 상세 페이지 이동
- [ ] 그리드 뷰 카드 디자인
  - [ ] shadcn Card 사용
  - [ ] 상단: 카테고리 배지 + 기간
  - [ ] 중앙: 제목 (text-lg font-semibold, 2줄 말줄임)
  - [ ] STAR 미리보기: Situation 첫 2줄 말줄임 표시
  - [ ] 하단: 무기 태그 배지 (최대 3개, 나머지 "+N" 표시)
  - [ ] hover: shadow-md 효과 + 커서 pointer
- [ ] 리스트 뷰 카드 디자인
  - [ ] 가로 방향 레이아웃 (flex-row)
  - [ ] 좌측: 카테고리 아이콘 + 제목 + 기간
  - [ ] 중앙: Situation 요약 (1줄 말줄임)
  - [ ] 우측: 무기 배지 + 생성일
- [ ] 카테고리별 아이콘/색상 매핑
  - [ ] 동아리: Users 아이콘, blue
  - [ ] 인턴: Briefcase 아이콘, green
  - [ ] 프로젝트: FolderOpen 아이콘, purple
  - [ ] 대외활동: Globe 아이콘, orange
  - [ ] 아르바이트: Clock 아이콘, yellow
  - [ ] 봉사활동: Heart 아이콘, pink
  - [ ] 기타: MoreHorizontal 아이콘, gray
- [ ] 즐겨찾기 표시
  - [ ] is_starred가 true면 카드 우상단에 ⭐ 표시
  - [ ] 클릭 시 즐겨찾기 토글 (Go Backend API 호출)

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| ExperienceCard | `src/components/experiences/experience-card.tsx` | experience, viewMode | 재사용 경험 카드 (그리드/리스트 대응) |
| CategoryIcon | `src/components/experiences/category-icon.tsx` | category | 카테고리별 아이콘+색상 |

### 검증 방법
- 그리드 뷰에서 카드 형태로 경험 표시
- 리스트 뷰에서 행 형태로 경험 표시
- 카드 클릭 → 상세 페이지 이동
- 무기 배지 색상이 W01~W07 각각 다르게 표시
- 카테고리 아이콘이 경험 유형에 맞게 표시
- 즐겨찾기 토글 동작

### 산출물
- `src/components/experiences/experience-card.tsx`
- `src/components/experiences/category-icon.tsx`

---

## Phase 완료 체크리스트

### 기능
- [ ] 경험 목록 페이지 정상 표시 (그리드/리스트 뷰)
- [ ] 경험 등록 폼 → STAR 구조 입력 → DB 저장 정상
- [ ] 경험 상세 페이지 → STAR 섹션 + 메타 정보 표시
- [ ] 경험 수정 → 기존 데이터 로드 → 수정 → 저장 정상
- [ ] 경험 삭제 → 확인 다이얼로그 → 삭제 → 목록 이동
- [ ] 정렬 (최신순/오래된순/제목순) 정상 동작
- [ ] 빈 상태 UI 정상 표시
- [ ] 경험 카드 그리드/리스트 뷰 전환 정상

### 테스트
- [ ] `pnpm run build` → 빌드 성공
- [ ] 경험 CRUD 전체 플로우 수동 테스트 (등록 → 목록확인 → 상세확인 → 수정 → 삭제)
- [ ] Zod 유효성 검증 테스트 (빈 제목, 너무 긴 텍스트)
- [ ] API 인증 테스트: 다른 사용자 경험 접근 불가 확인 (Go 미들웨어 JWT 검증)
- [ ] 반응형 테스트 (모바일/태블릿/데스크톱)

### 코드 품질
- [ ] TypeScript strict mode 에러 0건
- [ ] ESLint 경고/에러 0건
- [ ] Go Backend API에 입력 데이터 검증 적용
- [ ] 프론트엔드 Zod 스키마 검증 적용
- [ ] 모든 API 호출에 에러 핸들링
- [ ] React Query 캐시 무효화 정상 동작

---

## 다음 Phase

**→ Phase 2.1: AI 무기 자동 태깅** (Sprint 1, Day 4-5)
- AI 무기 분류 API 구현 (GPT-4.1 mini)
- 경험 등록/수정 후 자동 태깅 트리거
- 무기 태그 UI 및 수동 편집 기능
