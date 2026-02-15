# Phase 1.4: 프론트엔드 인증

## 목표

Next.js에서 로그인/회원가입 페이지, 인증 상태 관리 (Zustand), 보호된 라우트 미들웨어를 구현한다.

---

## 디렉토리 구조

```text
apps/web/src/
├── app/
│   ├── (auth)/
│   │   ├── layout.tsx          # 인증 페이지 레이아웃 (센터 카드)
│   │   ├── login/page.tsx      # 로그인 페이지
│   │   └── signup/page.tsx     # 회원가입 페이지
│   └── auth/
│       └── callback/page.tsx   # OAuth 콜백 처리 (token 수신)
├── lib/
│   ├── api-client.ts           # Axios 인스턴스 + 토큰 자동 첨부
│   └── auth.ts                 # 인증 유틸 함수
├── stores/
│   └── auth-store.ts           # Zustand 인증 스토어
├── hooks/
│   └── use-auth.ts             # 인증 관련 React hooks
└── middleware.ts               # Next.js 미들웨어 (보호 라우트)
```

---

## A. 로그인 페이지

### `src/app/(auth)/login/page.tsx`

```
┌─────────────────────────────┐
│                             │
│         Colight 로고         │
│    AI 자소서 코칭 플랫폼      │
│                             │
│  ┌─── Naver로 시작하기 ───┐  │  ← Naver 초록 버튼
│  └───────────────────────┘  │
│                             │
│  ─────── 또는 ────────      │
│                             │
│  이메일                      │
│  ┌───────────────────────┐  │
│  │                       │  │
│  └───────────────────────┘  │
│  비밀번호                    │
│  ┌───────────────────────┐  │
│  │                       │  │
│  └───────────────────────┘  │
│                             │
│  ┌────── 로그인 ─────────┐  │
│  └───────────────────────┘  │
│                             │
│  계정이 없으신가요? 회원가입   │
│                             │
└─────────────────────────────┘
```

### 기능 요구사항

- React Hook Form + Zod 유효성 검증
  - 이메일: 필수, 이메일 형식
  - 비밀번호: 필수, 최소 8자
- Naver 로그인 버튼: `window.location.href = API_URL + '/v1/auth/naver/login'`
- Email 로그인: POST `/v1/auth/login` → 토큰 저장 → 리디렉션
- 로그인 성공: `/(main)/experiences` 로 이동
- 로그인 실패: 인라인 에러 메시지 (sonner toast)

---

## B. 회원가입 페이지

### `src/app/(auth)/signup/page.tsx`

- React Hook Form + Zod
  - 이메일: 필수, 이메일 형식
  - 비밀번호: 필수, 최소 8자, 영문+숫자 조합
  - 비밀번호 확인: 비밀번호와 일치
  - 닉네임: 선택, 최대 50자
- POST `/v1/auth/signup` → 토큰 저장 → `/(main)/experiences` 이동

---

## C. OAuth 콜백 페이지

### `src/app/auth/callback/page.tsx`

Naver OAuth 완료 후 Go 백엔드가 리디렉션하는 페이지.

```
Go Backend → 302 Redirect → /auth/callback?access_token=xxx&refresh_token=yyy&expires_in=3600
```

```typescript
// src/app/auth/callback/page.tsx
"use client";

import { useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export default function AuthCallbackPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { setTokens, fetchUser } = useAuthStore();

  useEffect(() => {
    const accessToken = searchParams.get("access_token");
    const refreshToken = searchParams.get("refresh_token");

    if (accessToken && refreshToken) {
      setTokens(accessToken, refreshToken);
      fetchUser().then(() => {
        router.replace("/experiences");
      });
    } else {
      router.replace("/login?error=auth_failed");
    }
  }, []);

  return <div>로그인 처리 중...</div>;
}
```

---

## D. 인증 상태 관리 (Zustand)

### `src/stores/auth-store.ts`

```typescript
import { create } from "zustand";
import { persist } from "zustand/middleware";
import { apiClient } from "@/lib/api-client";

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: UserInfo | null;
  isLoading: boolean;

  // Actions
  setTokens: (access: string, refresh: string) => void;
  fetchUser: () => Promise<void>;
  logout: () => void;
  refreshAccessToken: () => Promise<boolean>;
}

interface UserInfo {
  id: string;
  email?: string;
  nickname?: string;
  role: "user" | "admin";
  auth_provider: "email" | "naver";
  onboarding_completed: boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isLoading: false,

      setTokens: (access, refresh) => {
        set({ accessToken: access, refreshToken: refresh });
      },

      fetchUser: async () => {
        set({ isLoading: true });
        try {
          const { data } = await apiClient.get("/v1/auth/me");
          set({ user: data.data, isLoading: false });
        } catch {
          set({ user: null, isLoading: false });
        }
      },

      logout: () => {
        apiClient.post("/v1/auth/logout").catch(() => {});
        set({
          accessToken: null,
          refreshToken: null,
          user: null,
        });
      },

      refreshAccessToken: async () => {
        const { refreshToken } = get();
        if (!refreshToken) return false;

        try {
          const { data } = await apiClient.post("/v1/auth/refresh", {
            refresh_token: refreshToken,
          });
          set({
            accessToken: data.data.access_token,
            refreshToken: data.data.refresh_token,
          });
          return true;
        } catch {
          get().logout();
          return false;
        }
      },
    }),
    {
      name: "colight-auth",
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    }
  )
);
```

