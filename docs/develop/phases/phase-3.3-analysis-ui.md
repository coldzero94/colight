# Phase 3.3: 분석 리포트 UI

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(API Routes, Supabase 직접 쿼리, Vercel AI SDK)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, hooks)는 그대로 참고하세요.
>
> - `src/app/api/analyze/` → Go 백엔드 `internal/controller/analysis_controller.go`
> - `createServerClient()` Supabase 쿼리 → Ent ORM 쿼리 (`internal/service/`)
> - 프론트엔드에서는 생성된 SDK (`@/api/generated`)로 Go API 호출

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | AI 기업 분석 결과를 시각적으로 표시하는 리포트 UI 구축 — URL 입력, 실시간 스트리밍 표시, 분석 결과 페이지, 분석 히스토리 |
| **선행 조건** | Phase 3.2 완료 (AI 기업 분석 스트리밍 API), shadcn/ui 컴포넌트 설치 |
| **스프린트** | Sprint 2 — Day 7~8 |
| **관련 기능** | F07 (채용공고 자동 분석), F08 (기업 프로필 자동 조회), F09 (인재상 분석 + 최근 동향) |
| **예상 공수** | 2일 (16시간) |
| **산출물** | 분석 메인 페이지, 스트리밍 디스플레이, 분석 결과 페이지, 분석 히스토리 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 3.3.1 | 분석 메인 페이지 | ⬜ |
| 3.3.2 | 스트리밍 디스플레이 | ⬜ |
| 3.3.3 | 분석 결과 페이지 | ⬜ |
| 3.3.4 | 분석 히스토리 | ⬜ |

---

## Step 3.3.1: 분석 메인 페이지

### 목표

기업 분석의 진입점 페이지를 구현한다. URL 입력 영역, 최근 분석 목록, 빈 상태(empty state) UI를 포함한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/analysis/__tests__/analysis-page.test.tsx`
  - `it('renders URL input area')` — URL 입력 영역 렌더링 확인
  - `it('renders recent analysis list when data exists')` — 최근 분석 목록 정상 표시 (데이터 있을 때)
  - `it('renders empty state when no analyses')` — 빈 상태 UI 정상 표시 (데이터 없을 때)
  - `it('shows loading state on submit')` — 분석 시작 버튼 클릭 → 로딩 상태 전환
  - `it('transitions to streaming display on analysis start')` — 분석 시작 시 스트리밍 디스플레이로 전환

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/analysis/__tests__/analysis-page.test.tsx`
- [ ] 구현 (GREEN)
  - [ ] `(main)/analysis/page.tsx` 페이지 생성
  - [ ] URL 입력 영역 (Phase 3.1의 `JobUrlInput` 컴포넌트 활용)
  - [ ] 최근 분석 목록 (최대 5건, 카드 형태)
  - [ ] 빈 상태 UI (첫 분석 안내 일러스트 + CTA)
  - [ ] 분석 시작 시 스트리밍 디스플레이로 전환
  - [ ] 페이지 레이아웃 및 반응형 디자인
  - [ ] 로딩 상태 처리
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `AnalysisPage` | `src/app/(main)/analysis/page.tsx` | — | 분석 메인 페이지 (Server Component) |
| `AnalysisClient` | `src/app/(main)/analysis/analysis-client.tsx` | `recentAnalyses: Analysis[]` | 클라이언트 인터랙션 담당 |
| `JobUrlInput` | `src/components/analysis/job-url-input.tsx` | `onSubmit, isLoading, error` | URL 입력 (Phase 3에서 구현) |
| `RecentAnalysisList` | `src/components/analysis/recent-analysis-list.tsx` | `analyses: Analysis[]` | 최근 분석 카드 목록 |
| `AnalysisCard` | `src/components/analysis/analysis-card.tsx` | `analysis: Analysis` | 개별 분석 요약 카드 |
| `EmptyAnalysisState` | `src/components/analysis/empty-state.tsx` | — | 분석 없을 때 빈 상태 UI |

### 페이지 구조

```
┌─────────────────────────────────────────┐
│  기업 분석                               │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │  채용공고 URL을 입력하세요        │    │
│  │  [https://...             ] [분석] │   │
│  │  🔵 잡코리아                     │    │
│  └─────────────────────────────────┘    │
│                                         │
│  최근 분석                               │
│  ┌──────────┐ ┌──────────┐             │
│  │ 삼성전자   │ │ 카카오    │             │
│  │ SW개발    │ │ 백엔드    │             │
│  │ 2일 전    │ │ 5일 전    │             │
│  └──────────┘ └──────────┘             │
│                                         │
│  (또는 빈 상태)                          │
│  ┌─────────────────────────────────┐    │
│  │  🔍                              │    │
│  │  아직 분석한 기업이 없어요         │    │
│  │  채용공고 URL을 입력하면           │    │
│  │  AI가 기업을 분석해드려요          │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### 데이터 페칭

```typescript
// src/app/(main)/analysis/page.tsx (Server Component)

