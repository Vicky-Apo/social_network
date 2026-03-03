package event

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domainevent "social-network/backend/internal/domain/event"
	domaingroup "social-network/backend/internal/domain/group"
	usecasenotification "social-network/backend/internal/usecase/notification"
	"social-network/backend/pkg/logger"
)

// Service errors.
var (
	ErrInvalidTitle     = errors.New("invalid event title")
	ErrInvalidEventTime = errors.New("invalid event time")
	ErrForbidden        = errors.New("event action forbidden")
	ErrInvalidResponse  = errors.New("invalid response")
	ErrNotCreator       = errors.New("only event creator can modify event")
)

// AccessService provides centralized access checks.
type AccessService interface {
	CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error)
	CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error)
}

// Notifier allows emitting notifications without coupling to transport details.
type Notifier interface {
	CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error)
}

// Service orchestrates event-related use cases.
type Service struct {
	repo      domainevent.Repository
	groupRepo domaingroup.Repository
	access    AccessService
	notifier  Notifier
	log       logger.Logger
}

// NewService builds an event service with the given repositories.
func NewService(repo domainevent.Repository, groupRepo domaingroup.Repository, access AccessService, notifier Notifier, log logger.Logger) *Service {
	return &Service{
		repo:      repo,
		groupRepo: groupRepo,
		access:    access,
		notifier:  notifier,
		log:       log.WithFields(logger.F("service", "event")),
	}
}

// CreateEvent creates a new group event.
func (s *Service) CreateEvent(ctx context.Context, creatorID, groupID int64, req CreateEventRequest) (EventDTO, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return EventDTO{}, ErrInvalidTitle
	}
	if req.EventTime.IsZero() {
		return EventDTO{}, ErrInvalidEventTime
	}
	if req.EventTime.Before(time.Now()) {
		return EventDTO{}, ErrInvalidEventTime
	}
	if s.access == nil {
		return EventDTO{}, errors.New("access service not configured")
	}
	canPost, err := s.access.CanPostInGroup(ctx, creatorID, groupID)
	if err != nil {
		return EventDTO{}, fmt.Errorf("check group access: %w", err)
	}
	if !canPost {
		return EventDTO{}, ErrForbidden
	}

	event := domainevent.Event{
		GroupID:     groupID,
		CreatorID:   creatorID,
		Title:       title,
		Description: req.Description,
		EventTime:   req.EventTime,
	}
	created, err := s.repo.Create(ctx, event)
	if err != nil {
		return EventDTO{}, fmt.Errorf("create event: %w", err)
	}

	s.notifyGroupMembers(ctx, created)

	return mapEvent(created), nil
}

// ListGroupEvents returns events for a group.
func (s *Service) ListGroupEvents(ctx context.Context, groupID, viewerID int64, limit, offset int) ([]EventDTO, error) {
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	canView, err := s.access.CanViewGroup(ctx, viewerID, groupID)
	if err != nil {
		return nil, fmt.Errorf("check group access: %w", err)
	}
	if !canView {
		return nil, ErrForbidden
	}
	items, err := s.repo.ListByGroup(ctx, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	out := make([]EventDTO, 0, len(items))
	for _, ev := range items {
		out = append(out, mapEvent(ev))
	}
	return out, nil
}

// GetEvent returns an event by ID after access checks.
func (s *Service) GetEvent(ctx context.Context, eventID, viewerID int64) (EventDTO, error) {
	ev, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return EventDTO{}, err
	}
	if s.access == nil {
		return EventDTO{}, errors.New("access service not configured")
	}
	canView, err := s.access.CanViewGroup(ctx, viewerID, ev.GroupID)
	if err != nil {
		return EventDTO{}, fmt.Errorf("check group access: %w", err)
	}
	if !canView {
		return EventDTO{}, ErrForbidden
	}
	return mapEvent(ev), nil
}

// UpdateEvent updates an event (creator only).
func (s *Service) UpdateEvent(ctx context.Context, eventID, userID int64, req UpdateEventRequest) (EventDTO, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return EventDTO{}, ErrInvalidTitle
	}
	if req.EventTime.IsZero() {
		return EventDTO{}, ErrInvalidEventTime
	}
	if req.EventTime.Before(time.Now()) {
		return EventDTO{}, ErrInvalidEventTime
	}

	ev, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return EventDTO{}, err
	}
	if ev.CreatorID != userID {
		return EventDTO{}, ErrNotCreator
	}

	ev.Title = title
	ev.Description = req.Description
	ev.EventTime = req.EventTime

	updated, err := s.repo.Update(ctx, ev)
	if err != nil {
		return EventDTO{}, fmt.Errorf("update event: %w", err)
	}
	return mapEvent(updated), nil
}

