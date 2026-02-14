# Phase 4: 경험-기업 적합도 매칭

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(API Routes, Supabase 직접 쿼리, OpenAI TS SDK)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, hooks)는 그대로 참고하세요.
>
> - `src/app/api/matching/` → Go 백엔드 `internal/controller/matching_controller.go`
> - `import OpenAI from 'openai'` → Go `openai` SDK (`internal/infrastructure/openai/`)
> - Supabase 직접 쿼리 → Ent ORM 쿼리 (`internal/service/`)
> - pgvector 쿼리 → Ent raw SQL 또는 custom predicate
> - 프론트엔드에서는 생성된 SDK (`@/api/generated`)로 Go API 호출

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 사용자의 경험 데이터와 기업 분석 결과를 AI로 매칭하여, 경험별 적합도 점수와 활용 근거를 산출하는 시스템 구축 |
| **선행 조건** | Phase 3.3 완료 (분석 리포트 UI), Phase 2 완료 (경험 CRUD), Phase 2.1 완료 (무기 태깅) |
| **스프린트** | Sprint 3 — Day 1~5 |
| **관련 기능** | F10 (내 경험 적합도 매칭) |
| **예상 공수** | 5일 (40시간) |
| **산출물** | 매칭 알고리즘, 매칭 API, 매칭 UI, 자동 매칭 트리거, 매칭 캐싱 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| | 4.1 | 매칭 알고리즘 | ⬜ | | ✅ |
| | 4.2 | 매칭 API | ✅ | | ✅ |
| | 4.3 | 매칭 UI | ✅ | | ✅ |
| | 4.4 | 자동 매칭 트리거 | ✅ | | ✅ |
| | 4.5 | 매칭 캐싱 | ✅ | | ✅ |

---

## Step 4.1: 매칭 알고리즘

### 목표

경량 모델 (Gemini/Groq)를 활용하여 사용자의 각 경험과 기업 분석 결과 사이의 적합도를 평가하는 매칭 알고리즘을 구현한다. 경험별로 직무 관련도(40%), 인재상 부합도(35%), 차별화 점수(25%)를 산출하고 종합 적합도를 계산한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/matching_service_test.go`
  - `TestMatchExperience_ReturnsThreeCategoryScores`: 단일 경험 매칭 → 3개 카테고리 점수 + 종합 점수 반환 확인
  - `TestCalculateOverallFit_WeightedAverage`: 종합 점수 = 가중 평균 계산 일치 확인 (cosine similarity with known vectors)
  - `TestMatchExperience_HighRelevance_ScoreAbove70`: 관련 높은 경험 → 70+ 점수 확인
  - `TestMatchExperience_LowRelevance_ScoreBelow30`: 무관한 경험 → 30 이하 점수 확인
  - `TestMatchAllExperiences_BatchProcessing`: 배치 매칭 (10개 경험) 정상 동작 확인
  - `TestMatchExperience_ReasoningNotEmpty`: 매칭 근거(reasoning) 구체성 확인
  - `TestMatchExperience_SchemaValidation`: 응답 스키마 검증 통과 확인
  - `TestMatchExperience_MockAIClient`: MockAIClient로 embedding API 호출 테스트

> MockAIClient 패턴 사용. testdata/ 디렉토리에 매칭 fixture 데이터 준비. cosine similarity는 known vectors로 검증.

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/matching_service_test.go` 작성
  - [ ] MockAIClient 및 testdata/ai/matching fixture 준비
- [ ] 구현 (GREEN)
  - [ ] `matching.ts` 모듈 생성
  - [ ] 매칭 프롬프트 설계 및 `prompt_templates` 등록
  - [ ] 경험별 개별 매칭 함수 구현
  - [ ] 가중 평균 종합 적합도 계산
  - [ ] 매칭 근거(reasoning) 생성
  - [ ] 배치 매칭 (여러 경험 동시 평가) 최적화
  - [ ] 매칭 결과 타입 정의 (Zod 스키마)
- [ ] 테스트 통과 확인

### 매칭 알고리즘 상세

