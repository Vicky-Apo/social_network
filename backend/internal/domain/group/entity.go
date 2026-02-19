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

// GroupInvitation represents a group invitation.
type GroupInvitation struct {
	ID        int64
	GroupID   int64
	InviterID int64
	InviteeID int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GroupJoinRequest represents a request to join a group.
type GroupJoinRequest struct {
	ID        int64
	GroupID   int64
	UserID    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GroupEvent represents a group event.
type GroupEvent struct {
	ID          int64
	GroupID     int64
	CreatorID   int64
	Title       string
	Description *string
	EventTime   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GroupEventResponse represents a user's response to an event.
type GroupEventResponse struct {
	EventID     int64
	UserID      int64
	Response    string
	RespondedAt *time.Time
}
