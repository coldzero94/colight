# Colight - 어드민 시스템 기획

> 시스템 관리자를 위한 어드민 대시보드 — API 키 관리, AI 모델 설정, 사용자 관리, 사용량 모니터링, 시스템 설정

---

## 1. 개요

### 목적

서비스 운영에 필요한 시스템 관리 기능을 어드민 페이지에서 제공한다. 서버 접속이나 환경변수 직접 수정 없이, 웹 UI에서 핵심 설정을 변경할 수 있게 한다.

### 현재 상태 (Phase 1.5 완료)

| 기능 | 상태 |
|------|------|
| 대시보드 (시스템 통계) | ✅ 구현 완료 |
| 사용자 관리 (목록/검색/역할) | ✅ 구현 완료 |
| 프롬프트 관리 (편집/활성화) | ✅ 구현 완료 |
| **역할 계층 (다단계 권한)** | ❌ 미기획 |
| **API 키 관리** | ❌ 미기획 |
| **AI 모델 설정** | ❌ 미기획 |
| **시스템 설정** | ❌ 미기획 |
| **사용량/비용 모니터링** | ❌ 미기획 |
| **감사 로그** | ❌ 미기획 |
| **시스템 헬스 모니터링** | ❌ 미기획 |
| **백그라운드 작업 관리** | ❌ 미기획 |
| **계정 정지/강제 로그아웃** | ❌ 미기획 |
| **피드백 관리** | ❌ 미기획 |
| **크롤러 모니터링** | ❌ 미기획 |
| **데이터 프라이버시 (내보내기/삭제)** | ❌ 미기획 |

---

## 2. 역할 계층 (Role Hierarchy)

### 2.1 역할 정의

기존 2단계(`user` | `admin`)를 4단계로 확장한다. 각 역할은 하위 역할의 모든 권한을 포함한다.

| 역할 | 코드 | 설명 | 대상 |
|------|------|------|------|
| **일반 사용자** | `user` | 서비스 이용 (경험 관리, 코칭, 분석) | 모든 가입자 |
| **매니저** | `manager` | 사용자 지원 + 읽기 전용 모니터링 | 운영 담당자, CS |
| **어드민** | `admin` | 프롬프트 관리, 사용자 역할 변경, 한도 설정 | 서비스 관리자 |
| **슈퍼 어드민** | `super_admin` | API 키 변경, 모델 설정, 어드민 승격, 모든 권한 | 시스템 최고 관리자 |

### 2.2 권한 매트릭스

| 기능 | user | manager | admin | super_admin |
|------|:----:|:-------:|:-----:|:-----------:|
| 서비스 이용 (경험, 코칭, 분석) | ✅ | ✅ | ✅ | ✅ |
| 어드민 대시보드 접근 | ❌ | ✅ | ✅ | ✅ |
| 사용자 목록 조회 | ❌ | ✅ | ✅ | ✅ |
| 사용량 모니터링 (읽기) | ❌ | ✅ | ✅ | ✅ |
| 감사 로그 조회 | ❌ | ✅ | ✅ | ✅ |
| 프롬프트 편집 | ❌ | ❌ | ✅ | ✅ |
| 사용자 역할 변경 (→ user, manager) | ❌ | ❌ | ✅ | ✅ |
| 사용량 한도 변경 | ❌ | ❌ | ✅ | ✅ |
| 사용자 계정 정지/해제 | ❌ | ❌ | ✅ | ✅ |
| 사용자 역할 변경 (→ admin) | ❌ | ❌ | ❌ | ✅ |
| API 키 변경 | ❌ | ❌ | ❌ | ✅ |
| AI 모델 설정 변경 | ❌ | ❌ | ❌ | ✅ |
| 슈퍼 어드민 승격 | ❌ | ❌ | ❌ | ✅ |
| 어드민/슈퍼어드민 강등 | ❌ | ❌ | ❌ | ✅ |

### 2.3 역할 승격/강등 규칙

```
승격 규칙:
  super_admin → admin, manager, user 모두 승격/강등 가능
  admin       → user ↔ manager 만 변경 가능 (admin 이상 불가)
  manager     → 역할 변경 불가
  user        → 역할 변경 불가

자기 자신 강등:
  자기 자신의 역할은 변경 불가 (실수 방지)

마지막 super_admin 보호:
  시스템에 super_admin이 1명만 남은 경우, 해당 계정은 강등 불가
```

### 2.4 데이터 모델 변경

```sql
-- user_profiles.role 확장
-- 기존: ENUM('user', 'admin')
-- 변경: ENUM('user', 'manager', 'admin', 'super_admin')

ALTER TABLE user_profiles
  ALTER COLUMN role TYPE VARCHAR(20);
-- Ent 스키마에서 enum values 추가
```

```go
// ent/schema/userprofile.go (변경)
field.Enum("role").
    Values("user", "manager", "admin", "super_admin").
    Default("user").
    Comment("User role: user < manager < admin < super_admin"),
```

### 2.5 백엔드 미들웨어 설계

```go
// internal/infrastructure/middleware/role.go

// RoleLevel returns the numeric level for a role (higher = more privileged).
func RoleLevel(role string) int {
    switch role {
    case "super_admin":
        return 4
    case "admin":
        return 3
    case "manager":
        return 2
    case "user":
        return 1
    default:
        return 0
    }
}

// RequireRole creates middleware that requires a minimum role level.
func RequireRole(minRole string) gin.HandlerFunc {
    minLevel := RoleLevel(minRole)
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || RoleLevel(role.(string)) < minLevel {
            c.AbortWithStatusJSON(403, gin.H{
                "error": gin.H{
                    "message": "접근 권한이 부족합니다.",
                    "code":    "AUTH_003",
                    "required_role": minRole,
                },
            })
            return
        }
        c.Next()
    }
}
```

### 2.6 라우트별 권한 적용

