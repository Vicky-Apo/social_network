package event

import "time"

// EventDTO represents an event in API responses.
type EventDTO struct {
	ID             int64     `json:"id"`
	GroupID        int64     `json:"group_id"`
	GroupTitle     *string   `json:"group_title,omitempty"`
	CreatorID      int64     `json:"creator_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	EventTime      time.Time `json:"event_time"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ResponsesCount int64     `json:"responses_count"`
}

// EventResponseDTO represents a user's response to an event.
type EventResponseDTO struct {
	EventID     int64      `json:"event_id"`
	UserID      int64      `json:"user_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Nickname    *string    `json:"nickname,omitempty"`
	AvatarPath  *string    `json:"avatar_path,omitempty"`
	Response    string     `json:"response"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// CreateEventRequest represents the request to create an event.
type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventTime   time.Time `json:"event_time"`
}

// UpdateEventRequest represents the request to update an event.
type UpdateEventRequest struct {
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventTime   time.Time `json:"event_time"`
}

// RespondRequest represents a request to respond to an event.
type RespondRequest struct {
	Response string `json:"response"`
}
