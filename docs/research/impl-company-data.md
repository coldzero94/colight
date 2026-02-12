# 기업 정보 수집 방법

> 작성일: 2026-02-09

---

## 1. 데이터 소스별 수집 방법

### DART 전자공시 OpenAPI (공공데이터, 무료)

| 항목 | 내용 |
|------|------|
| 기본 URL | `https://opendart.fss.or.kr/api/` |
| 인증 | API 인증키 발급 (이메일 인증 후 즉시 발급) |
| 비용 | 무료 |
| 제한 | 일 10,000건 |

**주요 API 엔드포인트**:
- 기업개황: `/api/company.json`
- 공시검색: `/api/list.json`
- 사업보고서: `/api/fnlttSinglAcnt.json`

**수집 가능 데이터**:
- 기업 기본정보: 대표자명, 설립일, 업종, 주소, 홈페이지 URL
- 재무정보: 매출액, 영업이익, 당기순이익
- 사업보고서: 사업 내용, 임직원 현황, 주주 현황

### 네이버 검색 API (뉴스 수집)

| 항목 | 내용 |
|------|------|
| API URL | `https://openapi.naver.com/v1/search/news.json` |
| 인증 | Client ID + Client Secret (네이버 개발자센터 발급) |
| 비용 | 무료 |
| 제한 | 일 25,000회, 1회 최대 100건 |

### 기업 공식 홈페이지 (인재상/핵심가치)

주요 대기업 인재상 페이지:
- 삼성: `samsung.com/sec/aboutsamsung/company/talent/`
- LG: `lgcareers.com/about/talentPromotion`
- SK: `recruit.sk.com/talent/`
- 현대: `talent.hyundai.com/htalent/`

**수집 전략**:
- 주요 대기업 인재상은 사전 DB로 구축 (수동 + AI 정리)
- 중소기업은 공고 내 키워드에서 AI가 추출

### 잡플래닛 (기업 리뷰/문화)

- 공식 API 없음, 로그인 필요한 데이터 많음
- **대안**: 사용자가 리뷰를 직접 복사-붙여넣기하면 AI가 분석

---

## 2. 데이터 소스 우선순위

```
[1순위: 공식 API] DART OpenAPI + 네이버 뉴스 API (무료, 합법, 안정적)
        |
[2순위: 사전 DB] 주요 기업 인재상/핵심가치 (수동 구축 + 주기적 업데이트)
        |
[3순위: 실시간 파싱] 사용자가 입력한 URL에서 정보 추출
        |
[4순위: 사용자 입력] 사용자가 직접 기업 정보를 입력/보완
```

### 참고 자료
- [DART OpenAPI](https://opendart.fss.or.kr/intro/main.do)
- [네이버 검색 API](https://developers.naver.com/docs/serviceapi/search/news/news.md)
