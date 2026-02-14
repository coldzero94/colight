# Phase 5.1: AI 초안 코칭

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(API Routes, Supabase 직접 쿼리, Vercel AI SDK)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, hooks)는 그대로 참고하세요.
>
> - `src/app/api/coaching/draft/route.ts` → Go 백엔드 `internal/controller/coaching_controller.go`
> - `useCompletion` (Vercel AI SDK) → 커스텀 SSE hook (`useSSE` 또는 `EventSource`)
> - Claude API 호출 → Anthropic Go SDK + Gin `c.SSEvent()` 스트리밍
> - `supabase/seed.sql` → Go 시드 스크립트 (`scripts/seed.go`)
> - 프론트엔드에서는 생성된 SDK (`@/api/generated`)로 Go API 호출

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 사용자가 선택한 경험과 문항 분석 결과를 기반으로 Claude Sonnet 4.5가 STAR 구조 초안을 스트리밍 생성한다 |
| **선행 조건** | Phase 5 (문항 분석) 완료, 문항 분석 결과 + 경험 선택 완료, `coaching_sessions` 테이블 존재 |
| **스프린트** | Sprint 4 |
| **관련 기능** | F13 (초안 코칭) |
| **예상 공수** | 2일 (Day 3-4) |
| **산출물** | 초안 생성 API (스트리밍), 경험 선택 UI, 스트리밍 표시 UI, 세션 기록 기능 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 5.1.1 | 초안 생성 API | ✅ 완료 |
| 5.1.2 | 경험 선택 UI | ✅ 완료 |
| 5.1.3 | 스트리밍 표시 | ✅ 완료 |
| 5.1.4 | 세션 기록 | ✅ 완료 |

---

## Step 5.1.1: 초안 생성 API

### 목표

사용자가 선택한 경험(1~3개)과 문항 분석 결과, 기업 분석 데이터를 조합하여 Claude Sonnet 4.5로 STAR 구조 자소서 초안을 스트리밍 생성하는 Go backend API를 구현한다. SSE를 통해 실시간으로 응답을 반환한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/coaching_service_test.go`
  - `TestGenerateDraft_CoachingPromptComposition`: 코칭 프롬프트가 경험 STAR 데이터 + 기업 분석 + 문항 분석을 올바르게 조합하는지 확인
  - `TestGenerateDraft_StreamingResponse`: MockAIClient로 스트리밍 응답 정상 수신 확인
  - `TestGenerateDraft_STARStructure`: 생성된 초안에 STAR 태그 ([상황], [과제], [행동], [결과]) 포함 확인
  - `TestGenerateDraft_CharLimitRespected`: 생성된 초안이 `char_limit` 이내인지 확인
  - `TestGenerateDraft_ErrorHandling`: AI 실패 시 적절한 에러 반환 확인
- `internal/controller/coaching_controller_test.go`
  - `TestPostDraft_Unauthorized`: 인증 없는 요청 시 401 반환
  - `TestPostDraft_StreamingHeaders`: SSE 스트리밍 응답 헤더(Content-Type: text/event-stream) 확인
  - `TestPostDraft_SavesOnCompletion`: 스트리밍 완료 후 `cover_letters` + `cover_letter_versions` 레코드 생성 확인
  - `TestPostDraft_GracefulDisconnect`: 클라이언트 연결 끊김 시 graceful 처리 확인

> MockAIClient 패턴 사용. testdata/ai/coaching fixture 데이터 준비.

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/coaching_service_test.go` 작성
  - [ ] `internal/controller/coaching_controller_test.go` 작성
  - [ ] MockAIClient 및 testdata/ai/coaching fixture 준비
- [ ] 구현 (GREEN)
  - [ ] Go backend API 엔드포인트 구현
  - [ ] `POST /v1/coaching/draft` 엔드포인트 구현 (SSE 스트리밍)
  - [ ] 인증 확인
  - [ ] 입력 검증 (application_id, experience_ids[], question_text, char_limit, analysis_result)
  - [ ] `prompt_templates`에서 `coaching_draft` 프롬프트 로드
  - [ ] 선택된 경험의 STAR 데이터 로드
  - [ ] 기업 분석 결과 로드
  - [ ] Claude Sonnet 4.5 API 호출 및 스트리밍
  - [ ] `cover_letters` 테이블에 초안 저장 (스트리밍 완료 후)
  - [ ] `coaching_sessions` 테이블에 세션 기록
  - [ ] 토큰 사용량 / 비용 로깅
  - [ ] 에러 처리 (스트리밍 중단, AI 실패 등)
  - [ ] 프론트엔드: EventSource로 SSE 수신
