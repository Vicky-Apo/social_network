# Social Network

A full-stack social platform built with Go and Next.js. Users can create profiles, post content, follow each other, and interact — backed by PostgreSQL with versioned migrations and deployed via Docker Compose.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Next.js (TypeScript), Tailwind CSS |
| Backend | Go — layered architecture (transport / usecase / domain) |
| Database | PostgreSQL + golang-migrate |
| Auth | Session-based authentication, CORS, rate limiting |
| DevOps | Docker Compose, multi-stage builds, non-root containers |

---

## Run

```bash
docker-compose up -d

# Frontend:    http://localhost:3000
# Backend API: http://localhost:8080
# PostgreSQL:  localhost:5433
```

---

## Local Development

**Backend** — create `backend/.env`:

```env
SERVER_ADDR=:8080
DATABASE_URL=postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH=pkg/db/migrations/postgres
UPLOAD_DIR=backend/uploads
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

```bash
cd backend && go run cmd/server/main.go
```

**Frontend** — create `frontend/.env.local`:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

```bash
cd frontend && npm install && npm run dev
```

---

## Architecture

```
social_network/
├── backend/
│   ├── cmd/server/        # entry point
│   └── internal/
│       ├── transport/     # HTTP handlers, routing
│       ├── usecase/       # business logic
│       ├── domain/        # entities and interfaces
│       └── app/           # wiring and startup
│   └── pkg/
│       ├── db/migrations/ # versioned SQL migrations
│       ├── logger/
│       └── utils/
├── frontend/              # Next.js app
└── docker-compose.yml
```

The backend follows clean layered architecture: HTTP handlers depend on use-case interfaces, use cases depend on domain interfaces — no layer imports the one above it. Migrations run automatically on startup via `golang-migrate`.

---

## License

MIT
