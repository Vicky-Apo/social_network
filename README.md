# Social Network

Database diagram:
https://dbdiagram.io/d/social-network-giannis-69415c3f6167ba741474389c

## Quick Start with Docker

The easiest way to run the entire application is using Docker Compose:

```bash
# Start all services (backend, frontend, database)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5433

## Development Setup

### Backend

Required environment variables for local development (create `backend/.env`):

```bash
SERVER_ADDR=:8080
DATABASE_URL=postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH=pkg/db/migrations/postgres
UPLOAD_DIR=backend/uploads
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

Run the backend:
```bash
cd backend
go run cmd/server/main.go
```

### Frontend

Create `frontend/.env.local`:
```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

Run the frontend:
```bash
cd frontend
npm install
npm run dev
```

## Migrations (Postgres)
- The backend uses Postgres + golang-migrate.
- Migrations are applied automatically on server startup via `MIGRATIONS_PATH`.
- For manual migration commands and best practices, see `DOCS/migrations.md`.

## Docker Configuration

The project includes:
- `docker-compose.yml` - Orchestrates all services
- `backend/Dockerfile` - Multi-stage Go build
- `frontend/Dockerfile` - Next.js standalone build
- `backend/.dockerignore` - Excludes unnecessary files from backend build
- `frontend/.dockerignore` - Excludes unnecessary files from frontend build

### Environment Variables (Docker)

All configuration is managed through `docker-compose.yml`. Key variables:

**Backend:**
- `SERVER_ADDR` - Server bind address (default: 0.0.0.0:8080)
- `DATABASE_URL` - PostgreSQL connection string
- `MIGRATIONS_PATH` - Path to migration files
- `UPLOAD_DIR` - Directory for uploaded files
- `CORS_ALLOWED_ORIGINS` - Frontend URL for CORS

**Frontend:**
- `NEXT_PUBLIC_API_URL` - Backend API URL

**Database:**
- `POSTGRES_USER` - Database user
- `POSTGRES_PASSWORD` - Database password
- `POSTGRES_DB` - Database name

## Initial server
The initial HTTP server lives at `backend/cmd/server/main.go`. It:
- Loads optional `.env` values.
- Opens Postgres and runs migrations.
- Starts an HTTP server with authentication, rate limiting, and CORS enabled.
