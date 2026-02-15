# Phase 0: 프로젝트 셋업 & 인프라

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 모노레포 초기화, Go 백엔드 + Ent ORM + Atlas 마이그레이션, TypeSpec 프로토콜, Next.js 프론트엔드, 로컬 개발 환경 구축 |
| **선행 조건** | 없음 (최초 Phase) |
| **스프린트** | Sprint 0 (Day 1-4) |
| **관련 기능** | F23 (프롬프트 DB), F24 (무기 카테고리), F25 (문항 패턴) |
| **예상 공수** | 4일 |
| **산출물** | 모노레포 (apps/backend + apps/web + packages/protocol), Go API 서버 기본 구조, Ent 스키마 15개, Atlas 마이그레이션, TypeSpec → OpenAPI → Go/TS 코드 생성 파이프라인, Next.js 프론트엔드, 로컬 PostgreSQL (Docker), 시드 데이터 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 0.1 | 모노레포 초기화 (pnpm + moon) | ✅ |
| 0.2 | Go 백엔드 프로젝트 | ✅ |
| 0.3 | Ent 초기화 + 전체 스키마 (15개) | ✅ |
| 0.4 | Atlas 설정 + 첫 마이그레이션 | ✅ |
| 0.5 | 로컬 PostgreSQL (Docker) | ✅ |
| 0.6 | TypeSpec 프로토콜 | ✅ |
| 0.7 | oapi-codegen (Go) + @hey-api (TS) | ✅ |
| 0.8 | Next.js 프로젝트 + ESLint 9 설정 | ✅ |
| 0.9 | 시드 데이터 삽입 | ✅ |
| 0.10 | 환경변수 확인 | ✅ |

> **Note**: Supabase 프로젝트 생성 및 Vercel/Koyeb 배포는 **Phase 6.2 (Landing & Beta)**로 이동. 로컬 개발 환경에서 모든 기능을 완성한 후 배포 설정을 진행합니다.

---

## Step 0.1: 모노레포 초기화

### 목표
pnpm 워크스페이스 기반 모노레포를 생성하고, 전체 디렉토리 구조를 스캐폴딩한다.

### 체크리스트

- [x] 프로젝트 루트 초기화
  - [x] `pnpm init` 실행
  - [x] `pnpm-workspace.yaml` 생성
  - [x] Node.js 20+ / pnpm 9+ 버전 확인
- [x] `pnpm-workspace.yaml` 작성
  ```yaml
  packages:
    - "apps/*"
    - "packages/*"
  ```
- [x] 전체 디렉토리 구조 생성
  - [x] `apps/backend/` (Go API 서버)
  - [x] `apps/web/` (Next.js 프론트엔드)
  - [x] `packages/protocol/` (TypeSpec API 정의)
  - [x] `docs/develop/phases/` (Phase 가이드)
  - [x] `supabase/` (참고용 마이그레이션)
- [x] moon (moonrepo) 설치 및 설정
  > moon은 Go + Node.js 폴리글랏 모노레포 태스크 러너. 루트에서 `moon run :lint` 등으로 전체 프로젝트 태스크를 통합 실행 가능.
  - [x] proto (moon 툴체인 매니저) 설치: `curl -fsSL https://moonrepo.dev/install/proto.sh | bash`
  - [x] moon 설치: `proto install moon`
  - [x] `.prototools` 생성 (프로젝트 루트)
    ```toml
    node = "20.11.0"
    pnpm = "9.15.0"
    go = "1.24.0"
    moon = "latest"
    ```
  - [x] `.moon/workspace.yml` 생성
    ```yaml
    projects:
      - "apps/*"
      - "packages/*"

    vcs:
      manager: git
      defaultBranch: main

    hasher:
      walkStrategy: glob
    ```
  - [x] `.moon/toolchain.yml` 생성
    ```yaml
    node:
      version: "20.11.0"
      packageManager: pnpm

    unstable_go:
      version: "1.24.0"
    ```
  - [x] `apps/backend/moon.yml` 생성
    ```yaml
    language: go
    type: application

    tasks:
      dev:
        command: air
        local: true
      build:
        command: go build -o bin/api ./cmd/api
      test:
        command: go test ./...
      lint:
        command: golangci-lint run ./...
      generate-ent:
        command: go generate ./ent
      generate-api:
        command: oapi-codegen -config oapi-codegen.yaml ../../packages/protocol/tsp-output/openapi/openapi.yaml
        deps:
          - "protocol:generate"
      migrate-diff:
        command: "atlas migrate diff ${name} --dir file://migrations --to ent://ent/schema --dev-url docker://postgres/16/dev?search_path=public"
      migrate-apply:
        command: "atlas migrate apply --dir file://migrations --url $DATABASE_URL"
      seed:
        command: go run ./scripts/seed.go all
    ```
  - [x] `apps/web/moon.yml` 생성
    ```yaml
    language: node
    type: application

    tasks:
      dev:
        command: pnpm run dev
        local: true
      build:
        command: pnpm run build
        deps:
          - "~:lint"
          - "~:typecheck"
      test:
        command: pnpm run test
      lint:
        command: pnpm run lint
      typecheck:
        command: pnpm exec tsc --noEmit
      generate-client:
        command: pnpm run generate:client
        deps:
          - "protocol:generate"
    ```
  - [x] `packages/protocol/moon.yml` 생성
    ```yaml
    language: node
    type: library

    tasks:
      generate:
        command: pnpm run generate
      validate:
        command: pnpm run validate
    ```
  - [x] `moon run :lint` → 전체 프로젝트 lint 실행 확인
  - [x] `moon run :test` → 전체 테스트 실행 확인
