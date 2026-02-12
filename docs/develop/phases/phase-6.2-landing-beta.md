# Phase 6.2: 랜딩 & 베타

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | 서비스 랜딩 페이지를 제작하고, 법적 페이지(개인정보처리방침/이용약관)를 추가하여 Vercel 프로덕션에 베타 배포한다 |
| **선행 조건** | Phase 6.1 (프리미엄 & 마무리) 완료, 핵심 기능 + 에러 처리 + 반응형 + 성능 최적화 완료 |
| **스프린트** | Sprint 5 |
| **관련 기능** | MVP 베타 출시 |
| **예상 공수** | 1일 (Day 5) |
| **산출물** | 랜딩 페이지, 법적 페이지 2개, Vercel 프로덕션 배포, 피드백 수집, 모니터링 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 6.2.1 | 랜딩 페이지 | ⬜ 대기 |
| 6.2.2 | 법적 페이지 | ⬜ 대기 |
| 6.2.3 | 베타 배포 | ⬜ 대기 |
| 6.2.4 | 피드백 수집 | ⬜ 대기 |
| 6.2.5 | 모니터링 | ⬜ 대기 |

---

## Step 6.2.1: 랜딩 페이지

### 목표

서비스의 가치를 한눈에 전달하는 랜딩 페이지를 구현한다. Hero 섹션, 핵심 기능 3가지 소개, 데모/스크린샷, CTA(회원가입) 버튼으로 구성한다. 비로그인 사용자가 처음 접하는 페이지이므로 임팩트 있게 제작한다.

### 체크리스트

- [ ] `src/app/page.tsx` 랜딩 페이지 구현 (기존 리다이렉트 대체)
- [ ] Hero 섹션: 메인 카피 + 서브 카피 + CTA 버튼
- [ ] 핵심 기능 소개 섹션 (3개 기능 카드)
- [ ] 데모/스크린샷 섹션 (서비스 미리보기)
- [ ] CTA 섹션: "지금 시작하기" → 회원가입
- [ ] 푸터: 서비스명, 법적 링크, 저작권
- [ ] 로그인 사용자 → `/dashboard`로 리다이렉트
- [ ] SEO 메타태그 (title, description, og:image)
- [ ] 반응형 (모바일 우선)

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `LandingPage` | `src/app/page.tsx` | - | 랜딩 페이지 (서버 컴포넌트) |
| `HeroSection` | `src/components/landing/hero-section.tsx` | - | Hero 섹션 (메인 카피 + CTA) |
| `FeatureSection` | `src/components/landing/feature-section.tsx` | - | 핵심 기능 3개 카드 |
| `DemoSection` | `src/components/landing/demo-section.tsx` | - | 데모/스크린샷 미리보기 |
| `CTASection` | `src/components/landing/cta-section.tsx` | - | 하단 CTA |
| `LandingFooter` | `src/components/landing/landing-footer.tsx` | - | 푸터 (법적 링크 포함) |

### 페이지 레이아웃

```
┌──────────────────────────────────────────────────────────────────┐
│  Colight                                    [로그인] [시작하기]  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│                    🔦 Colight                                     │
│                                                                  │
│            취업 준비, 이제 AI 코치와 함께                           │
│                                                                  │
│    경험 정리부터 기업 분석, 자소서 코칭까지                         │
│    당신만의 AI 취업 코치가 합격까지 이끌어드립니다                   │
│                                                                  │
│                [ 무료로 시작하기 →]                                │
│                                                                  │
│    무료: 경험 3개 등록 + 기업 분석 1회 + AI 코칭 1회               │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ✨ 핵심 기능                                                     │
│                                                                  │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐     │
│  │ 🏢 기업 분석     │  │ 🎯 경험 매칭    │  │ ✍️ AI 코칭     │     │
│  │                │  │                │  │                │     │
│  │ 채용공고 URL만  │  │ 내 경험이 어떤   │  │ STAR 구조 초안  │     │
│  │ 넣으면 인재상,  │  │ 기업에 맞는지    │  │ 부터 4점 첨삭    │     │
│  │ 핵심가치,      │  │ 적합도 점수로    │  │ 까지 AI가       │     │
│  │ 전략 키워드를   │  │ 한눈에!         │  │ 코칭해줍니다     │     │
│  │ 자동 분석!     │  │                │  │                │     │
│  └────────────────┘  └────────────────┘  └────────────────┘     │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  📸 이런 걸 할 수 있어요                                          │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                                                          │   │
│  │              [서비스 스크린샷 / 데모 GIF]                  │   │
│  │                                                          │   │
│  │  기업 분석 리포트 예시 → 문항 분석 → 초안 → 첨삭 결과       │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│              합격하는 자소서, 지금 바로 시작하세요                   │
│                                                                  │
│                    [ 무료로 시작하기 → ]                           │
│                                                                  │
│              이미 계정이 있으신가요? 로그인                         │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  Colight  |  개인정보처리방침  |  이용약관  |  © 2026             │
└──────────────────────────────────────────────────────────────────┘
```

