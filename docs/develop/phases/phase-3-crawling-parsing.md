# Phase 3: 크롤링 & 파싱

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시는 Next.js API Routes + Cheerio 기반으로 작성되었으나, 최종 아키텍처에서는 **Go 백엔드 + goquery**로 구현합니다. 코드 로직은 참고하되, 실제 구현은 Go로 작성하세요.
>
> - `npm install cheerio` → `go get github.com/PuerkitoBio/goquery`
> - `src/app/api/analyze/route.ts` → `apps/backend/internal/service/crawling.go`
> - `src/lib/crawling/` → `apps/backend/internal/infrastructure/crawler/`

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 채용공고 URL을 입력하면 구조화된 `JobPosting` 데이터로 변환하는 전체 파이프라인 구축 |
| **선행 조건** | Phase 2 완료 (경험 CRUD), Supabase DB 마이그레이션, 환경변수 설정 (OPENAI_API_KEY) |
| **스프린트** | Sprint 2 — Day 1~2 |
| **관련 기능** | F07 (채용공고 자동 분석) |
| **예상 공수** | 2일 (16시간) |
| **산출물** | URL 입력 UI, 잡코리아/캐치 파서, AI 정규화 함수, 도메인 라우터 API |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 3.1 | URL 입력 컴포넌트 | ⬜ |
| 3.2 | Cheerio 파서 — 잡코리아 | ⬜ |
| 3.3 | Cheerio 파서 — 캐치 | ⬜ |
| 3.4 | AI 구조화 (GPT-4.1 mini) | ⬜ |
| 3.5 | URL 도메인 라우터 API | ⬜ |

---

## Step 3.1: URL 입력 컴포넌트

### 목표

사용자가 채용공고 URL을 붙여넣으면, 도메인을 자동 감지하여 배지(badge)를 표시하고, 유효성 검증 후 분석을 시작할 수 있는 입력 UI를 제공한다.

### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/lib/validations/__tests__/job-url.test.ts`
  - `describe('JobUrlSchema')` — URL 유효성 검증 (정상 URL, 빈 값, 잘못된 형식)
  - `describe('detectDomain')` — 도메인 자동 감지 (잡코리아 → 'jobkorea', 캐치 → 'catch', 미지원 → 'unknown')
- `src/components/analysis/__tests__/job-url-input.test.tsx`
  - `it('renders domain badge for jobkorea URL')` — 잡코리아 URL 입력 시 파란색 배지 표시
  - `it('renders domain badge for catch URL')` — 캐치 URL 입력 시 초록색 배지 표시
  - `it('shows error for invalid URL')` — 잘못된 URL 입력 시 에러 메시지 표시
  - `it('shows unsupported domain message')` — 미지원 도메인 입력 시 안내 메시지 표시
  - `it('shows validation error for empty submission')` — 빈 값 제출 시 유효성 에러
  - `it('shows loading state when submitting')` — 분석 시작 시 로딩 상태

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/lib/validations/__tests__/job-url.test.ts`
  - [ ] `src/components/analysis/__tests__/job-url-input.test.tsx`
- [ ] 구현 (GREEN)
  - [ ] `JobUrlInput` 컴포넌트 생성
  - [ ] URL 유효성 검증 (정규식 + URL 파싱)
  - [ ] 도메인 자동 감지 로직 구현 (잡코리아, 캐치, 원티드, 사람인, 기타)
  - [ ] 도메인별 배지 UI (아이콘 + 색상 구분)
  - [ ] 지원 사이트 / 미지원 사이트 안내 메시지
  - [ ] 클립보드 붙여넣기 자동 감지
  - [ ] 분석 시작 버튼 + 로딩 상태
  - [ ] 에러 상태 표시 (잘못된 URL, 파싱 실패 등)
- [ ] 테스트 통과 확인

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `JobUrlInput` | `src/components/analysis/job-url-input.tsx` | `onSubmit: (url: string) => void`, `isLoading: boolean`, `error?: string` | URL 입력 + 도메인 배지 + 제출 버튼 |
| `DomainBadge` | `src/components/analysis/domain-badge.tsx` | `domain: SupportedDomain \| 'unknown'` | 잡코리아(파란색), 캐치(초록색), 원티드(보라색), 사람인(빨간색), 기타(회색) |

### 도메인 감지 로직

```typescript
// src/lib/crawling/domain-detector.ts