```go
// cmd/api/main.go (라우트 등록)

admin := r.Group("/v1/admin")
admin.Use(middleware.AuthMiddleware(tokenService))

// manager 이상 접근 가능 (읽기 전용)
admin.GET("/stats", middleware.RequireRole("manager"), adminCtrl.GetStats)
admin.GET("/users", middleware.RequireRole("manager"), adminCtrl.ListUsers)
admin.GET("/users/:id", middleware.RequireRole("manager"), adminCtrl.GetUser)
admin.GET("/usage/*", middleware.RequireRole("manager"), adminCtrl.GetUsage)
admin.GET("/audit-logs", middleware.RequireRole("manager"), adminCtrl.ListAuditLogs)

// admin 이상 접근 가능 (쓰기)
admin.PUT("/users/:id/role", middleware.RequireRole("admin"), adminCtrl.UpdateUserRole)
admin.PUT("/users/:id/suspend", middleware.RequireRole("admin"), adminCtrl.SuspendUser)
admin.GET("/prompts", middleware.RequireRole("admin"), adminCtrl.ListPrompts)
admin.PUT("/prompts/:id", middleware.RequireRole("admin"), adminCtrl.UpdatePrompt)
admin.PUT("/configs/limits/*", middleware.RequireRole("admin"), adminCtrl.UpdateLimits)

// super_admin 전용 (시스템 설정)
admin.GET("/configs", middleware.RequireRole("super_admin"), adminCtrl.ListConfigs)
admin.PUT("/configs/:key", middleware.RequireRole("super_admin"), adminCtrl.UpdateConfig)
admin.POST("/configs/validate-key", middleware.RequireRole("super_admin"), adminCtrl.ValidateApiKey)
```

### 2.7 프론트엔드 권한 처리

#### auth-store.ts 변경

```typescript
// 기존
role: "user" | "admin"

// 변경
role: "user" | "manager" | "admin" | "super_admin"
```

#### 권한 유틸리티

```typescript
// lib/auth-utils.ts

const ROLE_LEVELS: Record<string, number> = {
  user: 1,
  manager: 2,
  admin: 3,
  super_admin: 4,
};

export function hasRole(userRole: string, requiredRole: string): boolean {
  return (ROLE_LEVELS[userRole] ?? 0) >= (ROLE_LEVELS[requiredRole] ?? 0);
}

export function isManager(role: string): boolean {
  return hasRole(role, "manager");
}

export function isAdmin(role: string): boolean {
  return hasRole(role, "admin");
}

export function isSuperAdmin(role: string): boolean {
  return hasRole(role, "super_admin");
}
```

#### AdminGuard 확장

```typescript
// 기존: role === "admin" 체크
// 변경: hasRole(role, "manager") 체크 (manager 이상 어드민 접근)

export function AdminGuard({ children, requiredRole = "manager" }: {
  children: React.ReactNode;
  requiredRole?: "manager" | "admin" | "super_admin";
}) {
  const { user } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (user && !hasRole(user.role, requiredRole)) {
      router.replace("/experiences");
    }
  }, [user, requiredRole, router]);

  if (!user || !hasRole(user.role, requiredRole)) {
    return <LoadingSpinner />;
  }

  return <>{children}</>;
}
```

#### 사이드바 메뉴 권한 분기

```typescript
// admin-sidebar.tsx
const menuItems = [
  // manager 이상 (읽기)
  { label: "대시보드", href: "/admin", icon: "📊", minRole: "manager" },
  { label: "사용자 관리", href: "/admin/users", icon: "👥", minRole: "manager" },
  { label: "사용량 모니터링", href: "/admin/usage", icon: "📈", minRole: "manager" },
  { label: "감사 로그", href: "/admin/logs", icon: "📜", minRole: "manager" },

  // admin 이상 (쓰기)
  { label: "프롬프트 관리", href: "/admin/prompts", icon: "📝", minRole: "admin" },

  // super_admin 전용
  { label: "시스템 설정", href: "/admin/settings", icon: "⚙️", minRole: "super_admin" },
];

// 현재 사용자 역할에 따라 메뉴 필터링
const visibleItems = menuItems.filter(item => hasRole(user.role, item.minRole));
```

### 2.8 사용자 관리 페이지 역할 변경 UI

```
┌──────────────────────────────────────────────────────────────┐
│  사용자 관리                                    총 156명       │
├──────┬──────────┬────────┬────────────┬────────────────────┤
│ 이메일 │ 닉네임    │ 인증    │ 역할        │ 액션               │
├──────┼──────────┼────────┼────────────┼────────────────────┤
│ a@.. │ 홍길동    │ Naver  │ 🟢 user    │ [역할 변경 ▾]       │
│ b@.. │ 김철수    │ Email  │ 🔵 manager │ [역할 변경 ▾]       │
│ c@.. │ 관리자    │ Email  │ 🟣 admin   │ [역할 변경 ▾]       │
│ d@.. │ 최고관리  │ Email  │ 🔴 super   │ (변경 불가)          │
└──────┴──────────┴────────┴────────────┴────────────────────┘

역할 변경 드롭다운 (admin이 보는 경우):
  ┌────────────┐
  │ ○ User     │  ← 선택 가능
  │ ○ Manager  │  ← 선택 가능
  │ ○ Admin    │  ← 비활성 (admin은 admin 승격 불가)
  │ ○ Super    │  ← 비활성
  └────────────┘

역할 변경 드롭다운 (super_admin이 보는 경우):
  ┌────────────┐
  │ ○ User     │  ← 선택 가능
  │ ○ Manager  │  ← 선택 가능
  │ ○ Admin    │  ← 선택 가능
  │ ○ Super    │  ← 선택 가능 (⚠️ 확인 필요)
  └────────────┘
```

### 2.9 역할별 배지 색상

| 역할 | 배지 색상 | 표시 |
|------|----------|------|
| `user` | Gray (`bg-gray-100 text-gray-700`) | User |
| `manager` | Blue (`bg-blue-100 text-blue-700`) | Manager |
| `admin` | Purple (`bg-purple-100 text-purple-700`) | Admin |
| `super_admin` | Red (`bg-red-100 text-red-700`) | Super Admin |

### 2.10 초기 슈퍼 어드민 생성

```go
// scripts/seed_admin.go (변경)
func seedSuperAdmin(ctx context.Context, client *ent.Client) error {
    hash, _ := bcrypt.GenerateFromPassword([]byte("super-admin-change-me"), 12)

    _, err := client.UserProfile.Create().
        SetEmail("admin@colight.kr").
        SetPasswordHash(string(hash)).
        SetNickname("최고관리자").
        SetAuthProvider(userprofile.AuthProviderEmail).
        SetRole(userprofile.RoleSuperAdmin).  // super_admin으로 생성
        SetEmailVerified(true).
        SaveX(ctx)

    return err
}
```

