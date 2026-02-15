# Colight

AI cover letter coaching platform (Korean market).

## Structure

```
apps/backend/       # Go 1.24 (Gin + Ent + oapi-codegen)
apps/web/           # Next.js 15 (App Router, TypeScript strict)
packages/protocol/  # TypeSpec → OpenAPI (API single source of truth)
docs/               # Plan, develop, phase guides
```

## Architecture

- **Codegen**: TypeSpec → OpenAPI → Go server (oapi-codegen) + TS client (@hey-api/openapi-ts)
- **Source of truth**: Ent schemas for DB, TypeSpec for API
- **Generated code committed** — never edit `internal/generated/` or `src/api/generated/`
- **Auth**: Go backend (Naver OAuth + Email/Password), JWT HMAC-SHA256
- **AI**: Claude Sonnet 4.5 (heavy), Gemini Flash / Groq Llama 3.3 (light, `LLM_LIGHT_PROVIDER` env), text-embedding-3-small (embeddings). Prompts from DB `prompt_templates` — never hardcode
- **Jobs**: River (embedded, PostgreSQL-based)

## Conventions

- **UI text**: Korean. **Code/comments/commits**: English
- **Styling**: Tailwind only. **State**: React Query + Zustand + RHF/Zod
- **Commit**: `Phase X.Y: description`
- **Commands**: Always `moon` from project root — never `cd` into subdirs

```bash
moon run :dev    # All dev servers
moon run :test   # All tests
moon run :lint   # All linting
moon run :build  # All builds
```

## TDD & Phase Workflow

1. Read phase doc (`docs/develop/phases/phase-X.md`)
2. **Red**: Write test ONLY → run test → **confirm FAIL** (must see failure output)
3. **Green**: Write minimal implementation ONLY → run test → **confirm PASS**
4. **Never write test and implementation in the same step** — always verify failure first
5. **Gate**: `moon run backend:lint && moon run backend:test && moon run web:lint && moon run web:typecheck && moon run web:test && moon run web:build`
6. **Commit**: `git add <files> && git commit -m "Phase X.Y: description"`
7. Repeat steps 2-6 per step. Update Phase Progress table when phase completes

> TDD patterns per stack: see `apps/backend/CLAUDE.md` and `apps/web/CLAUDE.md`

## Phase Progress

| Phase | Name | Status |
|-------|------|--------|
| 0 | Project setup & infra | Complete |
| 1 | Auth & layout | Complete |
| 2 | Experience CRUD | Complete |
| 2.1 | Weapon auto-tagging | Complete |
| 3 | Crawling & parsing | Complete |
| 3.1 | Company data API | Complete |
| 3.2 | AI company analysis | Complete |
| 3.3 | Analysis report UI | Complete |
| 4 | Experience matching | Complete |
| 5 | Question analysis | Complete |
| 5.1 | Draft coaching | Complete |
| 5.2 | Coaching editor | Complete |
| 6 | Revision coaching | Complete |
| 6.1 | Premium & polish | Complete |
| 6.2 | Landing & beta | Complete |
| 7 | AI interview | Complete |
| 7.1 | Experience recommendation | Complete |
| 8 | Dashboard kanban | Complete |
| 8.1 | Version management | Complete |
| 9 | Payment integration | Pending |
| 10 | Growth features | Pending |

## Known Issues

- Phase docs 3-6.1 have TypeScript remnants — implement in Go
- `02-data-structure.md` is canonical if schema conflicts exist

## Loop Rules (Ralph / Continuous)

- NEVER skip verification gate. NEVER output false completion promises
- ALWAYS read files before editing. NEVER refactor unrelated code
- If stuck: document in `BLOCKED.md`, move to next possible task
- Progress state lives in files (SHARED_TASK_NOTES.md), not conversation
