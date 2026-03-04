package post

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a post does not exist.
var ErrNotFound = errors.New("post not found")

// Repository defines the data access contract for posts.
type Repository interface {
	List(ctx context.Context, viewerID int64, limit, offset int) ([]Post, error)
	Count(ctx context.Context, viewerID int64) (int, error)
	ListGroupsOnly(ctx context.Context, viewerID int64, limit, offset int) ([]Post, error)
	CountGroupsOnly(ctx context.Context, viewerID int64) (int, error)
	ListPublicOnly(ctx context.Context, limit, offset int) ([]Post, error)
	CountPublicOnly(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id int64) (Post, error)
	Create(ctx context.Context, post Post, allowedUserIDs []int64) (Post, error)
	Update(ctx context.Context, post Post, allowedUserIDs []int64) (Post, error)
	Delete(ctx context.Context, id int64) error
	ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]Post, error)
	CountByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool) (int, error)
	ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]Post, error)
	CountByGroup(ctx context.Context, groupID int64) (int, error)
	IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error)
}
