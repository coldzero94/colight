# AI 파이프라인

> 작성일: 2026-02-11

---

> **⚠️ 아키텍처 노트**: 이 문서의 일부 코드 예시가 TypeScript로 작성되어 있으나, 실제 구현은 Go 백엔드(Gin + Ent)에서 처리합니다. TypeScript 코드는 로직 설명을 위한 의사 코드(pseudocode)로 참고하세요. 실제 Go 구현 패턴은 `01-architecture.md`를 참조하세요.

## 1. 모델 티어링 전략

| 계층 | 모델 | 건당 비용 | 용도 | 선택 기준 |
|------|------|----------|------|----------|
| **경량 (Tier 1)** | Gemini 2.0 Flash / Groq Llama 3.3 70B | ~3~5원 | 공고 파싱, 경험 분류, 인터뷰 대화, 2차 매칭 | 빠른 응답(1~3초), 구조화 출력, 비용 최소화 |
| **고급 (Tier 2)** | Claude Sonnet 4.5 | ~65원 | 기업 종합 분석, 문항 분석, 초안/첨삭 코칭 | 한국어 분석 품질, 긴 컨텍스트, 심층 추론 |
| **임베딩** | text-embedding-3-small | ~0.5원 | 경험 벡터화, 유사도 1차 필터링 | 1536차원, 저비용, pgvector 호환 |

### 경량 모델 프로바이더 선택

`LLM_LIGHT_PROVIDER` 환경변수로 경량 모델 프로바이더를 선택합니다:

| 프로바이더 | 모델 | 장점 | 단점 |
|------------|------|------|------|
| **gemini** (기본) | Gemini 2.0 Flash | 무료 티어 넉넉, 구조화 출력 우수 | Google AI Studio 계정 필요 |
| **groq** | Llama 3.3 70B | 초고속 응답(\~0.5초), 무료 티어 관대 | 모델 품질 Gemini 대비 약간 하위 |

### 비용 최적화 원칙

1. **구조화 작업은 경량 모델 (Gemini/Groq)**: JSON 변환, 분류, 키워드 추출
2. **분석/코칭은 Claude Sonnet 4.5**: 맥락 이해, 한국어 품질이 중요한 작업
3. **대량 비교는 임베딩 우선**: 유사도 필터링 후 LLM 정밀 분석
4. **캐싱으로 중복 호출 방지**: 기업 분석 7일 TTL

---

## 2. Go 백엔드 AI 클라이언트 설정

**AI 호출은 Go 백엔드에서 처리**합니다. 프론트엔드는 Go API를 통해 간접 호출만 합니다.

### 2.1 공통 LLM 인터페이스

경량 모델을 Gemini/Groq 중 환경변수로 선택할 수 있도록 공통 인터페이스를 사용합니다.

```go
// apps/backend/internal/infrastructure/ai/llm.go
package ai

// LLMProvider — 경량 모델 공통 인터페이스
type LLMProvider interface {
    // Call sends a prompt and returns the response text
    Call(ctx context.Context, req LLMRequest) (LLMResponse, error)
}

type LLMRequest struct {
    SystemPrompt string
    UserPrompt   string
    Temperature  float64
    MaxTokens    int
    JSONMode     bool // structured output
}

type LLMResponse struct {
    Content      string
    InputTokens  int
    OutputTokens int
    Model        string
}
```

### 2.2 프로바이더 구현

```go
// apps/backend/internal/infrastructure/ai/gemini.go
type GeminiProvider struct {
    client *genai.Client
    model  string // "gemini-2.0-flash"
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    return &GeminiProvider{client: client, model: "gemini-2.0-flash"}, err
}

func (g *GeminiProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
    // Google AI SDK 호출
}
```

```go
// apps/backend/internal/infrastructure/ai/groq.go
type GroqProvider struct {
    client *http.Client
    apiKey string
    model  string // "llama-3.3-70b-versatile"
}

func NewGroqProvider(apiKey string) *GroqProvider {
    return &GroqProvider{apiKey: apiKey, model: "llama-3.3-70b-versatile"}
}

func (g *GroqProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
    // Groq REST API 호출 (OpenAI-compatible)
}
```

### 2.3 AIProvider (통합)

