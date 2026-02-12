# Phase 3.1: 기업 데이터 API

> **⚠️ 아키텍처 변경 사항**: 외부 API 연동(DART, 네이버 뉴스)은 Go 백엔드에서 구현합니다.
>
> - `src/app/api/analyze/company-data/route.ts` → `apps/backend/internal/infrastructure/external/`
> - `src/lib/external/` → `apps/backend/internal/infrastructure/external/`
> - `src/lib/cache/` → `apps/backend/internal/service/cache.go`

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | DART 전자공시 OpenAPI와 네이버 뉴스 API를 연동하여 기업 기본정보(개황, 재무)와 최근 뉴스를 자동 수집하는 시스템 구축 |
| **선행 조건** | Phase 3 완료 (크롤링 & 파싱), DART API 키 발급, 네이버 개발자 API 키 발급, 환경변수 설정 (DART_API_KEY, NAVER_CLIENT_ID, NAVER_CLIENT_SECRET) |
| **스프린트** | Sprint 2 — Day 3~4 |
| **관련 기능** | F08 (기업 프로필 자동 조회), F09 (인재상 분석 + 최근 동향) |
| **예상 공수** | 2일 (16시간) |
| **산출물** | DART 클라이언트, 네이버 뉴스 클라이언트, 기업 데이터 집계 API, 캐싱 시스템 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 3.1.1 | DART OpenAPI 클라이언트 | ⬜ |
| 3.1.2 | 네이버 뉴스 클라이언트 | ⬜ |
| 3.1.3 | 기업 데이터 집계 API | ⬜ |
| 3.1.4 | 캐싱 시스템 | ⬜ |

---

## Step 3.1.1: DART OpenAPI 클라이언트

### 목표

DART 전자공시 시스템의 OpenAPI를 활용하여 기업 개황(기본정보)과 재무제표(매출, 영업이익, 당기순이익)를 조회하는 클라이언트를 구현한다. 회사명 → 기업코드(corp_code) 조회를 위한 로컬 매핑도 포함한다.

### 체크리스트

