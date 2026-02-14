# Phase 5: 자소서 문항 분석

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(API Routes, Supabase 직접 쿼리)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, hooks)는 그대로 참고하세요.
>
> - 문항 분석 API → Go 백엔드 `internal/controller/question_controller.go`
> - Supabase SQL 쿼리 → Ent ORM 쿼리 (`internal/service/`)
> - Claude API 호출 → Anthropic Go SDK (`internal/infrastructure/anthropic/`)
> - 프론트엔드에서는 생성된 SDK (`@/api/generated`)로 Go API 호출

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 자소서 문항의 숨은 의도를 AI로 분석하고, 필요한 역량(무기)을 식별하여 적합한 경험을 추천한다 |
| **선행 조건** | Phase 4 (경험 매칭) 완료, `experience_weapons` 데이터 존재, `coaching_sessions` 테이블 생성 완료, Claude Sonnet 4.5 API 키 설정 |
| **스프린트** | Sprint 4 |
| **관련 기능** | F11 (문항 분석), F12 (경험 추천) |
| **예상 공수** | 2일 (Day 1-2) |
| **산출물** | 문항 입력 UI, 문항 분석 API, 분석 결과 UI, 경험 추천 기능 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 5.1 | 문항 입력 UI | ✅ 완료 |
| 5.2 | 문항 분석 API | ✅ 완료 |
| 5.3 | 분석 결과 UI | ✅ 완료 |
| 5.4 | 경험 추천 | ✅ 완료 |

---

## Step 5.1: 문항 입력 UI

### 목표

코칭 메인 페이지에서 사용자가 자소서 문항을 입력하고 분석을 요청할 수 있는 폼을 구현한다. 기업 분석 이력에서 기업을 선택하고, 해당 기업의 자소서 문항을 입력하는 직관적인 흐름을 제공한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/lib/validations/__tests__/coaching.test.ts`
  - `it('validates question_text minimum length (10 chars)')`
  - `it('rejects question_text under 10 chars')`
  - `it('validates char_limit range (200-2000)')`
  - `it('rejects char_limit below 200')`
  - `it('rejects char_limit above 2000')`
  - `it('requires valid UUID for application_id')`
- `src/components/coaching/__tests__/QuestionInputForm.test.tsx`
  - `it('renders company dropdown with analysis history')`
  - `it('shows empty state when no analysis history')`
  - `it('displays validation error for short question text')`
  - `it('shows loading spinner on form submit')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/lib/validations/__tests__/coaching.test.ts` 작성
  - [ ] `src/components/coaching/__tests__/QuestionInputForm.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] `(main)/coaching/page.tsx` 페이지 생성
  - [ ] 기업 선택 드롭다운 구현 (기존 분석 이력에서 불러오기)
  - [ ] 자소서 문항 텍스트 입력 영역 구현 (textarea)
  - [ ] 글자수 제한 입력 필드 구현 (number input, 200~2000자)
  - [ ] React Hook Form + Zod 스키마 검증
  - [ ] "분석 시작" 버튼 + 로딩 상태
  - [ ] 빈 상태 UI (분석 이력이 없을 때 → 기업 분석 페이지로 안내)
  - [ ] `loading.tsx`, `error.tsx` 추가
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `CoachingPage` | `src/app/(main)/coaching/page.tsx` | - | 코칭 메인 페이지 (서버 컴포넌트, 분석 이력 fetch) |
| `QuestionInputForm` | `src/components/coaching/question-input-form.tsx` | `companies: Company[]` | 문항 입력 폼 (클라이언트 컴포넌트) |
| `CompanySelect` | `src/components/coaching/company-select.tsx` | `companies: Company[], value: string, onChange: fn` | 기업 선택 드롭다운 (shadcn Select 기반) |
| `CharLimitInput` | `src/components/coaching/char-limit-input.tsx` | `value: number, onChange: fn` | 글자수 제한 입력 (200~2000 범위 검증) |

### Zod 스키마