type SupportedDomain = 'jobkorea' | 'catch' | 'wanted' | 'saramin';

interface DomainInfo {
  domain: SupportedDomain | 'unknown';
  label: string;
  supported: boolean;
  method: 'cheerio' | 'playwright' | 'api' | 'ai-fallback';
}

const DOMAIN_MAP: Record<string, DomainInfo> = {
  'jobkorea.co.kr': { domain: 'jobkorea', label: '잡코리아', supported: true, method: 'cheerio' },
  'catch.co.kr':    { domain: 'catch',    label: '캐치',     supported: true, method: 'cheerio' },
  'wanted.co.kr':   { domain: 'wanted',   label: '원티드',   supported: false, method: 'playwright' },
  'saramin.co.kr':  { domain: 'saramin',  label: '사람인',   supported: false, method: 'api' },
};

function detectDomain(url: string): DomainInfo;
```

### 산출물

- `src/components/analysis/job-url-input.tsx`
- `src/components/analysis/domain-badge.tsx`
- `src/lib/crawling/domain-detector.ts`

---

## Step 3.2: Cheerio 파서 — 잡코리아

### 목표

잡코리아 채용공고 URL에서 HTML을 가져와 Cheerio로 정적 파싱하여, 회사명, 포지션, 부서, 자격요건, 우대사항, 마감일 등의 데이터를 추출한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/infrastructure/crawler/jobkorea_parser_test.go`
  - `TestParseJobKorea_FullPosting` — testdata/crawling/jobkorea_full.html 픽스처로 전체 필드 추출 확인
  - `TestParseJobKorea_MissingSections` — 일부 섹션 누락 HTML에서 빈 값 반환 확인
  - `TestParseJobKorea_CompanyName` — 회사명 정상 추출 확인
  - `TestParseJobKorea_SelectorMismatch` — 셀렉터 미스매치 시 에러 없이 빈 값 반환 확인

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/infrastructure/crawler/jobkorea_parser_test.go`
  - [ ] HTML 픽스처 파일 생성: `testdata/crawling/jobkorea_full.html`, `testdata/crawling/jobkorea_minimal.html`
- [ ] 구현 (GREEN)
  - [ ] `goquery` 패키지 추가 (`go get github.com/PuerkitoBio/goquery`)
  - [ ] `jobkorea_parser.go` 생성
  - [ ] HTML fetch 함수 (User-Agent 헤더 설정, robots.txt 준수)
  - [ ] CSS 셀렉터 기반 데이터 추출 로직 구현
  - [ ] 추출 실패 시 graceful fallback (빈 문자열 반환, 에러 로깅)
  - [ ] `RawJobPosting` 타입 정의 및 반환
- [ ] 테스트 통과 확인

### CSS 셀렉터 매핑

```typescript
// src/lib/crawling/jobkorea-parser.ts

// 잡코리아 채용공고 페이지 구조:
// URL 패턴: https://www.jobkorea.co.kr/Recruit/GI_Read/{공고ID}

const SELECTORS = {
  companyName:  '.company-name a, .coName',           // 회사명
  position:     '.artReadJobTitle, .title-wrap h1',     // 포지션명
  department:   '.tbRow .dept',                         // 부서/팀
  career:       '.tbRow .career',                       // 경력 조건
  education:    '.tbRow .education',                    // 학력 조건
  jobType:      '.tbRow .jobtype',                      // 고용형태
  salary:       '.tbRow .salary',                       // 급여
  location:     '.tbRow .location',                     // 근무지
  deadline:     '.date .tahoma',                        // 마감일
  mainTasks:    '.artReadJobSecCont:nth-child(1)',       // 담당업무
  requirements: '.artReadJobSecCont:nth-child(2)',       // 자격요건
  preferred:    '.artReadJobSecCont:nth-child(3)',       // 우대사항
  skills:       '.skillWrap .skill',                    // 스킬 태그
};
```

### 파서 인터페이스

```typescript
interface RawJobPosting {
  source: SupportedDomain;
  sourceUrl: string;
  companyName: string;
  position: string;
  department?: string;
  career?: string;
  education?: string;
  jobType?: string;
  salary?: string;
  location?: string;
  deadline?: string;
  mainTasks?: string;
  requirements?: string;
  preferred?: string;
  skills?: string[];
  rawHtml?: string; // AI fallback용 (저장하지 않음, 분석 후 폐기)
}