```typescript
// src/lib/ai/matching.ts

// LLM_LIGHT_PROVIDER 환경변수로 gemini 또는 groq 선택 (공통 인터페이스)

/**
 * 단일 경험 vs 기업 분석 매칭
 *
 * 입력:
 * - experience: 사용자 경험 (STAR 구조)
 * - analysis: 기업 분석 결과 (핵심가치, 인재상, 전략키워드)
 * - jobPosting: 채용공고 구조화 데이터
 *
 * 출력:
 * - 3개 카테고리별 점수 + 종합 점수 + 근거
 */
async function matchExperience(
  experience: Experience,
  analysis: CompanyAnalysis,
  jobPosting: JobPosting
): Promise<ExperienceMatch>;

/**
 * 전체 경험 배치 매칭
 * - 사용자의 모든 경험을 한번에 매칭
 * - API 호출 최적화 (경험 5개 이하: 개별 호출, 6개 이상: 배치)
 */
async function matchAllExperiences(
  experiences: Experience[],
  analysis: CompanyAnalysis,
  jobPosting: JobPosting
): Promise<MatchingResult>;
```

### 매칭 점수 체계

```typescript
// src/lib/ai/matching-schemas.ts

import { z } from 'zod';

export const CategoryScoreSchema = z.object({
  jobRelevance: z.number().min(0).max(100),    // 직무 관련도 (40%)
  talentFit: z.number().min(0).max(100),       // 인재상 부합도 (35%)
  uniqueness: z.number().min(0).max(100),      // 차별화 점수 (25%)
});

export const ExperienceMatchSchema = z.object({
  experienceId: z.string().uuid(),
  experienceTitle: z.string(),
  scores: CategoryScoreSchema,
  overallFit: z.number().min(0).max(100),      // 가중 평균
  reasoning: z.object({
    jobRelevance: z.string(),    // 직무 관련도 평가 근거
    talentFit: z.string(),       // 인재상 부합도 평가 근거
    uniqueness: z.string(),      // 차별화 평가 근거
    summary: z.string(),         // 한줄 종합 평가
  }),
  recommendedFor: z.array(z.string()),  // 추천 활용처 (예: "도전정신 문항", "팀워크 문항")
  strengthKeywords: z.array(z.string()), // 이 경험의 강점 키워드
});

export const MatchingResultSchema = z.object({
  analysisId: z.string().uuid(),
  matchedAt: z.string().datetime(),
  experienceMatches: z.array(ExperienceMatchSchema),
  overallRecommendation: z.string(),  // 전체 종합 추천 (예: "3개 경험이 높은 적합도를 보입니다")
  gapAnalysis: z.string(),           // 부족한 역량 분석 (예: "리더십 관련 경험이 부족합니다")
  topExperiences: z.array(z.string().uuid()), // 상위 3개 경험 ID
});

export type ExperienceMatch = z.infer<typeof ExperienceMatchSchema>;
export type MatchingResult = z.infer<typeof MatchingResultSchema>;
```

### 종합 적합도 계산

```typescript
// 가중 평균 계산
function calculateOverallFit(scores: {
  jobRelevance: number;
  talentFit: number;
  uniqueness: number;
}): number {
  const weights = {
    jobRelevance: 0.40,  // 직무 관련도 40%
    talentFit: 0.35,     // 인재상 부합도 35%
    uniqueness: 0.25,    // 차별화 25%
  };
  return Math.round(
    scores.jobRelevance * weights.jobRelevance +
    scores.talentFit * weights.talentFit +
    scores.uniqueness * weights.uniqueness
  );
}
```

### 매칭 프롬프트

