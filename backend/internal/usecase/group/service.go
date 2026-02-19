package group

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domaingroup "social-network/backend/internal/domain/group"
	domainuser "social-network/backend/internal/domain/user"
	usecaseaccess "social-network/backend/internal/usecase/access"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

// Errors for group operations.
var (
	ErrInvalidRequest = errors.New("invalid group request")
	ErrForbidden      = errors.New("group action forbidden")
	ErrInvalidStatus  = errors.New("invalid status")
	ErrInvalidResponse = errors.New("invalid response")
)

// Service orchestrates group-related use cases.
type Service struct {
	groupRepo domaingroup.Repository
	userRepo  domainuser.Repository
	access    *usecaseaccess.Service
	notifier  Notifier
}

// Notifier allows emitting notifications without coupling to transport details.
type Notifier interface {
	CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error)
}

// NewService builds a group service with the given repositories.
func NewService(groupRepo domaingroup.Repository, userRepo domainuser.Repository, access *usecaseaccess.Service, notifier Notifier) *Service {
	return &Service{
		groupRepo: groupRepo,
		userRepo:  userRepo,
		access:    access,
		notifier:  notifier,
	}
}

// Create creates a new group and adds the creator as a member.
func (s *Service) Create(ctx context.Context, creatorID int64, req CreateGroupRequest) (GroupDTO, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return GroupDTO{}, ErrInvalidRequest
	}
	if _, err := s.userRepo.GetByID(ctx, creatorID); err != nil {
		return GroupDTO{}, err
	}
	group := domaingroup.Group{
		CreatorID:   creatorID,
		Title:       title,
		Description: req.Description,
	}
	created, err := s.groupRepo.Create(ctx, group)
	if err != nil {
		return GroupDTO{}, fmt.Errorf("create group: %w", err)
	}
	if err := s.groupRepo.AddMember(ctx, created.ID, creatorID); err != nil {
		return GroupDTO{}, fmt.Errorf("add creator as member: %w", err)
	}
	return mapGroup(created), nil
}

// List returns groups filtered by query.
func (s *Service) List(ctx context.Context, query string, limit, offset int) ([]GroupDTO, error) {
	groups, err := s.groupRepo.List(ctx, strings.TrimSpace(query), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	return mapGroups(groups), nil
}

// Get returns a group by ID.
func (s *Service) Get(ctx context.Context, id int64) (GroupDTO, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return GroupDTO{}, err
	}
	return mapGroup(group), nil
}

// Update updates a group if the actor is the creator.
func (s *Service) Update(ctx context.Context, groupID, actorID int64, req UpdateGroupRequest) (GroupDTO, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return GroupDTO{}, err
	}
	if group.CreatorID != actorID {
		return GroupDTO{}, ErrForbidden
	}
	if req.Title == nil && req.Description == nil {
		return GroupDTO{}, ErrInvalidRequest
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return GroupDTO{}, ErrInvalidRequest
		}
		group.Title = title
	}
	if req.Description != nil {
		group.Description = req.Description
	}
	updated, err := s.groupRepo.Update(ctx, group)
	if err != nil {
		return GroupDTO{}, err
	}
	return mapGroup(updated), nil
}

// Delete removes a group if the actor is the creator.
func (s *Service) Delete(ctx context.Context, groupID, actorID int64) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatorID != actorID {
		return ErrForbidden
	}
	return s.groupRepo.Delete(ctx, groupID)
}

// ListMembers returns members of a group.
func (s *Service) ListMembers(ctx context.Context, groupID, actorID int64, limit, offset int) ([]GroupMemberDTO, error) {
	if s.access != nil {
		ok, err := s.access.CanViewGroup(ctx, actorID, groupID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	members, err := s.groupRepo.ListMembers(ctx, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapMembers(members), nil
}

// Leave removes the actor from the group.
func (s *Service) Leave(ctx context.Context, groupID, actorID int64) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatorID == actorID {
		return ErrForbidden
	}
	return s.groupRepo.RemoveMember(ctx, groupID, actorID)
}

// RemoveMember removes a member from a group if actor is creator.
func (s *Service) RemoveMember(ctx context.Context, groupID, actorID, memberID int64) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatorID != actorID {
		return ErrForbidden
	}
	if group.CreatorID == memberID {
		return ErrForbidden
	}
	return s.groupRepo.RemoveMember(ctx, groupID, memberID)
}

// Invite creates a group invitation.
func (s *Service) Invite(ctx context.Context, groupID, inviterID, inviteeID int64) (GroupInvitationDTO, error) {
	if inviterID == inviteeID {
		return GroupInvitationDTO{}, ErrInvalidRequest
	}
	if s.access != nil {
		ok, err := s.access.CanInviteToGroup(ctx, inviterID, groupID)
		if err != nil {
			return GroupInvitationDTO{}, err
		}
		if !ok {
			return GroupInvitationDTO{}, ErrForbidden
		}
	}
	if _, err := s.userRepo.GetByID(ctx, inviteeID); err != nil {
		return GroupInvitationDTO{}, err
	}
	isMember, err := s.groupRepo.IsMember(ctx, groupID, inviteeID)
	if err != nil {
		return GroupInvitationDTO{}, err
	}
	if isMember {
		return GroupInvitationDTO{}, ErrInvalidRequest
	}
	inv, err := s.groupRepo.CreateInvitation(ctx, groupID, inviterID, inviteeID)
	if err != nil {
		return GroupInvitationDTO{}, err
	}
	if s.notifier != nil {
		_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
			UserID:     inviteeID,
			ActorID:    &inviterID,
			Type:       "group_invitation",
			EntityType: "group_invitation",
			EntityID:   inv.ID,
			Metadata: map[string]any{
				"group_id": groupID,
			},
		})
	}
	return mapInvitation(inv), nil
}