- [ ] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/v1/coaching/draft` | `{ application_id, experience_ids, question_text, char_limit, analysis_result }` | `text/event-stream` (SSE 스트리밍) |

### Request Body

```typescript
{
  application_id: string;              // UUID - 지원 ID
  experience_ids: string[];            // UUID[] - 선택한 경험 ID (1~3개)
  question_text: string;               // 자소서 문항 원문
  char_limit: number;                  // 글자수 제한
  analysis_result: {                   // Phase 5에서 받은 분석 결과
    surface_question: string;
    real_intents: Intent[];
    required_weapons: RequiredWeapons;
    writing_structure: WritingStructure;
    key_keywords: string[];
    avoid_list: string[];
  };
}
```

### Response (Server-Sent Events)

```
data: {"type":"text","content":"[상황]\n"}
data: {"type":"text","content":"저는 2023년 캡스톤 프로젝트에서..."}
data: {"type":"text","content":"팀장으로서..."}
...
data: {"type":"done","cover_letter_id":"uuid","session_id":"uuid"}
```

### 프롬프트 설계

```
당신은 한국 대기업 자소서 코칭 전문가입니다.
사용자의 실제 경험을 기반으로, 합격 자소서 수준의 초안을 작성합니다.

## 작성 규칙
1. 반드시 {{char_limit}}자 이내로 작성하세요 (현재 한국어 글자수 기준).
2. STAR 구조를 따르되, 자연스럽게 연결하세요 ([상황] [과제] 등 명시적 태그 사용).
3. 추상적 표현을 피하고, 구체적인 수치/사례를 포함하세요.
4. 아래 핵심 키워드를 자연스럽게 포함하세요: {{key_keywords}}
5. 아래 표현은 절대 사용하지 마세요: {{avoid_list}}
6. 기업의 인재상({{talent_keywords}})에 맞는 톤으로 작성하세요.

## 기업 맥락
- 기업명: {{company_name}}
- 직무: {{position}}
- 인재상: {{talent_keywords}}
- 핵심가치: {{values_keywords}}

## 문항
"{{question_text}}"

## 문항 분석
- 표면적 질문: {{surface_question}}
- 진짜 의도: {{real_intents}}

## 추천 구조
{{writing_structure}}