```sql
-- prompt_templates 시드
INSERT INTO prompt_templates (category, name, version, system_prompt, user_prompt_template, model, temperature, max_tokens)
VALUES (
  'matching',
  'experience_match',
  1,
  '당신은 취업 컨설턴트입니다. 지원자의 경험과 기업의 요구사항을 정밀하게 비교 분석합니다.
각 평가 기준에 대해 0~100점을 부여하고, 구체적인 근거를 제시합니다.
점수 기준:
- 90~100: 매우 높은 적합도 (핵심 경험)
- 70~89: 높은 적합도 (강력한 소재)
- 50~69: 보통 적합도 (활용 가능하나 보완 필요)
- 30~49: 낮은 적합도 (간접적 연관만 있음)
- 0~29: 거의 무관',
  '## 경험 정보
제목: {{experienceTitle}}
기간: {{period}}
역할: {{role}}
상황(S): {{situation}}
과제(T): {{task}}
행동(A): {{action}}
결과(R): {{result}}
역량 태그: {{tags}}

## 기업 분석 결과
회사명: {{companyName}}
포지션: {{position}}
핵심가치: {{coreValues}}
인재상: {{talentProfile}}
전략 키워드: {{strategyKeywords}}
자격요건: {{requirements}}
우대사항: {{preferred}}

## 평가 요청
위 경험을 다음 3가지 기준으로 평가해주세요:

1. **직무 관련도 (job_relevance, 0~100)**
   - 이 경험이 해당 직무의 요구사항/자격조건에 얼마나 부합하는가?

2. **인재상 부합도 (talent_fit, 0~100)**
   - 이 경험이 기업의 핵심가치/인재상에 얼마나 부합하는가?

3. **차별화 점수 (uniqueness, 0~100)**
   - 이 경험이 다른 지원자 대비 얼마나 독특하고 차별화되는가?

JSON 형식으로 응답해주세요.',
  'gemini-2.0-flash',
  0.2,
  1500
);
```

### 비용 관리

- 경량 모델 (Gemini/Groq) 사용: 경험 1건당 약 3~5원
- 경험 10개 매칭 시: 약 30~50원
- 배치 최적화 (5개 이하 경험은 하나의 프롬프트에 포함): 약 10~15원

### 산출물

- `src/lib/ai/matching.ts`
- `src/lib/ai/matching-schemas.ts`
- DB 시드: `prompt_templates` 1건 (matching/experience_match)
- `__tests__/lib/ai/matching.test.ts`

---

## Step 4.2: 매칭 API

### 목표

기업 분석 ID를 받아 해당 사용자의 모든 경험에 대한 매칭을 수행하고 결과를 반환하는 API를 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/controller/matching_controller_test.go`
  - `TestPostMatching_Success`: 정상 매칭 → 경험별 점수 + 종합 추천 반환 확인
  - `TestPostMatching_ResultSaved`: 결과가 `company_analyses.analysis_data.matching`에 저장 확인
  - `TestPostMatching_NoExperiences`: 경험 없는 사용자 → `NO_EXPERIENCES` 에러 확인
  - `TestPostMatching_AnalysisNotFound`: 존재하지 않는 분석 ID → 404 에러 확인
  - `TestGetMatching_ExistingResult`: GET `/api/matching/[analysisId]` → 기존 매칭 결과 조회 확인

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/controller/matching_controller_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] `/api/matching` POST 라우트 생성
  - [ ] 분석 데이터 + 사용자 경험 조회
  - [ ] 매칭 알고리즘 실행
  - [ ] 결과 저장 (`company_analyses.analysis_data`에 매칭 결과 병합)
  - [ ] 인증 및 권한 확인
  - [ ] 에러 처리 (분석 없음, 경험 없음)
  - [ ] Rate limiting
- [ ] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/api/matching` | `{ analysisId: string }` | `{ success: boolean, data: MatchingResult }` |
| `GET` | `/api/matching/[analysisId]` | — | `{ success: boolean, data: MatchingResult \| null }` |

### API 구현

```typescript
// src/app/api/matching/route.ts

export async function POST(request: Request) {
  const user = await getAuthUser(request);
  if (!user) return unauthorized();

  const { analysisId } = await request.json();

  // 1. 분석 데이터 조회
  const analysis = await getCompanyAnalysis(analysisId, user.id);
  if (!analysis) {
    return json({ success: false, error: '분석 데이터를 찾을 수 없습니다.' }, 404);
  }

  // 2. 사용자 경험 조회
  const experiences = await getUserExperiences(user.id);
  if (experiences.length === 0) {
    return json({
      success: false,
      error: '등록된 경험이 없습니다. 경험을 먼저 등록해주세요.',
      code: 'NO_EXPERIENCES',
    }, 400);
  }

  // 3. 매칭 실행
  const matchingResult = await matchAllExperiences(
    experiences,
    analysis.analysis_data,
    analysis.job_posting_data
  );

  // 4. 결과 저장 (analysis_data에 matching 필드 추가)
  await updateAnalysisWithMatching(analysisId, matchingResult);

  // 5. 결과 반환
  return json({ success: true, data: matchingResult });
}
```

### 매칭 결과 저장 구조

