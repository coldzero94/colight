# Phase 1: 인증 & 레이아웃

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | Naver OAuth 2.0 자체 구현 + 이메일/비밀번호 인증, JWT 토큰 시스템, 어드민 시스템, 공통 레이아웃 구축 |
| **선행 조건** | Phase 0 완료 |
| **스프린트** | Sprint 0 (Day 5-10) |
| **관련 기능** | 전체 기능 (인증 기반) |
| **예상 공수** | 5-6일 |
| **산출물** | Naver OAuth 로그인, Email/PW 로그인, JWT 미들웨어, 어드민 시스템, 보호된 라우트, 사이드바+헤더 레이아웃 |

> **⚠️ 아키텍처 변경**: Supabase Auth 완전 제거. Go 백엔드에서 OAuth + JWT 직접 구현.
> - 인증 주체: Go 백엔드 (NOT Supabase, NOT NextAuth.js)
> - Primary 로그인: Naver OAuth 2.0
> - Secondary 로그인: Email/Password (이메일 인증 없음)
> - 향후 확장: Kakao/Google은 Phase 6.3에서 추가

---

## 인증 아키텍처

```
[Next.js Frontend]
  ├── "Naver로 시작하기" 버튼 클릭
  │     └─→ GET /v1/auth/naver/login (Go Backend)
  │           └─→ 302 Redirect → nid.naver.com/oauth2.0/authorize
  │                 └─→ Naver 로그인 완료
  │                       └─→ GET /v1/auth/naver/callback (Go Backend)
  │                             ├─→ Exchange code → access_token
  │                             ├─→ Fetch user info (openapi.naver.com/v1/nid/me)
  │                             ├─→ Find or create user_profiles
  │                             ├─→ Issue JWT (access + refresh)
  │                             └─→ 302 Redirect → frontend (/auth/callback?token=...)
  │
  ├── Email/Password 로그인
  │     └─→ POST /v1/auth/login (Go Backend)
  │           ├─→ bcrypt 비밀번호 검증
  │           ├─→ Issue JWT
  │           └─→ JSON response { access_token, refresh_token }
  │
  └── 인증된 API 호출
        └─→ Authorization: Bearer <access_token>
              └─→ Go JWT 미들웨어 → user_id + role 추출 → context
```

---

## Sub-Phase 문서

Phase 1은 6개 하위 문서로 분리됩니다:

| Step | 문서 | 이름 | 공수 |
|------|------|------|------|
| 1.1 | [phase-1.1-db-schema.md](phase-1.1-db-schema.md) | DB 스키마 수정 | 1h |
| 1.2 | [phase-1.2-api-definition.md](phase-1.2-api-definition.md) | Auth API 정의 (TypeSpec) | 2h |
| 1.3 | [phase-1.3-backend-auth.md](phase-1.3-backend-auth.md) | Go 백엔드 인증 (Naver OAuth + Email/PW + JWT) | 12h |
| 1.4 | [phase-1.4-frontend-auth.md](phase-1.4-frontend-auth.md) | 프론트엔드 인증 (로그인 UI + 상태관리) | 8h |
| 1.5 | [phase-1.5-admin.md](phase-1.5-admin.md) | 어드민 시스템 | 6h |
| 1.6 | [phase-1.6-layout.md](phase-1.6-layout.md) | 공통 레이아웃 (사이드바/헤더) | 6h |

**총 공수: ~35시간 (5-6일)**

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 1.1 | DB 스키마 수정 (auth + admin 필드) | ✅ |
| 1.2 | Auth + Admin API 정의 (TypeSpec) | ✅ |
| 1.3 | Go 백엔드 인증 (Naver OAuth + Email/PW + JWT) | ✅ |
| 1.4 | 프론트엔드 인증 (로그인 UI + 상태관리 + 보호 라우트) | ✅ |
| 1.5 | 어드민 시스템 (어드민 미들웨어 + 어드민 페이지) | ✅ |
| 1.6 | 공통 레이아웃 (사이드바/헤더 + 공통 컴포넌트) | ✅ |

---

## 기술 스택 추가

| 패키지 | 용도 | 위치 |
|--------|------|------|
| `github.com/golang-jwt/jwt/v5` | JWT 생성/검증 | Go Backend |
| `golang.org/x/crypto/bcrypt` | 비밀번호 해싱 | Go Backend |
| `golang.org/x/oauth2` | OAuth 2.0 클라이언트 | Go Backend |
| `zustand` | 인증 상태 관리 | Next.js Frontend |

---

## Naver OAuth 2.0 Endpoints

| 용도 | URL |
|------|-----|
| Authorization | `https://nid.naver.com/oauth2.0/authorize` |
| Token Exchange | `https://nid.naver.com/oauth2.0/token` |
| User Info | `https://openapi.naver.com/v1/nid/me` |

> 출처: [Logto - Naver OAuth Endpoints](https://logto.io/oauth-providers-explorer/naver), [Auth.js Naver Provider](https://authjs.dev/reference/core/providers/naver)

---

## 어드민 시스템 개요

- `user_profiles.role` 필드: `"user"` (기본) | `"admin"`
- 초기 어드민: DB 시드 또는 직접 SQL로 생성 (어드민 회원가입 페이지 없음)
- 어드민 미들웨어: Go에서 JWT의 role claim 확인
- 어드민 페이지: `/admin/*` 라우트 (사용자 관리, 프롬프트 관리, 시스템 현황)
- 어드민도 일반 사용자 기능 접근 가능

---

## Phase 완료 체크리스트

### 기능
- [x] Naver OAuth 로그인/회원가입 정상
- [x] Email/Password 회원가입/로그인 정상
- [x] JWT access token + refresh token 발급/갱신 정상
- [x] 비인증 사용자 → `/login` 리디렉션 정상
- [x] 인증 사용자 → `/(main)/*` 접근 정상
- [x] 어드민 사용자 → `/admin/*` 접근 정상
- [x] 비어드민 사용자 → `/admin/*` 접근 차단
- [x] 사이드바 네비게이션 정상
- [x] 모바일 햄버거 메뉴 정상
- [x] 로그아웃 → 토큰 삭제 → `/login` 리디렉션

### 검증 게이트
- [x] `moon run backend:lint` → 통과
- [x] `moon run backend:test` → 통과
- [x] `moon run web:lint` → 통과
- [x] `moon run web:typecheck` → 통과
- [x] `moon run web:build` → 빌드 성공

---

## 다음 Phase

**→ Phase 2: 경험 CRUD** (Sprint 1, Day 1-3)
- 경험 목록/등록/상세/수정/삭제 CRUD 구현
- STAR 구조 입력 폼