---

## E. API 클라이언트

### `src/lib/api-client.ts`

```typescript
import axios from "axios";
import { useAuthStore } from "@/stores/auth-store";

export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000",
  headers: { "Content-Type": "application/json" },
});

// Request interceptor — 토큰 자동 첨부
apiClient.interceptors.request.use((config) => {
  const { accessToken } = useAuthStore.getState();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

// Response interceptor — 401 시 토큰 갱신
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      const success = await useAuthStore.getState().refreshAccessToken();
      if (success) {
        const { accessToken } = useAuthStore.getState();
        originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        return apiClient(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);
```

---

## F. Next.js 미들웨어 (보호 라우트)

### `src/middleware.ts`

```typescript
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Note: 실제 토큰 검증은 클라이언트 사이드에서 처리
  // 미들웨어는 기본적인 라우트 보호만 담당
  // (서버 사이드에서 JWT 검증하려면 jose 라이브러리 필요)

  // Static files — skip
  if (pathname.startsWith("/_next") || pathname.includes(".")) {
    return NextResponse.next();
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};
```

> **Note**: JWT 토큰은 Zustand (localStorage)에 저장되므로 Next.js 서버 사이드 미들웨어에서
> 직접 검증이 어렵습니다. 보호 라우트는 클라이언트 사이드 `AuthGuard` 컴포넌트로 처리합니다.

### AuthGuard 컴포넌트

```typescript
// src/components/auth/auth-guard.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { accessToken, user, isLoading, fetchUser } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!accessToken) {
      router.replace("/login");
      return;
    }
    if (!user && !isLoading) {
      fetchUser();
    }
  }, [accessToken]);

  if (!accessToken || isLoading) {
    return <LoadingSpinner />;
  }

  return <>{children}</>;
}

// src/components/auth/admin-guard.tsx
export function AdminGuard({ children }: { children: React.ReactNode }) {
  const { user } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (user && user.role !== "admin") {
      router.replace("/experiences");
    }
  }, [user]);

  if (!user || user.role !== "admin") {
    return <LoadingSpinner />;
  }

  return <>{children}</>;
}
```

---

## 체크리스트

### 페이지

- [x] `(auth)/layout.tsx` — 센터 카드 레이아웃, 로고
- [x] `(auth)/login/page.tsx` — Naver 로그인 + Email/PW 로그인
- [x] `(auth)/signup/page.tsx` — Email/PW 회원가입
- [x] `auth/callback/page.tsx` — OAuth 콜백 처리

### 상태 관리

- [x] `stores/auth-store.ts` — Zustand + persist
  - [x] `setTokens`, `fetchUser`, `logout`, `refreshAccessToken`
- [x] `lib/api-client.ts` — Axios + 토큰 자동 첨부 + 401 refresh

### 보호 라우트

- [x] `components/auth/auth-guard.tsx` — 인증 필요 라우트 보호
- [x] `components/auth/admin-guard.tsx` — 어드민 라우트 보호
- [x] `(main)/layout.tsx`에 `<AuthGuard>` 적용
- [x] `admin/layout.tsx`에 `<AdminGuard>` 적용 (Phase 1.5)

### UI 컴포넌트

- [x] Naver 로그인 버튼 (초록색, Naver 로고)
- [x] 이메일/비밀번호 폼 (shadcn Input + Label)
- [x] Zod 유효성 검증 스키마
- [x] 에러 메시지 표시 (sonner toast)
- [x] 로딩 상태 (버튼 Spinner)

---

## 검증 방법

1. Naver 로그인 버튼 클릭 → Naver 페이지 이동 → 로그인 → callback → /experiences
2. Email 회원가입 → 즉시 로그인 → /experiences
3. Email 로그인 → 성공 → /experiences
4. 비로그인 상태에서 /experiences 접속 → /login 리디렉션
5. 로그아웃 → 토큰 삭제 → /login 이동
6. 토큰 만료 → 자동 갱신 → API 호출 성공

---

## 산출물

- `src/app/(auth)/layout.tsx`
- `src/app/(auth)/login/page.tsx`
- `src/app/(auth)/signup/page.tsx`
- `src/app/auth/callback/page.tsx`
- `src/stores/auth-store.ts`
- `src/lib/api-client.ts`
- `src/components/auth/auth-guard.tsx`
- `src/components/auth/admin-guard.tsx`
