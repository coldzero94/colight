# Phase 6: 첨삭 코칭

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(API Routes, Supabase 직접 쿼리, Vercel AI SDK)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, hooks, Recharts)는 그대로 참고하세요.
>
> - `createClient` from `@/lib/supabase/server` → Go 백엔드 API 호출 (생성된 SDK 사용)
> - `generateObject` from `ai` (Vercel AI SDK) → Anthropic Go SDK (`internal/infrastructure/anthropic/`)
> - `anthropic` from `@ai-sdk/anthropic` → Go 네이티브 Anthropic SDK
> - 첨삭 API → Go 백엔드 `internal/controller/review_controller.go`
> - Supabase 직접 쿼리 → Ent ORM 쿼리 (`internal/service/`)

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 작성된 자소서를 AI가 4개 차원(구체성, 직무적합성, 기업맞춤도, 진정성)으로 평가하고, 구체적인 개선 제안을 제공한다 |
| **선행 조건** | Phase 5.2 (코칭 에디터) 완료, `cover_letters` + `cover_letter_versions`에 편집된 자소서 존재, Claude Sonnet 4.5 API 키 설정 |
| **스프린트** | Sprint 5 |
| **관련 기능** | F14 (첨삭 코칭) |
| **예상 공수** | 2일 (Day 1-2) |
| **산출물** | 첨삭 API, 4축 레이더 차트, 차원별 피드백 UI, 라인별 제안, 반복 코칭 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 6.1 | 첨삭 API | ✅ 완료 |
| 6.2 | 결과 UI | ✅ 완료 |
| 6.3 | 반복 코칭 | ✅ 완료 |

---

## Step 6.1: 첨삭 API

### 목표

자소서 전문을 Claude Sonnet 4.5에 입력하여 4개 차원별 0~100점 평가, 차원별 좋은점/개선점 피드백, 라인 단위 구체적 수정 제안을 포함하는 구조화된 첨삭 결과를 반환하는 API를 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `TestReviewService_ParseAIResponse` | `internal/service/review_service_test.go` | AI 응답 JSON 파싱 → 4개 차원 점수 + 피드백 구조체 변환 |
| `TestReviewService_ValidateScores` | `internal/service/review_service_test.go` | 각 scores 값이 0~100 범위, overall이 4개 평균 |
| `TestReviewController_Unauthorized` | `internal/controller/review_controller_test.go` | 인증 없는 요청 시 401 반환 |
| `TestReviewController_SuccessFlow` | `internal/controller/review_controller_test.go` | MockAIClient 사용, 유효 입력 → scores + feedback + suggestions 반환, coaching_sessions 저장 확인 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `internal/service/review_service_test.go` 작성
  - [x] `internal/controller/review_controller_test.go` 작성
- [x] 구현 (GREEN)
  - [x] `POST /v1/coaching/review` 엔드포인트 구현
  - [x] 인증 확인 (JWT middleware)
  - [x] 입력 검증 (cover_letter_id, content min=50)
  - [x] `prompt_templates`에서 `coaching/review` 프롬프트 로드
  - [x] 기업 분석 결과 + 문항 분석 결과 로드 (맥락 주입)
  - [x] Claude Sonnet 4.5 호출 (LLMProvider.Call)
  - [x] 응답 JSON 스키마 검증 + 파싱
  - [x] `coaching_sessions` 저장 (session_type = 'review')
  - [x] `cover_letter_versions.feedback` 컬럼에 결과 저장
  - [x] 토큰 사용량 로깅 (input_tokens, output_tokens)
  - [x] 에러 처리 (401/400/403/404/500)
- [x] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/api/coaching/review` | `{ cover_letter_id, content }` | `{ scores, overall, per_dimension_feedback, specific_suggestions }` |

### Request Body

```typescript
{
  cover_letter_id: string;   // UUID - 자소서 ID
  content: string;           // 현재 자소서 내용 (에디터에서)
}
```

### Response Body

```typescript
{
  scores: {
    specificity: number;       // 구체성 (0~100)
    job_fit: number;           // 직무적합성 (0~100)
    company_fit: number;       // 기업맞춤도 (0~100)
    authenticity: number;      // 진정성 (0~100)
  };
  overall: number;             // 종합 점수 (4개 평균, 0~100)
  per_dimension_feedback: [
    {
      dimension: string;       // 차원명 (예: "구체성")
      score: number;           // 해당 차원 점수
      good: string[];          // 잘한 점 (1~3개)
      improve: string[];       // 개선할 점 (1~3개)
    }
  ];
  specific_suggestions: [
    {
      line_ref: string;        // 원문 참조 구간 (예: "3번째 문장")
      original: string;        // 원문 발췌
      suggested: string;       // 수정 제안
      reason: string;          // 수정 이유
      dimension: string;       // 관련 차원 (specificity/job_fit/company_fit/authenticity)
    }
  ];
}
```

### 프롬프트 설계

```
당신은 한국 대기업 자소서 첨삭 전문가입니다.
다음 자소서를 4개 차원으로 평가하고 구체적인 개선 제안을 해주세요.

