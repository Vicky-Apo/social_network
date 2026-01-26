package follow

import (
	"context"
	"errors"
	"fmt"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
)

// ErrAlreadyFollowing is returned when a follow already exists.
var ErrAlreadyFollowing = errors.New("already following")

// ErrRequestExists is returned when a follow request already exists.
var ErrRequestExists = errors.New("follow request already exists")

// ErrCannotFollowSelf is returned when a user tries to follow themselves.
var ErrCannotFollowSelf = errors.New("cannot follow self")

// ErrNotFollowing is returned when trying to unfollow without an existing follow.
var ErrNotFollowing = errors.New("not following")

// ErrForbidden is returned when the action is not allowed.
var ErrForbidden = errors.New("follow action forbidden")

// ErrRequestNotPending is returned when a follow request is not pending.
var ErrRequestNotPending = errors.New("follow request is not pending")

// ErrInvalidStatus is returned when a follow request status is invalid.
var ErrInvalidStatus = errors.New("invalid status")

// Service orchestrates follow-related use cases.
type Service struct {
	userRepo   domainuser.Repository
	followRepo domainfollow.Repository
}

// NewService builds a follow service with the given repositories.
func NewService(userRepo domainuser.Repository, followRepo domainfollow.Repository) *Service {
	return &Service{
		userRepo:   userRepo,
		followRepo: followRepo,
	}
}

// RequestFollow requests to follow a target user or follows immediately if public.
func (s *Service) RequestFollow(ctx context.Context, requesterID, targetID int64) (FollowResultDTO, error) {
	if requesterID == targetID {
		return FollowResultDTO{}, ErrCannotFollowSelf
	}
	if _, err := s.userRepo.GetByID(ctx, requesterID); err != nil {
		return FollowResultDTO{}, err
	}
	targetUser, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return FollowResultDTO{}, err
	}
	alreadyFollowing, err := s.followRepo.IsFollowing(ctx, requesterID, targetID)
	if err != nil {
		return FollowResultDTO{}, fmt.Errorf("check follow: %w", err)
	}
	if alreadyFollowing {
		return FollowResultDTO{}, ErrAlreadyFollowing
	}
	requestExists, err := s.followRepo.RequestExists(ctx, requesterID, targetID)
	if err != nil {
		return FollowResultDTO{}, fmt.Errorf("check request: %w", err)
	}
	if requestExists {
		return FollowResultDTO{}, ErrRequestExists
	}
	if targetUser.IsPublic {
		if err := s.followRepo.CreateFollow(ctx, requesterID, targetID); err != nil {
			return FollowResultDTO{}, fmt.Errorf("create follow: %w", err)
		}
		return FollowResultDTO{Status: "followed"}, nil
	}
	req, err := s.followRepo.CreateRequest(ctx, requesterID, targetID)
	if err != nil {
		return FollowResultDTO{}, fmt.Errorf("create request: %w", err)
	}
	return FollowResultDTO{
		Status:  "requested",
		Request: mapRequest(req),
	}, nil
}

// AcceptRequest accepts a follow request and creates a follow.
func (s *Service) AcceptRequest(ctx context.Context, requestID, actorID int64) error {
	req, err := s.followRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.TargetID != actorID {
		return ErrForbidden
	}
	if req.Status != "pending" {
		return ErrRequestNotPending
	}
	alreadyFollowing, err := s.followRepo.IsFollowing(ctx, req.RequesterID, req.TargetID)
	if err != nil {
		return fmt.Errorf("check follow: %w", err)
	}
	if !alreadyFollowing {
		if err := s.followRepo.CreateFollow(ctx, req.RequesterID, req.TargetID); err != nil {
			return fmt.Errorf("create follow: %w", err)
		}
	}
	if err := s.followRepo.UpdateRequestStatus(ctx, requestID, "accepted"); err != nil {
		return fmt.Errorf("update request status: %w", err)
	}
	return nil
}

// DeclineRequest declines a follow request.
func (s *Service) DeclineRequest(ctx context.Context, requestID, actorID int64) error {
	req, err := s.followRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.TargetID != actorID {
		return ErrForbidden
	}
	if req.Status != "pending" {
		return ErrRequestNotPending
	}
	if err := s.followRepo.UpdateRequestStatus(ctx, requestID, "declined"); err != nil {
		return fmt.Errorf("update request status: %w", err)
	}
	return nil
}

// UpdateRequest updates a follow request status.
func (s *Service) UpdateRequest(ctx context.Context, requestID, actorID int64, status string) error {
	switch status {
	case "accepted":
		return s.AcceptRequest(ctx, requestID, actorID)
	case "declined":
		return s.DeclineRequest(ctx, requestID, actorID)
	default:
		return ErrInvalidStatus
	}
}

// ListRequests lists pending follow requests for a target user.
func (s *Service) ListRequests(ctx context.Context, targetID int64) ([]FollowRequestDTO, error) {
	requests, err := s.followRepo.ListRequestsByTarget(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	out := make([]FollowRequestDTO, 0, len(requests))
	for _, req := range requests {
		if dto := mapRequest(req); dto != nil {
			out = append(out, *dto)
		}
	}
	return out, nil
}

// ListSentRequests lists follow requests created by a requester.
func (s *Service) ListSentRequests(ctx context.Context, requesterID int64) ([]FollowRequestDTO, error) {
	requests, err := s.followRepo.ListRequestsByRequester(ctx, requesterID)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	out := make([]FollowRequestDTO, 0, len(requests))
	for _, req := range requests {
		if dto := mapRequest(req); dto != nil {
			out = append(out, *dto)
		}
	}
	return out, nil
}

// Unfollow removes a follow relationship.
func (s *Service) Unfollow(ctx context.Context, followerID, followingID int64) error {
	if followerID == followingID {
		return ErrCannotFollowSelf
	}
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return fmt.Errorf("check follow: %w", err)
	}
	if !isFollowing {
		return ErrNotFollowing
	}
	if err := s.followRepo.DeleteFollow(ctx, followerID, followingID); err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}
	return nil
}

func mapRequest(req domainfollow.FollowRequest) *FollowRequestDTO {
	return &FollowRequestDTO{
		ID:          req.ID,
		RequesterID: req.RequesterID,
		TargetID:    req.TargetID,
		Status:      req.Status,
		CreatedAt:   req.CreatedAt,
	}
}