```typescript
// src/types/coaching.ts
import { z } from 'zod';

export const questionAnalysisSchema = z.object({
  application_id: z.string().uuid('기업을 선택해주세요'),
  question_text: z.string()
    .min(10, '문항은 최소 10자 이상 입력해주세요')
    .max(500, '문항은 500자까지 입력 가능합니다'),
  char_limit: z.number()
    .min(200, '최소 200자 이상이어야 합니다')
    .max(2000, '최대 2000자까지 가능합니다'),
});

export type QuestionAnalysisInput = z.infer<typeof questionAnalysisSchema>;
```

### UI 레이아웃

```
┌─────────────────────────────────────────────────────┐
│  AI 자소서 코칭                                        │
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ 기업 선택                                        ││
│  │ [▼ 삼성전자 - 소프트웨어 개발직 (2024.03.15)  ] ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ 자소서 문항                                      ││
│  │ ┌───────────────────────────────────────────┐   ││
│  │ │ 본인이 팀 프로젝트에서 어려움을 극복한       │   ││
│  │ │ 경험을 구체적으로 기술하세요.                 │   ││
│  │ └───────────────────────────────────────────┘   ││
│  │                                    35/500자     ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ 글자수 제한   [  800  ] 자                       ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  [ 🔍 문항 분석 시작 ]                                │
│                                                     │
│  ─── 이전 분석 이력 ──────────────────────────────── │
│  ┌──────────────┐  ┌──────────────┐                 │
│  │ 삼성전자      │  │ LG전자       │                 │
│  │ 문항 1       │  │ 문항 2       │                 │
│  │ 2024.03.15   │  │ 2024.03.14   │                 │
│  └──────────────┘  └──────────────┘                 │
└─────────────────────────────────────────────────────┘
```

### 산출물

- `src/app/(main)/coaching/page.tsx`
- `src/app/(main)/coaching/loading.tsx`
- `src/app/(main)/coaching/error.tsx`
- `src/components/coaching/question-input-form.tsx`
- `src/components/coaching/company-select.tsx`
- `src/components/coaching/char-limit-input.tsx`

---

## Step 5.2: 문항 분석 API

### 목표

자소서 문항을 Claude Sonnet 4.5로 분석하여 표면적 질문, 진짜 의도, 필요 무기, 작성 구조(글자수 배분), 핵심 키워드, 피해야 할 표현 등을 구조화된 JSON으로 반환하는 Go backend API를 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/question_service_test.go`
  - `TestAnalyzeQuestion_ReturnsStructuredResult`: 유효한 문항 입력 시 구조화된 분석 결과 반환 확인
  - `TestAnalyzeQuestion_RealIntentsAlwaysThree`: `real_intents` 배열이 항상 3개인지 확인
  - `TestAnalyzeQuestion_PrimaryWeaponExists`: `required_weapons.primary`가 반드시 1개 존재하는지 확인
  - `TestAnalyzeQuestion_CharCountSumsCorrectly`: `writing_structure.sections`의 `char_count` 합이 `total_chars`와 일치하는지 확인
  - `TestAnalyzeQuestion_CategoryMatching`: question pattern classification + category matching 확인
- `internal/controller/question_controller_test.go`
  - `TestPostQuestionAnalysis_Unauthorized`: 인증 없는 요청 시 401 반환
  - `TestPostQuestionAnalysis_InvalidApplicationId`: 잘못된 `application_id` 시 404 반환
  - `TestPostQuestionAnalysis_AIFailure`: Claude API 실패 시 적절한 에러 메시지 반환
  - `TestPostQuestionAnalysis_KeywordExtraction`: keyword extraction 정상 동작 확인

> MockAIClient 패턴 사용. testdata/ai/question-analysis fixture 데이터 준비.

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/question_service_test.go` 작성
  - [ ] `internal/controller/question_controller_test.go` 작성
  - [ ] MockAIClient 및 testdata/ai/question-analysis fixture 준비