```go
// apps/backend/internal/infrastructure/ai/provider.go
package ai

type AIProvider struct {
    light           LLMProvider       // Gemini or Groq (configurable)
    anthropicClient *anthropic.Client // Claude (heavy)
}

func NewAIProvider(cfg AIConfig) (*AIProvider, error) {
    // 경량 프로바이더 선택 (LLM_LIGHT_PROVIDER 환경변수)
    var light LLMProvider
    switch cfg.LightProvider {
    case "groq":
        light = NewGroqProvider(cfg.GroqAPIKey)
    default: // "gemini"
        light, _ = NewGeminiProvider(cfg.GeminiAPIKey)
    }

    return &AIProvider{
        light:           light,
        anthropicClient: anthropic.NewClient(cfg.AnthropicAPIKey),
    }, nil
}

// 경량 작업용 (Gemini Flash / Groq Llama)
func (p *AIProvider) CallLight(ctx context.Context, req LLMRequest) (LLMResponse, error) {
    return p.light.Call(ctx, req)
}

// 고급 분석용 (Claude Sonnet 4.5)
func (p *AIProvider) CallHeavy(ctx context.Context, prompt string) (string, error) {
    resp, err := p.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
        Model: "claude-sonnet-4-5-20250929",
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
        },
    })
    return resp.Content[0].Text, err
}

// 스트리밍용 (SSE)
func (p *AIProvider) StreamHeavy(ctx context.Context, prompt string) (<-chan string, error) {
    stream := p.anthropicClient.Messages.NewStreaming(ctx, anthropic.MessageNewParams{...})
    // SSE 채널 반환
}
```

**프론트엔드는 SSE 수신만**:
```typescript
// apps/web/src/hooks/use-coaching-stream.ts
const response = await fetch('/api/coaching/draft', {
  method: 'POST',
  body: JSON.stringify(request),
});

const reader = response.body.getReader();
// SSE 파싱하여 UI 업데이트
```

---

## 3. DB 기반 프롬프트 관리

### 프롬프트 로드 + 변수 치환

```typescript
// src/lib/ai/prompts.ts
import { createClient } from '@/lib/supabase/server';

interface PromptTemplate {
  system_prompt: string;
  user_prompt_template: string;
  model: string;
  temperature: number;
  max_tokens: number;
}

// DB에서 프롬프트 로드
export async function loadPrompt(
  category: string,
  subCategory: string
): Promise<PromptTemplate> {
  const supabase = await createClient();

  const { data, error } = await supabase
    .from('prompt_templates')
    .select('system_prompt, user_prompt_template, model, temperature, max_tokens')
    .eq('category', category)
    .eq('sub_category', subCategory)
    .eq('is_active', true)
    .order('version', { ascending: false })
    .limit(1)
    .single();

  if (error || !data) throw new Error(`프롬프트 로드 실패: ${category}/${subCategory}`);
  return data;
}

// {{변수}} 치환
export function substituteVariables(
  template: string,
  variables: Record<string, string>
): string {
  return Object.entries(variables).reduce(
    (result, [key, value]) => result.replaceAll(`{{${key}}}`, value),
    template
  );
}
```

### 변수 목록

| 변수명 | 설명 | 주입 시점 |
|--------|------|----------|
| `{{experience_text}}` | 사용자 입력 경험 원문 | 경험 분류 시 |
| `{{experience_star}}` | STAR 구조화된 경험 | 코칭 시 |
| `{{weapon_categories}}` | DB에서 로드된 전체 무기 목록 | 경험 분류 시 |
| `{{weapon_name}}` | 특정 무기명 | 강화 코칭 시 |
| `{{weapon_code}}` | 무기 코드 | 강화 코칭 시 |
| `{{company_analysis}}` | 기업 분석 결과 JSON | 코칭 시 |
| `{{talent_profile}}` | 기업 인재상 | 코칭 시 |
| `{{question_text}}` | 자소서 문항 | 문항 분석/코칭 시 |
| `{{char_limit}}` | 글자수 제한 | 코칭 시 |
| `{{user_weapons_summary}}` | 사용자 보유 무기 요약 | 문항 분석 시 |
| `{{conversation_history}}` | 인터뷰 대화 이력 | 경험 인터뷰 시 |

---

## 4. 파이프라인 상세

### 파이프라인 1: 경험 무기 태깅

```
[입력: 경험 텍스트]
    │
    ▼
[weapon_categories 로드] ← DB에서 전체 무기 목록 조회 (35개)
    │
    ▼
[prompt_templates 로드] ← category='experience_classify', sub='weapon_tagging'
    │
    ▼
[변수 치환] ← {{experience_text}}, {{weapon_categories}}
    │
    ▼
[경량 모델 호출 (Gemini/Groq)] ← temperature: 0.2, max_tokens: 1500
    │
    ▼
[JSON 파싱]
    │
    ├── primary_weapon: {code, confidence, reasoning}
    ├── secondary_weapons: [{code, confidence, reasoning}, ...]
    ├── star: {situation, task, action, result}
    ├── matchable_questions: string[]
    └── strength_keywords: string[]
    │
    ▼
[DB 저장]
    ├── experience_weapons (주 무기 + 부 무기)
    └── experience_tags (키워드, STAR)
```

