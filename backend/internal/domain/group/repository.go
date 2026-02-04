package group

import (
	"context"
	"errors"
)

// Errors for the group domain.
var (
	ErrGroupNotFound = errors.New("group not found")
	ErrNotMember     = errors.New("user is not a member of this group")
)

// Repository defines the data access contract for group operations.
type Repository interface {
	GetByID(ctx context.Context, id int64) (Group, error)
	IsMember(ctx context.Context, groupID, userID int64) (bool, error)
	GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error)
}