```typescript
// company_analyses.analysis_data 필드 구조 (매칭 후)
{
  // 기존 분석 결과
  "coreValues": [...],
  "talentProfile": [...],
  "recentTrends": [...],
  "strategyKeywords": [...],
  "avoidExpressions": [...],

  // 매칭 결과 (추가)
  "matching": {
    "matchedAt": "2026-02-10T12:00:00Z",
    "experienceMatches": [...],
    "overallRecommendation": "...",
    "gapAnalysis": "...",
    "topExperiences": [...]
  }
}
```

### 에러 시나리오

| 시나리오 | 코드 | 메시지 |
|----------|------|--------|
| 분석 ID 없음 | 404 | "분석 데이터를 찾을 수 없습니다" |
| 타 사용자 분석 | 404 | "분석 데이터를 찾을 수 없습니다" (RLS) |
| 경험 0건 | 400 | "등록된 경험이 없습니다" |
| AI API 에러 | 500 | "매칭 처리 중 오류가 발생했습니다" |

### 산출물

- `src/app/api/matching/route.ts`
- `src/app/api/matching/[analysisId]/route.ts`
- `__tests__/app/api/matching/route.test.ts`

---

## Step 4.3: 매칭 UI

### 목표

분석 결과 페이지 내에 경험별 매칭 결과를 표시하는 UI를 구현한다. 각 경험에 대해 적합도 % 바와 색상 코딩(green/yellow/red), 매칭 근거를 보여준다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/matching/__tests__/ExperienceMatchCard.test.tsx`
  - `it('renders experience title and overall fit score')`
  - `it('applies green color class for score >= 70')`
  - `it('applies yellow color class for score 40-69')`
  - `it('applies red color class for score < 40')`
- `src/components/matching/__tests__/MatchingSection.test.tsx`
  - `it('displays matching trigger button when no matching result')`
  - `it('sorts experiences by score descending')`
  - `it('shows no-experiences prompt when user has no experiences')`
- `src/components/matching/__tests__/FitScoreBar.test.tsx`
  - `it('renders progress bar with correct width percentage')`
  - `it('displays score label text')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/matching/__tests__/ExperienceMatchCard.test.tsx` 작성
  - [ ] `src/components/matching/__tests__/MatchingSection.test.tsx` 작성
  - [ ] `src/components/matching/__tests__/FitScoreBar.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 매칭 결과 섹션 (분석 결과 페이지 하단에 추가)
  - [ ] 경험별 매칭 카드 컴포넌트
  - [ ] 적합도 % 바 (프로그레스 바)
  - [ ] 색상 코딩 (green: 70+, yellow: 40~69, red: 0~39)
  - [ ] 카테고리별 점수 세부 표시 (확장/접기)
  - [ ] 매칭 근거 표시
  - [ ] 종합 추천 & 갭 분석 표시
  - [ ] 매칭 실행 버튼 (매칭 결과 없을 때)
  - [ ] 경험 없음 안내 + 경험 등록 CTA
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `MatchingSection` | `src/components/matching/matching-section.tsx` | `analysisId: string`, `matching?: MatchingResult` | 매칭 전체 섹션 |
| `MatchingTriggerButton` | `src/components/matching/matching-trigger-button.tsx` | `analysisId: string`, `onComplete: () => void` | 매칭 실행 버튼 |
| `ExperienceMatchCard` | `src/components/matching/experience-match-card.tsx` | `match: ExperienceMatch` | 개별 경험 매칭 카드 |
| `FitScoreBar` | `src/components/matching/fit-score-bar.tsx` | `score: number`, `label: string` | 적합도 프로그레스 바 |
| `CategoryScores` | `src/components/matching/category-scores.tsx` | `scores: CategoryScore`, `reasoning: Reasoning` | 카테고리별 세부 점수 |
| `MatchingSummary` | `src/components/matching/matching-summary.tsx` | `recommendation: string`, `gapAnalysis: string`, `topExperiences: string[]` | 종합 추천 + 갭 분석 |
| `NoExperiencesPrompt` | `src/components/matching/no-experiences-prompt.tsx` | — | 경험 없음 안내 |

### 매칭 카드 레이아웃