### SEO 메타태그

```typescript
// src/app/page.tsx
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Colight - AI 취업 코치 | 경험 정리부터 자소서 코칭까지',
  description: '채용공고 URL만 넣으면 기업 분석, 경험 매칭, STAR 구조 초안, 4점 첨삭까지. 당신만의 AI 취업 코치가 합격까지 이끌어드립니다.',
  openGraph: {
    title: 'Colight - AI 취업 코치',
    description: '경험 정리부터 기업 분석, 자소서 코칭까지',
    type: 'website',
    url: 'https://colight.app',
    images: [{ url: '/og-image.png', width: 1200, height: 630 }],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Colight - AI 취업 코치',
    description: '경험 정리부터 기업 분석, 자소서 코칭까지',
    images: ['/og-image.png'],
  },
};
```

### 로그인 사용자 리다이렉트

```typescript
// src/app/page.tsx
import { createClient } from '@/lib/supabase/server';
import { redirect } from 'next/navigation';

export default async function LandingPage() {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();

  // 로그인 사용자는 대시보드로
  if (user) redirect('/dashboard');

  return (
    <main>
      <HeroSection />
      <FeatureSection />
      <DemoSection />
      <CTASection />
      <LandingFooter />
    </main>
  );
}
```

### 검증 방법

- [ ] 비로그인 사용자: 랜딩 페이지 정상 표시
- [ ] 로그인 사용자: `/dashboard`로 자동 리다이렉트
- [ ] "무료로 시작하기" → `/signup` 페이지로 이동
- [ ] "로그인" → `/login` 페이지로 이동
- [ ] 모바일 375px: 단일 컬럼 레이아웃 정상
- [ ] SEO: `<title>`, `<meta name="description">`, `og:image` 확인
- [ ] Lighthouse SEO ≥ 90

### 산출물

- `src/app/page.tsx`
- `src/components/landing/hero-section.tsx`
- `src/components/landing/feature-section.tsx`
- `src/components/landing/demo-section.tsx`
- `src/components/landing/cta-section.tsx`
- `src/components/landing/landing-footer.tsx`
- `public/og-image.png` (OG 이미지)

---

## Step 6.2.2: 법적 페이지

### 목표

서비스 운영에 필수적인 개인정보처리방침과 이용약관 페이지를 추가한다. 한국 법률(개인정보보호법, 전자상거래법)에 부합하는 최소 요건을 갖춘다.

### 체크리스트

- [ ] `/privacy` 개인정보처리방침 페이지 구현
- [ ] `/terms` 이용약관 페이지 구현
- [ ] 랜딩 페이지 푸터에 링크 추가
- [ ] 회원가입 페이지에 "이용약관 동의" 체크박스 추가
- [ ] 마크다운 기반 콘텐츠 (추후 수정 용이하도록)

### 개인정보처리방침 필수 항목

```
1. 개인정보의 처리 목적
2. 수집하는 개인정보의 항목
   - 필수: 이메일, 이름(닉네임)
   - 선택: 소셜 로그인 프로필
3. 개인정보의 보유 및 이용기간
4. 개인정보의 제3자 제공 (해당 시)
   - AI API (Anthropic, OpenAI)에 텍스트 데이터 전송
5. 개인정보처리의 위탁
   - Supabase (데이터 저장)
   - Vercel (호스팅)
6. 정보주체의 권리·의무 및 행사방법
7. 개인정보 파기절차 및 방법
8. 개인정보 보호책임자
```

### 이용약관 필수 항목