async function parseJobKorea(url: string): Promise<RawJobPosting>;
```

### 주의사항

- **robots.txt 준수**: 잡코리아는 AI/LLM 크롤러를 차단하지만, 채용정보 페이지 자체는 일반 크롤러에 부분 허용
- **사용자 URL 입력 기반 단건 파싱**: 대량 크롤링이 아니므로 법적 리스크 낮음
- **User-Agent**: 일반 브라우저 User-Agent 사용
- **Rate limiting**: 사용자 요청 기반이므로 별도 제한 불필요, 단 동일 URL 재요청 시 캐시 활용
- **HTML 원문 미저장**: 파싱 후 구조화된 데이터만 저장, rawHtml은 분석 완료 후 폐기

### 산출물

- `internal/infrastructure/crawler/jobkorea_parser.go`
- `internal/infrastructure/crawler/types.go` (`RawJobPosting`, `SupportedDomain` 타입 정의)
- `internal/infrastructure/crawler/jobkorea_parser_test.go`
- `testdata/crawling/jobkorea_full.html`

---

## Step 3.3: Cheerio 파서 — 캐치

### 목표

캐치(CATCH) 채용공고 URL에서 HTML을 가져와 Cheerio로 정적 파싱하여, 잡코리아 파서와 동일한 `RawJobPosting` 형태로 데이터를 추출한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/infrastructure/crawler/catch_parser_test.go`
  - `TestParseCatch_FullPosting` — testdata/crawling/catch_full.html 픽스처로 전체 필드 추출 확인
  - `TestParseCatch_RawJobPostingStructure` — 잡코리아 파서와 동일한 `RawJobPosting` 구조 반환 확인
  - `TestParseCatch_MissingSections` — 일부 섹션 누락 HTML에서 빈 값 반환 확인
  - `TestParseCatch_SelectorMismatch` — 셀렉터 미스매치 시 에러 없이 빈 값 반환 확인

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/infrastructure/crawler/catch_parser_test.go`
  - [ ] HTML 픽스처 파일 생성: `testdata/crawling/catch_full.html`, `testdata/crawling/catch_minimal.html`
- [ ] 구현 (GREEN)
  - [ ] `catch_parser.go` 생성
  - [ ] 캐치 페이지 구조 분석 및 CSS 셀렉터 매핑
  - [ ] HTML fetch + goquery 파싱 구현
  - [ ] `RawJobPosting` 타입으로 정규화
  - [ ] 잡코리아 파서와 동일한 에러 처리 패턴 적용
- [ ] 테스트 통과 확인

### CSS 셀렉터 매핑

```typescript
// src/lib/crawling/catch-parser.ts

// 캐치 채용공고 페이지 구조:
// URL 패턴: https://www.catch.co.kr/Comp/RecruitInfo/{회사코드}

const SELECTORS = {
  companyName:  '.company-name, .comp-name',            // 회사명
  position:     '.recruit-title, .job-title',            // 포지션명
  department:   '.recruit-info .dept',                   // 부서
  career:       '.recruit-info .career',                 // 경력
  education:    '.recruit-info .edu',                    // 학력
  jobType:      '.recruit-info .type',                   // 고용형태
  location:     '.recruit-info .location',               // 근무지
  deadline:     '.recruit-info .deadline, .dday',        // 마감일
  mainTasks:    '.job-description .task-section',         // 담당업무
  requirements: '.job-description .requirement-section',  // 자격요건
  preferred:    '.job-description .preferred-section',    // 우대사항
};
```

### 공통 파서 유틸리티

잡코리아/캐치 파서 간 공통 로직을 유틸리티로 추출:

```typescript
// src/lib/crawling/parser-utils.ts

/** HTML fetch with proper headers */
async function fetchHtml(url: string): Promise<string>;

/** 텍스트 정리 (불필요한 공백, 특수문자 제거) */
function cleanText(text: string): string;

/** 리스트 형태의 텍스트를 배열로 변환 */
function parseListText(text: string): string[];