- [ ] 구현 (GREEN)
  - [ ] Go backend API 엔드포인트 구현
  - [ ] `POST /v1/coaching/question-analysis` 엔드포인트 구현
  - [ ] 인증 확인
  - [ ] 입력 검증
  - [ ] `prompt_templates`에서 `question_analysis` 프롬프트 로드
  - [ ] `weapon_categories` 전체 목록 로드하여 프롬프트에 주입
  - [ ] `company_analyses` 결과 로드하여 기업 맥락 주입
  - [ ] Claude Sonnet 4.5 API 호출
  - [ ] 응답 JSON 스키마 검증 + 파싱
  - [ ] 에러 처리 (AI 호출 실패, 파싱 실패 등)
  - [ ] 토큰 사용량 로깅
  - [ ] 프론트엔드: 생성된 SDK를 통해 API 호출
- [ ] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/v1/coaching/question-analysis` | `{ application_id, question_text, char_limit }` | `{ surface_question, real_intents, required_weapons, writing_structure, key_keywords, avoid_list, good_structure_example }` |

### Request Body

```typescript
{
  application_id: string;     // UUID - 기업(지원) ID
  question_text: string;      // 자소서 문항 원문
  char_limit: number;         // 글자수 제한 (200~2000)
}
```

### Response Body

```typescript
{
  surface_question: string;   // 표면적으로 묻는 질문 (1줄 요약)
  real_intents: [             // 진짜 의도 3가지
    {
      intent: string;         // 의도 설명
      why: string;            // 왜 이것을 보고 싶어하는지
    }
  ];
  required_weapons: {
    primary: {                // 핵심 무기 1개
      weapon_id: string;
      weapon_name: string;
      reason: string;         // 왜 이 무기가 필요한지
    };
    secondary: [              // 보조 무기 1~2개
      {
        weapon_id: string;
        weapon_name: string;
        reason: string;
      }
    ];
  };
  writing_structure: {
    total_chars: number;      // 총 글자수
    sections: [               // 추천 구조 (STAR 기반)
      {
        name: string;         // 섹션명 (예: "상황 설정")
        char_ratio: number;   // 비율 (0.0~1.0)
        char_count: number;   // 배분 글자수
        guide: string;        // 작성 가이드
      }
    ];
  };
  key_keywords: string[];     // 핵심 키워드 5~8개
  avoid_list: string[];       // 피해야 할 표현 3~5개
  good_structure_example: string; // 좋은 구조 예시 (개요 수준)
}
```

### 프롬프트 설계

```
당신은 한국 대기업 자소서 전문가입니다.

## 기업 맥락
- 기업명: {{company_name}}
- 직무: {{position}}
- 인재상: {{talent_keywords}}
- 핵심가치: {{values_keywords}}

## 역량(무기) 카테고리
{{weapon_categories}}

## 분석 요청
자소서 문항: "{{question_text}}"
글자수 제한: {{char_limit}}자

## 지시사항
1. 이 문항이 표면적으로 묻는 것과 진짜 알고 싶어하는 의도 3가지를 분석하세요.
2. 위 역량 카테고리 중 이 문항에 가장 필요한 주 무기 1개, 부 무기 1~2개를 선택하세요.
3. {{char_limit}}자에 맞는 최적 작성 구조를 STAR 기반으로 설계하고, 각 섹션별 글자수를 배분하세요.
4. 합격 자소서에서 자주 사용되는 핵심 키워드 5~8개를 추출하세요.
5. 이 기업 문화에 맞지 않거나 감점 요인이 되는 표현 3~5개를 알려주세요.
6. 좋은 구조의 개요 예시를 작성하세요.
```

### 프론트엔드 API 호출

```typescript
// src/hooks/use-question-analysis.ts
import { useMutation } from '@tanstack/react-query';
import { api } from '@/api/generated/sdk.gen';