---

## 3. 어드민 페이지 구조 (확장)

```text
/admin                    # 대시보드 (기존)
/admin/users              # 사용자 관리 (기존)
/admin/prompts            # 프롬프트 관리 (기존)
/admin/settings           # 시스템 설정 (신규)
  ├── API Keys            # API 키 관리 탭
  ├── AI Models           # AI 모델 설정 탭
  └── Limits              # 사용량 한도 탭
/admin/usage              # 사용량 모니터링 (신규)
/admin/logs               # 감사 로그 (신규)
```

### 사이드바 메뉴 (확장)

| 메뉴 | 경로 | 아이콘 | 상태 |
|------|------|--------|------|
| 대시보드 | `/admin` | LayoutDashboard | 기존 |
| 사용자 관리 | `/admin/users` | Users | 기존 |
| 프롬프트 관리 | `/admin/prompts` | FileText | 기존 |
| **시스템 설정** | `/admin/settings` | Settings | **신규** |
| **사용량 모니터링** | `/admin/usage` | BarChart | **신규** |
| **감사 로그** | `/admin/logs` | ScrollText | **신규** |
| 구분선 | --- | --- | - |
| 메인 앱으로 | `/experiences` | ArrowLeft | 기존 |

---

## 4. 시스템 설정 (`/admin/settings`)

### 4.1 API 키 관리

관리자가 웹 UI에서 AI 프로바이더 API 키를 변경할 수 있다. 서버 재시작이나 환경변수 파일 수정 없이 런타임에 적용된다.

#### 데이터 모델

```sql
-- system_configs 테이블 (새로 생성)
CREATE TABLE system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,   -- 예: "api_key.anthropic"
    config_value TEXT NOT NULL,                 -- 암호화된 값
    description VARCHAR(500),                   -- 설명
    category VARCHAR(50) NOT NULL,              -- "api_key" | "model" | "limit" | "feature"
    is_secret BOOLEAN DEFAULT false,            -- true면 마스킹 표시
    updated_by UUID REFERENCES user_profiles(id),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### 관리 대상 API 키

| Config Key | 프로바이더 | 용도 | 필수 |
|------------|-----------|------|------|
| `api_key.anthropic` | Anthropic | Claude Sonnet 4.5 (Heavy AI) | ✅ |
| `api_key.gemini` | Google | Gemini 2.0 Flash (Light AI) | 조건부 |
| `api_key.groq` | Groq | Llama 3.3 70B (Light AI) | 조건부 |
| `api_key.openai` | OpenAI | text-embedding-3-small (임베딩) | ✅ |
| `api_key.dart` | DART | 기업 재무 데이터 | ✅ |
| `api_key.naver_id` | Naver | 뉴스 검색 API (Client ID) | ✅ |
| `api_key.naver_secret` | Naver | 뉴스 검색 API (Secret) | ✅ |

#### UI 와이어프레임

```
┌──────────────────────────────────────────────────────────────┐
│  시스템 설정                                                   │
│  [API Keys] [AI Models] [Limits]                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  API 키 관리                                                  │
│  ────────────────────────────────────────────────────────    │
│                                                              │
│  Anthropic (Claude)                              상태: ● 정상 │
│  ┌─────────────────────────────────────────────┐            │
│  │ sk-ant-••••••••••••••••••••••••••••abc       │ [변경]     │
│  └─────────────────────────────────────────────┘            │
│  마지막 수정: 2026-02-10 by 관리자                             │
│                                                              │
│  Gemini (Google)                                 상태: ● 정상 │
│  ┌─────────────────────────────────────────────┐            │
│  │ AI••••••••••••••••••••••••••••••xyz          │ [변경]     │
│  └─────────────────────────────────────────────┘            │
│  마지막 수정: 2026-02-08 by 관리자                             │
│                                                              │
│  OpenAI (임베딩)                                  상태: ● 정상 │
│  ┌─────────────────────────────────────────────┐            │
│  │ sk-••••••••••••••••••••••••••••••def         │ [변경]     │
│  └─────────────────────────────────────────────┘            │
│  마지막 수정: 2026-02-05 by 관리자                             │
│                                                              │
│  ...                                                         │
└──────────────────────────────────────────────────────────────┘
```

#### API 키 변경 플로우

```
관리자가 [변경] 클릭
    → 키 입력 모달 열림 (기존 키는 마스킹 상태)
    → 새 키 입력 + [검증] 클릭
    → 백엔드에서 해당 프로바이더 API 호출로 키 유효성 검증
    → 검증 성공 → 암호화 저장 + 감사 로그 기록
    → 검증 실패 → 에러 메시지 ("유효하지 않은 API 키입니다")
