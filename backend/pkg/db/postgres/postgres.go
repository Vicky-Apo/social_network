package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file" // enable file:// migrations
	_ "github.com/lib/pq"                               // postgres driver

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// Config wraps tunables for connecting to Postgres.
type Config struct {
	URL string
}

// Open connects to Postgres and verifies the connection.
func Open(cfg Config) (*sql.DB, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("postgres URL is required")
	}

	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// ApplyMigrations runs all up migrations from the given file:// path.
// Example sourceURL: file://backend/pkg/db/migrations/postgres
func ApplyMigrations(db *sql.DB, sourceURL string) error {
	if sourceURL == "" {
		return fmt.Errorf("sourceURL is required")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// WithDefaults builds a Config with standard defaults.
func WithDefaults(url string) Config {
	return Config{URL: url}
}
