# Phase 10: 성장 기능

> Sprint 8+ | 진행중 | 관련 기능: F04, F05, F06, F15, F16, F17

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | 크롤링 확장, 임베딩 매칭, AI 탐지, 대기업 인재상, 레이더 차트 등 성장 기능 |
| **선행 조건** | MVP 완료 (Phase 0~6.2) |
| **주요 산출물** | 추가 크롤러, 임베딩 매칭 고도화, 분석 도구 |
| **기술 스택** | Playwright, 사람인 API, pgvector, Recharts, 경량 모델 (Gemini/Groq) |

---

## 진행 상태

- [ ] 10.1 Koyeb Worker (Playwright)
- [ ] 10.2 사람인 API
- [ ] 10.3 임베딩 매칭
- [ ] 10.4 AI 탐지 체크
- [ ] 10.5 대기업 인재상 DB
- [ ] 10.6 무기 레이더 차트
- [ ] 10.7 글자수 조절 코칭

---

## 구현 단계

### 10.1 River Job Queue + Playwright (Embedded Mode)

**목표**: Go API 프로세스 내에서 River를 사용하여 Playwright 크롤링을 비동기 처리

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestEnqueueCrawlJob_Success` | `internal/service/crawl_service_test.go` | 크롤링 Job enqueue 및 job_id 반환 확인 |
| `TestCrawlWorker_ParsesHTML` | `internal/worker/crawl_wanted_test.go` | Playwright HTML 추출 및 파싱 정확성 확인 |
| `TestGetJobStatus_ReturnsCorrectState` | `internal/controller/job_controller_test.go` | Job 상태 조회 API 정상 응답 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/crawl_service_test.go` 작성
  - [ ] `internal/worker/crawl_wanted_test.go` 작성
  - [ ] `internal/controller/job_controller_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] River 설치: `go get github.com/riverqueue/river`
  - [ ] River 마이그레이션: `river migrate-up` (PostgreSQL에 river_job, river_leader 테이블 생성)
  - [ ] Playwright 크롤링 Worker 정의 (`internal/worker/crawl_wanted.go`)
    - [ ] PlaywrightCrawlArgs 구조체 (URL, Depth)
    - [ ] CrawlWantedWorker 구현 (Playwright 실행, HTML 추출, 파싱)
  - [ ] cmd/api/main.go에서 River 초기화 (embedded mode)
    - [ ] Workers 등록: `workers.Add(&CrawlWantedWorker{})`
    - [ ] Queues 설정: `"crawl": {MaxWorkers: 1}` (메모리 제약)
    - [ ] riverClient.Start() 백그라운드 실행
  - [ ] API 핸들러에서 Job enqueue
    - [ ] POST /v1/crawl/wanted → riverClient.Insert(ctx, PlaywrightCrawlArgs{...})
    - [ ] 즉시 응답: `{"status": "queued", "job_id": "..."}`
  - [ ] Job 상태 조회 엔드포인트: GET /v1/jobs/:id
  - [ ] River UI 통합 (선택, 모니터링용)
  - [ ] 메모리 모니터링 (Koyeb 512MB 제약)
- [ ] 테스트 통과 확인

**배포**: Koyeb Free 단일 서비스 (Go API + River Worker embedded), 비용 $0

### 10.2 사람인 API

**목표**: 사람인 API를 통한 채용공고 검색 및 조회 연동

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestSaraminClient_SearchJobs` | `internal/service/saramin_service_test.go` | 사람인 API 채용공고 검색 정상 응답 확인 |
| `TestSaraminClient_GetJobDetail` | `internal/service/saramin_service_test.go` | 사람인 API 상세 조회 정상 응답 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/saramin_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] 사람인 API 승인 신청
  - [ ] 사람인 API 클라이언트 구현 (`internal/external/saramin.go`)
  - [ ] 채용공고 검색/상세 조회
  - [ ] 기존 크롤링 파이프라인에 통합 (도메인 분기)
- [ ] 테스트 통과 확인

### 10.3 임베딩 매칭 고도화

**목표**: 벡터 임베딩 기반 경험-공고 매칭 정확도 개선

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestEmbeddingMatch_CosineSimilarity` | `internal/service/matching_service_test.go` | pgvector cosine similarity 검색 결과 정확성 확인 |
| `TestCombinedScore_WeightedAverage` | `internal/service/matching_service_test.go` | 무기 매칭 + 임베딩 매칭 결합 점수 계산 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/matching_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] text-embedding-3-small 기반 경험-공고 벡터 유사도
  - [ ] pgvector cosine similarity 검색
  - [ ] 기존 무기 매칭 + 임베딩 매칭 점수 결합
  - [ ] 매칭 정확도 평가 및 가중치 튜닝
- [ ] 테스트 통과 확인

### 10.4 AI 탐지 체크 (F15)

**목표**: 작성된 자소서의 AI 생성 여부를 판별하고 개선 제안 제공

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestAIDetection_ReturnsScore` | `internal/service/detection_service_test.go` | AI 탐지 점수 반환 확인 |
| `TestAIDetection_GeneratesSuggestions` | `internal/service/detection_service_test.go` | "사람다운" 표현 수정 제안 생성 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/detection_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] 작성된 자소서의 AI 생성 여부 판별
  - [ ] 탐지 점수 및 개선 제안 표시
  - [ ] "사람다운" 표현으로 수정 코칭