// ListInvitations returns invitations for the actor.
func (s *Service) ListInvitations(ctx context.Context, inviteeID int64, limit, offset int) ([]GroupInvitationDTO, error) {
	invs, err := s.groupRepo.ListInvitationsByInvitee(ctx, inviteeID, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapInvitations(invs), nil
}

// RespondInvitation accepts or declines an invitation.
func (s *Service) RespondInvitation(ctx context.Context, invitationID, actorID int64, status string) error {
	inv, err := s.groupRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return err
	}
	if inv.InviteeID != actorID {
		return ErrForbidden
	}
	switch status {
	case "accepted":
		if err := s.groupRepo.AddMember(ctx, inv.GroupID, actorID); err != nil {
			return err
		}
	case "declined":
	default:
		return ErrInvalidStatus
	}
	return s.groupRepo.DeleteInvitation(ctx, invitationID)
}

// RequestJoin creates a join request.
func (s *Service) RequestJoin(ctx context.Context, groupID, userID int64) (GroupJoinRequestDTO, error) {
	if s.access != nil {
		ok, err := s.access.CanViewGroup(ctx, userID, groupID)
		if err != nil {
			return GroupJoinRequestDTO{}, err
		}
		if ok {
			return GroupJoinRequestDTO{}, ErrInvalidRequest
		}
	}
	req, err := s.groupRepo.CreateJoinRequest(ctx, groupID, userID)
	if err != nil {
		return GroupJoinRequestDTO{}, err
	}
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err == nil && s.notifier != nil {
		creatorID := group.CreatorID
		_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
			UserID:     creatorID,
			ActorID:    &userID,
			Type:       "group_join_request",
			EntityType: "group_join_request",
			EntityID:   req.ID,
			Metadata: map[string]any{
				"group_id": groupID,
			},
		})
	}
	return mapJoinRequest(req), nil
}

// ListJoinRequests returns join requests for a group if actor is creator.
func (s *Service) ListJoinRequests(ctx context.Context, groupID, actorID int64, limit, offset int) ([]GroupJoinRequestDTO, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.CreatorID != actorID {
		return nil, ErrForbidden
	}
	reqs, err := s.groupRepo.ListJoinRequestsByGroup(ctx, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapJoinRequests(reqs), nil
}

// RespondJoinRequest approves or declines a join request.
func (s *Service) RespondJoinRequest(ctx context.Context, groupID, requestID, actorID int64, status string) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatorID != actorID {
		return ErrForbidden
	}
	req, err := s.groupRepo.GetJoinRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.GroupID != groupID {
		return ErrInvalidRequest
	}
	switch status {
	case "accepted":
		if err := s.groupRepo.AddMember(ctx, groupID, req.UserID); err != nil {
			return err
		}
	case "declined":
	default:
		return ErrInvalidStatus
	}
	return s.groupRepo.DeleteJoinRequest(ctx, requestID)
}

// CreateEvent creates a new group event.
func (s *Service) CreateEvent(ctx context.Context, groupID, creatorID int64, req CreateEventRequest) (GroupEventDTO, error) {
	if strings.TrimSpace(req.Title) == "" || req.EventTime == nil {
		return GroupEventDTO{}, ErrInvalidRequest
	}
	if s.access != nil {
		ok, err := s.access.CanPostInGroup(ctx, creatorID, groupID)
		if err != nil {
			return GroupEventDTO{}, err
		}
		if !ok {
			return GroupEventDTO{}, ErrForbidden
		}
	}
	event := domaingroup.GroupEvent{
		GroupID:     groupID,
		CreatorID:   creatorID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		EventTime:   req.EventTime.UTC(),
	}
	created, err := s.groupRepo.CreateEvent(ctx, event)
	if err != nil {
		return GroupEventDTO{}, err
	}
	if s.notifier != nil {
		memberIDs, err := s.groupRepo.GetMemberIDs(ctx, groupID)
		if err == nil {
			for _, memberID := range memberIDs {
				if memberID == creatorID {
					continue
				}
				_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
					UserID:     memberID,
					ActorID:    &creatorID,
					Type:       "event_created",
					EntityType: "event",
					EntityID:   created.ID,
					Metadata: map[string]any{
						"group_id": groupID,
					},
				})
			}
		}
	}
	return mapEvent(created), nil
}

// ListEvents lists events for a group.
func (s *Service) ListEvents(ctx context.Context, groupID, actorID int64, limit, offset int) ([]GroupEventDTO, error) {
	if s.access != nil {
		ok, err := s.access.CanViewGroup(ctx, actorID, groupID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	events, err := s.groupRepo.ListEventsByGroup(ctx, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapEvents(events), nil
}

// RSVP sets a response for a group event.
func (s *Service) RSVP(ctx context.Context, groupID, eventID, actorID int64, response string) error {
	switch response {
	case "going", "not_going":
	default:
		return ErrInvalidResponse
	}
	if s.access != nil {
		ok, err := s.access.CanViewGroup(ctx, actorID, groupID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrForbidden
		}
	}
	ev, err := s.groupRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}
	if ev.GroupID != groupID {
		return ErrInvalidRequest
	}
	return s.groupRepo.UpsertEventResponse(ctx, eventID, actorID, response)
}
