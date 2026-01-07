package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/source/file" // enable file:// migrations
	_ "github.com/mattn/go-sqlite3"                      // sqlite driver

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
)

// Config wraps tunables for connecting to sqlite.
type Config struct {
	Path string
	// Pragmas allows overriding defaults (e.g., foreign keys, journal mode).
	Pragmas map[string]string
}

// Open connects to the sqlite database and applies reasonable defaults.
func Open(cfg Config) (*sql.DB, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	dsn := normalizeDSN(cfg.Path)
	if len(cfg.Pragmas) > 0 {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		for k, v := range cfg.Pragmas {
			dsn += separator + fmt.Sprintf("%s=%s", k, v)
			separator = "&"
		}
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Fail fast on connection errors.
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	// Enforce foreign keys by default. Cause sqlite disables them by default.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return db, nil
}

// ApplyMigrations runs all up migrations from the given file:// path.
// Example sourceURL: file://backend/pkg/db/migrations/sqlite
func ApplyMigrations(db *sql.DB, sourceURL string) error {
	if sourceURL == "" {
		return fmt.Errorf("sourceURL is required")
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("create sqlite3 migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}

	// Apply up migrations; ErrNoChange means we are already current.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// WithDefaults builds a Config with pragmatic defaults.
func WithDefaults(path string) Config {
	return Config{
		Path: path,
		Pragmas: map[string]string{
			"_busy_timeout": "5000",
			"_foreign_keys": "on",
		},
	}
}

// normalizeDSN allows using either a plain path, a file: URI, or a sqlite3:// URL
// (useful for the migrate CLI) by stripping the sqlite3:// scheme for the driver.
func normalizeDSN(raw string) string {
	if strings.HasPrefix(raw, "sqlite3://") {
		return strings.TrimPrefix(raw, "sqlite3://")
	}
	return raw
}
