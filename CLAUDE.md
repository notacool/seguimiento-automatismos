# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

API REST in Go for centralized management of automation states from multiple teams. Built following **Clean Architecture (Hexagonal/Ports & Adapters)** principles with API-First design using OpenAPI 3.0.

## Technology Stack

- **Go 1.21+** with **Gin** framework
- **PostgreSQL 16** with pg_cron extension
- **pgx/v5** for database connectivity (connection pooling via pgxpool)
- **Docker & Docker Compose** for containerization
- **golang-migrate** for database migrations
- **Python CLI** (Click) with cross-platform binary support

## Common Commands

### Development
```bash
make deps              # Download Go dependencies
make build             # Compile binary to bin/api.exe
make run               # Run application locally
make test              # Run tests with race detector
make test-coverage     # Generate and open HTML coverage report
make fmt               # Format code
make lint              # Run golangci-lint
```

### Docker
```bash
make docker-up         # Start services (docker-compose up -d)
make docker-down       # Stop services
make docker-logs       # Follow logs
make docker-build      # Build Docker image
```

### Database Migrations
```bash
make migrate-up                        # Apply migrations
make migrate-down                      # Revert migrations
make migrate-create NAME=table_name    # Create new migration
```

Migrations are located in `internal/adapter/repository/postgres/migrations/` and require `DATABASE_URL` environment variable.

### Python CLI
```bash
make cli-deps                # Install Python dependencies
make cli-build-windows       # Generate Windows executable
make cli-build-linux         # Generate Linux executable
```

### Testing Single Package
```bash
go test -v ./internal/domain/entity/...
go test -v ./internal/usecase/task/...
```

## Architecture

The project follows Clean Architecture with clear separation of concerns:

```
internal/
├── domain/                      # Core business logic (innermost layer)
│   ├── entity/                 # Business entities: Task, Subtask
│   ├── service/                # Domain services: StateMachine (state transitions)
│   └── repository/             # Repository interfaces (ports)
├── usecase/                     # Application layer (use cases)
│   ├── task/                   # Task-related use cases
│   └── subtask/                # Subtask-related use cases
├── adapter/                     # External adapters (outermost layer)
│   ├── handler/http/           # HTTP handlers (Gin)
│   └── repository/postgres/    # PostgreSQL implementations
└── infrastructure/              # Cross-cutting concerns
    ├── config/                 # Configuration from env vars
    └── database/               # Database connection pool
```

### Dependency Flow
Dependencies point inward: `adapter` → `usecase` → `domain`. The domain layer has zero external dependencies.

### Key Architectural Patterns

1. **Dependency Injection**: Database pool (`*pgxpool.Pool`) is injected from main into handlers
2. **Repository Pattern**: Domain defines interfaces, adapters provide implementations
3. **State Machine**: Managed by domain service with strict transition rules
4. **Configuration**: All config loaded from environment variables via `config.Load()`

## Application Bootstrap

Entry point is [cmd/api/main.go](cmd/api/main.go):
1. Load configuration from environment variables
2. Create PostgreSQL connection pool with context
3. Setup router via `httpHandler.SetupRouter(dbPool, ginMode)`
4. Start HTTP server with graceful shutdown (5s timeout)

## Configuration

Environment variables (see [.env.example](.env.example)):
- `PORT` - Server port (default: 8080)
- `GIN_MODE` - debug/release
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME` - PostgreSQL connection
- `DATABASE_SSLMODE` - SSL mode (disable/require)
- `DATABASE_MAX_CONNS`, `DATABASE_MIN_CONNS` - Connection pool sizing (default: 25/5)
- `DATABASE_MAX_CONN_LIFETIME`, `DATABASE_MAX_CONN_IDLE_TIME` - Pool timeouts (default: 5m/1m)

Alternative: Use `DATABASE_URL` connection string.

## State Management

Task states: `PENDING` → `IN_PROGRESS` → `COMPLETED`/`FAILED`/`CANCELLED`

Critical rules (enforced by domain StateMachine):
- No backward transitions from final states (COMPLETED, FAILED, CANCELLED)
- Subtasks inherit parent task final states automatically
- Start date assigned on transition to IN_PROGRESS
- End date assigned on transition to final states

## API Specification

OpenAPI 3.0 specification should be in `api/openapi/spec.yaml` (API-First approach).

Current endpoints:
- `GET /health` - Health check with database connectivity test

Planned endpoints (per README):
- `POST /Automatizacion` - Create task
- `PUT /Automatizacion` - Update task
- `GET /Automatizacion/{uuid}` - Get task by ID
- `GET /AutomatizacionListado` - List tasks with filters/pagination
- `PUT /Subtask/{uuid}` - Update subtask
- `DELETE /Subtask/{uuid}` - Soft delete subtask

## Error Handling

API follows **RFC 7807 (Problem Details)** standard. See `docs/RFC7807.md` for full specification.

Error responses use structure:
```json
{
  "type": "https://api.example.com/problems/invalid-state-transition",
  "title": "Invalid State Transition",
  "status": 400,
  "detail": "Cannot transition from COMPLETED to PENDING"
}
```

## Database

- PostgreSQL 16 with pg_cron extension
- Soft deletes: Records marked deleted, purged after 30 days via pg_cron job
- Migrations managed with golang-migrate
- Connection pooling via pgx/v5 pgxpool

## Docker Environment

Services defined in [deployments/docker/docker-compose.yml](deployments/docker/docker-compose.yml):
- `api` service: Go application (port 8080)
- `db` service: PostgreSQL 16-alpine (port 5432)
- Health checks configured for both services
- Init SQL script: `deployments/docker/init-db.sql`

## CI/CD

GitHub Actions workflow ([.github/workflows/ci.yml](.github/workflows/ci.yml)):
- **test**: Run tests with race detector, upload coverage to Codecov
- **build**: Compile binary, upload artifact
- **docker**: Build Docker image with caching
- **lint**: Run golangci-lint (5m timeout)

Triggers on push/PR to `main` and `develop` branches.

## Development Principles

- **API-First**: OpenAPI specification before implementation
- **TDD**: Write tests first, especially for domain and use case layers
- **SOLID**: Single responsibility, dependency inversion
- **KISS**: Simple solutions over complex abstractions
- **DRY**: Avoid duplication
