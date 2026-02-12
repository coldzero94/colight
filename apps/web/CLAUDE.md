# Web

Next.js 15 frontend for Colight. See root CLAUDE.md for project-wide rules.

## Architecture

```text
src/
  app/(auth)/        # Login, signup, callback
  app/(main)/        # Dashboard, experiences, analysis, coaching (protected)
  api/generated/     # @hey-api/openapi-ts output (committed, do not edit)
  components/ui/     # shadcn/ui components
  hooks/             # React Query hooks
  stores/            # Zustand stores
  lib/api/client.ts  # Axios client with JWT interceptor
```

## Key Rules

- **Tailwind CSS only** — no CSS Modules
- **State**: React Query (server), Zustand (client, minimal), React Hook Form + Zod (forms)
- **Auth flow**: Supabase Auth (frontend) → JWT attached via Axios interceptor → Go backend verifies
- **UI text in Korean**, code/comments in English
- Generated `src/api/generated/` is committed — do not edit manually
- Use shadcn/ui components as base
- Server Components for data fetching, Client Components (`'use client'`) for interactivity

## Commands

```bash
pnpm run dev                  # Dev server
pnpm run build                # Build
pnpm run test                 # Tests
pnpm run lint                 # Lint
pnpm run generate:client      # OpenAPI → TS client
```