/** 날짜 문자열 정규화 (YYYY-MM-DD) */
function normalizeDate(dateStr: string): string | undefined;
```

### 산출물

- `internal/infrastructure/crawler/catch_parser.go`
- `internal/infrastructure/crawler/parser_utils.go`
- `internal/infrastructure/crawler/catch_parser_test.go`
- `testdata/crawling/catch_full.html`

---

## Step 3.4: AI 구조화 (GPT-4.1 mini)

### 목표

파서가 추출한 비정형 `RawJobPosting` 데이터를 GPT-4.1 mini로 정규화하여, 타입 안전한 `JobPosting` 구조로 변환한다. 미지원 사이트의 경우 raw HTML을 직접 LLM에 전달하여 구조화한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/service/crawling_service_test.go`
  - `TestNormalizeJobPosting_Success` — RawJobPosting → JobPosting 정상 변환 확인 (MockAIClient 사용)
  - `TestNormalizeJobPosting_RequiredFieldsMissing` — 필수 필드(companyName, position) 누락 시 에러 처리
  - `TestNormalizeJobPosting_RetryOnValidationFailure` — Zod 검증 실패 시 1회 재시도 확인
- `internal/infrastructure/ai/normalizer_test.go`
  - `TestParseHtmlWithAI_Fallback` — 미지원 사이트 HTML → AI fallback 파싱 확인 (testdata/crawling/unknown_site.html 픽스처)
  - `TestParseHtmlWithAI_TextNormalization` — 비정형 텍스트("경력 3년 이상 또는 석사") → 정규화된 값 변환 확인

> 패턴 참고: docs/develop/08-testing-strategy.md

- `src/lib/validations/__tests__/job-posting.test.ts`
  - `describe('JobPostingSchema')` — Zod 스키마 검증 (정상 데이터, 필수 필드 누락, 잘못된 enum 값)

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/crawling_service_test.go`
  - [ ] `internal/infrastructure/ai/normalizer_test.go`
  - [ ] `src/lib/validations/__tests__/job-posting.test.ts`
- [ ] 구현 (GREEN)
  - [ ] `JobPosting` Zod 스키마 정의
  - [ ] AI 정규화 프롬프트 작성 (`prompt_templates` 테이블에 저장)
  - [ ] `normalizeJobPosting()` 함수 구현 (OpenAI GPT-4.1 mini 호출)
  - [ ] AI fallback 파서 구현 (raw HTML → `JobPosting`)
  - [ ] 응답 Zod 검증 + 실패 시 재시도 (최대 1회)
  - [ ] 비용 추적 로깅
- [ ] 테스트 통과 확인

### DB 마이그레이션

```sql
-- prompt_templates 테이블에 시드 데이터 추가
INSERT INTO prompt_templates (category, name, version, system_prompt, user_prompt_template, model, temperature, max_tokens)
VALUES (
  'job_parsing',
  'normalize_job_posting',
  1,
  '당신은 채용공고 데이터를 정확하게 구조화하는 전문가입니다. 주어진 비정형 데이터를 정해진 JSON 스키마에 맞게 정규화하세요.',
  '다음 채용공고 데이터를 구조화된 JSON으로 변환해주세요.\n\n{{rawData}}\n\n반드시 아래 스키마를 따르세요:\n{{schema}}',
  'gpt-4.1-mini',
  0.1,
  2000
);

INSERT INTO prompt_templates (category, name, version, system_prompt, user_prompt_template, model, temperature, max_tokens)
VALUES (
  'job_parsing',
  'parse_html_fallback',
  1,
  '당신은 채용공고 웹페이지에서 핵심 정보를 추출하는 전문가입니다. HTML에서 채용공고 관련 정보만 정확하게 추출하세요.',
  '다음 HTML에서 채용공고 정보를 추출하여 JSON으로 구조화해주세요.\n\n{{html}}\n\n반드시 아래 스키마를 따르세요:\n{{schema}}',
  'gpt-4.1-mini',
  0.1,
  2000
);
```

### JobPosting Zod 스키마

```typescript
// src/lib/crawling/schemas.ts
import { z } from 'zod';