```
1. 목적
2. 용어의 정의
3. 서비스의 제공
4. 회원가입 및 탈퇴
5. 이용요금
6. AI 서비스 관련 면책
   - AI 생성 결과물의 정확성 보장하지 않음
   - 최종 결과물은 사용자 책임
7. 저작권 및 지적재산권
   - 사용자 입력 데이터: 사용자 소유
   - AI 생성 결과물: 사용자에게 이용 허락
8. 서비스의 중단
9. 면책 조항
10. 분쟁 해결
```

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `PrivacyPage` | `src/app/(legal)/privacy/page.tsx` | - | 개인정보처리방침 |
| `TermsPage` | `src/app/(legal)/terms/page.tsx` | - | 이용약관 |
| `LegalLayout` | `src/app/(legal)/layout.tsx` | - | 법적 페이지 공통 레이아웃 (간결한 헤더+푸터) |

### 검증 방법

- [ ] `/privacy` 페이지 접근 시 개인정보처리방침 표시
- [ ] `/terms` 페이지 접근 시 이용약관 표시
- [ ] 랜딩 페이지 푸터에서 링크 정상 동작
- [ ] 회원가입 시 이용약관 동의 체크박스 미체크 시 가입 불가
- [ ] 비로그인 상태에서도 접근 가능
- [ ] 모바일 반응형 확인

### 산출물

- `src/app/(legal)/layout.tsx`
- `src/app/(legal)/privacy/page.tsx`
- `src/app/(legal)/terms/page.tsx`

---

## Step 6.2.3: 베타 배포

### 목표

Vercel에 프로덕션 배포를 수행하고, 환경 변수를 설정하고, Supabase 프로덕션 프로젝트를 연결하여 5~10명의 베타 사용자가 실제 서비스를 사용할 수 있도록 한다.

### 체크리스트

- [ ] Vercel 프로젝트 프로덕션 환경 변수 설정
- [ ] Supabase 프로덕션 프로젝트 생성 + DB 마이그레이션 적용
- [ ] Supabase 프로덕션 Auth 설정 (이메일/소셜 로그인)
- [ ] 커스텀 도메인 연결 (선택: colight.app 또는 colight.vercel.app)
- [ ] SSL 인증서 확인 (Vercel 자동)
- [ ] 프로덕션 빌드 테스트 (`npm run build && npm run start`)
- [ ] 시드 데이터 프로덕션 적용 (weapon_categories, question_patterns, prompt_templates)
- [ ] 환경별 API 키 분리 확인 (dev/prod)
- [ ] 5~10명 베타 테스터 초대 (이메일)

### 환경 변수 체크리스트

```
# Vercel 프로덕션 환경 변수
NEXT_PUBLIC_SUPABASE_URL=https://xxxxx.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJ...
SUPABASE_SERVICE_ROLE_KEY=eyJ...

ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...

DART_API_KEY=...
NAVER_CLIENT_ID=...
NAVER_CLIENT_SECRET=...

NEXT_PUBLIC_APP_URL=https://colight.app (또는 vercel.app URL)
```

### 배포 절차

```bash
# 1. 프로덕션 빌드 확인
npm run build

# 2. Supabase 프로덕션 링크
supabase link --project-ref <project-id>

# 3. 프로덕션 DB 마이그레이션
supabase db push --linked

# 4. 시드 데이터 적용
psql $DATABASE_URL -f supabase/seed.sql

# 5. TypeScript 타입 재생성 (프로덕션 기준)
supabase gen types typescript --linked > src/types/database.ts

# 6. Vercel 배포 (main 브랜치 push)
git push origin main
# → Vercel 자동 배포
```

### 베타 테스터 초대 메시지 템플릿

```
안녕하세요!

AI 취업 코칭 서비스 Colight의 베타 테스트에 초대합니다.

🔗 서비스 주소: https://colight.app
📧 가입: 이메일로 회원가입 후 사용해주세요

🎁 무료 체험:
- 경험 3개 등록
- 기업 분석 1회/일
- AI 코칭 1회/일

📝 피드백:
- 서비스 내 피드백 버튼을 이용해주세요
- 버그, 개선 아이디어, 어색한 부분 모두 환영합니다!

감사합니다!
```

### 검증 방법

- [ ] 프로덕션 URL에서 랜딩 페이지 정상 표시
- [ ] 회원가입 → 로그인 → 대시보드 접근 정상
- [ ] 경험 등록 → 기업 분석 → 코칭 전체 흐름 동작
- [ ] SSL 인증서 유효 (HTTPS)
- [ ] 프로덕션 Supabase RLS 정책 동작 확인
- [ ] AI API (Claude, OpenAI) 프로덕션 키로 정상 호출
- [ ] 사용량 제한 프로덕션에서 동작 확인
- [ ] 베타 테스터 5명 이상 가입 확인

