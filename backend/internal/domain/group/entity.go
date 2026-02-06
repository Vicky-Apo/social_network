package group

import "time"

// Group represents a social group.
type Group struct {
	ID          int64
	CreatorID   int64
	Title       string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GroupMember represents a member of a group.
type GroupMember struct {
	GroupID  int64
	UserID   int64
	JoinedAt time.Time
}
