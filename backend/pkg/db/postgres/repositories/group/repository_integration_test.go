//go:build integration

package group

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestCreateGroup_CreatesConversationAndMember(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "message_reactions", "messages", "conversation_members", "group_conversations", "conversations", "group_members", "group_invitations", "group_join_requests", "groups", "users")

	var gcCount int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM group_conversations").Scan(&gcCount); err != nil {
		t.Fatalf("count group_conversations: %v", err)
	}
	if gcCount != 0 {
		t.Fatalf("expected group_conversations empty, got %d", gcCount)
	}

	// create user
	var userID int64
	email := "u1-" + time.Now().Format("20060102150405") + "@example.com"
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, is_public)
		VALUES ($1, 'hash', 'U', 'One', '2000-01-01', true)
		RETURNING id
	`, email).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	repo := NewRepository(db)
	g, err := repo.Create(context.Background(), userID, "Test Group", nil)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	// verify group conversation exists
	var convID int64
	err = db.QueryRowContext(context.Background(), `
		SELECT conversation_id FROM group_conversations WHERE group_id = $1
	`, g.ID).Scan(&convID)
	if err != nil {
		t.Fatalf("missing group_conversation: %v", err)
	}

	// verify creator is member of conversation
	var exists int
	err = db.QueryRowContext(context.Background(), `
		SELECT 1 FROM conversation_members WHERE conversation_id = $1 AND user_id = $2
	`, convID, userID).Scan(&exists)
	if err != nil {
		t.Fatalf("creator not in conversation: %v", err)
	}
}
