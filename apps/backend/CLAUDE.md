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
```

## Key Rules

- Controller → Service → Ent (never skip layers)
- No RLS — filter by `user_id` in service layer
- AI prompts from DB (`prompt_templates`) — never hardcode
- AI SDK: `github.com/openai/openai-go` (NOT sashabaranov)
- All routes use `/v1/` prefix
- Generated `internal/generated/` is committed — do not edit manually
- FK constraints use CASCADE DELETE

## Commands

```bash
# Use moon from project root (preferred): moon run backend:<task>
air                           # Dev server (hot reload)
go build -o bin/api ./cmd/api # Build binary
go generate ./ent             # Regenerate Ent code
oapi-codegen -config oapi-codegen.yaml ../../packages/protocol/tsp-output/openapi/openapi.yaml  # OpenAPI → Go
atlas migrate diff <name> --dir file://migrations --to ent://ent/schema --dev-url "docker://postgres/16/dev?search_path=public"
atlas migrate apply --dir file://migrations --url "$DATABASE_URL"
go run ./scripts/seed.go all  # Seed data
go test ./...                 # Tests
golangci-lint run ./...       # Lint
```

## Migration Workflow

1. Edit `ent/schema/*.go`
2. `go generate ./ent`
3. `atlas migrate diff <name> --dir file://migrations --to ent://ent/schema --dev-url "docker://postgres/16/dev?search_path=public"`
4. Review generated SQL
5. `atlas migrate apply --dir file://migrations --url "$DATABASE_URL"`
