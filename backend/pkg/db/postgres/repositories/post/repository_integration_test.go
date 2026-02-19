//go:build integration

package post

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

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
		_, _ = db.ExecContext(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE")
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

func TestList_ExcludesGroupPosts(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "post_allowed_users", "post_categories", "comments", "posts", "group_members", "groups", "users")

	// user
	var userID int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, is_public)
		VALUES ('u1@example.com', 'hash', 'U', 'One', '2000-01-01', true)
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	// group
	var groupID int64
	err = db.QueryRowContext(context.Background(), `
		INSERT INTO groups (creator_id, title, description)
		VALUES ($1, 'G', 'D')
		RETURNING id
	`, userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("insert group: %v", err)
	}

	// group post
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO posts (author_id, group_id, content, visibility)
		VALUES ($1, $2, 'group post', 'public')
	`, userID, groupID)
	if err != nil {
		t.Fatalf("insert group post: %v", err)
	}

	// normal post
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO posts (author_id, content, visibility)
		VALUES ($1, 'normal post', 'public')
	`, userID)
	if err != nil {
		t.Fatalf("insert normal post: %v", err)
	}

	repo := NewRepository(db)
	posts, err := repo.List(context.Background(), userID, 20, 0)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Content != "normal post" {
		t.Fatalf("expected normal post, got %q", posts[0].Content)
	}
}
