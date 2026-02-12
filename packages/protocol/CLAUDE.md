# Protocol

TypeSpec API definitions — single source of truth for all API contracts. See root CLAUDE.md for project-wide rules.

## Pipeline

```text
TypeSpec (.tsp) → OpenAPI (yaml) → Go server (oapi-codegen) + TS client (@hey-api/openapi-ts)
```

All API changes start here. Never edit generated OpenAPI directly.

## Structure

```text
src/         # *.tsp files (domain-organized: common, experience, analysis, coaching)
tsp-output/  # Generated OpenAPI spec (committed, do not edit)
main.tsp     # Entrypoint
```

## Key Rules

- All routes use `/v1/` prefix
- Use `@encodedName` for snake_case JSON fields (TypeSpec uses camelCase)
- Error responses use `Common.ErrorResponse` or `Common.ValidationError`
- **Non-breaking**: new endpoints, optional fields/params
- **Breaking** (needs `/v2/`): field removal/rename, type changes, new required params

## Commands

```bash
pnpm run generate      # TypeSpec → OpenAPI
pnpm run validate      # Validate OpenAPI spec
pnpm run generate:all  # Full pipeline (from project root)
```
