# Defaults for migration commands. Override as needed, e.g.:
# DB_URL=postgres://user:pass@localhost:5433/db?sslmode=disable make migrate-up
DB_URL ?= postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
MIGRATIONS_PATH ?= pkg/db/migrations/postgres
MIGRATE ?= go run -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

migrate-up:
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" up

migrate-down:
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" down

migrate-version:
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" version

migrate-goto:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required: make migrate-goto VERSION=2"; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" goto $(VERSION)

migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required: make migrate-force VERSION=2"; exit 1; fi
	$(MIGRATE) -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" force $(VERSION)

run-server:
	go run ./cmd/server

db-up:
	docker compose -f ../docker-compose.yml up -d

db-down:
	docker compose -f ../docker-compose.yml down

db-ps:
	docker compose -f ../docker-compose.yml ps

db-logs:
	docker compose -f ../docker-compose.yml logs -f postgres

db-reset:
	docker compose -f ../docker-compose.yml down -v
	docker compose -f ../docker-compose.yml up -d
