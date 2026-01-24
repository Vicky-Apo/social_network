package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainuser "social-network/backend/internal/domain/user"
)

// Repository implements the user repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres user repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByID returns a user by ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	const query = `
		SELECT id, email, first_name, last_name, date_of_birth, avatar_path, nickname, about, is_public, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domainuser.User
	var avatar, nickname, about sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.DateOfBirth,
		&avatar,
		&nickname,
		&about,
		&u.IsPublic,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainuser.User{}, domainuser.ErrNotFound
		}
		return domainuser.User{}, fmt.Errorf("get user: %w", err)
	}
	u.AvatarPath = nullableString(avatar)
	u.Nickname = nullableString(nickname)
	u.About = nullableString(about)
	return u, nil
}

// SetVisibility updates a user's public flag.
func (r *Repository) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	const query = `
		UPDATE 	
		SET is_public = $1
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, isPublic, id)
	if err != nil {
		return fmt.Errorf("set visibility: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set visibility: %w", err)
	}
	if affected == 0 {
		return domainuser.ErrNotFound
	}
	return nil
}

// CountFollowers returns the number of followers for a user.
func (r *Repository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM follows
		WHERE following_id = $1
	`
	var count int64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count followers: %w", err)
	}
	return count, nil
}

// CountFollowing returns the number of users a user follows.
func (r *Repository) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = $1
	`
	var count int64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count following: %w", err)
	}
	return count, nil
}

// ListFollowers returns the user profiles of followers.
func (r *Repository) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) {
	const query = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.date_of_birth, u.avatar_path, u.nickname, u.about, u.is_public, u.created_at, u.updated_at
		FROM follows f
		JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY u.id
	`
	return r.listUsers(ctx, query, userID)
}

// ListFollowing returns the user profiles of followed users.
func (r *Repository) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) {
	const query = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.date_of_birth, u.avatar_path, u.nickname, u.about, u.is_public, u.created_at, u.updated_at
		FROM follows f
		JOIN users u ON u.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY u.id
	`
	return r.listUsers(ctx, query, userID)
}

// ListUsers returns all user profiles.
func (r *Repository) ListUsers(ctx context.Context) ([]domainuser.User, error) {
	const query = `
		SELECT id, email, first_name, last_name, date_of_birth, avatar_path, nickname, about, is_public, created_at, updated_at
		FROM users
		ORDER BY id
	`
	return r.listUsersWithArgs(ctx, query)
}

// SearchUsers returns users matching the query in first name, last name, or nickname.
func (r *Repository) SearchUsers(ctx context.Context, query string) ([]domainuser.User, error) {
	const sqlQuery = `
		SELECT id, email, first_name, last_name, date_of_birth, avatar_path, nickname, about, is_public, created_at, updated_at
		FROM users
		WHERE first_name ILIKE $1 OR last_name ILIKE $1 OR nickname ILIKE $1
		ORDER BY id
	`
	pattern := "%" + query + "%"
	return r.listUsersWithArgs(ctx, sqlQuery, pattern)
}

func (r *Repository) listUsers(ctx context.Context, query string, userID int64) ([]domainuser.User, error) {
	return r.listUsersWithArgs(ctx, query, userID)
}

func (r *Repository) listUsersWithArgs(ctx context.Context, query string, args ...any) ([]domainuser.User, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domainuser.User
	for rows.Next() {
		var u domainuser.User
		var avatar, nickname, about sql.NullString
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.DateOfBirth,
			&avatar,
			&nickname,
			&about,
			&u.IsPublic,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		u.AvatarPath = nullableString(avatar)
		u.Nickname = nullableString(nickname)
		u.About = nullableString(about)
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
