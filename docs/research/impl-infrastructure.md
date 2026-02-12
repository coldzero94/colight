> **⚠️ Superseded**: 이 문서는 초기 아키텍처(Next.js 풀스택 + Vercel Serverless)를 기준으로 작성되었습니다. 현재 아키텍처(Go 백엔드 + Koyeb)의 인프라 설계는 `docs/develop/01-architecture.md`를 참조하세요.
> - ~~Vercel API Routes~~ → Go Gin API Server (Koyeb)
> - ~~Supabase RLS~~ → Go 미들웨어 JWT 검증 (RLS 미사용)
> - ~~Vercel Serverless 타임아웃~~ → Koyeb 자체 타임아웃
> - 무료 티어 스펙, 비용 추정 등 일부 정보는 여전히 유효합니다.

# 인프라 & 기술 스택

> 작성일: 2026-02-09

---

## 1. 무료 티어 스펙

### Vercel (프론트엔드 + API Routes)

| 항목 | 무료 티어 제한 |
|------|---------------|
| 배포 | 일 100회 |
| Serverless Function 타임아웃 | 10초 (Hobby) |
| Bandwidth | 월 100GB |
| Fluid Compute | 최대 60초 (무료) |

### Supabase (데이터베이스 + 인증)

| 항목 | 무료 티어 제한 |
|------|---------------|
| 프로젝트 수 | 2개 |
| DB 용량 | 500MB |
| Storage | 1GB |
| Bandwidth | 5GB |
| MAU | 50,000명 |
| 자동 중지 | 1주 비활동시 |

### Koyeb (백엔드 워커)

| 항목 | 무료 티어 제한 |
|------|---------------|
| 인스턴스 | 1개 |
| CPU | 0.1 vCPU |
| RAM | 512MB |
| SSD | 2GB |
| Scale-to-Zero | 1시간 무트래픽시 |

---

## 2. 요청 흐름

```
[사용자: 채용공고 URL 입력]
    |
    v
[Vercel API Route: /api/analyze]
    |
    +-- URL 도메인 판별
    |
    +-- [Case A: 정적 사이트] → Cheerio 파싱 (Vercel에서 직접)
    +-- [Case B: 동적 사이트] → Koyeb 워커에 크롤링 요청
    +-- [Case C: 사람인]     → 사람인 API 직접 호출
    |
    v
[기업 정보 병합] → DART API + 네이버 뉴스 + 인재상 사전 DB
    |
    v
[AI 종합 분석 (스트리밍)] → 결과 캐시 저장 + 사용자에게 반환
```

---

## 3. Vercel 10초 타임아웃 대응

- **전략 1: 스트리밍 응답** - AI 분석에 사용, 타임아웃에 유리
- **전략 2: 비동기 작업 + 폴링** - 무거운 크롤링은 Koyeb에 비동기 요청, 프론트에서 결과 폴링

---

## 4. 비용 최적화

### 기업 분석 결과 캐싱
- 같은 기업은 재분석 불필요 (7일 TTL)
- 채용공고 파싱 결과, 기업 기본정보 각각 개별 캐싱

### Supabase 500MB 용량 관리
- 만료 데이터 자동 정리 (Edge Function으로 주기적 실행)

---

## 5. 무료 티어 처리 가능 규모

| 항목 | 일일 처리 가능량 | 근거 |
|------|-----------------|------|
| 채용공고 분석 | ~100건/일 | Vercel 함수 호출 + AI API 비용 기준 |
| 사람인 API | 500건/일 | API 제한 |
| 네이버 뉴스 | 25,000건/일 | API 제한 |
| DART 기업정보 | 10,000건/일 | API 제한 |
| DB 저장 용량 | ~500개 기업 캐시 | 기업당 ~1MB 기준 |
| 동시 사용자 | ~50명 | Koyeb 0.1vCPU 제한 |

---

## 6. 기술 스택 요약

| 레이어 | 기술 | 용도 |
|--------|------|------|
| 프론트엔드 | Next.js (Vercel) | UI, API Routes |
| DB | Supabase PostgreSQL | 사용자 데이터, 캐시 |
| 인증 | Supabase Auth | 로그인/회원가입 |
| 정적 크롤링 | Cheerio | 잡코리아, 캐치 파싱 |
| 동적 크롤링 | Playwright (Koyeb) | 원티드 파싱 |
| AI 분석 | Claude/GPT API | 기업 분석, 매칭 |
| 임베딩 | text-embedding-3-small | 경험 유사도 계산 |
| 기업 데이터 | DART OpenAPI | 기업 기본정보, 재무 |
| 뉴스 수집 | 네이버 검색 API | 기업 관련 뉴스 |
| 채용 데이터 | 사람인 API | 채용공고 검색 |

### 참고 자료
- [Supabase Pricing](https://supabase.com/pricing)
- [Vercel Pricing](https://vercel.com/pricing)
- [Koyeb Pricing](https://www.koyeb.com/pricing)
- [Cheerio 공식 문서](https://cheerio.js.org/)
- [Playwright 공식 문서](https://playwright.dev/)
