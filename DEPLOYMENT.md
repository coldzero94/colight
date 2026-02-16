# Deployment Guide

Colight 배포 가이드 (무료 티어 전략)

## 아키텍처

```
Frontend (Next.js)  →  Vercel (무료)
Backend (Go)        →  Koyeb (무료)
Database            →  Supabase PostgreSQL (무료)
```

---

## 1. Supabase PostgreSQL 설정

### 1.1 프로젝트 생성
1. [Supabase](https://supabase.com) 로그인
2. "New Project" 클릭
3. 프로젝트 이름: `colight` (또는 원하는 이름)
4. Database Password 설정 (강력한 비밀번호 사용)
5. Region: `Northeast Asia (Seoul)` 또는 가까운 리전 선택

### 1.2 DATABASE_URL 획득
1. Supabase 프로젝트 대시보드 → Settings → Database
2. "Connection string" 섹션에서 **"URI"** 선택
3. `[YOUR-PASSWORD]`를 실제 비밀번호로 교체
4. 연결 문자열 예시:
   ```
   postgres://postgres.xxxxx:password@aws-0-ap-northeast-2.pooler.supabase.com:5432/postgres
   ```

### 1.3 마이그레이션 실행
로컬에서 DATABASE_URL을 Supabase로 설정하고 마이그레이션 실행:

```bash
# 환경변수 임시 설정
export DATABASE_URL="postgres://postgres.xxxxx:password@..."

# 마이그레이션 실행
moon run backend:migrate-up

# 초기 데이터 시딩 (선택적)
moon run backend:seed
```

---

## 2. Koyeb 백엔드 배포

### 2.1 프로젝트 생성
1. [Koyeb](https://www.koyeb.com) 로그인
2. "Create App" 클릭
3. Deployment method: **"GitHub"** 선택
4. Repository: `your-username/colight` 선택
5. Branch: `main`

### 2.2 빌드 설정
- **Builder**: `Docker`
- **Dockerfile path**: `apps/backend/Dockerfile`
- **Docker build context**: `apps/backend`
- **Port**: `9000`

### 2.3 환경변수 설정
Koyeb 대시보드에서 다음 환경변수 추가:

```bash
# Database
DATABASE_URL=postgres://postgres.xxxxx:password@aws-0-ap-northeast-2.pooler.supabase.com:5432/postgres

# JWT (IMPORTANT: 32자 이상의 강력한 랜덤 문자열 사용)
JWT_SECRET=your-production-jwt-secret-min-32-chars

# Naver OAuth
NAVER_CLIENT_ID=your-naver-client-id
NAVER_CLIENT_SECRET=your-naver-client-secret
NAVER_CALLBACK_URL=https://your-app.koyeb.app/v1/auth/naver/callback

# AI APIs
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
LLM_LIGHT_PROVIDER=gemini
GEMINI_API_KEY=...

# Frontend URL (CORS)
FRONTEND_URL=https://your-app.vercel.app

# Server
API_PORT=9000
```

### 2.4 배포
1. "Deploy" 클릭
2. 배포 완료 후 URL 확인: `https://your-app.koyeb.app`

---

## 3. Vercel 프론트엔드 배포

### 3.1 프로젝트 생성
1. [Vercel](https://vercel.com) 로그인
2. "Add New..." → "Project" 클릭
3. Repository: `your-username/colight` 선택
4. Import

### 3.2 빌드 설정

Vercel이 자동으로 Next.js를 감지하지만, pnpm workspace이므로 수동 설정 필요:

#### 옵션 1: vercel.json 사용 (권장)

`apps/web/vercel.json` 파일이 이미 설정되어 있습니다.

Vercel 대시보드 설정:

- **Framework Preset**: `Next.js`
- **Root Directory**: `apps/web`
- **Build Command**: 비워두기 (vercel.json 사용)
- **Install Command**: 비워두기 (vercel.json 사용)

#### 옵션 2: 수동 설정

- **Framework Preset**: `Next.js`
- **Root Directory**: `apps/web`
- **Build Command**: `cd ../.. && pnpm install && cd apps/web && pnpm run build`
- **Install Command**: `cd ../.. && pnpm install --frozen-lockfile`

### 3.3 환경변수 설정
Vercel 프로젝트 설정 → Environment Variables:

```bash
NEXT_PUBLIC_API_URL=https://your-app.koyeb.app
NEXT_PUBLIC_APP_URL=https://your-app.vercel.app
```

### 3.4 배포
1. "Deploy" 클릭
2. 배포 완료 후 URL 확인: `https://your-app.vercel.app`

---

## 4. Naver OAuth 설정

### 4.1 Naver Developers 콘솔
1. [Naver Developers](https://developers.naver.com) 로그인
2. "애플리케이션" → "애플리케이션 등록"
3. 애플리케이션 이름: `Colight`
4. 사용 API: "네이버 로그인"

### 4.2 Callback URL 설정
서비스 URL 설정:
```
https://your-app.vercel.app
```

Callback URL:
```
https://your-app.koyeb.app/v1/auth/naver/callback
```

### 4.3 Client ID/Secret 복사
생성된 Client ID와 Client Secret을 Koyeb 환경변수에 설정

---

## 5. 배포 후 확인 사항

### 5.1 헬스 체크
```bash
# Backend
curl https://your-app.koyeb.app/health

# Frontend
curl https://your-app.vercel.app
```

### 5.2 데이터베이스 연결 확인
Supabase 대시보드 → Table Editor에서 테이블들이 생성되었는지 확인:
- `user_profiles`
- `experiences`
- `applications`
- 기타 테이블들

### 5.3 로그 확인
- **Koyeb**: Dashboard → Logs 탭
- **Vercel**: Dashboard → Deployments → 최신 배포 → Logs

---

## 6. 무료 티어 제한사항

### Supabase (Free)
- Database: 500MB
- Bandwidth: 5GB/월
- API requests: Unlimited
- 프로젝트 일시정지: 7일 비활성 시

### Koyeb (Free)
- Apps: 2개
- vCPU: 0.1 (공유)
- RAM: 512MB
- Bandwidth: 100GB/월
- Sleep after 15분 비활성

### Vercel (Free)
- Bandwidth: 100GB/월
- Build time: 6,000분/월
- Serverless Functions: 무제한
- No sleep

---

## 7. 트러블슈팅

### Backend가 sleep 상태
Koyeb 무료 티어는 15분 비활성 후 sleep. 첫 요청 시 콜드 스타트 발생 (약 5-10초).

**해결책**:
- Uptime monitoring (UptimeRobot 등) 사용해서 주기적으로 핑
- 유료 플랜 고려 ($5/월)

### Database connection error
1. DATABASE_URL 형식 확인 (sslmode=require 필수)
2. Supabase 프로젝트가 일시정지되지 않았는지 확인
3. Connection pooler 사용: `pooler.supabase.com` (트랜잭션 모드)

### CORS error
1. Koyeb 환경변수에 `FRONTEND_URL` 올바르게 설정되었는지 확인
2. Backend 코드에서 CORS 미들웨어 설정 확인

### Vercel build failed
1. Root Directory가 `apps/web`로 설정되었는지 확인
2. Build Command가 monorepo 루트에서 실행되는지 확인
3. Moon CLI 설치가 포함되었는지 확인

---

## 8. 다음 단계

배포 완료 후:
1. [ ] 커스텀 도메인 연결 (선택적)
2. [ ] SSL 인증서 확인 (Vercel/Koyeb 자동)
3. [ ] 모니터링 설정 (Sentry, LogRocket 등)
4. [ ] 백업 전략 수립 (Supabase 자동 백업 7일)
5. [ ] Analytics 추가 (Vercel Analytics, Google Analytics)

---

## 참고 링크

- [Koyeb Documentation](https://www.koyeb.com/docs)
- [Vercel Documentation](https://vercel.com/docs)
- [Supabase Documentation](https://supabase.com/docs)
- [Moon Documentation](https://moonrepo.dev/docs)
