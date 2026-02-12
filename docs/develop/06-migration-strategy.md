# Atlas 마이그레이션 전략

> Atlas + Ent 스키마 기반 마이그레이션 관리 및 배포 워크플로

---

## 1. 도구 및 아키텍처

### Source of Truth

**Ent 스키마** (`apps/backend/ent/schema/`)가 데이터베이스 스키마의 단일 진실 공급원(Single Source of Truth)이다. SQL을 직접 작성하지 않고, Go의 Ent 스키마를 수정하면 Atlas가 자동으로 마이그레이션 SQL을 생성한다.

### atlas.hcl 설정

```hcl
env "local" {
  src = "ent://ent/schema"
  dev = "docker://postgres/16/dev"
  url = "postgres://user:pass@localhost:5532/colight_dev?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
}

env "prod" {
  src = "ent://ent/schema"
  dev = "docker://postgres/16/dev"
  url = "env://DATABASE_URL"  # Supabase connection string
  migration {
    dir = "file://migrations"
  }
}
```

### Atlas 핵심 명령어

```bash
# 마이그레이션 SQL 생성 (diff)
atlas migrate diff <name> --env local

# 로컬 DB에 마이그레이션 적용
atlas migrate apply --env local

# 마이그레이션 상태 확인
atlas migrate status --env local

# 프로덕션 DB에 마이그레이션 적용
DATABASE_URL=<supabase_url> atlas migrate apply --env prod

# 롤백 (최근 마이그레이션 되돌리기)
atlas migrate down --env local
```

### moon tasks (apps/backend/moon.yml)

```yaml
migrate-diff:
  command: "atlas migrate diff ${name} --dir file://migrations --to ent://ent/schema --dev-url docker://postgres/16/dev?search_path=public"
migrate-apply:
  command: "atlas migrate apply --dir file://migrations --url $DATABASE_URL"
```

### 파일 명명 규칙

```
migrations/YYYYMMDDHHMMSS_<description>.sql
```

**예시:**

```
migrations/
├── 20260301000000_phase0_initial_schema.sql
├── 20260315000000_phase2_add_experience_summary.sql
├── 20260320000000_phase3_add_cache_ttl_index.sql
└── 20260401000000_phase4_add_matching_score.sql
```

**규칙:**
- 타임스탬프는 `atlas migrate diff` 명령어가 자동 생성
- `description`은 영문 소문자 + 언더스코어 (`snake_case`)
- Phase 접두사를 붙여 어떤 단계의 마이그레이션인지 명시

---

## 2. 개발 워크플로

### 로컬 개발 (개인 작업)

```bash
# 1. Ent 스키마 수정
vim ent/schema/experience.go

# 2. 마이그레이션 SQL 생성
moon run backend:migrate-diff -- name=add_experience_summary

# 3. 생성된 SQL 리뷰
cat migrations/20260211_add_experience_summary.sql

# 4. 로컬 DB에 적용
moon run backend:migrate-apply

# 5. 로컬 테스트
go run ./cmd/api

# 6. 마이그레이션 파일 커밋
git add migrations/ && git commit -m "Phase 2.1: Add experience summary field"
```

### 프로덕션 배포

```bash
# 1. Ent 스키마 수정 → 마이그레이션 생성 → 로컬 테스트 (위 절차)
# 2. 코드 리뷰 (PR) — 마이그레이션 SQL 파일 포함
# 3. Supabase(프로덕션)에 적용
DATABASE_URL=<supabase_url> atlas migrate apply --env prod
```

---

## 3. Phase별 마이그레이션 계획

### Phase 0: 초기 스키마 (전체 15개 테이블 + pgvector 확장)

> Phase 0에서 모든 핵심 테이블을 한번에 생성한다. Ent 스키마에 전체 엔티티를 정의하고 `atlas migrate diff`로 초기 마이그레이션을 생성.

포함 대상:
- `user_profiles` — 사용자 프로필
- `experiences` — 경험 (STAR 구조, 임베딩 벡터)
- `experience_tags` — 경험 태그
- `experience_weapons` — 경험-무기 매핑
- `experience_usages` — 경험 활용 이력
- `weapon_categories` — 7대 무기 마스터 데이터
- `prompt_templates` — AI 프롬프트 템플릿
- `question_patterns` — 공통 문항 패턴
- `applications` — 지원 현황
- `company_analyses` — 기업 분석 결과
- `company_analysis_cache` — 분석 캐시
- `talent_profiles` — 인재상 DB
- `cover_letters` — 자소서 문항
- `cover_letter_versions` — 자소서 버전
- `coaching_sessions` — 코칭 세션 로그
- pgvector 확장 활성화

### Phase 2+: 이후 마이그레이션

이후 Phase에서는 Ent 스키마 변경 후 diff로 마이그레이션을 생성한다. 주로 필드 추가, 인덱스 생성 등의 변경.