### 산출물

- Vercel 프로덕션 배포 완료
- Supabase 프로덕션 DB 설정
- 베타 테스터 초대 완료

---

## Step 6.2.4: 피드백 수집

### 목표

베타 사용자로부터 실시간으로 피드백을 수집할 수 있는 플로팅 피드백 버튼을 구현한다. 최소한의 기능으로 텍스트 피드백 + 페이지 URL + 스크린샷 여부를 수집한다.

### 체크리스트

- [ ] 플로팅 피드백 버튼 구현 (우측 하단 고정)
- [ ] 피드백 모달: 카테고리(버그/개선/기타) + 텍스트 입력
- [ ] 현재 페이지 URL 자동 첨부
- [ ] 사용자 ID 자동 첨부 (로그인 시)
- [ ] 피드백 저장 (Supabase `feedback` 테이블 또는 외부 서비스)
- [ ] 제출 후 감사 토스트 메시지

### DB 스키마

```sql
CREATE TABLE public.feedback (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid REFERENCES auth.users(id),
  category text NOT NULL CHECK (category IN ('bug', 'improvement', 'other')),
  content text NOT NULL,
  page_url text,
  user_agent text,
  created_at timestamptz DEFAULT now()
);

-- 모든 인증 사용자 쓰기 가능
ALTER TABLE public.feedback ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Authenticated users can submit feedback"
  ON public.feedback FOR INSERT
  WITH CHECK (auth.role() = 'authenticated');
```

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `FeedbackButton` | `src/components/feedback/feedback-button.tsx` | - | 플로팅 피드백 버튼 (FAB) |
| `FeedbackModal` | `src/components/feedback/feedback-modal.tsx` | `onSubmit: fn, onClose: fn` | 피드백 입력 모달 |

### UI 레이아웃

```
┌─────────────────────────────────────┐
│  💬 피드백 보내기                ×   │
│                                     │
│  카테고리                           │
│  ○ 🐛 버그  ○ 💡 개선  ○ 💬 기타   │
│                                     │
│  내용                               │
│  ┌─────────────────────────────────┐│
│  │ 어떤 점이 불편하셨나요?          ││
│  │ 또는 어떤 기능이 있었으면         ││
│  │ 좋겠나요?                       ││
│  └─────────────────────────────────┘│
│                                     │
│  📍 현재 페이지: /coaching/edit     │
│                                     │
│  [ 보내기 ]                         │
└─────────────────────────────────────┘

// 플로팅 버튼 (우측 하단)
                              [💬]  ← 항상 표시
```

### 검증 방법

- [ ] 플로팅 버튼이 모든 페이지에서 표시 (랜딩 제외)
- [ ] 버튼 클릭 시 피드백 모달 열림
- [ ] 카테고리 선택 + 내용 입력 후 제출 성공
- [ ] `feedback` 테이블에 레코드 생성 확인
- [ ] 현재 페이지 URL 자동 첨부 확인
- [ ] 제출 후 "감사합니다" 토스트 표시
- [ ] 모바일에서 플로팅 버튼 + 모달 정상 동작

### 산출물

- `supabase/migrations/YYYYMMDD_create_feedback_table.sql`
- `src/components/feedback/feedback-button.tsx`
- `src/components/feedback/feedback-modal.tsx`
- `src/app/api/feedback/route.ts`

---

## Step 6.2.5: 모니터링

### 목표

프로덕션 환경에서 서비스 상태, 사용자 행동, AI 비용을 모니터링할 수 있는 기본 인프라를 설정한다.

### 체크리스트

- [ ] Vercel Analytics 활성화 (Web Vitals 자동 수집)
- [ ] Vercel Speed Insights 활성화 (성능 모니터링)
- [ ] Supabase Dashboard 확인 (DB 사용량, Auth 사용자 수)
- [ ] AI 비용 추적 대시보드 (coaching_sessions의 토큰/비용 집계)
- [ ] 에러 모니터링 기본 설정 (console.error → Vercel Logs)
- [ ] 주요 지표 수동 확인 루틴 문서화

### Vercel Analytics 설정

```typescript
// src/app/layout.tsx
import { Analytics } from '@vercel/analytics/react';
import { SpeedInsights } from '@vercel/speed-insights/next';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
```

```bash
npm install @vercel/analytics @vercel/speed-insights
```

### AI 비용 추적 쿼리

