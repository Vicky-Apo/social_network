# Local Postgres with Docker

This repo includes a Postgres service for development in `docker-compose.yml`.

## Start the database
From the repo root:
```
docker compose up -d
```
This creates the role, database, and password automatically on first run.

## Run migrations
```
make -C backend migrate-up
```
or start the server, which applies migrations automatically:
```
go run ./backend/cmd/server
```

## Connection string
```
postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable

then here in cli you can type psql commands 
```

## Reset the database (optional)
If you need a clean DB, remove the volume:
```
docker compose down -v
docker compose up -d
make -C backend migrate-up
```
