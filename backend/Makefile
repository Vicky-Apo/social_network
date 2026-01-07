# Defaults for migration commands. Override as needed, e.g.:
# DB_URL=sqlite3:///tmp/test.db make migrate-up
DB_URL ?= sqlite3://$(PWD)/data/social.db
MIGRATIONS_PATH ?= pkg/db/migrations/sqlite
MIGRATE ?= CGO_ENABLED=1 go run -tags "sqlite3" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

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
