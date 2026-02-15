# Phase 1.2: Auth & Admin API 정의 (TypeSpec)

## 목표

TypeSpec으로 인증 + 어드민 API 엔드포인트를 정의하고, OpenAPI 스펙 → Go 서버 코드 → TS 클라이언트를 생성한다.

---

## API 엔드포인트 목록

### Auth API (인증)

| 메서드 | 엔드포인트 | 설명 | 인증 |
|--------|-----------|------|------|
| `GET` | `/v1/auth/naver/login` | Naver OAuth 시작 (302 redirect) | 불필요 |
| `GET` | `/v1/auth/naver/callback` | Naver OAuth 콜백 처리 | 불필요 |
| `POST` | `/v1/auth/signup` | 이메일 회원가입 | 불필요 |
| `POST` | `/v1/auth/login` | 이메일 로그인 | 불필요 |
| `POST` | `/v1/auth/refresh` | JWT 토큰 갱신 | 필요 (refresh token) |
| `GET` | `/v1/auth/me` | 현재 사용자 정보 조회 | 필요 |
| `POST` | `/v1/auth/logout` | 로그아웃 | 필요 |

### Admin API (어드민)

| 메서드 | 엔드포인트 | 설명 | 인증 |
|--------|-----------|------|------|
| `GET` | `/v1/admin/users` | 사용자 목록 조회 | Admin |
| `GET` | `/v1/admin/users/{id}` | 사용자 상세 조회 | Admin |
| `PUT` | `/v1/admin/users/{id}/role` | 사용자 역할 변경 | Admin |
| `GET` | `/v1/admin/stats` | 시스템 통계 | Admin |
| `GET` | `/v1/admin/prompts` | 프롬프트 템플릿 목록 | Admin |
| `PUT` | `/v1/admin/prompts/{id}` | 프롬프트 템플릿 수정 | Admin |

---

## TypeSpec 정의

### `packages/protocol/src/auth/auth.tsp`

```typespec
import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.Auth;

// ─── Models ───

model SignupRequest {
  email: string;
  password: string;
  nickname?: string;
}

model LoginRequest {
  email: string;
  password: string;
}

model AuthTokens {
  @encodedName("application/json", "access_token")
  accessToken: string;

  @encodedName("application/json", "refresh_token")
  refreshToken: string;

  @encodedName("application/json", "expires_in")
  expiresIn: int32;
}

model AuthResponse {
  user: UserInfo;
  tokens: AuthTokens;
}

model UserInfo {
  id: string;
  email?: string;
  nickname?: string;
  role: "user" | "admin";

  @encodedName("application/json", "auth_provider")
  authProvider: "email" | "naver";

  @encodedName("application/json", "onboarding_completed")
  onboardingCompleted: boolean;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

// ─── Auth Interface ───

@route("/v1/auth")
@tag("Auth")
interface AuthAPI {
  /** Naver OAuth 시작 — Naver 로그인 페이지로 302 리디렉션 */
  @get
  @route("/naver/login")
  naverLogin(): void;

  /** Naver OAuth 콜백 — code 교환 후 프론트엔드로 리디렉션 */
  @get
  @route("/naver/callback")
  naverCallback(@query code: string, @query state: string): void;

  /** 이메일 회원가입 */
  @post
  @route("/signup")
  signup(@body body: SignupRequest): {
    @statusCode statusCode: 201;
    @body body: { data: AuthResponse };
  } | Colight.Common.ErrorResponse;

  /** 이메일 로그인 */
  @post
  @route("/login")
  login(@body body: LoginRequest): {
    @statusCode statusCode: 200;
    @body body: { data: AuthResponse };
  } | Colight.Common.ErrorResponse;

  /** JWT 토큰 갱신 */
  @post
  @route("/refresh")
  refresh(@body body: { @encodedName("application/json", "refresh_token") refreshToken: string }): {
    @statusCode statusCode: 200;
    @body body: { data: AuthTokens };
  } | Colight.Common.ErrorResponse;

  /** 현재 로그인 사용자 정보 */
  @get
  @route("/me")
  @useAuth(BearerAuth)
  me(): {
    @statusCode statusCode: 200;
    @body body: { data: UserInfo };
  } | Colight.Common.ErrorResponse;

  /** 로그아웃 */
  @post
  @route("/logout")
  @useAuth(BearerAuth)
  logout(): {
    @statusCode statusCode: 200;
    @body body: { success: boolean };
  };
}
```

### `packages/protocol/src/admin/admin.tsp`

