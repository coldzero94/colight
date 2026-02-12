# 배포 & CI/CD 가이드

> Koyeb(Go) + Vercel(Next.js) + Supabase(DB) 배포 아키텍처 및 GitHub Actions CI 파이프라인

---

## 1. 배포 아키텍처 개요

```
┌─────────────────────────────────────────────────────────┐
│                      사용자 (브라우저)                      │
└──────────────┬─────────────────────┬────────────────────┘
               │                     │
               ▼                     ▼
┌──────────────────────┐  ┌─────────────────────────────┐
│   Vercel (Next.js)   │  │     Koyeb (Go Backend)      │
│   - SSR/RSC 렌더링     │  │     - REST API 서버          │
│   - 정적 자산 서빙       │  │     - River 백그라운드 잡      │
│   - Edge Functions    │  │     - AI/외부 API 프록시       │
└──────────┬───────────┘  └──────────┬──────────────────┘
           │                         │
           │    NEXT_PUBLIC_API_URL   │
           │ ────────────────────────>│
           │                         │
           │                         ▼
           │              ┌────────────────────┐
           │              │  Supabase (DB)      │
           │              │  - PostgreSQL 16     │
           │              │  - pgvector          │
           └─────────────>│  - Auth (JWT 발급)    │
                          └────────────────────┘
```

### 배포 플랫폼 선택 이유

| 플랫폼 | 역할 | 선택 이유 |
|--------|------|----------|
| **Koyeb** | Go 백엔드 | Docker 컨테이너 무료 티어, GitHub 자동 배포, 512MB RAM |
| **Vercel** | Next.js 프론트엔드 | Next.js 공식 호스팅, Preview Deployments, Edge Network |
| **Supabase** | PostgreSQL + Auth | 무료 티어 충분, pgvector 내장, JWT 기반 인증 |

---

## 2. Go 백엔드 배포 (Koyeb)

### Dockerfile (Multi-stage Build)

```dockerfile
# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Dependency caching
COPY apps/backend/go.mod apps/backend/go.sum ./
RUN go mod download

# Copy source
COPY apps/backend/ .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /api ./cmd/api

# ============================================
# Stage 2: Runtime (distroless)
# ============================================
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /api /api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Koyeb은 8000 포트를 기본으로 사용
EXPOSE 8000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
```

**바이너리 크기 최적화:**
- `CGO_ENABLED=0`: 정적 링킹 (C 의존성 제거)
- `-ldflags="-w -s"`: 디버그 심볼 제거 (~40% 크기 감소)
- `distroless` 이미지: 셸/패키지 매니저 없는 최소 이미지 (~2MB base)
- 최종 이미지 예상 크기: **~20~30MB**

### Koyeb 서비스 설정

```yaml
# koyeb.yaml (참고용 — 실제로는 대시보드에서 설정)
service:
  name: colight-api
  type: web
  region: was  # Washington DC (한국 사용자 기준 가장 가까운 무료 리전)

  docker:
    dockerfile: apps/backend/Dockerfile
    context: .  # monorepo root

  instance_type: nano  # 무료 티어 (512MB RAM, shared vCPU)

  ports:
    - port: 8000
      protocol: http

  health_checks:
    - type: http
      port: 8000
      path: /health
      interval_seconds: 30
      timeout_seconds: 5
      healthy_threshold: 2
      unhealthy_threshold: 3

  env:
    - key: DATABASE_URL
      value: "{{secret.DATABASE_URL}}"
    - key: SUPABASE_JWT_SECRET
      value: "{{secret.SUPABASE_JWT_SECRET}}"
    - key: ANTHROPIC_API_KEY
      value: "{{secret.ANTHROPIC_API_KEY}}"
    - key: OPENAI_API_KEY
      value: "{{secret.OPENAI_API_KEY}}"
    - key: DART_API_KEY
      value: "{{secret.DART_API_KEY}}"
    - key: NAVER_CLIENT_ID
      value: "{{secret.NAVER_CLIENT_ID}}"
    - key: NAVER_CLIENT_SECRET
      value: "{{secret.NAVER_CLIENT_SECRET}}"
    - key: GIN_MODE
      value: release
    - key: PORT
      value: "8000"
```