- **모델**: Gemini 2.0 Flash / Groq Llama 3.3 (LLM_LIGHT_PROVIDER)
- **비용**: ~3~5원/건
- **응답 시간**: 1~3초

---

### 파이프라인 2: 채용공고 파싱

```
[입력: 채용공고 HTML/텍스트]
    │
    ▼
[URL 도메인 판별]
    │
    ├── [정적 사이트] → goquery 파싱 (Go 백엔드)
    │       └── 잡코리아, 캐치 등
    │
    ├── [동적 사이트] → Koyeb Worker (Playwright)
    │       └── 원티드 등 SPA
    │
    └── [사람인] → 사람인 API 직접 호출
    │
    ▼
[HTML → 텍스트 추출] (구조화 전 원문)
    │
    ▼
[경량 모델 호출 (Gemini/Groq)] ← 공고 구조화 프롬프트
    │
    ▼
[JSON 출력]
    ├── company_name: string
    ├── position: string
    ├── department: string
    ├── job_type: string
    ├── experience_level: string
    ├── main_tasks: string[]
    ├── requirements: string[]
    ├── preferred: string[]
    ├── required_skills: string[]
    ├── soft_skills: string[]
    ├── company_values_hints: string[]
    └── deadline: string
```

- **모델**: Gemini 2.0 Flash / Groq Llama 3.3 (LLM_LIGHT_PROVIDER)
- **비용**: ~3~5원/건
- **응답 시간**: 2~5초 (크롤링 포함)

---

### 파이프라인 3: 기업 종합 분석

```
[입력: 파싱된 공고 + 기업 데이터]
    │
    ▼
[캐시 확인] ← company_analysis_cache (7일 TTL)
    ├── [HIT] → 캐시 결과 즉시 반환
    └── [MISS] → 계속 진행
    │
    ▼
[기업 데이터 병렬 수집]
    ├── [DART OpenAPI] → 기업 기본정보, 재무제표, 임원 현황
    ├── [네이버 뉴스 API] → 최근 뉴스 5~10건 (3개월 이내)
    └── [talent_profiles] → 사전 DB 인재상 조회 (있으면)
    │
    ▼
[데이터 병합] → 공고 + DART + 뉴스 + 인재상 통합
    │
    ▼
[Claude Sonnet 4.5 스트리밍 호출]
    │
    ▼
[분석 결과 (스트리밍)]
    ├── core_values: [{keyword, description}]       ← 핵심가치 3~5개
    ├── talent_profile: [{trait, description, evidence}] ← 인재상 3~5가지
    ├── recent_trends: [{title, summary, relevance}] ← 최근 동향
    ├── strategy_keywords: string[]                 ← 자소서 전략 키워드 5~7개
    └── avoid_expressions: string[]                 ← 피해야 할 표현
    │
    ▼
[캐시 저장] ← company_analysis_cache (expires_at = NOW() + 7일)
    │
    ▼
[company_analyses 저장] ← 사용자별 분석 결과 영구 저장
```

- **모델**: Claude Sonnet 4.5
- **비용**: ~65원/건
- **응답 시간**: 5~15초 (스트리밍)

---

### 파이프라인 4: 경험 매칭 (하이브리드)

```
[입력: 기업 분석 결과 + 사용자 경험 DB]
    │
    ▼
[1차: 임베딩 유사도 필터링]
    │ 기업 요구사항 텍스트 → 임베딩 생성 (text-embedding-3-small)
    │ experiences.embedding과 코사인 유사도 계산 (pgvector)
    │ 상위 N개(5~10개) 경험 선별
    │
    ▼
[2차: LLM 정밀 매칭]
    │ 선별된 경험 × 기업 요구사항 → 경량 모델 (Gemini/Groq)
    │
    ▼
[매칭 결과]
    ├── overall_fit: number (0~100, 가중 평균)
    │     ├── job_relevance: 40%    ← 직무 관련도
    │     ├── talent_fit: 35%       ← 인재상 부합도
    │     └── uniqueness: 25%       ← 차별화 점수
    │
    ├── category_scores: [{category, score, matched_experiences}]
    │
    ├── experience_matches: [
    │     {experience_id, fit_score, reasoning, suggested_angle}
    │   ]
    │
    └── recommendations: string[]   ← 보완 제안
```

