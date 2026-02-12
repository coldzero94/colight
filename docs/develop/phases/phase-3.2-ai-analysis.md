# Phase 3.2: AI 기업 분석

> **⚠️ 아키텍처 변경 사항**: AI 분석은 Go 백엔드에서 Anthropic Go SDK를 사용하여 구현합니다. Vercel AI SDK(`@ai-sdk/anthropic`)는 사용하지 않습니다.
>
> - AI 호출: `apps/backend/internal/infrastructure/ai/anthropic.go`
> - SSE 스트리밍: Go Gin의 `c.SSEvent()` 사용
> - 프론트엔드는 `EventSource`로 SSE를 수신하여 UI를 업데이트합니다.

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 수집된 채용공고 + 기업 데이터를 Claude Sonnet 4.5에 전달하여 핵심가치, 인재상, 동향, 전략키워드, 피해야 할 표현을 스트리밍으로 분석하고, 전체 파이프라인(URL 파싱 → 데이터 수집 → AI 분석)을 오케스트레이션하는 메인 엔드포인트 구축 |
| **선행 조건** | Phase 3.1 완료 (기업 데이터 API), Anthropic API 키 발급, 환경변수 설정 (ANTHROPIC_API_KEY) |
| **스프린트** | Sprint 2 — Day 5~6 |
| **관련 기능** | F09 (인재상 분석 + 최근 동향 + 전략 키워드) |
| **예상 공수** | 2일 (16시간) |
| **산출물** | 분석 프롬프트 템플릿, 스트리밍 API, 파이프라인 오케스트레이터, Zod 스키마 검증 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 3.2.1 | 분석 프롬프트 설계 | ⬜ |
| 3.2.2 | 스트리밍 API | ⬜ |
| 3.2.3 | 파이프라인 오케스트레이션 | ⬜ |
| 3.2.4 | Zod 스키마 검증 | ⬜ |

---

## Step 3.2.1: 분석 프롬프트 설계

### 목표

`prompt_templates` 테이블에 `company_analysis` 카테고리의 프롬프트를 등록한다. 채용공고 정보 + 기업 프로필 + 재무 데이터 + 최근 뉴스를 입력으로 받아, 구조화된 분석 결과를 JSON으로 출력하도록 설계한다.

### 체크리스트

- [ ] `company_analysis` 프롬프트 시스템 프롬프트 작성
- [ ] 사용자 프롬프트 템플릿 작성 (변수 주입 포맷)
- [ ] 출력 JSON 스키마 정의
- [ ] `prompt_templates` 테이블에 시드 데이터 추가
- [ ] 프롬프트 버전 관리 체계 수립

### 프롬프트 상세

```sql
-- prompt_templates 시드 데이터
INSERT INTO prompt_templates (category, name, version, system_prompt, user_prompt_template, model, temperature, max_tokens)
VALUES (
  'company_analysis',
  'comprehensive_analysis',
  1,
  -- system_prompt (아래 참조)
  '',
  -- user_prompt_template (아래 참조)
  '',
  'claude-sonnet-4-5-20250929',
  0.3,
  4000
);
```

### 시스템 프롬프트

```
당신은 취업 준비생을 위한 기업 분석 전문가입니다.
주어진 채용공고, 기업 정보, 재무 데이터, 최근 뉴스를 종합 분석하여
자소서 작성에 실질적으로 도움이 되는 인사이트를 제공합니다.

분석 원칙:
1. 공식 데이터에 기반한 사실적 분석 (추측 최소화)
2. 자소서 작성에 즉시 활용 가능한 구체적 키워드와 전략 제시
3. 기업이 '실제로' 원하는 인재상을 공고 행간에서 읽어내기
4. 피해야 할 클리셰와 표현을 구체적으로 명시

반드시 지정된 JSON 스키마에 맞춰 응답하세요.
```

### 사용자 프롬프트 템플릿