export const JobPostingSchema = z.object({
  companyName: z.string().min(1, '회사명은 필수입니다'),
  position: z.string().min(1, '포지션명은 필수입니다'),
  department: z.string().optional(),
  jobType: z.enum(['정규직', '계약직', '인턴', '파견직', '기타']).optional(),
  experienceLevel: z.string().optional(),      // "신입", "경력 3년 이상" 등
  education: z.string().optional(),
  location: z.string().optional(),
  salary: z.string().optional(),
  mainTasks: z.array(z.string()).default([]),   // 주요 업무
  requirements: z.array(z.string()).default([]), // 자격요건
  preferred: z.array(z.string()).default([]),    // 우대사항
  requiredSkills: z.array(z.string()).default([]), // 필수 스킬
  softSkills: z.array(z.string()).default([]),   // 소프트 스킬
  companyValuesHints: z.array(z.string()).default([]), // 기업 가치 힌트
  deadline: z.string().optional(),               // YYYY-MM-DD
  sourceUrl: z.string().url(),
  sourceDomain: z.string(),
});

export type JobPosting = z.infer<typeof JobPostingSchema>;
```

### AI 정규화 함수

```typescript
// src/lib/crawling/ai-normalizer.ts
import OpenAI from 'openai';

interface NormalizeOptions {
  raw: RawJobPosting;
  promptTemplate?: PromptTemplate; // DB에서 로드
}

/**
 * RawJobPosting → JobPosting (GPT-4.1 mini)
 * - 비정형 텍스트를 배열/구조화 데이터로 변환
 * - 스킬 분류 (required vs soft)
 * - 기업 가치 힌트 추출
 * - 날짜 정규화
 */
async function normalizeJobPosting(options: NormalizeOptions): Promise<JobPosting>;

/**
 * Unknown 사이트의 HTML → JobPosting (AI fallback)
 * - HTML 전체를 LLM에 전달 (토큰 제한을 위해 body만 추출, max 8000자)
 * - 직접 JobPosting 스키마로 구조화
 */
async function parseHtmlWithAI(html: string, url: string): Promise<JobPosting>;
```

### 비용 관리

- GPT-4.1 mini 사용: 건당 약 5원
- 입력 토큰: ~1000 (RawJobPosting 데이터)
- 출력 토큰: ~500 (구조화된 JSON)
- AI fallback (HTML 직접 파싱): 입력 토큰 ~4000, 건당 약 15원

### 산출물

- `internal/infrastructure/ai/normalizer.go`
- `internal/infrastructure/ai/normalizer_test.go`
- `src/lib/validations/job-posting.ts`
- DB 시드: `prompt_templates` 2건 추가

---

## Step 3.5: URL 도메인 라우터 API

### 목표

`/api/analyze` POST 엔드포인트를 구현하여, 사용자가 입력한 URL의 도메인을 감지하고, 적절한 파서로 라우팅한 뒤, AI 정규화를 거쳐 `JobPosting` 결과를 반환한다. 미지원 도메인은 AI fallback으로 처리한다.

### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

- `internal/controller/analysis_controller_test.go`
  - `TestAnalyzeURL_JobKoreaRouting` — 잡코리아 URL → Cheerio 파서 → AI 정규화 → JobPosting 반환
  - `TestAnalyzeURL_CatchRouting` — 캐치 URL → Cheerio 파서 → AI 정규화 → JobPosting 반환
  - `TestAnalyzeURL_UnknownDomainFallback` — 미지원 도메인 URL → AI fallback → JobPosting 반환
  - `TestAnalyzeURL_InvalidURL` — 잘못된 URL → `INVALID_URL` 에러 반환
  - `TestAnalyzeURL_FetchFailed` — 존재하지 않는 페이지 → `FETCH_FAILED` 에러 반환
  - `TestAnalyzeURL_Unauthorized` — 미인증 요청 → `UNAUTHORIZED` 에러 반환
  - `TestAnalyzeURL_CacheHit` — 동일 URL 재요청 시 캐시 응답 반환
- `internal/service/crawling_service_test.go`
  - `TestDomainRouter_SelectsCorrectParser` — 도메인별 파서 라우팅 로직 단위 테스트

### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/controller/analysis_controller_test.go`
  - [ ] `internal/service/crawling_service_test.go` (도메인 라우팅)
