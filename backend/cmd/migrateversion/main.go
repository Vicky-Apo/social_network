package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3" // sqlite driver
	_ "github.com/golang-migrate/migrate/v4/source/file"      // file:// migrations
	_ "github.com/mattn/go-sqlite3"                           // sqlite driver for database/sql
)

// A tiny helper to print the current migration version using the bundled drivers.
func main() {
	dbURL := flag.String("database", os.Getenv("DATABASE_PATH"), "database URL (e.g. sqlite3:///path/to/db or file:./data/social.db)")
	migrationsPath := flag.String("path", os.Getenv("MIGRATIONS_PATH"), "migrations directory (e.g. pkg/db/migrations/sqlite)")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("database URL is required (flag -database or env DATABASE_PATH)")
	}
	if *migrationsPath == "" {
		log.Fatal("migrations path is required (flag -path or env MIGRATIONS_PATH)")
	}

	sourceURL := *migrationsPath
	if !strings.HasPrefix(sourceURL, "file://") {
		abs, err := filepath.Abs(sourceURL)
		if err != nil {
			log.Fatalf("resolve migrations path: %v", err)
		}
		sourceURL = "file://" + abs
	}

	m, err := migrate.New(sourceURL, *dbURL)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}

	version, dirty, err := m.Version()
	switch {
	case errors.Is(err, migrate.ErrNilVersion):
		fmt.Println("version: 0 (no migrations applied)")
	case err != nil:
		log.Fatalf("get version: %v", err)
	default:
		fmt.Printf("version: %d dirty: %t\n", version, dirty)
	}
}