export function useQuestionAnalysis() {
  return useMutation({
    mutationFn: async (data: {
      application_id: string;
      question_text: string;
      char_limit: number;
    }) => {
      const response = await api.POST('/v1/coaching/question-analysis', {
        body: data,
      });

      if (response.error) {
        throw new Error(response.error.message);
      }

      return response.data;
    },
  });
}
```

### 산출물

- Go backend: `POST /v1/coaching/question-analysis` 엔드포인트 (backend 레포지토리)
- 프론트엔드: `src/hooks/use-question-analysis.ts` (React Query 훅)
- 프론트엔드: `@/api/generated/sdk.gen.ts` (SDK 업데이트)
- DB: `supabase/seed.sql` (`question_analysis` 프롬프트 템플릿 추가)

---

## Step 5.3: 분석 결과 UI

### 목표

문항 분석 API 응답을 사용자가 직관적으로 이해할 수 있는 카드 기반 UI로 표시한다. 표면적 질문 vs 진짜 의도 대비, 필요 무기 배지, 글자수 배분 구조, 핵심 키워드 등을 시각적으로 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/coaching/__tests__/AnalysisResult.test.tsx`
  - `it('renders all analysis fields from API response')`
  - `it('highlights primary weapon badge visually')`
  - `it('displays structure chart with correct ratios summing to 100%')`
- `src/components/coaching/__tests__/IntentComparison.test.tsx`
  - `it('shows surface question and three real intents')`