- [ ] 테스트 통과 확인

### 10.5 대기업 인재상 DB (F17)

**목표**: Top 100 기업 인재상 데이터를 수집하고 자동 매칭에 활용

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestGetTalentProfile_ByCompany` | `internal/service/talent_service_test.go` | 기업별 인재상 조회 정상 응답 확인 |
| `TestMatchTalentProfile_PrioritizesDB` | `internal/service/talent_service_test.go` | 사전 DB 우선 조회 후 매칭 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/talent_service_test.go` 작성
- [ ] 구현 (GREEN)
  - [ ] talent_profiles 테이블에 Top 100 기업 데이터 수집
  - [ ] 기업별 인재상 키워드, 핵심 가치
  - [ ] 자동 매칭 시 사전 DB 우선 조회
  - [ ] 주기적 업데이트 파이프라인
- [ ] 테스트 통과 확인

### 10.6 무기 레이더 차트 (F05)

**목표**: 7대 무기 보유 점수를 레이더 차트로 시각화

**테스트 명세**:

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('WeaponRadarChart')` / `it('renders 7 weapon axes')` | `src/components/analytics/__tests__/WeaponRadarChart.test.tsx` | 7대 무기(W01-W07) 축 렌더링 확인 |
| `describe('WeaponRadarChart')` / `it('highlights strong and weak weapons')` | `src/components/analytics/__tests__/WeaponRadarChart.test.tsx` | 강한/약한 무기 하이라이트 확인 |
| `describe('WeaponRadarChart')` / `it('exports chart as image')` | `src/components/analytics/__tests__/WeaponRadarChart.test.tsx` | 차트 이미지 내보내기 기능 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `src/components/analytics/__tests__/WeaponRadarChart.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] Recharts RadarChart 컴포넌트
  - [ ] 7대 무기(W01-W07) 축 표시
  - [ ] 사용자 경험 기반 무기 보유 점수 시각화
  - [ ] 강한/약한 무기 하이라이트
  - [ ] 보완 추천 메시지
- [ ] 테스트 통과 확인

### 10.7 글자수 조절 코칭 (F16)

**목표**: 글자수 제한에 맞춘 축약/보강 코칭 제공

**테스트 명세**:

> 패턴 참고: docs/develop/12-backend-testing.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `TestCharCountCoaching_OverLimit` | `internal/service/coaching_service_test.go` | 글자수 초과 시 축약 코칭 생성 확인 |
| `TestCharCountCoaching_UnderLimit` | `internal/service/coaching_service_test.go` | 글자수 부족 시 보강 코칭 생성 확인 |

> 패턴 참고: docs/develop/08-testing-strategy.md

| 테스트 | 파일 | 설명 |
|--------|------|------|
| `describe('CharCountGuide')` / `it('displays current vs limit character count')` | `src/components/coaching/__tests__/CharCountGuide.test.tsx` | 현재 글자수 vs 제한 글자수 표시 확인 |

**구현 체크리스트**:

- [ ] 테스트 작성 (RED)
  - [ ] `internal/service/coaching_service_test.go`에 글자수 코칭 테스트 추가
  - [ ] `src/components/coaching/__tests__/CharCountGuide.test.tsx` 작성
- [ ] 구현 (GREEN)
  - [ ] 현재 글자수 vs 제한 글자수 표시
  - [ ] 초과 시 축약 코칭 / 부족 시 보강 코칭
  - [ ] 섹션별 글자수 배분 가이드
  - [ ] AI 기반 자동 조절 제안
- [ ] 테스트 통과 확인

---

## Phase 완료 체크리스트

**기능 검증**:

- [ ] River embedded mode 동작 확인 (Go API 내 Worker 실행)
- [ ] 원티드 크롤링 비동기 작업 정상 처리 (MaxWorkers: 1)
- [ ] 사람인 API 연동 정상 동작
- [ ] 임베딩 매칭 정확도 개선 확인
- [ ] AI 탐지 체크 결과 표시
- [ ] 인재상 DB 100개 기업 이상 등록
- [ ] 레이더 차트 7대 무기 시각화
- [ ] 글자수 코칭 동작

**테스트**:

- [ ] `moon run backend:test` → 전체 통과
- [ ] `moon run web:test` → 전체 통과
- [ ] `moon run :lint` → 경고 0건
- [ ] `moon run web:build` → 빌드 성공

**완료 처리**:

- [ ] phases/README.md 상태 업데이트

---

## 다음 Phase

성장 기능은 지속적으로 추가됩니다.