- [x] `.gitignore` 설정
  - [x] `node_modules/`
  - [x] `.env`, `.env.local`, `.env*.local`
  - [x] `.next/`
  - [x] `.vercel/`
  - [x] `supabase/.temp/`
  - [x] `apps/backend/tmp/` (air 핫 리로드)

### 검증 방법
- `pnpm install` → 정상 실행
- 디렉토리 구조가 아키텍처 문서와 일치
- `.gitignore`에 모든 민감 파일 포함

### 산출물
- 모노레포 루트 (`pnpm-workspace.yaml`, `package.json`)
- 전체 디렉토리 스캐폴딩
- `.gitignore` 보안 설정

---

## Step 0.2: Go 백엔드 프로젝트

### 목표
Go 모듈을 초기화하고, Gin + Ent + oapi-codegen 등 핵심 패키지를 설치하며, 표준 디렉토리 구조를 구축한다.

### 체크리스트

- [x] Go 모듈 초기화
  - [x] `cd apps/backend`
  - [x] `go mod init github.com/<username>/colight/apps/backend`
  - [x] Go 1.24+ 버전 확인
- [x] 핵심 패키지 설치
  - [x] HTTP 프레임워크: `go get github.com/gin-gonic/gin`
  - [x] ORM: `go get entgo.io/ent`
  - [x] 코드 생성: `go get github.com/oapi-codegen/oapi-codegen/v2`
  - [x] 환경 변수: `go get github.com/joho/godotenv`
  - [x] JWT: `go get github.com/golang-jwt/jwt/v5`
  - [x] UUID: `go get github.com/google/uuid`
  - [x] AI (Anthropic): `go get github.com/anthropics/anthropic-sdk-go`
  - [x] AI (경량 LLM 공통 인터페이스): Gemini SDK / Groq SDK (`LLM_LIGHT_PROVIDER` 환경변수로 선택)
  - [x] pgvector: `go get github.com/pgvector/pgvector-go`
  - [x] 로깅: `log/slog` (표준 라이브러리)
- [x] 디렉토리 구조 생성
  ```text
  apps/backend/
  ├── cmd/
  │   └── api/main.go              # 엔트리포인트
  ├── internal/
  │   ├── controller/              # HTTP 핸들러
  │   ├── service/                 # 비즈니스 로직
  │   ├── infrastructure/
  │   │   ├── config/              # 환경 설정
  │   │   ├── database/            # DB 연결
  │   │   ├── middleware/          # JWT, CORS, 로깅
  │   │   ├── ai/                  # Claude/OpenAI 클라이언트
  │   │   └── external/            # DART, 네이버, Koyeb
  │   └── generated/               # oapi-codegen 생성 코드
  ├── ent/
  │   └── schema/                  # Ent 스키마 정의
  ├── migrations/                  # Atlas SQL 마이그레이션
  ├── scripts/
  │   └── seed.go                  # 시드 데이터 스크립트
  ├── Dockerfile
  ├── go.mod
  └── go.sum
  ```
- [x] `cmd/api/main.go` 기본 엔트리포인트 작성
  ```go
  package main

  import (
      "log/slog"
      "os"

      "github.com/gin-gonic/gin"
  )

  func main() {
      slog.Info("starting colight api server")

      r := gin.Default()

      r.GET("/health", func(c *gin.Context) {
          c.JSON(200, gin.H{"status": "ok"})
      })

      port := os.Getenv("API_PORT")
      if port == "" {
          port = "8080"
      }

      if err := r.Run(":" + port); err != nil {
          slog.Error("failed to start server", "error", err)
          os.Exit(1)
      }
  }
  ```
- [x] `Dockerfile` 작성
  ```dockerfile
  FROM golang:1.24-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 go build -o /api ./cmd/api

  FROM alpine:3.19
  RUN apk add --no-cache ca-certificates
  COPY --from=builder /api /api
  EXPOSE 8080
  CMD ["/api"]
  ```

### 검증 방법
- `cd apps/backend && go run ./cmd/api` → 서버 시작
- `curl http://localhost:9000/health` → `{"status":"ok"}`
- `go build ./...` → 빌드 성공
- `go vet ./...` → 경고 없음