export default async function AnalysisPage() {
  const supabase = await createServerClient();
  const { data: { user } } = await supabase.auth.getUser();

  // 최근 분석 5건 조회
  const { data: recentAnalyses } = await supabase
    .from('company_analyses')
    .select('id, company_name, job_posting_data, created_at')
    .eq('user_id', user.id)
    .order('created_at', { ascending: false })
    .limit(5);

  return <AnalysisClient recentAnalyses={recentAnalyses ?? []} />;
}
```

### 산출물

- `src/app/(main)/analysis/page.tsx`
- `src/app/(main)/analysis/analysis-client.tsx`
- `src/components/analysis/recent-analysis-list.tsx`
- `src/components/analysis/analysis-card.tsx`
- `src/components/analysis/empty-state.tsx`

---

## Step 3.3.2: 스트리밍 디스플레이

### 목표

AI 분석이 진행되는 동안 각 섹션이 순차적으로 표시되는 progressive rendering UI를 구현한다. 각 섹션별 로딩 스켈레톤이 표시되며, 분석 완료 후 전체 결과가 렌더링된다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/analysis/__tests__/streaming-analysis.test.tsx`
  - `it('displays pipeline steps with correct status')` — 파이프라인 단계별 진행 표시
  - `it('shows checkmark when step completes')` — 각 단계 완료 시 체크마크 전환
  - `it('renders skeleton while section is loading')` — 각 섹션 스켈레톤 표시
  - `it('replaces skeleton with content when data arrives')` — 스켈레톤 → 실제 콘텐츠 교체
  - `it('navigates to result page on completion')` — 분석 완료 시 결과 페이지로 자동 이동
  - `it('handles cancel button click')` — 분석 취소 버튼 동작
  - `it('shows error UI with retry button on failure')` — 에러 시 에러 UI + 재시도 버튼

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/analysis/__tests__/streaming-analysis.test.tsx`
- [ ] 구현 (GREEN)
  - [ ] 파이프라인 진행 상태 표시 (단계별 체크마크)
  - [ ] 각 섹션별 로딩 스켈레톤 컴포넌트
  - [ ] 스트리밍 텍스트 progressive 렌더링
  - [ ] 분석 완료 시 결과 페이지로 자동 전환
  - [ ] 분석 취소 기능
  - [ ] 에러 상태 표시 + 재시도 버튼
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `StreamingAnalysis` | `src/components/analysis/streaming-analysis.tsx` | `url: string`, `onComplete: (id: string) => void` | 스트리밍 분석 전체 컨테이너 |
| `PipelineProgress` | `src/components/analysis/pipeline-progress.tsx` | `currentStep: PipelineStep`, `steps: StepStatus[]` | 단계별 진행률 표시 |
| `AnalysisSkeleton` | `src/components/analysis/analysis-skeleton.tsx` | `section: string` | 섹션별 로딩 스켈레톤 |
| `StreamingText` | `src/components/analysis/streaming-text.tsx` | `text: string`, `isStreaming: boolean` | 텍스트 스트리밍 애니메이션 |

### 파이프라인 진행 UI

```
┌─────────────────────────────────────────┐
│  분석 진행 중...                          │
│                                         │
│  ✅ 1. 채용공고 파싱 완료                  │
│  ✅ 2. 기업 데이터 수집 완료               │
│  🔄 3. AI 분석 진행 중... (62%)           │
│                                         │
│  ───────────────────────────────        │
│                                         │
│  📊 핵심가치                              │
│  ┌─────────────────────────────┐        │
│  │ 🔵 혁신                     │        │
│  │ 기술 혁신을 통한 고객 가치... │        │
│  └─────────────────────────────┘        │
│                                         │
│  👤 인재상                               │
│  ┌─────────────────────────────┐        │
│  │ ▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░ │  ← 스켈레톤 │
│  │ ▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░ │        │
│  └─────────────────────────────┘        │
│                                         │
│                          [분석 취소]      │
└─────────────────────────────────────────┘
```

### 스트리밍 훅 구현

```typescript
// src/hooks/use-analysis-stream.ts

import { useCallback, useState } from 'react';

interface UseAnalysisStreamOptions {
  onComplete?: (analysisId: string) => void;
  onError?: (error: string) => void;
}

