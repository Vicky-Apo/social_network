package group

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domaingroup "social-network/backend/internal/domain/group"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

// Service errors.
var (
	ErrInvalidTitle       = errors.New("invalid group title")
	ErrForbidden          = errors.New("group action forbidden")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrCannotInviteSelf   = errors.New("cannot invite self")
	ErrAlreadyMember      = errors.New("already a member")
	ErrInvitationExists   = errors.New("invitation already exists")
	ErrJoinRequestExists  = errors.New("join request already exists")
	ErrCannotLeaveCreator = errors.New("group creator cannot leave")
)

// AccessService provides centralized access checks.
type AccessService interface {
	CanInviteToGroup(ctx context.Context, userID, groupID int64) (bool, error)
	CanApproveGroupJoin(ctx context.Context, userID, groupID int64) (bool, error)
	CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error)
}

// Service orchestrates group-related use cases.
type Service struct {
	groupRepo domaingroup.Repository
	access    AccessService
	notifier  Notifier
}

// NewService builds a group service with the given repositories.
func NewService(groupRepo domaingroup.Repository, access AccessService, notifier Notifier) *Service {
	return &Service{groupRepo: groupRepo, access: access, notifier: notifier}
}

// Notifier allows emitting notifications without coupling to transport details.
type Notifier interface {
	CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error)
}

// CreateGroup creates a new group and enrolls the creator.
func (s *Service) CreateGroup(ctx context.Context, creatorID int64, req CreateGroupRequest) (GroupDTO, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return GroupDTO{}, ErrInvalidTitle
	}
	group, err := s.groupRepo.Create(ctx, creatorID, title, req.Description)
	if err != nil {
		return GroupDTO{}, fmt.Errorf("create group: %w", err)
	}
	return mapGroupSummary(domaingroup.GroupSummary{Group: group, MemberCount: 1, IsMember: true}), nil
}

// ListGroups lists groups with pagination and optional search.
func (s *Service) ListGroups(ctx context.Context, userID int64, query string, limit, offset int) ([]GroupDTO, error) {
	var (
		items []domaingroup.GroupSummary
		err   error
	)
	if strings.TrimSpace(query) == "" {
		items, err = s.groupRepo.List(ctx, userID, limit, offset)
	} else {
		items, err = s.groupRepo.Search(ctx, userID, query, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}

	out := make([]GroupDTO, 0, len(items))
	for _, g := range items {
		out = append(out, mapGroupSummary(g))
	}
	return out, nil
}

// GetGroup returns a group by ID.
func (s *Service) GetGroup(ctx context.Context, userID, groupID int64) (GroupDTO, error) {
	group, err := s.groupRepo.GetWithMeta(ctx, userID, groupID)
	if err != nil {
		return GroupDTO{}, err
	}
	return mapGroupSummary(group), nil
}

// ListMembers returns member info for a group (members only).
func (s *Service) ListMembers(ctx context.Context, userID, groupID int64) ([]GroupMemberDTO, error) {
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	canView, err := s.access.CanViewGroup(ctx, userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("check access: %w", err)
	}
	if !canView {
		return nil, ErrForbidden
	}

	members, err := s.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	out := make([]GroupMemberDTO, 0, len(members))
	for _, m := range members {
		out = append(out, GroupMemberDTO{
			UserID:     m.UserID,
			FirstName:  m.FirstName,
			LastName:   m.LastName,
			Nickname:   m.Nickname,
			AvatarPath: m.AvatarPath,
			JoinedAt:   m.JoinedAt,
		})
	}
	return out, nil
}

// InviteToGroup sends an invitation to a user.
func (s *Service) InviteToGroup(ctx context.Context, inviterID, groupID, inviteeID int64) (GroupInvitationDTO, error) {
	if inviterID == inviteeID {
		return GroupInvitationDTO{}, ErrCannotInviteSelf
	}
	if s.access == nil {
		return GroupInvitationDTO{}, errors.New("access service not configured")
	}
	allowed, err := s.access.CanInviteToGroup(ctx, inviterID, groupID)
	if err != nil {
		return GroupInvitationDTO{}, fmt.Errorf("check access: %w", err)
	}
	if !allowed {
		return GroupInvitationDTO{}, ErrForbidden
	}

	isMember, err := s.groupRepo.IsMember(ctx, groupID, inviteeID)
	if err != nil {
		return GroupInvitationDTO{}, fmt.Errorf("check member: %w", err)
	}
	if isMember {
		return GroupInvitationDTO{}, ErrAlreadyMember
	}

	exists, err := s.groupRepo.InvitationExists(ctx, groupID, inviteeID)
	if err != nil {
		return GroupInvitationDTO{}, fmt.Errorf("check invitation: %w", err)
	}
	if exists {
		return GroupInvitationDTO{}, ErrInvitationExists
	}

	inv, err := s.groupRepo.CreateInvitation(ctx, groupID, inviterID, inviteeID)
	if err != nil {
		return GroupInvitationDTO{}, fmt.Errorf("create invitation: %w", err)
	}
	if s.notifier != nil {
		_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
			UserID:     inviteeID,
			ActorID:    &inviterID,
			Type:       "group_invitation",
			EntityType: "group_invitation",
			EntityID:   inv.ID,
			Metadata: map[string]any{
				"group_id":   groupID,
				"inviter_id": inviterID,
			},
		})
	}
	return mapInvitation(inv), nil
}