### 산출물
- Go 프로젝트 (`go.mod`, 핵심 패키지)
- 표준 디렉토리 구조
- 기본 Gin 서버 (`cmd/api/main.go`)
- `Dockerfile`

---

## Step 0.3: Ent 초기화 + 전체 스키마

### 목표
Ent ORM을 초기화하고, 15개 DB 스키마를 정의한다.

> **⚠️ 스키마 정본(Single Source of Truth)**: Ent 스키마의 최종 정의는 `docs/develop/02-data-structure.md`입니다. 이 Phase 문서의 스키마 코드와 차이가 있을 경우, `02-data-structure.md`를 따르세요.

### 체크리스트

- [x] Ent 초기화
  - [x] `cd apps/backend`
  - [x] `go run -mod=mod entgo.io/ent/cmd/ent init --target ent/schema UserProfile Experience ExperienceTag ExperienceWeapon ExperienceUsage WeaponCategory PromptTemplate QuestionPattern Application CompanyAnalysis CompanyAnalysisCache TalentProfile CoverLetter CoverLetterVersion CoachingSession`
- [x] `ent/generate.go` 작성
  ```go
  package ent

  //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
  ```
- [x] UUID Mixin 작성
  ```go
  // ent/schema/mixin.go
  package schema

  import (
      "entgo.io/ent"
      "entgo.io/ent/schema/field"
      "entgo.io/ent/schema/mixin"
      "github.com/google/uuid"
  )

  type UUIDMixin struct {
      mixin.Schema
  }

  func (UUIDMixin) Fields() []ent.Field {
      return []ent.Field{
          field.UUID("id", uuid.UUID{}).Default(uuid.New),
      }
  }
  ```
- [x] 각 스키마 정의 (15개)

#### UserProfile 스키마

```go
// ent/schema/userprofile.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "entgo.io/ent/schema/mixin"
)

type UserProfile struct {
    ent.Schema
}

func (UserProfile) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.Time{},
        UUIDMixin{},
    }
}

func (UserProfile) Fields() []ent.Field {
    return []ent.Field{
        field.String("user_id").Unique().NotEmpty(),
        field.String("nickname").Optional().Nillable(),
        field.String("university").Optional().Nillable(),
        field.String("major").Optional().Nillable(),
        field.Int("graduation_year").Optional().Nillable(),
        field.String("target_industry").Optional().Nillable(),
        field.String("target_job").Optional().Nillable(),
        field.Enum("experience_level").
            Values("신입", "인턴", "경력").
            Optional().Nillable(),
    }
}

func (UserProfile) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id").Unique(),
    }
}

func (UserProfile) Edges() []ent.Edge {
    return nil
}
```

#### Experience 스키마

```go
// ent/schema/experience.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "entgo.io/ent/schema/mixin"
)

type Experience struct {
    ent.Schema
}

func (Experience) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.Time{},
        UUIDMixin{},
    }
}

func (Experience) Fields() []ent.Field {
    return []ent.Field{
        field.String("user_id").NotEmpty(),
        field.String("title").NotEmpty(),
        field.Time("period_start").Optional().Nillable(),
        field.Time("period_end").Optional().Nillable(),
        field.String("role").Optional().Nillable(),
        field.String("category").Optional().Nillable(),
        field.Text("situation").Optional().Nillable(),
        field.Text("task").Optional().Nillable(),
        field.Text("action").Optional().Nillable(),
        field.Text("result").Optional().Nillable(),
        field.Text("raw_content").Optional().Nillable(),
        field.Bool("is_starred").Default(false),
    }
}

func (Experience) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id"),
    }
}

func (Experience) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("tags", ExperienceTag.Type),
        edge.To("weapons", ExperienceWeapon.Type),
        edge.To("usages", ExperienceUsage.Type),
    }
}
```

#### WeaponCategory 스키마

```go
// ent/schema/weaponcategory.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "entgo.io/ent/schema/mixin"
)

type WeaponCategory struct {
    ent.Schema
}

func (WeaponCategory) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.Time{},
        UUIDMixin{},
    }
}

func (WeaponCategory) Fields() []ent.Field {
    return []ent.Field{
        field.String("code").Unique().NotEmpty(),
        field.String("parent_code").Optional().Nillable(),
        field.String("name").NotEmpty(),
        field.Text("description").Optional().Nillable(),
        field.JSON("keywords", []string{}).Default([]string{}),
        field.JSON("question_patterns", []string{}).Default([]string{}),
        field.Int("display_order").Default(0),
        field.String("icon").Optional().Nillable(),
        field.String("color").Optional().Nillable(),
    }
}

func (WeaponCategory) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("code").Unique(),
        index.Fields("parent_code"),
    }
}

func (WeaponCategory) Edges() []ent.Edge {
    return nil
}
```

