# 테스트 전략

> Vitest + Playwright 기반 테스트 계층, AI 모킹, CI/CD

---

## 1. 테스트 프레임워크

| 계층 | 프레임워크 | 용도 |
|------|-----------|------|
| Unit | Vitest | 유틸리티 함수, Zod 스키마, 데이터 변환 |
| Integration | Vitest + Testing Library | API Routes, React 컴포넌트 |
| E2E | Playwright | 사용자 시나리오 전체 흐름 |

### 설치

```bash
# Unit / Integration
pnpm add -D vitest @testing-library/react @testing-library/jest-dom
pnpm add -D @vitejs/plugin-react jsdom
pnpm add -D msw  # API 모킹

# E2E
pnpm add -D @playwright/test
pnpm exec playwright install
```

---

## 2. 테스트 계층별 범위

### Unit 테스트

**대상:** 외부 의존성 없는 순수 로직

```typescript
// src/lib/utils/__tests__/format.test.ts
import { describe, it, expect } from 'vitest';
import { truncateText, formatCharCount } from '../format';

describe('truncateText', () => {
  it('최대 길이 초과 시 말줄임', () => {
    expect(truncateText('안녕하세요 반갑습니다', 5)).toBe('안녕하세요…');
  });

  it('최대 길이 이하 시 원본 반환', () => {
    expect(truncateText('안녕', 5)).toBe('안녕');
  });
});
```

**Zod 스키마 테스트:**

```typescript
// src/lib/validations/__tests__/experience.test.ts
import { describe, it, expect } from 'vitest';
import { experienceSchema } from '../experience';

describe('experienceSchema', () => {
  it('유효한 경험 데이터 통과', () => {
    const result = experienceSchema.safeParse({
      title: '프로젝트 리더 경험',
      situation: '팀 프로젝트에서...',
      task: '팀장으로서...',
      action: 'Sprint 방식을 도입하여...',
      result: '기한 내 성공적으로 완료',
    });
    expect(result.success).toBe(true);
  });

  it('제목 누락 시 실패', () => {
    const result = experienceSchema.safeParse({ situation: '...' });
    expect(result.success).toBe(false);
  });
});
```

**데이터 변환 테스트:**

```typescript
// src/lib/ai/__tests__/parsing.test.ts
import { describe, it, expect } from 'vitest';
import { parseJobPosting, extractCompanyName } from '../parsing';

describe('parseJobPosting', () => {
  it('HTML에서 채용공고 데이터 추출', () => {
    const html = '<div class="job-title">프론트엔드 개발자</div>';
    const result = parseJobPosting(html, 'jobkorea');
    expect(result.position).toBe('프론트엔드 개발자');
  });
});
```

### Integration 테스트

**대상:** Go API 호출 + React 컴포넌트 + 상태

> **규칙**: 통합 테스트는 Go 백엔드 API를 통해 데이터를 조회/생성합니다. Supabase를 직접 쿼리하지 않습니다.

**API 호출 테스트 (Go 백엔드):**

```typescript
// src/hooks/__tests__/use-experiences.test.ts
import { describe, it, expect } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { useExperiences } from '../use-experiences';

describe('Experiences API', () => {
  it('경험 목록 조회 성공', async () => {
    const queryClient = new QueryClient();
    const { result } = renderHook(() => useExperiences({}), {
      wrapper: ({ children }) => (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      ),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeDefined();
  });
});
```

**React 컴포넌트 테스트:**

```typescript
// src/components/__tests__/ExperienceCard.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ExperienceCard } from '../ExperienceCard';

describe('ExperienceCard', () => {
  it('경험 제목과 태그 렌더링', () => {
    render(
      <ExperienceCard
        experience={{
          id: '1',
          title: '프로젝트 리더',
          tags: ['리더십', '협업'],
        }}
      />
    );

    expect(screen.getByText('프로젝트 리더')).toBeInTheDocument();
    expect(screen.getByText('리더십')).toBeInTheDocument();
  });
});
```

### E2E 테스트

**대상:** 주요 사용자 시나리오

