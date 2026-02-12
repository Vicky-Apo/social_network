# Testing Guide

This project includes unit tests, handler tests, WebSocket tests, repository integration tests (Postgres), and end-to-end (E2E) integration tests.

## Quick Start

From the `backend/` folder (Go module root):

```bash
# Usecase unit tests
go test ./internal/usecase/... -v

# HTTP handler tests
go test ./internal/transport/http/handler -v

# WebSocket tests
go test ./internal/transport/websocket -v
```

## Integration Tests (Postgres)

Integration tests require a running Postgres instance and migrations path.

Set environment variables:

```bash
export DATABASE_URL="postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable"
export MIGRATIONS_PATH="./pkg/db/migrations/postgres"
```

Run repo integration tests (recommended serial to avoid deadlocks):

```bash
go test -tags=integration -p 1 -count=1 ./pkg/db/postgres/repositories/... -v
```

Note: These tests share the same database and truncate tables. Running them in
parallel can cause deadlocks or FK violations. Keep `-p 1 -count=1` when you
run them together.

This includes the media repository integration tests (media path lookups for
posts, comments, messages, and avatars).

## E2E Integration Test

The E2E flow runs a full two-user scenario across HTTP + WebSocket:
register, login, upload, create post, react, comment, follow request/accept,
profile access, group create/invite/accept/post, event create/respond,
notifications, websocket chat, and message reactions.

```bash
go test -tags=integration ./internal/transport/http -v
```

## Test Groups and Commands

Use these grouped commands depending on what you want to verify:

### 1. Unit/Usecase Tests
```bash
go test ./internal/usecase/... -v
```

### 2. Handler Tests
```bash
go test ./internal/transport/http/handler -v
```

### 3. WebSocket Tests
```bash
go test ./internal/transport/websocket -v
```

### 4. Repository Integration Tests (Postgres)
```bash
go test -tags=integration -p 1 -count=1 ./pkg/db/postgres/repositories/... -v
```

### 5. E2E Integration Test (HTTP + WebSocket)
```bash
go test -tags=integration ./internal/transport/http -v
```

## Notes
- Integration tests require Postgres + migrations.
- E2E tests reuse the integration database and will truncate tables.
