# 리서치 자료

> Colight 서비스 기획을 위한 시장 조사, 기술 검토, 도메인 연구 자료
>
> **참고**: 일부 문서는 아키텍처 변경(Next.js fullstack → Go 백엔드) 이후 superseded 처리되었습니다.
> 해당 문서 상단에 안내가 있으며, 최신 내용은 `docs/develop/`를 참조하세요.

---

## 자소서 작성법 리서치

| 파일 | 내용 | 핵심 |
|------|------|------|
| [jasoseo-writing-techniques.md](./jasoseo-writing-techniques.md) | 실전 작성 기법 | STAR, 두괄식, 스토리텔링, 합격/불합격 분석, HR 평가 기준 |
| [jasoseo-ai-era-strategy.md](./jasoseo-ai-era-strategy.md) | AI 시대 전략 | AI 활용법, 탐지 현황, 차별화 전략, 업종별 포인트 |

## 시장 리서치

| 파일 | 내용 | 핵심 |
|------|------|------|
| [market-ai-services.md](./market-ai-services.md) | 기존 AI 자소서 서비스 분석 | 잡메이커, 딥레쥬메, 사람인, 코멘토 등 8개 서비스 비교 |
| [market-insights.md](./market-insights.md) | 시장 인사이트 & 기회 | 시장 Gap, 기술 동향, 경쟁 분석 |

## 기술 구현 리서치

| 파일 | 내용 | 상태 |
|------|------|------|
| [impl-crawling-parsing.md](./impl-crawling-parsing.md) | 채용공고 크롤링/파싱 | ⚠️ Superseded — Cheerio→goquery 변경. `04-ai-pipeline.md` 참조 |
| [impl-company-data.md](./impl-company-data.md) | 기업 정보 수집 | 유효 — DART API, 네이버 뉴스 API, 인재상 DB |
| [impl-ai-pipeline.md](./impl-ai-pipeline.md) | AI 분석 파이프라인 | ⚠️ Superseded — TS→Go 변경. `04-ai-pipeline.md` 참조 |
| [impl-infrastructure.md](./impl-infrastructure.md) | 인프라 & 기술 스택 | ⚠️ Superseded — Next.js fullstack→Go+Next.js. `01-architecture.md` 참조 |

## 비동기 작업 & River 리서치

| 파일 | 내용 | 핵심 |
|------|------|------|
| [async-job-processing-options.md](./async-job-processing-options.md) | 비동기 작업 처리 옵션 비교 | River vs Asynq vs Temporal, 결론: River embedded mode |
| [river-job-queue-licensing.md](./river-job-queue-licensing.md) | River 라이선싱 조사 | MPL-2.0 무료 사용 가능, Pro는 선택적 |

## 경험 무기 시스템

| 파일 | 내용 | 핵심 |
|------|------|------|
| [weapon-classification.md](./weapon-classification.md) | 무기 분류 체계 | 7대 무기 + 소분류, 자동 분류 로직 |
| [weapon-db-prompts.md](./weapon-db-prompts.md) | DB 레벨 프롬프트 설계 | 프롬프트 DB 스키마, 핵심 템플릿 4종 |
| [weapon-features-flow.md](./weapon-features-flow.md) | 기능 아이디어 & 플로우 | 문항 해독기, 레이더 차트, 전체 플로우 |

## 원본 계획 자료

| 파일 | 내용 | 상태 |
|------|------|------|
| [plan-tech-requirements.md](./plan-tech-requirements.md) | 기술 구현 요구사항 원본 | ⚠️ Superseded — `docs/develop/` 기술 문서로 확장됨 |
| [plan-dev-roadmap.md](./plan-dev-roadmap.md) | 개발 로드맵 원본 | ⚠️ Partially Superseded — `docs/develop/phases/`로 분리됨. RLS 관련 내용 제거 필요 |
