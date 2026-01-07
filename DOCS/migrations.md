# SQLite migrations guide

This project uses SQLite with golang-migrate. Migrations must run on server startup and can also be driven manually for local work.

## Prereqs
- SQLite headers available (e.g. `sudo apt-get install libsqlite3-dev build-essential` on Debian/Ubuntu).
- Go toolchain installed.

## Environment
Set the database and migrations paths (already in `backend/.env`):
```
DATABASE_PATH=file:./data/social.db
MIGRATIONS_PATH=pkg/db/migrations/sqlite
```
You can also use a URL form that the CLI accepts:
```
DATABASE_PATH=sqlite3://$(pwd)/data/social.db
```
The server uses these values to open the DB and apply migrations automatically.

## Applying migrations on server start
Simply run the server; `backend/cmd/server/main.go` calls `sqlite.ApplyMigrations` on startup using `MIGRATIONS_PATH`.

## Using the golang-migrate CLI locally
Install the CLI with SQLite support (important: include the `sqlite3` tag):
```
cd backend
CGO_ENABLED=1 go install -tags "sqlite3" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
```
Ensure you are using that binary (check with `which migrate`), or call it directly:
```
$(go env GOPATH)/bin/migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite version
```

Common commands (run from `backend`):
```
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite up     # apply all pending
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite down   # rollback one step
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite goto N # move to version N
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite force N # set version after fixing dirty state
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite version # show version
```
If your PATH binary lacks the driver, use `go run` instead:
```
CGO_ENABLED=1 go run -tags "sqlite3" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
  -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite version
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
DB_URL=sqlite3:///tmp/test.db make migrate-up
MIGRATIONS_PATH=/abs/path/to/migrations make migrate-version
```

## Checking version via bundled helper
There is a small helper that uses the project’s drivers:
```
cd backend
go run ./cmd/migrateversion -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite
```
It prints the current version and whether the schema is dirty.

## Adding migrations (best practice)
1) Create paired files in `backend/pkg/db/migrations/sqlite`:
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
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite force <last_good_version>
```
Then re-run `up`.

## Resetting locally
For a clean slate in dev:
```
rm -f data/social.db
migrate -database "sqlite3://$(pwd)/data/social.db" -path pkg/db/migrations/sqlite up
```
or just start the server to reapply from scratch.
