//go:build integration

package user

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainuser "social-network/backend/internal/domain/user"
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

func TestGetByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "users")

	repo := NewRepository(db)
	if _, err := repo.GetByID(context.Background(), 999); !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestUpdateProfileAndVisibility(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "users")

	userID := createUser(t, db, "user1@example.com", "Alice", "One", "alice")
	repo := NewRepository(db)

	nickname := "newnick"
	about := "about me"
	avatar := "/uploads/avatar.png"
	updated, err := repo.UpdateProfile(context.Background(), userID, &nickname, &about, &avatar)
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.Nickname == nil || *updated.Nickname != nickname {
		t.Fatalf("expected nickname updated")
	}
	if updated.About == nil || *updated.About != about {
		t.Fatalf("expected about updated")
	}
	if updated.AvatarPath == nil || *updated.AvatarPath != avatar {
		t.Fatalf("expected avatar updated")
	}

	empty := "   "
	updated, err = repo.UpdateProfile(context.Background(), userID, &empty, nil, nil)
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.Nickname != nil {
		t.Fatalf("expected nickname cleared")
	}

	if err := repo.SetVisibility(context.Background(), userID, false); err != nil {
		t.Fatalf("set visibility: %v", err)
	}
	got, err := repo.GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.IsPublic {
		t.Fatalf("expected private")
	}
}

func TestFollowersFollowingAndLists(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "follows", "users")

	u1 := createUser(t, db, "user2@example.com", "Bob", "Two", "bobby")
	u2 := createUser(t, db, "user3@example.com", "Charlie", "Three", "")
	u3 := createUser(t, db, "user4@example.com", "Dora", "Four", "")

	if err := addFollow(db, u2, u1); err != nil {
		t.Fatalf("add follow: %v", err)
	}
	if err := addFollow(db, u1, u3); err != nil {
		t.Fatalf("add follow: %v", err)
	}

	repo := NewRepository(db)
	followers, err := repo.CountFollowers(context.Background(), u1)
	if err != nil || followers != 1 {
		t.Fatalf("expected 1 follower, got %d, err %v", followers, err)
	}
	following, err := repo.CountFollowing(context.Background(), u1)
	if err != nil || following != 1 {
		t.Fatalf("expected 1 following, got %d, err %v", following, err)
	}

	listFollowers, err := repo.ListFollowers(context.Background(), u1)
	if err != nil || len(listFollowers) != 1 {
		t.Fatalf("expected 1 follower in list")
	}
	listFollowing, err := repo.ListFollowing(context.Background(), u1)
	if err != nil || len(listFollowing) != 1 {
		t.Fatalf("expected 1 following in list")
	}
}

func TestListAndSearchUsers(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	cleanup(t, db, "users")

	_ = createUser(t, db, "user5@example.com", "Eve", "Five", "eve")
	_ = createUser(t, db, "user6@example.com", "Frank", "Six", "franky")

	repo := NewRepository(db)
	users, err := repo.ListUsers(context.Background())
	if err != nil || len(users) != 2 {
		t.Fatalf("expected 2 users")
	}

	search, err := repo.SearchUsers(context.Background(), "frank")
	if err != nil {
		t.Fatalf("search users: %v", err)
	}
	if len(search) != 1 || search[0].Email != "user6@example.com" {
		t.Fatalf("unexpected search results")
	}
}

func createUser(t *testing.T, db *sql.DB, email, first, last, nickname string) int64 {
	t.Helper()
	var userID int64
	var nick any
	if nickname != "" {
		nick = nickname
	}
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth, nickname, is_public)
		VALUES ($1, 'hash', $2, $3, $4, $5, true)
		RETURNING id
	`, email, first, last, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), nick).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return userID
}

func addFollow(db *sql.DB, followerID, followingID int64) error {
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO follows (follower_id, following_id)
		VALUES ($1, $2)
	`, followerID, followingID)
	return err
}
