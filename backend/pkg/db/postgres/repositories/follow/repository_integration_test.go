//go:build integration

package follow

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	domainfollow "social-network/backend/internal/domain/follow"
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

func TestFollowRequestsLifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "follow_requests", "users")

	u1 := createUser(t, db, "follow1@example.com")
	u2 := createUser(t, db, "follow2@example.com")

	repo := NewRepository(db)
	exists, err := repo.RequestExists(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("request exists: %v", err)
	}
	if exists {
		t.Fatalf("expected no request")
	}

	req, err := repo.CreateRequest(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if req.ID == 0 || req.Status != "pending" {
		t.Fatalf("unexpected request")
	}

	if _, err := repo.GetRequestByID(context.Background(), req.ID); err != nil {
		t.Fatalf("get request: %v", err)
	}

	exists, err = repo.RequestExists(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("request exists: %v", err)
	}
	if !exists {
		t.Fatalf("expected request exists")
	}

	if err := repo.UpdateRequestStatus(context.Background(), req.ID, "accepted"); err != nil {
		t.Fatalf("update status: %v", err)
	}

	pending, err := repo.ListRequestsByTarget(context.Background(), u2)
	if err != nil {
		t.Fatalf("list by target: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected no pending requests")
	}

	if err := repo.UpdateRequestStatus(context.Background(), 999, "accepted"); !errors.Is(err, domainfollow.ErrRequestNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	// add another pending request for list checks (use fresh users)
	u3 := createUser(t, db, "follow3@example.com")
	u4 := createUser(t, db, "follow4@example.com")
	if _, err := repo.CreateRequest(context.Background(), u3, u4); err != nil {
		t.Fatalf("create request: %v", err)
	}
	pending, err = repo.ListRequestsByTarget(context.Background(), u4)
	if err != nil {
		t.Fatalf("list by target: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending")
	}
	pending, err = repo.ListRequestsByRequester(context.Background(), u3)
	if err != nil {
		t.Fatalf("list by requester: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending")
	}
}

func TestFollowsNetwork(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "follows", "users")

	u1 := createUser(t, db, "net1@example.com")
	u2 := createUser(t, db, "net2@example.com")
	u3 := createUser(t, db, "net3@example.com")

	repo := NewRepository(db)
	if err := repo.CreateFollow(context.Background(), u1, u2); err != nil {
		t.Fatalf("create follow: %v", err)
	}
	if err := repo.CreateFollow(context.Background(), u3, u1); err != nil {
		t.Fatalf("create follow: %v", err)
	}

	isFollowing, err := repo.IsFollowing(context.Background(), u1, u2)
	if err != nil || !isFollowing {
		t.Fatalf("expected following")
	}

	network, err := repo.GetFollowNetwork(context.Background(), u1)
	if err != nil {
		t.Fatalf("get network: %v", err)
	}
	if len(network) != 2 {
		t.Fatalf("expected 2 in network")
	}

	if err := repo.DeleteFollow(context.Background(), u1, u2); err != nil {
		t.Fatalf("delete follow: %v", err)
	}
	isFollowing, err = repo.IsFollowing(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("check follow: %v", err)
	}
	if isFollowing {
		t.Fatalf("expected not following")
	}
}

func createUser(t *testing.T, db *sql.DB, email string) int64 {
	t.Helper()
	var userID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, is_public)
		VALUES ($1, 'hash', 'U', 'Test', '2000-01-01', true)
		RETURNING id
	`, email).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return userID
}
