package notification

import (
	"encoding/json"
	"time"
)

// NotificationType represents a notification category.
type NotificationType string

const (
	FollowRequest    NotificationType = "follow_request"
	GroupInvitation  NotificationType = "group_invitation"
	GroupJoinRequest NotificationType = "group_join_request"
	EventCreated     NotificationType = "event_created"
	PostReaction     NotificationType = "post_reaction"
	CommentReaction  NotificationType = "comment_reaction"
	CommentOnPost    NotificationType = "comment_on_post"
)

// Notification represents a user notification.
type Notification struct {
	ID         int64
	UserID     int64
	ActorID    *int64
	Type       NotificationType
	EntityType string
	EntityID   int64
	Metadata   json.RawMessage
	IsRead     bool
	ReadAt     *time.Time
	CreatedAt  time.Time
}
