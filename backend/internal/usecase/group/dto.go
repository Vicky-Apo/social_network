package group

import "time"

// GroupDTO represents a group in responses.
type GroupDTO struct {
	ID          int64     `json:"id"`
	CreatorID   int64     `json:"creator_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMemberDTO represents a group member.
type GroupMemberDTO struct {
	GroupID  int64     `json:"group_id"`
	UserID   int64     `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// GroupInvitationDTO represents a group invitation.
type GroupInvitationDTO struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	InviterID int64     `json:"inviter_id"`
	InviteeID int64     `json:"invitee_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupJoinRequestDTO represents a group join request.
type GroupJoinRequestDTO struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupEventDTO represents a group event.
type GroupEventDTO struct {
	ID          int64     `json:"id"`
	GroupID     int64     `json:"group_id"`
	CreatorID   int64     `json:"creator_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventTime   time.Time `json:"event_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupEventResponseDTO represents an RSVP response.
type GroupEventResponseDTO struct {
	EventID     int64      `json:"event_id"`
	UserID      int64      `json:"user_id"`
	Response    string     `json:"response"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// CreateGroupRequest is the request to create a group.
type CreateGroupRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

// UpdateGroupRequest is the request to update a group.
type UpdateGroupRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateEventRequest is the request to create an event.
type CreateEventRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	EventTime   *time.Time `json:"event_time"`
}
