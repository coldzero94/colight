# 에러 처리 컨벤션

> API 응답 형식, 에러 코드 카탈로그, AI/외부 API 에러 핸들링
>
> **⚠️ 아키텍처 노트**: 이 문서의 일부 코드(NextResponse, Vercel 타임아웃 등)가 이전 아키텍처(Next.js API Routes) 기준입니다. 실제 에러 처리는 Go 백엔드(Gin 미들웨어)에서 구현합니다. TypeScript 코드는 프론트엔드 에러 표시 패턴만 참고하세요.

---

## 1. API 응답 형식

### 성공 응답

```json
{
  "data": { ... }
}
```

### 에러 응답

```json
{
  "error": {
    "message": "사용자에게 보여줄 메시지",
    "code": "AUTH_001"
  }
}
```

### TypeScript 타입

```typescript
// src/types/api.ts

type ApiSuccessResponse<T> = {
  data: T;
};

type ApiErrorResponse = {
  error: {
    message: string;
    code?: string;
  };
};

type ApiResponse<T> = ApiSuccessResponse<T> | ApiErrorResponse;
```

---

## 2. 에러 코드 카탈로그

### 인증 (AUTH)

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `AUTH_001` | 미인증 | 401 | 로그인이 필요합니다 |
| `AUTH_002` | 세션 만료 | 401 | 세션이 만료되었습니다. 다시 로그인해 주세요 |

### 입력 검증 (VALIDATION)

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `VALIDATION_001` | 잘못된 입력 | 400 | 입력값을 확인해 주세요 |

### 속도 제한 (RATE)

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `RATE_001` | 요청 제한 초과 | 429 | 잠시 후 다시 시도해 주세요 |

### AI (AI)

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `AI_001` | AI 응답 타임아웃 | 504 | AI 분석이 지연되고 있습니다. 다시 시도해 주세요 |
| `AI_002` | AI API 속도 제한 | 429 | AI 서비스가 바쁩니다. 잠시 후 다시 시도해 주세요 |

### 크롤링 (CRAWL)

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `CRAWL_001` | URL 파싱 실패 | 400 | 해당 URL을 분석할 수 없습니다. URL을 확인해 주세요 |

### 외부 API (EXTERNAL)

| 코드 | 설명 | HTTP 상태 | 사용자 메시지 |
|------|------|----------|-------------|
| `EXTERNAL_001` | DART API 오류 | 502 | 기업 재무정보를 가져올 수 없습니다 |
| `EXTERNAL_002` | 네이버 API 오류 | 502 | 뉴스 정보를 가져올 수 없습니다 |

---

## 3. 에러 유틸리티 함수

```typescript
// src/lib/errors.ts

import { NextResponse } from 'next/server';

export class AppError extends Error {
  constructor(
    message: string,
    public code: string,
    public status: number = 500
  ) {
    super(message);
    this.name = 'AppError';
  }
}

// 에러 팩토리
export const Errors = {
  unauthorized: () =>
    new AppError('로그인이 필요합니다', 'AUTH_001', 401),
  sessionExpired: () =>
    new AppError('세션이 만료되었습니다. 다시 로그인해 주세요', 'AUTH_002', 401),
  invalidInput: (detail?: string) =>
    new AppError(detail ?? '입력값을 확인해 주세요', 'VALIDATION_001', 400),
  rateLimited: () =>
    new AppError('잠시 후 다시 시도해 주세요', 'RATE_001', 429),
  aiTimeout: () =>
    new AppError('AI 분석이 지연되고 있습니다. 다시 시도해 주세요', 'AI_001', 504),
  aiRateLimited: () =>
    new AppError('AI 서비스가 바쁩니다. 잠시 후 다시 시도해 주세요', 'AI_002', 429),
  crawlFailed: () =>
    new AppError('해당 URL을 분석할 수 없습니다. URL을 확인해 주세요', 'CRAWL_001', 400),
  dartError: () =>
    new AppError('기업 재무정보를 가져올 수 없습니다', 'EXTERNAL_001', 502),
  naverError: () =>
    new AppError('뉴스 정보를 가져올 수 없습니다', 'EXTERNAL_002', 502),
} as const;

// API Route 에러 응답 헬퍼
export function errorResponse(error: unknown): NextResponse {
  if (error instanceof AppError) {
    return NextResponse.json(
      { error: { message: error.message, code: error.code } },
      { status: error.status }
    );
  }

  console.error('Unexpected error:', error);
  return NextResponse.json(
    { error: { message: '서버 오류가 발생했습니다' } },
    { status: 500 }
  );
}
```

