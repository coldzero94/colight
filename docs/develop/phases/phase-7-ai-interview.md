# Phase 7: AI 경험 인터뷰

> Sprint 7 | 예상 공수: 2일 | 관련 기능: F02

> **Architecture Note**: 원본 문서의 TypeScript/Next.js API Routes 경로는 Go 백엔드 구현으로 대체됨.
> 실제 구현: `internal/service/interview_service.go`, `internal/controller/interview_controller.go`

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | AI와 대화하며 경험을 발굴하고 STAR 구조 경험 카드를 자동 생성 |
| **선행 조건** | Phase 2 (경험 CRUD), Phase 2.1 (무기 태깅) |
| **주요 산출물** | 인터뷰 채팅 UI, 인터뷰 API, 경험 카드 자동 생성 |
| **기술 스택** | 경량 모델 (Gemini/Groq), Go backend (Gin), React Query |

---

## 진행 상태

- [x] 7.1 채팅 UI
- [x] 7.2 인터뷰 API
- [x] 7.3 경험 카드 자동 생성
- [x] 7.4 자동 무기 태깅

---

## 구현 단계

### 7.1 채팅 UI

**경로**: `src/app/(main)/experiences/interview/page.tsx`

#### 테스트 명세

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `describe('InterviewChat')` | `src/components/interview/__tests__/interview-chat.test.tsx` | 메시지 버블 렌더링 (AI/사용자 구분), 입력 전송 동작 |
| `describe('InterviewProgress')` | `src/components/interview/__tests__/interview-progress.test.tsx` | 5단계 진행 인디케이터 렌더링, 현재 단계 하이라이트 |
| `describe('useInterviewChat')` | `src/hooks/__tests__/use-interview-chat.test.ts` | 스트리밍 메시지 수신, 메시지 배열 누적, 에러 처리 |
| `describe('InterviewTimer')` | `src/components/interview/__tests__/interview-timer.test.tsx` | 타이머 표시, 시간 경과 포맷팅 |

#### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `src/components/interview/__tests__/interview-chat.test.tsx` 작성
  - [x] `src/components/interview/__tests__/interview-progress.test.tsx` 작성
  - [x] `src/hooks/__tests__/use-interview-chat.test.ts` 작성
  - [x] `src/components/interview/__tests__/interview-timer.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 인터뷰 페이지 라우트 생성
  - [x] 메시지 버블 컴포넌트 (AI / 사용자 구분)
  - [x] 텍스트 입력 + 전송 버튼
  - [x] useInterviewChat 커스텀 훅으로 상태 관리
  - [x] 인터뷰 단계 진행 인디케이터
  - [x] 스크롤 자동 하단 이동
- [x] 테스트 통과 확인

### 7.2 인터뷰 API

**경로**: `internal/controller/interview_controller.go`, `internal/service/interview_service.go`

#### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `TestInterviewService_GenerateQuestion` | `internal/service/interview_service_test.go` | MockAIClient 사용, 단계별 질문 생성 확인 (5단계 각각) |
| `TestInterviewService_ParseStage` | `internal/service/interview_service_test.go` | 대화 컨텍스트에서 현재 인터뷰 단계 판별 |
| `TestInterviewController_Unauthorized` | `internal/controller/interview_controller_test.go` | 인증 없는 요청 시 401 반환 |
| `TestInterviewController_StreamResponse` | `internal/controller/interview_controller_test.go` | MockAIClient 사용, 스트리밍 응답 형식 검증 |

#### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `internal/service/interview_service_test.go` 작성 (8 tests)
  - [x] `internal/controller/interview_controller_test.go` 작성 (7 tests)
- [x] 구현 (GREEN)
  - [x] 경량 모델 (Gemini/Groq) 기반 멀티턴 대화형 API
  - [x] 인터뷰 단계 관리 (5단계: warmup → memory → challenge → solution → outcome)
  - [x] 대화 컨텍스트 누적 (messages 배열, 클라이언트 관리)
  - [x] 3개 엔드포인트: POST /v1/interview/{question,extract,save}
- [x] 테스트 통과 확인

### 7.3 경험 카드 자동 생성

#### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `TestInterviewService_ExtractSTAR` | `internal/service/interview_service_test.go` | MockAIClient 사용, 대화 내용에서 STAR 구조 추출 (Situation/Task/Action/Result 각 필드 비어있지 않음) |

#### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `internal/service/interview_service_test.go`에 STAR 추출 테스트 추가
  - [x] `src/components/interview/__tests__/star-preview.test.tsx` 작성
- [x] 구현 (GREEN)
  - [x] 인터뷰 완료 시 AI가 대화에서 STAR 구조 자동 추출 (POST /v1/interview/extract)
  - [x] 추출된 STAR 데이터를 사용자에게 미리보기 표시 (StarPreview 컴포넌트)
  - [x] 사용자 확인/수정 후 experiences 테이블에 저장 (POST /v1/interview/save)
- [x] 테스트 통과 확인

### 7.4 자동 무기 태깅

#### 테스트 명세

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 검증 내용 |
|--------|------|----------|
| `TestInterviewService_TriggerAutoTag` | `internal/service/interview_service_test.go` | 경험 저장 완료 후 무기 태깅 API 자동 호출 확인 |

#### 구현 체크리스트

- [x] 테스트 작성 (RED)
  - [x] `internal/service/interview_service_test.go`에 자동 태깅 트리거 테스트 추가
- [x] 구현 (GREEN)
  - [x] SaveExperience 시 WeaponTaggingService.TagExperience 자동 호출
  - [x] 태깅 결과 응답에 `tagged: true/false` 반환
- [x] 테스트 통과 확인

---

## 완료 체크리스트

- [x] AI 인터뷰 5단계 대화 흐름 정상 동작
- [x] AI 질문 생성 및 응답 표시
- [x] STAR 구조 자동 추출 및 사용자 확인 플로우
- [x] 경험 저장 후 무기 태깅 자동 실행
- [x] 에러 시 재시도 가능
- [x] phases/README.md 상태 업데이트
- [x] `moon run backend:test` → 전체 통과
- [x] `moon run web:test` → 전체 통과
- [x] `moon run :lint` → 경고 0건
- [x] `moon run web:build` → 빌드 성공

---

## 다음 Phase

→ [Phase 7.1: 경험 추천 강화](./phase-7.1-experience-recommend.md)