- [x] 나머지 12개 스키마 작성
  - [x] `ExperienceTag` — experience_id FK, tag_name, tag_category, confidence
  - [x] `ExperienceWeapon` — experience_id FK, weapon_code FK, confidence, is_primary, reasoning, user_confirmed, user_modified
  - [x] `ExperienceUsage` — experience_id FK, cover_letter_id FK, application_id FK, used_at
  - [x] `PromptTemplate` — category, sub_category, name, description, system_prompt, user_prompt_template, output_schema JSONB, model, temperature, max_tokens, version, is_active, usage_count, avg_latency_ms, avg_quality_score
  - [x] `QuestionPattern` — pattern_type, pattern_name, detection_keywords JSON, detection_regex, primary_weapons JSON, secondary_weapons JSON, writing_guide JSONB, coaching_prompt_id FK
  - [x] `Application` — user_id, company_name, job_title, job_url, status enum(preparing/writing/submitted/interview/accepted/rejected), deadline, notes
  - [x] `CompanyAnalysis` — user_id, application_id FK, company_name, job_url, job_posting_parsed JSONB, company_profile JSONB, talent_analysis JSONB, news_summary JSONB, strategy_keywords JSON, matching_score
  - [x] `CompanyAnalysisCache` — company_name, analysis_type, data JSONB, source, expires_at (unique: company_name + analysis_type)
  - [x] `TalentProfile` — company_name, industry, core_values JSON, talent_keywords JSON, culture_keywords JSON, interview_topics JSON, source_urls JSON, last_updated
  - [x] `CoverLetter` — user_id, application_id FK, company_name, question_text, question_type, char_limit, current_content, status enum(draft/coaching/reviewing/final), ai_detection_score
  - [x] `CoverLetterVersion` — cover_letter_id FK, version_number, content, change_summary, coaching_feedback JSONB, scores JSONB
  - [x] `CoachingSession` — user_id, cover_letter_id FK, session_type enum(question_analysis/draft_coaching/review_coaching/weapon_enhance), messages JSONB, prompt_template_id FK, model_used, total_tokens
- [x] `go generate ./ent` 실행 → 생성 코드 확인
- [x] `go build ./...` → 빌드 성공

### 검증 방법
- `go generate ./ent` → 에러 없음
- `ls ent/` → 생성된 파일 확인 (client.go, mutation.go, ...)
- `go build ./...` → 빌드 성공
- 15개 스키마 파일이 `ent/schema/`에 존재

### 산출물
- `ent/schema/` 15개 스키마 파일 + mixin.go
- `ent/generate.go`
- Ent 생성 코드 (`ent/client.go`, `ent/*.go`)

---

## Step 0.4: Atlas 설정 + 첫 마이그레이션

### 목표
Atlas CLI로 Ent 스키마에서 SQL 마이그레이션을 자동 생성하고, 로컬 DB에 적용한다.

### 체크리스트

- [x] Atlas CLI 설치
  - [x] `curl -sSf https://atlasgo.sh | sh` 또는 `brew install ariga/tap/atlas`
  - [x] `atlas version` 확인
- [x] `atlas.hcl` 설정 파일 생성
  ```hcl
  // apps/backend/atlas.hcl

  data "composite_schema" "app" {
    schema "public" {
      url = "ent://ent/schema"
    }
  }

  env "local" {
    src = data.composite_schema.app.url
    dev = "docker://postgres/16/dev?search_path=public"

    migration {
      dir = "file://migrations"
    }

    format {
      migrate {
        diff = "{{ sql . \"  \" }}"
      }
    }
  }

  env "prod" {
    src = data.composite_schema.app.url
    url = getenv("DATABASE_URL")

    migration {
      dir = "file://migrations"
    }
  }
  ```
- [x] 첫 마이그레이션 생성
  - [x] `cd apps/backend`
  - [x] `moon run backend:migrate-diff -- name=initial_schema` (or direct: `atlas migrate diff initial_schema --dir file://migrations --to ent://ent/schema --dev-url "docker://postgres/16/dev?search_path=public"`)
  - [x] `migrations/` 디렉토리에 SQL 파일 생성 확인
- [x] pgvector 확장 활성화 SQL 추가
  - [x] 마이그레이션 파일 상단에 `CREATE EXTENSION IF NOT EXISTS vector;` 추가
  - [x] `CREATE EXTENSION IF NOT EXISTS pg_trgm;` 추가
- [x] experiences 테이블에 벡터 컬럼 수동 추가 (Ent에서 직접 지원하지 않는 경우)
  - [x] `ALTER TABLE experiences ADD COLUMN embedding vector(1536);`
  - [x] `CREATE INDEX idx_experiences_embedding ON experiences USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);`
- [x] 생성된 SQL 파일 검토
  - [x] 15개 테이블 CREATE TABLE 확인
  - [x] FK 관계 확인
  - [x] 인덱스 확인

