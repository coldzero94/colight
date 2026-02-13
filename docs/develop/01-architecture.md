# 시스템 아키텍처

> 작성일: 2026-02-11

---

## 1. 전체 시스템 다이어그램

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              사용자 브라우저                              │
│                        (React, Next.js CSR)                            │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │ HTTPS
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     Vercel (Next.js 15 App Router)                     │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │  Pages/SSR   │  │  Generated   │  │  Components  │  │ Middleware  │  │
│  │  (RSC)       │  │  API Client  │  │  (shadcn/ui) │  │ (Auth)     │  │
│  └─────────────┘  └──────┬───────┘  └──────────────┘  └────────────┘  │
│                          │                                            │
│  ┌───────────────────────┘                                            │
│  │  src/api/generated/  ← @hey-api/openapi-ts 생성                    │
│  │  src/lib/api/        ← Axios 클라이언트 래퍼                        │
│  └───────────────────────┬────────────────────────────────────────── │
└──────────────────────────┼──────────────────────────────────────────────┘
                           │ REST API (TypeSpec contract)
                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      Koyeb (Go API Server)                             │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │  Controller  │  │   Service    │  │Infrastructure│  │ Middleware  │  │
│  │  (Gin HTTP)  │  │  (비즈니스)   │  │ (DB, AI)     │  │ (JWT)      │  │
│  └─────────────┘  └──────────────┘  └──────┬───────┘  └────────────┘  │
│                                            │                          │
│  ┌─────────────────────────────────────────┘                          │
│  │  internal/generated/  ← oapi-codegen 생성                          │
│  │  ent/schema/          ← Ent ORM 스키마 (15 tables)                 │
│  └─────────────────────────────────────────────────────────────────── │
└──────────────────────────┼──────────────────────────────────────────────┘
                           │ SQL
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    Supabase PostgreSQL + pgvector                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │ 15 Tables    │  │ pgvector     │  │ Auth (JWT)   │  │ Edge Funcs │  │
│  │ (Ent 관리)   │  │ (벡터 검색)   │  │ (토큰 발급)   │  │ (캐시 정리) │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘

External Services (Go에서 직접 호출):
┌──────────────────┐ ┌──────────────┐
│   AI Providers   │ │ 기업 데이터    │
│                  │ │              │
│ • Claude API     │ │ • DART API   │
│ • Gemini API     │ │ • 네이버 API │
│ • Groq API       │ │ • 사람인 API │
│ • Embedding API  │ │              │
└──────────────────┘ └──────────────┘

