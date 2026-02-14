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
- **Auth**: Go backend handles Naver OAuth 2.0 + Email/Password directly, JWT (HMAC-SHA256) for access/refresh tokens
- **Migrations**: Atlas (not Supabase CLI)
- **AI prompts**: loaded from `prompt_templates` table — never hardcode
- **AI models**: Claude Sonnet 4.5 (analysis/coaching), Gemini Flash / Groq Llama 3.3 (parsing/tagging, configurable via `LLM_LIGHT_PROVIDER` env), text-embedding-3-small (embeddings)
- **Go AI SDK**: Common LLM abstraction (`internal/infrastructure/ai/`) supporting multiple providers (Anthropic, Google Gemini, Groq)
- **Background jobs**: River (embedded mode, PostgreSQL-based)

## Conventions

- **UI text**: Korean. **Code, comments, commits**: English
- **Styling**: Tailwind CSS only — no CSS Modules
- **State**: React Query (server), Zustand (client, minimal), React Hook Form + Zod (forms)
- **Commit format**: `Phase X.Y: description`
- **Deploy**: Koyeb (Go), Vercel (Next.js), Supabase (PostgreSQL only)

## Commands

**CRITICAL: Always use `moon` from project root for ALL operations.** Never `cd` into subdirectories to run commands. See each sub-project CLAUDE.md for full command reference.

```bash
# Essential shortcuts
moon run :dev          # All dev servers
moon run :test         # All tests
moon run :lint         # All linting
moon run :build        # All builds
```

## Phase Completion Rules

1. Read phase doc (`docs/develop/phases/phase-X.md`), verify prerequisites
2. Follow phase doc checklist, update checkboxes, commit as `Phase X.Y: description`
3. **Verification gate** — all must pass before marking complete:
   ```bash
   moon run backend:lint && moon run backend:test && moon run web:lint && moon run web:typecheck && moon run web:test && moon run web:build
   ```
4. Update `phases/README.md` status + Phase Progress table below

## Phase Progress

| Phase | Name | Status | Sprint |
|-------|------|--------|--------|
| 0 | Project setup & infra | Complete | 0 |
| 1 | Auth & layout | Complete | 0 |
| 2 | Experience CRUD | Complete | 1 |
| 2.1 | Weapon auto-tagging | Complete | 1 |
| 3 | Crawling & parsing | Complete | 2 |
| 3.1 | Company data API | Complete | 2 |
| 3.2 | AI company analysis | Complete | 3 |
| 3.3 | Analysis report UI | Complete | 3 |
| 4 | Experience matching | Complete | 4 |
| 5 | Question analysis | Complete | 5 |
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

## Ralph Loop

### Rules
- Run full verification gate before outputting any completion promise — zero errors required
- Format: `<promise>PROMISE_TEXT</promise>` — ONLY when verifiably TRUE
- NEVER output a false promise to escape the loop

### TDD Cycle
1. Write failing test → `moon run :test` (red)
2. Implement minimal code → `moon run :test` (green)
3. `moon run :lint` → refactor if needed → commit `Phase X.Y: description`
4. Repeat. Output promise only when ALL requirements have passing tests

### Guardrails
- ALWAYS read files before editing — never assume content
- NEVER skip failing tests. Do NOT refactor unrelated code
- Re-read task requirements before declaring completion
- If stuck: document in `BLOCKED.md`, let `--max-iterations` handle exit