## 평가 기준

### 1. 구체성 (Specificity) - 0~100점
- 추상적 표현 대신 구체적 수치, 사례, 상황이 포함되어 있는가?
- "열심히 했다" → "3주간 매일 2시간씩 추가 학습하여" 수준의 구체성
- 행동의 과정이 단계별로 서술되어 있는가?
- 결과가 측정 가능한 형태로 제시되어 있는가?

### 2. 직무적합성 (Job Fit) - 0~100점
- 지원 직무({{position}})에 필요한 역량을 보여주는가?
- 직무 관련 키워드({{job_keywords}})가 자연스럽게 포함되어 있는가?
- 경험이 지원 직무와 연결되는 논리가 명확한가?

### 3. 기업맞춤도 (Company Fit) - 0~100점
- 기업 인재상({{talent_keywords}})에 부합하는 태도/행동이 드러나는가?
- 기업 핵심가치({{values_keywords}})가 자연스럽게 녹아 있는가?
- 기업명을 바꿔도 통하는 범용적 내용이 아닌, 이 기업에 맞춤화되어 있는가?

### 4. 진정성 (Authenticity) - 0~100점
- 실제 경험에 기반한 것처럼 느껴지는가? (날짜, 장소, 인물 등 디테일)
- AI가 작성한 것 같은 패턴(과도한 수사, 비인간적 완벽함)이 없는가?
- 지원자만의 고유한 시각/성찰이 드러나는가?
- 감정이나 고민이 자연스럽게 녹아 있는가?

## 기업 맥락
- 기업명: {{company_name}}
- 직무: {{position}}
- 인재상: {{talent_keywords}}
- 핵심가치: {{values_keywords}}

## 자소서 문항
"{{question_text}}"

## 자소서 전문
{{content}}

## 출력 형식
JSON으로 응답해주세요. (스키마 설명 ...)
```

### 구현 코드 구조

```typescript
// src/app/api/coaching/review/route.ts
import { createClient } from '@/lib/supabase/server';
import { anthropic } from '@/lib/ai/providers';
import { loadPrompt } from '@/lib/ai/prompts';
import { generateObject } from 'ai';