- [ ] 구현 (GREEN)
  - [ ] `/api/analyze` POST 라우트 생성
  - [ ] 도메인 감지 → 파서 선택 라우팅 로직
  - [ ] 파서 실행 → AI 정규화 → 결과 반환 파이프라인
  - [ ] 미지원 도메인 AI fallback 처리
  - [ ] 에러 처리 (fetch 실패, 파싱 실패, AI 실패)
  - [ ] 인증 미들웨어 (로그인 사용자만)
  - [ ] Rate limiting (사용자당 분당 5회)
  - [ ] 응답 캐싱 (동일 URL은 1시간 내 재파싱 방지)
- [ ] 테스트 통과 확인

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/api/analyze` | `{ url: string }` | `{ success: boolean, data: JobPosting, meta: { source, parsedAt, method } }` |

### 라우팅 로직 상세

```typescript
// src/app/api/analyze/route.ts

export async function POST(request: Request) {
  // 1. 인증 확인
  const user = await getAuthUser(request);
  if (!user) return unauthorized();

  // 2. URL 유효성 검증
  const { url } = await request.json();
  const validation = validateUrl(url);
  if (!validation.valid) return badRequest(validation.error);

  // 3. 도메인 감지
  const domainInfo = detectDomain(url);

  // 4. 캐시 확인 (동일 URL 최근 1시간 이내)
  const cached = await checkUrlCache(url);
  if (cached) return json({ success: true, data: cached, meta: { cached: true } });

  // 5. 파서 선택 및 실행
  let rawPosting: RawJobPosting;
  switch (domainInfo.domain) {
    case 'jobkorea':
      rawPosting = await parseJobKorea(url);
      break;
    case 'catch':
      rawPosting = await parseCatch(url);
      break;
    case 'wanted':
    case 'saramin':
      // Phase 3+ 에서 구현 예정, 현재는 AI fallback
      rawPosting = await fetchAndPrepareForAI(url);
      break;
    default:
      // AI fallback
      rawPosting = await fetchAndPrepareForAI(url);
  }

  // 6. AI 정규화
  let jobPosting: JobPosting;
  if (domainInfo.method === 'ai-fallback' || domainInfo.domain === 'unknown') {
    jobPosting = await parseHtmlWithAI(rawPosting.rawHtml!, url);
  } else {
    jobPosting = await normalizeJobPosting({ raw: rawPosting });
  }

  // 7. 캐시 저장
  await saveUrlCache(url, jobPosting);

  // 8. 결과 반환
  return json({
    success: true,
    data: jobPosting,
    meta: {
      source: domainInfo.domain,
      method: domainInfo.method,
      parsedAt: new Date().toISOString(),
      cached: false,
    },
  });
}
```

### 에러 응답 포맷

```typescript
// 에러 코드 정의
type AnalyzeErrorCode =
  | 'INVALID_URL'          // URL 형식 오류
  | 'FETCH_FAILED'         // HTML 가져오기 실패 (404, 차단 등)
  | 'PARSE_FAILED'         // 파서 데이터 추출 실패
  | 'AI_NORMALIZE_FAILED'  // AI 정규화 실패
  | 'RATE_LIMITED'          // 요청 제한 초과
  | 'UNAUTHORIZED';         // 인증 필요

// 에러 응답
interface AnalyzeError {
  success: false;
  error: {
    code: AnalyzeErrorCode;
    message: string;
    details?: string;
  };
}
```

### 산출물

- `internal/controller/analysis_controller.go`
- `internal/service/crawling_service.go` (파서 라우팅 로직)
- `internal/controller/analysis_controller_test.go`

---

## Phase 완료 체크리스트

- [ ] 잡코리아 URL 파싱 → 구조화된 데이터 추출 성공
- [ ] 캐치 URL 파싱 → 구조화된 데이터 추출 성공
- [ ] 미지원 사이트 URL → AI fallback 파싱 성공
- [ ] JobPosting Zod 스키마 검증 통과
- [ ] `/api/analyze` POST API 정상 동작
- [ ] 도메인별 배지가 표시되는 URL 입력 UI 완성
- [ ] 에러 핸들링 (잘못된 URL, 파싱 실패, AI 실패)
- [ ] 캐시 동작 확인 (동일 URL 1시간 이내 재요청 방지)
- [ ] prompt_templates 시드 데이터 추가 완료
- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

---

## 다음 Phase

→ [Phase 3.1: 기업 데이터 API](./phase-3.1-company-data.md) — DART OpenAPI + 네이버 뉴스 연동으로 기업 기본정보와 최근 뉴스를 수집한다.
