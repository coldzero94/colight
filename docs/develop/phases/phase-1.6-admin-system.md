# Phase 1.6: 어드민 시스템 확장

> Sprint 8-9 | 예상 공수: 5일 | 관련 기능: 시스템 관리, 사용자 관리, 모니터링

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | 역할 계층, API 키/모델 설정, 사용량 모니터링, 시스템 헬스, 감사 로그, 계정 관리 등 프로덕션 수준의 어드민 시스템 구현 |
| **선행 조건** | Phase 1.5 (기존 어드민), Phase 6.1 (Fair Use Policy) |
| **주요 산출물** | 역할 계층 미들웨어, 시스템 설정 UI, 사용량 대시보드, 감사 로그, 헬스 체크 |
| **기획 문서** | [docs/plan/13-admin-system.md](../../plan/13-admin-system.md) |

---

## 진행 상태

| Step | 이름 | 상태 |
|------|------|------|
| 1.6.1 | 역할 계층 (Role Hierarchy) | ✅ 완료 |
| 1.6.2 | 시스템 설정 (API 키 & 모델) | ✅ 완료 |
| 1.6.3 | 사용량 모니터링 | ✅ 완료 |
| 1.6.4 | 사용자 관리 확장 (정지/로그아웃/상세) | ✅ 완료 |
| 1.6.5 | 감사 로그 & 한도 설정 | ✅ 완료 |
| 1.6.6 | 시스템 헬스 & 작업 모니터링 | ⬜ 대기 |
| 1.6.7 | 피드백 관리 & 크롤러 모니터링 | ⬜ 대기 |
| 1.6.8 | 데이터 프라이버시 (내보내기/삭제) | ⬜ 대기 |

---

## 구현 단계

### 1.6.1 역할 계층 (Role Hierarchy)

**목표**: 기존 2단계 역할(`user` | `admin`)을 4단계(`user` | `manager` | `admin` | `super_admin`)로 확장

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestRoleLevel` | `internal/infrastructure/middleware/role_test.go` | 역할별 레벨 값 검증 (user=1, manager=2, admin=3, super_admin=4) |
| `TestRequireRole_ManagerAccess` | `internal/infrastructure/middleware/role_test.go` | manager 이상만 접근 가능 |
| `TestRequireRole_AdminAccess` | `internal/infrastructure/middleware/role_test.go` | admin 이상만 접근 가능 |
| `TestRequireRole_SuperAdminAccess` | `internal/infrastructure/middleware/role_test.go` | super_admin만 접근 가능 |
| `TestRequireRole_InsufficientRole` | `internal/infrastructure/middleware/role_test.go` | 권한 부족 시 403 반환 |
| `TestUpdateUserRole_AdminCanPromoteToManager` | `internal/controller/admin_controller_test.go` | admin이 user→manager 변경 가능 |
| `TestUpdateUserRole_AdminCannotPromoteToAdmin` | `internal/controller/admin_controller_test.go` | admin이 admin 이상 승격 불가 |
| `TestUpdateUserRole_SuperAdminCanPromoteToAdmin` | `internal/controller/admin_controller_test.go` | super_admin은 admin 승격 가능 |
| `TestUpdateUserRole_CannotChangeSelf` | `internal/controller/admin_controller_test.go` | 자기 자신 역할 변경 불가 |
| `TestUpdateUserRole_LastSuperAdminProtected` | `internal/controller/admin_controller_test.go` | 마지막 super_admin 강등 불가 |

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `hasRole returns true for equal or higher role` | `src/lib/__tests__/auth-utils.test.ts` | 권한 비교 유틸리티 검증 |
| `isManager/isAdmin/isSuperAdmin helpers` | `src/lib/__tests__/auth-utils.test.ts` | 역할 확인 헬퍼 검증 |
| `AdminGuard redirects insufficient role` | `src/components/auth/__tests__/admin-guard.test.tsx` | requiredRole별 접근 제어 검증 |
| `AdminSidebar filters menu by role` | `src/components/admin/__tests__/admin-sidebar.test.tsx` | 역할별 메뉴 필터링 검증 |
| `Users page shows role dropdown` | `src/app/(admin)/admin/users/__tests__/page.test.tsx` | 역할 변경 드롭다운 UI 검증 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
  - [ ] `internal/infrastructure/middleware/role_test.go`
  - [ ] `internal/controller/admin_controller_test.go` 역할 검증 테스트 추가
  - [ ] `src/lib/__tests__/auth-utils.test.ts`
  - [ ] `src/components/auth/__tests__/admin-guard.test.tsx` 확장
  - [ ] `src/components/admin/__tests__/admin-sidebar.test.tsx` 확장
- [ ] 구현 (GREEN)
  - [ ] **DB**: Ent 스키마 `role` enum 확장 (`manager`, `super_admin` 추가)
  - [ ] **DB**: `moon run backend:generate-ent && moon run backend:migrate-diff -- name=add_role_hierarchy`
  - [ ] **Backend**: `middleware/role.go` — `RoleLevel()`, `RequireRole(minRole)` 미들웨어
  - [ ] **Backend**: `admin_controller.go` — `UpdateUserRole` 역할 검증 로직 (승격 권한, 자기 자신, 마지막 super_admin)
  - [ ] **Backend**: `cmd/api/main.go` — 기존 `AdminMiddleware()` → 라우트별 `RequireRole()` 교체
  - [ ] **Backend**: `scripts/seed_admin.go` — super_admin으로 변경
  - [ ] **Frontend**: `stores/auth-store.ts` — role 타입 확장
  - [ ] **Frontend**: `lib/auth-utils.ts` — `hasRole()`, `isManager()`, `isAdmin()`, `isSuperAdmin()`
  - [ ] **Frontend**: `components/auth/admin-guard.tsx` — `requiredRole` prop 추가
  - [ ] **Frontend**: `components/admin/admin-sidebar.tsx` — 메뉴 역할별 필터링
  - [ ] **Frontend**: `app/(admin)/admin/users/page.tsx` — 역할 변경 드롭다운, 배지 색상
- [ ] 테스트 통과 확인

**산출물**:
- `internal/infrastructure/middleware/role.go`
- `internal/infrastructure/middleware/role_test.go`
- `src/lib/auth-utils.ts`
- 수정: Ent 스키마, admin_controller, auth-store, admin-guard, admin-sidebar, users page

---

### 1.6.2 시스템 설정 (API 키 & 모델)

**목표**: 어드민 UI에서 API 키 변경, AI 모델 선택, 사용량 한도 설정 (super_admin 전용)

**DB 스키마**:

```sql
CREATE TABLE system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,        -- 암호화 저장 (is_secret=true인 경우)
    description VARCHAR(500),
    category VARCHAR(50) NOT NULL,     -- "api_key" | "model" | "limit" | "cost"
    is_secret BOOLEAN DEFAULT false,
    updated_by UUID REFERENCES user_profiles(id),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);