## 사용자 경험 (STAR 원본)
{{#each experiences}}
### 경험 {{@index}}: {{title}}
- 상황(S): {{situation}}
- 과제(T): {{task}}
- 행동(A): {{action}}
- 결과(R): {{result}}
- 보유 무기: {{weapons}}
{{/each}}

## 지시사항
위 경험을 조합하여 "{{question_text}}" 문항에 대한 {{char_limit}}자 이내 자소서 초안을 작성하세요.
각 STAR 섹션을 [상황], [과제], [행동], [결과] 태그로 구분하세요.
```

### 프론트엔드 스트리밍 수신

```typescript
// src/hooks/use-draft-coaching.ts
import { useState } from 'react';

export function useDraftCoaching() {
  const [content, setContent] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const startDraft = async (data: {
    application_id: string;
    experience_ids: string[];
    question_text: string;
    char_limit: number;
    analysis_result: any;
  }) => {
    setIsLoading(true);
    setError(null);
    setContent('');

    try {
      const response = await fetch('/v1/coaching/draft', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') {
              setIsLoading(false);
            } else {
              setContent(prev => prev + data);
            }
          }
        }
      }
    } catch (err) {
      setError('초안 생성 중 오류가 발생했습니다.');
      setIsLoading(false);
    }
  };

  return { content, isLoading, error, startDraft };
}
```

### 산출물

- `src/app/api/coaching/draft/route.ts`
- `src/lib/ai/coaching.ts` (초안 생성 로직 추가)
- `supabase/seed.sql` (`coaching_draft` 프롬프트 템플릿 추가)

---

## Step 5.1.2: 경험 선택 UI

### 목표

Phase 5.4에서 추천된 경험 목록에서 사용자가 1~3개 경험을 선택하고, 선택한 경험의 매칭 점수와 STAR 요약을 확인한 뒤 초안 생성을 요청하는 인터랙션을 구현한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/coaching/__tests__/ExperienceSelector.test.tsx`
  - `it('toggles experience selection on card click')`
  - `it('blocks selection beyond 3 experiences with message')`
  - `it('shows validation error when submitting with no selection')`
  - `it('expands card to show full STAR content on click')`
  - `it('updates weapon coverage display on selection change')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/coaching/__tests__/ExperienceSelector.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 경험 선택 상태 관리 (최대 3개)
  - [ ] 선택/해제 시 시각적 피드백 (체크박스 + 카드 보더 하이라이트)
  - [ ] 선택된 경험 수 표시 ("1/3개 선택됨")
  - [ ] 경험 카드 확장 시 STAR 전문 표시
  - [ ] "초안 작성 시작" 버튼 (최소 1개 선택 필수)
  - [ ] 선택한 경험 간 무기 중복 여부 표시 (다양한 무기 조합 권장)
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `ExperienceSelector` | `src/components/coaching/experience-selector.tsx` | `experiences: RecommendedExp[], maxSelect: 3, onConfirm: fn` | 경험 선택 래퍼 |
| `SelectableExpCard` | `src/components/coaching/selectable-exp-card.tsx` | `experience: Experience, matchScore: number, selected: boolean, onToggle: fn, expanded: boolean` | 선택 가능 경험 카드 |
| `WeaponCoverage` | `src/components/coaching/weapon-coverage.tsx` | `selectedWeapons: Weapon[], requiredWeapons: RequiredWeapons` | 선택한 경험의 무기 커버리지 표시 |

### 산출물

- `src/components/coaching/experience-selector.tsx`
- `src/components/coaching/selectable-exp-card.tsx`
- `src/components/coaching/weapon-coverage.tsx`

---

## Step 5.1.3: 스트리밍 표시

### 목표

초안 생성 API의 SSE 스트리밍 응답을 에디터 영역에 실시간으로 표시한다. STAR 섹션 태그를 감지하여 시각적으로 구분하고, 타이핑 애니메이션 효과를 적용한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/components/coaching/__tests__/DraftStreaming.test.tsx`
  - `it('shows loading indicator when streaming starts')`
  - `it('renders streamed text in real-time')`
  - `it('highlights STAR tags with correct colors')`
  - `it('updates character counter during streaming')`
  - `it('shows edit button after streaming completes')`

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/coaching/__tests__/DraftStreaming.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] Vercel AI SDK `useChat()` 또는 `useCompletion()` 훅 활용
  - [ ] 스트리밍 텍스트를 에디터 영역에 실시간 렌더링
  - [ ] STAR 태그 (`[상황]`, `[과제]`, `[행동]`, `[결과]`) 감지 → 색상 하이라이트
  - [ ] 스트리밍 진행 중 로딩 인디케이터 표시
  - [ ] 스트리밍 완료 시 "편집 모드로 전환" 버튼 활성화
  - [ ] 글자수 카운터 실시간 업데이트
  - [ ] 스트리밍 중단 버튼 (선택사항)
  - [ ] 스트리밍 완료 후 자동으로 에디터 페이지로 이동 옵션
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `DraftStreaming` | `src/components/coaching/draft-streaming.tsx` | `coverLetterId: string, onComplete: fn` | 스트리밍 표시 래퍼 |
| `StreamingEditor` | `src/components/coaching/streaming-editor.tsx` | `content: string, isStreaming: boolean` | 스트리밍 텍스트 표시 영역 (읽기 전용 에디터) |
| `StarHighlighter` | `src/components/coaching/star-highlighter.tsx` | `text: string` | STAR 태그 하이라이트 렌더러 |
| `StreamingProgress` | `src/components/coaching/streaming-progress.tsx` | `charCount: number, charLimit: number, isStreaming: boolean` | 글자수 + 스트리밍 상태 |

### 구현 코드

```typescript
// src/components/coaching/draft-streaming.tsx
'use client';

import { useCompletion } from 'ai/react';
import { useState } from 'react';

export function DraftStreaming({ coverLetterId, onComplete }) {
  const { completion, isLoading, complete, stop } = useCompletion({
    api: '/api/coaching/draft',
    onFinish: (prompt, completion) => {
      onComplete(completion);
    },
  });

  return (
    <div className="space-y-4">
      <StreamingProgress
        charCount={completion.length}
        charLimit={charLimit}
        isStreaming={isLoading}
      />
      <StreamingEditor
        content={completion}
        isStreaming={isLoading}
      />
      {isLoading && (
        <Button variant="outline" onClick={stop}>
          생성 중단
        </Button>
      )}
      {!isLoading && completion && (
        <Button onClick={() => router.push(`/coaching/${coverLetterId}/edit`)}>
          에디터에서 편집하기
        </Button>
      )}
    </div>
  );
}
```

### STAR 태그 하이라이트 규칙

| 태그 | 색상 | 배경색 |
|------|------|--------|
| `[상황]` | `text-blue-700` | `bg-blue-50` |
| `[과제]` | `text-amber-700` | `bg-amber-50` |
| `[행동]` | `text-green-700` | `bg-green-50` |
| `[결과]` | `text-purple-700` | `bg-purple-50` |

### 산출물

- `src/components/coaching/draft-streaming.tsx`
- `src/components/coaching/streaming-editor.tsx`
- `src/components/coaching/star-highlighter.tsx`
- `src/components/coaching/streaming-progress.tsx`
- `src/hooks/use-coaching.ts` (스트리밍 훅 추가)