Async Jobs (Go API 내부, River embedded):
┌─────────────────────────────────────┐
│  River Job Queue (PostgreSQL 기반)   │
│  • Playwright 크롤링 (원티드, Phase 10)│
│  • 대용량 임베딩 생성                 │
│  • 백그라운드 분석 작업               │
│  → MaxWorkers: 1 (메모리 제약)       │
└─────────────────────────────────────┘
```

---

## 2. 모노레포 구조

```text
colight/
├── apps/
│   ├── backend/                    # Go API 서버
│   │   ├── cmd/
│   │   │   └── api/main.go        # 엔트리포인트 (Gin 서버 설정)
│   │   ├── internal/
│   │   │   ├── controller/        # HTTP 핸들러 (StrictServerInterface 구현)
│   │   │   ├── service/           # 비즈니스 로직
│   │   │   │   ├── experience.go  # ExperienceService
│   │   │   │   ├── analysis.go    # AnalysisService
│   │   │   │   ├── coaching.go    # CoachingService
│   │   │   │   ├── matching.go    # MatchingService
│   │   │   │   ├── crawling.go    # CrawlingService
│   │   │   │   └── auth.go        # AuthService
│   │   │   ├── infrastructure/    # 인프라 레이어
│   │   │   │   ├── config/        # 환경 설정
│   │   │   │   ├── database/      # DB 연결
│   │   │   │   ├── middleware/     # JWT 검증, CORS, 로깅
│   │   │   │   ├── ai/            # 공통 LLM (Gemini/Groq), Claude, Embedding 클라이언트
│   │   │   │   └── external/      # DART, 네이버, 사람인, Koyeb 클라이언트
│   │   │   └── generated/         # oapi-codegen 생성 코드
│   │   ├── ent/                   # Ent ORM
│   │   │   ├── schema/            # DB 스키마 정의 (15개)
│   │   │   └── migrate/           # Atlas 마이그레이션
│   │   ├── migrations/            # Atlas SQL 마이그레이션 파일
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   └── web/                       # Next.js 15 프론트엔드
│       ├── src/
│       │   ├── app/               # App Router (페이지)
│       │   │   ├── layout.tsx
│       │   │   ├── page.tsx       # 랜딩 페이지
│       │   │   ├── (auth)/        # 인증 (login, signup, callback)
│       │   │   └── (main)/        # 메인 앱 (dashboard, experiences, analysis, coaching)
│       │   ├── api/
│       │   │   └── generated/     # @hey-api/openapi-ts 생성 코드
│       │   │       ├── types.gen.ts    # API 타입
│       │   │       ├── sdk.gen.ts      # API SDK
│       │   │       └── zod.gen.ts      # Zod 스키마
│       │   ├── lib/
│       │   │   └── api/           # Axios 클라이언트 래퍼 (Generated SDK 사용)
│       │   ├── components/        # UI 컴포넌트
│       │   │   ├── ui/            # shadcn/ui
│       │   │   ├── layout/        # 사이드바, 헤더
│       │   │   ├── experience/    # 경험 관련
│       │   │   ├── analysis/      # 분석 관련
│       │   │   ├── coaching/      # 코칭 관련
│       │   │   └── dashboard/     # 대시보드 관련
│       │   ├── hooks/             # React Query 훅
│       │   └── stores/            # Zustand 스토어
│       ├── package.json
│       └── next.config.ts
├── packages/
│   └── protocol/                  # TypeSpec API 정의 (Single Source of Truth)
│       ├── src/
│       │   ├── main.tsp           # 엔트리포인트
│       │   ├── auth.tsp           # 인증 API
│       │   ├── experiences.tsp    # 경험 CRUD API
│       │   ├── analysis.tsp       # 기업 분석 API
│       │   ├── coaching.tsp       # 코칭 API
│       │   ├── matching.tsp       # 매칭 API
│       │   ├── applications.tsp   # 지원 관리 API
│       │   └── models/            # 공통 모델 정의
│       ├── tsp-output/
│       │   └── openapi/
│       │       └── openapi.yaml   # 생성된 OpenAPI 스펙
│       ├── tspconfig.yaml
│       └── package.json
├── docs/
│   ├── plan/                      # 기획 문서
│   └── develop/                   # 개발 문서 + Phase 가이드
│       └── phases/
├── supabase/
│   ├── migrations/                # Supabase SQL 마이그레이션 (참고용)
│   └── seed.sql                   # 시드 데이터
├── CLAUDE.md
├── package.json                   # 워크스페이스 루트
└── pnpm-workspace.yaml
```

---

## 3. 백엔드 아키텍처 (Go)

### 3.1 엔트리포인트

```text
cmd/api/main.go
    │
    ├── 환경 설정 로드 (config)
    ├── DB 연결 (Ent 클라이언트)
    ├── AI 클라이언트 초기화 (Claude, OpenAI, Embedding)
    ├── External 클라이언트 초기화 (DART, 네이버, Koyeb)
    ├── Service 레이어 생성
    ├── Controller 레이어 생성
    ├── Middleware 설정 (JWT, CORS, 로깅)
    ├── Router 설정 (Gin + oapi-codegen StrictHandler)
    └── 서버 시작 (Graceful Shutdown)
```

### 3.2 레이어 구조

```text
[HTTP Request]
    │
    ▼
[Middleware] ← JWT 검증, CORS, Request ID, 로깅
    │
    ▼
[Controller] ← StrictServerInterface 구현, 요청/응답 변환
    │
    ▼
[Service] ← 비즈니스 로직, 트랜잭션 관리
    │
    ├── [Ent ORM] ← DB 쿼리, 마이그레이션
    ├── [AI Client] ← Claude/OpenAI 호출, 스트리밍
    └── [External Client] ← DART, 네이버, Koyeb
