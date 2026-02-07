package websocket

import (
	"context"

	usecasenotification "social-network/backend/internal/usecase/notification"
)

// NotificationPublisher pushes notifications to connected users.
type NotificationPublisher struct {
	hub *Hub
}

// NewNotificationPublisher creates a notification publisher backed by the hub.
func NewNotificationPublisher(hub *Hub) *NotificationPublisher {
	return &NotificationPublisher{hub: hub}
}

// Publish sends a notification payload to a user over WebSocket.
func (p *NotificationPublisher) Publish(ctx context.Context, userID int64, n usecasenotification.NotificationDTO) error {
	payload := NotificationPayload{
		ID:         n.ID,
		UserID:     n.UserID,
		ActorID:    n.ActorID,
		Type:       n.Type,
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		Metadata:   n.Metadata,
		IsRead:     n.IsRead,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}
	msg, err := NewWSMessage(MessageTypeNotification, payload)
	if err != nil {
		return err
	}
	p.hub.SendToUser(userID, msg)
	return nil
}