- **모델**: text-embedding-3-small (~0.5원) + 경량 모델 (~3~5원)
- **총 비용**: ~3.5~5.5원/건
- **응답 시간**: 2~5초

---

### 파이프라인 5: 문항 분석

```
[입력: 자소서 문항 + 기업 분석 결과 + 사용자 무기 보유 현황]
    │
    ▼
[question_patterns 조회] ← 문항 키워드/정규식 매칭으로 패턴 탐지
    │
    ▼
[prompt_templates 로드] ← coaching_draft/question_analysis
    │
    ▼
[변수 치환]
    ├── {{question_text}} ← 자소서 문항
    ├── {{company_analysis}} ← 기업 분석 결과
    ├── {{talent_profile}} ← 인재상
    ├── {{char_limit}} ← 글자수 제한
    └── {{user_weapons_summary}} ← 보유 무기 요약
    │
    ▼
[Claude Sonnet 4.5 호출]
    │
    ▼
[문항 분석 결과]
    ├── surface_question: string        ← 표면적 질문
    ├── real_intent: string             ← 진짜 평가 의도
    ├── required_weapons: [{code, reason}] ← 필요 무기
    ├── writing_structure: {intro, body, conclusion, sections}
    ├── key_keywords: string[]          ← 핵심 키워드
    ├── avoid_list: string[]            ← 피해야 할 것
    └── good_example_structure: string  ← 좋은 답변 구조
```

- **모델**: Claude Sonnet 4.5
- **비용**: ~65원/건
- **응답 시간**: 3~8초

---

### 파이프라인 6: 초안 코칭

```
[입력: 문항 분석 + 선택 경험(STAR) + 기업 분석 + 인재상]
    │
    ▼
[prompt_templates 로드] ← coaching_draft/weapon_enhance
    │
    ▼
[변수 치환]
    ├── {{experience_star}} ← 선택 경험의 STAR 구조
    ├── {{weapon_name}}, {{weapon_code}} ← 주 무기
    ├── {{company_analysis}} ← 기업 분석
    ├── {{talent_profile}} ← 인재상
    ├── {{question_text}} ← 문항
    └── {{char_limit}} ← 글자수
    │
    ▼
[Claude Sonnet 4.5 스트리밍 호출]
    │ Go Backend → SSE (Server-Sent Events)
    ▼
[스트리밍 코칭 출력]
    ├── 1. 역량 포인트 (이 경험에서 드러나는 핵심 역량)
    ├── 2. 구체성 보강 제안 (수치, 에피소드, 감정 추가)
    ├── 3. 인재상 연결 포인트
    ├── 4. STAR 약한 부분 보강 질문
    ├── 5. 추상적→구체적 표현 변환 예시
    └── 6. 글자수 배분에 맞춘 초안 구조
    │
    ▼
[coaching_sessions 저장] ← 토큰/비용 로깅
[cover_letter_versions 저장] ← 초안 버전 생성
```

- **모델**: Claude Sonnet 4.5
- **비용**: ~65원/건
- **응답 시간**: 5~15초 (스트리밍)

---

### 파이프라인 7: 첨삭 코칭

```
[입력: 작성된 자소서 + 문항 분석 + 기업 분석]
    │
    ▼
[Claude Sonnet 4.5 스트리밍 호출]
    │
    ▼
[첨삭 결과 (스트리밍)]
    │
    ├── 4개 지표 점수 (각 0~100)
    │   ├── specificity: 구체성 (수치, 에피소드, 감정이 있는가)
    │   ├── job_fit: 직무적합성 (직무 역량과 연결되는가)
    │   ├── company_fit: 기업맞춤 (인재상/핵심가치가 녹아있는가)
    │   └── authenticity: 진정성 (AI스러움 없이 자연스러운가)
    │
    ├── 문장별 피드백
    │   └── [{sentence, issue, suggestion, category}]
    │
    ├── 전체 개선 제안
    │   └── [priority순 개선사항]
    │
    └── 강점 분석
        └── [잘 쓴 부분과 이유]
    │
    ▼
[coaching_sessions 저장] ← 토큰/비용 로깅
[cover_letter_versions 저장] ← 첨삭 결과 + 점수 저장
```

- **모델**: Claude Sonnet 4.5
- **비용**: ~65원/건
- **응답 시간**: 5~15초 (스트리밍)

---

## 5. 비용 모니터링