```

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestListConfigs_ByCategory` | `internal/controller/admin_controller_test.go` | 카테고리별 설정 조회, secret 마스킹 |
| `TestUpdateConfig_ApiKey` | `internal/controller/admin_controller_test.go` | API 키 변경 + 암호화 저장 |
| `TestUpdateConfig_NonSuperAdminForbidden` | `internal/controller/admin_controller_test.go` | super_admin 외 403 |
| `TestValidateApiKey_Valid` | `internal/controller/admin_controller_test.go` | 키 유효성 검증 성공 |
| `TestValidateApiKey_Invalid` | `internal/controller/admin_controller_test.go` | 잘못된 키 에러 반환 |
| `TestConfigCrypto_EncryptDecrypt` | `internal/infrastructure/crypto/config_crypto_test.go` | AES-256-GCM 암복호화 |
| `TestConfigCrypto_Mask` | `internal/infrastructure/crypto/config_crypto_test.go` | 키 마스킹 (앞3자 + ••• + 뒤3자) |

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `Settings page renders tabs` | `src/app/(admin)/admin/settings/__tests__/page.test.tsx` | API Keys / AI Models / Limits 탭 렌더링 |
| `API key shows masked value` | `src/app/(admin)/admin/settings/__tests__/page.test.tsx` | 키 마스킹 표시 검증 |
| `Model select changes provider` | `src/app/(admin)/admin/settings/__tests__/page.test.tsx` | 모델 선택 변경 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `ent/schema/systemconfig.go` Ent 스키마
  - [ ] **DB**: 마이그레이션 생성 + 적용
  - [ ] **Backend**: `internal/infrastructure/crypto/config_crypto.go` (AES-256-GCM)
  - [ ] **Backend**: `internal/service/config_service.go` (CRUD + 캐시 + 폴백)
  - [ ] **Backend**: `admin_controller.go` — `ListConfigs`, `UpdateConfig`, `ValidateApiKey`
  - [ ] **Backend**: AI provider가 DB 설정 우선 조회하도록 수정
  - [ ] **Frontend**: `/admin/settings` 페이지 (3개 탭)
  - [ ] **Frontend**: API 키 변경 모달 (검증 → 저장)
  - [ ] **Frontend**: AI 모델 설정 폼
  - [ ] **Frontend**: 사용량 한도 설정 테이블
  - [ ] **Frontend**: 사이드바에 "시스템 설정" 메뉴 추가 (`super_admin`)