### 검증 방법
- `ls migrations/` → SQL 파일 존재
- SQL 파일 내용에 15개 테이블 DDL 포함
- `atlas migrate validate --dir file://migrations` → 유효

### 산출물
- `atlas.hcl` (Atlas 설정)
- `migrations/YYYYMMDDHHMMSS_initial_schema.sql` (첫 마이그레이션)
- pgvector 확장 + 벡터 인덱스 포함

---

## Step 0.5: 로컬 PostgreSQL

### 목표
Docker Compose로 로컬 PostgreSQL 16을 실행하고, Atlas 마이그레이션을 적용한다.

### 체크리스트

- [x] `docker-compose.yml` 생성 (프로젝트 루트)
  ```yaml
  services:
    postgres:
      image: pgvector/pgvector:pg16
      ports:
        - "5532:5432"
      environment:
        POSTGRES_DB: colight
        POSTGRES_USER: postgres
        POSTGRES_PASSWORD: password
      volumes:
        - pgdata:/var/lib/postgresql/data

  volumes:
    pgdata:
  ```
- [x] Docker Compose 실행
  - [x] `docker compose up -d`
  - [x] `docker compose ps` → postgres 실행 확인
- [x] 마이그레이션 적용
  - [x] `cd apps/backend`
  - [x] `DATABASE_URL="postgres://postgres:password@localhost:5532/colight?sslmode=disable" moon run backend:migrate-apply`
  - [x] 15개 테이블 생성 확인
- [x] DB 연결 테스트
  - [x] `psql -h localhost -U postgres -d colight -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"`
  - [x] 결과: 15개 이상

### 검증 방법
- `docker compose ps` → postgres 실행 중
- `psql` 접속 정상
- 15개 테이블 존재 확인
- pgvector 확장: `SELECT extname FROM pg_extension WHERE extname = 'vector';` → 결과 있음

### 산출물
- `docker-compose.yml` (로컬 PostgreSQL)
- 로컬 DB에 15개 테이블 생성 완료

---

## Step 0.6: TypeSpec 프로토콜

### 목표
TypeSpec으로 API 정의를 작성하고, OpenAPI 스펙을 생성한다.

### 체크리스트

- [x] TypeSpec 초기화
  - [x] `cd packages/protocol`
  - [x] `pnpm init`
  - [x] `pnpm add -D @typespec/compiler @typespec/http @typespec/rest @typespec/openapi3`
  - [x] `tspconfig.yaml` 작성
  ```yaml
  emit:
    - "@typespec/openapi3"
  options:
    "@typespec/openapi3":
      output-dir: "{project-root}/tsp-output/openapi"
      emitter-output-dir: "{output-dir}"
  ```
- [x] API 정의 작성
  - [x] `src/main.tsp` (엔트리포인트, 서버 URL, 공통 모델)
  - [x] `src/models/` (공통 타입: ErrorResponse, PaginatedResponse 등)
  - [x] `src/experiences.tsp` (경험 CRUD API — Phase 2에서 상세화, 지금은 기본 구조만)
- [x] `src/main.tsp` 기본 구조
  ```typespec
  import "@typespec/http";
  import "@typespec/rest";
  import "@typespec/openapi3";

  using Http;
  using Rest;

  @service({
    title: "Colight API",
    version: "1.0.0",
  })
  @server("http://localhost:9000", "Local development")
  @server("https://colight-api.koyeb.app", "Production")
  namespace Colight;

  // Health check
  @route("/health")
  @tag("System")
  interface Health {
    @get op check(): { status: string };
  }
  ```
- [x] OpenAPI 생성
  - [x] `pnpm run generate:protocol` (tsp compile)
  - [x] `tsp-output/openapi/openapi.yaml` 파일 생성 확인

### 검증 방법
- `tsp compile .` → 에러 없음
- `tsp-output/openapi/openapi.yaml` 파일 존재
- OpenAPI 스펙에 `/health` 엔드포인트 포함

### 산출물
- `packages/protocol/src/main.tsp` (API 정의)
- `packages/protocol/tspconfig.yaml`
- `packages/protocol/tsp-output/openapi/openapi.yaml` (생성된 OpenAPI)

---

## Step 0.7: oapi-codegen (Go) + @hey-api (TS)

### 목표
OpenAPI 스펙에서 Go 서버 코드와 TypeScript 클라이언트 코드를 자동 생성한다.

### 체크리스트

- [x] oapi-codegen 설정 (Go)
  - [x] `cd apps/backend`
  - [x] `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`
  - [x] `oapi-codegen.yaml` 작성
  ```yaml
  package: generated
  output: internal/generated/api.gen.go
  generate:
    gin-server: true
    strict-server: true
    models: true
    embedded-spec: true
  ```
  - [x] `moon run backend:generate-api` 실행
  - [x] `internal/generated/api.gen.go` 파일 생성 확인
  - [x] `StrictServerInterface` 인터페이스 확인