### coaching_sessions 기반 토큰 로깅

**Go 백엔드에서 AI 호출 후 자동 로깅**:

```go
// apps/backend/internal/service/coaching.go
func (s *CoachingService) GenerateDraft(ctx context.Context, req DraftRequest) (*DraftResponse, error) {
    // 1. AI 호출 (Claude Sonnet 4.5)
    startTime := time.Now()
    stream, usage, err := s.aiProvider.StreamHeavy(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // 2. 스트리밍 응답을 클라이언트로 전송 (SSE)
    // ... SSE 처리 ...

    // 3. 완료 후 토큰 로깅
    _, err = s.entClient.CoachingSession.Create().
        SetUserID(userID).
        SetCoverLetterID(req.CoverLetterID).
        SetSessionType("draft").
        SetPromptTemplateID(promptID).
        SetInputData(req). // JSON
        SetModelUsed("claude-sonnet-4-5").
        SetInputTokens(usage.InputTokens).
        SetOutputTokens(usage.OutputTokens).
        SetTotalCostKrw(calculateCost("claude-sonnet-4-5", usage)).
        SetLatencyMs(time.Since(startTime).Milliseconds()).
        Save(ctx)

    return &DraftResponse{...}, nil
}
```

### 비용 계산 함수

```typescript
function calculateCost(model: string, usage: { promptTokens: number; completionTokens: number }): number {
  const rates: Record<string, { input: number; output: number }> = {
    'gemini-2.0-flash':   { input: 0.0001,  output: 0.0004 },  // 원/1K tokens
    'llama-3.3-70b':      { input: 0.00084, output: 0.00084 }, // 원/1K tokens (Groq)
    'claude-sonnet-4-5':  { input: 0.042,   output: 0.21 },    // 원/1K tokens
    'text-embedding-3-small': { input: 0.000028, output: 0 },
  };

  const rate = rates[model];
  if (!rate) return 0;

  return (
    (usage.promptTokens / 1000) * rate.input +
    (usage.completionTokens / 1000) * rate.output
  );
}
```

---

## 6. 캐싱 전략

| 대상 | TTL | 캐시 키 | 무효화 조건 |
|------|-----|---------|------------|
| 기업 분석 결과 | 7일 | URL 해시 + 기업명 | TTL 만료 또는 수동 무효화 |
| 채용공고 파싱 | 7일 | URL 해시 | TTL 만료 |
| DART 기업 정보 | 7일 | 기업 코드 | TTL 만료 |
| 코칭 결과 | 캐시 안함 | - | 매번 새로 생성 (개인화) |
| 경험 분류 | 캐시 안함 | - | 경험 수정 시 재분류 |
| 무기/프롬프트 마스터 | 앱 시작 시 | - | 관리자 수정 시 |

### 캐시 구현

```typescript
async function getCachedOrFetch<T>(
  cacheKey: string,
  cacheType: string,
  fetcher: () => Promise<T>,
  ttlDays: number = 7
): Promise<T> {
  const supabase = await createClient();

  // 캐시 조회
  const { data: cached } = await supabase
    .from('company_analysis_cache')
    .select('data')
    .eq('cache_key', cacheKey)
    .eq('cache_type', cacheType)
    .gt('expires_at', new Date().toISOString())
    .single();

  if (cached) return cached.data as T;

  // 캐시 미스 → 데이터 fetch
  const result = await fetcher();

  // 캐시 저장
  await supabase.from('company_analysis_cache').upsert({
    cache_key: cacheKey,
    cache_type: cacheType,
    data: result,
    expires_at: new Date(Date.now() + ttlDays * 24 * 60 * 60 * 1000).toISOString(),
  });

  return result;
}
```

---

## 7. 전체 비용 요약 (1건 풀 파이프라인)

```
┌─────────────────────────────────────────────────┐
│  1건 자소서 완성 파이프라인 총 비용              │
│                                                 │
│  공고 파싱 (Gemini/Groq)      :   ~3~5원        │
│  기업 종합 분석 (Claude)      :    ~65원        │
│  경험 매칭 (임베딩+Gemini)    : ~3.5~5.5원      │
│  문항 분석 (Claude)           :    ~65원        │
│  초안 코칭 (Claude)           :    ~65원        │
│  ──────────────────────────────────────         │
│  합계                         : ~201~206원      │
│                                                 │
│  ※ 첨삭 코칭 추가 시: +~65원                    │
│  ※ 캐시 HIT 시 기업 분석 0원                    │
└─────────────────────────────────────────────────┘
```
