.PHONY: backend-run backend-migrate-up backend-migrate-down backend-migrate-version backend-migrate-goto backend-migrate-force \
	db-up db-down db-ps db-logs db-reset \
	docker-help docker-status docker-build docker-up docker-rebuild-up-app docker-down docker-restart docker-ps docker-logs docker-backend-logs docker-frontend-logs docker-db-logs docker-clean \
	docker-build-backend docker-build-frontend docker-up-backend docker-up-frontend docker-up-postgres docker-down-backend docker-down-frontend docker-down-postgres docker-reset-postgres \
	backend-test frontend-test
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DB_URL ?= postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH ?= $(BACKEND_DIR)/pkg/db/migrations/postgres
MIGRATE ?= go run -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
DOCKER_COMPOSE ?= docker/docker-compose.dev.yml

# --------------------------------------------------------------------
# Local Development (non-Docker)
# Postgres runs in Docker only. Use db-* targets to manage it.
# --------------------------------------------------------------------

### BACKEND TASKS ###
backend-run:
	(cd $(BACKEND_DIR) && go run ./cmd/server)

backend-migrate-up:
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" up

backend-migrate-down:
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" down

backend-migrate-version:
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" version

backend-migrate-goto:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required: make backend-migrate-goto VERSION=2"; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" goto $(VERSION)

backend-migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required: make backend-migrate-force VERSION=2"; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" force $(VERSION)


### DATABASE TASKS ###
db-up:
	docker compose -f $(DOCKER_COMPOSE) up -d postgres

db-down:
	docker compose -f $(DOCKER_COMPOSE) down

db-ps:
	docker compose -f $(DOCKER_COMPOSE) ps

db-logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f postgres

db-reset:
	docker compose -f $(DOCKER_COMPOSE) down -v postgres
	docker compose -f $(DOCKER_COMPOSE) up -d postgres

backend-test:
	(cd $(BACKEND_DIR) && go test ./...)

frontend-test:
	npm --prefix $(FRONTEND_DIR) test

# --------------------------------------------------------------------
# Docker Compose (build/run)
# --------------------------------------------------------------------

### DOCKER COMPOSE TASKS ###

## Docker: Build
docker-build:
	docker compose -f $(DOCKER_COMPOSE) build

## Docker: Run (All Services)
docker-up:
	docker compose -f $(DOCKER_COMPOSE) up -d

## Docker: Rebuild + Run (Backend + Frontend only)
docker-rebuild-up-app:
	docker compose -f $(DOCKER_COMPOSE) up -d --build backend frontend

## Docker: Stop/Down (All Services)
docker-down:
	docker compose -f $(DOCKER_COMPOSE) down

## Docker: Restart (All Services)
docker-restart:
	docker compose -f $(DOCKER_COMPOSE) restart

## Docker: Help/Status
docker-help:
	@echo "Top 5 Docker dev commands:"
	@echo "  make docker-build"
	@echo "  make docker-up"
	@echo "  make docker-status"
	@echo "  make docker-logs"
	@echo "  make docker-down"

docker-status:
	docker compose -f $(DOCKER_COMPOSE) ps

## Docker: Status/Logs (All Services)
docker-ps:
	docker compose -f $(DOCKER_COMPOSE) ps

docker-logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f

docker-backend-logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f backend

docker-frontend-logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f frontend

docker-db-logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f postgres

## Docker: Clean (All Services)
docker-clean:
	docker compose -f $(DOCKER_COMPOSE) down -v --remove-orphans

## Docker: Build (Per Service)
docker-build-backend:
	docker compose -f $(DOCKER_COMPOSE) build backend

docker-build-frontend:
	docker compose -f $(DOCKER_COMPOSE) build frontend

## Docker: Run (Per Service)
docker-up-postgres:
	docker compose -f $(DOCKER_COMPOSE) up -d postgres

docker-up-backend:
	docker compose -f $(DOCKER_COMPOSE) up -d backend

docker-up-frontend:
	docker compose -f $(DOCKER_COMPOSE) up -d frontend

## Docker: Stop (Per Service)
docker-down-postgres:
	docker compose -f $(DOCKER_COMPOSE) stop postgres

docker-down-backend:
	docker compose -f $(DOCKER_COMPOSE) stop backend

docker-down-frontend:
	docker compose -f $(DOCKER_COMPOSE) stop frontend

## Docker: Reset (Per Service)
docker-reset-postgres:
	docker compose -f $(DOCKER_COMPOSE) down -v postgres
	docker compose -f $(DOCKER_COMPOSE) up -d postgres