- [x] @hey-api/openapi-ts 설정 (TS)
  - [x] `cd apps/web`
  - [x] `pnpm add -D @hey-api/openapi-ts`
  - [x] `openapi-ts.config.ts` 작성
  ```typescript
  import { defineConfig } from "@hey-api/openapi-ts";

  export default defineConfig({
    client: "axios",
    input: "../../packages/protocol/tsp-output/openapi/openapi.yaml",
    output: {
      path: "src/api/generated",
      format: "prettier",
    },
    plugins: [
      "@hey-api/typescript",
      "@hey-api/sdk",
      {
        name: "@hey-api/zod",
        output: "zod.gen",
      },
    ],
  });
  ```
  - [x] `package.json`에 생성 스크립트 추가: `"generate": "openapi-ts"`
  - [x] `pnpm run generate` 실행
  - [x] `src/api/generated/` 디렉토리에 파일 생성 확인
    - [x] `types.gen.ts`
    - [x] `sdk.gen.ts`
    - [x] `zod.gen.ts`
- [x] 생성된 코드 Git 커밋 (코드 리뷰를 위해 커밋)

### 검증 방법
- Go: `go build ./...` → 빌드 성공
- Go: `internal/generated/api.gen.go` 존재, `StrictServerInterface` 인터페이스 포함
- TS: `src/api/generated/types.gen.ts` 존재
- TS: `src/api/generated/sdk.gen.ts` 존재

### 산출물
- `apps/backend/oapi-codegen.yaml` (Go 코드 생성 설정)
- `apps/backend/internal/generated/api.gen.go` (생성된 Go 코드)
- `apps/web/openapi-ts.config.ts` (TS 코드 생성 설정)
- `apps/web/src/api/generated/` (생성된 TS 코드)

---

## Step 0.8: Next.js 프로젝트

### 목표
Next.js 15 프론트엔드 프로젝트를 생성하고, Go 백엔드 API를 호출하는 Axios 클라이언트를 설정한다.

### 체크리스트

- [x] Next.js 프로젝트 생성
  - [x] `cd apps`
  - [x] `npx create-next-app@latest web --typescript --tailwind --eslint --app --src-dir --import-alias "@/*"`
  - [x] `cd web && pnpm install`
  - [x] `pnpm run dev` 정상 동작 확인
- [x] 핵심 의존성 설치
  - [x] Supabase: `pnpm add @supabase/supabase-js @supabase/ssr`
  - [x] HTTP: `pnpm add axios`
  - [x] 상태 관리: `pnpm add zustand @tanstack/react-query`
  - [x] 유효성 검증: `pnpm add zod`
  - [x] 폼: `pnpm add react-hook-form @hookform/resolvers`
- [x] shadcn/ui 초기화
  - [x] `npx shadcn@latest init`
  - [x] 기본 컴포넌트 설치: `npx shadcn@latest add button card input label textarea select dialog sheet dropdown-menu tabs badge separator avatar skeleton toast sonner`
- [x] API 클라이언트 설정
  - [x] `src/lib/api/client.ts` 작성 (Axios 인스턴스)
  ```typescript
  import axios from "axios";

  export const apiClient = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000",
    headers: {
      "Content-Type": "application/json",
    },
  });

  // Supabase JWT 토큰을 자동으로 헤더에 추가
  apiClient.interceptors.request.use(async (config) => {
    // Phase 1에서 Supabase Auth 연동 시 구현
    return config;
  });
  ```
- [x] App Router 페이지 구조 생성 (빈 페이지)
  - [x] `src/app/(auth)/login/page.tsx`
  - [x] `src/app/(auth)/signup/page.tsx`
  - [x] `src/app/(auth)/layout.tsx`
  - [x] `src/app/(main)/dashboard/page.tsx`
  - [x] `src/app/(main)/experiences/page.tsx`
  - [x] `src/app/(main)/analysis/page.tsx`
  - [x] `src/app/(main)/coaching/page.tsx`
  - [x] `src/app/(main)/layout.tsx`
  - [x] `src/app/(main)/settings/page.tsx`
- [x] **API Routes 없음** — `src/app/api/` 디렉토리 생성하지 않음 (Go 백엔드로 이관)
- [x] 환경변수 설정
  - [x] `NEXT_PUBLIC_API_URL` — Go API 서버 URL
  - [x] `NEXT_PUBLIC_SUPABASE_URL` — Supabase 프로젝트 URL
  - [x] `NEXT_PUBLIC_SUPABASE_ANON_KEY` — Supabase Anon Key

### 검증 방법
- `pnpm run dev` → `http://localhost:4000` 정상 접속
- `pnpm run build` → 빌드 에러 없음
- API 클라이언트: `apiClient.get('/health')` → Go 서버 응답 확인

### 산출물
- Next.js 15 프로젝트 (`apps/web/`)
- shadcn/ui 컴포넌트 설치
- Axios API 클라이언트 (`src/lib/api/client.ts`)
- App Router 페이지 스캐폴딩 (빈 페이지)

