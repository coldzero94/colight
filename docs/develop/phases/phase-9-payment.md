# Phase 9: 결제 연동

> Sprint 7 | 예상 공수: 2일

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | Toss Payments 연동으로 유료 상품 결제 및 크레딧 시스템 구현 |
| **선행 조건** | Phase 6.1 (프리미엄 & 마무리) |
| **주요 산출물** | 상품 페이지, 결제 플로우, 크레딧 관리 |
| **기술 스택** | @tosspayments/payment-sdk, Webhook, user_profiles 크레딧 |

---

## 진행 상태

- [ ] 9.1 결제 상품 설계
- [ ] 9.2 Toss Payments SDK 연동
- [ ] 9.3 결제 확인 + 크레딧
- [ ] 9.4 결제 UI

---

## 구현 단계

### 9.1 결제 상품 설계

**목표**: 결제 상품 구성을 정의하고 상품 정보를 관리

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestGetPlan_ReturnsCorrectCredits` | `internal/service/payment_service_test.go` | 각 플랜(Starter/Pro/Season Pass)별 크레딧 수 정확성 확인 |
| `TestUpgradePlan_Success` | `internal/service/payment_service_test.go` | 플랜 업그레이드 시 크레딧 차액 정산 확인 |
| `TestDowngradePlan_RefundCalculation` | `internal/service/payment_service_test.go` | 플랜 다운그레이드 시 환불 금액 계산 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/payment_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] 상품 구성:
    | 상품 | 가격 | 내용 |
    |------|------|------|
    | **Starter** | TBD | 코칭 10건 |
    | **Pro** | TBD | 코칭 30건 |
    | **Season Pass** | TBD | 무제한 (기간제) |
  - [ ] 상품 정보 DB 또는 상수 관리
- [ ] 테스트 통과 확인

### 9.2 Toss Payments SDK 연동

**목표**: Toss Payments SDK를 통한 결제 요청 및 콜백 처리

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestInitiatePayment_CreatesPaymentRecord` | `internal/service/payment_service_test.go` | 결제 시작 시 결제 레코드 생성 확인 |
| `TestConfirmPayment_Success` | `internal/controller/payment_controller_test.go` | 결제 확인 API 정상 응답 확인 |

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('PaymentForm')` / `it('validates required fields before submission')` | `src/components/payment/__tests__/PaymentForm.test.tsx` | 결제 양식 필수 필드 검증 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/payment_service_test.go`에 결제 시작 테스트 추가
  - [ ] `internal/controller/payment_controller_test.go` 작성
  - [ ] `src/components/payment/__tests__/PaymentForm.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] @tosspayments/payment-sdk 설치
  - [ ] 프론트엔드 결제 위젯 초기화
  - [ ] 결제 요청 → Toss 결제 페이지 → 콜백 처리
  - [ ] `/api/payments/confirm` 웹훅 엔드포인트 생성
  - [ ] 환경변수: TOSS_CLIENT_KEY, TOSS_SECRET_KEY
- [ ] 테스트 통과 확인

### 9.3 결제 확인 + 크레딧

**목표**: Toss 웹훅을 검증하고 결제 성공 시 크레딧을 추가

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestVerifyWebhookSignature_ValidSignature` | `internal/service/payment_service_test.go` | 유효한 웹훅 서명 검증 통과 확인 |
| `TestVerifyWebhookSignature_InvalidSignature` | `internal/service/payment_service_test.go` | 잘못된 웹훅 서명 시 거부 확인 |
| `TestProcessPaymentSuccess_AddsCredits` | `internal/service/payment_service_test.go` | 결제 성공 시 user_profiles 크레딧 추가 확인 |
| `TestProcessPaymentFailure_NoCreditsAdded` | `internal/service/payment_service_test.go` | 결제 실패 시 크레딧 미추가 확인 |
| `TestValidateReceipt_MatchesPaymentRecord` | `internal/service/payment_service_test.go` | 영수증 검증 시 결제 기록 일치 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/payment_service_test.go`에 웹훅 서명 검증 테스트 추가
  - [ ] `internal/service/payment_service_test.go`에 크레딧 처리 테스트 추가
- [ ] 구현 (GREEN)
  - [ ] Toss 웹훅 검증 (서명 확인)
  - [ ] 결제 성공 시 user_profiles에 크레딧 추가
  - [ ] 결제 실패/취소 처리
  - [ ] 결제 이력 로깅
- [ ] 테스트 통과 확인

### 9.4 결제 UI

**목표**: 가격표 페이지와 결제 결과 화면 구현

**경로**: `src/app/(main)/pricing/page.tsx`

**테스트 명세**:

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('PlanSelection')` / `it('renders 3 product cards with correct details')` | `src/components/payment/__tests__/PlanSelection.test.tsx` | 3개 상품 카드 렌더링 확인 |
| `describe('PlanSelection')` / `it('highlights current plan')` | `src/components/payment/__tests__/PlanSelection.test.tsx` | 현재 플랜 하이라이트 확인 |
| `describe('BillingHistory')` / `it('displays payment history list')` | `src/components/payment/__tests__/BillingHistory.test.tsx` | 결제 이력 목록 표시 확인 |
| `describe('UpgradeFlow')` / `it('navigates through upgrade steps')` | `src/components/payment/__tests__/UpgradeFlow.test.tsx` | 업그레이드 플로우 단계 탐색 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/payment/__tests__/PlanSelection.test.tsx` 작성
  - [ ] `src/components/payment/__tests__/BillingHistory.test.tsx` 작성
  - [ ] `src/components/payment/__tests__/UpgradeFlow.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 가격표 페이지 (상품 카드 3개)
  - [ ] 현재 플랜 / 잔여 크레딧 표시
  - [ ] 결제 버튼 → Toss 결제 플로우
  - [ ] 결제 완료/실패 결과 페이지
- [ ] 테스트 통과 확인

---

## Phase 완료 체크리스트

**기능 검증**:

- [ ] Toss Payments 테스트 결제 성공
- [ ] 웹훅 검증 후 크레딧 정상 추가
- [ ] 결제 실패 시 에러 처리
- [ ] 가격표 페이지 렌더링
- [ ] 크레딧 잔여량 표시

**테스트**:

- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

**완료 처리**:

- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

→ [Phase 10: 성장 기능](./phase-10-growth.md)