- [ ] DART OpenAPI 키 발급 (https://opendart.fss.or.kr)
- [ ] `dart.ts` 클라이언트 모듈 생성
- [ ] corp_code 조회 기능 구현 (회사명 → 고유번호 매핑)
- [ ] 기업 개황 API (`/api/company.json`) 연동
- [ ] 재무제표 API (`/api/fnlttSinglAcnt.json`) 연동
- [ ] API 응답 타입 정의 (Zod 스키마)
- [ ] 에러 처리 (API 키 만료, 일일 한도 초과, 비상장 기업 등)
- [ ] 단위 테스트 (모킹 기반)

### DART API 상세

```typescript
// src/lib/external/dart.ts

const DART_BASE_URL = 'https://opendart.fss.or.kr/api';

interface DartClient {
  /**
   * 회사명으로 corp_code (고유번호) 조회
   * DART는 corp_code를 ZIP 파일로 제공 → 로컬 JSON 매핑 필요
   * 방법: corpCode.xml 다운로드 → JSON 변환 → 정적 파일 또는 DB 저장
   */
  findCorpCode(companyName: string): Promise<string | null>;

  /**
   * 기업 개황 조회
   * GET /api/company.json?crtfc_key={key}&corp_code={code}
   */
  getCompanyInfo(corpCode: string): Promise<DartCompanyInfo>;

  /**
   * 단일회사 주요계정 (재무제표)
   * GET /api/fnlttSinglAcnt.json?crtfc_key={key}&corp_code={code}&bsns_year={year}&reprt_code=11011
   * reprt_code: 11011(사업보고서), 11012(반기), 11013(1분기), 11014(3분기)
   */
  getFinancials(corpCode: string, year?: number): Promise<DartFinancials>;
}
```

### corp_code 매핑 전략

```typescript
/**
 * DART corp_code 조회 전략:
 *
 * 1. DART 고유번호 XML 다운로드 (https://opendart.fss.or.kr/api/corpCode.xml)
 *    - ZIP 파일 → XML 파싱 → JSON 변환
 *    - 전체 기업 약 80,000건
 *
 * 2. 매핑 저장 위치:
 *    - 옵션 A: src/data/corp-codes.json (빌드 시 포함, ~2MB)
 *    - 옵션 B: Supabase DB corp_codes 테이블 (런타임 조회)
 *    - 권장: 옵션 A (빠른 조회, 월 1회 업데이트 스크립트)
 *
 * 3. 검색 로직:
 *    - 정확한 회사명 매칭 우선
 *    - 부분 매칭 (포함 검색) fallback
 *    - 주식회사/㈜ 등 접두사 정규화 후 비교
 */

// 회사명 정규화 함수
function normalizeCompanyName(name: string): string {
  return name
    .replace(/^(주식회사|㈜|\(주\))\s*/g, '')
    .replace(/\s*(주식회사|㈜|\(주\))$/g, '')
    .trim();
}
```

### Zod 응답 스키마

```typescript
// src/lib/external/dart-schemas.ts

export const DartCompanyInfoSchema = z.object({
  corp_name: z.string(),           // 정식명칭
  corp_name_eng: z.string().optional(), // 영문명칭
  stock_name: z.string().optional(),    // 종목명
  ceo_nm: z.string(),              // 대표자명
  corp_cls: z.enum(['Y', 'K', 'N', 'E']), // Y:유가, K:코스닥, N:코넥스, E:기타
  est_dt: z.string(),              // 설립일 (YYYYMMDD)
  adres: z.string(),               // 주소
  hm_url: z.string().optional(),   // 홈페이지 URL
  induty_code: z.string().optional(), // 업종코드
});

export const DartFinancialsSchema = z.object({
  list: z.array(z.object({
    rcept_no: z.string(),           // 접수번호
    bsns_year: z.string(),          // 사업연도
    account_nm: z.string(),         // 계정명 (매출액, 영업이익 등)
    thstrm_amount: z.string(),      // 당기금액
    frmtrm_amount: z.string(),      // 전기금액
    bfefrmtrm_amount: z.string(),   // 전전기금액
  })),
});

// 정규화된 기업 정보
export const CompanyProfileSchema = z.object({
  name: z.string(),
  nameEng: z.string().optional(),
  ceo: z.string(),
  industry: z.string().optional(),
  founded: z.string().optional(),    // YYYY-MM-DD
  address: z.string().optional(),
  homepage: z.string().optional(),
  listingType: z.enum(['코스피', '코스닥', '코넥스', '비상장']).optional(),
  financials: z.object({
    year: z.number(),
    revenue: z.string().optional(),       // 매출액
    operatingProfit: z.string().optional(), // 영업이익
    netIncome: z.string().optional(),      // 당기순이익
    revenueGrowth: z.string().optional(),  // 매출 성장률
  }).optional(),
});

export type CompanyProfile = z.infer<typeof CompanyProfileSchema>;
```

### 에러 처리

```typescript
// DART API 에러 코드 매핑
const DART_ERROR_CODES: Record<string, string> = {
  '000': '정상',
  '010': '등록되지 않은 키',
  '011': '사용할 수 없는 키',
  '012': '접근할 수 없는 IP',
  '013': '조회된 데이터가 없음',     // 비상장 중소기업에서 빈번
  '020': '요청 제한 초과',
  '100': '필드 오류',
  '800': '원활한 공시서비스를 위해 접속 제한',
  '900': '정의되지 않은 오류',
};
```

### 검증 방법

- [ ] 상장 대기업 (삼성전자) corp_code 조회 → 기업 개황 정상 반환
- [ ] 재무제표 조회 → 매출액, 영업이익, 당기순이익 정상 파싱
- [ ] 비상장 기업 → "조회된 데이터가 없음" graceful 처리 확인
- [ ] 회사명 정규화 ("주식회사 삼성전자" → "삼성전자") 동작 확인
- [ ] API 키 누락 시 적절한 에러 메시지 확인

### 산출물

- `src/lib/external/dart.ts`
- `src/lib/external/dart-schemas.ts`
- `src/data/corp-codes.json` (또는 생성 스크립트)
- `scripts/update-corp-codes.ts` (corp_code JSON 업데이트 스크립트)
- `__tests__/lib/external/dart.test.ts`

---

## Step 3.1.2: 네이버 뉴스 클라이언트

### 목표

네이버 검색 API를 활용하여 특정 기업의 최근 3개월 뉴스를 검색하고, 상위 5~10건의 뉴스 제목/요약/링크를 수집한다. 기업 분석 시 최근 동향 파악의 입력 데이터로 활용한다.

### 체크리스트

- [ ] 네이버 개발자 API 키 발급 (https://developers.naver.com)
- [ ] `naver-news.ts` 클라이언트 모듈 생성
- [ ] 뉴스 검색 API 연동 (`/v1/search/news.json`)
- [ ] 검색 쿼리 최적화 (기업명 + 채용/경영/실적 키워드)
- [ ] 검색 결과 정리 (HTML 태그 제거, 날짜 정규화)
- [ ] 최근 3개월 필터링
- [ ] 상위 5~10건 반환
- [ ] 에러 처리 (API 한도 초과, 네트워크 오류)

### API 상세

```typescript
// src/lib/external/naver-news.ts

const NAVER_SEARCH_URL = 'https://openapi.naver.com/v1/search/news.json';

interface NaverNewsClient {
  /**
   * 기업명으로 뉴스 검색
   *
   * @param companyName - 검색할 기업명
   * @param options - 검색 옵션
   * @returns 정제된 뉴스 목록
   *
   * API 파라미터:
   * - query: 검색어 (기업명)
   * - display: 검색 결과 출력 건수 (최대 100)
   * - start: 검색 시작 위치 (1~1000)
   * - sort: sim(정확도순) | date(최신순)
   */
  searchNews(companyName: string, options?: SearchOptions): Promise<NewsItem[]>;
}

interface SearchOptions {
  maxResults?: number;    // 기본값 10
  monthsBack?: number;    // 기본값 3
  sort?: 'sim' | 'date';  // 기본값 'date'
  additionalKeywords?: string[]; // 추가 검색어 (예: ['채용', '경영'])
}

interface NewsItem {
  title: string;          // HTML 태그 제거된 제목
  description: string;    // HTML 태그 제거된 요약
  link: string;           // 원문 URL
  pubDate: string;        // ISO 8601 날짜
  source?: string;        // 출처 (가능한 경우)
}
```

### 검색 쿼리 전략

```typescript
/**
 * 기업 뉴스 검색 쿼리 생성
 *
 * 전략:
 * 1. 기본 검색: "삼성전자" (기업명만)
 * 2. 정제된 검색: "삼성전자 (채용 OR 경영 OR 실적 OR 투자 OR 사업)"
 * 3. 노이즈 제외: 주식/시세 관련 뉴스 필터링 (후처리)
 *
 * 검색 결과 후처리:
 * - HTML 태그 제거 (<b>, </b> 등)
 * - 3개월 이전 뉴스 필터링
 * - 중복 뉴스 제거 (제목 유사도 기반)
 * - 주식/시세 관련 뉴스 필터링 (키워드 기반)
 */

function buildSearchQuery(companyName: string, keywords?: string[]): string {
  const defaultKeywords = ['채용', '경영', '실적', '투자', '사업', '성장'];
  const kw = keywords ?? defaultKeywords;
  return `"${companyName}" (${kw.join(' OR ')})`;
}

function cleanHtml(text: string): string {
  return text.replace(/<[^>]*>/g, '').replace(/&[a-z]+;/g, '');
}

function filterStockNews(news: NewsItem[]): NewsItem[] {
  const STOCK_KEYWORDS = ['주가', '시세', '종목', '매수', '매도', '코스피', '코스닥', '상한가', '하한가'];
  return news.filter(item =>
    !STOCK_KEYWORDS.some(kw => item.title.includes(kw))
  );
}
```

### 검증 방법

- [ ] 대기업명 (삼성전자) 검색 → 최근 뉴스 5건 이상 반환 확인
- [ ] HTML 태그 제거 정상 동작 확인 (`<b>삼성</b>` → `삼성`)
- [ ] 3개월 이전 뉴스 필터링 확인
- [ ] 주식/시세 뉴스 필터링 확인
- [ ] 존재하지 않는 기업명 → 빈 배열 반환 확인
- [ ] API 키 누락 시 적절한 에러 메시지 확인

### 산출물

- `src/lib/external/naver-news.ts`
- `__tests__/lib/external/naver-news.test.ts`

---

## Step 3.1.3: 기업 데이터 집계 API

### 목표

DART 기업정보와 네이버 뉴스를 `Promise.all`로 병렬 호출하여 통합된 기업 데이터를 반환하는 API를 구현한다. 일부 소스 실패 시에도 나머지 데이터로 partial response를 반환한다.

### 체크리스트

- [ ] `/api/analyze/company-data` POST 라우트 생성
- [ ] DART + 네이버 뉴스 병렬 호출 (`Promise.allSettled`)
- [ ] Partial failure 처리 (일부 소스 실패해도 결과 반환)
- [ ] 통합 응답 스키마 정의
- [ ] 각 소스별 상태 표시 (성공/실패/건너뜀)
- [ ] 인증 미들웨어 적용
- [ ] 에러 핸들링 (전체 실패 시 에러 응답)

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `POST` | `/api/analyze/company-data` | `{ companyName: string, corpCode?: string }` | `{ success: boolean, data: CompanyData, sources: SourceStatus[] }` |

### 통합 응답 스키마

```typescript
// src/lib/external/company-data-schemas.ts

interface CompanyData {
  profile: CompanyProfile | null;   // DART 기업 개황
  financials: Financials | null;    // DART 재무제표
  news: NewsItem[];                 // 네이버 뉴스 (빈 배열 가능)
}

interface SourceStatus {
  source: 'dart_company' | 'dart_financials' | 'naver_news';
  status: 'success' | 'failed' | 'skipped';
  error?: string;
  fetchedAt: string;
}
```

### Promise.allSettled 패턴

```typescript
// src/app/api/analyze/company-data/route.ts

export async function POST(request: Request) {
  const { companyName, corpCode } = await request.json();

  // 1. corp_code 조회 (없으면 회사명으로 검색)
  const resolvedCorpCode = corpCode ?? await dartClient.findCorpCode(companyName);

  // 2. 데이터 소스 병렬 호출
  const [companyResult, financialsResult, newsResult] = await Promise.allSettled([
    // DART 기업 개황 (corp_code 있을 때만)
    resolvedCorpCode
      ? dartClient.getCompanyInfo(resolvedCorpCode)
      : Promise.reject(new Error('corp_code not found')),

    // DART 재무제표 (corp_code 있을 때만)
    resolvedCorpCode
      ? dartClient.getFinancials(resolvedCorpCode)
      : Promise.reject(new Error('corp_code not found')),

    // 네이버 뉴스 (항상 시도)
    naverNewsClient.searchNews(companyName, { maxResults: 10, monthsBack: 3 }),
  ]);

  // 3. 결과 조합 (partial failure 허용)
  const data: CompanyData = {
    profile: companyResult.status === 'fulfilled' ? companyResult.value : null,
    financials: financialsResult.status === 'fulfilled' ? financialsResult.value : null,
    news: newsResult.status === 'fulfilled' ? newsResult.value : [],
  };

  // 4. 소스 상태 생성
  const sources: SourceStatus[] = [
    buildSourceStatus('dart_company', companyResult),
    buildSourceStatus('dart_financials', financialsResult),
    buildSourceStatus('naver_news', newsResult),
  ];

  // 5. 전체 실패 확인 (모든 소스 실패 시 에러)
  const allFailed = sources.every(s => s.status === 'failed');
  if (allFailed) {
    return json({ success: false, error: '모든 데이터 소스 조회에 실패했습니다.' }, 500);
  }

  return json({ success: true, data, sources });
}
```

### 에러 시나리오

| 시나리오 | DART 개황 | DART 재무 | 뉴스 | 결과 |
|----------|-----------|-----------|------|------|
| 상장 대기업 | ✅ | ✅ | ✅ | 완전한 데이터 |
| 비상장 기업 | ✅ | ❌ (재무 없음) | ✅ | profile + news (재무 null) |
| 미등록 스타트업 | ❌ (corp_code 없음) | ❌ | ✅ | news만 (profile, financials null) |
| 전체 실패 | ❌ | ❌ | ❌ | 500 에러 |

### 검증 방법

- [ ] 상장 대기업 → DART + 뉴스 모두 성공 확인
- [ ] 비상장 기업 → DART 개황 성공, 재무 실패 (partial) 확인
- [ ] 미등록 기업 → 뉴스만 성공 (partial) 확인
- [ ] 각 소스 상태(success/failed/skipped) 정확히 반환 확인
- [ ] 전체 실패 시 500 에러 응답 확인
- [ ] 응답 시간 3초 이내 확인 (병렬 호출)

### 산출물

- `src/app/api/analyze/company-data/route.ts`
- `src/lib/external/company-data-schemas.ts`
- `__tests__/app/api/analyze/company-data.test.ts`

---

## Step 3.1.4: 캐싱 시스템

### 목표

기업 분석 결과를 `company_analysis_cache` 테이블에 7일 TTL로 캐싱하여, 동일 기업에 대한 반복 API 호출을 방지하고 응답 속도를 개선한다.

### 체크리스트

- [ ] `company_analysis_cache` 테이블 생성 마이그레이션
- [ ] 캐시 조회 함수 구현 (TTL 확인 포함)
- [ ] 캐시 저장 함수 구현
- [ ] 캐시 무효화 함수 구현 (수동 새로고침용)
- [ ] 만료 캐시 자동 정리 (Supabase cron 또는 조회 시 lazy deletion)
- [ ] `/api/analyze/company-data`에 캐시 로직 통합
- [ ] 캐시 히트/미스 로깅

### DB 마이그레이션

```sql
-- supabase/migrations/YYYYMMDDHHMMSS_create_company_analysis_cache.sql

CREATE TABLE company_analysis_cache (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_name TEXT NOT NULL,
  company_name_normalized TEXT NOT NULL, -- 정규화된 회사명 (검색용)
  corp_code TEXT,                         -- DART 고유번호

  -- 캐시 데이터
  profile_data JSONB,                     -- DART 기업 개황
  financials_data JSONB,                  -- DART 재무제표
  news_data JSONB,                        -- 네이버 뉴스
  source_statuses JSONB NOT NULL DEFAULT '[]', -- 각 소스 상태

  -- TTL 관리
  cached_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '7 days'),

  -- 메타
  hit_count INTEGER NOT NULL DEFAULT 0,
  last_hit_at TIMESTAMPTZ,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 인덱스
CREATE INDEX idx_cache_company_name ON company_analysis_cache (company_name_normalized);
CREATE INDEX idx_cache_expires_at ON company_analysis_cache (expires_at);
CREATE INDEX idx_cache_corp_code ON company_analysis_cache (corp_code) WHERE corp_code IS NOT NULL;

-- 만료 캐시 정리 함수 (Supabase pg_cron으로 일 1회 실행)
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS void AS $$
BEGIN
  DELETE FROM company_analysis_cache
  WHERE expires_at < now();
END;
$$ LANGUAGE plpgsql;
```

### 캐시 유틸리티

```typescript
// src/lib/cache/company-cache.ts

interface CacheOptions {
  ttlDays?: number;          // 기본값 7
  forceRefresh?: boolean;    // true면 캐시 무시
}

/**
 * 캐시 조회
 * - 정규화된 회사명으로 검색
 * - TTL 만료 확인
 * - 히트 카운트 증가
 */
async function getCachedCompanyData(
  companyName: string
): Promise<CompanyData | null>;

/**
 * 캐시 저장
 * - UPSERT (동일 회사명 존재 시 업데이트)
 * - TTL 자동 설정
 */
async function cacheCompanyData(
  companyName: string,
  corpCode: string | null,
  data: CompanyData,
  sources: SourceStatus[]
): Promise<void>;

/**
 * 캐시 무효화
 * - 특정 기업의 캐시 삭제
 * - 사용자가 "새로고침" 요청 시 사용
 */
async function invalidateCache(companyName: string): Promise<void>;
```

### 캐시 통합 흐름

```
[/api/analyze/company-data 요청]
    │
    ▼
[캐시 조회] ──── 히트 ──→ [캐시 데이터 반환] (hit_count++)
    │
    │ 미스 또는 만료
    ▼
[DART + 네이버 뉴스 API 호출]
    │
    ▼
[결과 캐시 저장] (expires_at = now + 7일)
    │
    ▼
[결과 반환]
```

### 검증 방법

- [ ] 첫 요청 시 API 호출 + 캐시 저장 확인
- [ ] 두 번째 요청 시 캐시 히트 확인 (API 미호출)
- [ ] 7일 후 캐시 만료 → API 재호출 확인
- [ ] 캐시 강제 새로고침 (`forceRefresh: true`) 동작 확인
- [ ] 회사명 정규화 ("주식회사 카카오" = "카카오") 동일 캐시 조회 확인
- [ ] hit_count 정상 증가 확인
- [ ] 만료 캐시 정리 함수 동작 확인

### 산출물

- `supabase/migrations/YYYYMMDDHHMMSS_create_company_analysis_cache.sql`
- `src/lib/cache/company-cache.ts`
- `__tests__/lib/cache/company-cache.test.ts`

---

## Phase 완료 체크리스트

- [ ] DART OpenAPI로 상장 기업 개황 조회 성공
- [ ] DART OpenAPI로 재무제표 (매출, 영업이익, 순이익) 조회 성공
- [ ] 네이버 뉴스 API로 기업 관련 최근 뉴스 검색 성공
- [ ] 비상장/미등록 기업에 대한 partial failure 정상 처리
- [ ] 병렬 호출 (`Promise.allSettled`) 정상 동작
- [ ] `/api/analyze/company-data` API 응답 정상
- [ ] 캐시 저장/조회/만료/무효화 정상 동작
- [ ] 캐시 히트 시 API 미호출 확인
- [ ] 환경변수 설정 문서 업데이트 (DART_API_KEY, NAVER_CLIENT_ID, NAVER_CLIENT_SECRET)
- [ ] 단위 테스트 + 통합 테스트 통과

---

## 다음 Phase

→ [Phase 3.2: AI 기업 분석](./phase-3.2-ai-analysis.md) — 수집된 기업 데이터를 바탕으로 Claude Sonnet 4.5을 활용하여 인재상, 핵심가치, 전략 키워드 등의 종합 분석 리포트를 스트리밍으로 생성한다.
