# GIDP — Copilot Instructions

## What this project is

GIDP (Golang Internal Developer Portal) is a **Go modular monolith** — a single deployable binary organized into
self-contained business modules, backed by Postgres. It uses:

- **gin** for HTTP routing (`github.com/gin-gonic/gin`)
- **sqlc** for type-safe, generated database access (never hand-write query code)
- **goose** for SQL migrations
- **air** for live reload during local development
- **JWT** (`golang-jwt/jwt/v5`) for stateless authentication, **bcrypt** for password hashing
- **pgx/v5** (`pgxpool`) as the Postgres driver

All API routes are versioned under `/api/v1` (see [internal/server/server.go](../internal/server/server.go)).
`/health` is the only unversioned route (infra liveness check).

## Project structure

```
cmd/api/              entrypoint (main.go) — wires config, logger, db, server
internal/config/      env-based configuration (Config struct + Load())
internal/platform/    shared, framework-level infra: db connection, logger,
                       password hashing, JWT token helpers, uuid helpers
internal/store/       sqlc-generated code (package `store`). DO NOT hand-edit;
                       regenerate with `make sqlc-generate`
internal/server/       gin engine setup, route groups, module registration
internal/modules/      one folder per business module (auth, user, health, ...)
db/migrations/         goose SQL migrations (numbered, sequential)
db/queries/            sqlc query definitions (.sql files, one per domain)
pkg/                   small, reusable, framework-agnostic packages (e.g. response)
```

## Module conventions (follow exactly — consistency matters)

Every module under `internal/modules/<name>/` follows the same shape:

- `handler.go` — HTTP layer. Defines `Module` struct, `NewModule(pool *pgxpool.Pool, ...)` constructor that
  builds its own `Repository` and `Service` internally, `RegisterRoutes(rg *gin.RouterGroup)`, and gin handler
  methods. Handlers only: bind/validate input, call the service, translate errors to HTTP status codes, respond
  via `pkg/response`. No business logic or SQL in handlers.
- `service.go` — business logic. Takes a `*Repository` (and any other deps like JWT secret/TTL). Returns
  password-free / API-safe response DTOs (see `user.Response`, `auth.UserResponse`) — **never return sqlc models
  with sensitive fields (e.g. `password_hash`) directly from a handler.**
- `repository.go` — thin wrapper around `*store.Queries` methods needed by this module. No SQL here, only calls
  into the generated `internal/store` package.
- `dto.go` (optional) — request/response structs with gin `binding` tags for validation.

Constructor pattern to copy for every new module:

```go
func NewModule(pool *pgxpool.Pool, ...extraDeps) *Module {
    repo := NewRepository(pool)
    svc := NewService(repo, ...extraDeps)
    return &Module{service: svc}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
    rg.GET("/things", m.list)
}
```

### Adding a new module — checklist

1. Add schema to `db/migrations/<next_number>_<name>.sql` (goose `-- +goose Up` / `-- +goose Down`).
2. Add queries to `db/queries/<name>.sql` (sqlc `-- name: QueryName :one|:many|:exec`).
3. Run `make sqlc-generate` to regenerate `internal/store`.
4. Run `make migrate-up` to apply the new migration to the local/dev database.
5. Create `internal/modules/<name>/{repository,service,handler}.go` following the pattern above.
6. Register the module in `internal/server/server.go` under the `/api/v1` group (or `protected` group if it
   requires auth — see below).
7. Update `.env.example` / `docker-compose.yml` only if new env vars are introduced.

### Protecting routes with JWT

`auth.Module` exposes `RequireAuth() gin.HandlerFunc`. To protect a group of routes:

```go
protected := v1.Group("")
protected.Use(authModule.RequireAuth())
someModule.NewModule(s.db).RegisterRoutes(protected)
```

Inside a protected handler, read the authenticated user via
`c.GetString(auth.ContextUserIDKey)` / `ContextUsernameKey` / `ContextRoleIDKey`.

## Go / backend best practices to follow

- Keep handlers thin; put logic in services and data access in repositories.
- Always accept `context.Context` as the first argument through service/repository layers and pass
  `c.Request.Context()` from handlers — never `context.Background()` in request paths.
- Never log or return secrets: password hashes, JWT secrets, raw DB errors that may leak schema details to
  clients. Map internal errors to generic messages via `response.Error`.
- Validate all external input with gin `binding` tags; don't trust client-supplied IDs/roles.
- Use `pgtype.UUID` for Postgres `uuid` columns (sqlc default) and the `internal/platform/uuidutil` helpers to
  convert to/from string — don't add ad-hoc UUID handling per module.
- Prefer small, focused packages under `internal/platform` for cross-cutting concerns (hashing, tokens, db,
  logging) instead of duplicating logic inside modules.
- Return errors, don't panic; only `gin.Recovery()` middleware should be the last line of defense.
- Keep sqlc-generated code untouched — if a query needs to change, edit the `.sql` file and regenerate.
- Follow existing naming: package names singular and lowercase (`user`, `auth`), no stutter
  (`user.Repository`, not `user.UserRepository`).

## What NOT to do

- Do not run builds, tests, or start containers/services unless explicitly asked to. Prioritize making the
  requested code changes correctly over verifying them — only verify when the user asks you to.
- Do not create new markdown docs for every change; only update existing docs (README, this file) when
  relevant.
- Do not introduce a new web framework, ORM, or migration tool — stick to gin, sqlc, and goose.