### 프론트엔드 API 호출에서 사용

```typescript
// src/hooks/use-experiences.ts

import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/generated/sdk.gen';
import { toast } from 'sonner';

export function useExperiences() {
  return useQuery({
    queryKey: ['experiences'],
    queryFn: async () => {
      const response = await api.GET('/v1/experiences');

      if (response.error) {
        // Go backend에서 반환한 에러 처리
        if (response.error.code === 'AUTH_001') {
          toast.error('로그인이 필요합니다');
          // 로그인 페이지로 리다이렉트
        } else {
          toast.error(response.error.message || '경험을 불러오는데 실패했습니다');
        }
        throw new Error(response.error.message);
      }

      return response.data;
    },
  });
}
```

---

## 4. Supabase 에러 처리 패턴

```typescript
// src/lib/supabase/helpers.ts

import { PostgrestError } from '@supabase/supabase-js';
import { AppError } from '@/lib/errors';

export function handleSupabaseError(error: PostgrestError): never {
  // RLS 위반
  if (error.code === '42501') {
    throw new AppError('접근 권한이 없습니다', 'AUTH_001', 403);
  }

  // 고유 제약 위반
  if (error.code === '23505') {
    throw new AppError('이미 존재하는 데이터입니다', 'VALIDATION_001', 409);
  }

  // FK 위반
  if (error.code === '23503') {
    throw new AppError('참조하는 데이터가 존재하지 않습니다', 'VALIDATION_001', 400);
  }

  // 기타
  console.error('Supabase error:', error);
  throw new AppError('데이터베이스 오류가 발생했습니다', 'DB_001', 500);
}
```

---

## 5. AI 에러 처리

### 재시도 (3회, 지수 백오프)

```typescript
// src/lib/ai/retry.ts

export async function withRetry<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3
): Promise<T> {
  let lastError: Error | undefined;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;

      // Rate limit (429) 또는 서버 에러 (5xx)만 재시도
      if (error instanceof Error && 'status' in error) {
        const status = (error as { status: number }).status;
        if (status !== 429 && status < 500) throw error;
      }

      // 지수 백오프: 1초 → 2초 → 4초
      const delay = Math.pow(2, attempt) * 1000;
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }

  throw lastError;
}
```

### Vercel 타임아웃 대응

| 환경 | 타임아웃 | 대응 |
|------|---------|------|
| Hobby (무료) | 10초 | 스트리밍 응답 사용 |
| Fluid Compute | 60초 | 복잡한 분석 파이프라인 |

```typescript
// src/app/api/analyze/route.ts

// Fluid Compute 60초 허용
export const maxDuration = 60;

export async function POST(req: Request) {
  // 스트리밍 응답으로 타임아웃 방지
  const stream = new ReadableStream({
    async start(controller) {
      try {
        const result = await withRetry(() => analyzeCompany(data));
        controller.enqueue(encoder.encode(JSON.stringify(result)));
      } catch (error) {
        controller.enqueue(
          encoder.encode(JSON.stringify({
            error: { message: 'AI 분석에 실패했습니다', code: 'AI_001' }
          }))
        );
      } finally {
        controller.close();
      }
    },
  });

  return new Response(stream, {
    headers: { 'Content-Type': 'text/event-stream' },
  });
}
```

### 프로바이더 폴백

```typescript
// src/lib/ai/provider.ts

export async function analyzeWithFallback(prompt: string): Promise<string> {
  try {
    // 1차: Claude Sonnet 4.5
    return await callAnthropic(prompt);
  } catch (error) {
    console.warn('Anthropic failed, falling back to OpenAI:', error);

    try {
      // 2차: GPT-4.1 mini (폴백)
      return await callOpenAI(prompt);
    } catch (fallbackError) {
      console.error('All AI providers failed:', fallbackError);
      throw Errors.aiTimeout();
    }
  }
}
```

