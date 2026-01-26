package profile

import (
	"context"
	"errors"
	"fmt"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
)

// ErrForbidden is returned when the viewer cannot access a profile.
var ErrForbidden = errors.New("profile access forbidden")

// Service orchestrates profile-related use cases.
type Service struct {
	userRepo   domainuser.Repository
	followRepo domainfollow.Repository
}

// NewService builds a profile service with the given repositories.
func NewService(userRepo domainuser.Repository, followRepo domainfollow.Repository) *Service {
	return &Service{
		userRepo:   userRepo,
		followRepo: followRepo,
	}
}

// GetProfile returns a profile with follower stats if the viewer can access it.
func (s *Service) GetProfile(ctx context.Context, profileID, viewerID int64) (ProfileDTO, error) {
	user, err := s.userRepo.GetByID(ctx, profileID)
	if err != nil {
		return ProfileDTO{}, err
	}
	if err := s.ensureAccess(ctx, user, viewerID); err != nil {
		if errors.Is(err, ErrForbidden) {
			return ProfileDTO{
				User:           mapUserLimited(user),
				FollowersCount: nil,
				FollowingCount: nil,
				IsFollowing:    false,
				IsFollowedBy:   false,
				Limited:        true,
			}, nil
		}
		return ProfileDTO{}, err
	}

	// TODO: Add user activity and posts once post timelines are implemented for profiles.
	followers, err := s.userRepo.CountFollowers(ctx, profileID)
	if err != nil {
		return ProfileDTO{}, fmt.Errorf("count followers: %w", err)
	}
	following, err := s.userRepo.CountFollowing(ctx, profileID)
	if err != nil {
		return ProfileDTO{}, fmt.Errorf("count following: %w", err)
	}

	isFollowing := false
	isFollowedBy := false
	if viewerID != 0 && viewerID != user.ID {
		if follow, err := s.followRepo.IsFollowing(ctx, viewerID, user.ID); err != nil {
			return ProfileDTO{}, fmt.Errorf("check follow: %w", err)
		} else {
			isFollowing = follow
		}
		if follow, err := s.followRepo.IsFollowing(ctx, user.ID, viewerID); err != nil {
			return ProfileDTO{}, fmt.Errorf("check follow: %w", err)
		} else {
			isFollowedBy = follow
		}
	}

	return ProfileDTO{
		User:           mapUser(user),
		FollowersCount: &followers,
		FollowingCount: &following,
		IsFollowing:    isFollowing,
		IsFollowedBy:   isFollowedBy,
	}, nil
}

// ListFollowers returns profile data for followers when allowed.
func (s *Service) ListFollowers(ctx context.Context, profileID, viewerID int64) ([]UserDTO, error) {
	user, err := s.userRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureAccess(ctx, user, viewerID); err != nil {
		return nil, err
	}
	followers, err := s.userRepo.ListFollowers(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("list followers: %w", err)
	}
	return mapUsers(followers), nil
}

// ListFollowing returns profile data for followed users when allowed.
func (s *Service) ListFollowing(ctx context.Context, profileID, viewerID int64) ([]UserDTO, error) {
	user, err := s.userRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureAccess(ctx, user, viewerID); err != nil {
		return nil, err
	}
	following, err := s.userRepo.ListFollowing(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("list following: %w", err)
	}
	return mapUsers(following), nil
}

// SetVisibility updates the public/private flag for a user profile.
func (s *Service) SetVisibility(ctx context.Context, profileID, actorID int64, isPublic bool) error {
	if profileID != actorID {
		return ErrForbidden
	}
	if err := s.userRepo.SetVisibility(ctx, profileID, isPublic); err != nil {
		return err
	}
	return nil
}
