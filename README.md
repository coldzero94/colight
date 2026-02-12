# Colight

AI cover letter coaching platform (Korean market).

## Quick Start (Local Development)

### Prerequisites
- Node.js 20+
- Go 1.24+
- Docker & Docker Compose
- pnpm 9+
- moon (task runner)

### 1. Install dependencies
```bash
pnpm install
```

### 2. Start PostgreSQL
```bash
docker compose up -d
```

### 3. Start all dev servers with moon
```bash
# Start all dev servers in parallel (backend + frontend)
moon run :dev

# Or start individually:
moon run backend:dev    # API server (port 9000)
moon run web:dev        # Next.js (port 4000)
```

### 4. Access
- **API**: http://localhost:9000/health
- **Frontend**: http://localhost:4000
- **Database**: localhost:5532 (postgres/password)

## Moon Commands

moon is our unified task runner. All commands run from project root.

### Development
```bash
moon run :dev              # Start all dev servers (persistent, parallel)
moon run backend:dev       # Start Go API server only
moon run web:dev           # Start Next.js only
```

### Code Generation
```bash
moon run protocol:generate          # TypeSpec → OpenAPI
moon run backend:generate-api       # OpenAPI → Go server code
moon run web:generate-client        # OpenAPI → TS client
moon run backend:generate-ent       # Regenerate Ent code
```

### Database Migrations
```bash
moon run backend:migrate-diff -- name=add_feature   # Create migration
moon run backend:migrate-apply                      # Apply migrations
moon run backend:seed                               # Insert seed data
```

### Testing & Quality
```bash
moon run :test             # Run all tests
moon run :lint             # Lint all projects
moon run backend:test      # Go tests only
moon run web:test          # Next.js tests only
moon run web:typecheck     # TypeScript type check
```

### Build
```bash
moon run :build            # Build all projects
moon run backend:build     # Build Go binary
moon run web:build         # Build Next.js
```

## Project Structure

```
colight/
├── apps/
│   ├── backend/              # Go API server (port 9000)
│   └── web/                  # Next.js frontend (port 4000)
├── packages/
│   └── protocol/             # TypeSpec API definitions
├── docs/                     # Documentation
├── .moon/                    # Moon configuration
├── docker-compose.yml        # PostgreSQL (port 5532)
└── CLAUDE.md                # Development guidelines
```

## Tech Stack

- **Backend**: Go 1.24 + Gin + Ent ORM + Atlas migrations
- **Frontend**: Next.js 15 + TypeScript + Tailwind CSS + shadcn/ui
- **API Protocol**: TypeSpec → OpenAPI → oapi-codegen (Go) + @hey-api (TS)
- **Database**: PostgreSQL 16 + pgvector
- **Task Runner**: moon (moonrepo)
- **Package Manager**: pnpm

## Documentation

- [CLAUDE.md](./CLAUDE.md) - Project-wide development rules
- [docs/plan/](./docs/plan/) - Planning documents
- [docs/develop/](./docs/develop/) - Technical documentation
- [docs/develop/phases/](./docs/develop/phases/) - Phase implementation guides

## References

- [moon documentation](https://moonrepo.dev/docs)
- [moon run command](https://moonrepo.dev/docs/commands/run)
- [moon tasks](https://moonrepo.dev/docs/concepts/task)