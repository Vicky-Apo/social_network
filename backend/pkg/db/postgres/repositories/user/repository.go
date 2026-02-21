package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domainuser "social-network/backend/internal/domain/user"
	reposhared "social-network/backend/pkg/db/postgres/repositories/shared"
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
	u.AvatarPath = reposhared.NullableString(avatar)
	u.Nickname = reposhared.NullableString(nickname)
	u.About = reposhared.NullableString(about)
	return u, nil
}

// UpdateProfile updates optional profile fields and returns the updated user.
func (r *Repository) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if nickname != nil {
		setClauses = append(setClauses, fmt.Sprintf("nickname = $%d", len(args)+1))
		args = append(args, reposhared.NullableStringValue(nickname))
	}
	if about != nil {
		setClauses = append(setClauses, fmt.Sprintf("about = $%d", len(args)+1))
		args = append(args, reposhared.NullableStringValue(about))
	}
	if avatarPath != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_path = $%d", len(args)+1))
		args = append(args, reposhared.NullableStringValue(avatarPath))
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
		RETURNING id, email, first_name, last_name, date_of_birth, avatar_path, nickname, about, is_public, created_at, updated_at
	`, strings.Join(setClauses, ", "), len(args))

	var u domainuser.User
	var avatar, nick, aboutStr sql.NullString
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.DateOfBirth,
		&avatar,
		&nick,
		&aboutStr,
		&u.IsPublic,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainuser.User{}, domainuser.ErrNotFound
		}
		return domainuser.User{}, fmt.Errorf("update profile: %w", err)
	}
	u.AvatarPath = reposhared.NullableString(avatar)
	u.Nickname = reposhared.NullableString(nick)
	u.About = reposhared.NullableString(aboutStr)
	return u, nil
}

// SetVisibility updates a user's public flag.
func (r *Repository) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	const query = `
		UPDATE users
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

// ListUsers returns user profiles visible to the viewer with pagination.
func (r *Repository) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]domainuser.User, error) {
	const query = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.date_of_birth, u.avatar_path, u.nickname, u.about, u.is_public, u.created_at, u.updated_at
		FROM users u
		WHERE u.id <> $1
		ORDER BY u.id
		LIMIT $2 OFFSET $3
	`
	return r.listUsersWithArgs(ctx, query, viewerID, limit, offset)
}

// SearchUsers returns users matching the query in first name, last name, or nickname.
func (r *Repository) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]domainuser.User, error) {
	const sqlQuery = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.date_of_birth, u.avatar_path, u.nickname, u.about, u.is_public, u.created_at, u.updated_at
		FROM users u
		WHERE u.id <> $1
		  AND (u.first_name ILIKE $2 ESCAPE '\' OR u.last_name ILIKE $2 ESCAPE '\' OR u.nickname ILIKE $2 ESCAPE '\')
		ORDER BY u.id
		LIMIT $3 OFFSET $4
	`
	escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(query)
	pattern := "%" + escaped + "%"
	return r.listUsersWithArgs(ctx, sqlQuery, viewerID, pattern, limit, offset)
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
		u.AvatarPath = reposhared.NullableString(avatar)
		u.Nickname = reposhared.NullableString(nickname)
		u.About = reposhared.NullableString(about)
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// nullableString helpers moved to repositories/shared.