---

## Step 0.9: 시드 데이터 삽입

### 목표
무기 카테고리 7대분류 + 28소분류, 프롬프트 템플릿 4종, 문항 패턴 7종의 초기 데이터를 삽입한다.

### 체크리스트

- [x] `scripts/seed.go` 작성
  - [x] `all` 명령어: 전체 시드 실행
  - [x] `weapons` 명령어: weapon_categories 시드
  - [x] `prompts` 명령어: prompt_templates 시드
  - [x] `patterns` 명령어: question_patterns 시드
- [x] weapon_categories 삽입 (35건: 대분류 7 + 소분류 28)
  - [x] W01~W07 대분류 7건 (code, name, description, keywords, question_patterns, display_order, icon, color)
  - [x] W01-A~W01-D, W02-A~W02-D, ..., W07-A~W07-D 소분류 28건 (parent_code 연결)
- [x] prompt_templates 삽입 (4건)
  - [x] 경험 무기 자동 분류 (`experience_classify` / `weapon_tagging`, model: `gemini-2.0-flash`)
  - [x] 경험 인터뷰 (`experience_classify` / `interview`, model: `gemini-2.0-flash`)
  - [x] 자소서 문항 분석 (`coaching_draft` / `question_analysis`, model: `claude-sonnet-4-5`)
  - [x] 무기별 경험 강화 코칭 (`coaching_draft` / `weapon_enhance`, model: `claude-sonnet-4-5`)
- [x] question_patterns 삽입 (7건)
  - [x] 지원동기, 장단점, 위기극복, 리더십, 팀워크, 목표달성, 성장과정
- [x] 시드 실행
  - [x] `moon run backend:seed` (or direct: `cd apps/backend && go run ./scripts/seed.go all`)
  - [x] 데이터 건수 확인

### 시드 데이터 상세

> 시드 데이터의 상세 내용 (weapon_categories 35건, prompt_templates 4건, question_patterns 7건)은 Go 코드에서 Ent 클라이언트를 사용하여 삽입. 각 항목의 상세 값은 아래 참고:

<details>
<summary>weapon_categories 대분류 7건 상세</summary>

| code | name | icon | color | keywords (일부) |
|------|------|------|-------|----------------|
| W01 | 위기극복 | 🛡️ | #EF4444 | 실패, 좌절, 극복, 위기, 역경 |
| W02 | 리더십 | 👑 | #F59E0B | 리더, 팀장, 주도, 이끌, 방향 |
| W03 | 팀워크/협업 | 🤝 | #10B981 | 협업, 팀워크, 갈등, 조율, 소통 |
| W04 | 도전정신 | 🚀 | #8B5CF6 | 도전, 목표, 달성, 시도, 새로운 |
| W05 | 문제해결 | 🔧 | #3B82F6 | 문제, 해결, 분석, 원인, 개선 |
| W06 | 소통/설득 | 💬 | #EC4899 | 소통, 설득, 협상, 발표, 경청 |
| W07 | 성장/학습 | 📈 | #6366F1 | 성장, 학습, 배움, 자격증, 전문 |

</details>

<details>
<summary>prompt_templates 4건 상세</summary>

| category | sub_category | name | model |
|----------|-------------|------|-------|
| experience_classify | weapon_tagging | 경험 무기 자동 분류 | gemini-2.0-flash |
| experience_classify | interview | 경험 인터뷰 (대화형) | gemini-2.0-flash |
| coaching_draft | question_analysis | 자소서 문항 분석 | claude-sonnet-4-5 |
| coaching_draft | weapon_enhance | 무기별 경험 강화 코칭 | claude-sonnet-4-5 |

</details>

<details>
<summary>question_patterns 7건 상세</summary>

| pattern_type | pattern_name | primary_weapons | secondary_weapons |
|-------------|-------------|----------------|-------------------|
| growth | 성장과정/자기소개 | W07 | W01, W04 |
| crisis | 위기극복/실패경험 | W01 | W05, W04 |
| leadership | 리더십/주도적 경험 | W02 | W03, W06 |
| teamwork | 팀워크/협업/갈등해결 | W03 | W06, W02 |
| challenge | 도전/목표달성 | W04 | W05, W07 |
| problem_solving | 문제해결/창의성 | W05 | W04, W07 |
| motivation | 지원동기/입사 후 포부 | W07, W04 | W05 |

</details>

### 검증 방법
- `SELECT COUNT(*) FROM weapon_categories;` → 35
- `SELECT COUNT(*) FROM weapon_categories WHERE parent_code IS NULL;` → 7
- `SELECT COUNT(*) FROM weapon_categories WHERE parent_code IS NOT NULL;` → 28
- `SELECT COUNT(*) FROM prompt_templates;` → 4
- `SELECT COUNT(*) FROM question_patterns;` → 7