- [ ] 테스트 통과 확인

**산출물**:
- `ent/schema/systemconfig.go`
- `internal/infrastructure/crypto/config_crypto.go`
- `internal/service/config_service.go`
- `src/app/(admin)/admin/settings/page.tsx`

---

### 1.6.3 사용량 모니터링

**목표**: AI API 사용량, 토큰 소모량, 추정 비용을 대시보드에서 확인 (manager 이상)

**DB 스키마**:

```sql
-- usage_logs 테이블 확장 (기존 테이블에 필드 추가)
ALTER TABLE usage_logs ADD COLUMN provider VARCHAR(20);
ALTER TABLE usage_logs ADD COLUMN model VARCHAR(100);
ALTER TABLE usage_logs ADD COLUMN input_tokens INTEGER DEFAULT 0;
ALTER TABLE usage_logs ADD COLUMN output_tokens INTEGER DEFAULT 0;
ALTER TABLE usage_logs ADD COLUMN total_tokens INTEGER DEFAULT 0;
ALTER TABLE usage_logs ADD COLUMN estimated_cost_krw DECIMAL(10,2);
ALTER TABLE usage_logs ADD COLUMN latency_ms INTEGER;
ALTER TABLE usage_logs ADD COLUMN status VARCHAR(20) DEFAULT 'success';
ALTER TABLE usage_logs ADD COLUMN error_message TEXT;
```

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestGetUsageSummary` | `internal/controller/admin_controller_test.go` | 기간별 요약 (호출 수, 토큰, 비용, 에러율) |
| `TestGetUsageDaily` | `internal/controller/admin_controller_test.go` | 일별 기능별 사용량 |
| `TestGetUsageCosts` | `internal/controller/admin_controller_test.go` | 프로바이더별 비용 비율 |
| `TestGetUsageTopUsers` | `internal/controller/admin_controller_test.go` | 상위 사용자 토큰/비용 순위 |

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `Usage page renders summary cards` | `src/app/(admin)/admin/usage/__tests__/page.test.tsx` | 요약 카드 렌더링 |
| `Usage page renders daily chart` | `src/app/(admin)/admin/usage/__tests__/page.test.tsx` | 일별 차트 표시 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `usage_logs` Ent 스키마 확장 + 마이그레이션
  - [ ] **Backend**: AI 호출 시 usage_logs에 토큰/비용 자동 기록
  - [ ] **Backend**: `admin_controller.go` — 사용량 집계 API 4개
  - [ ] **Frontend**: `/admin/usage` 페이지
  - [ ] **Frontend**: 요약 카드 (총 호출, 토큰, 비용, 에러율)
  - [ ] **Frontend**: 일별 사용량 바 차트 (recharts)
  - [ ] **Frontend**: 프로바이더별 비용 비율 바
  - [ ] **Frontend**: 상위 사용자 테이블
- [ ] 테스트 통과 확인

**산출물**:
- `src/app/(admin)/admin/usage/page.tsx`
- 수정: `usage_logs` Ent 스키마, AI 호출 미들웨어, admin_controller

---

### 1.6.4 사용자 관리 확장

**목표**: 계정 정지/해제, 강제 로그아웃, 사용자 상세 모달, 플랜 수동 변경

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestSuspendUser_Success` | `internal/controller/admin_controller_test.go` | 계정 정지 성공 |
| `TestSuspendUser_AlreadySuspended` | `internal/controller/admin_controller_test.go` | 이미 정지된 계정 |
| `TestUnsuspendUser` | `internal/controller/admin_controller_test.go` | 정지 해제 |
| `TestSuspendedUser_AuthBlocked` | `internal/infrastructure/middleware/auth_test.go` | 정지된 사용자 API 접근 차단 |
| `TestForceLogout` | `internal/controller/admin_controller_test.go` | 강제 로그아웃 (토큰 무효화) |
| `TestGetUserDetail` | `internal/controller/admin_controller_test.go` | 사용자 상세 (경험 수, 코칭 수, 사용량) |
| `TestUpdateUserPlan` | `internal/controller/admin_controller_test.go` | 플랜 수동 변경 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `user_profiles`에 `suspended`, `suspended_at`, `suspended_reason` 필드 추가
  - [ ] **Backend**: `POST /v1/admin/users/:id/suspend` — 계정 정지
  - [ ] **Backend**: `DELETE /v1/admin/users/:id/suspend` — 정지 해제
  - [ ] **Backend**: `POST /v1/admin/users/:id/force-logout` — 강제 로그아웃
  - [ ] **Backend**: `GET /v1/admin/users/:id/detail` — 상세 정보 (경험 수, 코칭 수 등)
  - [ ] **Backend**: `PUT /v1/admin/users/:id/plan` — 플랜 변경
  - [ ] **Backend**: AuthMiddleware에 suspended 체크 추가
  - [ ] **Frontend**: 사용자 상세 모달 컴포넌트
  - [ ] **Frontend**: 정지/해제 버튼 + 사유 입력
  - [ ] **Frontend**: 플랜 변경 드롭다운