// ListInvitations lists invitations for the user.
func (s *Service) ListInvitations(ctx context.Context, userID int64) ([]GroupInvitationDTO, error) {
	items, err := s.groupRepo.ListInvitationsByInvitee(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	out := make([]GroupInvitationDTO, 0, len(items))
	for _, inv := range items {
		out = append(out, mapInvitation(inv))
	}
	return out, nil
}

// UpdateInvitation accepts or declines an invitation.
func (s *Service) UpdateInvitation(ctx context.Context, invitationID, actorID int64, status string) error {
	inv, err := s.groupRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return err
	}
	if inv.InviteeID != actorID {
		return ErrForbidden
	}

	switch status {
	case "accepted":
		isMember, err := s.groupRepo.IsMember(ctx, inv.GroupID, actorID)
		if err != nil {
			return fmt.Errorf("check member: %w", err)
		}
		if !isMember {
			if err := s.groupRepo.AddMember(ctx, inv.GroupID, actorID); err != nil {
				return fmt.Errorf("add member: %w", err)
			}
		}
		return s.groupRepo.DeleteInvitation(ctx, invitationID)
	case "declined":
		return s.groupRepo.DeleteInvitation(ctx, invitationID)
	default:
		return ErrInvalidStatus
	}
}

// RequestJoin creates a join request for a group.
func (s *Service) RequestJoin(ctx context.Context, groupID, userID int64) (GroupJoinRequestDTO, error) {
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return GroupJoinRequestDTO{}, fmt.Errorf("check member: %w", err)
	}
	if isMember {
		return GroupJoinRequestDTO{}, ErrAlreadyMember
	}

	exists, err := s.groupRepo.JoinRequestExists(ctx, groupID, userID)
	if err != nil {
		return GroupJoinRequestDTO{}, fmt.Errorf("check join request: %w", err)
	}
	if exists {
		return GroupJoinRequestDTO{}, ErrJoinRequestExists
	}

	req, err := s.groupRepo.CreateJoinRequest(ctx, groupID, userID)
	if err != nil {
		return GroupJoinRequestDTO{}, fmt.Errorf("create join request: %w", err)
	}
	if s.notifier != nil {
		if group, err := s.groupRepo.GetByID(ctx, groupID); err == nil {
			creatorID := group.CreatorID
			if creatorID != userID {
				_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
					UserID:     creatorID,
					ActorID:    &userID,
					Type:       "group_join_request",
					EntityType: "group_join_request",
					EntityID:   req.ID,
					Metadata: map[string]any{
						"group_id": groupID,
						"user_id":  userID,
					},
				})
			}
		}
	}
	return mapJoinRequest(req), nil
}

// ListJoinRequests lists join requests for a group (creator only).
func (s *Service) ListJoinRequests(ctx context.Context, groupID, actorID int64) ([]GroupJoinRequestDTO, error) {
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	allowed, err := s.access.CanApproveGroupJoin(ctx, actorID, groupID)
	if err != nil {
		return nil, fmt.Errorf("check access: %w", err)
	}
	if !allowed {
		return nil, ErrForbidden
	}

	items, err := s.groupRepo.ListJoinRequestsByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	out := make([]GroupJoinRequestDTO, 0, len(items))
	for _, req := range items {
		out = append(out, mapJoinRequest(req))
	}
	return out, nil
}

// UpdateJoinRequest accepts or declines a join request.
func (s *Service) UpdateJoinRequest(ctx context.Context, requestID, actorID int64, status string) error {
	req, err := s.groupRepo.GetJoinRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if s.access == nil {
		return errors.New("access service not configured")
	}
	allowed, err := s.access.CanApproveGroupJoin(ctx, actorID, req.GroupID)
	if err != nil {
		return fmt.Errorf("check access: %w", err)
	}
	if !allowed {
		return ErrForbidden
	}

	switch status {
	case "accepted":
		isMember, err := s.groupRepo.IsMember(ctx, req.GroupID, req.UserID)
		if err != nil {
			return fmt.Errorf("check member: %w", err)
		}
		if !isMember {
			if err := s.groupRepo.AddMember(ctx, req.GroupID, req.UserID); err != nil {
				return fmt.Errorf("add member: %w", err)
			}
		}
		return s.groupRepo.DeleteJoinRequest(ctx, requestID)
	case "declined":
		return s.groupRepo.DeleteJoinRequest(ctx, requestID)
	default:
		return ErrInvalidStatus
	}
}

// LeaveGroup removes a user from a group.
func (s *Service) LeaveGroup(ctx context.Context, groupID, userID int64) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatorID == userID {
		return ErrCannotLeaveCreator
	}
	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

func mapGroupSummary(g domaingroup.GroupSummary) GroupDTO {
	return GroupDTO{
		ID:          g.ID,
		CreatorID:   g.CreatorID,
		Title:       g.Title,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		MemberCount: g.MemberCount,
		IsMember:    g.IsMember,
	}
}

func mapInvitation(inv domaingroup.GroupInvitation) GroupInvitationDTO {
	return GroupInvitationDTO{
		ID:        inv.ID,
		GroupID:   inv.GroupID,
		InviterID: inv.InviterID,
		InviteeID: inv.InviteeID,
		CreatedAt: inv.CreatedAt,
	}
}

func mapJoinRequest(req domaingroup.GroupJoinRequest) GroupJoinRequestDTO {
	return GroupJoinRequestDTO{
		ID:        req.ID,
		GroupID:   req.GroupID,
		UserID:    req.UserID,
		CreatedAt: req.CreatedAt,
	}
}