### 산출물
- `scripts/seed.go` (시드 스크립트)
- weapon_categories 35건, prompt_templates 4건, question_patterns 7건 삽입 완료

---

## Step 0.10: 환경변수 및 보안 설정

### 목표
로컬/프로덕션 환경변수를 설정하고, 민감 정보가 Git에 포함되지 않도록 보안을 확보한다.

### 체크리스트

- [x] 루트 `.env` 파일 생성 (`.gitignore`에 포함)
  ```env
  # ── Backend (Go) ──
  API_PORT=9000
  DATABASE_URL=postgres://postgres:password@localhost:5532/colight?sslmode=disable
  SUPABASE_JWT_SECRET=<supabase-jwt-secret>
  ANTHROPIC_API_KEY=<anthropic-key>
  OPENAI_API_KEY=<openai-key>
  DART_API_KEY=<dart-key>
  NAVER_CLIENT_ID=<naver-id>
  NAVER_CLIENT_SECRET=<naver-secret>

  # ── Frontend (Next.js) ──
  NEXT_PUBLIC_API_URL=http://localhost:9000
  NEXT_PUBLIC_SUPABASE_URL=https://<project-ref>.supabase.co
  NEXT_PUBLIC_SUPABASE_ANON_KEY=<anon-key>
  ```
- [x] `.env.example` 파일 생성 (Git 커밋)
  ```env
  # ── Backend (Go) ──
  API_PORT=9000
  DATABASE_URL=postgres://postgres:password@localhost:5532/colight?sslmode=disable
  SUPABASE_JWT_SECRET=your-supabase-jwt-secret
  ANTHROPIC_API_KEY=your-anthropic-api-key
  OPENAI_API_KEY=your-openai-api-key
  DART_API_KEY=your-dart-api-key
  NAVER_CLIENT_ID=your-naver-client-id
  NAVER_CLIENT_SECRET=your-naver-client-secret

  # ── Frontend (Next.js) ──
  NEXT_PUBLIC_API_URL=http://localhost:9000
  NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
  NEXT_PUBLIC_SUPABASE_ANON_KEY=your-supabase-anon-key
  ```
- [x] Go 백엔드 환경변수 로드 설정
  - [x] `internal/infrastructure/config/config.go` 작성
  - [x] `.env` 파일 로드 (godotenv, 루트 `.env` 경로: `../../.env`)
- [x] Next.js 환경변수 로드 설정
  - [x] `next.config.ts`에서 dotenv로 루트 `.env` 로드

### 검증 방법
- `cd apps/backend && go run ./cmd/api` → 환경변수 정상 로드
- `cd apps/web && pnpm run dev` → 환경변수 정상 로드
- `.env`가 `git status`에 표시되지 않음
- `.env.example`이 `git status`에 표시됨

### 산출물
- `.env` (로컬 환경변수 — Git 미포함)
- `.env.example` (환경변수 템플릿 — Git 포함)
- `apps/backend/internal/infrastructure/config/config.go`

---

---

> **⚠️ Supabase 프로젝트 생성 및 Vercel/Koyeb 배포는 Phase 6.2 (Landing & Beta)로 이동**
> 로컬 개발 환경에서 모든 기능을 완성한 후 배포 설정을 진행합니다.

---

## Phase 완료 체크리스트

### 인프라
- [x] 모노레포 구조 (apps/backend, apps/web, packages/protocol)
- [x] moon 태스크 러너 설정 (pnpm + Go 통합)
- [x] Go 백엔드 기본 서버 동작 (`/health` 포트 9000)
- [x] Ent ORM 15개 스키마 정의 + 코드 생성
- [x] Atlas 마이그레이션 생성 + 적용
- [x] 로컬 PostgreSQL (Docker Compose, 포트 5532) 동작
- [x] TypeSpec → OpenAPI → Go/TS 코드 생성 파이프라인
- [x] Next.js 16 프론트엔드 동작 (포트 4000)
- [x] ESLint 9 flat config 설정
- [x] 시드 데이터 삽입 완료

### 데이터
- [x] 15개 DB 테이블 생성 (로컬)
- [x] pgvector 확장 활성화
- [x] weapon_categories 35건 시드
- [x] prompt_templates 4건 시드
- [x] question_patterns 7건 시드

### 코드 품질
- [x] `moon run backend:build` → 빌드 성공
- [x] `moon run web:build` → 빌드 성공
- [x] `moon run web:lint` → lint 통과
- [x] `moon run web:typecheck` → 타입 체크 통과
- [x] `.env`가 Git에 포함되지 않음
- [x] 생성된 코드 (oapi-codegen, @hey-api) 커밋 준비됨

---

## 다음 Phase

**→ Phase 1: 인증 & 레이아웃** (Sprint 0, Day 5-6)
- Supabase Auth (프론트엔드) + JWT 검증 (Go 미들웨어) 연동
- 공통 사이드바/헤더 레이아웃 구축
- 보호된 라우트 미들웨어
