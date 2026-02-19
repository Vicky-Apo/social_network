//go:build integration

package comment

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	domaincomment "social-network/backend/internal/domain/comment"
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

func TestCreateAndGetByPostID_WithReactions(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "comment_reactions", "comments", "posts", "users")

	userID := createUser(t, db, "comment1@example.com")
	userID2 := createUser(t, db, "comment2@example.com")
	postID := createPost(t, db, userID)

	repo := NewRepository(db)
	comment, err := repo.Create(context.Background(), domaincomment.Comment{
		PostID:   postID,
		AuthorID: userID,
		Content:  "hello",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	if err := addCommentReaction(db, comment.ID, userID, "like"); err != nil {
		t.Fatalf("add reaction: %v", err)
	}
	if err := addCommentReaction(db, comment.ID, userID2, "dislike"); err != nil {
		t.Fatalf("add reaction: %v", err)
	}

	items, err := repo.GetByPostID(context.Background(), postID, 10, 0)
	if err != nil {
		t.Fatalf("get by post: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 comment")
	}
	if items[0].LikeCount != 1 || items[0].DislikeCount != 1 {
		t.Fatalf("unexpected reaction counts")
	}

	got, err := repo.GetByID(context.Background(), comment.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ID != comment.ID {
		t.Fatalf("unexpected comment")
	}
}

func TestDeleteAndNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "comments", "posts", "users")

	userID := createUser(t, db, "comment2@example.com")
	postID := createPost(t, db, userID)

	repo := NewRepository(db)
	comment, err := repo.Create(context.Background(), domaincomment.Comment{
		PostID:   postID,
		AuthorID: userID,
		Content:  "bye",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	if err := repo.Delete(context.Background(), comment.ID); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	if _, err := repo.GetByID(context.Background(), comment.ID); !errors.Is(err, domaincomment.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := repo.Delete(context.Background(), 999); !errors.Is(err, domaincomment.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
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

func createPost(t *testing.T, db *sql.DB, authorID int64) int64 {
	t.Helper()
	var postID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO posts (author_id, content, visibility, created_at, updated_at)
		VALUES ($1, 'post', 'public', $2, $2)
		RETURNING id
	`, authorID, time.Now()).Scan(&postID)
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}
	return postID
}

func addCommentReaction(db *sql.DB, commentID, userID int64, reaction string) error {
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO comment_reactions (comment_id, user_id, reaction, created_at)
		VALUES ($1, $2, $3, $4)
	`, commentID, userID, reaction, time.Now())
	return err
}