```

#### 키 유효성 검증 방법

| 프로바이더 | 검증 방법 |
|-----------|----------|
| Anthropic | `POST /v1/messages` (minimal request, max_tokens=1) |
| Gemini | `POST /v1beta/models/gemini-2.0-flash:generateContent` (minimal) |
| Groq | `POST /openai/v1/chat/completions` (minimal) |
| OpenAI | `GET /v1/models` (list models) |
| DART | `GET /api/corpCode.xml` (기업 코드 조회) |
| Naver | `GET /v1/search/news.json?query=test` |

#### 보안 요구사항

- API 키는 **AES-256-GCM으로 암호화** 후 DB에 저장
- 암호화 키(`SYSTEM_CONFIG_ENCRYPTION_KEY`)는 환경변수로 관리 (이것만 서버에서 직접 관리)
- 프론트엔드에 키 전체값이 절대 노출되지 않음 (마스킹: 앞 3자 + ••• + 뒤 3자)
- 키 조회 API는 마스킹된 값만 반환
- 키 변경 시 감사 로그 필수 기록

---

### 4.2 AI 모델 설정

관리자가 각 AI 파이프라인에서 사용할 모델을 선택하고 파라미터를 조정할 수 있다.

#### 설정 항목

| Config Key | 현재 기본값 | 선택 옵션 | 설명 |
|------------|-----------|-----------|------|
| `model.heavy_provider` | `anthropic` | `anthropic` | Heavy AI 프로바이더 |
| `model.heavy_model` | `claude-sonnet-4-5-20250929` | Claude 모델 목록 | Heavy AI 모델 |
| `model.light_provider` | `gemini` | `gemini`, `groq` | Light AI 프로바이더 |
| `model.light_model_gemini` | `gemini-2.0-flash` | Gemini 모델 목록 | Gemini 모델 |
| `model.light_model_groq` | `llama-3.3-70b-versatile` | Groq 모델 목록 | Groq 모델 |
| `model.embedding_model` | `text-embedding-3-small` | OpenAI 임베딩 모델 | 임베딩 모델 |

#### UI 와이어프레임

```
┌──────────────────────────────────────────────────────────────┐
│  시스템 설정                                                   │
│  [API Keys] [AI Models] [Limits]                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  AI 모델 설정                                                 │
│  ────────────────────────────────────────────────────────    │
│                                                              │
│  Heavy AI (기업 분석, 코칭, 첨삭)                                │
│  ┌────────────────────────────────────────────────────┐      │
│  │ 프로바이더:  Anthropic ▾                             │      │
│  │ 모델:       claude-sonnet-4-5-20250929 ▾            │      │
│  └────────────────────────────────────────────────────┘      │
│                                                              │
│  Light AI (무기 태깅, 문항 분석)                                 │
│  ┌────────────────────────────────────────────────────┐      │
│  │ 프로바이더:  ○ Gemini  ● Groq                        │      │
│  │ Gemini 모델: gemini-2.0-flash ▾                     │      │
│  │ Groq 모델:  llama-3.3-70b-versatile ▾              │      │
│  └────────────────────────────────────────────────────┘      │
│                                                              │
│  임베딩                                                       │
│  ┌────────────────────────────────────────────────────┐      │
│  │ 모델: text-embedding-3-small ▾                      │      │
│  │ 차원: 1536                                          │      │
│  └────────────────────────────────────────────────────┘      │
│                                                              │
│                                            [저장]            │
│                                                              │
│  ⚠️ 모델 변경 시 즉시 적용됩니다.                                │
│     프롬프트별 temperature/max_tokens는 프롬프트 관리에서         │
│     조정하세요.                                                │
└──────────────────────────────────────────────────────────────┘
```

#### 모델 변경 규칙

- **즉시 적용**: 설정 변경 후 새 요청부터 변경된 모델 사용
- **프롬프트별 파라미터**: temperature, max_tokens 등은 프롬프트 관리 페이지에서 관리 (이미 구현됨)
- **모델 변경 시 감사 로그 기록**

---

### 4.3 사용량 한도 설정

Fair Use Policy의 플랜별 한도를 어드민에서 조정할 수 있다.

#### 설정 항목

| Config Key | 기본값 | 설명 |
|------------|-------|------|
| `limit.free.experiences` | 3 | 무료 플랜 경험 등록 한도 |
| `limit.free.analysis` | 1 | 무료 플랜 기업 분석 한도 |
| `limit.free.coaching` | 1 | 무료 플랜 AI 코칭 한도 |
| `limit.free.interview` | 1 | 무료 플랜 AI 인터뷰 한도 |
| `limit.free.daily_analysis` | 1 | 무료 플랜 일일 분석 한도 |
| `limit.free.daily_coaching` | 1 | 무료 플랜 일일 코칭 한도 |
| `limit.starter.analysis` | 10 | 스타터 플랜 기업 분석 한도 |
| `limit.starter.coaching` | 10 | 스타터 플랜 AI 코칭 한도 |
| `limit.starter.interview` | 5 | 스타터 플랜 AI 인터뷰 한도 |
| `limit.starter.daily_analysis` | 5 | 스타터 플랜 일일 분석 한도 |
| `limit.starter.daily_coaching` | 5 | 스타터 플랜 일일 코칭 한도 |
| `limit.pro.analysis` | 30 | 프로 플랜 기업 분석 한도 |
| `limit.pro.coaching` | 30 | 프로 플랜 AI 코칭 한도 |
| `limit.pro.interview` | 15 | 프로 플랜 AI 인터뷰 한도 |
| `limit.pro.daily_analysis` | 10 | 프로 플랜 일일 분석 한도 |
| `limit.pro.daily_coaching` | 10 | 프로 플랜 일일 코칭 한도 |
| `limit.season.daily_analysis` | 20 | 시즌패스 일일 분석 한도 |
| `limit.season.daily_coaching` | 20 | 시즌패스 일일 코칭 한도 |

#### UI 와이어프레임

```
┌──────────────────────────────────────────────────────────────┐
│  시스템 설정                                                   │
│  [API Keys] [AI Models] [Limits]                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  플랜별 사용 한도 설정                                           │
│  ────────────────────────────────────────────────────────    │
│                                                              │
│  ┌───────────┬────────┬──────────┬────────┬──────────┐      │
│  │ 항목       │ 무료   │ 스타터    │ 프로    │ 시즌패스   │      │
│  ├───────────┼────────┼──────────┼────────┼──────────┤      │
│  │ 경험 등록   │ [3]   │ 무제한    │ 무제한   │ 무제한    │      │
│  │ 기업 분석   │ [1]   │ [10]    │ [30]   │ 무제한    │      │
│  │ AI 코칭    │ [1]   │ [10]    │ [30]   │ 무제한    │      │
│  │ AI 인터뷰  │ [1]   │ [5]     │ [15]   │ 무제한    │      │
│  │ 일일 분석   │ [1]   │ [5]     │ [10]   │ [20]    │      │
│  │ 일일 코칭   │ [1]   │ [5]     │ [10]   │ [20]    │      │
│  └───────────┴────────┴──────────┴────────┴──────────┘      │
│                                                              │
│                                            [저장]            │
│  ⚠️ 변경 시 즉시 적용됩니다. 기존 사용자의 누적 사용량은            │
│     변하지 않습니다.                                            │
└──────────────────────────────────────────────────────────────┘
```

---

## 5. 사용량 모니터링 (`/admin/usage`)

### 5.1 개요

서비스 전체의 AI API 사용량과 비용을 실시간으로 모니터링한다.

### 5.2 모니터링 대상

| 메트릭 | 단위 | 소스 |
|--------|------|------|
| AI 호출 횟수 | 건/일 | `usage_logs` 테이블 |
| 토큰 사용량 | 토큰/일 | AI 응답의 usage 필드 |
| 추정 비용 | 원/일 | 토큰 × 단가 계산 |
| 기능별 사용량 | 건/기능/일 | feature 필드 기준 |
| 사용자별 사용량 | 건/사용자/일 | user_id 기준 |

### 5.3 데이터 모델

```sql
-- usage_logs 테이블 (확장)
-- 기존 usage_tracking에 비용 추적 필드 추가
CREATE TABLE usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES user_profiles(id),
    feature VARCHAR(50) NOT NULL,           -- "analysis", "coaching", "interview", "tagging"
    provider VARCHAR(20) NOT NULL,          -- "anthropic", "gemini", "groq", "openai"
    model VARCHAR(100) NOT NULL,            -- "claude-sonnet-4-5-20250929"
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    estimated_cost_krw DECIMAL(10,2),       -- 추정 비용 (원화)
    latency_ms INTEGER,                     -- 응답 시간
    status VARCHAR(20) DEFAULT 'success',   -- "success", "error", "timeout"
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 인덱스
CREATE INDEX idx_usage_logs_created_at ON usage_logs(created_at);
CREATE INDEX idx_usage_logs_user_id ON usage_logs(user_id);
CREATE INDEX idx_usage_logs_feature ON usage_logs(feature);
```

### 5.4 UI 와이어프레임

```
┌──────────────────────────────────────────────────────────────┐
│  사용량 모니터링                     기간: [최근 7일 ▾] [조회]   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  요약 카드                                                    │
│  ┌──────────┬──────────┬──────────┬──────────┐              │
│  │ 총 호출   │ 총 토큰   │ 추정 비용  │ 오류율    │              │
│  │ 1,234건  │ 2.3M     │ ₩45,600  │ 0.3%    │              │
│  └──────────┴──────────┴──────────┴──────────┘              │
│                                                              │
│  일별 사용량 차트                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  ■ 기업분석  ■ 코칭  ■ 인터뷰  ■ 태깅                  │    │
│  │  ╷    ╷                                              │    │
│  │  │ ╷  │ ╷                                            │    │
│  │  ████ ████ ████ ████ ████ ████ ████                  │    │
│  │  2/9  2/10 2/11 2/12 2/13 2/14 2/15                 │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              │
│  프로바이더별 비용                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Anthropic  ████████████████████████  ₩38,400 (84%)   │    │
│  │ OpenAI     ████                      ₩4,200  (9%)    │    │
│  │ Gemini     ███                       ₩2,100  (5%)    │    │
│  │ Groq       █                         ₩900    (2%)    │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              │
│  상위 사용자 (토큰 기준)                                        │
│  ┌─────┬──────────┬───────┬────────┬──────────┐            │
│  │ #   │ 사용자     │ 호출수 │ 토큰    │ 추정 비용  │            │
│  ├─────┼──────────┼───────┼────────┼──────────┤            │
│  │ 1   │ user@... │ 45    │ 234K   │ ₩5,600   │            │
│  │ 2   │ foo@...  │ 38    │ 198K   │ ₩4,200   │            │
│  │ ...                                                      │
│  └─────┴──────────┴───────┴────────┴──────────┘            │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 5.5 비용 단가 테이블