```

### 3.3 Ent ORM 스키마 (15개)

| 스키마 | 설명 | Phase |
|--------|------|-------|
| `UserProfile` | 사용자 프로필 (목표 직무, 산업, 크레딧) | Phase 1 |
| `Experience` | 경험 (STAR 구조, 임베딩 벡터) | Phase 2 |
| `ExperienceTag` | 경험 태그 (STAR, 키워드) | Phase 2 |
| `ExperienceWeapon` | 경험-무기 매핑 (주/부, confidence) | Phase 2.1 |
| `ExperienceUsage` | 경험 활용 이력 | Phase 10 |
| `WeaponCategory` | 7대 무기 마스터 데이터 | Phase 0 |
| `PromptTemplate` | AI 프롬프트 템플릿 (DB 관리) | Phase 0 |
| `QuestionPattern` | 공통 문항 패턴 DB | Phase 0 |
| `Application` | 지원 현황 (상태, 마감일) | Phase 3 |
| `CompanyAnalysis` | 기업 분석 결과 | Phase 3.2 |
| `CompanyAnalysisCache` | 분석 캐시 (7일 TTL) | Phase 3 |
| `TalentProfile` | 대기업 인재상 DB | Phase 10 |
| `CoverLetter` | 자소서 문항 | Phase 5 |
| `CoverLetterVersion` | 자소서 버전 (내용, 점수, 피드백) | Phase 5.1 |
| `CoachingSession` | 코칭 세션 로그 | Phase 5.1 |

### 3.4 Service 레이어

| Service | 역할 | 주요 메서드 |
|---------|------|-----------|
| `AuthService` | 인증 처리 | JWT 검증, 사용자 조회 |
| `ExperienceService` | 경험 CRUD + 태깅 | Create, List, Get, Update, Delete, Tag |
| `AnalysisService` | 기업 분석 파이프라인 | Analyze (크롤링→파싱→데이터→AI) |
| `CoachingService` | 코칭 (문항분석, 초안, 첨삭) | QuestionAnalysis, Draft, Review |
| `MatchingService` | 경험-공고 벡터 매칭 | Match (pgvector cosine similarity) |
| `CrawlingService` | 공고 크롤링/파싱 | Crawl (Cheerio/Koyeb 분기) |

### 3.5 Infrastructure 레이어

| 패키지 | 역할 |
|--------|------|
| `config/` | 환경 변수 로드, 설정 구조체 |
| `database/` | Ent 클라이언트 생성, DB 연결 관리 |
| `middleware/` | JWT 검증, CORS, Request ID, HTTP 로깅, Rate Limiting |
| `ai/` | 공통 LLM 인터페이스 (Gemini/Groq), Claude 클라이언트 (스트리밍), Embedding 클라이언트 |
| `external/` | DART API, 네이버 뉴스 API, 사람인 API, Koyeb Worker 트리거 |

---

## 4. 프론트엔드 아키텍처 (Next.js 15)

### 4.1 App Router 구조

```text
src/app/
├── layout.tsx                    # Root Layout (Providers, 폰트)
├── page.tsx                      # 랜딩 페이지
├── (auth)/
│   ├── layout.tsx                # Auth 레이아웃 (비로그인 전용)
│   ├── login/page.tsx            # 로그인
│   ├── signup/page.tsx           # 회원가입
│   └── callback/route.ts         # OAuth 콜백
├── (main)/
│   ├── layout.tsx                # Main 레이아웃 (사이드바, 네비게이션)
│   ├── dashboard/
│   │   ├── page.tsx              # 대시보드 메인 (칸반보드)
│   │   └── loading.tsx           # 스켈레톤 로딩
│   ├── experiences/
│   │   ├── page.tsx              # 경험 목록
│   │   ├── new/page.tsx          # 경험 등록
│   │   ├── [id]/page.tsx         # 경험 상세
│   │   ├── interview/page.tsx    # AI 인터뷰
│   │   └── loading.tsx
│   ├── analysis/
│   │   ├── page.tsx              # 기업 분석 메인
│   │   ├── [id]/page.tsx         # 분석 결과 상세
│   │   └── loading.tsx
│   ├── coaching/
│   │   ├── page.tsx              # 코칭 메인
│   │   ├── [id]/page.tsx         # 코칭 세션
│   │   └── loading.tsx
│   ├── pricing/page.tsx          # 가격표
│   └── settings/page.tsx         # 설정
└── (no api/ directory)           # API Routes 없음 — Go 백엔드로 이관
```

### 4.2 API 클라이언트 구조

```text
src/api/generated/                # @hey-api/openapi-ts 자동 생성
├── types.gen.ts                  # API 요청/응답 타입
├── sdk.gen.ts                    # API SDK (함수 기반)
└── zod.gen.ts                    # Zod 스키마 (런타임 검증)

