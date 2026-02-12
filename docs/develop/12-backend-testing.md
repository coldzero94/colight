# Go 백엔드 테스트 전략

> Go 표준 testing + testify + enttest 기반 테스트 계층, AI 모킹, DB 테스트 패턴

---

## 1. 테스트 계층

| 계층 | 대상 | 프레임워크 | DB |
|------|------|-----------|-----|
| **단위 테스트** | Service 레이어 (비즈니스 로직) | `testing` + `testify` | Mock |
| **통합 테스트** | Controller 레이어 (HTTP → Service → DB) | `testing` + `httptest` | SQLite in-memory |
| **Ent 테스트** | 스키마 정합성, 관계, 제약조건 | `enttest` + SQLite | SQLite in-memory |

### 테스트 디렉토리 구조

```
apps/backend/
├── internal/
│   ├── service/
│   │   ├── experience.go
│   │   ├── experience_test.go        # Service 단위 테스트
│   │   ├── coaching.go
│   │   └── coaching_test.go
│   ├── controller/
│   │   ├── experience.go
│   │   ├── experience_test.go        # Controller 통합 테스트
│   │   └── testutil_test.go          # 테스트 헬퍼
│   └── infrastructure/
│       └── ai/
│           ├── client.go
│           ├── client_test.go         # AI 응답 파싱 테스트
│           └── mock_client.go         # Mock 구현
├── ent/
│   └── schema/
│       └── experience_test.go        # Ent 스키마 테스트
├── testdata/                          # AI 응답 fixture 파일
│   ├── ai/
│   │   ├── weapon_tagging_response.json
│   │   ├── company_analysis_response.json
│   │   └── draft_coaching_response.json
│   └── crawling/
│       ├── jobkorea_sample.html
│       └── wanted_sample.html
└── testutil/                          # 공통 테스트 유틸리티
    ├── db.go                          # enttest 헬퍼
    ├── seed.go                        # 테스트 시드 데이터
    └── fixtures.go                    # fixture 로더
```

---

## 2. 테스트 프레임워크

### 의존성

```bash
# apps/backend
go get github.com/stretchr/testify
go get github.com/mattn/go-sqlite3  # enttest SQLite 드라이버
```

### 사용하는 패키지

| 패키지 | 용도 |
|--------|------|
| `testing` | Go 표준 테스트 프레임워크 |
| `testify/assert` | 가독성 높은 assertion (실패해도 테스트 계속) |
| `testify/require` | 필수 assertion (실패 시 테스트 즉시 중단) |
| `testify/mock` | Interface mocking |
| `net/http/httptest` | HTTP 핸들러 테스트 |
| `enttest` | Ent 스키마 기반 in-memory DB 테스트 |

### assertion 사용 규칙

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // require: 실패 시 즉시 중단 (전제 조건)
    result, err := someFunction()
    require.NoError(t, err)          // err가 nil이 아니면 테스트 중단
    require.NotNil(t, result)

    // assert: 실패해도 계속 실행 (검증)
    assert.Equal(t, "expected", result.Name)
    assert.Len(t, result.Tags, 3)
    assert.True(t, result.IsActive)
}
```

---

## 3. AI 클라이언트 모킹

### Interface 정의

```go
// internal/infrastructure/ai/client.go
package ai

import "context"

type ChatRequest struct {
    Model    string
    Messages []Message
    MaxTokens int
}

type ChatResponse struct {
    Content      string
    InputTokens  int
    OutputTokens int
    Model        string
}

type StreamChunk struct {
    Content string
    Done    bool
    Error   error
}

type EmbeddingRequest struct {
    Model string
    Input []string
}

type EmbeddingResponse struct {
    Embeddings [][]float64
}

// AIClient defines the interface for AI API calls
type AIClient interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    StreamCompletion(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
}
```

### Mock 구현

```go
// internal/infrastructure/ai/mock_client.go
package ai

import (
    "context"

    "github.com/stretchr/testify/mock"
)

type MockAIClient struct {
    mock.Mock
}

func (m *MockAIClient) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*ChatResponse), args.Error(1)
}

func (m *MockAIClient) StreamCompletion(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(<-chan StreamChunk), args.Error(1)
}

func (m *MockAIClient) CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*EmbeddingResponse), args.Error(1)
}
```

### Mock 사용 예시 (Service 테스트)

```go
// internal/service/weapon_tagging_test.go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "colight/internal/infrastructure/ai"
)