| 프로바이더 | 모델 | Input (1M 토큰) | Output (1M 토큰) | 원화 환산 (1 USD = ₩1,400) |
|-----------|------|-----------------|------------------|--------------------------|
| Anthropic | Claude Sonnet 4.5 | $3.00 | $15.00 | ₩4,200 / ₩21,000 |
| Google | Gemini 2.0 Flash | $0.10 | $0.40 | ₩140 / ₩560 |
| Groq | Llama 3.3 70B | $0.59 | $0.79 | ₩826 / ₩1,106 |
| OpenAI | text-embedding-3-small | $0.02 | - | ₩28 |

> 단가는 `system_configs`에 `cost.anthropic_input`, `cost.anthropic_output` 등으로 저장하여 어드민에서 수정 가능

---

## 6. 감사 로그 (`/admin/logs`)

### 6.1 목적

관리자의 모든 시스템 설정 변경 사항을 기록하여 추적한다.

### 6.2 데이터 모델

```sql
CREATE TABLE admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES user_profiles(id),
    action VARCHAR(50) NOT NULL,              -- "update_api_key", "update_model", "update_role", "update_prompt", "update_limit"
    target_type VARCHAR(50) NOT NULL,         -- "system_config", "user", "prompt_template"
    target_id VARCHAR(255),                   -- 대상 ID
    old_value TEXT,                            -- 이전 값 (API 키는 마스킹)
    new_value TEXT,                            -- 새 값 (API 키는 마스킹)
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_admin_id ON admin_audit_logs(admin_id);
CREATE INDEX idx_audit_logs_created_at ON admin_audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON admin_audit_logs(action);
```

### 6.3 UI 와이어프레임

```
┌──────────────────────────────────────────────────────────────┐
│  감사 로그                           [액션 필터: 전체 ▾]        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┬──────────┬──────────┬────────────────┐ │
│  │ 시간             │ 관리자    │ 액션      │ 상세            │ │
│  ├─────────────────┼──────────┼──────────┼────────────────┤ │
│  │ 02-15 14:30:22  │ 관리자    │ 모델 변경  │ light_provider │ │
│  │                 │          │          │ gemini → groq  │ │
│  ├─────────────────┼──────────┼──────────┼────────────────┤ │
│  │ 02-15 11:15:03  │ 관리자    │ 키 변경   │ api_key.       │ │
│  │                 │          │          │ anthropic      │ │
│  │                 │          │          │ sk-•••abc →    │ │
│  │                 │          │          │ sk-•••xyz      │ │
│  ├─────────────────┼──────────┼──────────┼────────────────┤ │
│  │ 02-14 09:45:11  │ 관리자    │ 역할 변경  │ user@test.com │ │
│  │                 │          │          │ user → admin   │ │
│  └─────────────────┴──────────┴──────────┴────────────────┘ │
│                                                              │
│  < 1 2 3 >                                                   │
└──────────────────────────────────────────────────────────────┘
```

---

## 7. API 설계

### 7.1 역할 변경 API (확장)

