# Web

Next.js 15 frontend. See root CLAUDE.md for project rules.

## Structure

```
src/
  app/(auth)/        # Login, signup, callback
  app/(main)/        # Dashboard, experiences, analysis, coaching
  app/(admin)/       # Admin pages
  api/generated/     # @hey-api/openapi-ts (do not edit)
  components/        # auth/, layout/, common/, admin/
  hooks/             # React Query hooks
  stores/            # Zustand (auth-store)
  lib/               # api-client, schemas, utilities
```

## Rules

- Tailwind only, no CSS Modules
- Server Components for data, Client Components (`'use client'`) for interactivity
- Auth: JWT in Zustand (localStorage) → Axios interceptor

## Commands

```bash
moon run web:dev              # Dev (port 4000)
moon run web:build            # Build
moon run web:lint             # ESLint
moon run web:typecheck        # tsc --noEmit
moon run web:test             # vitest
moon run web:generate-client  # OpenAPI → TS client
```

## Testing

Vitest + @testing-library/react + jsdom. Tests in `__tests__/*.test.ts(x)`.
