package group

import "time"

// GroupDTO represents a group in API responses.
type GroupDTO struct {
	ID          int64     `json:"id"`
	CreatorID   int64     `json:"creator_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MemberCount int64     `json:"member_count"`
	IsMember    bool      `json:"is_member"`
}

// GroupMemberDTO represents a member in a group list.
type GroupMemberDTO struct {
	UserID     int64     `json:"user_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Nickname   *string   `json:"nickname,omitempty"`
	AvatarPath *string   `json:"avatar_path,omitempty"`
	JoinedAt   time.Time `json:"joined_at"`
}

// GroupInvitationDTO represents a group invitation.
type GroupInvitationDTO struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	GroupTitle *string `json:"group_title,omitempty"`
	InviterID int64     `json:"inviter_id"`
	InviteeID int64     `json:"invitee_id"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupJoinRequestDTO represents a group join request.
type GroupJoinRequestDTO struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	User      *GroupMemberDTO `json:"user,omitempty"`
}

// CreateGroupRequest represents the request to create a group.
type CreateGroupRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}