export async function POST(request: Request) {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return Response.json({ error: 'Unauthorized' }, { status: 401 });

  const body = await request.json();

  // cover_letter + application + analysis 로드
  const { data: coverLetter } = await supabase
    .from('cover_letters')
    .select(`
      *,
      applications!inner (
        company_name, position,
        company_analyses (result)
      )
    `)
    .eq('id', body.cover_letter_id)
    .single();

  const analysis = coverLetter.applications.company_analyses[0]?.result;

  // 프롬프트 로드 + 변수 치환
  const prompt = await loadPrompt('coaching_review', {
    company_name: coverLetter.applications.company_name,
    position: coverLetter.applications.position,
    talent_keywords: analysis?.talent_keywords?.join(', ') || '',
    values_keywords: analysis?.values_keywords?.join(', ') || '',
    job_keywords: analysis?.job_keywords?.join(', ') || '',
    question_text: coverLetter.question_text,
    content: body.content,
  });

  // Claude Sonnet 4.5 호출
  const result = await generateObject({
    model: anthropic('claude-sonnet-4-5-20250929'),
    prompt,
    schema: reviewResponseSchema,
  });

  // 결과 저장 (cover_letter_versions.feedback)
  await supabase
    .from('cover_letter_versions')
    .update({ feedback: result.object })
    .eq('cover_letter_id', body.cover_letter_id)
    .order('version_number', { ascending: false })
    .limit(1);

  // 코칭 세션 저장
  await supabase.from('coaching_sessions').insert({
    cover_letter_id: body.cover_letter_id,
    session_type: 'review',
    messages: [
      { role: 'system', content: prompt },
      { role: 'assistant', content: JSON.stringify(result.object) },
    ],
  });

  return Response.json(result.object);
}
```

### 산출물

- `src/app/api/coaching/review/route.ts`
- `src/lib/ai/coaching.ts` (첨삭 로직 추가)
- `supabase/seed.sql` (`coaching_review` 프롬프트 템플릿 추가)

---

## Step 6.2: 결과 UI

### 목표

첨삭 결과를 4축 레이더 차트, 차원별 확장 카드, 에디터 내 라인별 수정 제안 하이라이트로 직관적으로 표시한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('ScoreRadarChart')` | `src/components/coaching/__tests__/score-radar-chart.test.tsx` | 4개 차원 점수 렌더링, 이전 점수 오버레이 표시 |
| `describe('DimensionCard')` | `src/components/coaching/__tests__/dimension-card.test.tsx` | 접기/펼치기 동작, 좋은점/개선점 표시 |
| `describe('SuggestionList')` | `src/components/coaching/__tests__/suggestion-list.test.tsx` | 수정 제안 목록 렌더링, "적용" 콜백 호출 |
| `describe('OverallScore')` | `src/components/coaching/__tests__/overall-score.test.tsx` | 종합 점수 표시, 등급 라벨 매핑 (S/A/B/C/D) |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/coaching/__tests__/score-radar-chart.test.tsx` — 스킵 (Recharts JSDOM 호환 불가, lazy-load로 커버)
  - [x] `src/components/coaching/__tests__/DimensionCard.test.tsx` 작성
  - [x] `src/components/coaching/__tests__/SuggestionList.test.tsx` 작성
  - [x] `src/components/coaching/__tests__/OverallScore.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 4축 레이더 차트 구현 (Recharts RadarChart, next/dynamic lazy-load)
  - [x] 종합 점수 표시 (큰 숫자 + 등급 라벨 S/A/B/C/D)
  - [x] 차원별 피드백 카드 (접기/펼치기)
  - [x] 각 카드에 점수 + 좋은점 + 개선점 표시
  - [x] 라인별 수정 제안 목록 구현
  - [x] "수정 적용" 버튼 (제안된 텍스트로 자동 교체)
  - [x] 첨삭 결과 패널 (Radix Tabs 탭 사이드바로 에디터 통합)
- [x] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `ReviewResult` | `src/components/coaching/review-result.tsx` | `review: ReviewResponse` | 첨삭 결과 래퍼 |
| `ScoreRadarChart` | `src/components/coaching/score-radar-chart.tsx` | `scores: Scores` | 4축 레이더 차트 (Recharts) |
| `OverallScore` | `src/components/coaching/overall-score.tsx` | `score: number` | 종합 점수 표시 (숫자 + 등급) |
| `DimensionCard` | `src/components/coaching/dimension-card.tsx` | `feedback: DimensionFeedback` | 차원별 피드백 카드 (접기/펼치기) |
| `SuggestionList` | `src/components/coaching/suggestion-list.tsx` | `suggestions: Suggestion[], onApply: fn` | 라인별 수정 제안 목록 |
| `SuggestionItem` | `src/components/coaching/suggestion-item.tsx` | `suggestion: Suggestion, onApply: fn` | 개별 수정 제안 (원문 → 수정안) |

### 레이더 차트 구현

```typescript
// src/components/coaching/score-radar-chart.tsx
'use client';

import {
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
  ResponsiveContainer,
} from 'recharts';

interface ScoreRadarChartProps {
  scores: {
    specificity: number;
    job_fit: number;
    company_fit: number;
    authenticity: number;
  };
  previousScores?: typeof scores;  // 이전 점수 (비교용)
}

export function ScoreRadarChart({ scores, previousScores }: ScoreRadarChartProps) {
  const data = [
    { dimension: '구체성', score: scores.specificity, prev: previousScores?.specificity },
    { dimension: '직무적합', score: scores.job_fit, prev: previousScores?.job_fit },
    { dimension: '기업맞춤', score: scores.company_fit, prev: previousScores?.company_fit },
    { dimension: '진정성', score: scores.authenticity, prev: previousScores?.authenticity },
  ];

  return (
    <ResponsiveContainer width="100%" height={300}>
      <RadarChart data={data}>
        <PolarGrid />
        <PolarAngleAxis dataKey="dimension" />
        <PolarRadiusAxis angle={90} domain={[0, 100]} />
        {previousScores && (
          <Radar
            name="이전"
            dataKey="prev"
            stroke="#94a3b8"
            fill="#94a3b8"
            fillOpacity={0.1}
            strokeDasharray="5 5"
          />
        )}
        <Radar
          name="현재"
          dataKey="score"
          stroke="#3b82f6"
          fill="#3b82f6"
          fillOpacity={0.2}
        />
      </RadarChart>
    </ResponsiveContainer>
  );
}
```

