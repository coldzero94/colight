# Phase 1.5: 어드민 시스템

## 목표

어드민 전용 페이지를 구현하여 사용자 관리, 프롬프트 템플릿 관리, 시스템 통계를 확인할 수 있게 한다.

---

## 어드민 설계 원칙

1. **초기 어드민 생성**: DB 시드로 생성 (어드민 회원가입 UI 없음)
2. **같은 로그인 흐름**: Naver/Email 로그인 후 role 체크
3. **라우트 분리**: `/admin/*` 전용 라우트 그룹
4. **역할 기반 접근**: Go `AdminMiddleware` + Next.js `AdminGuard`
5. **어드민도 일반 사용자 기능 접근 가능**

---

## 초기 어드민 생성

### 방법 1: DB 시드 (추천)

`scripts/seed.go`에 어드민 계정 시드 추가:

```go
// scripts/seed_admin.go
func seedAdmin(ctx context.Context, client *ent.Client) error {
    hash, _ := bcrypt.GenerateFromPassword([]byte("admin-password-change-me"), 12)

    _, err := client.UserProfile.Create().
        SetEmail("admin@colight.kr").
        SetPasswordHash(string(hash)).
        SetNickname("관리자").
        SetAuthProvider(userprofile.AuthProviderEmail).
        SetRole(userprofile.RoleAdmin).
        SetEmailVerified(true).
        SaveX(ctx)

    return err
}
```

### 방법 2: SQL 직접 실행

```sql
-- 기존 사용자를 어드민으로 승격
UPDATE user_profiles SET role = 'admin' WHERE email = 'your-email@example.com';
```

---

## 어드민 페이지 구조

```text
apps/web/src/app/
└── (admin)/
    ├── layout.tsx              # 어드민 레이아웃 (AdminGuard + 어드민 사이드바)
    ├── admin/
    │   ├── page.tsx            # 대시보드 (시스템 통계)
    │   ├── users/
    │   │   └── page.tsx        # 사용자 관리
    │   └── prompts/
    │       └── page.tsx        # 프롬프트 관리
```

---

## A. 어드민 대시보드 (`/admin`)

### 시스템 통계 카드

```
┌──────────────────────────────────────────────────────┐
│  Colight Admin Dashboard                              │
├──────────────┬──────────────┬───────────────┬────────┤
│ 전체 사용자    │ 오늘 활성     │ 전체 경험      │ 코칭   │
│    156       │     23       │    1,234      │  89    │
├──────────────┴──────────────┴───────────────┴────────┤
│                                                      │
│  인증 제공자 분포                                       │
│  ┌─ Naver: 120명 (77%) ──────────────────────┐       │
│  ├─ Email:  36명 (23%) ────────┘             │       │
│  └───────────────────────────────────────────┘       │
│                                                      │
│  빠른 링크                                             │
│  [사용자 관리] [프롬프트 관리] [메인 앱으로]              │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### API: `GET /v1/admin/stats`

```json
{
  "data": {
    "total_users": 156,
    "active_users_today": 23,
    "total_experiences": 1234,
    "total_coaching_sessions": 89,
    "total_analyses": 45,
    "naver_auth_users": 120,
    "email_auth_users": 36
  }
}
```

---

## B. 사용자 관리 (`/admin/users`)

### 사용자 목록 테이블

```
┌──────────────────────────────────────────────────────────────┐
│  사용자 관리                           [검색: _____________]  │
├──────┬──────────┬────────┬──────┬────────────┬──────────────┤
│ 이메일 │ 닉네임    │ 인증방식 │ 역할  │ 최근 로그인  │ 가입일       │
├──────┼──────────┼────────┼──────┼────────────┼──────────────┤
│ a@.. │ 홍길동    │ Naver  │ user │ 2026-02-13 │ 2026-02-10  │
│ b@.. │ 김철수    │ Email  │ admin│ 2026-02-13 │ 2026-02-08  │
│ ...  │ ...      │ ...    │ ...  │ ...        │ ...         │
├──────┴──────────┴────────┴──────┴────────────┴──────────────┤
│  < 1 2 3 ... 8 >                           총 156명         │
└──────────────────────────────────────────────────────────────┘
```

### 기능

- 사용자 목록 조회 (페이지네이션)
- 검색 (이메일, 닉네임)
- 역할 필터 (전체 / user / admin)
- 역할 변경 (user ↔ admin) — 확인 다이얼로그
- 사용자 상세 정보 보기 (경험 수, 코칭 세션 수 등)

### API 사용

```
GET /v1/admin/users?limit=20&offset=0&role=user&search=홍길동
PUT /v1/admin/users/{id}/role  { "role": "admin" }
```

---

## C. 프롬프트 관리 (`/admin/prompts`)

### 프롬프트 목록 + 편집

```
┌──────────────────────────────────────────────────────────────┐
│  프롬프트 템플릿 관리                   [카테고리: 전체 ▾]     │
├──────────┬──────────────┬──────────────┬──────┬─────────────┤
│ 카테고리   │ 서브카테고리   │ 이름         │ 모델  │ 사용횟수     │
├──────────┼──────────────┼──────────────┼──────┼─────────────┤
│ exp_cls  │ weapon_tag   │ 경험 무기 분류 │ gpt  │ 234         │
│ exp_cls  │ interview    │ 경험 인터뷰   │ gpt  │ 56          │
│ coaching │ q_analysis   │ 문항 분석     │ claude│ 123         │
│ coaching │ weapon_enh   │ 경험 강화     │ claude│ 89          │
├──────────┴──────────────┴──────────────┴──────┴─────────────┤
│  클릭하면 편집 패널 열림                                       │
└──────────────────────────────────────────────────────────────┘

