# CLAUDE.md

## Project Overview

Compass is a context engine that builds temporal knowledge graphs. Written in Go (1.25+), it uses PostgreSQL with pgvector, pg_trgm, and tsvector extensions. Single binary with no external dependencies beyond Postgres.

## API Development Workflow

Proto definitions live in a separate repo: `raystack/proton` (at `../proton` relative to compass). Never edit files in `gen/` directly.

**Workflow for API changes:**
1. Edit proto definitions in the proton repo first
2. Regenerate: `buf generate ../proton --template buf.gen.yaml --path raystack/compass`
3. Update compass handler/service code to match

Alternative: `make proto` fetches proton from GitHub using the PROTON_COMMIT pinned in the Makefile.

Generated code lands in `gen/raystack/compass/v1beta1/`. The API layer uses ConnectRPC (not plain gRPC) -- handlers implement the connect service interface.

## Architecture

- `core/` -- Domain logic (entity, principal, document, embedding, namespace, pipeline)
- `store/` -- PostgreSQL repositories and migrations
- `handler/` -- ConnectRPC handlers
- `internal/mcp/` -- MCP server for AI agent tool access
- `internal/middleware/` -- Connect interceptors (namespace, principal, logging, recovery)
- `internal/server/` -- Server bootstrap and wiring
- `internal/config/` -- Configuration
- `cli/` -- CLI commands
- `gen/` -- Auto-generated protobuf/connect code (DO NOT EDIT)

## Key Conventions

- Identity model uses `Principal` (not User/Caller) -- supports types: user, agent, service
- Multi-tenancy via `Namespace` with PostgreSQL Row-Level Security
- Temporal data: entities and edges use `valid_from`/`valid_to` timestamps
- Soft deletes: set `valid_to = now()` instead of DELETE
- Database migrations in `store/migrations/` -- sequential numbering (000001, 000002, etc.)
- MCP server reads namespace and principal from context (injected by middleware)

## Common Commands

```
make build          # Build binary
make test           # Run tests with coverage
make proto          # Regenerate proto from proton (GitHub, uses PROTON_COMMIT)
make lint           # Run golangci-lint
go vet ./...        # Run vet
go build ./...      # Check compilation
```

Local proto generation from a local proton checkout:
```
buf generate ../proton --template buf.gen.yaml --path raystack/compass
```

## Do NOT

- Edit files in `gen/` directly
- Add Elasticsearch, Kafka, or other infrastructure dependencies -- Postgres only
- Use `User` or `Caller` naming -- the identity type is `Principal`
- Skip database migrations -- always add a new migration file for schema changes
