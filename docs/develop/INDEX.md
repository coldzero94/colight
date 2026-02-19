# 개발 가이드

> Colight 개발을 위한 기술 문서 + Phase별 구현 가이드

---

## Quick Start

```bash
# 1. 프로젝트 클론 + 의존성 설치
git clone <repo> && cd colight

# 2. 환경변수 설정
cp .env.example .env.local  # 값 채우기

# 3. 개발 서버 (moon 사용 — 프로젝트 루트에서 실행)
moon run :dev
```

---

## 기술 문서

| # | 파일 | 내용 |
|---|------|------|
| 01 | [시스템 아키텍처](./01-architecture.md) | 전체 구조도, 요청 흐름, 프로젝트 디렉토리 |
| 02 | [DB 스키마](./02-data-structure.md) | 15개 테이블 DDL, RLS 정책, 인덱스, 시드 데이터 |
| 03 | [API 설계](./03-api-design.md) | 엔드포인트 목록, 요청/응답 포맷, 스트리밍 패턴 |
| 04 | [AI 파이프라인](./04-ai-pipeline.md) | 모델 전략, 프롬프트 관리, 파이프라인 정의 |
| 05 | [프론트엔드 패턴](./05-frontend-patterns.md) | 라우트 구조, 상태관리, 컴포넌트, 스트리밍 UI |
| 06 | [마이그레이션 전략](./06-migration-strategy.md) | Supabase CLI 워크플로, Phase별 마이그레이션 |
| 07 | [환경 설정](./07-environment-config.md) | API 키 7개, .env.local 템플릿 |
| 08 | [테스트 전략](./08-testing-strategy.md) | Vitest + Playwright, AI 모킹, CI/CD |
| 09 | [에러 처리](./09-error-handling.md) | API/AI/외부API 에러 컨벤션 |
| 10 | [배포 & CI/CD](./10-deployment-ci.md) | Koyeb/Vercel 배포, GitHub Actions CI |
| 11 | [로깅 & 모니터링](./11-logging-monitoring.md) | slog 구조화 로깅, 헬스체크, 알림 |
| 12 | [Go 백엔드 테스트](./12-backend-testing.md) | testify + enttest, AI 모킹, 커버리지 |
| 13 | [데이터 내보내기 & 백업](./13-data-export-backup.md) | 사용자 내보내기, 계정 삭제, DB 백업, 재해 복구 |
| 14 | [크롤러 모니터링](./14-crawler-monitoring.md) | HTML 변경 감지, 실패 대응, 헬스체크, 관리자 대시보드 |
| 15 | [UX 개선 백로그](./15-ux-improvement-backlog.md) | 접근성, 성능, 모바일 UX 개선 항목 추적 |

## Phase 가이드

> [Phase 상태 테이블 & 의존관계](./phases/README.md)

### MVP (Sprint 0-6, ~6주)

| Phase | 이름 | 스프린트 | 공수 | 상태 |
|-------|------|----------|------|------|
| [Phase 0](./phases/phase-0-project-setup.md) | 프로젝트 셋업 & 인프라 | Sprint 0 | 4일 | ✅ |
| [Phase 1](./phases/phase-1-auth-layout.md) | 인증 & 레이아웃 | Sprint 0 | 4일 | ✅ |
| [Phase 2](./phases/phase-2-experience-crud.md) | 경험 CRUD | Sprint 1 | 3일 | ✅ |
| [Phase 2.1](./phases/phase-2.1-weapon-tagging.md) | 무기 자동 태깅 | Sprint 1 | 2일 | ✅ |
| [Phase 3](./phases/phase-3-crawling-parsing.md) | 크롤링 & 파싱 | Sprint 2 | 2일 | ✅ |
| [Phase 3.1](./phases/phase-3.1-company-data.md) | 기업 데이터 API | Sprint 2 | 2일 | ✅ |
| [Phase 3.2](./phases/phase-3.2-ai-analysis.md) | AI 기업 분석 | Sprint 3 | 2일 | ✅ |
| [Phase 3.3](./phases/phase-3.3-analysis-ui.md) | 분석 리포트 UI | Sprint 3 | 2일 | ✅ |
| [Phase 4](./phases/phase-4-matching.md) | 경험 매칭 | Sprint 4 | 5일 | ✅ |
| [Phase 5](./phases/phase-5-question-analysis.md) | 문항 분석 | Sprint 5 | 2일 | ✅ |
| [Phase 5.1](./phases/phase-5.1-draft-coaching.md) | 초안 코칭 | Sprint 5 | 2일 | ✅ |
| [Phase 5.2](./phases/phase-5.2-coaching-editor.md) | 코칭 에디터 | Sprint 5 | 1일 | ✅ |
| [Phase 6](./phases/phase-6-review-coaching.md) | 첨삭 코칭 | Sprint 6 | 2일 | ✅ |
| [Phase 6.1](./phases/phase-6.1-freemium-polish.md) | 프리미엄 & 마무리 | Sprint 6 | 2일 | ✅ |
| [Phase 6.2](./phases/phase-6.2-landing-beta.md) | 랜딩 & 베타 | Sprint 6 | 1일 | ✅ |

### Post-MVP (Sprint 7+)

| Phase | 이름 | 스프린트 | 공수 | 상태 |
|-------|------|----------|------|------|
| [Phase 7](./phases/phase-7-ai-interview.md) | AI 인터뷰 | Sprint 7 | 2일 | ✅ |
| [Phase 7.1](./phases/phase-7.1-experience-recommend.md) | 경험 추천 | Sprint 7 | 1일 | ⬜ |
| [Phase 1.6](./phases/phase-1.6-admin-system.md) | 어드민 시스템 확장 | Sprint 8-9 | 5일 | ⬜ |
| [Phase 1.7](./phases/phase-1.7-admin-dashboard.md) | 어드민 관찰성 대시보드 | Sprint 9-10 | 6일 | ⬜ |
| [Phase 8](./phases/phase-8-dashboard-kanban.md) | 대시보드 칸반 | Sprint 7-8 | 1.5일 | ⬜ |
| [Phase 8.1](./phases/phase-8.1-version-management.md) | 버전 관리 | Sprint 8 | 1일 | ⬜ |
| [Phase 9](./phases/phase-9-payment.md) | 결제 연동 | Sprint 9 | 2일 | ⬜ |
| [Phase 10](./phases/phase-10-growth.md) | 성장 기능 | Sprint 10+ | 진행중 | ⬜ |

---

## 읽기 순서

1. [CLAUDE.md](../../CLAUDE.md) - 프로젝트 컨텍스트
2. [01-architecture.md](./01-architecture.md) - 전체 구조 이해
3. [02-data-structure.md](./02-data-structure.md) - DB 스키마
4. [phases/README.md](./phases/README.md) - Phase 상태 확인
5. 현재 Phase 문서 → 구현 시작
