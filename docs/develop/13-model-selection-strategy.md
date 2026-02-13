# AI 모델 선택 전략

> 작성일: 2026-02-13

---

## 1. 모델 티어링 및 선택 가능 구조

### 모델 분류

| 계층 | 기본 모델 | 대체 모델 | 용도 | 선택 방식 |
|------|---------|---------|------|----------|
| **경량 (Light)** | Gemini Flash | Groq Llama | 파싱, 태깅, 인터뷰 | 환경변수 (LLM_LIGHT_PROVIDER) |
| **고급 (Heavy)** | Claude Sonnet 4.5 | Gemini Pro, GPT-4 | 기업 분석, 코칭 | **프롬프트 템플릿별 설정** |
| **임베딩** | text-embedding-3-small | - | 벡터 검색 | 고정 |

### 핵심 설계: **프롬프트 템플릿 기반 모델 선택**

```sql
prompt_templates 테이블:
├─ category (experience_classify, company_analysis, coaching_draft 등)
├─ sub_category (weapon_tagging, talent_analysis 등)
├─ model VARCHAR(50) ← "claude-sonnet-4-5", "gemini-2.0-flash", "gpt-4"
├─ temperature
├─ max_tokens
└─ is_active
```

**어드민이 프롬프트별로 모델 선택:**
- 경험 태깅: `gemini-2.0-flash` (저렴)
- 인재상 분석: `claude-sonnet-4-5` (품질)
- 초안 코칭: `claude-sonnet-4-5` (한국어)

---

## 2. AI Provider 통합 구조

### Before (현재):
```go
AIProvider {
    light: LLMProvider  // Gemini or Groq
}
```

### After (목표):
```go
AIProvider {
    light: LLMProvider       // Gemini or Groq
    heavy: LLMProvider       // Claude, Gemini Pro, GPT-4
    embedding: EmbeddingProvider
}

// 프롬프트 템플릿의 model 필드에 따라 자동 선택
func (p *AIProvider) CallByModelName(model string, req LLMRequest) {
    switch model {
    case "gemini-2.0-flash", "gemini-flash":
        return p.light.Call(req)
    case "claude-sonnet-4-5", "claude":
        return p.heavy.Call(req)
    case "gpt-4":
        return p.gpt.Call(req)  // Future
    }
}
```

---

## 3. 모델별 사용 예시 (어드민 설정)

### 경량 작업 (Gemini Flash - 기본):
```
category: experience_classify
sub_category: weapon_tagging
model: gemini-2.0-flash
temperature: 0.2
비용: ~3원/건
```

### 심층 분석 (Claude Sonnet - 한국어 품질):
```
category: company_analysis
sub_category: talent_analysis
model: claude-sonnet-4-5
temperature: 0.3
비용: ~65원/건
```

### 코칭 (Claude Sonnet - 긴 컨텍스트):
```
category: coaching_draft
sub_category: weapon_enhance
model: claude-sonnet-4-5
temperature: 0.5
비용: ~65원/건
```

### 실험 (Gemini Pro - 중간):
```
category: coaching_draft
sub_category: revision
model: gemini-2.0-pro  ← 어드민이 실험적으로 변경 가능
temperature: 0.4
비용: ~30원/건
```

---

## 4. 어드민 UI 설계 (Phase 10)

### 프롬프트 관리 페이지 (`/admin/prompts`)

**기능:**
1. 프롬프트 템플릿 목록
2. 프롬프트 편집 모달
   - System prompt 편집
   - User prompt template 편집
   - **모델 선택 드롭다운** ← 핵심
   - Temperature 슬라이더
   - Max tokens 입력
3. 사용 통계
   - 호출 횟수 (usage_count)
   - 평균 응답 시간 (avg_latency_ms)
   - 평균 품질 점수 (avg_quality_score)
4. A/B 테스트 지원
   - 같은 category에 여러 버전 활성화
   - 랜덤 분배 또는 사용자 그룹별 분배

### 모델 선택 UI 예시:
```
┌─────────────────────────────────────┐
│ 프롬프트 설정                         │
├─────────────────────────────────────┤
│ Category: company_analysis          │
│ Sub-category: talent_analysis       │
│                                     │
│ AI Model: [▼ 선택]                  │
│   ○ claude-sonnet-4-5 (권장)        │
│     - 한국어 분석 품질 최상          │
│     - 비용: ~65원/건                │
│                                     │
│   ○ gemini-2.0-pro                  │
│     - 중간 품질, 저렴함              │
│     - 비용: ~30원/건                │
│                                     │
│   ○ gemini-2.0-flash                │
│     - 빠름, 매우 저렴                │
│     - 비용: ~3원/건                 │
│                                     │
│ Temperature: [====|----] 0.3        │
│ Max Tokens: [2000]                  │
│                                     │
│ [미리보기] [저장]                    │
└─────────────────────────────────────┘
```

---

## 5. Config 구조 업데이트

### 환경변수:
```bash
# Light model provider
LLM_LIGHT_PROVIDER=gemini  # or groq

# API Keys
GEMINI_API_KEY=...
GROQ_API_KEY=...
ANTHROPIC_API_KEY=...      # ← 추가
OPENAI_API_KEY=...         # embeddings

# Optional: Default heavy model (프롬프트 미지정 시)
LLM_HEAVY_DEFAULT=claude-sonnet-4-5
```

### Config 구조체:
```go
type Config struct {
    // Light model
    LLMLightProvider string
    GeminiAPIKey     string
    GroqAPIKey       string

    // Heavy model (추가)
    AnthropicAPIKey  string

    // Embedding
    OpenAIAPIKey     string
}
```

---

## 6. 서비스 레이어에서 사용

### 기존 (WeaponTaggingService):
```go
// Prompt template에서 model 로드
promptTemplate := loadPrompt("experience_classify", "weapon_tagging")
// model = "gemini-2.0-flash"

// AIProvider가 자동으로 적절한 provider 선택
resp := aiProvider.CallByModelName(promptTemplate.Model, request)
```

### 새로운 (CompanyAnalysisService):
```go
// Prompt template에서 model 로드
promptTemplate := loadPrompt("company_analysis", "talent_analysis")
// model = "claude-sonnet-4-5"

// AIProvider가 Claude 사용
resp := aiProvider.CallByModelName(promptTemplate.Model, request)
```

---

## 7. 구현 우선순위

### Phase 3.2 (지금):
1. ✅ Claude provider 구현
2. ✅ AIProvider에 heavy field 추가
3. ✅ CallByModelName() 메서드 구현
4. ⬜ CompanyAnalysisService (캐싱 포함)

### Phase 10 (나중):
5. ⬜ 어드민 프롬프트 관리 UI
6. ⬜ 모델 선택 드롭다운
7. ⬜ A/B 테스트 지원
8. ⬜ 사용 통계 대시보드

---

## 8. 마이그레이션 경로

**단계적 전환:**
1. Phase 3.2: Claude 추가, company_analysis만 Claude 사용
2. Phase 5.1: 코칭도 Claude 사용
3. Phase 10: 어드민 UI에서 모든 프롬프트 모델 변경 가능

**하위 호환성:**
- 기존 Gemini 파이프라인 유지
- 새 기능만 Claude 사용
- 점진적 확장
