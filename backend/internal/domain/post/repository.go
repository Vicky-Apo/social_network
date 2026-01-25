package post

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a post does not exist.
var ErrNotFound = errors.New("post not found")

// Repository defines the data access contract for posts.
type Repository interface {
	List(ctx context.Context, limit, offset int) ([]Post, error)
	GetByID(ctx context.Context, id int64) (Post, error)
	Create(ctx context.Context, post Post, categoryIDs []int64) (Post, error)
	ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]Post, error)
	ListByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]Post, error)
	IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error)
}