- `src/components/coaching/__tests__/KeywordTags.test.tsx`
  - `it('renders keyword tags and avoid-list in warning style')`
  - `it('displays avoid expressions with red/orange styling')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/coaching/__tests__/AnalysisResult.test.tsx` 작성
  - [ ] `src/components/coaching/__tests__/IntentComparison.test.tsx` 작성
  - [ ] `src/components/coaching/__tests__/KeywordTags.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 분석 결과 섹션 컴포넌트 구현
  - [ ] 의도 비교 카드 (표면적 질문 vs 진짜 의도)
  - [ ] 무기 배지 표시 (주 무기 강조 + 부 무기)
  - [ ] 작성 구조 시각화 (섹션별 비율 바 + 글자수)
  - [ ] 핵심 키워드 태그 클라우드
  - [ ] 피해야 할 표현 목록 (경고 스타일)
  - [ ] 좋은 구조 예시 접기/펼치기
  - [ ] 분석 결과 → 경험 추천 섹션 연결
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `AnalysisResult` | `src/components/coaching/analysis-result.tsx` | `analysis: QuestionAnalysisResponse` | 분석 결과 래퍼 컴포넌트 |
| `IntentComparison` | `src/components/coaching/intent-comparison.tsx` | `surface: string, intents: Intent[]` | 표면 질문 vs 진짜 의도 카드 |
| `WeaponBadges` | `src/components/coaching/weapon-badges.tsx` | `primary: Weapon, secondary: Weapon[]` | 필요 무기 배지 (주 무기 강조) |
| `StructureChart` | `src/components/coaching/structure-chart.tsx` | `sections: Section[], totalChars: number` | 글자수 배분 시각화 바 |
| `KeywordTags` | `src/components/coaching/keyword-tags.tsx` | `keywords: string[], avoidList: string[]` | 키워드 + 피해야 할 표현 |
| `StructureExample` | `src/components/coaching/structure-example.tsx` | `example: string` | 좋은 구조 예시 (접기/펼치기) |

### UI 레이아웃

```
┌─────────────────────────────────────────────────────┐
│  📊 문항 분석 결과                                    │
│                                                     │
│  ┌──────────────────┬──────────────────────────────┐│
│  │ 표면적 질문       │ 진짜 의도                     ││
│  │                  │                              ││
│  │ "팀 프로젝트에서   │ 1. 🎯 갈등 해결 능력         ││
│  │  어려움을 극복한   │    → 의견 충돌 시 어떻게      ││
│  │  경험"            │      조율하는지                ││
│  │                  │ 2. 🎯 주도적 문제 인식         ││
│  │                  │    → 문제를 먼저 발견하고      ││
│  │                  │      해결하려 했는지            ││
│  │                  │ 3. 🎯 성과 측정 역량           ││
│  │                  │    → 극복 결과를 수치로         ││
│  │                  │      말할 수 있는지             ││
│  └──────────────────┴──────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ 필요 무기                                        ││
│  │ [🏆 문제해결]  [협업]  [리더십]                    ││
│  │   주 무기        부 무기                          ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ 추천 작성 구조 (800자 기준)                       ││
│  │                                                 ││
│  │ 상황 설정  ████████░░░░░░░░░░░░░░  20% (160자)  ││
│  │ 과제 정의  ██████░░░░░░░░░░░░░░░░  15% (120자)  ││
│  │ 실행 과정  ████████████████░░░░░░  40% (320자)  ││
│  │ 성과/교훈  ██████████░░░░░░░░░░░░  25% (200자)  ││
│  │                                                 ││
│  │ 💡 실행 과정에 가장 많은 비중을!                   ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ 핵심 키워드                                      ││
│  │ [데이터 기반] [개선율] [주도적] [소통] [성과]     ││
│  │                                                 ││
│  │ ⚠️ 피해야 할 표현                                ││
│  │ • "열심히 했습니다" (추상적, 구체성 부족)          ││
│  │ • "팀원들과 잘 지냈습니다" (성과 불분명)           ││
│  │ • "많은 것을 배웠습니다" (구체적 학습 내용 필요)   ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ▼ 좋은 구조 예시 보기                               │
│                                                     │
│  [ 🚀 경험 추천 받기 ]                                │
└─────────────────────────────────────────────────────┘
```

### 산출물

- `src/components/coaching/analysis-result.tsx`
- `src/components/coaching/intent-comparison.tsx`
- `src/components/coaching/weapon-badges.tsx`
- `src/components/coaching/structure-chart.tsx`
- `src/components/coaching/keyword-tags.tsx`
- `src/components/coaching/structure-example.tsx`

---

## Step 5.4: 경험 추천

### 목표

문항 분석에서 도출된 `required_weapons`를 기반으로 사용자의 `experience_weapons` 테이블에서 매칭되는 경험을 검색하여 적합도 순으로 Top 3를 추천한다. 사용자가 경험을 선택하면 다음 단계(초안 코칭)로 넘어간다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/question_service_test.go`
  - `TestRecommendExperiences_Top3Sorted`: 적합도 점수가 높은 순서로 Top 3 반환 확인
  - `TestRecommendExperiences_PrimaryWeaponDoubleWeight`: 주 무기 매칭 가중치 2배 적용 확인
  - `TestRecommendExperiences_NoExperiences`: 경험 없을 때 빈 배열 반환 확인

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/coaching/__tests__/ExperienceRecommend.test.tsx`
  - `it('renders top 3 recommended experience cards sorted by score')`
  - `it('shows empty state with registration CTA when no experiences')`
  - `it('enables CTA button only when at least 1 experience selected')`
  - `it('limits selection to maximum 3 experiences')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/question_service_test.go`에 추천 테스트 추가
  - [ ] `src/components/coaching/__tests__/ExperienceRecommend.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 경험 추천 로직 구현 (`src/lib/ai/matching.ts` 활용)
  - [ ] `experience_weapons` 테이블에서 무기 ID 기반 쿼리
  - [ ] 주 무기 매칭 가중치 2배, 부 무기 1배 적용
  - [ ] 적합도 점수 계산 (0~100%)
  - [ ] Top 3 경험 카드 UI 구현
  - [ ] 경험 카드 선택 기능 (체크박스, 1~3개 선택)
  - [ ] "이 경험으로 초안 작성" CTA 버튼
  - [ ] 추천 경험이 없을 때 → 경험 등록 안내
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `ExperienceRecommend` | `src/components/coaching/experience-recommend.tsx` | `requiredWeapons: Weapons, onSelect: fn` | 경험 추천 래퍼 |
| `RecommendCard` | `src/components/coaching/recommend-card.tsx` | `experience: Experience, matchScore: number, selected: boolean, onToggle: fn` | 추천 경험 카드 (선택 가능) |

