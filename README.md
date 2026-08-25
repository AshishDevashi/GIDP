# wolf-platform

Go modular monolith starter with live reload ([air](https://github.com/air-verse/air)) and type-safe SQL ([sqlc](https://sqlc.dev)).

## Structure

```
cmd/api/              entrypoint
internal/config/      env-based configuration
internal/platform/    shared infra (db, logger)
internal/server/      HTTP server wiring, mounts all modules
internal/modules/     one folder per business module (health, user, ...)
  user/db/            sqlc-generated code (do not edit by hand)
db/migrations/        SQL schema migrations
db/queries/           sqlc query definitions
pkg/                  shared, reusable, framework-agnostic packages
```

## Getting started

```bash
cp .env.example .env
make sqlc-generate   # generate DB access code from db/queries + db/migrations
make tidy
make dev             # run with live reload via air
# or
make run
```

## Adding a module

1. Create `internal/modules/<name>/` with `handler.go`, `service.go`, `repository.go`.
2. Register its routes in `internal/server/server.go`.
3. Add SQL schema to `db/migrations/`, queries to `db/queries/`, then run `make sqlc-generate`.
