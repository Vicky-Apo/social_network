package follow

import "time"

// FollowRequest represents a follow request with status.
type FollowRequest struct {
	ID          int64
	RequesterID int64
	TargetID    int64
	Status      string
	CreatedAt   time.Time
	Requester   *UserInfo
	Target      *UserInfo
}

// UserInfo is a lightweight user view for follow requests.
type UserInfo struct {
	ID         int64
	FirstName  string
	LastName   string
	Nickname   *string
	AvatarPath *string
}

// Follow represents a follower -> following relationship.
type Follow struct {
	FollowerID  int64
	FollowingID int64
	CreatedAt   time.Time
}