### 배포 트리거

Koyeb은 GitHub 연동 시 push 이벤트로 자동 배포된다.

```
GitHub push (main 브랜치)
    │
    ▼
Koyeb 자동 빌드
    │
    ├── 1. Docker build (multi-stage)
    ├── 2. Health check 통과 대기
    └── 3. Blue-Green 전환 (무중단 배포)
```

### Koyeb 무료 티어 제약 및 대응

| 제약 | 값 | 대응 |
|------|---|------|
| RAM | 512MB | Go 바이너리 경량화, 메모리 프로파일링 |
| CPU | shared vCPU | AI 호출은 외부 API이므로 CPU 부담 적음 |
| Cold Start | ~2~5초 | Health check로 warm 유지, 무료 티어는 슬립 있음 |
| 대역폭 | 100GB/월 | JSON API만 서빙 (이미지 없음), 충분 |

### 메모리 최적화 (512MB 제한)

```go
// cmd/api/main.go — 메모리 제한 설정
import "runtime/debug"

func init() {
    // GC를 더 자주 실행하여 메모리 사용량 제한
    debug.SetGCPercent(50) // 기본값 100에서 50으로 낮춤

    // 최대 메모리 소프트 리밋 (450MB, 512MB 중 여유분 확보)
    debug.SetMemoryLimit(450 * 1024 * 1024)
}
```

---

## 3. Next.js 프론트엔드 배포 (Vercel)

### Monorepo 설정

Vercel 프로젝트 설정에서 Root Directory를 `apps/web`으로 지정한다.

```
Vercel Dashboard → Project Settings → General
├── Root Directory: apps/web
├── Framework Preset: Next.js
├── Build Command: pnpm run build  (자동 감지)
└── Output Directory: .next  (자동 감지)
```

### 환경변수 설정

```bash
# Vercel Dashboard → Settings → Environment Variables
# 또는 Vercel CLI:

# 클라이언트 노출 (NEXT_PUBLIC_ prefix)
vercel env add NEXT_PUBLIC_SUPABASE_URL production
vercel env add NEXT_PUBLIC_SUPABASE_ANON_KEY production
vercel env add NEXT_PUBLIC_API_URL production  # Go backend URL

# 서버 전용 (Server Components / API Routes에서만 접근)
# Next.js 프론트엔드에는 AI API 키가 불필요 (Go 백엔드가 처리)
```

### 환경별 API URL

| 환경 | `NEXT_PUBLIC_API_URL` | 용도 |
|------|----------------------|------|
| Development | `http://localhost:8000` | 로컬 Go 서버 |
| Preview | `https://colight-api-staging.koyeb.app` | PR 미리보기용 |
| Production | `https://colight-api.koyeb.app` | 운영 |

### Preview Deployments

PR이 생성되면 Vercel이 자동으로 Preview URL을 배포한다.

```
PR #42 생성
    │
    ▼
Vercel Preview 배포
    │
    ├── URL: https://colight-git-feature-42.vercel.app
    ├── 환경변수: Preview 설정 사용
    └── PR 코멘트에 미리보기 URL 자동 게시
```

**주의사항:**
- Preview 환경에서는 staging DB를 사용해야 함 (production DB 연결 금지)
- `NEXT_PUBLIC_API_URL`을 staging backend로 설정

### vercel.json 설정

```json
{
  "buildCommand": "pnpm run build",
  "framework": "nextjs",
  "regions": ["icn1"],
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Cache-Control", "value": "no-store" }
      ]
    }
  ]
}
```

---

## 4. DB 마이그레이션 배포

### 마이그레이션 파이프라인