```
┌──────────────────────────────────────────────┐
│  🏆 매칭 결과                                 │
│                                              │
│  📊 종합 추천                                 │
│  ┌──────────────────────────────────────┐    │
│  │ 3개 경험이 높은 적합도를 보입니다.     │    │
│  │ 리더십 관련 경험이 부족합니다.         │    │
│  └──────────────────────────────────────┘    │
│                                              │
│  ── 경험별 적합도 (높은 순) ──                 │
│                                              │
│  ┌──────────────────────────────────────┐    │
│  │  데이터 분석 프로젝트 리더            │    │
│  │  ████████████████████████░░░ 85%  🟢│    │
│  │                                      │    │
│  │  ▼ 세부 점수                          │    │
│  │  직무 관련도   ████████████████░░ 90% │    │
│  │  인재상 부합도 █████████████████░ 82% │    │
│  │  차별화 점수   ████████████████░░ 80% │    │
│  │                                      │    │
│  │  💬 "데이터 분석 역량이 해당 직무의    │    │
│  │  핵심 요구사항과 높은 부합도..."        │    │
│  │                                      │    │
│  │  🎯 추천 활용: 문제해결력 문항, 전문성 문항│   │
│  └──────────────────────────────────────┘    │
│                                              │
│  ┌──────────────────────────────────────┐    │
│  │  동아리 마케팅 기획                   │    │
│  │  ██████████████░░░░░░░░░░ 55%  🟡  │    │
│  │  ...                                 │    │
│  └──────────────────────────────────────┘    │
│                                              │
│  ┌──────────────────────────────────────┐    │
│  │  편의점 아르바이트                    │    │
│  │  ████████░░░░░░░░░░░░░░░░ 28%  🔴  │    │
│  │  ...                                 │    │
│  └──────────────────────────────────────┘    │
└──────────────────────────────────────────────┘
```

### 색상 코딩 로직

```typescript
// src/lib/utils/score-color.ts

type ScoreColor = 'green' | 'yellow' | 'red';

function getScoreColor(score: number): ScoreColor {
  if (score >= 70) return 'green';
  if (score >= 40) return 'yellow';
  return 'red';
}

function getScoreLabel(score: number): string {
  if (score >= 90) return '매우 높음';
  if (score >= 70) return '높음';
  if (score >= 50) return '보통';
  if (score >= 30) return '낮음';
  return '매우 낮음';
}

// Tailwind 클래스 매핑
const colorClasses: Record<ScoreColor, string> = {
  green: 'bg-green-500 text-green-700',
  yellow: 'bg-yellow-500 text-yellow-700',
  red: 'bg-red-500 text-red-700',
};
```

### 산출물

- `src/components/matching/matching-section.tsx`
- `src/components/matching/matching-trigger-button.tsx`
- `src/components/matching/experience-match-card.tsx`
- `src/components/matching/fit-score-bar.tsx`
- `src/components/matching/category-scores.tsx`
- `src/components/matching/matching-summary.tsx`
- `src/components/matching/no-experiences-prompt.tsx`
- `src/lib/utils/score-color.ts`

---

## Step 4.4: 자동 매칭 트리거

### 목표

기업 분석이 완료된 후, 사용자에게 등록된 경험이 있으면 자동으로 매칭을 트리거한다. 분석 결과 페이지에서 매칭 섹션이 자동으로 로딩되도록 한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/hooks/__tests__/useAutoMatching.test.ts`
  - `it('triggers matching when no existing result and experiences exist')`
  - `it('skips matching when no experiences')`
  - `it('skips matching when result already exists')`
  - `it('shows loading state during matching')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/hooks/__tests__/useAutoMatching.test.ts` 작성
- [ ] 구현 (GREEN)
  - [ ] 분석 완료 후 경험 존재 여부 확인 로직
  - [ ] 자동 매칭 트리거 (경험 있으면 매칭 자동 실행)
  - [ ] 매칭 진행 중 로딩 표시
  - [ ] 자동 매칭 비활성화 옵션 (사용자 설정)
  - [ ] 분석 결과 페이지 진입 시 매칭 결과 없으면 자동 트리거
- [ ] 테스트 통과 확인

### 자동 트리거 구현

```typescript
// src/hooks/use-auto-matching.ts

