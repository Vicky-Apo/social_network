package comment

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a comment doesn't exist
var ErrNotFound = errors.New("comment not found")

// Repository defines data access for comments
type Repository interface {
	Create(ctx context.Context, comment Comment) (Comment, error)
	GetByPostID(ctx context.Context, postID int64, limit, offset int) ([]Comment, error)
	CountByPostID(ctx context.Context, postID int64) (int, error)
	GetByID(ctx context.Context, id int64) (Comment, error)
	Update(ctx context.Context, comment Comment) (Comment, error)
	Delete(ctx context.Context, id int64) error
}
