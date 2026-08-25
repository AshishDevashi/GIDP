# GIDP (Golang Internal Developer Portal)

Go modular monolith starter with live reload ([air](https://github.com/air-verse/air)) and type-safe SQL ([sqlc](https://sqlc.dev)).

## Structure

```
cmd/api/              entrypoint
internal/config/      env-based configuration
internal/platform/    shared infra (db, logger)
internal/store/       sqlc-generated data access code (do not edit by hand)
internal/server/      HTTP server wiring, mounts all modules
internal/modules/     one folder per business module (health, user, ...)
db/migrations/        SQL schema migrations
db/queries/           sqlc query definitions
pkg/                  shared, reusable, framework-agnostic packages
```

## Getting started

```bash
cp .env.example .env
make sqlc-generate   # generate DB access code from db/queries + db/migrations
make tidy
make migrate-up      # apply db/migrations to the database (requires goose)
make dev             # run with live reload via air
# or
make run
```

Migrations use [goose](https://github.com/pressly/goose). Tables (e.g. `users`) won't exist until `make migrate-up` has been run against the target database — run it once against localhost and again (or with `DB_URL` set) whenever you point at a different database.

## Running with Docker Compose

Starts Postgres and the app together, app connects to Postgres via the `db` service hostname.

```bash
make up      # docker compose up -d --build
make logs    # follow logs
make down    # stop and remove containers
```

Postgres is also published on `localhost:5432` (user/pass/db: `postgres`/`postgres`/`wolf_platform`) so you can run the app locally with `make dev` against the same database.

## Adding a module

1. Create `internal/modules/<name>/` with `handler.go`, `service.go`, `repository.go`.
2. Register its routes in `internal/server/server.go`.
3. Add SQL schema to `db/migrations/`, queries to `db/queries/`, then run `make sqlc-generate`.
