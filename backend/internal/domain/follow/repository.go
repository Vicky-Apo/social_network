package follow

import (
	"context"
	"errors"
)

// ErrRequestNotFound is returned when a follow request does not exist.
var ErrRequestNotFound = errors.New("follow request not found")

// Repository defines the data access contract for follow relationships.
type Repository interface {
	IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error)
	RequestExists(ctx context.Context, requesterID, targetID int64) (bool, error)
	CreateRequest(ctx context.Context, requesterID, targetID int64) (FollowRequest, error)
	GetRequestByID(ctx context.Context, id int64) (FollowRequest, error)
	DeleteRequest(ctx context.Context, id int64) error
	ListRequestsByTarget(ctx context.Context, targetID int64) ([]FollowRequest, error)
	CreateFollow(ctx context.Context, followerID, followingID int64) error
	DeleteFollow(ctx context.Context, followerID, followingID int64) error
}