src/lib/api/
├── client.ts                     # Axios 인스턴스 설정 (baseURL, 인터셉터)
├── auth.ts                       # 인증 헤더 주입 (Supabase JWT → Authorization)
└── index.ts                      # Generated SDK 래퍼 (에러 핸들링 통합)
```

---

## 5. 주요 요청 흐름

### 5.1 인증 흐름

```
[사용자] → 로그인/회원가입 요청
    │
    ▼
[Next.js Middleware]
    │ 세션 확인 (Supabase Auth)
    │
    ├── [세션 없음] → (auth)/login 페이지 렌더링
    │       │
    │       ▼
    │   [Supabase Auth] ← 이메일/소셜 로그인 처리
    │       │
    │       ▼
    │   [JWT 토큰 발급] ← Supabase에서 JWT 발급
    │       │
    │       ▼
    │   [/(main)/dashboard 리다이렉트]
    │
    └── [세션 있음] → API 요청 시 JWT 포함
            │
            ▼
        [Go API Server] ← Authorization: Bearer <JWT>
            │
            ▼
        [JWT Middleware] ← Supabase JWT 서명 검증
            │
            ▼
        [Controller → Service] ← user_id 추출하여 비즈니스 로직 실행
```

### 5.2 경험 등록 + 자동 태깅 흐름

```
[사용자] → 경험 입력 (제목, 기간, 역할, 내용, 성과)
    │
    ▼
[Next.js] → POST /v1/experiences (Generated SDK)
    │
    ▼
[Go API Server]
    │
    ├── [Controller] → 요청 검증 (oapi-codegen)
    │
    ├── [ExperienceService.Create]
    │       │
    │       ├── 1. DB 저장 (Ent → experiences 테이블)
    │       │
    │       ├── 2. 임베딩 생성 (OpenAI text-embedding-3-small → vector(1536))
    │       │       └── experiences.embedding 컬럼 업데이트
    │       │
    │       └── 3. 무기 자동 태깅 (비동기)
    │               │
    │               ▼
    │           [prompt_templates 로드] → category='experience_classify'
    │               │
    │               ▼
    │           [weapon_categories 로드] → 전체 무기 목록 조회
    │               │
    │               ▼
    │           [경량 모델 (Gemini/Groq) 호출] → {{experience_text}} + {{weapon_categories}} 주입
    │               │
    │               ▼
    │           [결과 파싱]
    │               ├── experience_weapons 저장 (주 무기 + 부 무기, confidence 점수)
    │               └── experience_tags 저장 (STAR 구조, 키워드)
    │
    └── [응답] → { data: Experience }
```

### 5.3 기업 분석 파이프라인 (URL → 파싱 → 데이터 → AI)

```
[사용자] → 채용공고 URL 입력
    │
    ▼
[Next.js] → POST /v1/analysis (Generated SDK)
    │
    ▼
[Go API Server - AnalysisService]
    │
    ├── 1. 캐시 확인 (company_analysis_cache, 7일 TTL)
    │       ├── [캐시 HIT] → 캐시 결과 즉시 반환
    │       └── [캐시 MISS] → 계속 진행
    │
    ├── 2. URL 도메인 판별 → 크롤링 전략 결정
    │       ├── [정적: 잡코리아/캐치] → Go에서 Cheerio 스타일 파싱 (goquery)
    │       ├── [동적: 원티드]       → Koyeb Worker에 Playwright 요청 트리거
    │       └── [사람인]             → 사람인 API 직접 호출
    │
    ├── 3. 공고 파싱 (경량 모델 (Gemini/Groq), ~5원)
    │       └── 직무, 자격요건, 우대사항, 키워드 구조화
    │
    ├── 4. 기업 정보 병렬 수집 (Go goroutine)
    │       ├── [DART OpenAPI]     → 기업 기본정보, 재무제표
    │       ├── [네이버 뉴스 API]  → 최근 뉴스 5~10건
    │       └── [talent_profiles]  → 사전 DB 인재상 (있으면)
    │
    ├── 5. AI 종합 분석 (Claude Sonnet 4.5, SSE 스트리밍, ~65원)
    │       └── 인재상 추론 + 핵심가치 + 전략 키워드 + 피해야 할 표현
    │
    ├── 6. 결과 캐시 저장 (company_analysis_cache)
    │
    └── 7. SSE 스트리밍 응답 → 프론트엔드에서 점진적 표시
