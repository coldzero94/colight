# Backend

Go API server. See root CLAUDE.md for project rules.

## Structure

```
cmd/api/           # Entrypoint
ent/schema/        # DB schema (source of truth)
internal/
  controller/      # HTTP handlers (StrictServerInterface)
  service/         # Business logic
  infrastructure/  # AI, DB, middleware, external APIs
  generated/       # oapi-codegen (do not edit)
testutil/          # Shared test helpers
```

## Rules

- Controller → Service → Ent (never skip layers)
- No RLS — filter by `user_id` in service layer
- AI: `LLMProvider` interface (`internal/infrastructure/ai/`)
- Routes: `/v1/` prefix. FK: CASCADE DELETE

## Commands

```bash
moon run backend:dev              # Dev (air hot reload)
moon run backend:build            # Build binary
moon run backend:lint             # golangci-lint
moon run backend:test             # go test ./...
moon run backend:generate-ent     # Ent codegen
moon run backend:generate-api     # OpenAPI → Go server
moon run backend:migrate-diff -- name=<desc>
moon run backend:migrate-apply
moon run backend:seed             # Seed data
```

## Testing

Go `testing` + testify. Tests co-located: `foo.go` → `foo_test.go`. Helpers in `testutil/`.

## Migration

Edit `ent/schema/*.go` → `generate-ent` → `migrate-diff` → review SQL → `migrate-apply`
