package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"social-network/backend/pkg/db/sqlite"
	"social-network/backend/pkg/utils"
)

const envFileName = ".env"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := utils.LoadDotEnv(envFileName); err != nil {
		log.Printf("warning: could not load %s: %v", envFileName, err)
	}

	dbPath := utils.GetString("DATABASE_PATH", "")
	if dbPath == "" {
		log.Fatalf("DATABASE_PATH is required (set in .env or environment)")
	}
	if err := ensureDir(dbPath); err != nil {
		log.Fatalf("ensure data dir: %v", err)
	}

	db, err := sqlite.Open(sqlite.WithDefaults(dbPath))
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if maxOpen := utils.GetInt("MAX_OPEN_CONNS", 0); maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
	if maxIdle := utils.GetInt("MAX_IDLE_CONNS", 0); maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	}

	migrationsDir := utils.GetString("MIGRATIONS_PATH", "")
	if migrationsDir == "" {
		log.Fatalf("MIGRATIONS_PATH is required (set in .env or environment)")
	}
	migrationsPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		log.Fatalf("resolve migrations path: %v", err)
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	if err := sqlite.ApplyMigrations(db, sourceURL); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	log.Println("server boot completed, starting HTTP listener")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:              requiredString("SERVER_ADDR"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func ensureDir(dsn string) error {
	// Extract path portion from a sqlite file:... dsn if present.
	path := dsn
	switch {
	case strings.HasPrefix(dsn, "file:"):
		path = dsn[len("file:"):]
	case strings.HasPrefix(dsn, "sqlite3://"):
		path = dsn[len("sqlite3://"):]
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func requiredString(key string) string {
	val := utils.GetString(key, "")
	if val == "" {
		log.Fatalf("%s is required (set in .env or environment)", key)
	}
	return val
}