```

### 5.4 자소서 코칭 (스트리밍)

```
[사용자] → 자소서 문항 + 선택한 경험 + 기업 분석 결과
    │
    ▼
[Next.js] → POST /v1/coaching/question-analysis (Generated SDK)
    │
    ▼
[Go API Server - CoachingService.QuestionAnalysis]
    │ Claude Sonnet 4.5 → 문항 의도, 필요 무기, 작성 구조(글자수 배분), 핵심 키워드
    │
    ▼
[Next.js] → POST /v1/coaching/draft (SSE 연결)
    │
    ▼
[Go API Server - CoachingService.Draft]
    │
    ├── [prompt_templates 로드] → coaching_draft/weapon_enhance
    ├── [변수 주입] → {{experience_star}}, {{company_analysis}}, {{question_text}}, {{char_limit}}
    ├── [Claude Sonnet 4.5 스트리밍 호출]
    │       │
    │       ▼
    │   [Go SSE Writer]
    │       │ Server-Sent Events
    │       ▼
    │   [사용자 브라우저] ← EventSource로 실시간 텍스트 수신 → Tiptap 에디터 표시
    │
    └── [coaching_sessions 저장] → 토큰 사용량, 비용 로깅
    │
    ▼
[Next.js] → POST /v1/coaching/review (SSE 연결)
    │
    ▼
[Go API Server - CoachingService.Review]
    │ 구체성/직무적합/기업맞춤/진정성 4개 지표 점수 + 개선 제안
    │
    └── [cover_letter_versions 저장] → 버전 이력 관리
