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

- [ ] 상품 구성:
  | 상품 | 가격 | 내용 |
  |------|------|------|
  | **Starter** | TBD | 코칭 10건 |
  | **Pro** | TBD | 코칭 30건 |
  | **Season Pass** | TBD | 무제한 (기간제) |
- [ ] 상품 정보 DB 또는 상수 관리

### 9.2 Toss Payments SDK 연동

- [ ] @tosspayments/payment-sdk 설치
- [ ] 프론트엔드 결제 위젯 초기화
- [ ] 결제 요청 → Toss 결제 페이지 → 콜백 처리
- [ ] `/api/payments/confirm` 웹훅 엔드포인트 생성
- [ ] 환경변수: TOSS_CLIENT_KEY, TOSS_SECRET_KEY

### 9.3 결제 확인 + 크레딧

- [ ] Toss 웹훅 검증 (서명 확인)
- [ ] 결제 성공 시 user_profiles에 크레딧 추가
- [ ] 결제 실패/취소 처리
- [ ] 결제 이력 로깅

### 9.4 결제 UI

**경로**: `src/app/(main)/pricing/page.tsx`

- [ ] 가격표 페이지 (상품 카드 3개)
- [ ] 현재 플랜 / 잔여 크레딧 표시
- [ ] 결제 버튼 → Toss 결제 플로우
- [ ] 결제 완료/실패 결과 페이지

---

## 완료 체크리스트

- [ ] Toss Payments 테스트 결제 성공
- [ ] 웹훅 검증 후 크레딧 정상 추가
- [ ] 결제 실패 시 에러 처리
- [ ] 가격표 페이지 렌더링
- [ ] 크레딧 잔여량 표시
- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

→ [Phase 10: 성장 기능](./phase-10-growth.md)