```
Ent 스키마 수정 (apps/backend/ent/schema/)
    │
    ├── 1. go generate ./ent
    │       → Ent 코드 재생성
    │
    ├── 2. moon run backend:migrate-diff -- name=add_feature
    │       → atlas migrate diff --env local
    │       → migrations/YYYYMMDDHHMMSS_add_feature.sql 생성
    │
    ├── 3. 코드 리뷰 (SQL 파일 확인)
    │
    └── 4. CI에서 atlas migrate apply
            → production DB에 마이그레이션 적용
```

### CI에서 마이그레이션 자동 적용

```yaml
# .github/workflows/migrate.yml
name: Database Migration

on:
  push:
    branches: [main]
    paths:
      - 'apps/backend/migrations/**'

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Atlas
        run: |
          curl -sSf https://atlasgo.sh | sh

      - name: Apply migrations
        working-directory: apps/backend
        run: |
          atlas migrate apply \
            --url "${{ secrets.DATABASE_URL }}" \
            --dir "file://migrations"
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}

      - name: Verify migration status
        working-directory: apps/backend
        run: |
          atlas migrate status \
            --url "${{ secrets.DATABASE_URL }}" \
            --dir "file://migrations"
```

### 롤백 전략

Atlas는 자동 롤백을 지원하지 않으므로 수동 롤백 마이그레이션을 준비한다.

```bash
# 1. 문제 확인
atlas migrate status --url $DATABASE_URL --dir file://migrations

# 2. 롤백 마이그레이션 생성
# 이전 스키마 상태로 Ent 스키마를 되돌린 후:
moon run backend:migrate-diff -- name=rollback_add_feature

# 3. 롤백 적용
atlas migrate apply --url $DATABASE_URL --dir file://migrations
```

**롤백 원칙:**
- 비파괴적 마이그레이션 우선 (ADD COLUMN, CREATE TABLE)
- DROP/ALTER 시 반드시 롤백 SQL을 먼저 작성
- 대규모 스키마 변경은 여러 단계로 분리 (expand-contract pattern)

---

## 5. GitHub Actions CI 파이프라인

