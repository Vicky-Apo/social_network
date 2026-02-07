package notification

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainnotification "social-network/backend/internal/domain/notification"
)

// Publisher sends real-time notifications (e.g., WebSocket).
type Publisher interface {
	Publish(ctx context.Context, userID int64, payload NotificationDTO) error
}

// Service handles notification business logic.
type Service struct {
	repo      domainnotification.Repository
	publisher Publisher
}

// NewService creates a notification service.
func NewService(repo domainnotification.Repository, publisher Publisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

// CreateForUser creates a notification for a user and publishes it if possible.
func (s *Service) CreateForUser(ctx context.Context, req CreateRequest) (NotificationDTO, error) {
	if req.UserID <= 0 {
		return NotificationDTO{}, errors.New("invalid user id")
	}
	if req.EntityID <= 0 {
		return NotificationDTO{}, errors.New("invalid entity id")
	}
	if req.EntityType == "" || req.Type == "" {
		return NotificationDTO{}, errors.New("type and entity_type are required")
	}

	var metaBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			return NotificationDTO{}, err
		}
		metaBytes = b
	}

	n := domainnotification.Notification{
		UserID:     req.UserID,
		ActorID:    req.ActorID,
		Type:       domainnotification.NotificationType(req.Type),
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Metadata:   metaBytes,
	}

	created, err := s.repo.Create(ctx, n)
	if err != nil {
		return NotificationDTO{}, err
	}

	dto := mapNotification(created)

	if s.publisher != nil {
		_ = s.publisher.Publish(context.Background(), req.UserID, dto)
	}

	return dto, nil
}

// List returns notifications for a user.
func (s *Service) List(ctx context.Context, userID int64, limit, offset int, unreadOnly bool) ([]NotificationDTO, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	items, err := s.repo.ListByUser(ctx, userID, limit, offset, unreadOnly)
	if err != nil {
		return nil, err
	}
	out := make([]NotificationDTO, 0, len(items))
	for _, n := range items {
		out = append(out, mapNotification(n))
	}
	return out, nil
}

// MarkRead marks a notification as read.
func (s *Service) MarkRead(ctx context.Context, userID, notificationID int64) error {
	if userID <= 0 || notificationID <= 0 {
		return errors.New("invalid user id or notification id")
	}
	return s.repo.MarkRead(ctx, userID, notificationID, time.Now())
}

// MarkAllRead marks all notifications as read.
func (s *Service) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}
	return s.repo.MarkAllRead(ctx, userID, time.Now())
}

// UnreadCount returns unread notifications count.
func (s *Service) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}
	return s.repo.UnreadCount(ctx, userID)
}

func mapNotification(n domainnotification.Notification) NotificationDTO {
	var meta map[string]any
	if len(n.Metadata) > 0 {
		_ = json.Unmarshal(n.Metadata, &meta)
	}
	return NotificationDTO{
		ID:         n.ID,
		UserID:     n.UserID,
		ActorID:    n.ActorID,
		Type:       string(n.Type),
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		Metadata:   meta,
		IsRead:     n.IsRead,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}
}
