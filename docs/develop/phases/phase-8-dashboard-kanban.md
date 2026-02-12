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

**경로**: `src/app/(main)/dashboard/page.tsx`

- [ ] @hello-pangea/dnd 설치 및 설정
- [ ] 칸반 컬럼 정의:
  - 관심 → 작성중 → 제출완료 → 서류통과 → 면접 → 최종
- [ ] DragDropContext + Droppable 컬럼 레이아웃
- [ ] Draggable 지원 카드 컴포넌트 (회사명, 직무, 마감일)
- [ ] 컬럼별 카드 카운트 표시
- [ ] 반응형 레이아웃 (모바일: 수평 스크롤)

### 8.2 상태 업데이트 API

- [ ] 드래그앤드롭 시 `PUT /api/applications/[id]` 호출
- [ ] Optimistic UI 적용 (드롭 즉시 UI 반영, 실패 시 롤백)
- [ ] applications.status 업데이트
- [ ] updated_at 자동 갱신

### 8.3 통계 요약

- [ ] 총 지원 건수
- [ ] 상태별 건수 (컬럼 헤더에 표시)
- [ ] 다가오는 마감일 목록 (3일 이내 하이라이트)
- [ ] 대시보드 상단 요약 카드

---

## 완료 체크리스트

- [ ] 드래그앤드롭으로 상태 변경 정상 동작
- [ ] Optimistic UI 적용 (실패 시 롤백)
- [ ] 6개 컬럼 칸반 렌더링
- [ ] 통계 요약 표시
- [ ] 모바일 레이아웃 대응
- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

→ [Phase 8.1: 버전 관리](./phase-8.1-version-management.md)
