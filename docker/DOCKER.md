# Docker Setup

This project ships three containers: Postgres, Backend (Go), and Frontend (Next.js).
Use the Makefile targets below to get a clean, repeatable setup.

## Quick Start (Clean)

1. Reset database only:
```
make db-reset
```

2. Build and start backend:
```
make docker-build-backend
make docker-up-backend
```

3. Build and start frontend:
```
make docker-build-frontend
make docker-up-frontend
```

## Two Commands (All Services)

```
make docker-build
make docker-up
```

## Day-to-Day Commands

Start individual services:
```
make docker-up-postgres
make docker-up-backend
make docker-up-frontend
```

Status:
```
make docker-status
```

Stop individual services:
```
make docker-down-postgres
make docker-down-backend
make docker-down-frontend
```

Logs:
```
make docker-db-logs
make docker-backend-logs
make docker-frontend-logs
```

## Clean Everything

Remove containers + volumes:
```
make docker-clean
```


## Services and Ports

- Postgres: `localhost:5433`
- Backend: `http://localhost:8080`
- Frontend: `http://localhost:3000`

## Network

All services join the same Docker network:
`social-network`

This is created automatically by Docker Compose from `docker/docker-compose.dev.yml`.
