# 환경 설정

> API 키 관리, 환경변수 설정, 보안 규칙

---

## 1. 환경변수 전체 목록 (9개)

| 변수명 | 발급처 | 비용 | 필요 Phase | 설명 |
|--------|--------|------|-----------|------|
| `NEXT_PUBLIC_SUPABASE_URL` | [supabase.com](https://supabase.com) | 무료 | Phase 0 | Supabase 프로젝트 URL |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | [supabase.com](https://supabase.com) | 무료 | Phase 0 | Supabase 익명 키 (클라이언트용) |
| `SUPABASE_SERVICE_ROLE_KEY` | [supabase.com](https://supabase.com) | 무료 | Phase 0 | Supabase 서비스 역할 키 (서버용, RLS 우회) |
| `LLM_LIGHT_PROVIDER` | - | - | Phase 2.1 | 경량 모델 프로바이더 선택 (`gemini` 또는 `groq`, 기본: `gemini`) |
| `GEMINI_API_KEY` | [aistudio.google.com](https://aistudio.google.com) | 무료 티어 | Phase 2.1 | Gemini 2.0 Flash (경량 작업) |
| `GROQ_API_KEY` | [console.groq.com](https://console.groq.com) | 무료 티어 | Phase 2.1 | Groq Llama 3.3 (경량 작업, `groq` 선택 시) |
| `OPENAI_API_KEY` | [platform.openai.com](https://platform.openai.com) | 종량제 | Phase 2.1 | text-embedding-3-small (임베딩 전용) |
| `ANTHROPIC_API_KEY` | [console.anthropic.com](https://console.anthropic.com) | 종량제 | Phase 3.2 | Claude Sonnet 4.5 |
| `DART_API_KEY` | [opendart.fss.or.kr](https://opendart.fss.or.kr) | 무료 | Phase 3.1 | DART OpenAPI (기업 재무정보) |
| `NAVER_CLIENT_ID` | [developers.naver.com](https://developers.naver.com) | 무료 | Phase 3.1 | 네이버 검색 API |
| `NAVER_CLIENT_SECRET` | [developers.naver.com](https://developers.naver.com) | 무료 | Phase 3.1 | 네이버 검색 API |
| `SARAMIN_ACCESS_KEY` | [oapi.saramin.co.kr](https://oapi.saramin.co.kr) | 무료 (승인 필요) | Phase 10 | 사람인 채용공고 API |
| `KOYEB_API_TOKEN` | [koyeb.com](https://koyeb.com) | 무료 티어 | Phase 10 | Koyeb 워커 배포/관리 |

> **참고**: RLS를 사용하지 않으므로(`02-data-structure.md` 참조), `SUPABASE_SERVICE_ROLE_KEY`는 Supabase Management API 호출이 필요한 경우에만 사용합니다. 일반 CRUD는 Go 백엔드가 `DATABASE_URL`로 직접 연결합니다.

---

## 2. .env.local 템플릿

```bash
# ============================================
# Colight 환경변수
# ============================================

# --- Supabase (Phase 0) ---
# https://supabase.com → 프로젝트 Settings → API
NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# --- 경량 모델 프로바이더 (Phase 2.1) ---
# "gemini" (기본) 또는 "groq"
LLM_LIGHT_PROVIDER=gemini

# --- Gemini (Phase 2.1) ---
# https://aistudio.google.com/apikey
GEMINI_API_KEY=your-gemini-key

# --- Groq (Phase 2.1, groq 선택 시) ---
# https://console.groq.com/keys
GROQ_API_KEY=your-groq-key

# --- OpenAI (Phase 2.1, 임베딩 전용) ---
# https://platform.openai.com/api-keys
OPENAI_API_KEY=sk-...

# --- Anthropic (Phase 3.2) ---
# https://console.anthropic.com/settings/keys
ANTHROPIC_API_KEY=sk-ant-...

# --- DART OpenAPI (Phase 3.1) ---
# https://opendart.fss.or.kr → 인증키 신청
DART_API_KEY=your-dart-key

# --- 네이버 검색 API (Phase 3.1) ---
# https://developers.naver.com → 애플리케이션 등록
NAVER_CLIENT_ID=your-naver-client-id
NAVER_CLIENT_SECRET=your-naver-client-secret

# --- 사람인 API (Phase 10) ---
# https://oapi.saramin.co.kr → API 사용 신청 (승인 필요)
SARAMIN_ACCESS_KEY=your-saramin-key

# --- Koyeb (Phase 10) ---
# https://app.koyeb.com → Account → API
KOYEB_API_TOKEN=your-koyeb-token
```

---

## 3. NEXT_PUBLIC_ 접두사 규칙

### 클라이언트 노출 가능 (`NEXT_PUBLIC_` 접두사)

```
NEXT_PUBLIC_SUPABASE_URL        → 브라우저에서 Supabase 접속
NEXT_PUBLIC_SUPABASE_ANON_KEY   → 브라우저에서 인증 요청 (RLS로 보호됨)
```

- 브라우저 번들에 포함됨
- 개발자 도구에서 확인 가능
- **RLS 정책으로 데이터 접근을 제한**하므로 노출되어도 안전

### 서버 전용 (접두사 없음)

```
SUPABASE_SERVICE_ROLE_KEY       → RLS 우회, 관리 작업용
LLM_LIGHT_PROVIDER              → 경량 모델 프로바이더 선택
GEMINI_API_KEY                  → Gemini Flash 호출
GROQ_API_KEY                    → Groq Llama 호출
OPENAI_API_KEY                  → 임베딩 API 호출
ANTHROPIC_API_KEY               → Claude API 호출
DART_API_KEY                    → 기업 데이터 조회
NAVER_CLIENT_ID                 → 뉴스 검색
NAVER_CLIENT_SECRET             → 뉴스 검색
SARAMIN_ACCESS_KEY              → 채용공고 검색
KOYEB_API_TOKEN                 → 워커 관리
```

- 서버 사이드에서만 접근 가능 (API Routes, Server Components)
- **절대 클라이언트에 노출되면 안 됨**
- `NEXT_PUBLIC_` 접두사를 붙이지 않을 것

### 코드에서 접근

```typescript
// Server Component / API Route (안전)
const serviceKey = process.env.SUPABASE_SERVICE_ROLE_KEY;
const openaiKey = process.env.OPENAI_API_KEY;

// Client Component (NEXT_PUBLIC_ 만 접근 가능)
const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL;
const anonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;
```

---

## 4. Vercel 환경변수 설정

### 설정 방법

1. [Vercel Dashboard](https://vercel.com/dashboard) → 프로젝트 선택
2. **Settings** → **Environment Variables**
3. 각 변수 추가 (환경별 분리)

### 환경별 설정

| 변수 | Development | Preview | Production |
|------|:-----------:|:-------:|:----------:|
| `NEXT_PUBLIC_SUPABASE_URL` | 로컬 URL | 스테이징 URL | 프로덕션 URL |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | 로컬 키 | 스테이징 키 | 프로덕션 키 |
| `SUPABASE_SERVICE_ROLE_KEY` | 로컬 키 | 스테이징 키 | 프로덕션 키 |
| `OPENAI_API_KEY` | 동일 | 동일 | 동일 |
| `ANTHROPIC_API_KEY` | 동일 | 동일 | 동일 |
| 나머지 API 키 | 동일 | 동일 | 동일 |

> Development: `vercel dev` 실행 시 사용. 로컬 개발은 `.env.local` 사용.

### Vercel CLI로 환경변수 설정

```bash
# 환경변수 추가
vercel env add OPENAI_API_KEY production

# 환경변수 목록 확인
vercel env ls

# 환경변수 가져오기 (.env.local로)
vercel env pull .env.local
```

---

## 5. .gitignore 설정

```gitignore
# 환경변수 (절대 커밋 금지)
.env
.env.local
.env.*.local

# Supabase 로컬 설정
supabase/.temp/

# IDE
.idea/
.vscode/
*.swp
```

**반드시 커밋에서 제외할 파일:**
- `.env.local` — 모든 API 키 포함
- `.env` — 환경변수 원본
- `supabase/.temp/` — 로컬 DB 임시 파일

---

## 6. Phase별 API 키 필요 시점

```
Phase 0   프로젝트 셋업
          ├── NEXT_PUBLIC_SUPABASE_URL
          ├── NEXT_PUBLIC_SUPABASE_ANON_KEY
          └── SUPABASE_SERVICE_ROLE_KEY

Phase 2.1 무기 자동 태깅
          ├── LLM_LIGHT_PROVIDER
          ├── GEMINI_API_KEY (또는 GROQ_API_KEY)
          └── OPENAI_API_KEY (임베딩)

Phase 3.1 기업 데이터 API
          ├── DART_API_KEY
          ├── NAVER_CLIENT_ID
          └── NAVER_CLIENT_SECRET

Phase 3.2 AI 기업 분석
          └── ANTHROPIC_API_KEY

Phase 10  성장 기능
          ├── SARAMIN_ACCESS_KEY
          └── KOYEB_API_TOKEN
```

> 각 Phase를 시작하기 전에 해당 API 키를 발급받아 `.env.local`에 추가한다.

---

## 7. 환경변수 검증

앱 시작 시 필수 환경변수가 설정되었는지 검증한다.

```typescript
// src/lib/env.ts

function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

// Phase 0 필수
export const SUPABASE_URL = requireEnv('NEXT_PUBLIC_SUPABASE_URL');
export const SUPABASE_ANON_KEY = requireEnv('NEXT_PUBLIC_SUPABASE_ANON_KEY');
export const SUPABASE_SERVICE_ROLE_KEY = requireEnv('SUPABASE_SERVICE_ROLE_KEY');

// Phase 2.1+ (사용 시점에 검증)
export const getLightProvider = () => process.env.LLM_LIGHT_PROVIDER || 'gemini';
export const getGeminiKey = () => requireEnv('GEMINI_API_KEY');
export const getGroqKey = () => requireEnv('GROQ_API_KEY');
export const getOpenAIKey = () => requireEnv('OPENAI_API_KEY'); // 임베딩 전용
export const getAnthropicKey = () => requireEnv('ANTHROPIC_API_KEY');
export const getDartKey = () => requireEnv('DART_API_KEY');
export const getNaverCredentials = () => ({
  clientId: requireEnv('NAVER_CLIENT_ID'),
  clientSecret: requireEnv('NAVER_CLIENT_SECRET'),
});
```
