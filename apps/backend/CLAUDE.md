# Backend

Go API server for Colight. See root CLAUDE.md for project-wide rules.

## Architecture

```text
cmd/api/           # Entrypoint
ent/schema/        # DB schema (source of truth)
internal/
  controller/      # HTTP handlers (implements generated StrictServerInterface)
  service/         # Business logic
  infrastructure/  # External integrations (AI, DB, middleware, external APIs)
  generated/       # oapi-codegen output (committed, do not edit)
migrations/        # Atlas SQL migrations
scripts/           # Seed data
testutil/          # Shared test helpers
```

## Key Rules

- Controller → Service → Ent (never skip layers)
- No RLS — filter by `user_id` in service layer
- AI prompts from DB (`prompt_templates`) — never hardcode
- AI SDK: `github.com/openai/openai-go` (NOT sashabaranov)
- All routes use `/v1/` prefix
- Generated `internal/generated/` is committed — do not edit manually
- FK constraints use CASCADE DELETE

## Commands (always from project root)

```bash
moon run backend:dev              # Dev server (hot reload via air)
moon run backend:build            # Build binary
moon run backend:lint             # golangci-lint run ./...
moon run backend:test             # go test ./...
moon run backend:generate-ent     # Regenerate Ent code
moon run backend:generate-api     # OpenAPI → Go server code
moon run backend:migrate-diff -- name=<desc>  # Create migration
moon run backend:migrate-apply    # Apply migrations
moon run backend:seed             # Seed data (all | admin | prompts)
```

## Verification Gate

**Run after every code change, before marking any phase complete:**

```bash
moon run backend:lint && moon run backend:test
```

Both must pass with **zero errors**. Do not skip. Do not proceed to the next phase if tests fail.

## Testing Conventions

- Framework: Go standard `testing` + `github.com/stretchr/testify`
- Test files: live next to source — `foo.go` → `foo_test.go`
- Shared helpers: `testutil/` package
- Use table-driven tests where appropriate
- **Every new service/controller/middleware must have tests**
- Run `moon run backend:test` after writing or modifying any code

## Migration Workflow

1. Edit `ent/schema/*.go`
2. `moon run backend:generate-ent`
3. `moon run backend:migrate-diff -- name=<description>`
4. Review generated SQL
5. `moon run backend:migrate-apply`
