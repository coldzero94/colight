# Phase 1.1: DB 스키마 수정

## 목표

user_profiles 테이블에 인증 관련 필드를 추가하고, Naver OAuth / Email+Password / 어드민 역할을 지원한다.

## 변경 사항

### user_profiles 테이블 — 추가 필드

| 필드 | 타입 | 제약조건 | 설명 |
|------|------|----------|------|
| `email` | `string(255)` | UNIQUE, Optional, Nillable | 이메일 (Email/PW 로그인 또는 Naver에서 가져온 이메일) |
| `password_hash` | `string(255)` | Optional, Nillable, Sensitive | bcrypt 해시 (Email/PW 로그인 전용) |
| `naver_id` | `string(255)` | UNIQUE, Optional, Nillable | Naver OAuth 사용자 고유 ID |
| `auth_provider` | `enum` | Default("email") | 인증 제공자: `"email"`, `"naver"` |
| `role` | `enum` | Default("user") | 역할: `"user"`, `"admin"` |
| `email_verified` | `bool` | Default(false) | 이메일 인증 여부 (Naver 로그인 시 자동 true) |
| `last_login_at` | `time` | Optional, Nillable | 마지막 로그인 시각 |

### user_profiles 테이블 — 제거 필드

| 필드 | 이유 |
|------|------|
| `user_id` (UUID) | Supabase auth.users 참조용이었음. 자체 인증이므로 `id` (BaseMixin PK)를 직접 사용 |

> **중요**: 기존 `user_id` 필드는 Supabase `auth.users(id)` 참조용이었습니다.
> 자체 인증으로 전환하면서 `user_profiles.id` (UUID PK)가 곧 user_id가 됩니다.
> 다른 테이블의 `user_id` FK는 `user_profiles.id`를 직접 참조하도록 변경합니다.

---

## Ent 스키마 수정

### `ent/schema/userprofile.go` 변경

```go
func (UserProfile) Fields() []ent.Field {
    return []ent.Field{
        // --- 인증 필드 (Phase 1 추가) ---
        field.String("email").
            Optional().
            Nillable().
            MaxLen(255).
            Comment("Email address"),
        field.String("password_hash").
            Optional().
            Nillable().
            MaxLen(255).
            Sensitive().
            Comment("bcrypt hashed password (email auth only)"),
        field.String("naver_id").
            Optional().
            Nillable().
            MaxLen(255).
            Comment("Naver OAuth user ID"),
        field.Enum("auth_provider").
            Values("email", "naver").
            Default("email").
            Comment("Authentication provider"),
        field.Enum("role").
            Values("user", "admin").
            Default("user").
            Comment("User role"),
        field.Bool("email_verified").
            Default(false).
            Comment("Whether email is verified"),
        field.Time("last_login_at").
            Optional().
            Nillable().
            Comment("Last login timestamp"),

        // --- 기존 프로필 필드 ---
        field.String("nickname").
            Optional().
            MaxLen(50).
            Comment("Display name"),
        field.String("target_job").
            Optional().
            MaxLen(100).
            Comment("Target job position"),
        field.String("target_industry").
            Optional().
            MaxLen(100).
            Comment("Target industry"),
        field.String("education_level").
            Optional().
            MaxLen(20).
            Comment("Education level"),
        field.Int("graduation_year").
            Optional().
            Nillable().
            Comment("Graduation year"),
        field.Int("experience_years").
            Default(0).
            Comment("Years of experience"),
        field.Bool("onboarding_completed").
            Default(false).
            Comment("Whether onboarding is completed"),
    }
}

func (UserProfile) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("email").Unique(),
        index.Fields("naver_id").Unique(),
        index.Fields("role"),
    }
}
```

### 다른 테이블의 `user_id` FK 변경

기존: `field.UUID("user_id", uuid.UUID{}).Comment("References auth.users(id)")` (Supabase 참조)
변경: `user_id`를 `user_profiles.id` 직접 참조 (Ent edge로 관리)

> **주의**: 이 변경은 experience, application, company_analysis, cover_letter, coaching_session, experience_usage 테이블에 모두 적용됩니다.

---

## 체크리스트

- [ ] `ent/schema/userprofile.go` 수정
  - [ ] `user_id` 필드 제거
  - [ ] 인증 필드 7개 추가 (email, password_hash, naver_id, auth_provider, role, email_verified, last_login_at)
  - [ ] 인덱스 추가 (email UNIQUE, naver_id UNIQUE, role)
- [ ] 다른 스키마의 `user_id` FK 정리
  - [ ] experience.go — edge.From("user", UserProfile.Type) 유지, field("user_id") 참조가 user_profiles.id를 가리키도록 확인
  - [ ] application.go — 동일
  - [ ] company_analysis.go — 동일
  - [ ] cover_letter.go — 동일
  - [ ] coaching_session.go — 동일
  - [ ] experience_usage.go — 동일
- [ ] Ent 코드 재생성
  - [ ] `moon run backend:generate-ent`
  - [ ] 컴파일 에러 확인 및 수정
- [ ] 마이그레이션 생성
  - [ ] `moon run backend:migrate-diff -- name=add_auth_fields`
  - [ ] 생성된 SQL 검토 (ALTER TABLE 확인)
- [ ] 마이그레이션 적용
  - [ ] `moon run backend:migrate-apply`
  - [ ] 테이블 구조 확인

---

## 검증 방법

```sql
-- user_profiles 테이블 구조 확인
\d user_profiles

-- 인덱스 확인
\di *user_profiles*

-- Expected: email, password_hash, naver_id, auth_provider, role, email_verified, last_login_at 컬럼 존재
-- Expected: user_profiles_email_key (UNIQUE), user_profiles_naver_id_key (UNIQUE) 인덱스 존재
```

---

## 산출물

- 수정된 `ent/schema/userprofile.go`
- 마이그레이션 SQL 파일 (`migrations/YYYYMMDD_add_auth_fields.sql`)