편집 패널:
┌──────────────────────────────────────────────────────────────┐
│  프롬프트 편집: 경험 무기 자동 분류                              │
│                                                              │
│  System Prompt:                                              │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ You are an expert career counselor specializing...   │    │
│  │ (텍스트 에디터)                                       │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              │
│  User Prompt Template:                                       │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ {{experience_content}} 분석하여...                     │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              │
│  Temperature: [0.3]  Max Tokens: [2000]  Active: [✓]         │
│                                                              │
│  [취소]                                         [저장]        │
└──────────────────────────────────────────────────────────────┘
```

### 기능

- 프롬프트 목록 조회 (카테고리 필터)
- 프롬프트 편집 (system_prompt, user_prompt_template, temperature, max_tokens, is_active)
- 변경 사항 저장 → version 자동 증가
- 프롬프트 활성/비활성 토글

### API 사용

```
GET /v1/admin/prompts?category=coaching
PUT /v1/admin/prompts/{id}  { "system_prompt": "...", "temperature": 0.3 }
```

---

## Go 백엔드 (Admin Controller)

### `internal/controller/admin_controller.go`

```go
type AdminController struct {
    db *ent.Client
}

// ListUsers — 사용자 목록 (페이지네이션, 검색, 역할 필터)
func (c *AdminController) ListUsers(ctx *gin.Context) {
    query := c.db.UserProfile.Query()

    // Filter by role
    if role := ctx.Query("role"); role != "" {
        query = query.Where(userprofile.RoleEQ(userprofile.Role(role)))
    }

    // Search by email or nickname
    if search := ctx.Query("search"); search != "" {
        query = query.Where(
            userprofile.Or(
                userprofile.EmailContains(search),
                userprofile.NicknameContains(search),
            ),
        )
    }

    // Count
    count, _ := query.Count(ctx)

    // Pagination
    limit, offset := parsePagination(ctx)
    users, err := query.
        Limit(limit).
        Offset(offset).
        Order(ent.Desc(userprofile.FieldCreatedAt)).
        All(ctx)

    // ... return response
}

// GetStats — 시스템 통계
func (c *AdminController) GetStats(ctx *gin.Context) {
    totalUsers, _ := c.db.UserProfile.Query().Count(ctx)
    naverUsers, _ := c.db.UserProfile.Query().
        Where(userprofile.AuthProviderEQ(userprofile.AuthProviderNaver)).
        Count(ctx)
    totalExperiences, _ := c.db.Experience.Query().Count(ctx)
    // ... etc

    ctx.JSON(200, gin.H{
        "data": gin.H{
            "total_users":              totalUsers,
            "naver_auth_users":         naverUsers,
            "email_auth_users":         totalUsers - naverUsers,
            "total_experiences":        totalExperiences,
            // ...
        },
    })
}
```

---

## 어드민 레이아웃

### `src/app/(admin)/layout.tsx`

```typescript
import { AdminGuard } from "@/components/auth/admin-guard";
import { AdminSidebar } from "@/components/admin/admin-sidebar";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminGuard>
      <div className="flex h-screen">
        <AdminSidebar />
        <main className="flex-1 overflow-auto p-6">
          {children}
        </main>
      </div>
    </AdminGuard>
  );
}
```

### 어드민 사이드바 메뉴

| 메뉴 | 경로 | 아이콘 |
|------|------|--------|
| 대시보드 | `/admin` | LayoutDashboard |
| 사용자 관리 | `/admin/users` | Users |
| 프롬프트 관리 | `/admin/prompts` | FileText |
| 구분선 | --- | --- |
| 메인 앱으로 | `/experiences` | ArrowLeft |

---

## 체크리스트

### Go 백엔드

- [ ] `controller/admin_controller.go` 구현
  - [ ] `ListUsers` — 사용자 목록 (검색, 필터, 페이지네이션)
  - [ ] `GetUser` — 사용자 상세
  - [ ] `UpdateUserRole` — 역할 변경
  - [ ] `GetStats` — 시스템 통계
  - [ ] `ListPrompts` — 프롬프트 목록
  - [ ] `UpdatePrompt` — 프롬프트 수정

### 어드민 시드

- [ ] `scripts/seed_admin.go` — 기본 어드민 계정 시드
- [ ] `moon run backend:seed` 명령어에 admin 시드 추가

### Next.js 프론트엔드

- [ ] `(admin)/layout.tsx` — 어드민 레이아웃 + AdminGuard
- [ ] `(admin)/admin/page.tsx` — 대시보드 (통계 카드)
- [ ] `(admin)/admin/users/page.tsx` — 사용자 관리 테이블
- [ ] `(admin)/admin/prompts/page.tsx` — 프롬프트 관리
- [ ] `components/admin/admin-sidebar.tsx` — 어드민 사이드바
- [ ] `hooks/use-admin.ts` — 어드민 API React Query hooks

---

## 검증 방법

1. 어드민 계정으로 로그인 → `/admin` 접속 가능
2. 일반 사용자로 로그인 → `/admin` 접속 시 리디렉션
3. 사용자 목록 조회 → 테이블 정상 표시
4. 사용자 역할 변경 → 확인 다이얼로그 → 변경 성공
5. 프롬프트 편집 → 저장 → 변경 사항 반영
6. 시스템 통계 → 수치 정상 표시

---

## 산출물

- `internal/controller/admin_controller.go`
- `scripts/seed_admin.go`
- `src/app/(admin)/layout.tsx`
- `src/app/(admin)/admin/page.tsx`
- `src/app/(admin)/admin/users/page.tsx`
- `src/app/(admin)/admin/prompts/page.tsx`
- `src/components/admin/admin-sidebar.tsx`
