//go:build integration

package chat

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

func TestCreateMessage_ReactionsPersist(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "message_reactions", "messages", "conversation_members", "conversations", "users")

	// users
	var u1, u2 int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, is_public)
		VALUES ('u1@example.com', 'hash', 'U', 'One', '2000-01-01', true)
		RETURNING id
	`).Scan(&u1)
	if err != nil {
		t.Fatalf("insert user1: %v", err)
	}
	err = db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, is_public)
		VALUES ('u2@example.com', 'hash', 'U', 'Two', '2000-01-01', true)
		RETURNING id
	`).Scan(&u2)
	if err != nil {
		t.Fatalf("insert user2: %v", err)
	}

	repo := NewRepository(db)
	conv, err := repo.GetOrCreateDirectConversation(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	msg, err := repo.CreateMessage(context.Background(), conv.ID, u1, strPtr("hello"), nil)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	if err := repo.AddMessageReaction(context.Background(), msg.ID, u2, "😀"); err != nil {
		t.Fatalf("add reaction: %v", err)
	}

	items, err := repo.ListMessageReactions(context.Background(), msg.ID)
	if err != nil {
		t.Fatalf("list reactions: %v", err)
	}
	if len(items) != 1 || items[0].Emoji != "😀" {
		t.Fatalf("unexpected reactions")
	}
}

func TestGetOrCreateDirectConversation_Idempotent(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "conversation_members", "conversations", "users")

	u1 := createUser(t, db, "direct1@example.com")
	u2 := createUser(t, db, "direct2@example.com")

	repo := NewRepository(db)
	conv1, err := repo.GetOrCreateDirectConversation(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	conv2, err := repo.GetOrCreateDirectConversation(context.Background(), u2, u1)
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if conv1.ID != conv2.ID {
		t.Fatalf("expected same conversation id")
	}

	members, err := repo.GetConversationMembers(context.Background(), conv1.ID)
	if err != nil {
		t.Fatalf("get members: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members")
	}
}

func TestMessages_PaginationAndOrder(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "users")

	u1 := createUser(t, db, "msg1@example.com")
	u2 := createUser(t, db, "msg2@example.com")
	repo := NewRepository(db)
	conv, err := repo.GetOrCreateDirectConversation(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	msg1, err := repo.CreateMessage(context.Background(), conv.ID, u1, strPtr("first"), nil)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	msg2, err := repo.CreateMessage(context.Background(), conv.ID, u2, strPtr("second"), nil)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	items, err := repo.GetMessagesByConversation(context.Background(), conv.ID, 1, 0)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(items) != 1 || items[0].ID != msg2.ID {
		t.Fatalf("expected latest message first")
	}

	items, err = repo.GetMessagesByConversation(context.Background(), conv.ID, 2, 0)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(items) != 2 || items[1].ID != msg1.ID {
		t.Fatalf("expected message order")
	}
}

func TestUnreadCountsAndMarkAsRead(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "users")

	u1 := createUser(t, db, "unread1@example.com")
	u2 := createUser(t, db, "unread2@example.com")
	repo := NewRepository(db)
	conv, err := repo.GetOrCreateDirectConversation(context.Background(), u1, u2)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	if _, err := repo.CreateMessage(context.Background(), conv.ID, u1, strPtr("m1"), nil); err != nil {
		t.Fatalf("create message: %v", err)
	}
	if _, err := repo.CreateMessage(context.Background(), conv.ID, u2, strPtr("m2"), nil); err != nil {
		t.Fatalf("create message: %v", err)
	}

	unread, err := repo.GetUnreadCount(context.Background(), conv.ID, u2)
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if unread != 2 {
		t.Fatalf("expected 2 unread, got %d", unread)
	}

	mapping, err := repo.GetUnreadConversations(context.Background(), u2)
	if err != nil {
		t.Fatalf("get unread conversations: %v", err)
	}
	if mapping[conv.ID] != 2 {
		t.Fatalf("expected unread map to include conversation")
	}

	if err := repo.MarkAsRead(context.Background(), conv.ID, u2); err != nil {
		t.Fatalf("mark as read: %v", err)
	}
	unread, err = repo.GetUnreadCount(context.Background(), conv.ID, u2)
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if unread != 0 {
		t.Fatalf("expected 0 unread, got %d", unread)
	}
}

func TestGroupConversationLookup(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "group_conversations", "conversations", "groups", "users")

	creatorID := createUser(t, db, "groupchat@example.com")
	groupID := createGroup(t, db, creatorID)

	repo := NewRepository(db)
	convID, err := repo.GetGroupConversationID(context.Background(), groupID)
	if err != nil {
		t.Fatalf("get group conversation: %v", err)
	}

	gotGroupID, err := repo.GetGroupIDByConversationID(context.Background(), convID)
	if err != nil {
		t.Fatalf("get group id: %v", err)
	}
	if gotGroupID == nil || *gotGroupID != groupID {
		t.Fatalf("expected group id")
	}
}

func strPtr(s string) *string { return &s }

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

func createGroup(t *testing.T, db *sql.DB, creatorID int64) int64 {
	t.Helper()
	var groupID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO groups (creator_id, title, description)
		VALUES ($1, 'Test Group', 'Desc')
		RETURNING id
	`, creatorID).Scan(&groupID)
	if err != nil {
		t.Fatalf("insert group: %v", err)
	}
	return groupID
}
