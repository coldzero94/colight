# Phase 8.1: 자소서 버전 관리

> Sprint 7 | 예상 공수: 1일 | 관련 기능: F20

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | 자소서 버전 히스토리 관리, 버전 비교, 복원 기능 |
| **선행 조건** | Phase 5.1 (초안 코칭), Phase 5.2 (코칭 에디터) |
| **주요 산출물** | 버전 목록 UI, 사이드바이사이드 비교, 버전 복원 |
| **기술 스택** | cover_letter_versions 테이블, diff 알고리즘 |

---

## 진행 상태

- [ ] 8.1.1 버전 히스토리 UI
- [ ] 8.1.2 버전 비교
- [ ] 8.1.3 버전 복원

---

## 구현 단계

### 8.1.1 버전 히스토리 UI

**목표**: 코칭 에디터 사이드바에 버전 목록을 표시하고 버전 간 탐색 지원

**테스트 명세**:

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('VersionList')` / `it('renders version entries with number, score, and date')` | `src/components/version/__tests__/VersionList.test.tsx` | 버전 번호, 점수, 날짜 표시 확인 |
| `describe('VersionList')` / `it('highlights the current version')` | `src/components/version/__tests__/VersionList.test.tsx` | 현재 버전 하이라이트 확인 |
| `describe('VersionList')` / `it('loads version content on click')` | `src/components/version/__tests__/VersionList.test.tsx` | 버전 클릭 시 해당 내용 로드 확인 |

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestCreateVersion_Success` | `internal/service/version_service_test.go` | 새 버전 생성 및 version_number 증가 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/version/__tests__/VersionList.test.tsx` 작성
  - [ ] `internal/service/version_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] 버전 목록 패널 (코칭 에디터 사이드바)
  - [ ] 각 버전 항목: 버전 번호, 점수, 날짜, 내용 미리보기
  - [ ] 현재 버전 하이라이트
  - [ ] 버전 클릭 시 해당 버전 내용 로드
- [ ] 테스트 통과 확인

### 8.1.2 버전 비교

**목표**: 두 버전을 사이드바이사이드로 비교하여 변경 사항을 시각적으로 표시

**테스트 명세**:

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('DiffViewer')` / `it('renders side-by-side comparison of two versions')` | `src/components/version/__tests__/DiffViewer.test.tsx` | 두 버전 사이드바이사이드 비교 뷰 렌더링 확인 |
| `describe('DiffViewer')` / `it('highlights added and deleted content')` | `src/components/version/__tests__/DiffViewer.test.tsx` | 추가(초록)/삭제(빨강) 하이라이트 확인 |
| `describe('DiffViewer')` / `it('displays change statistics')` | `src/components/version/__tests__/DiffViewer.test.tsx` | 변경 통계(추가/삭제/수정 글자수) 표시 확인 |

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestCompareVersions_DiffResult` | `internal/service/version_service_test.go` | 두 버전 비교 시 diff 결과 정확성 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/version/__tests__/DiffViewer.test.tsx` 작성
  - [ ] `internal/service/version_service_test.go`에 비교 테스트 추가
- [ ] 구현 (GREEN)
  - [ ] 두 버전 선택 → 사이드바이사이드 비교 뷰
  - [ ] 추가된 내용 하이라이트 (초록)
  - [ ] 삭제된 내용 하이라이트 (빨강)
  - [ ] 변경 통계 (추가/삭제/수정 글자수)
- [ ] 테스트 통과 확인

### 8.1.3 버전 복원

**목표**: 이전 버전을 새 버전으로 복원하여 기존 히스토리를 보존

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestRollbackVersion_CreatesNewVersion` | `internal/service/version_service_test.go` | 복원 시 새 버전 생성 확인 (덮어쓰기 아님) |
| `TestRollbackVersion_IncrementsVersionNumber` | `internal/service/version_service_test.go` | 복원 시 version_number 자동 증가 확인 |

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('RollbackConfirmDialog')` / `it('displays rollback confirmation dialog')` | `src/components/version/__tests__/RollbackConfirmDialog.test.tsx` | 복원 확인 다이얼로그 표시 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/version_service_test.go`에 복원 테스트 추가
  - [ ] `src/components/version/__tests__/RollbackConfirmDialog.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 이전 버전 복원 버튼
  - [ ] 복원 시 새 버전으로 생성 (기존 버전 덮어쓰기 아님)
  - [ ] 복원 확인 다이얼로그
  - [ ] version_number 자동 증가
- [ ] 테스트 통과 확인

---

## Phase 완료 체크리스트

**기능 검증**:

- [ ] 버전 히스토리 목록 정상 표시
- [ ] 두 버전 간 비교 뷰 동작
- [ ] 추가/삭제 하이라이트 표시
- [ ] 복원 시 새 버전 생성 확인

**테스트**:

- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

**완료 처리**:

- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

→ [Phase 9: 결제 연동](./phase-9-payment.md)
