# Social Network

Database diagram:
https://dbdiagram.io/d/social-network-giannis-69415c3f6167ba741474389c

## Migrations (Postgres)
- The backend uses Postgres + golang-migrate.
- Migrations are applied automatically on server startup via `MIGRATIONS_PATH`.
- For manual migration commands and best practices, see `DOCS/migrations.md`.

Required env vars (load from `backend/.env` if present):
```
DATABASE_URL=postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH=pkg/db/migrations/postgres
```

## Initial server
The initial HTTP server lives at `backend/cmd/server/main.go`. It:
- Loads optional `.env` values.
- Opens Postgres and runs migrations.
- Starts an HTTP server with a `/healthz` endpoint.

Required env var:
```
SERVER_ADDR=:8080
```