---

## 6. 프론트엔드 에러 바운더리

### Route별 error.tsx

```typescript
// src/app/(main)/experiences/error.tsx
'use client';

export default function ExperienceError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20">
      <h2 className="text-lg font-semibold">경험 데이터를 불러올 수 없습니다</h2>
      <p className="text-sm text-muted-foreground">{error.message}</p>
      <button
        onClick={reset}
        className="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground"
      >
        다시 시도
      </button>
    </div>
  );
}
```

### 주요 라우트별 error.tsx 구성

```
src/app/
├── (main)/
│   ├── experiences/error.tsx     # 경험 관리 에러
│   ├── analysis/error.tsx        # 기업 분석 에러
│   ├── coaching/error.tsx        # 코칭 에러
│   └── dashboard/error.tsx       # 대시보드 에러
├── (auth)/error.tsx              # 인증 에러
└── global-error.tsx              # 전체 앱 폴백
```

---

## 7. 스트리밍 에러 처리

Go backend SSE 스트리밍 중 에러 발생 시:

```typescript
// 클라이언트: SSE 스트리밍 에러 처리
import { useState } from 'react';
import { toast } from 'sonner';

function CoachingStream() {
  const [content, setContent] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const startStreaming = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('/v1/coaching/draft', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ /* ... */ }),
      });

      if (!response.ok) {
        throw new Error('스트리밍 시작 실패');
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        setContent(prev => prev + chunk);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'AI 응답 중 오류가 발생했습니다';
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  if (error) {
    return (
      <div>
        <p>오류: {error}</p>
        <button onClick={startStreaming}>다시 시도</button>
      </div>
    );
  }

  return <>{/* content 렌더링 */}</>;
}
```

---

## 8. 외부 API Graceful Degradation

외부 API 실패 시 전체 기능이 멈추지 않도록 부분 결과를 반환한다.

```typescript
// src/lib/external/aggregate.ts

type CompanyData = {
  dart: DartInfo | null;
  news: NewsItem[] | null;
  talent: TalentProfile | null;
};

export async function fetchCompanyData(
  companyName: string
): Promise<CompanyData> {
  const [dart, news, talent] = await Promise.allSettled([
    fetchDartInfo(companyName),
    fetchNaverNews(companyName),
    fetchTalentProfile(companyName),
  ]);

  return {
    dart: dart.status === 'fulfilled' ? dart.value : null,
    news: news.status === 'fulfilled' ? news.value : null,
    talent: talent.status === 'fulfilled' ? talent.value : null,
  };
}
```

**원칙:**
- `Promise.allSettled`로 개별 API 실패를 격리
- 실패한 항목은 `null`로 표시, 나머지 결과는 정상 반환
- UI에서 `null`인 섹션은 "정보를 가져올 수 없습니다" 표시

---

## 9. Toast 알림 컨벤션

[sonner](https://sonner.emilkowal.dev/) 라이브러리 사용.

### 사용 패턴

```typescript
import { toast } from 'sonner';

// 성공
toast.success('경험이 저장되었습니다');

// 에러
toast.error('저장에 실패했습니다. 다시 시도해 주세요');

// 경고 (외부 API 부분 실패)
toast.warning('일부 기업 정보를 가져오지 못했습니다');

// 로딩 → 완료
const id = toast.loading('분석 중...');
// ... 작업 완료 후
toast.success('분석이 완료되었습니다', { id });
```

### 규칙

| 상황 | 토스트 타입 | 예시 |
|------|-----------|------|
| CRUD 성공 | `success` | 경험이 저장되었습니다 |
| 사용자 입력 에러 | `error` | 제목을 입력해 주세요 |
| 서버 에러 | `error` | 서버 오류가 발생했습니다 |
| AI 타임아웃 | `error` | AI 분석이 지연되고 있습니다 |
| 외부 API 부분 실패 | `warning` | 일부 정보를 불러오지 못했습니다 |
| 장시간 작업 진행 | `loading` → `success` | 분석 중... → 완료 |

### 레이아웃 설정

```typescript
// src/app/layout.tsx
import { Toaster } from 'sonner';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Toaster position="top-right" richColors closeButton />
      </body>
    </html>
  );
}
```