```typespec
import "@typespec/http";
import "../common/errors.tsp";
import "../auth/auth.tsp";

using TypeSpec.Http;

namespace Colight.Admin;

// ─── Models ───

model AdminUserListItem {
  id: string;
  email?: string;
  nickname?: string;
  role: "user" | "admin";

  @encodedName("application/json", "auth_provider")
  authProvider: "email" | "naver";

  @encodedName("application/json", "experience_count")
  experienceCount: int32;

  @encodedName("application/json", "last_login_at")
  lastLoginAt?: utcDateTime;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model UpdateRoleRequest {
  role: "user" | "admin";
}

model SystemStats {
  @encodedName("application/json", "total_users")
  totalUsers: int32;

  @encodedName("application/json", "active_users_today")
  activeUsersToday: int32;

  @encodedName("application/json", "total_experiences")
  totalExperiences: int32;

  @encodedName("application/json", "total_coaching_sessions")
  totalCoachingSessions: int32;

  @encodedName("application/json", "total_analyses")
  totalAnalyses: int32;

  @encodedName("application/json", "naver_auth_users")
  naverAuthUsers: int32;

  @encodedName("application/json", "email_auth_users")
  emailAuthUsers: int32;
}

model PromptTemplateDetail {
  id: string;
  category: string;

  @encodedName("application/json", "sub_category")
  subCategory: string;

  name: string;

  @encodedName("application/json", "system_prompt")
  systemPrompt: string;

  @encodedName("application/json", "user_prompt_template")
  userPromptTemplate: string;

  model: string;
  temperature: float32;

  @encodedName("application/json", "max_tokens")
  maxTokens: int32;

  version: int32;

  @encodedName("application/json", "is_active")
  isActive: boolean;

  @encodedName("application/json", "usage_count")
  usageCount: int32;
}

model UpdatePromptRequest {
  @encodedName("application/json", "system_prompt")
  systemPrompt?: string;

  @encodedName("application/json", "user_prompt_template")
  userPromptTemplate?: string;

  temperature?: float32;

  @encodedName("application/json", "max_tokens")
  maxTokens?: int32;

  @encodedName("application/json", "is_active")
  isActive?: boolean;
}

// ─── Admin Interface ───

@route("/v1/admin")
@tag("Admin")
@useAuth(BearerAuth)
interface AdminAPI {
  /** 사용자 목록 조회 */
  @get
  @route("/users")
  listUsers(
    @query limit?: int32 = 20,
    @query offset?: int32 = 0,
    @query role?: "user" | "admin",
    @query search?: string,
  ): {
    @statusCode statusCode: 200;
    @body body: { data: AdminUserListItem[]; count: int32 };
  } | Colight.Common.ErrorResponse;

  /** 사용자 상세 조회 */
  @get
  @route("/users/{id}")
  getUser(@path id: string): {
    @statusCode statusCode: 200;
    @body body: { data: AdminUserListItem };
  } | Colight.Common.ErrorResponse;

  /** 사용자 역할 변경 */
  @put
  @route("/users/{id}/role")
  updateUserRole(@path id: string, @body body: UpdateRoleRequest): {
    @statusCode statusCode: 200;
    @body body: { data: AdminUserListItem };
  } | Colight.Common.ErrorResponse;

  /** 시스템 통계 */
  @get
  @route("/stats")
  getStats(): {
    @statusCode statusCode: 200;
    @body body: { data: SystemStats };
  } | Colight.Common.ErrorResponse;

  /** 프롬프트 템플릿 목록 */
  @get
  @route("/prompts")
  listPrompts(@query category?: string): {
    @statusCode statusCode: 200;
    @body body: { data: PromptTemplateDetail[] };
  } | Colight.Common.ErrorResponse;

  /** 프롬프트 템플릿 수정 */
  @put
  @route("/prompts/{id}")
  updatePrompt(@path id: string, @body body: UpdatePromptRequest): {
    @statusCode statusCode: 200;
    @body body: { data: PromptTemplateDetail };
  } | Colight.Common.ErrorResponse;
}
```

### `packages/protocol/src/main.tsp` 수정

```typespec
// 기존 import에 추가:
import "./auth/auth.tsp";
import "./admin/admin.tsp";
```

---

## 체크리스트

- [x] TypeSpec 모델 정의
  - [x] `packages/protocol/src/auth/auth.tsp` 생성
  - [x] `packages/protocol/src/admin/admin.tsp` 생성
- [x] `main.tsp`에 import 추가
- [x] OpenAPI 생성
  - [x] `moon run protocol:generate`
  - [x] OpenAPI 스펙에 `/v1/auth/*` + `/v1/admin/*` 엔드포인트 포함 확인
- [x] Go 서버 코드 생성
  - [x] `moon run backend:generate-api`
  - [x] `StrictServerInterface`에 Auth + Admin 메서드 추가 확인
- [x] TS 클라이언트 생성
  - [x] `moon run web:generate-client`
  - [x] `src/api/generated/` 에 Auth + Admin 타입/SDK 생성 확인

---

## 검증 방법

- `tsp compile .` → 에러 없음
- OpenAPI 스펙에 Auth 7개 + Admin 6개 = 13개 엔드포인트 포함
- Go `StrictServerInterface`에 Auth + Admin 메서드 존재
- TS `sdk.gen.ts`에 Auth + Admin 함수 존재

---

## 산출물

- `packages/protocol/src/auth/auth.tsp`
- `packages/protocol/src/admin/admin.tsp`
- 업데이트된 OpenAPI 스펙
- 업데이트된 Go 서버 인터페이스
- 업데이트된 TS 클라이언트
