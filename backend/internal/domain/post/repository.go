package post

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a post does not exist.
var ErrNotFound = errors.New("post not found")

// Repository defines the data access contract for posts.
type Repository interface {
	List(ctx context.Context) ([]Post, error)
	GetByID(ctx context.Context, id int64) (Post, error)
}