```

---

## 6. TypeSpec → 코드 생성 파이프라인

```
packages/protocol/src/*.tsp         ← API 정의 (Single Source of Truth)
    │
    ▼ tsp compile
    │
packages/protocol/tsp-output/openapi/openapi.yaml
    │
    ├──────────────────────────────────────────────┐
    │                                              │
    ▼ oapi-codegen (Go)                            ▼ @hey-api/openapi-ts (TS)
    │                                              │
apps/backend/internal/generated/                   apps/web/src/api/generated/
├── api.gen.go                                     ├── types.gen.ts
│   ├── StrictServerInterface                      ├── sdk.gen.ts
│   ├── Request/Response types                     └── zod.gen.ts
│   └── Gin route registration
└── spec.gen.go
```

### 생성 명령어

```bash
# 전체 생성 (TypeSpec → OpenAPI → Go/TS)
pnpm run generate

# TypeSpec → OpenAPI만
pnpm run generate:protocol

# OpenAPI → Go 서버 코드
pnpm run generate:api:go

# OpenAPI → TypeScript (타입 + SDK + Zod)
pnpm run generate:api:ts
```

### 코드 생성 규칙

| 항목 | 정책 |
|------|------|
| 생성된 코드 | Git에 커밋 (코드 리뷰 가능, CI 생성 단계 불필요) |
| API 변경 시 | TypeSpec 먼저 수정 → `pnpm run generate` → Go/TS 코드 자동 반영 |
| 수동 수정 | 생성된 코드 직접 수정 금지 (재생성 시 덮어쓰기됨) |
| Controller 구현 | `StrictServerInterface` 메서드 구현으로 타입 안전 보장 |

---

## 7. 기술 선택 근거

| 기술 | 선택 이유 | 대안 (미선택 이유) |
|------|----------|------------------|
| **Go 백엔드 분리** | 타입 안전, DB 마이그레이션 직접 제어, API 버전 관리 용이, 테스트 용이, Go의 동시성(goroutine)으로 AI 병렬 호출 최적 | Next.js API Routes (관심사 분리 어려움, 60초 제한, 스케일링 제약) |
| **Ent ORM** | 타입 안전 ORM, Atlas 마이그레이션 자동 생성, 그래프 쿼리 지원 | GORM (마이그레이션 불안정), sqlc (ORM 기능 부족) |
| **TypeSpec** | API 계약의 Single Source of Truth, Go/TS 동시 생성, 스키마 일관성 보장 | OpenAPI 수동 작성 (동기화 어려움), gRPC (브라우저 호환성) |
| **oapi-codegen** | Go 표준, StrictServerInterface 패턴, Gin 통합 우수 | ogen (생태계 작음), go-swagger (무거움) |
| **@hey-api/openapi-ts** | Zod v4 네이티브 지원, 타입+SDK+검증 단일 도구, Java 의존성 없음 | openapi-generator-cli (Java 필요, 커스텀 어려움) |
| **Koyeb (Go 서버)** | 무료 512MB RAM, Docker 배포, Scale-to-Zero | Railway (무료 제한적), Fly.io (메모리 이슈) |
| **Next.js 15 App Router** | SSR/RSC 지원, Vercel 최적 배포, 스트리밍 네이티브 지원 | Remix (Vercel 최적화 부족), SvelteKit (생태계 작음) |
| **Vercel** | Next.js 최적 배포, 정적 사이트에 집중 가능 (API 없이) | Netlify (SSR 제약), AWS Amplify (설정 복잡) |
| **Supabase** | PostgreSQL + pgvector + Auth 통합, 무료 500MB | Firebase (NoSQL 부적합, 벡터 미지원), PlanetScale (벡터 미지원) |
| **pgvector** | PostgreSQL 네이티브 벡터 검색, Supabase 내장, 별도 벡터 DB 불필요 | Pinecone (추가 인프라, 무료 제한적) |
| **Claude Sonnet 4.5** | 한국어 분석/코칭 품질 최상, 구조화 출력 우수, 긴 컨텍스트 | GPT-4o (한국어 코칭 품질 열세) |
| **Gemini Flash / Groq Llama** | 경량 작업 최적, ~3~5원/건 저비용, 무료 티어 넉넉, 공통 LLM 인터페이스로 교체 용이 | GPT-4.1 mini (유료 전용), Claude Haiku (비용 유사하나 API 통합 복잡) |
| **shadcn/ui** | 커스터마이징 자유, Radix UI 기반, Tailwind 호환, 번들 최소 | Material UI (무겁다), Ant Design (한국 서비스 UX 부적합) |
| **Tiptap** | 리치 텍스트 에디터, ProseMirror 기반, React 통합 우수 | Quill (확장성 제한), Slate (학습곡선 높음) |
| **@hello-pangea/dnd** | 칸반보드용 DnD, 접근성, react-beautiful-dnd 후속 | dnd-kit (칸반 UX 직접 구현 필요) |
| **Recharts** | React 네이티브 차트, 선언적 API, 레이더 차트 지원 | Chart.js (React 래퍼 필요), D3 (과도한 저수준) |
| **Toss Payments** | 한국 결제 특화, SDK 간편, 스타트업 친화 | Stripe (한국 결제 제약) |

---

## 8. 배포 아키텍처

```
┌──────────────────────────────────────────────────────────────────────┐
│                         배포 구성                                     │
│                                                                      │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────────────┐ │
│  │    Vercel        │    │     Koyeb       │    │    Supabase      │ │
│  │                  │    │                  │    │                  │ │
│  │  Next.js 15      │    │  Go API Server   │    │  PostgreSQL     │ │
│  │  (SSG/ISR)       │    │  (Docker)        │    │  + pgvector     │ │
│  │                  │    │                  │    │  + Auth (JWT)    │ │
│  │  정적 페이지      │    │  512MB RAM       │    │  500MB (Free)   │ │
│  │  + CSR          │    │  (Free Tier)     │    │                  │ │
│  │                  │    │                  │    │                  │ │
│  │  비용: 무료       │    │  비용: 무료       │    │  비용: 무료       │ │
│  └────────┬─────────┘    └────────┬─────────┘    └────────┬─────────┘ │
│           │                       │                       │           │
│           │  REST API (HTTPS)     │  SQL (TCP)            │           │
│           └───────────────────────┘───────────────────────┘           │
│                                                                      │
│  ┌─────────────────┐                                                 │
│  │  Koyeb Worker    │  ← Go API Server에서 트리거                     │
│  │  (Playwright)    │                                                 │
│  │  동적 크롤링      │                                                 │
│  │  비용: 무료       │                                                 │
│  └─────────────────┘                                                 │
└──────────────────────────────────────────────────────────────────────┘
```

### 배포 역할 분담

| 서비스 | 역할 | 특이사항 |
|--------|------|---------|
| **Vercel** | 정적 Next.js (SSG/ISR, CSR) | API Routes 없음, 순수 프론트엔드만 담당 |
| **Koyeb** | Go 바이너리 (Docker, 512MB RAM) | 모든 API 로직, AI 호출, 크롤링 처리 |
| **Supabase** | PostgreSQL + pgvector 전용 | Auth SDK는 프론트엔드에서만 사용, 백엔드는 JWT 검증만 |
| **Koyeb Worker** | Playwright 동적 크롤링 | Go 서버에서 HTTP 트리거 |

### 인증 분리 구조

```
[프론트엔드 (Vercel)]
    │
    ├── Supabase Auth SDK 사용 (로그인, 회원가입, OAuth)
    │       └── JWT 토큰 발급받음
    │
    └── API 요청 시 JWT 포함 → Authorization: Bearer <JWT>

[백엔드 (Koyeb)]
    │
    ├── Supabase Auth SDK 미사용
    ├── JWT 서명만 검증 (Supabase JWT Secret으로)
    └── user_id 추출하여 DB 쿼리에 사용
```

---

## 9. 비용 구조 (무료 티어 한도)

| 서비스 | 무료 티어 한도 | 주요 제약 | 초과 시 비용 |
|--------|--------------|----------|-------------|
| **Vercel** | 배포 100회/일, 대역폭 100GB/월 | 정적 사이트 전용 (API 없음) | Pro $20/월 |
| **Supabase** | DB 500MB, Storage 1GB, MAU 50K | 1주 비활동 시 자동 중지, 프로젝트 2개 | Pro $25/월 |
| **Koyeb (API)** | 인스턴스 1개, 0.1vCPU, 512MB RAM | 1시간 무트래픽 시 Scale-to-Zero | Starter $5.6/월 |
| **Koyeb (Worker)** | 인스턴스 1개 (API와 별도) | 동시 크롤링 1건 | Starter $5.6/월 |
| **Gemini API** | 무료 티어 넉넉 (15 RPM) | 경량 작업 ~3원/건 | 종량제 |
| **Groq API** | 무료 티어 (30 RPM) | 경량 작업 ~5원/건 | 종량제 |
| **OpenAI API** | 없음 (종량제) | 임베딩 전용 ~0.5원/건 | 사용량 비례 |
| **Anthropic API** | 없음 (종량제) | Claude Sonnet 4.5 ~65원/건 | 사용량 비례 |
| **DART OpenAPI** | 10,000건/일 | 승인 후 사용 | 무료 |
| **네이버 API** | 25,000건/일 | Client ID/Secret 필요 | 무료 |
| **사람인 API** | 500건/일 | 사전 승인 필요 | 무료 |

### 예상 AI 비용 (사용자 1명 기준)

| 작업 | 모델 | 건당 비용 | 월간 예상 (10건) |
|------|------|----------|-----------------|
| 공고 파싱 | Gemini Flash / Groq Llama | ~3~5원 | ~30~50원 |
| 경험 인터뷰 (5턴) | Gemini Flash / Groq Llama | ~15~25원 | ~150~250원 |
| 경험 무기 분류 | Gemini Flash / Groq Llama | ~3~5원 | ~30~50원 |
| 임베딩 생성 | text-embedding-3-small | ~0.5원 | ~5원 |
| 기업 종합 분석 | Claude Sonnet 4.5 | ~65원 | ~650원 |
| 문항 분석 | Claude Sonnet 4.5 | ~65원 | ~650원 |
| 초안/첨삭 코칭 | Claude Sonnet 4.5 | ~65원 | ~650원 |
| **1건 풀 파이프라인** | **혼합** | **~140~200원** | **~2,000원** |

### 호스팅 비용 요약

| 항목 | 무료 운영 | 유료 전환 시 |
|------|----------|-------------|
| **프론트엔드 (Vercel)** | $0 | $20/월 (Pro) |
| **백엔드 (Koyeb)** | $0 | $5.6/월 (Starter) |
| **DB (Supabase)** | $0 | $25/월 (Pro) |
| **AI API** | 종량제 | ~2,000원/월 (10건) |
| **합계** | ~2,000원/월 | ~$50 + 2,000원/월 |
