# Phase 7.1: 경험 추천 강화

> Sprint 6 | 예상 공수: 1일 | 관련 기능: F12

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | 문항별 경험 추천 로직 고도화 및 UI 개선 |
| **선행 조건** | Phase 5 (문항 분석), Phase 2.1 (무기 태깅) |
| **주요 산출물** | 강화된 추천 알고리즘, 추천 사유 표시 UI |
| **기술 스택** | pgvector, React Query |

---

## 진행 상태

- [x] 7.1.1 추천 로직 강화
- [x] 7.1.2 추천 UI 개선

---

## 구현 단계

### 7.1.1 추천 로직 강화

#### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `TestRecommendService_WeaponMatchScore` | `internal/service/recommend_service_test.go` | 문항 필요 무기와 경험 보유 무기 매칭 점수 계산 정확성 |
| `TestRecommendService_KeywordOverlap` | `internal/service/recommend_service_test.go` | 문항 키워드와 경험 키워드 오버랩 점수 계산 |
| `TestRecommendService_UsageDedup` | `internal/service/recommend_service_test.go` | 이미 다른 문항에 사용된 경험 감점 처리, 유니크 보너스 적용 |
| `TestRecommendService_RankTopN` | `internal/service/recommend_service_test.go` | 최종 가중 합산 점수 기준 상위 3~5개 정렬 반환 |

#### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/recommend_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] 무기 매칭 점수 (문항 필요 무기 ↔ 경험 보유 무기)
  - [ ] 키워드 오버랩 점수 (문항 키워드 ↔ 경험 키워드)
  - [ ] 사용 이력 중복 제거 (이미 다른 문항에 사용된 경험 감점)
  - [ ] 유니크 보너스 (덜 사용된 경험 우선)
  - [ ] 최종 점수 = 가중 합산, 상위 3~5개 반환
- [ ] 테스트 통과 확인

### 7.1.2 추천 UI 개선

#### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('RecommendationList')` | `src/components/recommend/__tests__/recommendation-list.test.tsx` | 추천 경험 목록 렌더링, 점수순 정렬 확인 |
| `describe('RelevanceIndicator')` | `src/components/recommend/__tests__/relevance-indicator.test.tsx` | 적합도 점수 시각적 표시 (프로그레스 바, 색상) |
| `describe('RecommendationCard')` | `src/components/recommend/__tests__/recommendation-card.test.tsx` | 추천 사유 텍스트 표시, "이미 사용됨" 뱃지, 카드 확장 미리보기 |

#### 구현 체크리스트

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/recommend/__tests__/recommendation-list.test.tsx` 작성
  - [ ] `src/components/recommend/__tests__/relevance-indicator.test.tsx` 작성
  - [ ] `src/components/recommend/__tests__/recommendation-card.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 추천 사유 텍스트 표시 (예: "리더십 무기 일치", "직무 키워드 3개 매칭")
  - [ ] "이미 사용됨" 뱃지 표시 (다른 문항에서 사용 중인 경험)
  - [ ] 추천 순서 재정렬 (점수순 + 사용 여부 고려)
  - [ ] 경험 카드 미리보기 (hover 또는 expand)
- [ ] 테스트 통과 확인

---

## 완료 체크리스트

- [x] 추천 결과에 무기 매칭 + 키워드 오버랩 반영
- [x] 이미 사용된 경험 구분 표시
- [x] 추천 사유 사용자에게 표시
- [x] 기존 추천 대비 적합도 향상 확인
- [x] phases/README.md 상태 업데이트
- [x] `moon run backend:test` → 전체 통과
- [x] `moon run web:test` → 전체 통과
- [x] `moon run :lint` → 경고 0건
- [x] `moon run web:build` → 빌드 성공

---

## 다음 Phase

→ [Phase 8: 대시보드 칸반](./phase-8-dashboard-kanban.md)