func TestWeaponTagging_Success(t *testing.T) {
    // Arrange
    mockAI := new(ai.MockAIClient)
    svc := NewWeaponTaggingService(mockAI)

    mockAI.On("ChatCompletion", mock.Anything, mock.MatchedBy(func(req ai.ChatRequest) bool {
        return req.Model == "gpt-4.1-mini"
    })).Return(&ai.ChatResponse{
        Content: `{"weapons": [{"code": "W01", "confidence": 0.85}, {"code": "W03", "confidence": 0.72}]}`,
        InputTokens:  500,
        OutputTokens: 100,
    }, nil)

    // Act
    result, err := svc.TagExperience(context.Background(), "프로젝트 리더로서 팀을 이끌었습니다...")

    // Assert
    require.NoError(t, err)
    assert.Len(t, result.Weapons, 2)
    assert.Equal(t, "W01", result.Weapons[0].Code)
    assert.InDelta(t, 0.85, result.Weapons[0].Confidence, 0.01)

    mockAI.AssertExpectations(t)
}

func TestWeaponTagging_AIError(t *testing.T) {
    mockAI := new(ai.MockAIClient)
    svc := NewWeaponTaggingService(mockAI)

    mockAI.On("ChatCompletion", mock.Anything, mock.Anything).
        Return(nil, errors.New("rate limit exceeded"))

    _, err := svc.TagExperience(context.Background(), "경험 내용...")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "rate limit")
}
```

### Fixture 파일 활용

```go
// testutil/fixtures.go
package testutil

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"
)

// LoadFixture reads a fixture file from testdata directory
func LoadFixture(t *testing.T, path string) []byte {
    t.Helper()

    _, filename, _, _ := runtime.Caller(0)
    rootDir := filepath.Dir(filepath.Dir(filename)) // apps/backend/
    fullPath := filepath.Join(rootDir, "testdata", path)

    data, err := os.ReadFile(fullPath)
    if err != nil {
        t.Fatalf("failed to load fixture %s: %v", path, err)
    }
    return data
}
```

```go
// 사용 예시
func TestParseCompanyAnalysis(t *testing.T) {
    fixture := testutil.LoadFixture(t, "ai/company_analysis_response.json")

    result, err := ParseCompanyAnalysis(string(fixture))
    require.NoError(t, err)
    assert.NotEmpty(t, result.CompanyName)
    assert.Len(t, result.Strengths, 3)
}
```

---

## 4. DB 테스트 패턴

### enttest를 사용한 in-memory DB

```go
// testutil/db.go
package testutil

import (
    "context"
    "testing"

    "colight/ent"
    "colight/ent/enttest"

    _ "github.com/mattn/go-sqlite3"
)

// NewTestClient creates an Ent client with SQLite in-memory DB
func NewTestClient(t *testing.T) *ent.Client {
    t.Helper()

    client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1",
        enttest.WithOptions(ent.Log(t.Log)),
    )

    t.Cleanup(func() {
        client.Close()
    })

    return client
}
```

### 테스트별 격리

각 테스트는 독립적인 DB 인스턴스를 사용하여 테스트 간 간섭을 방지한다.

```go
func TestExperienceCreate(t *testing.T) {
    // 이 테스트만의 독립적인 DB
    client := testutil.NewTestClient(t)
    ctx := context.Background()

    // 시드 데이터 삽입
    testutil.SeedWeapons(t, client)

    // 테스트 실행
    exp, err := client.Experience.Create().
        SetTitle("테스트 경험").
        SetSituation("상황 설명").
        SetTask("과제 설명").
        SetAction("행동 설명").
        SetResult("결과 설명").
        SetUserID(testutil.TestUserID).
        Save(ctx)

    require.NoError(t, err)
    assert.Equal(t, "테스트 경험", exp.Title)
}

func TestExperienceList(t *testing.T) {
    // 별도의 독립적인 DB — 위 테스트의 데이터 영향 없음
    client := testutil.NewTestClient(t)
    // ...
}
```

### 시드 데이터 함수

```go
// testutil/seed.go
package testutil

import (
    "context"
    "testing"

    "colight/ent"

    "github.com/google/uuid"
)

var TestUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// SeedWeapons inserts the 7 weapon categories for testing
func SeedWeapons(t *testing.T, client *ent.Client) {
    t.Helper()
    ctx := context.Background()

    weapons := []struct {
        Code string
        Name string
    }{
        {"W01", "리더십/조직관리"},
        {"W02", "문제해결/분석"},
        {"W03", "소통/협업"},
        {"W04", "도전/추진력"},
        {"W05", "창의/혁신"},
        {"W06", "전문성/학습"},
        {"W07", "고객지향/서비스"},
    }

    bulk := make([]*ent.WeaponCategoryCreate, len(weapons))
    for i, w := range weapons {
        bulk[i] = client.WeaponCategory.Create().
            SetCode(w.Code).
            SetName(w.Name)
    }

    _, err := client.WeaponCategory.CreateBulk(bulk...).Save(ctx)
    if err != nil {
        t.Fatalf("failed to seed weapons: %v", err)
    }
}

