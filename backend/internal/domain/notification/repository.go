package notification

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a notification doesn't exist.
var ErrNotFound = errors.New("notification not found")

// Repository defines data access for notifications.
type Repository interface {
	Create(ctx context.Context, n Notification) (Notification, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int, unreadOnly bool) ([]Notification, error)
	MarkRead(ctx context.Context, userID, notificationID int64, readAt time.Time) error
	MarkAllRead(ctx context.Context, userID int64, readAt time.Time) (int64, error)
	UnreadCount(ctx context.Context, userID int64) (int64, error)
}
