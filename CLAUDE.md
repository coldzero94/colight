# Colight

AI cover letter coaching platform (Korean market). Job posting URL → company analysis → experience matching → STAR-structured draft coaching.

## Monorepo Structure

```text
apps/backend/    # Go 1.24 (Gin + Ent + oapi-codegen)
apps/web/        # Next.js 15 (App Router, TypeScript strict)
packages/protocol/  # TypeSpec → OpenAPI (API single source of truth)
docs/            # Plan, develop, phase guides, research
```

## Architecture

- **Code generation pipeline**: TypeSpec → OpenAPI → Go server (oapi-codegen) + TS client (@hey-api/openapi-ts)
- **Source of truth**: Ent schemas for DB (`02-data-structure.md` is canonical), TypeSpec for API
- **Generated code is committed** — do not manually edit `internal/generated/` or `src/api/generated/`
- **Auth**: Supabase Auth on frontend only, Go backend verifies JWT (no Supabase SDK on backend)
- **Migrations**: Atlas (not Supabase CLI)
- **AI prompts**: loaded from `prompt_templates` table — never hardcode
- **AI models**: Claude Sonnet 4.5 (analysis/coaching), GPT-4.1 mini (parsing/tagging), text-embedding-3-small (embeddings)
- **Go AI SDK**: `github.com/openai/openai-go` (NOT sashabaranov/go-openai)
- **Background jobs**: River (embedded mode, PostgreSQL-based)

## Conventions

- **UI text**: Korean. **Code, comments, commits**: English
- **Styling**: Tailwind CSS only — no CSS Modules
- **State**: React Query (server), Zustand (client, minimal), React Hook Form + Zod (forms)
- **Commit format**: `Phase X.Y: description`
- **Deploy**: Koyeb (Go), Vercel (Next.js), Supabase (PostgreSQL + Auth)

## Phase Completion Rules

### Before starting
1. Read phase doc (`docs/develop/phases/phase-X.md`)
2. Verify prerequisite phases are complete
3. Update `phases/README.md` status to in-progress

### During implementation
1. Follow phase doc checklist
2. Update checkboxes as steps complete
3. Commit format: `Phase X.Y: description`

### Before marking complete — verification gate
1. **Backend**: `cd apps/backend && golangci-lint run ./... && go test ./...`
2. **Frontend**: `cd apps/web && pnpm run lint && pnpm exec tsc --noEmit && pnpm run build`
3. Or use moon from root: `moon run :lint && moon run :test && moon run web:build`
4. Fix all lint errors, type errors, and build failures before proceeding

### After verification passes
1. Confirm all phase doc checkboxes are checked
2. Update `phases/README.md` status to complete
3. Update Phase Progress table below

## Commands

**⚠️ CRITICAL: Always use `moon` from project root for ALL operations.**

This includes:
- Running dev servers
- Building projects
- Running tests
- Code generation
- Database operations
- **Even when testing or verifying steps**

❌ **NEVER do this:**
```bash
cd apps/backend && go run ./cmd/api
cd apps/backend && go test ./...
cd apps/web && pnpm run dev
```

✅ **ALWAYS do this:**
```bash
moon run backend:dev
moon run backend:test
moon run web:dev
```

```bash
# Development (all commands from project root)
moon run :dev                     # Start all dev servers (backend + frontend, parallel)
moon run backend:dev              # Start Go API server only (port 9000)
moon run web:dev                  # Start Next.js only (port 4000)

# Code Generation
moon run protocol:generate        # TypeSpec → OpenAPI
moon run backend:generate-api     # OpenAPI → Go server code
moon run web:generate-client      # OpenAPI → TS client
moon run backend:generate-ent     # Regenerate Ent code

# Database Migrations
moon run backend:migrate-diff -- name=<description>   # Create migration
moon run backend:migrate-apply                        # Apply migrations
moon run backend:seed                                 # Insert seed data

# Testing & Quality
moon run :test                    # Run all tests
moon run :lint                    # Lint all projects
moon run backend:test             # Go tests only
moon run web:test                 # Next.js tests only
moon run web:typecheck            # TypeScript type check

# Build
moon run :build                   # Build all projects
moon run backend:build            # Build Go binary
moon run web:build                # Build Next.js

# Direct Commands (reference only - prefer moon)
# Backend: cd apps/backend && air (or go run ./cmd/api)
# Frontend: cd apps/web && pnpm run dev
# Database: docker compose up -d
```

## Phase Progress

| Phase | Name | Status | Sprint |
|-------|------|--------|--------|
| 0 | Project setup & infra | Pending | 0 |
| 1 | Auth & layout | Pending | 0 |
| 2 | Experience CRUD | Pending | 1 |
| 2.1 | Weapon auto-tagging | Pending | 1 |
| 3 | Crawling & parsing | Pending | 2 |
| 3.1 | Company data API | Pending | 2 |
| 3.2 | AI company analysis | Pending | 3 |
| 3.3 | Analysis report UI | Pending | 3 |
| 4 | Experience matching | Pending | 4 |
| 5 | Question analysis | Pending | 5 |
| 5.1 | Draft coaching | Pending | 5 |
| 5.2 | Coaching editor | Pending | 5 |
| 6 | Revision coaching | Pending | 6 |
| 6.1 | Premium & polish | Pending | 6 |
| 6.2 | Landing & beta | Pending | 6 |
| 7 | AI interview | Pending | 7 |
| 7.1 | Experience recommendation | Pending | 7 |
| 8 | Dashboard kanban | Pending | 7-8 |
| 8.1 | Version management | Pending | 8 |
| 9 | Payment integration | Pending | 8 |
| 10 | Growth features | Pending | 9+ |

## Known Issues

- **Phase docs have TypeScript remnants** (Phase 3–6.1): architecture change notes added at top. Implement in Go, not TypeScript.
- **Schema duality**: `02-data-structure.md` is canonical if it conflicts with phase-0 doc.