- [ ] 테스트 통과 확인

**산출물**:
- 수정: Ent 스키마, admin_controller, auth middleware, users page

---

### 1.6.5 감사 로그 & 한도 설정

**목표**: 모든 어드민 설정 변경 사항을 기록하고 조회, Fair Use 한도를 어드민에서 조정

**DB 스키마**:

```sql
CREATE TABLE admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES user_profiles(id),
    action VARCHAR(50) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id VARCHAR(255),
    old_value TEXT,
    new_value TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT NOW()
);
```

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestListAuditLogs` | `internal/controller/admin_controller_test.go` | 감사 로그 목록 (필터, 페이지네이션) |
| `TestAuditLog_RoleChangeRecorded` | `internal/controller/admin_controller_test.go` | 역할 변경 시 감사 로그 자동 기록 |
| `TestAuditLog_ConfigChangeRecorded` | `internal/controller/admin_controller_test.go` | 설정 변경 시 감사 로그 자동 기록 |

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `Audit logs page renders table` | `src/app/(admin)/admin/logs/__tests__/page.test.tsx` | 감사 로그 테이블 렌더링 |
| `Audit logs page filters by action` | `src/app/(admin)/admin/logs/__tests__/page.test.tsx` | 액션 필터 동작 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `ent/schema/adminauditlog.go` Ent 스키마
  - [ ] **Backend**: 감사 로그 자동 기록 헬퍼 (`recordAuditLog()`)
  - [ ] **Backend**: 기존 역할 변경, 설정 변경, 프롬프트 변경에 감사 로그 추가
  - [ ] **Backend**: `GET /v1/admin/audit-logs` — 감사 로그 목록 API
  - [ ] **Frontend**: `/admin/logs` 페이지
  - [ ] **Frontend**: 액션/관리자 필터, 페이지네이션
- [ ] 테스트 통과 확인

**산출물**:
- `ent/schema/adminauditlog.go`
- `src/app/(admin)/admin/logs/page.tsx`

---

### 1.6.6 시스템 헬스 & 작업 모니터링

**목표**: DB, AI 프로바이더, River 작업 큐의 건강 상태를 대시보드에 표시

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestHealthCheck_AllHealthy` | `internal/controller/admin_controller_test.go` | 모든 서비스 정상 |
| `TestHealthCheck_DegradedProvider` | `internal/controller/admin_controller_test.go` | 일부 프로바이더 느린 경우 degraded |
| `TestJobsSummary` | `internal/controller/admin_controller_test.go` | 작업 큐 상태 요약 |
| `TestJobRetry` | `internal/controller/admin_controller_test.go` | 실패 작업 재시도 |

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `Dashboard shows health status cards` | `src/app/(admin)/admin/__tests__/page.test.tsx` | 시스템 상태 카드 렌더링 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **Backend**: `GET /v1/admin/health` — DB, AI 프로바이더, River 상태 체크
  - [ ] **Backend**: `GET /v1/admin/jobs/summary` — 작업 큐 상태 요약
  - [ ] **Backend**: `GET /v1/admin/jobs/failed` — 실패 작업 목록
  - [ ] **Backend**: `POST /v1/admin/jobs/:id/retry` — 작업 재시도
  - [ ] **Backend**: `DELETE /v1/admin/jobs/:id` — 작업 삭제
  - [ ] **Frontend**: 대시보드에 시스템 상태 카드 추가
  - [ ] **Frontend**: 작업 큐 모니터링 섹션 (실패 작업 재시도/삭제)
- [ ] 테스트 통과 확인

**산출물**:
- 수정: admin_controller, admin dashboard page

---

### 1.6.7 피드백 관리 & 크롤러 모니터링

