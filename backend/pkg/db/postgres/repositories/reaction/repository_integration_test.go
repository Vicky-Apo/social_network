//go:build integration

package reaction

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainreaction "social-network/backend/internal/domain/reaction"
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

func TestPostReactions_AddUpdateRemove(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "post_reactions", "posts", "users")

	userID := createUser(t, db, "reaction-post@example.com")
	postID := createPost(t, db, userID)

	repo := NewRepository(db)
	err := repo.AddPostReaction(context.Background(), domainreaction.PostReaction{
		PostID:   postID,
		UserID:   userID,
		Reaction: domainreaction.Like,
	})
	if err != nil {
		t.Fatalf("add reaction: %v", err)
	}

	got, err := repo.GetPostReaction(context.Background(), postID, userID)
	if err != nil {
		t.Fatalf("get reaction: %v", err)
	}
	if got.Reaction != domainreaction.Like {
		t.Fatalf("expected like")
	}

	err = repo.AddPostReaction(context.Background(), domainreaction.PostReaction{
		PostID:   postID,
		UserID:   userID,
		Reaction: domainreaction.Dislike,
	})
	if err != nil {
		t.Fatalf("update reaction: %v", err)
	}

	got, err = repo.GetPostReaction(context.Background(), postID, userID)
	if err != nil {
		t.Fatalf("get reaction: %v", err)
	}
	if got.Reaction != domainreaction.Dislike {
		t.Fatalf("expected dislike")
	}

	if err := repo.RemovePostReaction(context.Background(), postID, userID); err != nil {
		t.Fatalf("remove reaction: %v", err)
	}
	_, err = repo.GetPostReaction(context.Background(), postID, userID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected no rows, got %v", err)
	}
}

func TestCommentReactions_AddListRemove(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "comment_reactions", "comments", "posts", "users")

	userID := createUser(t, db, "reaction-comment@example.com")
	postID := createPost(t, db, userID)
	commentID := createComment(t, db, postID, userID)

	repo := NewRepository(db)
	err := repo.AddCommentReaction(context.Background(), domainreaction.CommentReaction{
		CommentID: commentID,
		UserID:    userID,
		Reaction:  domainreaction.Like,
	})
	if err != nil {
		t.Fatalf("add comment reaction: %v", err)
	}

	reactions, err := repo.GetCommentReactions(context.Background(), commentID)
	if err != nil {
		t.Fatalf("list comment reactions: %v", err)
	}
	if len(reactions) != 1 || reactions[0].Reaction != domainreaction.Like {
		t.Fatalf("unexpected reactions")
	}

	if err := repo.RemoveCommentReaction(context.Background(), commentID, userID); err != nil {
		t.Fatalf("remove comment reaction: %v", err)
	}
	reactions, err = repo.GetCommentReactions(context.Background(), commentID)
	if err != nil {
		t.Fatalf("list comment reactions: %v", err)
	}
	if len(reactions) != 0 {
		t.Fatalf("expected no reactions")
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

func createPost(t *testing.T, db *sql.DB, authorID int64) int64 {
	t.Helper()
	var postID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO posts (author_id, content, visibility, created_at, updated_at)
		VALUES ($1, 'hello', 'public', $2, $2)
		RETURNING id
	`, authorID, time.Now()).Scan(&postID)
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}
	return postID
}

func createComment(t *testing.T, db *sql.DB, postID, authorID int64) int64 {
	t.Helper()
	var commentID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO comments (post_id, author_id, content, created_at, updated_at)
		VALUES ($1, $2, 'comment', $3, $3)
		RETURNING id
	`, postID, authorID, time.Now()).Scan(&commentID)
	if err != nil {
		t.Fatalf("insert comment: %v", err)
	}
	return commentID
}
