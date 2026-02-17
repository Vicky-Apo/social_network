# Run Frontend and Backend

This guide is for contributors or external reviewers who want to run the project locally.

## Prerequisites

- `git`
- `go` (1.22+ recommended)
- `node` (20+ recommended) and `npm`
- `docker` + `docker compose`

## 1) Clone and install

```bash
git clone <your-repo-url>
cd social-network
cd frontend && npm install && cd ..
```

## 2) Start database (Postgres)

From project root:

```bash
make -C backend db-up
```

This starts Postgres from `docker-compose.yml` on port `5433`.

## 3) Configure backend environment

Create `backend/.env` (local only, do not commit) with at least:

```env
SERVER_ADDR=:8080
DATABASE_URL=postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH=pkg/db/migrations/postgres
```

Optional settings have defaults in `backend/internal/config/config.go`.

## 4) Run backend API

From project root:

```bash
make -C backend run-server
```

Expected: backend starts on `http://localhost:8080`.

Health check:

```bash
curl http://localhost:8080/healthz
```

## 5) Run frontend app

In a new terminal:

```bash
cd frontend
node ./node_modules/next/dist/bin/next dev
```

Open:

- `http://localhost:3000`

Notes:

- The frontend uses `NEXT_PUBLIC_API_BASE_URL` if provided.
- If unset, it defaults to `http://localhost:8080`.

Example (optional) `frontend/.env.local`:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

## 6) Stop services

- Stop frontend/backend with `Ctrl + C`
- Stop database:

```bash
make -C backend db-down
```

## Troubleshooting

- **`next: Permission denied`**
  - Use: `node ./node_modules/next/dist/bin/next dev`
- **Port 3000 busy**
  - Next.js may auto-switch to 3001; use the printed URL.
- **Backend says `SERVER_ADDR` / `DATABASE_URL` missing**
  - Verify `backend/.env` exists and contains required keys.
- **No DB connection**
  - Ensure Postgres container is up: `make -C backend db-ps`

## Security and private data

- Never commit `backend/.env` or `frontend/.env.local`.
- Never commit API keys, tokens, secrets, or credentials.
- Before pushing, review with:

```bash
git status
git diff --staged
```