interface UseAutoMatchingOptions {
  analysisId: string;
  hasMatching: boolean;         // 이미 매칭 결과 존재 여부
  experienceCount: number;      // 사용자 경험 수
  enabled?: boolean;            // 자동 매칭 활성화 여부 (기본 true)
}

function useAutoMatching(options: UseAutoMatchingOptions) {
  const { analysisId, hasMatching, experienceCount, enabled = true } = options;
  const [isMatching, setIsMatching] = useState(false);
  const [matchingResult, setMatchingResult] = useState<MatchingResult | null>(null);

  useEffect(() => {
    // 조건: 매칭 결과 없음 + 경험 1개 이상 + 자동 매칭 활성화
    if (!hasMatching && experienceCount > 0 && enabled) {
      triggerMatching();
    }
  }, [analysisId, hasMatching, experienceCount, enabled]);

  const triggerMatching = async () => {
    setIsMatching(true);
    try {
      const response = await fetch('/api/matching', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ analysisId }),
      });
      const data = await response.json();
      if (data.success) {
        setMatchingResult(data.data);
      }
    } catch (error) {
      console.error('Auto-matching failed:', error);
    } finally {
      setIsMatching(false);
    }
  };

  return { isMatching, matchingResult, triggerMatching };
}
```

### 파이프라인 통합

```typescript
// src/lib/pipeline/analyze-orchestrator.ts (확장)

// Step 3 완료 후 자동 매칭 추가
async function runAnalysisPipeline(options: PipelineOptions): Promise<void> {
  // ... 기존 Step 1~3 ...

  // Step 4: 자동 매칭 (경험 존재 시)
  const experiences = await getUserExperiences(options.userId);
  if (experiences.length > 0) {
    onEvent({ step: 'matching', status: 'started', progress: 85 });
    const matchingResult = await matchAllExperiences(experiences, analysis, jobPosting);
    await updateAnalysisWithMatching(analysisId, matchingResult);
    onEvent({ step: 'matching', status: 'completed', progress: 100, data: matchingResult });
  }

  onEvent({ step: 'complete', status: 'completed', data: { analysisId } });
}
```

### 산출물

- `src/hooks/use-auto-matching.ts`
- `src/lib/pipeline/analyze-orchestrator.ts` (확장)

---

## Step 4.5: 매칭 캐싱

### 목표

매칭 결과를 `company_analyses.analysis_data` 내에 저장하고, 경험이 추가/수정/삭제되면 매칭 결과에 "outdated" 배지를 표시하여 재매칭을 유도한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/matching_service_test.go`
  - `TestCheckMatchingOutdated_ExperienceAdded`: 매칭 후 경험 추가 → outdated 판정
  - `TestCheckMatchingOutdated_ExperienceModified`: 매칭 후 경험 수정 → outdated 판정
  - `TestCheckMatchingOutdated_ExperienceDeleted`: 매칭 후 경험 삭제 → outdated 판정
  - `TestCheckMatchingOutdated_SevenDaysExpired`: 매칭 후 7일 경과 → outdated 판정
  - `TestCheckMatchingOutdated_Fresh`: 최신 매칭 → outdated 아님

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/matching/__tests__/OutdatedMatchingBanner.test.tsx`
  - `it('displays outdated banner with reason text')`
  - `it('calls onRefresh when re-match button clicked')`
  - `it('hides banner when matching is fresh')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/matching_service_test.go`에 outdated 테스트 추가
  - [ ] `src/components/matching/__tests__/OutdatedMatchingBanner.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 매칭 결과 저장 구조 정의 (`analysis_data.matching`)
  - [ ] 매칭 결과에 `matchedAt` 타임스탬프 저장
  - [ ] 경험 변경 감지 로직 (최종 경험 수정일 vs 매칭일 비교)
  - [ ] "outdated" 배지 표시 조건 정의
  - [ ] 재매칭 버튼 + 확인 다이얼로그
  - [ ] 매칭 결과 갱신 시 기존 결과 대체
- [ ] 테스트 통과 확인

### Outdated 감지 로직