```
# 사용자 역할 변경 (기존 API 확장)
PUT /v1/admin/users/{id}/role
← { "role": "manager" }  // "user" | "manager" | "admin" | "super_admin"
→ 200 { "data": { "id": "...", "role": "manager", ... } }

# 권한 검증 규칙 (백엔드)
# - 요청자 role이 대상의 새 role보다 높아야 함
# - admin은 user↔manager만 변경 가능
# - super_admin만 admin/super_admin 승격 가능
# - 자기 자신 변경 불가
# - 마지막 super_admin 강등 불가

→ 403 { "error": { "code": "AUTH_003", "message": "해당 역할로 변경할 권한이 없습니다." } }
→ 400 { "error": { "code": "AUTH_010", "message": "자기 자신의 역할은 변경할 수 없습니다." } }
→ 409 { "error": { "code": "AUTH_011", "message": "마지막 슈퍼 어드민은 강등할 수 없습니다." } }
```

### 7.2 시스템 설정 API

```
# 설정 목록 조회
GET /v1/admin/configs?category=api_key
→ 200 { "data": [{ "config_key": "api_key.anthropic", "masked_value": "sk-ant-•••abc", "description": "...", "category": "api_key", "is_secret": true, "updated_at": "..." }] }

# 설정 값 변경
PUT /v1/admin/configs/{config_key}
← { "value": "new-api-key-value" }
→ 200 { "data": { "config_key": "api_key.anthropic", "masked_value": "sk-ant-•••xyz", "updated_at": "..." } }

# API 키 유효성 검증
POST /v1/admin/configs/validate-key
← { "provider": "anthropic", "key": "sk-ant-..." }
→ 200 { "data": { "valid": true, "provider": "anthropic" } }
→ 400 { "error": { "code": "ADMIN_001", "message": "유효하지 않은 API 키입니다" } }
```

### 7.3 사용량 모니터링 API

```
# 사용량 요약
GET /v1/admin/usage/summary?period=7d
→ 200 {
    "data": {
      "total_calls": 1234,
      "total_tokens": 2300000,
      "estimated_cost_krw": 45600,
      "error_rate": 0.003
    }
  }

# 일별 사용량
GET /v1/admin/usage/daily?period=7d
→ 200 {
    "data": [
      { "date": "2026-02-15", "analysis": 45, "coaching": 38, "interview": 12, "tagging": 89 },
      ...
    ]
  }

# 프로바이더별 비용
GET /v1/admin/usage/costs?period=7d
→ 200 {
    "data": [
      { "provider": "anthropic", "cost_krw": 38400, "percentage": 84 },
      ...
    ]
  }

# 상위 사용자
GET /v1/admin/usage/top-users?period=7d&limit=10
→ 200 {
    "data": [
      { "user_id": "...", "email": "...", "calls": 45, "tokens": 234000, "cost_krw": 5600 },
      ...
    ]
  }
```

### 7.4 감사 로그 API

```
# 감사 로그 목록
GET /v1/admin/audit-logs?action=update_api_key&limit=20&offset=0
→ 200 {
    "data": [...],
    "total": 150
  }
```

---

## 8. 에러 코드

| 코드 | HTTP | 설명 |
|------|------|------|
| `AUTH_003` | 403 | 접근 권한 부족 (역할 미달) |
| `AUTH_010` | 400 | 자기 자신의 역할 변경 불가 |
| `AUTH_011` | 409 | 마지막 슈퍼 어드민 강등 불가 |
| `AUTH_012` | 403 | 해당 역할로 변경할 권한 없음 |
| `ADMIN_001` | 400 | 유효하지 않은 API 키 |
| `ADMIN_002` | 400 | 지원하지 않는 프로바이더 |
| `ADMIN_003` | 400 | 유효하지 않은 설정 값 |
| `ADMIN_004` | 404 | 설정 키를 찾을 수 없음 |
| `ADMIN_005` | 409 | 설정 값 충돌 (동시 수정) |

---

## 9. 구현 우선순위

### Phase 0: 역할 계층 (기반 작업)

1. Ent 스키마 `role` enum 확장: `user`, `manager`, `admin`, `super_admin`
2. `RequireRole(minRole)` 미들웨어 구현 (기존 `AdminMiddleware` 대체)
3. 프론트엔드 `auth-store.ts` role 타입 확장
4. `hasRole()`, `isManager()`, `isAdmin()`, `isSuperAdmin()` 유틸리티
5. `AdminGuard` 확장 (`requiredRole` prop)
6. 사이드바 메뉴 역할별 필터링
7. 사용자 관리 페이지 역할 변경 UI (드롭다운 + 권한 체크)
8. 라우트별 최소 역할 적용
9. 초기 슈퍼 어드민 시드 스크립트 수정

### Phase 1: 시스템 설정 (핵심, super_admin 전용)

1. `system_configs` 테이블 + Ent 스키마
2. API 키 관리 API (CRUD + 유효성 검증)
3. AI 모델 설정 API
4. 시스템 설정 프론트엔드 (`/admin/settings`)
5. 사이드바 메뉴 확장

### Phase 2: 모니터링 (manager 이상 읽기)

1. `usage_logs` 테이블 확장 (토큰/비용 필드)
2. AI 호출 시 usage_logs 자동 기록 미들웨어
3. 사용량 모니터링 API
4. 사용량 모니터링 프론트엔드 (`/admin/usage`)

### Phase 3: 사용자 관리 확장 (admin 이상)

1. 계정 정지/해제 (`suspended` 필드, 인증 미들웨어 체크)
2. 강제 로그아웃 (refresh_token 무효화)
3. 사용자 상세 모달 (경험 수, 코칭 수, 사용량)
4. 플랜 수동 변경 API + UI

### Phase 4: 감사 & 한도 (admin 이상)

1. `admin_audit_logs` 테이블
2. 감사 로그 자동 기록 미들웨어
3. 사용량 한도 설정 API + UI (admin 이상)
4. 감사 로그 프론트엔드 (`/admin/logs`, manager 이상 읽기)

### Phase 5: 시스템 헬스 & 작업 모니터링

1. 시스템 헬스 체크 API (DB, AI 프로바이더, River)
2. 대시보드 시스템 상태 카드
3. River 작업 큐 모니터링 API + UI
4. 실패 작업 재시도/삭제

