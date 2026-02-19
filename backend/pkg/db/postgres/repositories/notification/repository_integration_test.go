//go:build integration

package notification

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainnotification "social-network/backend/internal/domain/notification"
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

func TestCreateAndUnreadCount(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "notifications", "users")

	userID := createUser(t, db, "n1@example.com")

	repo := NewRepository(db)
	n, err := repo.Create(context.Background(), domainnotification.Notification{
		UserID:     userID,
		Type:       "follow_request",
		EntityType: "follow_request",
		EntityID:   1,
		Metadata:   []byte(`{}`),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	if n.ID == 0 {
		t.Fatalf("expected notification id")
	}

	count, err := repo.UnreadCount(context.Background(), userID)
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 unread, got %d", count)
	}
}

func TestListUnreadOnlyAndActorID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "notifications", "users")

	userID := createUser(t, db, "n2@example.com")
	actorID := createUser(t, db, "actor@example.com")

	repo := NewRepository(db)
	_, err := repo.Create(context.Background(), domainnotification.Notification{
		UserID:     userID,
		ActorID:    &actorID,
		Type:       "group_invitation",
		EntityType: "group",
		EntityID:   10,
		Metadata:   []byte(`{}`),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	n2, err := repo.Create(context.Background(), domainnotification.Notification{
		UserID:     userID,
		Type:       "follow_request",
		EntityType: "follow_request",
		EntityID:   11,
		Metadata:   []byte(`{}`),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if err := repo.MarkRead(context.Background(), userID, n2.ID, time.Now()); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	items, err := repo.ListByUser(context.Background(), userID, 10, 0, true)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 unread, got %d", len(items))
	}
	if items[0].ActorID == nil || *items[0].ActorID != actorID {
		t.Fatalf("expected actor id to be set")
	}
}

func TestMarkRead_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "notifications", "users")

	userID := createUser(t, db, "n3@example.com")
	repo := NewRepository(db)

	err := repo.MarkRead(context.Background(), userID, 999, time.Now())
	if !errors.Is(err, domainnotification.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestMarkAllRead_CountAndReadAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "notifications", "users")

	userID := createUser(t, db, "n4@example.com")
	repo := NewRepository(db)

	n1, err := repo.Create(context.Background(), domainnotification.Notification{
		UserID:     userID,
		Type:       "follow_request",
		EntityType: "follow_request",
		EntityID:   21,
		Metadata:   []byte(`{}`),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	if err := repo.MarkRead(context.Background(), userID, n1.ID, time.Now()); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	_, err = repo.Create(context.Background(), domainnotification.Notification{
		UserID:     userID,
		Type:       "group_join_request",
		EntityType: "group",
		EntityID:   22,
		Metadata:   []byte(`{}`),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	count, err := repo.MarkAllRead(context.Background(), userID, time.Now())
	if err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 updated, got %d", count)
	}

	var unread int64
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false",
		userID,
	).Scan(&unread); err != nil {
		t.Fatalf("count unread: %v", err)
	}
	if unread != 0 {
		t.Fatalf("expected 0 unread, got %d", unread)
	}
}

func createUser(t *testing.T, db *sql.DB, email string) int64 {
	t.Helper()
	var userID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, is_public)
		VALUES ($1, 'hash', 'N', 'One', '2000-01-01', true)
		RETURNING id
	`, email).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return userID
}
