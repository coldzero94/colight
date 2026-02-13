# Phase 8: 대시보드 칸반보드

> Sprint 6-7 | 예상 공수: 1.5일 | 관련 기능: F18

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | 지원 현황을 칸반보드로 시각화하고 드래그앤드롭으로 상태 관리 |
| **선행 조건** | Phase 3 (크롤링/파싱), Phase 6.1 (프리미엄) |
| **주요 산출물** | 칸반보드 UI, 상태 업데이트 API, 통계 요약 |
| **기술 스택** | @hello-pangea/dnd, React Query, Optimistic UI |

---

## 진행 상태

- [ ] 8.1 칸반 UI
- [ ] 8.2 상태 업데이트 API
- [ ] 8.3 통계 요약

---

## 구현 단계

### 8.1 칸반 UI

**목표**: 6개 컬럼의 칸반보드 UI 구성 및 드래그앤드롭 인터랙션 구현

**경로**: `src/app/(main)/dashboard/page.tsx`

**테스트 명세**:

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('KanbanBoard')` / `it('renders 6 columns')` | `src/components/dashboard/__tests__/KanbanBoard.test.tsx` | 6개 컬럼(관심~최종) 렌더링 확인 |
| `describe('KanbanBoard')` / `it('displays card count per column')` | `src/components/dashboard/__tests__/KanbanBoard.test.tsx` | 컬럼별 카드 카운트 표시 확인 |
| `describe('KanbanCard')` / `it('displays company name, position, and deadline')` | `src/components/dashboard/__tests__/KanbanCard.test.tsx` | 카드에 회사명, 직무, 마감일 표시 확인 |
| `describe('KanbanBoard')` / `it('applies responsive layout on mobile')` | `src/components/dashboard/__tests__/KanbanBoard.test.tsx` | 모바일 반응형 레이아웃(수평 스크롤) 확인 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
  - [ ] `src/components/dashboard/__tests__/KanbanBoard.test.tsx` 작성
  - [ ] `src/components/dashboard/__tests__/KanbanCard.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] @hello-pangea/dnd 설치 및 설정
  - [ ] 칸반 컬럼 정의:
    - 관심 → 작성중 → 제출완료 → 서류통과 → 면접 → 최종
  - [ ] DragDropContext + Droppable 컬럼 레이아웃
  - [ ] Draggable 지원 카드 컴포넌트 (회사명, 직무, 마감일)
  - [ ] 컬럼별 카드 카운트 표시
  - [ ] 반응형 레이아웃 (모바일: 수평 스크롤)
- [ ] 테스트 통과 확인

### 8.2 상태 업데이트 API

**목표**: 드래그앤드롭으로 지원 상태를 변경하고 Optimistic UI를 적용

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestUpdateApplicationStatus_Success` | `internal/controller/application_controller_test.go` | PUT /api/applications/:id 상태 변경 정상 응답 |
| `TestUpdateApplicationStatus_InvalidTransition` | `internal/service/application_service_test.go` | 잘못된 상태 전환 시 에러 반환 |
| `TestUpdateApplicationStatus_UpdatesTimestamp` | `internal/service/application_service_test.go` | 상태 변경 시 updated_at 자동 갱신 확인 |

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('KanbanBoard')` / `it('updates card position on drag and drop')` | `src/components/dashboard/__tests__/KanbanBoard.test.tsx` | 드래그앤드롭 시 카드 위치 업데이트 확인 |
| `describe('KanbanBoard')` / `it('rolls back on API failure')` | `src/components/dashboard/__tests__/KanbanBoard.test.tsx` | API 실패 시 Optimistic UI 롤백 확인 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/application_service_test.go` 작성
  - [ ] `internal/controller/application_controller_test.go` 작성
  - [ ] `src/components/dashboard/__tests__/KanbanBoard.test.tsx`에 드래그앤드롭 테스트 추가
- [ ] 구현 (GREEN)
  - [ ] 드래그앤드롭 시 `PUT /api/applications/[id]` 호출
  - [ ] Optimistic UI 적용 (드롭 즉시 UI 반영, 실패 시 롤백)
  - [ ] applications.status 업데이트
  - [ ] updated_at 자동 갱신
- [ ] 테스트 통과 확인

### 8.3 통계 요약

**목표**: 대시보드 상단에 지원 현황 통계 요약 표시

**테스트 명세**:

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('DashboardSummary')` / `it('displays total application count')` | `src/components/dashboard/__tests__/DashboardSummary.test.tsx` | 총 지원 건수 표시 확인 |
| `describe('DashboardSummary')` / `it('highlights deadlines within 3 days')` | `src/components/dashboard/__tests__/DashboardSummary.test.tsx` | 3일 이내 마감일 하이라이트 확인 |

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestGetApplicationStats_FilterByStatus` | `internal/service/application_service_test.go` | 상태별 필터링 및 건수 집계 확인 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
  - [ ] `src/components/dashboard/__tests__/DashboardSummary.test.tsx` 작성
  - [ ] `internal/service/application_service_test.go`에 통계 테스트 추가
- [ ] 구현 (GREEN)
  - [ ] 총 지원 건수
  - [ ] 상태별 건수 (컬럼 헤더에 표시)
  - [ ] 다가오는 마감일 목록 (3일 이내 하이라이트)
  - [ ] 대시보드 상단 요약 카드
- [ ] 테스트 통과 확인

---

## Phase 완료 체크리스트

**기능 검증**:
- [ ] 드래그앤드롭으로 상태 변경 정상 동작
- [ ] Optimistic UI 적용 (실패 시 롤백)
- [ ] 6개 컬럼 칸반 렌더링
- [ ] 통계 요약 표시
- [ ] 모바일 레이아웃 대응

**테스트**:
- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

**완료 처리**:
- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

→ [Phase 8.1: 버전 관리](./phase-8.1-version-management.md)