---

## Step 5.1.4: 세션 기록

### 목표

코칭 세션의 전체 과정을 `coaching_sessions` 테이블에 기록한다. 세션 유형(draft/review/refine), 메시지 이력, 토큰 사용량을 저장하여 추후 히스토리 조회와 비용 추적에 활용한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/coaching_service_test.go`
  - `TestRecordSession_CreatesRecord`: 초안 생성 완료 후 `coaching_sessions` 레코드 생성 확인
  - `TestRecordSession_DraftType`: `session_type`이 'draft'로 저장 확인
  - `TestRecordSession_MessagesStored`: `messages` JSONB에 system/assistant 메시지 포함 확인
  - `TestRecordSession_TokenUsageTracked`: 토큰 사용량이 정확히 기록되는지 확인
- `internal/controller/coaching_controller_test.go`
  - `TestGetSessions_Success`: 세션 이력 조회 API 정상 동작 확인
  - `TestGetSessions_OnlyOwnSessions`: RLS 정책: 본인 세션만 조회 가능 확인

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/coaching_service_test.go`에 세션 기록 테스트 추가
  - [ ] `internal/controller/coaching_controller_test.go`에 세션 조회 테스트 추가
- [ ] 구현 (GREEN)
  - [ ] `coaching_sessions` 테이블에 세션 생성 로직 구현
  - [ ] 세션 타입별 분류 (`draft`, `review`, `refine`)
  - [ ] 메시지 이력 저장 (JSONB: system/user/assistant 메시지)
  - [ ] 토큰 사용량 기록 (input_tokens, output_tokens, total_cost)
  - [ ] 세션 이력 조회 API 구현
  - [ ] 코칭 이력 목록 UI (최근 세션 표시)
- [ ] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `GET` | `/api/coaching/sessions` | query: `cover_letter_id` | `{ sessions: CoachingSession[] }` |
| `GET` | `/api/coaching/sessions/[id]` | - | `{ session: CoachingSession }` |

### 데이터 구조 (coaching_sessions.messages JSONB)

```typescript
// messages 컬럼의 JSONB 구조
[
  {
    role: 'system',
    content: '프롬프트 전문...',
    timestamp: '2024-03-15T10:00:00Z'
  },
  {
    role: 'user',
    content: '선택한 경험 요약...',
    timestamp: '2024-03-15T10:00:01Z'
  },
  {
    role: 'assistant',
    content: '생성된 초안 전문...',
    timestamp: '2024-03-15T10:00:30Z',
    usage: {
      input_tokens: 2500,
      output_tokens: 1200,
      total_cost: 65  // 원 단위
    }
  }
]
```

### 비용 추적 로직

```typescript
// src/lib/ai/cost-tracker.ts
const COST_PER_TOKEN = {
  'claude-sonnet-4-5-20250929': {
    input: 0.003,   // $/1K tokens
    output: 0.015,  // $/1K tokens
  },
};

export function calculateCost(
  model: string,
  inputTokens: number,
  outputTokens: number
): number {
  const rate = COST_PER_TOKEN[model];
  const usdCost = (inputTokens / 1000 * rate.input) + (outputTokens / 1000 * rate.output);
  const krwCost = Math.round(usdCost * 1350); // 대략적 환율
  return krwCost;
}
```

### 산출물

- `src/lib/ai/cost-tracker.ts`
- `src/app/api/coaching/sessions/route.ts` (GET)
- `src/app/api/coaching/sessions/[id]/route.ts` (GET)

---

## Phase 완료 체크리스트

- [x] 경험 선택(1~3개) → 초안 생성 API 호출 → 스트리밍 표시 전체 흐름 동작
- [x] 초안이 STAR 구조로 생성되고, 글자수 제한 이내
- [x] STAR 태그가 시각적으로 하이라이트
- [x] 스트리밍 완료 후 `cover_letters` + `cover_letter_versions` + `coaching_sessions` 저장 확인
- [x] 세션 이력 조회 정상 동작
- [x] 스트리밍 완료 후 에디터 페이지로 이동 가능
- [x] 에러 상태 (스트리밍 실패, 네트워크 에러 등) 처리
- [ ] Claude API 비용: ~65원/건 이내 확인 (실 API 연동 후 검증 필요)
- [x] 테스트
  - [x] `moon run backend:test` → 전체 통과
  - [x] `moon run web:test` → 전체 통과
  - [x] `moon run :lint` → 경고 0건
  - [x] `moon run web:build` → 빌드 성공

---

## 다음 Phase

**[Phase 5.2: 코칭 에디터](./phase-5.2-coaching-editor.md)** — Tiptap 리치 텍스트 에디터 + 글자수 카운터 + 자동 저장