// SeedTestUser inserts a test user profile
func SeedTestUser(t *testing.T, client *ent.Client) *ent.UserProfile {
    t.Helper()
    ctx := context.Background()

    user, err := client.UserProfile.Create().
        SetID(TestUserID).
        SetEmail("test@colight.kr").
        SetTargetJob("백엔드 개발자").
        SetTargetIndustry("IT").
        SetCredits(10).
        Save(ctx)
    if err != nil {
        t.Fatalf("failed to seed test user: %v", err)
    }
    return user
}

// SeedExperiences inserts sample experiences for testing
func SeedExperiences(t *testing.T, client *ent.Client, userID uuid.UUID, count int) []*ent.Experience {
    t.Helper()
    ctx := context.Background()

    var experiences []*ent.Experience
    for i := 0; i < count; i++ {
        exp, err := client.Experience.Create().
            SetTitle(fmt.Sprintf("테스트 경험 %d", i+1)).
            SetSituation(fmt.Sprintf("상황 %d", i+1)).
            SetTask(fmt.Sprintf("과제 %d", i+1)).
            SetAction(fmt.Sprintf("행동 %d", i+1)).
            SetResult(fmt.Sprintf("결과 %d", i+1)).
            SetUserID(userID).
            Save(ctx)
        if err != nil {
            t.Fatalf("failed to seed experience %d: %v", i+1, err)
        }
        experiences = append(experiences, exp)
    }
    return experiences
}
```

### Controller 통합 테스트

```go
// internal/controller/experience_test.go
package controller

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "colight/testutil"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *ent.Client) {
    t.Helper()
    gin.SetMode(gin.TestMode)

    client := testutil.NewTestClient(t)
    testutil.SeedWeapons(t, client)
    testutil.SeedTestUser(t, client)

    r := gin.New()

    // 테스트용 인증 미들웨어 (TestUserID 주입)
    r.Use(func(c *gin.Context) {
        c.Set("user_id", testutil.TestUserID.String())
        c.Next()
    })

    ctrl := NewExperienceController(client)
    v1 := r.Group("/v1")
    v1.GET("/experiences", ctrl.List)
    v1.POST("/experiences", ctrl.Create)
    v1.GET("/experiences/:id", ctrl.Get)
    v1.PUT("/experiences/:id", ctrl.Update)
    v1.DELETE("/experiences/:id", ctrl.Delete)

    return r, client
}

func TestExperienceController_Create(t *testing.T) {
    router, _ := setupTestRouter(t)

    body := map[string]string{
        "title":     "프로젝트 리더 경험",
        "situation": "대학교 캡스톤 프로젝트에서",
        "task":      "팀장으로서 일정 관리를 맡았다",
        "action":    "Jira를 도입하여 Sprint를 운영했다",
        "result":    "2주 앞당겨 프로젝트를 완료했다",
    }
    bodyJSON, _ := json.Marshal(body)

    req := httptest.NewRequest(http.MethodPost, "/v1/experiences", bytes.NewReader(bodyJSON))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var resp map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)
    assert.Equal(t, "프로젝트 리더 경험", resp["data"].(map[string]interface{})["title"])
}