interface AnalysisStreamState {
  isLoading: boolean;
  currentStep: PipelineStep | null;
  steps: StepStatus[];
  streamingText: string;
  error: string | null;
  analysisId: string | null;
}

type PipelineStep = 'parsing' | 'company_data' | 'analysis' | 'complete';

interface StepStatus {
  step: PipelineStep;
  label: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
}

function useAnalysisStream(options?: UseAnalysisStreamOptions) {
  const [state, setState] = useState<AnalysisStreamState>({
    isLoading: false,
    currentStep: null,
    steps: [
      { step: 'parsing', label: '채용공고 파싱', status: 'pending' },
      { step: 'company_data', label: '기업 데이터 수집', status: 'pending' },
      { step: 'analysis', label: 'AI 분석', status: 'pending' },
    ],
    streamingText: '',
    error: null,
    analysisId: null,
  });

  const startAnalysis = useCallback(async (url: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const response = await fetch('/api/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, mode: 'full' }),
      });

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;

        const text = decoder.decode(value);
        const events = parseSSEEvents(text);

        for (const event of events) {
          handlePipelineEvent(event, setState, options);
        }
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: '분석 중 오류가 발생했습니다.',
      }));
    }
  }, [options]);

  const cancelAnalysis = useCallback(() => {
    // AbortController로 요청 취소
    setState(prev => ({ ...prev, isLoading: false }));
  }, []);

  return { ...state, startAnalysis, cancelAnalysis };
}
```

### 산출물

- `src/components/analysis/streaming-analysis.tsx`
- `src/components/analysis/pipeline-progress.tsx`
- `src/components/analysis/analysis-skeleton.tsx`
- `src/components/analysis/streaming-text.tsx`
- `src/hooks/use-analysis-stream.ts`
- `src/components/analysis/__tests__/streaming-analysis.test.tsx`

---

## Step 3.3.3: 분석 결과 페이지

### 목표

완성된 기업 분석 결과를 보기 좋게 표시하는 상세 페이지를 구현한다. 기업정보 → 핵심가치 → 인재상 → 최근 동향 → 전략 키워드 → 피해야 할 표현 순서로 섹션을 구성한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/analysis/report/__tests__/analysis-report.test.tsx`
  - `it('renders company info section')` — 기업정보 섹션 (회사명, 업종, 재무 등) 표시
  - `it('renders core values cards (3-5)')` — 핵심가치 카드 표시
  - `it('renders talent profile with priority badges')` — 인재상 카드 (우선순위 배지 포함) 표시
  - `it('renders recent trends timeline')` — 최근 동향 타임라인 표시
  - `it('renders strategy keyword tags')` — 전략 키워드 태그 표시
  - `it('renders avoid expressions warning cards')` — 피해야 할 표현 경고 카드 표시
  - `it('scrolls to section on nav click')` — 섹션 네비게이션 클릭 → 해당 섹션으로 스크롤
  - `it('shows 404 for non-existent analysis')` — 존재하지 않는 ID → 404 페이지

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/analysis/report/__tests__/analysis-report.test.tsx`
- [ ] 구현 (GREEN)
  - [ ] `(main)/analysis/[id]/page.tsx` 페이지 생성
  - [ ] 기업 기본정보 섹션 (회사명, 업종, 설립일, 대표자, 재무 요약)
  - [ ] 핵심가치 섹션 (키워드 카드 3~5개)
  - [ ] 인재상 섹션 (특성 카드 + 우선순위 배지)
  - [ ] 최근 동향 섹션 (뉴스 타임라인)
  - [ ] 전략 키워드 섹션 (태그 클라우드)
  - [ ] 피해야 할 표현 섹션 (경고 카드)
  - [ ] 섹션 간 네비게이션 (사이드바 또는 상단 탭)
  - [ ] 공유/복사 기능
  - [ ] 분석 새로고침 버튼
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `AnalysisDetailPage` | `src/app/(main)/analysis/[id]/page.tsx` | — | 분석 결과 상세 페이지 (Server Component) |
| `AnalysisReport` | `src/components/analysis/report/analysis-report.tsx` | `analysis: FullAnalysis` | 전체 리포트 컨테이너 |
| `CompanyInfoSection` | `src/components/analysis/report/company-info-section.tsx` | `profile: CompanyProfile, financials: Financials` | 기업 기본정보 |
| `CoreValuesSection` | `src/components/analysis/report/core-values-section.tsx` | `values: CoreValue[]` | 핵심가치 카드 |
| `TalentProfileSection` | `src/components/analysis/report/talent-profile-section.tsx` | `traits: TalentTrait[]` | 인재상 카드 |
| `RecentTrendsSection` | `src/components/analysis/report/recent-trends-section.tsx` | `trends: Trend[]` | 최근 동향 타임라인 |
| `StrategyKeywordsSection` | `src/components/analysis/report/strategy-keywords-section.tsx` | `keywords: string[]` | 전략 키워드 태그 |
| `AvoidExpressionsSection` | `src/components/analysis/report/avoid-expressions-section.tsx` | `expressions: AvoidExpression[]` | 피해야 할 표현 |
| `AnalysisSectionNav` | `src/components/analysis/report/section-nav.tsx` | `sections: string[]`, `activeSection: string` | 섹션 네비게이션 |

### 페이지 레이아웃

```
┌─────────────────────────────────────────────────┐
│  ← 분석 목록    삼성전자 SW개발 분석   [새로고침] │
│                                                 │
│  ┌──────┐  ┌──────────────────────────────┐    │
│  │ 목차  │  │                              │    │
│  │      │  │  📋 기업 정보                  │    │
│  │ 기업  │  │  ┌────────────────────────┐  │    │
│  │ 정보  │  │  │ 삼성전자               │  │    │
│  │      │  │  │ 반도체/전자 | 코스피     │  │    │
│  │ 핵심  │  │  │ 대표: 경계현           │  │    │
│  │ 가치  │  │  │ 매출 302조 (+12%)      │  │    │
│  │      │  │  └────────────────────────┘  │    │
│  │ 인재상│  │                              │    │
│  │      │  │  💎 핵심가치                  │    │
│  │ 동향  │  │  ┌────┐ ┌────┐ ┌────┐     │    │
│  │      │  │  │혁신 │ │상생 │ │인재 │     │    │
│  │ 전략  │  │  │... │ │... │ │... │     │    │
│  │ 키워드│  │  └────┘ └────┘ └────┘     │    │
│  │      │  │                              │    │
│  │ 주의  │  │  👤 인재상                   │    │
│  │ 표현  │  │  ┌──────────────────────┐  │    │
│  │      │  │  │ 🔴 높음 | 도전정신     │  │    │
│  └──────┘  │  │ 끊임없이 새로운 기술... │  │    │
│            │  └──────────────────────┘  │    │
│            │                              │    │
│            │  📰 최근 동향                 │    │
│            │  • 2026.02 AI 반도체 투자     │    │
│            │  • 2026.01 파운드리 수주 확대  │    │
│            │                              │    │
│            │  🎯 전략 키워드               │    │
│            │  [AI 반도체] [시스템반도체]    │    │
│            │  [글로벌 리더십] [혁신 문화]   │    │
│            │                              │    │
│            │  ⚠️ 피해야 할 표현            │    │
│            │  ❌ "대기업이라 안정적"        │    │
│            │  → "글로벌 시장 선도" 추천     │    │
│            └──────────────────────────────┘    │
└─────────────────────────────────────────────────┘
```

### 데이터 페칭

```typescript
// src/app/(main)/analysis/[id]/page.tsx