### 전체 CI 구조

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  # ============================================
  # Go Backend
  # ============================================
  backend-lint:
    name: Backend Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: apps/backend/go.sum
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          working-directory: apps/backend
          version: latest

  backend-test:
    name: Backend Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: apps/backend/go.sum
      - name: Run tests
        working-directory: apps/backend
        run: go test -race -cover ./...

  # ============================================
  # Next.js Frontend
  # ============================================
  frontend-lint:
    name: Frontend Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          cache-dependency-path: apps/web/pnpm-lock.yaml
      - name: Install dependencies
        working-directory: apps/web
        run: pnpm install --frozen-lockfile
      - name: Lint
        working-directory: apps/web
        run: pnpm run lint

  frontend-typecheck:
    name: Frontend Type Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          cache-dependency-path: apps/web/pnpm-lock.yaml
      - name: Install dependencies
        working-directory: apps/web
        run: pnpm install --frozen-lockfile
      - name: Type check
        working-directory: apps/web
        run: pnpm exec tsc --noEmit

  frontend-test:
    name: Frontend Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          cache-dependency-path: apps/web/pnpm-lock.yaml
      - name: Install dependencies
        working-directory: apps/web
        run: pnpm install --frozen-lockfile
      - name: Test
        working-directory: apps/web
        run: pnpm run test

  frontend-build:
    name: Frontend Build
    runs-on: ubuntu-latest
    needs: [frontend-lint, frontend-typecheck, frontend-test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          cache-dependency-path: apps/web/pnpm-lock.yaml
      - name: Install dependencies
        working-directory: apps/web
        run: pnpm install --frozen-lockfile
      - name: Build
        working-directory: apps/web
        run: pnpm run build
        env:
          NEXT_PUBLIC_SUPABASE_URL: ${{ vars.NEXT_PUBLIC_SUPABASE_URL }}
          NEXT_PUBLIC_SUPABASE_ANON_KEY: ${{ vars.NEXT_PUBLIC_SUPABASE_ANON_KEY }}
          NEXT_PUBLIC_API_URL: ${{ vars.NEXT_PUBLIC_API_URL }}

  # ============================================
  # Protocol (TypeSpec)
  # ============================================
  protocol-check:
    name: Protocol Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - name: Install TypeSpec
        working-directory: packages/protocol
        run: pnpm install
      - name: Verify generated code is up to date
        working-directory: packages/protocol
        run: |
          pnpm run generate
          git diff --exit-code tsp-output/ || \
            (echo "Generated code is out of date. Run 'pnpm run generate' and commit." && exit 1)
```

### CI 파이프라인 흐름

```
PR 생성 / push to main
    │
    ├── backend-lint        (golangci-lint)        ─┐
    ├── backend-test        (go test -race)         │ 병렬 실행
    ├── frontend-lint       (eslint)                │
    ├── frontend-typecheck  (tsc --noEmit)          │
    ├── frontend-test       (vitest)                │
    └── protocol-check      (typespec 정합성)       ─┘
                                                     │
                                                     ▼
                                          frontend-build (next build)
                                          (lint, typecheck, test 통과 후)
```

---

## 6. 환경별 설정

### 환경 분리 전략

| 환경 | DB | Backend | Frontend | 용도 |
|------|---|---------|----------|------|
| **development** | 로컬 PostgreSQL (Docker) | `go run ./cmd/api` | `pnpm run dev` | 로컬 개발 |
| **staging** | Supabase (별도 프로젝트) | Koyeb (staging 브랜치) | Vercel Preview | PR 검증, QA |
| **production** | Supabase (운영) | Koyeb (main 브랜치) | Vercel Production | 실서비스 |

### Go 백엔드 환경 감지

```go
// internal/config/config.go

type Environment string

const (
    EnvDevelopment Environment = "development"
    EnvStaging     Environment = "staging"
    EnvProduction  Environment = "production"
)

type Config struct {
    Env          Environment
    Port         string
    DatabaseURL  string
    JWTSecret    string
    // AI API Keys
    AnthropicKey string
    OpenAIKey    string
    // ...
}

func Load() *Config {
    env := Environment(getEnv("APP_ENV", "development"))

    return &Config{
        Env:          env,
        Port:         getEnv("PORT", "8000"),
        DatabaseURL:  requireEnv("DATABASE_URL"),
        JWTSecret:    requireEnv("SUPABASE_JWT_SECRET"),
        AnthropicKey: getEnv("ANTHROPIC_API_KEY", ""),
        OpenAIKey:    getEnv("OPENAI_API_KEY", ""),
    }
}

func (c *Config) IsDevelopment() bool { return c.Env == EnvDevelopment }
func (c *Config) IsProduction() bool  { return c.Env == EnvProduction }
```

### 환경별 동작 차이

| 항목 | Development | Staging | Production |
|------|-------------|---------|------------|
| 로그 레벨 | DEBUG | INFO | INFO |
| 로그 포맷 | Text (읽기 쉬운) | JSON | JSON |
| CORS | `localhost:3000` | `*.vercel.app` | `colight.kr` |
| GIN_MODE | debug | release | release |
| AI 호출 | Mock 가능 | 실제 API | 실제 API |
| Rate Limit | 비활성화 | 완화 | 활성화 |

### 로컬 개발 환경 셋업

```bash
# 1. 로컬 PostgreSQL (Docker)
docker run -d \
  --name colight-db \
  -e POSTGRES_DB=colight_dev \
  -e POSTGRES_USER=colight \
  -e POSTGRES_PASSWORD=colight \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# 2. 마이그레이션 적용
cd apps/backend
atlas migrate apply --url "postgres://colight:colight@localhost:5432/colight_dev?sslmode=disable" --dir file://migrations

# 3. 시드 데이터
go run ./scripts/seed.go

# 4. 백엔드 서버 실행
go run ./cmd/api

# 5. 프론트엔드 서버 실행 (별도 터미널)
cd apps/web
pnpm run dev
```
