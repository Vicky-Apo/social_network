package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	domainauth "social-network/backend/internal/domain/auth"
)

// Common column definitions to avoid duplication
const (
	userColumns = `id, email, password_hash, first_name, last_name, date_of_birth,
	               avatar_path, nickname, about, is_public, created_at, updated_at`

	sessionColumns = `id, user_id, session_token, user_agent, ip_address, expires_at, created_at`
)

// Repository implements the auth repository interface for PostgreSQL
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// scanner interface for scanning rows
type scanner interface {
	Scan(dest ...any) error
}

// scanUser scans a row into a User struct
func scanUser(s scanner) (domainauth.User, error) {
	var user domainauth.User
	err := s.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.AvatarPath,
		&user.Nickname,
		&user.About,
		&user.IsPublic,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

// scanSession scans a row into a Session struct
func scanSession(s scanner) (domainauth.Session, error) {
	var session domainauth.Session
	err := s.Scan(
		&session.ID,
		&session.UserID,
		&session.SessionToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	return session, err
}

// CreateUser inserts a new user record
func (r *Repository) CreateUser(ctx context.Context, user domainauth.User) (int64, error) {
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, date_of_birth,
		                   avatar_path, nickname, about, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.AvatarPath,
		user.Nickname,
		user.About,
		user.IsPublic,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		// Check for unique constraint violation (duplicate email)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return 0, domainauth.ErrEmailAlreadyExists
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (domainauth.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE email = $1`, userColumns)

	user, err := scanUser(r.db.QueryRowContext(ctx, query, email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainauth.User{}, domainauth.ErrUserNotFound
		}
		return domainauth.User{}, fmt.Errorf("query user by email: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (domainauth.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userColumns)

	user, err := scanUser(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainauth.User{}, domainauth.ErrUserNotFound
		}
		return domainauth.User{}, fmt.Errorf("query user by id: %w", err)
	}

	return user, nil
}

// CreateSession inserts a new session record
func (r *Repository) CreateSession(ctx context.Context, session domainauth.Session) (int64, error) {
	query := `
		INSERT INTO sessions (user_id, session_token, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		session.UserID,
		session.SessionToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}

	return id, nil
}

// GetSessionByToken retrieves a session by token
func (r *Repository) GetSessionByToken(ctx context.Context, token string) (domainauth.Session, error) {
	query := fmt.Sprintf(`SELECT %s FROM sessions WHERE session_token = $1 AND expires_at > NOW()`, sessionColumns)

	session, err := scanSession(r.db.QueryRowContext(ctx, query, token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainauth.Session{}, domainauth.ErrSessionNotFound
		}
		return domainauth.Session{}, fmt.Errorf("query session by token: %w", err)
	}

	return session, nil
}

// DeleteSession removes a session by token
func (r *Repository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE session_token = $1`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	// Note: It's OK if the session doesn't exist (idempotent operation)
	return nil
}

// DeleteUserSessions removes all sessions for a user
func (r *Repository) DeleteUserSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}

	return nil
}
