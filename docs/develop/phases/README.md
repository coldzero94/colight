# Colight 개발 Phase 가이드

## 개요

각 Phase는 독립적인 기능 단위로, **TDD(Test-Driven Development)** 기반으로 진행됩니다.
각 스텝은 테스트 명세 → 테스트 작성 (RED) → 구현 (GREEN) → 리팩터링 순서를 따릅니다.

## Phase 상태

| Phase | 이름 | 상태 | 스프린트 | 공수 | 관련 기능 |
|-------|------|------|----------|------|----------|
| Phase 0 | 프로젝트 셋업 & 인프라 | ✅ | Sprint 0 | 4일 | - |
| Phase 1 | 인증 & 레이아웃 | ✅ | Sprint 0 | 4일 | Sub-phases: 1.1~1.6 |
| Phase 2 | 경험 CRUD | ✅ | Sprint 1 | 3일 | F01 |
| Phase 2.1 | 무기 자동 태깅 | ✅ | Sprint 1 | 2일 | F03 |
| Phase 3 | 크롤링 & 파싱 | ✅ | Sprint 2 | 2일 | F07 |
| Phase 3.1 | 기업 데이터 API | ✅ | Sprint 2 | 2일 | F08 |
| Phase 3.2 | AI 기업 분석 | ✅ | Sprint 3 | 2일 | F09 |
| Phase 3.3 | 분석 리포트 UI | ✅ | Sprint 3 | 2일 | F09 |
| Phase 4 | 경험 매칭 | ✅ | Sprint 4 | 5일 | F10 |
| Phase 5 | 문항 분석 | ✅ | Sprint 5 | 2일 | F11 |
| Phase 5.1 | 초안 코칭 | ✅ | Sprint 5 | 2일 | F12, F13 |
| Phase 5.2 | 코칭 에디터 | ✅ | Sprint 5 | 1일 | F13 |
| Phase 6 | 첨삭 코칭 | ✅ | Sprint 6 | 2일 | F14 |
| Phase 6.1 | 프리미엄 & 마무리 | ✅ | Sprint 6 | 2일 | - |
| Phase 6.2 | 랜딩 & 베타 | ✅ | Sprint 6 | 1일 | - |
| Phase 7 | AI 인터뷰 | ✅ | Sprint 7 | 2일 | F02 |
| Phase 7.1 | 경험 추천 강화 | ✅ | Sprint 7 | 1일 | F12 |
| Phase 8 | 대시보드 칸반 | ✅ | Sprint 7-8 | 1.5일 | F18 |
| Phase 8.1 | 버전 관리 | ✅ | Sprint 8 | 1일 | F20 |
| Phase 9 | 결제 연동 | ⬜ | Sprint 8 | 2일 | - |
| Phase 10 | 성장 기능 | ⬜ | Sprint 9+ | 진행중 | F04, F05, F15-F17 |

상태: ⬜ 대기 | 🟡 진행중 | ✅ 완료

---

## 의존 관계

```
Phase 0 (셋업)
  └─→ Phase 1 (인증)
        └─→ Phase 2 (경험 CRUD)
              ├─→ Phase 2.1 (무기 태깅)     ← 병렬 분기 A
              │     ├─→ Phase 7 (AI 인터뷰)
              │     └─→ Phase 7.1 (추천 강화)
              │
              └─→ Phase 3 (크롤링)           ← 병렬 분기 B
                    │
                    ├─→ Phase 3.1 (기업 데이터)  ⚡ Phase 3과 병렬 가능
                    │     └─→ Phase 3.2 (AI 분석)
                    │           └─→ Phase 3.3 (분석 UI)
                    │
                    └─→ Phase 8 (칸반)

Phase 2.1 (무기 태깅) + Phase 3.3 (분석 UI)
  └─→ Phase 4 (매칭)    ← ⚠️ Phase 2.1 필수 (무기 태그가 매칭 점수의 40% 차지)
        └─→ Phase 5 (문항 분석)
              └─→ Phase 5.1 (초안 코칭)
                    └─→ Phase 5.2 (코칭 에디터)
                          └─→ Phase 6 (첨삭 코칭)
                                └─→ Phase 6.1 (프리미엄)
                                      ├─→ Phase 6.2 (랜딩)
                                      ├─→ Phase 8.1 (버전 관리)
                                      └─→ Phase 9 (결제)

Phase 10 (성장) ← MVP 완료 후 독립 진행
```

### 주요 의존성 참고 사항

| 항목 | 설명 |
|------|------|
| **Phase 3 ↔ 3.1 병렬** | Phase 3(채용공고 크롤링)과 Phase 3.1(기업 데이터 API)은 서로 다른 소스를 크롤링하므로 **동시 진행 가능**. Phase 3: 채용공고 URL → 구조화. Phase 3.1: DART/뉴스 → 기업 프로필. 단, Phase 3.2(AI 분석)는 둘 다 필요. |
| **Phase 4 ← Phase 2.1 필수** | Phase 4(매칭) 알고리즘에서 무기 태그가 talent_fit 점수(35%)와 uniqueness 점수(25%)에 핵심 입력. Phase 2.1 미완료 시 매칭 품질 저하. |
| **Phase 0.9 시드 데이터** | Phase 0.9에서 시드되는 weapon_categories(35행), prompt_templates(4행), question_patterns(7행)은 Phase 2.1(무기 태깅)에서 사용. |

---

## Phase 진행 규칙

### 1. 시작 전

- [ ] Phase 문서 읽기 (개요, 구현 단계 확인)
- [ ] 선행 조건 Phase 완료 확인
- [ ] 이 README 상태를 🟡 진행중으로 변경

### 2. 진행 중 (TDD)

각 스텝마다 아래 순서를 반복:

- [ ] 테스트 명세 확인 (Phase 문서의 각 스텝 "테스트 명세" 섹션)
- [ ] 테스트 작성 (RED — 실패하는 테스트 먼저 작성)
- [ ] 구현 (GREEN — 테스트를 통과시키는 최소 코드)
- [ ] 리팩터링 (필요시 — 테스트 통과를 유지하면서 개선)
- [ ] 테스트 통과 확인 (`moon run backend:test` / `moon run web:test`)
- [ ] 커밋 메시지: `Phase X.Y: description`

### 3. 완료 전 검증 게이트

- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

### 4. 완료 후

- [ ] Phase 문서의 모든 체크박스 완료
- [ ] 이 README 상태를 ✅ 완료로 변경
- [ ] CLAUDE.md Phase Progress 테이블 업데이트

---

## 테스트 참고 문서

- `docs/develop/08-testing-strategy.md` — 프론트엔드 테스트 전략 (Vitest, Testing Library, MSW, Playwright)
- `docs/develop/12-backend-testing.md` — 백엔드 테스트 전략 (testify, enttest, MockAIClient, fixtures)
