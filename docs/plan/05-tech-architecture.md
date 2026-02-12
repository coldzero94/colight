# Colight - 기술 아키텍처

> 작성일: 2026-02-09

---

## 1. 인프라 구성

```
[사용자 브라우저]
    │
    ▼
[Vercel - Next.js 15]
    │  - App Router / React Server Components
    │  - AI SDK (스트리밍 응답)
    │  - API Routes (10초 이내 작업)
    │
    ├──▶ [Supabase]
    │      - PostgreSQL: 사용자, 경험, 자소서, 기업분석 캐시
    │      - Auth: 이메일/소셜 로그인
    │      - Row Level Security
    │
    ├──▶ [Koyeb - Go API + River (embedded)]
    │      - Gin API 서버
    │      - River embedded: 백그라운드 작업 큐 (별도 Worker 불필요)
    │      - Playwright 동적 크롤링 (원티드 등, River 비동기)
    │      - 배치 처리 (무거운 AI 분석)
    │
    └──▶ [외부 API]
           - Claude/GPT API (AI 분석/코칭)
           - DART OpenAPI (기업 정보)
           - 네이버 뉴스 API (기업 뉴스)
           - 사람인 API (채용 공고)
```

---

## 2. 핵심 데이터 모델

```
user_profiles
  ├── experiences (경험 관리)
  │     ├── experience_tags (역량 태그)
  │     ├── experience_weapons (경험-무기 매핑)
  │     └── experience_usages (경험 사용 이력)
  │
  ├── applications (지원 현황)
  │     ├── cover_letters (자소서)
  │     │     └── cover_letter_versions (버전 관리)
  │     └── company_analyses (기업 분석 결과)
  │
  └── coaching_sessions (코칭 대화 이력)

company_analysis_cache (기업 분석 캐시, 7일 TTL)
talent_profiles (주요 기업 인재상 사전 DB)
weapon_categories (7대 무기 마스터 데이터)
prompt_templates (AI 프롬프트 템플릿)
question_patterns (공통 문항 패턴 DB)
```

---

## 3. API 비용 전략

| 작업 | 모델 | 건당 비용 |
|------|------|----------|
| 공고 파싱 + 구조화 | GPT-4.1 mini | ~5원 |
| 기업 종합 분석 | Claude Sonnet 4.5 | ~65원 |
| 경험 매칭 | GPT-4.1 mini | ~5원 |
| 자소서 초안 코칭 | Claude Sonnet 4.5 | ~65원 |
| 자소서 첨삭 | Claude Sonnet 4.5 | ~65원 |
| 경험 인터뷰 (대화형) | GPT-4.1 mini | ~5원/턴 |
| 임베딩 생성 | text-embedding-3-small | ~0.5원 |

**1건 풀 분석 (공고 파싱 → 기업 분석 → 매칭 → 코칭)**: 약 140~200원