### 등급 기준

| 점수 | 등급 | 색상 | 라벨 |
|------|------|------|------|
| 90~100 | S | `text-green-600` | "합격 수준" |
| 75~89 | A | `text-blue-600` | "우수" |
| 60~74 | B | `text-amber-600` | "양호 (개선 여지 있음)" |
| 40~59 | C | `text-orange-600` | "보완 필요" |
| 0~39 | D | `text-red-600` | "대폭 수정 필요" |

### UI 레이아웃

```
┌──────────────────────────────────────────────────────────────────┐
│  📝 첨삭 결과                                                     │
│                                                                  │
│  ┌─────────────────────┬────────────────────────────────────────┐│
│  │                     │                                        ││
│  │   [레이더 차트]      │  종합 점수                              ││
│  │                     │                                        ││
│  │    구체성            │      72점                              ││
│  │   78 ╱╲             │      등급: B (양호)                    ││
│  │     ╱    ╲  직무적합 │                                        ││
│  │    ╱  ● ●  ╲ 68    │  구체성: 78  ████████░░ +12↑           ││
│  │   ╱        ╲        │  직무적합: 68  ███████░░░               ││
│  │  진정성 70   기업 72 │  기업맞춤: 72  ███████░░░               ││
│  │                     │  진정성: 70  ███████░░░                 ││
│  └─────────────────────┴────────────────────────────────────────┘│
│                                                                  │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │ ▼ 구체성 (78점)                                    ████████ ││
│  │                                                              ││
│  │ ✅ 잘한 점                                                   ││
│  │ • 프로젝트 기간(3주)과 팀 규모(4명)를 구체적으로 명시          ││
│  │ • 문제 해결 과정을 단계별로 서술                               ││
│  │                                                              ││
│  │ 🔧 개선할 점                                                  ││
│  │ • 결과 수치가 부족 → "매출 20% 증가" 같은 정량적 성과 추가     ││
│  │ • 본인의 구체적 행동과 팀원의 행동을 더 명확히 구분             ││
│  └──────────────────────────────────────────────────────────────┘│
│  ▶ 직무적합성 (68점) ...                                         │
│  ▶ 기업맞춤도 (72점) ...                                         │
│  ▶ 진정성 (70점) ...                                             │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │ 💡 구체적 수정 제안 (5건)                                     ││
│  │                                                              ││
│  │ ┌────────────────────────────────────────────────────────┐   ││
│  │ │ 📍 3번째 문장 (구체성)                                  │   ││
│  │ │ 원문: "많은 노력을 기울여 문제를 해결했습니다"            │   ││
│  │ │ 제안: "매일 2시간씩 추가 학습하며 API 설계를 3차례        │   ││
│  │ │       수정하여 응답 시간을 40% 개선했습니다"              │   ││
│  │ │ 이유: 추상적 표현을 구체적 수치와 행동으로 교체           │   ││
│  │ │                                         [✅ 적용]        │   ││
│  │ └────────────────────────────────────────────────────────┘   ││
│  │                                                              ││
│  │ ┌────────────────────────────────────────────────────────┐   ││
│  │ │ 📍 5번째 문장 (기업맞춤)                                │   ││
│  │ │ 원문: "이 경험을 통해 성장했습니다"                      │   ││
│  │ │ 제안: "이 경험은 삼성전자가 추구하는 '도전정신'과         │   ││
│  │ │       맞닿아 있으며..."                                  │   ││
│  │ │ 이유: 기업 인재상과 직접 연결하여 맞춤도 향상             │   ││
│  │ │                                         [✅ 적용]        │   ││
│  │ └────────────────────────────────────────────────────────┘   ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  [ 🔄 수정 후 재첨삭 ]                                            │
└──────────────────────────────────────────────────────────────────┘
```

### Recharts Lazy Loading

```typescript
// Recharts 번들이 크므로 lazy load
import dynamic from 'next/dynamic';

const ScoreRadarChart = dynamic(
  () => import('@/components/coaching/score-radar-chart').then(mod => mod.ScoreRadarChart),
  {
    loading: () => <div className="h-[300px] animate-pulse bg-gray-100 rounded" />,
    ssr: false,
  }
);
```