interface Props {
  params: { id: string };
}

export default async function AnalysisDetailPage({ params }: Props) {
  const supabase = await createServerClient();
  const { data: { user } } = await supabase.auth.getUser();

  const { data: analysis } = await supabase
    .from('company_analyses')
    .select('*')
    .eq('id', params.id)
    .eq('user_id', user.id)
    .single();

  if (!analysis) notFound();

  return <AnalysisReport analysis={analysis} />;
}
```

### 산출물

- `src/app/(main)/analysis/[id]/page.tsx`
- `src/components/analysis/report/analysis-report.tsx`
- `src/components/analysis/report/company-info-section.tsx`
- `src/components/analysis/report/core-values-section.tsx`
- `src/components/analysis/report/talent-profile-section.tsx`
- `src/components/analysis/report/recent-trends-section.tsx`
- `src/components/analysis/report/strategy-keywords-section.tsx`
- `src/components/analysis/report/avoid-expressions-section.tsx`
- `src/components/analysis/report/section-nav.tsx`

---

## Step 3.3.4: 분석 히스토리

### 목표

사용자가 과거에 수행한 기업 분석 목록을 보여주는 히스토리 UI를 구현한다. 목록에서 클릭하면 해당 분석 결과 페이지로 이동한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/analysis/history/__tests__/analysis-history.test.tsx`
  - `it('renders analysis history list')` — 분석 히스토리 목록 정상 표시
  - `it('paginates results')` — 페이지네이션 동작 (다음/이전 페이지)
  - `it('filters by company name search')` — 회사명 검색 → 필터링 결과
  - `it('navigates to analysis detail on card click')` — 히스토리 카드 클릭 → 분석 결과 페이지 이동
  - `it('shows delete confirmation dialog')` — 삭제 버튼 → 확인 다이얼로그
  - `it('removes card after delete confirmation')` — 삭제 확인 → 카드 제거
  - `it('shows empty state when no history')` — 빈 상태 UI

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/analysis/history/__tests__/analysis-history.test.tsx`
- [ ] 구현 (GREEN)
  - [ ] 분석 목록 컴포넌트 (페이지네이션 포함)
  - [ ] 분석 카드 (회사명, 포지션, 분석일, 요약 정보)
  - [ ] 검색/필터 기능 (회사명 검색)
  - [ ] 분석 삭제 기능 (확인 다이얼로그 포함)
  - [ ] 빈 상태 UI
  - [ ] 정렬 옵션 (최신순, 회사명순)
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `AnalysisHistoryList` | `src/components/analysis/history/analysis-history-list.tsx` | `analyses: Analysis[]`, `onDelete: (id: string) => void` | 분석 히스토리 목록 |
| `AnalysisHistoryCard` | `src/components/analysis/history/analysis-history-card.tsx` | `analysis: Analysis` | 히스토리 개별 카드 |
| `AnalysisSearchBar` | `src/components/analysis/history/analysis-search-bar.tsx` | `onSearch: (query: string) => void` | 검색 바 |
| `DeleteAnalysisDialog` | `src/components/analysis/history/delete-analysis-dialog.tsx` | `analysisId: string`, `companyName: string`, `onConfirm: () => void` | 삭제 확인 다이얼로그 |

### 히스토리 카드 레이아웃

```
┌──────────────────────────────────────┐
│  삼성전자 - SW개발 엔지니어           │
│  📅 2026.02.09  |  📊 핵심가치 4개    │
│  🎯 전략 키워드: AI 반도체, 혁신...    │
│                          [보기] [삭제]│
└──────────────────────────────────────┘
```

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `GET` | `/api/analyze/history` | `?page=1&limit=10&search=삼성` | `{ analyses: Analysis[], total: number, page: number }` |
| `DELETE` | `/api/analyze/[id]` | — | `{ success: boolean }` |

### 히스토리 API 구현

```typescript
// src/app/api/analyze/history/route.ts