```typescript
// e2e/full-flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('핵심 사용자 흐름', () => {
  test('회원가입 → 경험 등록 → 분석 → 코칭', async ({ page }) => {
    // 1. 회원가입
    await page.goto('/signup');
    await page.fill('[name="email"]', 'e2e@test.com');
    await page.fill('[name="password"]', 'test1234!');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/experiences');

    // 2. 경험 등록
    await page.click('text=경험 추가');
    await page.fill('[name="title"]', 'E2E 테스트 경험');
    await page.fill('[name="situation"]', '테스트 상황');
    await page.fill('[name="action"]', '테스트 행동');
    await page.fill('[name="result"]', '테스트 결과');
    await page.click('button:has-text("저장")');
    await expect(page.locator('text=E2E 테스트 경험')).toBeVisible();

    // 3. 기업 분석
    await page.goto('/analysis');
    await page.fill('[name="jobUrl"]', 'https://example.com/job/123');
    await page.click('button:has-text("분석")');
    await expect(page.locator('[data-testid="analysis-result"]')).toBeVisible({
      timeout: 30000,
    });

    // 4. 코칭
    await page.click('text=자소서 코칭');
    await expect(page.locator('[data-testid="coaching-editor"]')).toBeVisible();
  });
});
```

---

## 3. 테스트 DB

### 설정 (프론트엔드: MSW 모킹, 백엔드: 로컬 DB)

```bash
# Go 백엔드 테스트: SQLite in-memory (enttest) 또는 로컬 PostgreSQL
# 프론트엔드 테스트: MSW로 Go API 응답 모킹 (실제 DB 불필요)
```

