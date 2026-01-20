# Postgres migrations guide

This project uses Postgres with golang-migrate. Migrations must run on server startup and can also be driven manually for local work.

## Prereqs
- Postgres installed locally and accepting connections.
- Go toolchain installed.

## Database setup (local)
Create the role and database once:
```
psql -U postgres -c "CREATE ROLE \"social-network-role\" WITH LOGIN PASSWORD '123456';"
psql -U postgres -c "CREATE DATABASE \"social-network-db\" OWNER \"social-network-role\";"
```

## Environment
Set the database and migrations paths (already in `backend/.env`):
```
DATABASE_URL=postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH=pkg/db/migrations/postgres
```
The server uses these values to open the DB and apply migrations automatically.

## Applying migrations on server start
Simply run the server; `backend/cmd/server/main.go` calls `postgres.ApplyMigrations` on startup using `MIGRATIONS_PATH`.

## Using the golang-migrate CLI locally
Install the CLI:
```
cd backend
go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
```
Ensure you are using that binary (check with `which migrate`), or call it directly:
```
$(go env GOPATH)/bin/migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres version
```

Common commands (run from `backend`):
```
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres up     # apply all pending
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres down   # rollback one step
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres goto N # move to version N
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres force N # set version after fixing dirty state
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres version # show version
```

## Short commands (Makefile)
From `backend/`, the Makefile wraps the common commands with sensible defaults:
```
make migrate-up          # up
make migrate-down        # down
make migrate-version     # show version
make migrate-goto VERSION=2
make migrate-force VERSION=2
```
Override paths/DB if needed:
```
DB_URL=postgres://user:pass@localhost:5433/db?sslmode=disable make migrate-up
MIGRATIONS_PATH=/abs/path/to/migrations make migrate-version
```

## Checking version via bundled helper
There is a small helper that uses the project’s drivers:
```
cd backend
go run ./cmd/migrateversion -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres
```
It prints the current version and whether the schema is dirty.

## Adding migrations (best practice)
1) Create paired files in `backend/pkg/db/migrations/postgres`:
```
000003_create_followers_table.up.sql
000003_create_followers_table.down.sql
```
2) Put only the forward change in `.up.sql` and the exact rollback in `.down.sql`.
3) Keep increments sequential and zero-padded.
4) Favor narrow, reversible changes; include indexes and foreign keys; keep `updated_at` triggers if you need them.
5) After adding, run `migrate ... up` (or start the server) and verify with `version`.

## Dirty state recovery
If a migration fails, the database can become dirty. Fix the SQL or DB manually, then run:
```
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres force <last_good_version>
```
Then re-run `up`.

## Resetting locally
For a clean slate in dev:
```
dropdb --if-exists social-network-db
createdb -O social-network-role social-network-db
migrate -database "$(DATABASE_URL)" -path pkg/db/migrations/postgres up
```
or just start the server to reapply from scratch.
