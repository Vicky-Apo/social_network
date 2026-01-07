# Social Network

Database diagram:
https://dbdiagram.io/d/social-network-giannis-69415c3f6167ba741474389c

## Migrations (SQLite)
- The backend uses SQLite + golang-migrate.
- Migrations are applied automatically on server startup via `MIGRATIONS_PATH`.
- For manual migration commands and best practices, see `backend/DOCS/migrations.md`.

Required env vars (load from `backend/.env` if present):
```
DATABASE_PATH=file:./data/social.db
MIGRATIONS_PATH=pkg/db/migrations/sqlite
```

## Initial server
The initial HTTP server lives at `backend/cmd/server/main.go`. It:
- Loads optional `.env` values.
- Opens SQLite and runs migrations.
- Starts an HTTP server with a `/healthz` endpoint.

Required env var:
```
SERVER_ADDR=:8080
```


