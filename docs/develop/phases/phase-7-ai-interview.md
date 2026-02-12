# Phase 7: AI 경험 인터뷰

> Sprint 6 | 예상 공수: 2일 | 관련 기능: F02

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | AI와 대화하며 경험을 발굴하고 STAR 구조 경험 카드를 자동 생성 |
| **선행 조건** | Phase 2 (경험 CRUD), Phase 2.1 (무기 태깅) |
| **주요 산출물** | 인터뷰 채팅 UI, 인터뷰 API (스트리밍), 경험 카드 자동 생성 |
| **기술 스택** | GPT-4.1 mini, Vercel AI SDK streamText, React Query |

---

## 진행 상태

- [ ] 7.1 채팅 UI
- [ ] 7.2 인터뷰 API
- [ ] 7.3 경험 카드 자동 생성
- [ ] 7.4 자동 무기 태깅

---

## 구현 단계

### 7.1 채팅 UI

**경로**: `src/app/(main)/experiences/interview/page.tsx`

- [ ] 인터뷰 페이지 라우트 생성
- [ ] 메시지 버블 컴포넌트 (AI / 사용자 구분)
- [ ] 텍스트 입력 + 전송 버튼
- [ ] 스트리밍 응답 실시간 표시 (useChat 또는 커스텀 훅)
- [ ] 인터뷰 단계 진행 인디케이터
- [ ] 스크롤 자동 하단 이동

### 7.2 인터뷰 API

**경로**: `src/app/api/experiences/interview/route.ts` (POST, streaming)

- [ ] GPT-4.1 mini 기반 멀티턴 대화형 API
- [ ] prompt_templates에서 인터뷰 프롬프트 로드
- [ ] 인터뷰 단계 관리:
  1. **가볍게** — 최근 활동, 관심사 탐색
  2. **기억에 남는 순간** — 구체적 경험 발굴
  3. **어려웠던 점** — 도전/갈등 상황 파악
  4. **해결법** — 실제 행동과 과정
  5. **결과/배운 점** — 성과와 인사이트
- [ ] 대화 컨텍스트 누적 (messages 배열)
- [ ] Vercel AI SDK streamText로 스트리밍 응답

### 7.3 경험 카드 자동 생성

- [ ] 인터뷰 완료 시 AI가 대화에서 STAR 구조 자동 추출
- [ ] 추출된 STAR 데이터를 사용자에게 미리보기 표시
- [ ] 사용자 확인/수정 후 experiences 테이블에 저장
- [ ] 임베딩 생성 (text-embedding-3-small)

### 7.4 자동 무기 태깅

- [ ] 경험 저장 완료 시 Phase 2.1 무기 태깅 API 자동 호출
- [ ] `POST /api/experiences/[id]/tag` 트리거
- [ ] 태깅 결과 경험 카드에 즉시 반영

---

## 완료 체크리스트

- [ ] AI 인터뷰 5단계 대화 흐름 정상 동작
- [ ] 스트리밍 응답 실시간 표시
- [ ] STAR 구조 자동 추출 및 사용자 확인 플로우
- [ ] 경험 저장 후 무기 태깅 자동 실행
- [ ] 에러 시 재시도 가능
- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

→ [Phase 7.1: 경험 추천 강화](./phase-7.1-experience-recommend.md)