// DeleteEvent deletes an event (creator only).
func (s *Service) DeleteEvent(ctx context.Context, eventID, userID int64) error {
	ev, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}
	if ev.CreatorID != userID {
		return ErrNotCreator
	}
	return s.repo.Delete(ctx, eventID)
}

// Respond sets the user's response to an event.
func (s *Service) Respond(ctx context.Context, eventID, userID int64, response string) (EventResponseDTO, error) {
	resp := strings.TrimSpace(response)
	if resp == "" {
		return EventResponseDTO{}, ErrInvalidResponse
	}
	switch resp {
	case "going", "not_going":
	default:
		return EventResponseDTO{}, ErrInvalidResponse
	}

	ev, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return EventResponseDTO{}, err
	}
	if s.access == nil {
		return EventResponseDTO{}, errors.New("access service not configured")
	}
	canView, err := s.access.CanViewGroup(ctx, userID, ev.GroupID)
	if err != nil {
		return EventResponseDTO{}, fmt.Errorf("check group access: %w", err)
	}
	if !canView {
		return EventResponseDTO{}, ErrForbidden
	}

	stored, err := s.repo.UpsertResponse(ctx, eventID, userID, resp)
	if err != nil {
		return EventResponseDTO{}, fmt.Errorf("respond to event: %w", err)
	}

	return EventResponseDTO{
		EventID:     stored.EventID,
		UserID:      stored.UserID,
		Response:    stored.Response,
		RespondedAt: stored.RespondedAt,
	}, nil
}

// ListResponses lists responses for an event.
func (s *Service) ListResponses(ctx context.Context, eventID, viewerID int64) ([]EventResponseDTO, error) {
	ev, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	canView, err := s.access.CanViewGroup(ctx, viewerID, ev.GroupID)
	if err != nil {
		return nil, fmt.Errorf("check group access: %w", err)
	}
	if !canView {
		return nil, ErrForbidden
	}
	items, err := s.repo.ListResponses(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list responses: %w", err)
	}
	out := make([]EventResponseDTO, 0, len(items))
	for _, r := range items {
		out = append(out, EventResponseDTO{
			EventID:     r.EventID,
			UserID:      r.UserID,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			Nickname:    r.Nickname,
			AvatarPath:  r.AvatarPath,
			Response:    r.Response,
			RespondedAt: r.RespondedAt,
		})
	}
	return out, nil
}

func (s *Service) notifyGroupMembers(ctx context.Context, ev domainevent.Event) {
	if s.notifier == nil || s.groupRepo == nil {
		return
	}
	memberIDs, err := s.groupRepo.GetMemberIDs(ctx, ev.GroupID)
	if err != nil {
		return
	}
	groupTitle := ""
	if group, err := s.groupRepo.GetByID(ctx, ev.GroupID); err == nil {
		groupTitle = group.Title
	}
	for _, id := range memberIDs {
		if id == ev.CreatorID {
			continue
		}
		if _, err := s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
			UserID:     id,
			ActorID:    &ev.CreatorID,
			Type:       "event_created",
			EntityType: "event",
			EntityID:   ev.ID,
			Metadata: map[string]any{
				"group_id":   ev.GroupID,
				"group_name": groupTitle,
				"title":      ev.Title,
				"event_time": ev.EventTime.Format(time.RFC3339),
			},
		}); err != nil && s.log != nil {
			s.log.Debug("failed to notify event created", logger.F("error", err.Error()), logger.F("user_id", id))
		}
	}
}

func mapEvent(e domainevent.Event) EventDTO {
	return EventDTO{
		ID:             e.ID,
		GroupID:        e.GroupID,
		GroupTitle:     e.GroupTitle,
		CreatorID:      e.CreatorID,
		Title:          e.Title,
		Description:    e.Description,
		EventTime:      e.EventTime,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		ResponsesCount: e.ResponsesCount,
	}
}