```
다음 정보를 바탕으로 기업 종합 분석을 수행해주세요.

## 채용공고 정보
{{jobPosting}}

## 기업 프로필
{{companyProfile}}

## 재무 정보
{{financials}}

## 최근 뉴스 (최근 3개월)
{{recentNews}}

---

다음 항목을 분석하여 JSON으로 응답해주세요:

1. **핵심가치 (core_values)**: 이 기업이 가장 중시하는 가치 3~5개
   - 각 가치에 대한 keyword, description, evidence (근거 출처)

2. **인재상 (talent_profile)**: 이 기업/직무가 원하는 인재 특성 3~5가지
   - 각 특성의 trait, description, evidence, priority (high/medium/low)

3. **최근 동향 (recent_trends)**: 기업의 최근 주요 이슈 3~5건
   - 각 동향의 title, summary, relevance_to_job (이 직무와의 연관성)
   - 뉴스 데이터가 없으면 공시/재무 데이터 기반으로 분석

4. **전략 키워드 (strategy_keywords)**: 자소서에 반드시 녹여야 할 키워드 5~7개
   - 기업 가치 + 직무 요건 + 최근 동향에서 도출
   - 구체적이고 차별화된 키워드 (예: "디지털 전환" 보다는 "AI 기반 업무 자동화 경험")

5. **피해야 할 표현 (avoid_expressions)**: 이 기업/직무에 어울리지 않는 표현 3~5개
   - 왜 피해야 하는지 이유 포함
   - 대안 표현 제안
```

### 입력 변수 정의

| 변수 | 타입 | 설명 | 소스 |
|------|------|------|------|
| `{{jobPosting}}` | JSON string | 구조화된 채용공고 데이터 | Phase 3 `/api/analyze` 결과 |
| `{{companyProfile}}` | JSON string | DART 기업 개황 | Phase 3.1 DART API |
| `{{financials}}` | JSON string | DART 재무제표 | Phase 3.1 DART API |
| `{{recentNews}}` | JSON string | 네이버 뉴스 상위 10건 | Phase 3.1 네이버 뉴스 API |

### 출력 스키마

```typescript
// src/lib/ai/analysis-schemas.ts

interface CompanyAnalysis {
  coreValues: {
    keyword: string;
    description: string;
    evidence: string;
  }[];

  talentProfile: {
    trait: string;
    description: string;
    evidence: string;
    priority: 'high' | 'medium' | 'low';
  }[];

  recentTrends: {
    title: string;
    summary: string;
    relevanceToJob: string;
  }[];

  strategyKeywords: string[];

  avoidExpressions: {
    expression: string;
    reason: string;
    alternative: string;
  }[];
}
```

### 검증 방법

- [ ] 프롬프트 템플릿 → DB 정상 저장 확인
- [ ] 변수 주입 후 완성된 프롬프트 형식 확인
- [ ] Claude API에 직접 테스트 → 유효한 JSON 응답 확인
- [ ] 출력 스키마와 실제 응답 구조 일치 확인
- [ ] 다양한 기업(대기업, 중견, 스타트업)에 대해 분석 품질 확인

### 산출물

- `src/lib/ai/analysis-schemas.ts` (CompanyAnalysis 타입 + Zod 스키마)
- `src/lib/ai/prompts/company-analysis.ts` (프롬프트 빌더)
- DB 시드: `prompt_templates` 1건 추가 (company_analysis/comprehensive_analysis)

---

## Step 3.2.2: 스트리밍 API

### 목표

Go backend에서 Claude Sonnet 4.5의 기업 분석 결과를 Server-Sent Events (SSE)로 실시간 스트리밍하여 전달하는 API를 구현한다. 분석 완료 후 결과를 `company_analyses` 테이블에 저장한다.

### 체크리스트

- [ ] Go backend API 엔드포인트 구현
- [ ] `POST /v1/analyze/comprehensive` SSE 스트리밍 구현
- [ ] Claude Sonnet 4.5 API 호출 및 스트리밍 처리
- [ ] 스트림 완료 후 `company_analyses` 테이블에 결과 저장
- [ ] 스트리밍 중 에러 처리 (타임아웃, API 에러)
- [ ] 토큰 사용량 로깅
- [ ] 인증 미들웨어 적용
- [ ] 프론트엔드: EventSource를 사용한 SSE 수신

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/v1/analyze/comprehensive` | `{ jobPosting: JobPosting, companyData: CompanyData }` | `text/event-stream` (SSE 스트리밍 → JSON) |

### 스트리밍 구현

프론트엔드에서 EventSource를 사용하여 SSE를 수신:

```typescript
// src/hooks/use-analysis-stream.ts
import { useState, useEffect } from 'react';