```typescript
// src/lib/matching/outdated-checker.ts

interface OutdatedCheckResult {
  isOutdated: boolean;
  reason?: string;
  lastMatchedAt: string;
  lastExperienceUpdatedAt: string;
}

/**
 * 매칭 결과가 outdated인지 확인
 *
 * Outdated 조건:
 * 1. 매칭 이후 경험이 추가됨
 * 2. 매칭 이후 경험이 수정됨
 * 3. 매칭 이후 경험이 삭제됨
 * 4. 매칭 이후 7일 이상 경과
 */
async function checkMatchingOutdated(
  analysisId: string,
  userId: string
): Promise<OutdatedCheckResult> {
  // 매칭 시점
  const analysis = await getCompanyAnalysis(analysisId, userId);
  const matchedAt = analysis.analysis_data?.matching?.matchedAt;
  if (!matchedAt) return { isOutdated: true, reason: '매칭 결과가 없습니다' };

  // 경험 최종 변경일
  const { data: latestExperience } = await supabase
    .from('experiences')
    .select('updated_at')
    .eq('user_id', userId)
    .order('updated_at', { ascending: false })
    .limit(1)
    .single();

  const lastExperienceUpdate = latestExperience?.updated_at;

  // 비교
  if (lastExperienceUpdate && new Date(lastExperienceUpdate) > new Date(matchedAt)) {
    return {
      isOutdated: true,
      reason: '매칭 이후 경험이 변경되었습니다',
      lastMatchedAt: matchedAt,
      lastExperienceUpdatedAt: lastExperienceUpdate,
    };
  }

  // 7일 경과 체크
  const daysSinceMatch = (Date.now() - new Date(matchedAt).getTime()) / (1000 * 60 * 60 * 24);
  if (daysSinceMatch > 7) {
    return {
      isOutdated: true,
      reason: '매칭 결과가 7일 이상 지났습니다',
      lastMatchedAt: matchedAt,
      lastExperienceUpdatedAt: lastExperienceUpdate ?? '',
    };
  }

  return {
    isOutdated: false,
    lastMatchedAt: matchedAt,
    lastExperienceUpdatedAt: lastExperienceUpdate ?? '',
  };
}
```

### Outdated 배지 UI

```
┌──────────────────────────────────────────┐
│  🏆 매칭 결과                             │
│  ┌────────────────────────────────┐      │
│  │ ⚠️ 경험이 변경되어 매칭 결과가   │      │
│  │ 최신이 아닐 수 있습니다.         │      │
│  │              [다시 매칭하기]     │      │
│  └────────────────────────────────┘      │
│                                          │
│  (기존 매칭 결과 표시, 약간 흐리게)       │
└──────────────────────────────────────────┘
```

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `OutdatedMatchingBanner` | `src/components/matching/outdated-matching-banner.tsx` | `reason: string`, `onRefresh: () => void` | Outdated 경고 배너 |

### 산출물

- `src/lib/matching/outdated-checker.ts`
- `src/components/matching/outdated-matching-banner.tsx`
- `__tests__/lib/matching/outdated-checker.test.ts`

---

## Phase 완료 체크리스트

- [ ] 경량 모델 (Gemini/Groq) 기반 경험-기업 매칭 알고리즘 동작
- [ ] 경험별 3개 카테고리 점수 (직무 40%, 인재상 35%, 차별화 25%) 산출
- [ ] 종합 적합도 (가중 평균) 계산 정확
- [ ] 매칭 근거(reasoning) 구체적으로 생성
- [ ] `/api/matching` POST API 정상 동작
- [ ] 매칭 UI: 경험별 카드 + 적합도 바 + 색상 코딩 표시
- [ ] 분석 완료 후 자동 매칭 트리거 동작
- [ ] 경험 변경 시 outdated 배지 표시
- [ ] 재매칭 기능 동작
- [ ] 경험 없는 사용자에 대한 적절한 안내
- [ ] 종합 추천 + 갭 분석 표시
- [ ] Zod 스키마 검증 통과
- [ ] 테스트
  - [ ] `moon run backend:test` → 전체 통과
  - [ ] `moon run web:test` → 전체 통과
  - [ ] `moon run :lint` → 경고 0건
  - [ ] `moon run web:build` → 빌드 성공
- [ ] E2E: URL 입력 → 분석 → 매칭 → 결과 표시 전체 플로우

---

## 다음 Phase

→ [Phase 5: 문항 분석](./phase-5-question-analysis.md) — 자소서 문항의 표면적 질문과 진짜 의도를 분석하고, 매칭된 경험 중 최적의 경험을 추천한다.
