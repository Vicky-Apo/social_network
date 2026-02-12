//go:build integration

package media

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainmedia "social-network/backend/internal/domain/media"
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

func createUser(t *testing.T, db *sql.DB, email string, avatarPath *string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, avatar_path, is_public, created_at, updated_at)
		VALUES ($1, 'hash', 'Test', 'User', '1990-01-01', $2, true, $3, $3)
		RETURNING id
	`, email, avatarPath, time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func createPost(t *testing.T, db *sql.DB, authorID int64, mediaPath *string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO posts (author_id, content, media_path, visibility)
		VALUES ($1, 'hello', $2, 'public')
		RETURNING id
	`, authorID, mediaPath).Scan(&id)
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}
	return id
}

func createComment(t *testing.T, db *sql.DB, postID, authorID int64, mediaPath *string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO comments (post_id, author_id, content, media_path)
		VALUES ($1, $2, 'comment', $3)
		RETURNING id
	`, postID, authorID, mediaPath).Scan(&id)
	if err != nil {
		t.Fatalf("insert comment: %v", err)
	}
	return id
}

func createConversation(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO conversations (type)
		VALUES ('direct')
		RETURNING id
	`).Scan(&id)
	if err != nil {
		t.Fatalf("insert conversation: %v", err)
	}
	return id
}

func createMessage(t *testing.T, db *sql.DB, conversationID, senderID int64, mediaPath *string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO messages (conversation_id, sender_id, content, media_path)
		VALUES ($1, $2, 'hi', $3)
		RETURNING id
	`, conversationID, senderID, mediaPath).Scan(&id)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}
	return id
}

func TestFindByPath_Post(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "comments", "posts", "users")

	userID := createUser(t, db, "post@example.com", nil)
	mediaPath := "/uploads/post/post.jpg"
	postID := createPost(t, db, userID, &mediaPath)

	repo := NewRepository(db)
	ref, err := repo.FindByPath(context.Background(), mediaPath)
	if err != nil {
		t.Fatalf("find by path: %v", err)
	}
	if ref.Type != domainmedia.MediaTypePost || ref.PostID != postID {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestFindByPath_Comment(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "comments", "posts", "users")

	userID := createUser(t, db, "comment@example.com", nil)
	postID := createPost(t, db, userID, nil)
	mediaPath := "/uploads/comment/comment.jpg"
	commentID := createComment(t, db, postID, userID, &mediaPath)

	repo := NewRepository(db)
	ref, err := repo.FindByPath(context.Background(), mediaPath)
	if err != nil {
		t.Fatalf("find by path: %v", err)
	}
	if ref.Type != domainmedia.MediaTypeComment || ref.CommentID != commentID || ref.PostID != postID {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestFindByPath_Message(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "comments", "posts", "users")

	userID := createUser(t, db, "msg@example.com", nil)
	conversationID := createConversation(t, db)
	mediaPath := "/uploads/message/msg.jpg"
	messageID := createMessage(t, db, conversationID, userID, &mediaPath)

	repo := NewRepository(db)
	ref, err := repo.FindByPath(context.Background(), mediaPath)
	if err != nil {
		t.Fatalf("find by path: %v", err)
	}
	if ref.Type != domainmedia.MediaTypeMessage || ref.MessageID != messageID || ref.ConversationID != conversationID {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestFindByPath_Avatar(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "comments", "posts", "users")

	mediaPath := "/uploads/avatar/avatar.jpg"
	userID := createUser(t, db, "avatar@example.com", &mediaPath)

	repo := NewRepository(db)
	ref, err := repo.FindByPath(context.Background(), mediaPath)
	if err != nil {
		t.Fatalf("find by path: %v", err)
	}
	if ref.Type != domainmedia.MediaTypeAvatar || ref.UserID != userID {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestFindByPath_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "messages", "conversation_members", "conversations", "comments", "posts", "users")

	repo := NewRepository(db)
	_, err := repo.FindByPath(context.Background(), "/uploads/missing.jpg")
	if !errors.Is(err, domainmedia.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