### 추천 알고리즘

```typescript
// 경험 매칭 점수 계산
function calculateMatchScore(
  experience: ExperienceWithWeapons,
  requiredWeapons: RequiredWeapons
): number {
  let score = 0;
  const maxScore = 100;

  // 주 무기 매칭 (50점)
  const primaryMatch = experience.weapons.find(
    w => w.weapon_category_id === requiredWeapons.primary.weapon_id
  );
  if (primaryMatch) {
    score += 50 * primaryMatch.relevance_score;
  }

  // 부 무기 매칭 (각 25점)
  for (const secondary of requiredWeapons.secondary) {
    const match = experience.weapons.find(
      w => w.weapon_category_id === secondary.weapon_id
    );
    if (match) {
      score += 25 * match.relevance_score;
    }
  }

  return Math.min(Math.round(score), maxScore);
}
```

### Supabase 쿼리

```sql
-- 필요 무기에 매칭되는 경험 조회
SELECT
  e.*,
  ew.weapon_category_id,
  ew.relevance_score,
  wc.name AS weapon_name
FROM experiences e
JOIN experience_weapons ew ON ew.experience_id = e.id
JOIN weapon_categories wc ON wc.id = ew.weapon_category_id
WHERE e.user_id = $1
  AND ew.weapon_category_id IN ($2, $3, $4)  -- primary + secondary IDs
ORDER BY ew.relevance_score DESC;
```

### UI 레이아웃

```
┌─────────────────────────────────────────────────────┐
│  🎯 추천 경험 (3건)                                  │
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ ☑ 캡스톤 프로젝트 팀장 경험           적합도 92% ││
│  │   2023.03 ~ 2023.12                            ││
│  │   [🏆 문제해결] [협업]                           ││
│  │   "팀원 간 기술 수준 차이로 발생한 병목을..."     ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ ☐ 해커톤 우승 경험                    적합도 78% ││
│  │   2023.07                                      ││
│  │   [문제해결] [창의성]                            ││
│  │   "36시간 동안 새로운 아이디어를..."              ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  ┌─────────────────────────────────────────────────┐│
│  │ ☐ 인턴십 프로젝트 경험                 적합도 65% ││
│  │   2024.01 ~ 2024.02                            ││
│  │   [실행력] [문제해결]                            ││
│  │   "데이터 파이프라인 오류를..."                   ││
│  └─────────────────────────────────────────────────┘│
│                                                     │
│  [ 🚀 선택한 경험으로 초안 작성 (1개 선택됨) ]        │
└─────────────────────────────────────────────────────┘
```

### 산출물

- `src/components/coaching/experience-recommend.tsx`
- `src/components/coaching/recommend-card.tsx`
- `src/lib/ai/matching.ts` (매칭 점수 계산 함수 추가)

---

## Phase 완료 체크리스트

- [ ] 코칭 메인 페이지에서 기업 선택 → 문항 입력 → 분석 요청 전체 흐름 동작
- [ ] 분석 결과가 구조화된 카드 UI로 정상 표시
- [ ] 경험 추천이 무기 매칭 기반으로 적합도 순 정렬
- [ ] 1~3개 경험 선택 후 "초안 작성" 페이지로 이동 가능
- [ ] 인증 없는 접근 시 로그인 페이지로 리다이렉트
- [ ] 에러 상태 (API 실패, 빈 데이터 등) 정상 처리
- [ ] 모바일 반응형 확인 (375px)
- [ ] Claude API 비용: ~65원/건 이내 확인
- [ ] 테스트
  - [ ] `moon run backend:test` → 전체 통과
  - [ ] `moon run web:test` → 전체 통과
  - [ ] `moon run :lint` → 경고 0건
  - [ ] `moon run web:build` → 빌드 성공

---

## 다음 Phase

**[Phase 5.1: 초안 코칭](./phase-5.1-draft-coaching.md)** — 선택한 경험을 바탕으로 AI가 STAR 구조 초안을 스트리밍 생성
