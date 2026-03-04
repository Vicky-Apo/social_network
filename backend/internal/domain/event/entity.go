package event

import "time"

// Event represents a group event.
type Event struct {
	ID          int64
	GroupID     int64
	CreatorID   int64
	Title       string
	Description *string
	EventTime   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	GroupTitle  *string
	ResponsesCount int64
}

// EventResponse represents a user's response to an event.
type EventResponse struct {
	EventID     int64
	UserID      int64
	Response    string
	RespondedAt *time.Time
}

// EventResponseInfo includes user metadata for response lists.
type EventResponseInfo struct {
	EventID     int64
	UserID      int64
	FirstName   string
	LastName    string
	Nickname    *string
	AvatarPath  *string
	Response    string
	RespondedAt *time.Time
}