export async function GET(request: Request) {
  const user = await getAuthUser(request);
  const { searchParams } = new URL(request.url);

  const page = Number(searchParams.get('page') ?? 1);
  const limit = Number(searchParams.get('limit') ?? 10);
  const search = searchParams.get('search') ?? '';
  const offset = (page - 1) * limit;

  let query = supabase
    .from('company_analyses')
    .select('id, company_name, job_posting_data, analysis_data, created_at', { count: 'exact' })
    .eq('user_id', user.id)
    .order('created_at', { ascending: false })
    .range(offset, offset + limit - 1);

  if (search) {
    query = query.ilike('company_name', `%${search}%`);
  }

  const { data, count } = await query;

  return json({
    analyses: data ?? [],
    total: count ?? 0,
    page,
  });
}
```

### 산출물

- `src/components/analysis/history/analysis-history-list.tsx`
- `src/components/analysis/history/analysis-history-card.tsx`
- `src/components/analysis/history/analysis-search-bar.tsx`
- `src/components/analysis/history/delete-analysis-dialog.tsx`
- `src/app/api/analyze/history/route.ts`
- `src/app/api/analyze/[id]/route.ts` (DELETE)

---

## Phase 완료 체크리스트

- [ ] 분석 메인 페이지: URL 입력 + 최근 분석 + 빈 상태 렌더링
- [ ] 스트리밍 디스플레이: 파이프라인 진행 상태 + 실시간 텍스트 표시
- [ ] 분석 결과 페이지: 6개 섹션 (기업정보, 핵심가치, 인재상, 동향, 키워드, 주의표현) 렌더링
- [ ] 분석 히스토리: 목록 + 검색 + 삭제 동작
- [ ] E2E 플로우: URL 입력 → 스트리밍 분석 → 결과 페이지 표시 → 히스토리 저장
- [ ] 반응형 디자인 (모바일/태블릿/데스크톱)
- [ ] 에러 상태 처리 (네트워크 오류, 분석 실패)
- [ ] 로딩 상태 처리 (스켈레톤, 스피너)
- [ ] RLS 보호 확인 (타 사용자 데이터 접근 불가)
- [ ] Lighthouse 접근성 점수 80+ 확인
- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

---

## 다음 Phase

→ [Phase 4: 경험 매칭](./phase-4-matching.md) — 사용자의 경험 데이터와 기업 분석 결과를 매칭하여 적합도를 산출한다.
