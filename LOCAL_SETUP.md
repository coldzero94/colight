# 로컬 개발 환경 설정 가이드

Colight를 로컬에서 실행하는 방법입니다.

## 사전 준비

필수 설치:
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) ✅ (설치됨)
- [Node.js 20+](https://nodejs.org/) ✅
- [pnpm](https://pnpm.io/) ✅
- [Moon CLI](https://moonrepo.dev/) ✅
- [Go 1.24+](https://go.dev/) ✅

## 빠른 시작 (5분)

### 1. 환경변수 설정

`.env` 파일에서 최소한 다음 값을 설정하세요:

```bash
# JWT Secret (32자 이상 아무 문자열)
JWT_SECRET=local-development-secret-key-32chars-minimum

# 나머지는 기본값 사용 가능
```

**선택적 (특정 기능 테스트 시 필요)**:
```bash
# AI 기능 테스트 시
ANTHROPIC_API_KEY=sk-ant-...  # 코칭 기능
OPENAI_API_KEY=sk-...          # 임베딩/매칭 기능
GEMINI_API_KEY=...             # 경량 AI 기능

# Naver 로그인 테스트 시
NAVER_CLIENT_ID=...
NAVER_CLIENT_SECRET=...
```

### 2. PostgreSQL 시작

```bash
# Docker Compose로 PostgreSQL 시작 (pgvector 포함)
docker-compose up -d

# 확인
docker-compose ps
# postgres가 "Up" 상태여야 함
```

### 3. 데이터베이스 마이그레이션

```bash
# 스키마 적용
moon run backend:migrate-apply

# 초기 데이터 시딩 (어드민 계정, 무기 카테고리 등)
moon run backend:seed
```

### 4. 개발 서버 실행

```bash
# 백엔드 + 프론트엔드 동시 실행
moon run :dev
```

실행 후:
- **Frontend**: http://localhost:4000
- **Backend**: http://localhost:9000
- **API Docs**: http://localhost:9000/swagger (있다면)

### 5. 초기 로그인

시딩된 어드민 계정:
```
Email: admin@colight.kr
Password: admin123
```

---

## 세부 명령어

### PostgreSQL 관리

```bash
# 시작
docker-compose up -d

# 중지
docker-compose down

# 로그 확인
docker-compose logs -f postgres

# 데이터베이스 초기화 (주의: 모든 데이터 삭제)
docker-compose down -v
docker-compose up -d
moon run backend:migrate-apply
moon run backend:seed
```

### 백엔드 (Go)

```bash
# 개발 모드 (hot reload)
moon run backend:dev

# 테스트
moon run backend:test

# 린팅
moon run backend:lint

# Ent 코드 생성 (스키마 변경 시)
moon run backend:generate-ent

# API 코드 생성 (TypeSpec 변경 시)
moon run backend:generate-api
```

### 프론트엔드 (Next.js)

```bash
# 개발 모드
moon run web:dev

# 테스트
moon run web:test

# 타입 체크
moon run web:typecheck

# 린팅
moon run web:lint

# 빌드
moon run web:build

# API 클라이언트 생성 (TypeSpec 변경 시)
moon run web:generate-client
```

### 전체 프로젝트

```bash
# 모든 개발 서버 실행
moon run :dev

# 모든 테스트
moon run :test

# 모든 린팅
moon run :lint

# 전체 빌드
moon run :build
```

---

## 트러블슈팅

### PostgreSQL 연결 실패

```bash
# PostgreSQL이 실행 중인지 확인
docker-compose ps

# 로그 확인
docker-compose logs postgres

# 재시작
docker-compose restart postgres
```

### 포트 충돌

다른 서비스가 포트를 사용 중일 경우:

```bash
# 5532 포트 (PostgreSQL)
lsof -ti:5532 | xargs kill -9

# 9000 포트 (백엔드)
lsof -ti:9000 | xargs kill -9

# 4000 포트 (프론트엔드)
lsof -ti:4000 | xargs kill -9
```

### 마이그레이션 에러

```bash
# 마이그레이션 상태 확인
atlas migrate status --env local

# 데이터베이스 초기화 후 재시도
docker-compose down -v
docker-compose up -d
moon run backend:migrate-apply
```

### pnpm 캐시 문제

```bash
# pnpm 캐시 삭제
pnpm store prune

# node_modules 재설치
rm -rf node_modules apps/*/node_modules packages/*/node_modules
pnpm install
```

### Moon 캐시 문제

```bash
# Moon 캐시 삭제
rm -rf .moon/cache

# 재실행
moon run :dev
```

---

## 개발 워크플로우

### 새 기능 개발

1. Phase 문서 읽기 (`docs/develop/phases/phase-X.md`)
2. 브랜치 생성 (선택적)
3. **TDD Red**: 테스트 작성 → 실행 → 실패 확인
4. **TDD Green**: 구현 → 테스트 → 통과 확인
5. **Gate**: 린팅 + 테스트 + 빌드 통과
6. **Commit**: `Phase X.Y: description`

### 스키마 변경

```bash
# 1. Ent 스키마 수정 (apps/backend/ent/schema/*.go)
# 2. Ent 코드 생성
moon run backend:generate-ent

# 3. 마이그레이션 생성
moon run backend:migrate-diff -- name=add_new_field

# 4. 마이그레이션 적용
moon run backend:migrate-apply

# 5. 테스트
moon run backend:test
```

### API 변경

```bash
# 1. TypeSpec 수정 (packages/protocol/main.tsp)
# 2. OpenAPI 생성
moon run protocol:generate

# 3. Go 서버 코드 생성
moon run backend:generate-api

# 4. TS 클라이언트 생성
moon run web:generate-client

# 5. 구현 및 테스트
moon run :test
```

---

## 환경변수 상세

### 필수

| 변수 | 설명 | 로컬 기본값 |
|------|------|-------------|
| `DATABASE_URL` | PostgreSQL 연결 문자열 | `postgres://postgres:password@localhost:5532/colight?sslmode=disable` |
| `JWT_SECRET` | JWT 토큰 서명 키 (32자 이상) | 설정 필요 |

### 선택적

| 변수 | 설명 | 필요한 기능 |
|------|------|-------------|
| `ANTHROPIC_API_KEY` | Claude API 키 | 코칭 기능 |
| `OPENAI_API_KEY` | OpenAI API 키 | 임베딩/매칭 |
| `GEMINI_API_KEY` | Gemini API 키 | 경량 AI (기본값) |
| `GROQ_API_KEY` | Groq API 키 | 경량 AI (대안) |
| `NAVER_CLIENT_ID` | Naver OAuth 클라이언트 ID | Naver 로그인 |
| `NAVER_CLIENT_SECRET` | Naver OAuth 시크릿 | Naver 로그인 |
| `LLM_LIGHT_PROVIDER` | 경량 AI 프로바이더 | `gemini` (기본) 또는 `groq` |

---

## VSCode 설정 (추천)

`.vscode/settings.json`:
```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

권장 Extensions:
- Go (golang.go)
- ESLint (dbaeumer.vscode-eslint)
- Tailwind CSS IntelliSense (bradlc.vscode-tailwindcss)
- Prettier (esbenp.prettier-vscode)

---

## 다음 단계

- [ ] Phase 1.6 구현 시작 (`docs/develop/phases/phase-1.6-admin-system.md`)
- [ ] TDD 워크플로우 숙지 (`CLAUDE.md`)
- [ ] 배포 준비 (`DEPLOYMENT.md`)

문제가 있으면 GitHub Issues에 등록해주세요!