### Phase 6: 피드백 & 크롤러 & 프라이버시

1. 피드백 관리 UI (`admin_status`, `admin_note` 필드 확장)
2. 크롤러 모니터링 API + UI
3. 사용자 데이터 내보내기 (PIPA 준수)
4. 계정 삭제 요청 큐 관리

---

## 10. 기술 결정 사항

### 10.1 설정값 로딩 전략

```
서버 시작 시:
  1. 환경변수에서 기본값 로드
  2. DB system_configs에서 오버라이드 값 로드
  3. 메모리 캐시에 저장 (5분 TTL)

설정 변경 시:
  1. DB 업데이트
  2. 메모리 캐시 즉시 무효화
  3. 다음 요청부터 새 값 사용
```

### 10.2 환경변수와 DB 설정의 관계

| 우선순위 | 소스 | 설명 |
|---------|------|------|
| 1 (최우선) | DB `system_configs` | 어드민에서 변경한 값 |
| 2 (폴백) | 환경변수 (`.env`) | 초기 설정값 |
| 3 (기본) | 코드 내 하드코딩 기본값 | 최종 폴백 |

> DB에 값이 없으면 환경변수 → 기본값 순으로 폴백. 어드민에서 한 번이라도 설정하면 DB 값이 우선.

### 10.3 암호화

```go
// internal/infrastructure/crypto/config_crypto.go
type ConfigCrypto struct {
    key []byte // SYSTEM_CONFIG_ENCRYPTION_KEY (32 bytes for AES-256)
}

func (c *ConfigCrypto) Encrypt(plaintext string) (string, error)
func (c *ConfigCrypto) Decrypt(ciphertext string) (string, error)
func (c *ConfigCrypto) Mask(value string) string // "sk-ant-•••abc"
```

---

## 11. 사용자 관리 확장 (기존 페이지 보강)

현재 사용자 관리 페이지에 추가할 기능:

| 기능 | 설명 | 우선순위 |
|------|------|---------|
| 사용자 상세 모달 | 경험 수, 코칭 수, 마지막 활동, 플랜 정보 | 높음 |
| 사용량 이력 | 해당 사용자의 기능별 사용 이력 | 중간 |
| 계정 정지/해제 | 어뷰징 사용자 일시 정지 (soft-ban) | 높음 |
| 강제 로그아웃 | 계정 침해 시 모든 세션 무효화 | 높음 |
| 플랜 수동 변경 | 관리자가 사용자 플랜을 수동 조정 | 중간 |
| 사용자 데이터 내보내기 | PIPA 준수, CSV 형식 | 중간 |
| 계정 삭제 처리 | 탈퇴 요청 큐 관리, 30일 보존 후 영구 삭제 | 중간 |

### 11.1 계정 정지 (Soft-Ban)

```sql
-- user_profiles에 필드 추가
ALTER TABLE user_profiles ADD COLUMN suspended BOOLEAN DEFAULT false;
ALTER TABLE user_profiles ADD COLUMN suspended_at TIMESTAMP;
ALTER TABLE user_profiles ADD COLUMN suspended_reason VARCHAR(500);
```

```go
// 인증 미들웨어에서 정지 체크
func AuthMiddleware(...) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... JWT 검증 후
        if user.Suspended {
            c.AbortWithStatusJSON(403, gin.H{
                "error": gin.H{
                    "code": "AUTH_013",
                    "message": "계정이 정지되었습니다.",
                    "suspended_reason": user.SuspendedReason,
                },
            })
            return
        }
    }
}
```

### 11.2 강제 로그아웃

```
POST /v1/admin/users/{id}/force-logout
→ 해당 사용자의 모든 refresh_token 무효화
→ 다음 토큰 갱신 시 자동 로그아웃
```

---

## 12. 시스템 헬스 모니터링 (`/admin` 대시보드 확장)

### 12.1 시스템 상태 카드

기존 대시보드에 시스템 상태 섹션을 추가한다.

```
┌──────────────────────────────────────────────────────────────┐
│  시스템 상태                                                   │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ 🟢 DB        │ 🟢 Anthropic │ 🟡 Gemini    │ 🟢 River   │ │
│  │ 응답: 3ms    │ 응답: 450ms  │ 응답: 1200ms │ 대기: 2건   │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
│                                                              │
│  최근 에러율: 0.3% (24h)  │  평균 응답: 230ms  │  가동률: 99.8% │
└──────────────────────────────────────────────────────────────┘
```

### 12.2 헬스 체크 대상

| 대상 | 체크 방법 | 정상 기준 | 주기 |
|------|----------|----------|------|
| PostgreSQL | `SELECT 1` | < 100ms | 30초 |
| Anthropic API | 키 유효성 체크 | < 3000ms | 5분 |
| Gemini API | 키 유효성 체크 | < 3000ms | 5분 |
| Groq API | 키 유효성 체크 | < 3000ms | 5분 |
| OpenAI API | `GET /v1/models` | < 3000ms | 5분 |
| River (Job Queue) | 대기 작업 수 조회 | 대기 < 100건 | 1분 |
| DART API | 간단 조회 | < 5000ms | 10분 |

### 12.3 API

```
GET /v1/admin/health
→ 200 {
    "data": {
      "status": "healthy",           // "healthy" | "degraded" | "unhealthy"
      "checks": {
        "database": { "status": "up", "latency_ms": 3 },
        "anthropic": { "status": "up", "latency_ms": 450 },
        "gemini": { "status": "degraded", "latency_ms": 1200 },
        "groq": { "status": "up", "latency_ms": 280 },
        "openai": { "status": "up", "latency_ms": 320 },
        "river": { "status": "up", "pending_jobs": 2, "failed_jobs": 0 }
      },
      "error_rate_24h": 0.003,
      "avg_response_ms": 230,
      "uptime_percentage": 99.8
    }
  }
```

---

## 13. 백그라운드 작업 모니터링

### 13.1 River Job Queue 관리

River (Go 내장 작업 큐)의 상태를 어드민에서 모니터링하고 관리한다.

