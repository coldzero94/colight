> **⚠️ Superseded**: 이 문서는 초기 기획 시점의 기술 요구사항으로, Next.js 풀스택 구조를 기준으로 작성되었습니다. 최종 기술 스택 및 프로젝트 구조는 `CLAUDE.md`와 `docs/develop/01-architecture.md`를 참조하세요.
> - 프로젝트 구조: 모노레포(apps/backend + apps/web + packages/protocol)
> - 백엔드: Go 1.24 + Gin + Ent (TypeSpec → OpenAPI → oapi-codegen)
> - AI 호출: Go 백엔드에서 처리 (Vercel AI SDK 미사용)

# Colight 기술 구현 요구사항 (개발자 관점)

> 작성일: 2026-02-09
> 총 공수: 39~51일 | DB 테이블: 15개 | API 키: 8개

---

## 1. 기능별 기술 요구사항 요약

### 인증/회원 관리 | 2~3일 | 난이도: 하
- Supabase Auth (이메일/소셜 로그인)
- Next.js Middleware, RLS
- 테이블: users(기본), user_profiles

### 경험 관리 | 7~9일 | 난이도: 중
- Next.js 15, Vercel AI SDK (스트리밍 대화형 인터뷰)
- Gemini Flash / Groq Llama (인터뷰, 분류), text-embedding-3-small (임베딩)
- Supabase pgvector
- 테이블: experiences, experience_tags, experience_usages, experience_weapons
- 공수 분배: CRUD+UI 2일, AI 인터뷰 3일, 태그 분류+무기 매핑 1.5일, 임베딩 1.5일

### 기업 분석 | 12~15일 | 난이도: 상
- Cheerio (잡코리아/캐치), Playwright→Koyeb (원티드)
- Claude Sonnet 4.5 (종합 분석), Gemini Flash / Groq Llama (공고 파싱)
- DART OpenAPI, 네이버 뉴스 API, 사람인 API
- 테이블: company_analysis_cache, talent_profiles, applications, company_analyses
- 공수 분배: 크롤링 엔진 4일, API 연동 2일, AI 파이프라인 3일, 매칭 알고리즘 2일, 캐싱 1일, UI 2일

### AI 자소서 코칭 | 10~13일 | 난이도: 상
- Vercel AI SDK (스트리밍), Claude Sonnet 4.5 (코칭/첨삭)
- 리치 텍스트 에디터 (Tiptap)
- 테이블: cover_letters, cover_letter_versions, coaching_sessions, prompt_templates, question_patterns
- 공수 분배: 문항 분석+추천 2일, 초안 코칭 2.5일, 첨삭 코칭 2.5일, 에디터 UI 2일, 버전 관리 1일, AI 탐지+글자수 1일, 프롬프트 DB 1.5일

### 지원 대시보드 | 5~7일 | 난이도: 중
- @hello-pangea/dnd (칸반), Recharts (차트)
- 공수 분배: 칸반 UI 2일, 캘린더 1일, 통계 시각화 2일, 버전 관리 UI 1일

### 공통 인프라 | 3~4일 | 난이도: 중

---

## 2. DB 스키마 전체 (15개 테이블)

```
── 인증/사용자 ──
users                    (Supabase Auth 기본)
user_profiles            (프로필 확장)

── 경험 관리 ──
experiences              (경험 등록)
experience_tags          (역량 태그)
experience_usages        (사용 이력)
experience_weapons       (경험→무기 매핑)

── 무기/프롬프트 ──
weapon_categories        (무기 카테고리 마스터)
prompt_templates         (AI 프롬프트 템플릿)
question_patterns        (공통 문항 패턴)

── 기업 분석 ──
company_analysis_cache   (분석 캐시, 7일 TTL)
talent_profiles          (기업 인재상 사전 DB)

── 지원 관리 ──
applications             (지원 현황)
company_analyses         (분석 결과)
cover_letters            (자소서)
cover_letter_versions    (버전 관리)

── 코칭 ──
coaching_sessions        (코칭 대화 이력)
```

