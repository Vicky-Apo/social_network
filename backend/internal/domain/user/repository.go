package user

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// Repository defines the data access contract for users and profile data.
type Repository interface {
	GetByID(ctx context.Context, id int64) (User, error)
	SetVisibility(ctx context.Context, id int64, isPublic bool) error
	CountFollowers(ctx context.Context, userID int64) (int64, error)
	CountFollowing(ctx context.Context, userID int64) (int64, error)
	ListFollowers(ctx context.Context, userID int64) ([]User, error)
	ListFollowing(ctx context.Context, userID int64) ([]User, error)
	ListUsers(ctx context.Context) ([]User, error)
	SearchUsers(ctx context.Context, query string) ([]User, error)
}