func TestExperienceController_List(t *testing.T) {
    router, client := setupTestRouter(t)

    // 테스트 데이터 삽입
    testutil.SeedExperiences(t, client, testutil.TestUserID, 3)

    req := httptest.NewRequest(http.MethodGet, "/v1/experiences", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var resp map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)

    data := resp["data"].([]interface{})
    assert.Len(t, data, 3)
}
```

---

## 5. Phase별 최소 테스트

| Phase | 테스트 대상 | 최소 테스트 수 | 우선순위 |
|-------|-----------|-------------|---------|
| **Phase 0** | Ent 스키마 생성, 마이그레이션, 시드 데이터 삽입 | 5 | 필수 |
| **Phase 1** | JWT 미들웨어 (유효/만료/누락 토큰), 프로필 CRUD | 5 | 필수 |
| **Phase 2** | 경험 CRUD, 목록 필터/검색, STAR 필드 유효성 | 8 | 필수 |
| **Phase 2.1** | 무기 태깅 AI 응답 파싱, confidence 정렬 | 5 | 필수 |
| **Phase 3** | 크롤러 (잡코리아/캐치 HTML 파싱), URL 정규화 | 10 | 필수 |
| **Phase 3.1** | DART API 응답 파싱, 네이버 뉴스 변환 | 5 | 필수 |
| **Phase 3.2** | AI 분석 파이프라인, 캐시 히트/미스, 프롬프트 조합 | 8 | 필수 |
| **Phase 4** | 임베딩 유사도 계산, 매칭 정렬, 벡터 검색 | 8 | 필수 |
| **Phase 5** | 문항 패턴 분류, 코칭 프롬프트 템플릿 조합 | 8 | 필수 |
| **Phase 5.1~6** | 초안/첨삭 코칭 스트리밍, 버전 저장, 점수 계산 | 10 | 필수 |

### Phase 0 테스트 상세

```go
// ent/schema/experience_test.go
func TestEntSchema_ExperienceCreate(t *testing.T) {
    client := testutil.NewTestClient(t)
    testutil.SeedTestUser(t, client)

    exp, err := client.Experience.Create().
        SetTitle("테스트").
        SetSituation("상황").
        SetTask("과제").
        SetAction("행동").
        SetResult("결과").
        SetUserID(testutil.TestUserID).
        Save(context.Background())

    require.NoError(t, err)
    assert.NotZero(t, exp.ID)
}

func TestEntSchema_ExperienceRequiredFields(t *testing.T) {
    client := testutil.NewTestClient(t)

    // title 누락 시 에러
    _, err := client.Experience.Create().
        SetSituation("상황").
        SetUserID(testutil.TestUserID).
        Save(context.Background())

    assert.Error(t, err)
}

func TestSeedData_WeaponCategories(t *testing.T) {
    client := testutil.NewTestClient(t)
    testutil.SeedWeapons(t, client)

    weapons, err := client.WeaponCategory.Query().All(context.Background())
    require.NoError(t, err)
    assert.Len(t, weapons, 7)
    assert.Equal(t, "W01", weapons[0].Code)
}
```

### Phase 1 테스트 상세

```go
// internal/middleware/auth_test.go
func TestAuthMiddleware_ValidToken(t *testing.T) {
    // 유효한 JWT 토큰으로 요청 시 user_id가 context에 설정됨
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
    // 만료된 토큰으로 요청 시 401 반환
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
    // Authorization 헤더 누락 시 401 반환
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
    // "Bearer " 접두사 없는 토큰 시 401 반환
}

func TestProfileController_GetOrCreate(t *testing.T) {
    // 첫 로그인 시 프로필 자동 생성
}
```

---

## 6. 실행 커맨드

```bash
# ============================================
# 전체 테스트
# ============================================
cd apps/backend
go test ./...

# ============================================
# 커버리지 리포트
# ============================================
go test -cover ./...

# 상세 커버리지 (HTML 리포트)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# ============================================
# Race condition 검사 (CI 필수)
# ============================================
go test -race ./...

# ============================================
# 특정 테스트만 실행
# ============================================
# 패키지 단위
go test ./internal/service/...

# 테스트 함수 이름으로 필터
go test -run TestExperience ./internal/service/...
go test -run TestExperienceController_Create ./internal/controller/...

# 특정 서브테스트
go test -run TestWeaponTagging/Success ./internal/service/...

# ============================================
# 상세 출력 (실패 디버깅)
# ============================================
go test -v ./internal/service/...

# ============================================
# 타임아웃 설정 (AI 관련 테스트)
# ============================================
go test -timeout 30s ./...

# ============================================
# 벤치마크 (매칭 알고리즘 등)
# ============================================
go test -bench=BenchmarkMatching ./internal/service/...
```

### 커버리지 목표

| 대상 | 목표 커버리지 | 비고 |
|------|-------------|------|
| `internal/service/` (비즈니스 로직) | **70%** | 핵심 로직, 최우선 |
| `internal/controller/` (HTTP) | **50%** | 주요 엔드포인트 |
| `internal/infrastructure/` (외부 연동) | **40%** | 파싱/변환 로직 위주 |
| `ent/schema/` (DB 스키마) | **60%** | 제약조건, 관계 검증 |
| 전체 | **55%** | - |

**커버리지에서 제외:**
- `internal/generated/` — oapi-codegen 자동 생성 코드
- `ent/` (schema 제외) — Ent 자동 생성 코드
- `cmd/` — 진입점 (설정만)
- `scripts/` — 시드/마이그레이션 스크립트

### CI에서 테스트 실행

```yaml
# .github/workflows/ci.yml (backend-test job)
- name: Run tests
  working-directory: apps/backend
  run: |
    go test -race -cover -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | tail -1
  # 마지막 줄에 전체 커버리지 % 출력
```
