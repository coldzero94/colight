# Web

Next.js 15 frontend for Colight. See root CLAUDE.md for project-wide rules.

## Architecture

```text
src/
  app/(auth)/        # Login, signup, callback
  app/(main)/        # Dashboard, experiences, analysis, coaching (protected)
  app/(admin)/       # Admin pages (admin-only)
  app/auth/callback/ # OAuth callback handler
  api/generated/     # @hey-api/openapi-ts output (committed, do not edit)
  components/
    auth/            # AuthGuard, AdminGuard
    layout/          # Sidebar, Header, MobileNav, UserDropdown
    common/          # LoadingSpinner, EmptyState
    admin/           # AdminSidebar
  hooks/             # React Query hooks
  stores/            # Zustand stores (auth-store)
  lib/               # api-client, auth schemas, utilities
  test/              # Test setup (setup.ts)
```

## Key Rules

- **Tailwind CSS only** — no CSS Modules
- **State**: React Query (server), Zustand (client, minimal), React Hook Form + Zod (forms)
- **Auth flow**: Go backend handles OAuth/Email auth → JWT stored in Zustand (localStorage persist) → Axios interceptor attaches token
- **UI text in Korean**, code/comments in English
- Generated `src/api/generated/` is committed — do not edit manually
- Server Components for data fetching, Client Components (`'use client'`) for interactivity

## Commands (always from project root)

```bash
moon run web:dev              # Dev server (port 4000)
moon run web:build            # Build (runs lint + typecheck first)
moon run web:lint             # ESLint
moon run web:typecheck        # tsc --noEmit
moon run web:test             # vitest run
moon run web:generate-client  # OpenAPI → TS client
```

## Verification Gate

**Run after every code change, before marking any phase complete:**

```bash
moon run web:lint && moon run web:typecheck && moon run web:test && moon run web:build
```

All must pass with **zero errors**. Do not skip. Do not proceed to the next phase if any check fails.

## Testing Conventions

- Framework: Vitest + @testing-library/react + jsdom
- Config: `vitest.config.ts` (jsdom environment, `@/` alias, setup file)
- Setup: `src/test/setup.ts` (imports @testing-library/jest-dom matchers)
- Test files: `src/**/__tests__/*.test.ts(x)` (co-located `__tests__` directories)
- **Every new store, lib utility, and component with logic must have tests**
- Run `moon run web:test` after writing or modifying any code
