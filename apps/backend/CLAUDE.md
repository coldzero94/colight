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

## TDD Rules

### Workflow

1. **Red**: Write `_test.go` ONLY → `moon run backend:test` → confirm FAIL
2. **Green**: Write minimal `.go` → `moon run backend:test` → confirm PASS
3. Never write test + implementation together

### Test Structure

```
service/foo.go         → service/foo_test.go        (unit, mock dependencies)
controller/foo.go      → controller/foo_test.go     (integration, httptest + gin)
```

### Patterns

**Service test** (unit — mock AI, mock DB via interface):
```go
func TestFoo_Success(t *testing.T) {
    mockAI := new(ai.MockAIClient)
    svc := NewFooService(mockAI)
    mockAI.On("ChatCompletion", mock.Anything, mock.Anything).
        Return(&ai.ChatResponse{Content: `{"result":"ok"}`}, nil)

    result, err := svc.DoSomething(context.Background(), input)
    require.NoError(t, err)       // prerequisite — fail fast
    assert.Equal(t, expected, result) // assertion — continue on fail
}
```

**Controller test** (integration — httptest + real gin router):
```go
func TestFoo_Unauthorized(t *testing.T) {
    router := setupTestRouter(t) // no auth token
    req := httptest.NewRequest(http.MethodPost, "/v1/foo", body)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

### Conventions

- `require` for preconditions (fail immediately), `assert` for checks (continue)
- DB tests: `enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")` per test
- AI fixtures in `testdata/` — don't hardcode JSON in tests
- Helpers in `testutil/` (NewTestClient, SeedTestUser, SeedWeapons)
- Don't test: `internal/generated/`, `cmd/`, `scripts/`

## Migration

Edit `ent/schema/*.go` → `generate-ent` → `migrate-diff` → review SQL → `migrate-apply`