```
┌──────────────────────────────────────────────────────────────┐
│  백그라운드 작업                                                │
│  ┌──────────┬──────────┬──────────┬──────────┐              │
│  │ 대기 중    │ 처리 중    │ 완료      │ 실패      │              │
│  │    5     │    2     │  1,234   │    3     │              │
│  └──────────┴──────────┴──────────┴──────────┘              │
│                                                              │
│  실패한 작업                                                   │
│  ┌──────────┬──────────────────┬──────────┬────────────────┐│
│  │ 작업 유형  │ 에러 메시지        │ 시도 횟수  │ 액션           ││
│  ├──────────┼──────────────────┼──────────┼────────────────┤│
│  │ 태깅     │ API rate limit   │ 3/5     │ [재시도] [삭제]  ││
│  │ 분석     │ timeout          │ 5/5     │ [재시도] [삭제]  ││
│  └──────────┴──────────────────┴──────────┴────────────────┘│
└──────────────────────────────────────────────────────────────┘
```

### 13.2 API

```
# 작업 상태 요약
GET /v1/admin/jobs/summary
→ 200 { "data": { "pending": 5, "running": 2, "completed": 1234, "failed": 3 } }

# 실패한 작업 목록
GET /v1/admin/jobs/failed?limit=20
→ 200 { "data": [{ "id": "...", "kind": "weapon_tagging", "error": "...", "attempts": 3, "created_at": "..." }] }

# 작업 재시도
POST /v1/admin/jobs/{id}/retry
→ 200 { "data": { "id": "...", "status": "pending" } }

# 작업 삭제
DELETE /v1/admin/jobs/{id}
→ 204
```

---

## 14. 피드백 관리

### 14.1 사용자 피드백 조회/응답

`feedback` 테이블에 저장된 사용자 피드백을 어드민에서 조회하고 관리한다.

```
┌──────────────────────────────────────────────────────────────┐
│  피드백 관리                    [상태: 전체 ▾] [유형: 전체 ▾]    │
├──────────────────────────────────────────────────────────────┤
│  ┌─────────┬──────────┬──────────┬──────────┬─────────────┐ │
│  │ 날짜     │ 사용자    │ 유형      │ 내용      │ 상태         │ │
│  ├─────────┼──────────┼──────────┼──────────┼─────────────┤ │
│  │ 02-15   │ user@..  │ 버그      │ 코칭이... │ 🟡 확인 중   │ │
│  │ 02-14   │ foo@..   │ 기능 요청  │ 다크모... │ ⬜ 미확인    │ │
│  │ 02-13   │ bar@..   │ 칭찬      │ 좋아요!  │ ✅ 처리 완료  │ │
│  └─────────┴──────────┴──────────┴──────────┴─────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### 14.2 데이터 모델 변경

```sql
-- feedback 테이블에 어드민 관리 필드 추가
ALTER TABLE feedbacks ADD COLUMN admin_status VARCHAR(20) DEFAULT 'pending';
  -- "pending" | "reviewing" | "resolved" | "dismissed"
ALTER TABLE feedbacks ADD COLUMN admin_note TEXT;
ALTER TABLE feedbacks ADD COLUMN reviewed_by UUID REFERENCES user_profiles(id);
ALTER TABLE feedbacks ADD COLUMN reviewed_at TIMESTAMP;
```

### 14.3 API

```
# 피드백 목록
GET /v1/admin/feedbacks?status=pending&type=bug&limit=20
→ 200 { "data": [...], "total": 45 }

# 피드백 상태 변경
PUT /v1/admin/feedbacks/{id}
← { "admin_status": "resolved", "admin_note": "v2.1에서 수정 예정" }
→ 200 { "data": { ... } }
```

---

## 15. 크롤러 모니터링

> 상세 설계: [docs/develop/14-crawler-monitoring.md](../develop/14-crawler-monitoring.md)

어드민에서 크롤러의 건강 상태와 실패 이력을 모니터링한다.

| 메트릭 | 설명 |
|--------|------|
| 사이트별 성공률 | 최근 7일 기준 크롤링 성공/실패 비율 |
| 평균 크롤링 시간 | 사이트별 평균 소요 시간 |
| 최근 실패 로그 | 실패 URL, 에러 메시지, 시도 횟수 |
| HTML 변경 감지 | 파싱 규칙 업데이트 필요 여부 |

### API

```
GET /v1/admin/crawler/status
GET /v1/admin/crawler/failures?limit=50
POST /v1/admin/crawler/health-check/:domain
```

---

## 16. 데이터 프라이버시 관리

### 16.1 사용자 데이터 내보내기

PIPA(개인정보보호법) 준수를 위한 데이터 내보내기 기능.

```
# 사용자 데이터 내보내기 요청 생성
POST /v1/admin/users/{id}/export
→ 202 { "data": { "export_id": "...", "status": "processing" } }

# 내보내기 상태 확인
GET /v1/admin/exports/{export_id}
→ 200 { "data": { "status": "completed", "download_url": "...", "expires_at": "..." } }
```

### 16.2 계정 삭제 처리

```
# 계정 삭제 요청 (30일 보존 기간)
POST /v1/admin/users/{id}/delete-request
← { "reason": "사용자 탈퇴 요청" }
→ 200 { "data": { "scheduled_deletion_at": "2026-03-17T00:00:00Z" } }

# 삭제 대기 목록
GET /v1/admin/deletion-queue
→ 200 { "data": [{ "user_id": "...", "email": "...", "scheduled_at": "...", "reason": "..." }] }

# 삭제 취소 (보존 기간 내)
DELETE /v1/admin/users/{id}/delete-request
→ 204
```

---

## 17. 비기능 요구사항

| 항목 | 요구사항 |
|------|---------|
| 보안 | API 키 AES-256-GCM 암호화, 마스킹 표시, HTTPS 필수 |
| 접근 제어 | 역할 계층 (user < manager < admin < super_admin), RequireRole 미들웨어 |
| 감사 | 모든 설정 변경 감사 로그 기록, 민감 정보 마스킹 |
| 가용성 | 설정 캐시로 DB 장애 시에도 기존 설정 유지 |
| 성능 | 설정 조회 < 100ms (캐시), 사용량 집계 < 3s |
| 프라이버시 | PIPA 준수, 데이터 내보내기, 30일 보존 후 영구 삭제 |
| 모니터링 | 시스템 헬스 체크 30초~10분 주기, 장애 시 알림 |
