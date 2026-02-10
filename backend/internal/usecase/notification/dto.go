package notification

import "time"

// NotificationDTO is the API-facing notification.
type NotificationDTO struct {
	ID         int64            `json:"id"`
	UserID     int64            `json:"user_id"`
	ActorID    *int64           `json:"actor_id,omitempty"`
	Type       string           `json:"type"`
	EntityType string           `json:"entity_type"`
	EntityID   int64            `json:"entity_id"`
	Metadata   map[string]any   `json:"metadata,omitempty"`
	IsRead     bool             `json:"is_read"`
	ReadAt     *time.Time       `json:"read_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

// CreateRequest is the input for creating a notification.
type CreateRequest struct {
	UserID     int64
	ActorID    *int64
	Type       string
	EntityType string
	EntityID   int64
	Metadata   map[string]any
}