### 산출물

- `src/components/coaching/review-result.tsx`
- `src/components/coaching/score-radar-chart.tsx`
- `src/components/coaching/overall-score.tsx`
- `src/components/coaching/dimension-card.tsx`
- `src/components/coaching/suggestion-list.tsx`
- `src/components/coaching/suggestion-item.tsx`

---

## Step 6.3: 반복 코칭

### 목표

수정 후 재첨삭을 요청하여 이전 점수와 현재 점수를 비교하는 반복 코칭 루프를 구현한다. 사용자가 개선 과정을 시각적으로 확인할 수 있도록 점수 변화를 표시한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('ScoreComparison')` | `src/components/coaching/__tests__/score-comparison.test.tsx` | 점수 변화 표시 (상승/하락/동일 아이콘), diff 계산 정확성 |
| `describe('ReviewTimeline')` | `src/components/coaching/__tests__/review-timeline.test.tsx` | 첨삭 이력 타임라인 렌더링, 버전별 점수 표시 |

### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/coaching/__tests__/ScoreComparison.test.tsx` 작성
  - [x] `src/components/coaching/__tests__/ReviewTimeline.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] "첨삭 요청" 버튼으로 재첨삭 가능 (에디터 페이지)
  - [x] 이전 첨삭 점수 상태 저장 (previousScores)
  - [x] 재첨삭 시 이전 점수와 현재 점수 비교 표시 (ScoreComparison)
  - [x] 레이더 차트에 이전/현재 점수 오버레이
  - [x] 차원별 점수 변화 표시 (+N green / -N red / - gray)
  - [x] 첨삭 이력 타임라인 (1차 첨삭 → 2차 첨삭 → ...)
  - [ ] 최대 5회 첨삭 제한 → Phase 6.1 (프리미엄)으로 이관
- [x] 테스트 통과 확인

### 점수 비교 표시

```typescript
// 점수 변화 표시 로직
function ScoreChange({ current, previous }: { current: number; previous?: number }) {
  if (!previous) return null;

  const diff = current - previous;
  if (diff > 0) return <span className="text-green-600">+{diff} ↑</span>;
  if (diff < 0) return <span className="text-red-600">{diff} ↓</span>;
  return <span className="text-gray-400">→ 동일</span>;
}
```

### 첨삭 이력 타임라인

```
┌────────────────────────────────────────────────────────┐
│  📊 첨삭 이력                                           │
│                                                        │
│  v1 초안        → v2 1차 첨삭      → v3 2차 첨삭       │
│  (14:30)          (15:12)             (16:05)          │
│                                                        │
│  종합 52점       종합 68점 (+16)     종합 78점 (+10)    │
│  ●───────────────●───────────────────●                 │
│                                                        │
│  구체성 45→72(+27) 직무 60→68(+8) 기업 55→78(+23)      │
└────────────────────────────────────────────────────────┘
```

### 산출물

- `src/components/coaching/score-comparison.tsx`
- `src/components/coaching/review-timeline.tsx`
- 에디터 페이지에 "재첨삭" 버튼 통합

---

## Phase 완료 체크리스트

- [x] 에디터에서 "첨삭 요청" → API 호출 → 결과 표시 전체 흐름 동작
- [x] 4축 레이더 차트 (Recharts) 정상 렌더링
- [x] 종합 점수 + 등급 라벨 정상 표시
- [x] 차원별 피드백 카드 접기/펼치기 동작
- [x] 라인별 수정 제안 → "적용" → 에디터 내용 자동 교체
- [x] 재첨삭 시 이전/현재 점수 비교 표시
- [x] 첨삭 이력 타임라인 정상 표시
- [x] `coaching_sessions` (review) + `cover_letter_versions.feedback` 저장 확인
- [ ] Claude API 비용: ~65원/건 이내 확인 (런타임 검증 필요)
- [x] Recharts lazy load 적용
- [x] `moon run backend:test` → 전체 통과
- [x] `moon run web:test` → 전체 통과 (63 files, 275 tests)
- [x] `moon run :lint` → 경고 0건
- [x] `moon run web:build` → 빌드 성공

---

## 다음 Phase

**[Phase 6.1: 프리미엄 & 마무리](./phase-6.1-freemium-polish.md)** — 사용량 제한, 페이월, 에러 처리, 반응형, 성능 최적화