```sql
-- 일별 AI 사용 비용 집계
SELECT
  DATE(created_at) AS date,
  session_type,
  COUNT(*) AS sessions,
  SUM((messages->-1->'usage'->>'total_cost')::int) AS total_cost_krw
FROM coaching_sessions
WHERE created_at >= now() - interval '30 days'
GROUP BY DATE(created_at), session_type
ORDER BY date DESC;
```

### 모니터링 대시보드 (수동 확인)

| 지표 | 확인 위치 | 주기 |
|------|----------|------|
| 일별 가입자 수 | Supabase Dashboard > Auth | 매일 |
| 일별 활성 사용자 | Vercel Analytics | 매일 |
| AI API 비용 | coaching_sessions 쿼리 | 매일 |
| 에러율 | Vercel Functions Logs | 매일 |
| 페이지 로드 속도 | Vercel Speed Insights | 주간 |
| DB 용량 | Supabase Dashboard | 주간 |
| 피드백 건수 | feedback 테이블 | 매일 |
| Lighthouse 점수 | 수동 측정 | 주간 |

### 비용 알람 설정

```typescript
// src/lib/monitoring/cost-alert.ts
// 일별 AI 비용이 임계치를 넘으면 알림
const DAILY_COST_ALERT_THRESHOLD = 5000; // 원

export async function checkDailyCost() {
  const supabase = await createClient();
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const { data: sessions } = await supabase
    .from('coaching_sessions')
    .select('messages')
    .gte('created_at', today.toISOString());

  const totalCost = sessions?.reduce((sum, s) => {
    const lastMsg = s.messages?.[s.messages.length - 1];
    return sum + (lastMsg?.usage?.total_cost || 0);
  }, 0) || 0;

  if (totalCost > DAILY_COST_ALERT_THRESHOLD) {
    console.warn(`[COST ALERT] Daily AI cost: ${totalCost}원 (threshold: ${DAILY_COST_ALERT_THRESHOLD}원)`);
    // 추후: Slack/이메일 알림 추가
  }

  return totalCost;
}
```

### 검증 방법

- [ ] Vercel Analytics 대시보드에서 페이지뷰 확인
- [ ] Vercel Speed Insights에서 Web Vitals 확인
- [ ] Supabase Dashboard에서 사용자 수/DB 용량 확인
- [ ] AI 비용 쿼리 실행 결과 확인
- [ ] Vercel Functions Logs에서 에러 로그 확인 가능
- [ ] 비용 알람 임계치 동작 확인 (로그 출력)

### 산출물

- `src/app/layout.tsx` (Analytics/SpeedInsights 추가)
- `src/lib/monitoring/cost-alert.ts`
- 모니터링 확인 루틴 문서

---

## Phase 완료 체크리스트

- [ ] 랜딩 페이지 완성 (Hero + 기능 3개 + 데모 + CTA)
- [ ] 비로그인 → 랜딩, 로그인 → 대시보드 리다이렉트
- [ ] SEO 메타태그 설정 완료
- [ ] 개인정보처리방침 + 이용약관 페이지 게시
- [ ] Vercel 프로덕션 배포 완료
- [ ] Supabase 프로덕션 DB 설정 + 마이그레이션 + 시드 완료
- [ ] SSL 인증서 유효
- [ ] 환경 변수 프로덕션 설정 완료
- [ ] 플로팅 피드백 버튼 동작
- [ ] Vercel Analytics + Speed Insights 활성화
- [ ] AI 비용 추적 가능
- [ ] 베타 테스터 5~10명 초대 완료
- [ ] 전체 플로우 프로덕션 동작 확인

---

## 다음 Phase

**[Phase 7: AI 인터뷰](./phase-7-ai-interview.md)** — AI 대화형 경험 발굴 인터뷰 (Post-MVP)

---

## MVP 완료!

Phase 6.2를 완료하면 Colight MVP가 베타 출시됩니다.

```
✅ 핵심 사용자 여정 완성:
   경험 등록 → 무기 태깅 → 채용공고 분석 → 기업 분석
   → 경험 매칭 → 문항 분석 → 경험 추천
   → AI 초안 코칭 → 에디터 편집 → AI 첨삭 → 반복 개선

✅ Freemium 모델 작동:
   무료: 경험 3개 + 분석 1회/일 + 코칭 1회/일
   유료: 스타터(4,900원) / 프로(12,900원) / 시즌패스(24,900원)

✅ 베타 배포 완료:
   Vercel + Supabase 프로덕션
   5~10명 베타 테스터 피드백 수집 중
```