---

## 3. AI 모델 사용 전략

| 작업 | 모델 | 건당 비용 |
|------|------|----------|
| 공고 파싱/구조화 | Gemini Flash / Groq Llama | ~3~5원 |
| 경험 인터뷰 (대화형) | Gemini Flash / Groq Llama | ~3~5원/턴 |
| 경험 무기 자동 분류 | Gemini Flash / Groq Llama | ~3~5원 |
| 경험 매칭 (1차) | text-embedding-3-small | ~0.5원 |
| 기업 종합 분석 | Claude Sonnet 4.5 | ~65원 |
| 문항 분석 + 추천 | Claude Sonnet 4.5 | ~65원 |
| 초안/첨삭 코칭 | Claude Sonnet 4.5 | ~65원 |
| 경험 매칭 (2차 정밀) | Gemini Flash / Groq Llama | ~3~5원 |

**1건 풀 파이프라인**: 약 140~200원

---

## 4. 외부 API 키 목록 (7개)

| API | 발급처 | 비용 |
|-----|--------|------|
| Anthropic API Key | console.anthropic.com | 종량제 |
| Google AI API Key | aistudio.google.dev | 종량제 |
| Groq API Key | console.groq.com | 종량제 |
| DART OpenAPI Key | opendart.fss.or.kr | 무료 |
| 네이버 Client ID/Secret | developers.naver.com | 무료 |
| 사람인 Access Key | oapi.saramin.co.kr | 무료 (승인 필요) |
| Supabase URL/Anon Key | supabase.com | 무료 티어 |
| Koyeb API Token | koyeb.com | 무료 티어 |

---

## 5. Next.js 프로젝트 구조

```
src/
├── app/
│   ├── (auth)/login/, signup/
│   ├── (main)/
│   │   ├── experiences/        # 경험 관리
│   │   ├── analysis/           # 기업 분석
│   │   ├── coaching/           # 자소서 코칭
│   │   └── dashboard/          # 대시보드
│   └── api/
│       ├── analyze/            # 기업 분석
│       ├── crawl/              # 크롤링
│       ├── coaching/           # 코칭
│       └── experiences/        # 경험 CRUD
├── lib/
│   ├── supabase/               # DB 클라이언트
│   ├── ai/                     # AI 유틸 (prompts, parsing, analysis, matching, coaching)
│   ├── crawling/               # cheerio-parser, koyeb-client
│   └── external/               # dart, naver-news, saramin
├── components/                 # UI 컴포넌트
├── hooks/                      # 커스텀 훅
└── types/                      # TypeScript 타입
```

---

## 6. 기술 스택 전체 요약

| 카테고리 | 기술 | 비용 |
|----------|------|------|
| 프레임워크 | Next.js 15 (App Router) | 무료 |
| 배포 (프론트) | Vercel | 무료 |
| 배포 (백엔드) | Koyeb | 무료 |
| DB | Supabase PostgreSQL + pgvector | 무료 (500MB) |
| 인증 | Supabase Auth | 무료 |
| AI (경량) | Gemini Flash / Groq Llama | 종량제 |
| AI (고급) | Claude Sonnet 4.5 | 종량제 |
| AI 통합 | Vercel AI SDK | 무료 |
| 크롤링 | Cheerio + Playwright | 무료 |
| UI | Tailwind CSS + shadcn/ui | 무료 |
| 칸반 | @hello-pangea/dnd | 무료 |
| 차트 | Recharts | 무료 |

---

## 7. 기술 리스크 TOP 5

1. **Vercel 10초 타임아웃** → 스트리밍 응답 + Fluid Compute(60초) 활용
2. **Koyeb 무료 티어 제약** → Playwright 동시 1건, cold start 지연
3. **크롤링 사이트 구조 변경** → AI fallback 파싱 + 모니터링
4. **Supabase 500MB 제한** → 만료 데이터 자동 정리 Edge Function
5. **AI 비용 관리** → 1건 ~200원, 캐싱/임베딩 필터링으로 최소화
