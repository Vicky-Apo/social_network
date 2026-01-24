package auth

import (
	"context"
	"errors"
)

// Domain-level errors (following post pattern)
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

// Repository defines data access contract for authentication
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id int64) (User, error)

	// Session operations
	CreateSession(ctx context.Context, session Session) (int64, error)
	GetSessionByToken(ctx context.Context, token string) (Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID int64) error
}