**목표**: 사용자 피드백을 어드민에서 조회/관리, 크롤러 건강 상태 모니터링

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestListFeedbacks` | `internal/controller/admin_controller_test.go` | 피드백 목록 (상태/유형 필터) |
| `TestUpdateFeedbackStatus` | `internal/controller/admin_controller_test.go` | 피드백 상태 변경 + 메모 |
| `TestCrawlerStatus` | `internal/controller/admin_controller_test.go` | 크롤러 사이트별 상태 |

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `Feedback page renders list` | `src/app/(admin)/admin/feedbacks/__tests__/page.test.tsx` | 피드백 목록 렌더링 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `feedbacks` 테이블에 `admin_status`, `admin_note`, `reviewed_by`, `reviewed_at` 필드 추가
  - [ ] **Backend**: `GET /v1/admin/feedbacks` — 피드백 목록 API
  - [ ] **Backend**: `PUT /v1/admin/feedbacks/:id` — 피드백 상태/메모 변경
  - [ ] **Backend**: `GET /v1/admin/crawler/status` — 크롤러 상태
  - [ ] **Backend**: `GET /v1/admin/crawler/failures` — 크롤러 실패 로그
  - [ ] **Frontend**: `/admin/feedbacks` 페이지
  - [ ] **Frontend**: 피드백 상태 변경 + 관리자 메모
  - [ ] **Frontend**: 크롤러 상태 (대시보드 또는 별도 페이지)
- [ ] 테스트 통과 확인

**산출물**:
- `src/app/(admin)/admin/feedbacks/page.tsx`
- 수정: feedbacks Ent 스키마, admin_controller

---

### 1.6.8 데이터 프라이버시 (내보내기/삭제)

**목표**: PIPA 준수를 위한 사용자 데이터 내보내기, 계정 삭제 처리

**테스트 명세**:

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestExportUserData` | `internal/controller/admin_controller_test.go` | 사용자 데이터 내보내기 (JSON/CSV) |
| `TestDeleteRequest_Create` | `internal/controller/admin_controller_test.go` | 삭제 요청 생성 (30일 보존) |
| `TestDeleteRequest_Cancel` | `internal/controller/admin_controller_test.go` | 삭제 요청 취소 |
| `TestDeletionQueue_List` | `internal/controller/admin_controller_test.go` | 삭제 대기 목록 조회 |

**구현 체크리스트**:
- [ ] 테스트 작성 (RED)
- [ ] 구현 (GREEN)
  - [ ] **DB**: `deletion_requests` 테이블 (user_id, reason, scheduled_at, status)
  - [ ] **Backend**: `POST /v1/admin/users/:id/export` — 데이터 내보내기 (비동기, River job)
  - [ ] **Backend**: `POST /v1/admin/users/:id/delete-request` — 삭제 요청
  - [ ] **Backend**: `DELETE /v1/admin/users/:id/delete-request` — 삭제 취소
  - [ ] **Backend**: `GET /v1/admin/deletion-queue` — 삭제 대기 목록
  - [ ] **Backend**: River 주기 작업 — 보존 기간 만료 계정 영구 삭제
  - [ ] **Frontend**: 사용자 상세 모달에 내보내기/삭제 버튼 추가
  - [ ] **Frontend**: 삭제 대기 목록 (대시보드 또는 별도)
- [ ] 테스트 통과 확인

**산출물**:
- `ent/schema/deletionrequest.go`
- 수정: admin_controller, users page

---

## Phase 완료 체크리스트

- [ ] 역할 계층 동작 (user < manager < admin < super_admin)
- [ ] manager가 대시보드/사용자목록/모니터링/감사로그 읽기 가능
- [ ] admin이 프롬프트/한도/계정정지/역할(→manager) 변경 가능
- [ ] super_admin이 API 키/모델 설정/admin 승격/모든 권한 가능
- [ ] 자기 자신 역할 변경 불가, 마지막 super_admin 강등 불가
- [ ] API 키 암호화 저장 + 마스킹 표시
- [ ] 사용량/비용 모니터링 대시보드 표시
- [ ] 감사 로그 모든 어드민 액션 기록
- [ ] 시스템 헬스 체크 동작
- [ ] 계정 정지 시 API 접근 차단
- [ ] 피드백 관리 동작
- [ ] 데이터 내보내기/삭제 처리 동작
- [ ] `moon run backend:lint` → 경고 0건
- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:lint` → 경고 0건
- [ ] `moon run web:typecheck` → 에러 0건
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run web:build` → 빌드 성공

---

## 다음 Phase

→ [Phase 7.1: 경험 추천](./phase-7.1-experience-recommend.md)
