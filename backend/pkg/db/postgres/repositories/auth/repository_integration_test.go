//go:build integration

package auth

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainauth "social-network/backend/internal/domain/auth"
	"social-network/backend/pkg/db/postgres"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	migrations := os.Getenv("MIGRATIONS_PATH")
	if migrations == "" {
		t.Skip("MIGRATIONS_PATH not set")
	}
	abs, err := filepath.Abs(migrations)
	if err != nil {
		t.Fatalf("migrations path: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		root := findModuleRoot(t)
		abs = filepath.Join(root, migrations)
	}

	db, err := postgres.Open(postgres.WithDefaults(url))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sourceURL := "file://" + abs
	if err := postgres.ApplyMigrations(db, sourceURL); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func cleanup(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if _, err := db.ExecContext(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "sessions", "users")

	repo := NewRepository(db)
	user := domainauth.User{
		Email:        "auth1@example.com",
		PasswordHash: "hash",
		FirstName:    "A",
		LastName:     "One",
		DateOfBirth:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		IsPublic:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if _, err := repo.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := repo.CreateUser(context.Background(), user); !errors.Is(err, domainauth.ErrEmailAlreadyExists) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestGetUserByEmailAndID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "sessions", "users")

	repo := NewRepository(db)
	user := domainauth.User{
		Email:        "auth2@example.com",
		PasswordHash: "hash",
		FirstName:    "B",
		LastName:     "Two",
		DateOfBirth:  time.Date(1999, 2, 2, 0, 0, 0, 0, time.UTC),
		IsPublic:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	id, err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := repo.GetUserByEmail(context.Background(), user.Email); err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if _, err := repo.GetUserByID(context.Background(), id); err != nil {
		t.Fatalf("get by id: %v", err)
	}

	if _, err := repo.GetUserByEmail(context.Background(), "missing@example.com"); !errors.Is(err, domainauth.ErrUserNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if _, err := repo.GetUserByID(context.Background(), 999); !errors.Is(err, domainauth.ErrUserNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "sessions", "users")

	repo := NewRepository(db)
	user := domainauth.User{
		Email:        "auth3@example.com",
		PasswordHash: "hash",
		FirstName:    "C",
		LastName:     "Three",
		DateOfBirth:  time.Date(2001, 3, 3, 0, 0, 0, 0, time.UTC),
		IsPublic:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userID, err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	sess := domainauth.Session{
		UserID:       userID,
		SessionToken: "token-1",
		ExpiresAt:    time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}
	if _, err := repo.CreateSession(context.Background(), sess); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := repo.GetSessionByToken(context.Background(), "token-1"); err != nil {
		t.Fatalf("get session: %v", err)
	}

	if err := repo.DeleteSession(context.Background(), "token-1"); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := repo.GetSessionByToken(context.Background(), "token-1"); !errors.Is(err, domainauth.ErrSessionNotFound) {
		t.Fatalf("expected session not found, got %v", err)
	}

	expired := domainauth.Session{
		UserID:       userID,
		SessionToken: "token-expired",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		CreatedAt:    time.Now(),
	}
	if _, err := repo.CreateSession(context.Background(), expired); err != nil {
		t.Fatalf("create expired session: %v", err)
	}
	if _, err := repo.GetSessionByToken(context.Background(), "token-expired"); !errors.Is(err, domainauth.ErrSessionNotFound) {
		t.Fatalf("expected expired session not found, got %v", err)
	}
}

func TestDeleteUserSessions(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "sessions", "users")

	repo := NewRepository(db)
	user := domainauth.User{
		Email:        "auth4@example.com",
		PasswordHash: "hash",
		FirstName:    "D",
		LastName:     "Four",
		DateOfBirth:  time.Date(2002, 4, 4, 0, 0, 0, 0, time.UTC),
		IsPublic:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userID, err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := repo.CreateSession(context.Background(), domainauth.Session{
		UserID:       userID,
		SessionToken: "token-a",
		ExpiresAt:    time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := repo.CreateSession(context.Background(), domainauth.Session{
		UserID:       userID,
		SessionToken: "token-b",
		ExpiresAt:    time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := repo.DeleteUserSessions(context.Background(), userID); err != nil {
		t.Fatalf("delete user sessions: %v", err)
	}
	if _, err := repo.GetSessionByToken(context.Background(), "token-a"); !errors.Is(err, domainauth.ErrSessionNotFound) {
		t.Fatalf("expected session not found, got %v", err)
	}
	if _, err := repo.GetSessionByToken(context.Background(), "token-b"); !errors.Is(err, domainauth.ErrSessionNotFound) {
		t.Fatalf("expected session not found, got %v", err)
	}
}