export function useAnalysisStream() {
  const [content, setContent] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const startAnalysis = async (jobPosting: JobPosting, companyData: CompanyData) => {
    setIsLoading(true);
    setError(null);
    setContent('');

    try {
      const response = await fetch('/v1/analyze/comprehensive', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ jobPosting, companyData }),
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
            setContent(prev => prev + data);
          }
        }
      }
    } catch (err) {
      setError('분석 중 오류가 발생했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  return { content, isLoading, error, startAnalysis };
}
```

### DB 테이블 (company_analyses)

```sql
-- 이 테이블은 Phase 0 마이그레이션에서 이미 생성됨
-- 여기서는 analysis_data 컬럼의 실제 구조를 정의

-- company_analyses.analysis_data JSONB 구조:
-- {
--   "coreValues": [...],
--   "talentProfile": [...],
--   "recentTrends": [...],
--   "strategyKeywords": [...],
--   "avoidExpressions": [...]
-- }
```

### Vercel AI SDK 설정

```typescript
// src/lib/ai/config.ts

import { createAnthropic } from '@ai-sdk/anthropic';

export const anthropicProvider = createAnthropic({
  apiKey: process.env.ANTHROPIC_API_KEY,
});

// 모델 선택 유틸리티
export function getModel(purpose: 'analysis' | 'coaching' | 'parsing') {
  switch (purpose) {
    case 'analysis':
    case 'coaching':
      return anthropicProvider('claude-sonnet-4-5-20250929');
    case 'parsing':
      // OpenAI GPT-4.1 mini는 별도 설정
      return null;
  }
}
```

### 비용 관리

- Claude Sonnet 4.5 사용: 건당 약 65원
- 입력 토큰: ~3000 (공고 + 기업정보 + 뉴스)
- 출력 토큰: ~2000 (분석 결과 JSON)
- 스트리밍으로 체감 대기 시간 단축 (실제 7~15초 → 체감 2~3초)

### 검증 방법

- [ ] 스트리밍 응답이 실시간으로 전달되는지 확인 (SSE 이벤트)
- [ ] 스트림 완료 후 전체 JSON 파싱 성공 확인
- [ ] `company_analyses` 테이블에 결과 정상 저장 확인
- [ ] 토큰 사용량 로깅 확인
- [ ] Anthropic API 에러 시 적절한 에러 응답 확인
- [ ] 미인증 요청 시 401 반환 확인

### 산출물

- Go backend: `POST /v1/analyze/comprehensive` 엔드포인트 (backend 레포지토리)
- 프론트엔드: `src/hooks/use-analysis-stream.ts` (SSE 스트리밍 훅)
- 프론트엔드: `@/api/generated/sdk.gen.ts` (SDK 업데이트)

---

## Step 3.2.3: 파이프라인 오케스트레이션

### 목표

`/api/analyze` 메인 엔드포인트를 확장하여, URL 파싱 → 기업 데이터 수집 → AI 종합 분석의 전체 파이프라인을 하나의 요청으로 오케스트레이션한다. 각 단계의 진행 상태를 스트리밍으로 전달한다.

### 체크리스트

- [ ] `/api/analyze` 메인 엔드포인트 확장 (파이프라인 모드)
- [ ] 단계별 진행 상태 스트리밍 이벤트 정의
- [ ] 단계별 에러 처리 및 fallback 로직
- [ ] 파이프라인 상태 관리 (진행중/완료/실패)
- [ ] 중간 결과 저장 (URL 파싱 결과, 기업 데이터)
- [ ] 타임아웃 처리 (전체 60초 제한)

### 파이프라인 흐름

```
[Step 1: URL 파싱]
    │ 이벤트: { step: 'parsing', status: 'started' }
    │ → 도메인 감지 → 파서 선택 → HTML 파싱 → AI 정규화
    │ 이벤트: { step: 'parsing', status: 'completed', data: JobPosting }
    │
    ▼
[Step 2: 기업 데이터 수집]
    │ 이벤트: { step: 'company_data', status: 'started' }
    │ → DART API + 네이버 뉴스 (병렬)
    │ 이벤트: { step: 'company_data', status: 'completed', data: CompanyData }
    │
    ▼
[Step 3: AI 종합 분석]
    │ 이벤트: { step: 'analysis', status: 'started' }
    │ → Claude Sonnet 4.5 스트리밍
    │ 이벤트: { step: 'analysis', status: 'streaming', chunk: '...' }
    │ ...
    │ 이벤트: { step: 'analysis', status: 'completed', data: CompanyAnalysis }
    │
    ▼
[완료]
    이벤트: { step: 'complete', analysisId: '...' }
```

### API 엔드포인트 (확장)

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/api/analyze` | `{ url: string, mode?: 'parse_only' \| 'full' }` | `ReadableStream` (단계별 진행 이벤트) |

### 스트리밍 이벤트 포맷

```typescript
// SSE (Server-Sent Events) 포맷

// 진행 상태 이벤트
interface PipelineEvent {
  step: 'parsing' | 'company_data' | 'analysis' | 'complete' | 'error';
  status: 'started' | 'completed' | 'streaming' | 'failed';
  data?: any;
  error?: string;
  progress?: number; // 0~100
}

// 스트림 이벤트 예시:
// data: {"step":"parsing","status":"started","progress":0}
// data: {"step":"parsing","status":"completed","progress":30,"data":{...JobPosting}}
// data: {"step":"company_data","status":"started","progress":30}
// data: {"step":"company_data","status":"completed","progress":60,"data":{...CompanyData}}
// data: {"step":"analysis","status":"started","progress":60}
// data: {"step":"analysis","status":"streaming","chunk":"핵심가치 분석 결과..."}
// data: {"step":"analysis","status":"completed","progress":100,"data":{...CompanyAnalysis}}
// data: {"step":"complete","analysisId":"uuid-here"}
```

### 오케스트레이터 구현

```typescript
// src/lib/pipeline/analyze-orchestrator.ts

interface PipelineOptions {
  url: string;
  userId: string;
  mode: 'parse_only' | 'full';
  onEvent: (event: PipelineEvent) => void;
}

async function runAnalysisPipeline(options: PipelineOptions): Promise<void> {
  const { url, userId, mode, onEvent } = options;

  try {
    // Step 1: URL 파싱
    onEvent({ step: 'parsing', status: 'started', progress: 0 });
    const jobPosting = await parseJobUrl(url);
    onEvent({ step: 'parsing', status: 'completed', progress: 30, data: jobPosting });

    if (mode === 'parse_only') {
      onEvent({ step: 'complete', status: 'completed' });
      return;
    }

    // Step 2: 기업 데이터 수집
    onEvent({ step: 'company_data', status: 'started', progress: 30 });
    const companyData = await fetchCompanyData(jobPosting.companyName);
    onEvent({ step: 'company_data', status: 'completed', progress: 60, data: companyData });

    // Step 3: AI 종합 분석 (스트리밍)
    onEvent({ step: 'analysis', status: 'started', progress: 60 });
    const analysis = await runComprehensiveAnalysis({
      jobPosting,
      companyData,
      onChunk: (chunk) => {
        onEvent({ step: 'analysis', status: 'streaming', data: { chunk } });
      },
    });
    onEvent({ step: 'analysis', status: 'completed', progress: 100, data: analysis });

    // 결과 저장
    const analysisId = await saveAnalysisResult({
      userId,
      jobPosting,
      companyData,
      analysis,
    });

    onEvent({ step: 'complete', status: 'completed', data: { analysisId } });

  } catch (error) {
    onEvent({
      step: 'error',
      status: 'failed',
      error: error instanceof Error ? error.message : '알 수 없는 오류가 발생했습니다.',
    });
  }
}
```

### 에러 복구 전략

| 단계 | 에러 시나리오 | 대응 |
|------|-------------|------|
| URL 파싱 | HTML fetch 실패 | 에러 반환 (파이프라인 중단) |
| URL 파싱 | 파서 데이터 추출 실패 | AI fallback 시도 |
| 기업 데이터 | DART API 실패 | partial data로 계속 (뉴스만으로 분석) |
| 기업 데이터 | 네이버 뉴스 실패 | partial data로 계속 (DART만으로 분석) |
| AI 분석 | Claude API 에러 | 1회 재시도 → 실패 시 에러 반환 |
| AI 분석 | JSON 파싱 실패 | 1회 재시도 (temperature 0.1로 낮춤) |
| 전체 | 타임아웃 (60초) | 현재까지 결과 저장 + 에러 이벤트 |

### 검증 방법

- [ ] 전체 파이프라인 (URL → 파싱 → 데이터 수집 → AI 분석) E2E 동작 확인
- [ ] 각 단계별 진행 이벤트가 순서대로 스트리밍되는지 확인
- [ ] `mode: 'parse_only'` → URL 파싱만 수행 후 완료 확인
- [ ] DART 실패 시 뉴스만으로 AI 분석 수행 확인
- [ ] AI 분석 스트리밍 청크가 실시간으로 전달되는지 확인
- [ ] 분석 완료 후 `company_analyses` 테이블에 저장 확인
- [ ] 60초 타임아웃 동작 확인

### 산출물

- `src/lib/pipeline/analyze-orchestrator.ts`
- `src/app/api/analyze/route.ts` (확장)

---

## Step 3.2.4: Zod 스키마 검증

### 목표

AI 분석 결과의 JSON 출력을 Zod 스키마로 검증하여 타입 안전성을 보장한다. 검증 실패 시 1회 재시도하고, 재시도에도 실패하면 부분 결과를 반환한다.

### 체크리스트

- [ ] `CompanyAnalysis` Zod 스키마 정의 (전체 분석 결과)
- [ ] AI 응답 JSON 파싱 + Zod 검증 함수 구현
- [ ] 검증 실패 시 재시도 로직 (에러 메시지를 피드백으로 포함)
- [ ] partial parsing (일부 필드 누락 시 기본값 적용)
- [ ] 검증 결과 로깅 (성공률 추적)

### Zod 스키마 정의

```typescript
// src/lib/ai/analysis-schemas.ts

import { z } from 'zod';

export const CoreValueSchema = z.object({
  keyword: z.string().min(1, '키워드는 필수입니다'),
  description: z.string().min(10, '설명은 최소 10자 이상이어야 합니다'),
  evidence: z.string().min(1, '근거는 필수입니다'),
});

export const TalentTraitSchema = z.object({
  trait: z.string().min(1),
  description: z.string().min(10),
  evidence: z.string().min(1),
  priority: z.enum(['high', 'medium', 'low']),
});

export const TrendSchema = z.object({
  title: z.string().min(1),
  summary: z.string().min(10),
  relevanceToJob: z.string().min(1),
});

export const AvoidExpressionSchema = z.object({
  expression: z.string().min(1),
  reason: z.string().min(5),
  alternative: z.string().min(1),
});

export const CompanyAnalysisSchema = z.object({
  coreValues: z.array(CoreValueSchema).min(3).max(5),
  talentProfile: z.array(TalentTraitSchema).min(3).max(5),
  recentTrends: z.array(TrendSchema).min(1).max(5),
  strategyKeywords: z.array(z.string().min(1)).min(5).max(7),
  avoidExpressions: z.array(AvoidExpressionSchema).min(3).max(5),
});

export type CompanyAnalysis = z.infer<typeof CompanyAnalysisSchema>;
```

### 검증 + 재시도 로직

```typescript
// src/lib/ai/analysis-validator.ts

interface ValidationResult {
  success: boolean;
  data?: CompanyAnalysis;
  errors?: z.ZodError;
  retried: boolean;
}

/**
 * AI 응답을 파싱 + 검증
 * 1차 실패 시 에러 내용을 AI에 피드백하여 재시도
 */
async function validateAnalysisResponse(
  rawText: string,
  retryFn?: (feedback: string) => Promise<string>
): Promise<ValidationResult> {
  // 1. JSON 파싱 시도
  let parsed: unknown;
  try {
    // JSON 블록 추출 (```json ... ``` 패턴 처리)
    const jsonStr = extractJsonFromText(rawText);
    parsed = JSON.parse(jsonStr);
  } catch {
    if (!retryFn) return { success: false, retried: false };

    // JSON 파싱 실패 → 재시도
    const retryText = await retryFn(
      '이전 응답이 유효한 JSON이 아닙니다. 반드시 순수 JSON만 응답해주세요. 마크다운 코드 블록이나 설명 없이 JSON만 반환하세요.'
    );
    return validateAnalysisResponse(retryText); // 재귀 (재시도 없음)
  }

  // 2. Zod 검증
  const result = CompanyAnalysisSchema.safeParse(parsed);
  if (result.success) {
    return { success: true, data: result.data, retried: false };
  }

  // 3. 검증 실패 → 재시도
  if (retryFn) {
    const errorMessages = result.error.issues
      .map(i => `- ${i.path.join('.')}: ${i.message}`)
      .join('\n');

    const retryText = await retryFn(
      `이전 응답의 JSON 구조에 문제가 있습니다:\n${errorMessages}\n\n위 오류를 수정하여 올바른 JSON을 다시 반환해주세요.`
    );
    const retryResult = CompanyAnalysisSchema.safeParse(JSON.parse(extractJsonFromText(retryText)));
    if (retryResult.success) {
      return { success: true, data: retryResult.data, retried: true };
    }
  }

  // 4. 재시도 실패 → partial parsing (가능한 필드만 추출)
  const partialData = parsePartialAnalysis(parsed);
  return {
    success: false,
    data: partialData,
    errors: result.error,
    retried: true,
  };
}

/**
 * 부분 파싱 (일부 필드 누락 시 기본값 적용)
 */
function parsePartialAnalysis(data: unknown): CompanyAnalysis {
  const obj = data as Record<string, unknown>;
  return {
    coreValues: safeParseArray(obj.coreValues, CoreValueSchema, []),
    talentProfile: safeParseArray(obj.talentProfile, TalentTraitSchema, []),
    recentTrends: safeParseArray(obj.recentTrends, TrendSchema, []),
    strategyKeywords: Array.isArray(obj.strategyKeywords) ? obj.strategyKeywords.filter(Boolean) : [],
    avoidExpressions: safeParseArray(obj.avoidExpressions, AvoidExpressionSchema, []),
  };
}
```

### 검증 방법

- [ ] 정상 JSON → Zod 검증 통과 확인
- [ ] 필수 필드 누락 JSON → 에러 목록 정확히 반환 확인
- [ ] 마크다운 코드 블록 (```json ... ```) → JSON 추출 성공 확인
- [ ] 검증 실패 → 재시도 → 성공 흐름 확인
- [ ] 재시도에도 실패 → partial parsing 동작 확인
- [ ] 검증 성공률 로깅 확인

### 산출물

- `src/lib/ai/analysis-schemas.ts` (Zod 스키마 완성)
- `src/lib/ai/analysis-validator.ts`
- `__tests__/lib/ai/analysis-validator.test.ts`

---

## Phase 완료 체크리스트

- [ ] `company_analysis` 프롬프트가 `prompt_templates`에 등록됨
- [ ] Claude Sonnet 4.5 기반 기업 분석 스트리밍 API 동작
- [ ] 분석 결과가 `company_analyses` 테이블에 정상 저장됨
- [ ] 전체 파이프라인 (URL → 파싱 → 데이터 수집 → AI 분석) 오케스트레이션 동작
- [ ] 단계별 진행 이벤트가 SSE로 정상 전달됨
- [ ] Zod 스키마 검증 통과 (재시도 포함)
- [ ] Partial failure 시 가능한 데이터로 분석 진행
- [ ] 토큰 사용량 로깅 동작
- [ ] 환경변수 문서 업데이트 (ANTHROPIC_API_KEY)
- [ ] 단위 테스트 + 통합 테스트 통과

---

## 다음 Phase

→ [Phase 3.3: 분석 리포트 UI](./phase-3.3-analysis-ui.md) — 스트리밍 분석 결과를 시각적으로 표시하는 리포트 UI를 구축한다.
