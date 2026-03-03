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

// GroupSummary represents a group with extra metadata.
type GroupSummary struct {
	Group
	MemberCount int64
	IsMember    bool
}

// GroupMemberInfo represents a group member with user details.
type GroupMemberInfo struct {
	UserID     int64
	FirstName  string
	LastName   string
	Nickname   *string
	AvatarPath *string
	JoinedAt   time.Time
}

// GroupInvitation represents an invitation to join a group.
type GroupInvitation struct {
	ID        int64
	GroupID   int64
	GroupTitle *string
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
	User      *GroupMemberInfo
}