### Vitest 설정

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      include: ['src/lib/**', 'src/app/api/**'],
      exclude: ['src/types/**', 'src/test/**'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
```

### 테스트 셋업

```typescript
// src/test/setup.ts
import '@testing-library/jest-dom';
import { server } from './mocks/server';

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

---

## 4. AI 응답 모킹

### MSW (Mock Service Worker) 핸들러

```typescript
// src/test/mocks/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
  // OpenAI Chat Completion 모킹
  http.post('https://api.openai.com/v1/chat/completions', () => {
    return HttpResponse.json({
      choices: [{
        message: {
          content: JSON.stringify({
            tags: ['리더십', '문제해결'],
            confidence: 0.85,
          }),
        },
      }],
    });
  }),

  // OpenAI Embedding 모킹
  http.post('https://api.openai.com/v1/embeddings', () => {
    return HttpResponse.json({
      data: [{
        embedding: new Array(1536).fill(0.1),
      }],
    });
  }),

  // Anthropic Messages 모킹
  http.post('https://api.anthropic.com/v1/messages', () => {
    return HttpResponse.json({
      content: [{
        type: 'text',
        text: '기업 분석 결과: 이 기업은 IT 분야의 중견 기업으로...',
      }],
    });
  }),

  // DART API 모킹
  http.get('https://opendart.fss.or.kr/api/*', () => {
    return HttpResponse.json({
      status: '000',
      corp_name: '테스트기업',
      stock_code: '000000',
    });
  }),

  // 네이버 뉴스 검색 모킹
  http.get('https://openapi.naver.com/v1/search/news.json', () => {
    return HttpResponse.json({
      items: [{
        title: '테스트기업 관련 뉴스',
        link: 'https://example.com/news/1',
        description: '테스트기업이 신규 사업을 발표했다.',
      }],
    });
  }),
];
```

### MSW 서버 설정

```typescript
// src/test/mocks/server.ts
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);
```

### AI 응답 Fixture 파일

```
src/test/fixtures/
├── ai/
│   ├── weapon-tagging-response.json       # 무기 태깅 결과
│   ├── company-analysis-response.json     # 기업 분석 결과
│   ├── coaching-draft-response.json       # 초안 코칭 결과
│   ├── coaching-review-response.json      # 첨삭 코칭 결과
│   └── embedding-response.json            # 임베딩 벡터
├── crawling/
│   ├── jobkorea-sample.html               # 잡코리아 샘플 HTML
│   └── wanted-sample.html                 # 원티드 샘플 HTML
└── api/
    ├── dart-company-info.json             # DART 기업정보
    └── naver-news-results.json            # 네이버 뉴스 결과
```

**Fixture 사용 예시:**

```typescript
import weaponTaggingResponse from '@/test/fixtures/ai/weapon-tagging-response.json';

it('무기 태깅 결과 파싱', () => {
  const result = parseWeaponTags(weaponTaggingResponse);
  expect(result).toHaveLength(3);
  expect(result[0].category).toBe('리더십');
});
```

---

## 5. Phase별 최소 테스트

| Phase | 테스트 대상 | 최소 테스트 수 |
|-------|-----------|--------------|
| Phase 0 | Supabase 연결, 마이그레이션 적용 확인 | 3 |
| Phase 1 | 회원가입, 로그인, 로그아웃, 미인증 리다이렉트 | 5 |
| Phase 2 | 경험 CRUD, Zod 검증 | 8 |
| Phase 2.1 | 무기 태깅 파싱, AI 응답 처리 | 5 |
| Phase 3 | HTML 파싱 (잡코리아/캐치), 크롤링 결과 구조화 | 6 |
| Phase 3.1 | DART API 호출, 네이버 뉴스 변환 | 4 |
| Phase 3.2 | AI 분석 파이프라인, 캐시 히트/미스 | 5 |
| Phase 4 | 임베딩 유사도 계산, 경험 매칭 정렬 | 5 |
| Phase 5+ | 문항 분류, 코칭 프롬프트 조합, 버전 관리 | 8 |
| E2E | 핵심 사용자 흐름 (회원가입 → 코칭) | 3 |

---

## 6. GitHub Actions CI

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  ci:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      # 1. Lint
      - name: Lint
        run: pnpm run lint

      # 2. Type check
      - name: Type check
        run: pnpm exec tsc --noEmit

      # 3. Unit & Integration tests
      - name: Test
        run: pnpm exec vitest run --coverage

      # 4. Build
      - name: Build
        run: pnpm run build
        env:
          NEXT_PUBLIC_SUPABASE_URL: ${{ secrets.NEXT_PUBLIC_SUPABASE_URL }}
          NEXT_PUBLIC_SUPABASE_ANON_KEY: ${{ secrets.NEXT_PUBLIC_SUPABASE_ANON_KEY }}

  e2e:
    runs-on: ubuntu-latest
    needs: ci

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Install Playwright browsers
        run: pnpm exec playwright install --with-deps

      - name: Start Supabase local
        uses: supabase/setup-cli@v1
      - run: supabase start

      - name: Run E2E tests
        run: pnpm exec playwright test
        env:
          NEXT_PUBLIC_SUPABASE_URL: http://127.0.0.1:54321
          NEXT_PUBLIC_SUPABASE_ANON_KEY: ${{ secrets.LOCAL_ANON_KEY }}
```

### CI 파이프라인 순서

```
PR 생성 / push
    │
    ├── 1. lint          (ESLint)
    ├── 2. type-check    (tsc --noEmit)
    ├── 3. test          (Vitest + coverage)
    ├── 4. build         (next build)
    │
    └── 5. e2e           (Playwright, ci job 성공 후)
```

### npm scripts

```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "lint": "next lint",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui"
  }
}
```

---

## 7. 커버리지 목표

| 대상 | 목표 커버리지 | 비고 |
|------|-------------|------|
| `src/lib/**` (비즈니스 로직) | **60%** | AI 파싱, 유틸리티, 검증 |
| `src/app/api/**` (API Routes) | **50%** | 주요 엔드포인트 |
| `src/components/**` (UI) | **30%** | 핵심 컴포넌트만 |
| 전체 | **50%** | - |

**커버리지에서 제외:**

- `src/types/` — 타입 정의만
- `src/test/` — 테스트 헬퍼
- `*.config.*` — 설정 파일
- `src/app/**/layout.tsx`, `page.tsx` — 단순 레이아웃

### Playwright 설정

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 60000,
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: 'http://localhost:4000',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'pnpm run dev',
    port: 3000,
    reuseExistingServer: !process.env.CI,
  },
});
```

---

## 8. Go 백엔드 테스트

### 테스트 프레임워크

- **단위 테스트**: Go 표준 `testing` 패키지 + `testify/assert`
- **HTTP 테스트**: `net/http/httptest` 기반 API 통합 테스트
- **Ent 테스트**: `enttest` 패키지로 SQLite in-memory DB 사용

### 테스트 구조

```text
apps/backend/
├── internal/
│   ├── service/
│   │   └── experience_test.go     # Service 단위 테스트
│   └── controller/
│       └── experience_test.go     # Controller 통합 테스트
└── ent/
    └── enttest/                   # Ent 스키마 테스트
```

### AI 클라이언트 모킹

```go
// AI 클라이언트 인터페이스 정의
type AIClient interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// 테스트용 Mock 구현
type MockAIClient struct {
    Response *ChatResponse
    Err      error
}

func (m *MockAIClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    return m.Response, m.Err
}
```

### 실행 커맨드

```bash
# 전체 테스트
go test ./...

# 특정 패키지
go test ./internal/service/...

# 커버리지
go test -cover ./...

# Race condition 검사
go test -race ./...
```