```bash
# Phase 2: 경험 필드 추가
moon run backend:migrate-diff -- name=phase2_add_experience_fields

# Phase 3: 분석 캐시 인덱스 추가
moon run backend:migrate-diff -- name=phase3_add_cache_indexes
```

---

## 4. 로컬 PostgreSQL 환경

로컬 개발/테스트에는 Docker Compose로 PostgreSQL 16을 사용한다. Supabase에 배포하기 전에 로컬에서 충분히 테스트한다.

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16
    ports:
      - "5532:5432"
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: colight_dev
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

---

## 5. 시드 데이터

시드 데이터는 마이그레이션과 **완전히 분리**하여 Go 스크립트로 관리한다.

```
apps/backend/scripts/seed.go
```

### 시드 실행

```bash
# 전체 시드 데이터 적용
go run ./scripts/seed.go all

# 특정 카테고리만 시드
go run ./scripts/seed.go weapons        # 무기 카테고리
go run ./scripts/seed.go patterns       # 문항 패턴
go run ./scripts/seed.go prompts        # 프롬프트 템플릿
```

### 시드 데이터 대상

| 테이블 | 설명 | Phase |
|--------|------|-------|
| `weapon_categories` | 7대 무기 마스터 데이터 | Phase 0 |
| `question_patterns` | 공통 문항 패턴 | Phase 0 |
| `prompt_templates` | AI 프롬프트 템플릿 | Phase 0 |

---

## 6. 롤백 전략

### Atlas 롤백

Atlas는 `atlas migrate down` 명령으로 마이그레이션 롤백을 지원한다.

```bash
# 최근 1개 마이그레이션 롤백
atlas migrate down --env local

# 프로덕션 롤백
DATABASE_URL=<supabase_url> atlas migrate down --env prod
```

### 롤백 가능 여부

| 변경 유형 | 롤백 가능? | 대응 방법 |
|-----------|-----------|----------|
| 인덱스 추가/삭제 | O | `drop index` / `create index` |
| 컬럼 추가 | O | `alter table drop column` |
| 컬럼 삭제 | X | 데이터 손실, 사전 백업 필수 |
| 테이블 삭제 | X | 데이터 손실, 사전 백업 필수 |
| 데이터 타입 변경 | 부분적 | 원복 가능한 경우만 |

**원칙:** 마이그레이션은 가능한 한 **되돌릴 수 있는 형태**로 작성한다.

---

## 7. 안전한 스키마 변경 규칙

### 컬럼 추가

```sql
-- 항상 DEFAULT 값을 지정하여 기존 행 영향 최소화
ALTER TABLE experiences ADD COLUMN summary TEXT DEFAULT '';
```

Ent 스키마에서는 `.Default("")`로 지정:

```go
field.String("summary").Default("").Optional()
```

### 컬럼 삭제 (2단계)

```
Step 1 (마이그레이션 N): 새 컬럼 추가 + 데이터 마이그레이션
Step 2 (마이그레이션 N+1): 구 컬럼 삭제
```

### 인덱스 생성 (무중단)

```sql
-- CONCURRENTLY 옵션으로 테이블 잠금 없이 인덱스 생성
CREATE INDEX CONCURRENTLY idx_name ON table_name(column_name);
```

> `CONCURRENTLY`는 트랜잭션 내에서 실행 불가. 해당 마이그레이션 파일에는 하나의 concurrent index만 포함할 것.

### 금지 사항

- `DROP TABLE` → 데이터 백업 확인 후에만 사용
- `ALTER COLUMN TYPE` → 데이터 손실 가능, 새 컬럼 추가 방식으로 대체
- `TRUNCATE` → 프로덕션에서 절대 사용 금지

---

## 8. Supabase 연동

### 역할 분담

| 관리 주체 | 대상 |
|-----------|------|
| **Supabase Auth** | `auth.users` 테이블 (인증/사용자 관리) |
| **Go/Atlas** | 나머지 모든 테이블 (비즈니스 로직) |

### 연결 방식

- `DATABASE_URL` 환경변수로 Supabase Connection Pooler에 연결
- RLS 정책은 Atlas 마이그레이션에 포함하거나 Supabase Dashboard에서 관리
- `user_profiles.id`는 `auth.users(id)`를 FK로 참조

---

## 9. 마이그레이션 체크리스트

새 마이그레이션 작성 시:

- [ ] Ent 스키마가 올바르게 수정되었는가?
- [ ] `moon run backend:migrate-diff`로 SQL이 정상 생성되었는가?
- [ ] 생성된 SQL을 리뷰했는가? (불필요한 변경 없는지)
- [ ] `moon run backend:migrate-apply` (로컬)에서 에러 없이 적용되는가?
- [ ] 롤백이 가능한 변경인가?
- [ ] `CONCURRENTLY` 인덱스가 별도 마이그레이션에 분리되었는가?
- [ ] 시드 데이터 변경이 필요하면 `seed.go`도 업데이트했는가?
