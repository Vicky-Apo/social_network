package profile

import (
	"context"
	"errors"
	"fmt"

	domainuser "social-network/backend/internal/domain/user"
)

// ErrForbidden is returned when the viewer cannot access a profile.
var ErrForbidden = errors.New("profile access forbidden")

// Service orchestrates profile-related use cases.
type Service struct {
	userRepo domainuser.Repository
	access   AccessService
}

// AccessService provides centralized access checks.
type AccessService interface {
	CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error)
	IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error)
}

// NewService builds a profile service with the given repositories.
func NewService(userRepo domainuser.Repository, access AccessService) *Service {
	return &Service{
		userRepo: userRepo,
		access:   access,
	}
}

// GetProfile returns a profile with follower stats if the viewer can access it.
func (s *Service) GetProfile(ctx context.Context, profileID, viewerID int64) (ProfileDTO, error) {
	user, err := s.userRepo.GetByID(ctx, profileID)
	if err != nil {
		return ProfileDTO{}, err
	}
	accessSnapshot, err := s.buildAccessSnapshot(ctx, user, viewerID)
	if err != nil {
		return ProfileDTO{}, err
	}
	if !accessSnapshot.CanView {
		return ProfileDTO{
			User:           mapProfileUserLimited(user),
			FollowersCount: nil,
			FollowingCount: nil,
			IsFollowing:    false,
			IsFollowedBy:   false,
			Limited:        true,
		}, nil
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

	return ProfileDTO{
		User:           mapProfileUser(user),
		FollowersCount: &followers,
		FollowingCount: &following,
		IsFollowing:    accessSnapshot.IsFollower,
		IsFollowedBy:   accessSnapshot.IsFollowedBy,
	}, nil
}

type accessSnapshot struct {
	CanView      bool
	IsOwner      bool
	IsFollower   bool
	IsFollowedBy bool
	IsPublic     bool
}

func (s *Service) buildAccessSnapshot(ctx context.Context, user domainuser.User, viewerID int64) (accessSnapshot, error) {
	snapshot := accessSnapshot{
		IsPublic: user.IsPublic,
		IsOwner:  viewerID != 0 && viewerID == user.ID,
	}
	if snapshot.IsOwner || user.IsPublic {
		snapshot.CanView = true
	}
	if viewerID == 0 {
		return snapshot, nil
	}
	if s.access == nil {
		return accessSnapshot{}, errors.New("access service not configured")
	}
	if !snapshot.IsOwner {
		isFollower, err := s.access.IsFollowing(ctx, viewerID, user.ID)
		if err != nil {
			return accessSnapshot{}, fmt.Errorf("check follow: %w", err)
		}
		snapshot.IsFollower = isFollower
		if !snapshot.CanView && isFollower {
			snapshot.CanView = true
		}
		isFollowedBy, err := s.access.IsFollowing(ctx, user.ID, viewerID)
		if err != nil {
			return accessSnapshot{}, fmt.Errorf("check follow back: %w", err)
		}
		snapshot.IsFollowedBy = isFollowedBy
	}
	return snapshot, nil
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

// UpdateProfile updates nickname/about/avatar for the profile owner.
func (s *Service) UpdateProfile(ctx context.Context, profileID, actorID int64, req UpdateProfileRequest) (UserDTO, error) {
	if profileID != actorID {
		return UserDTO{}, ErrForbidden
	}
	updated, err := s.userRepo.UpdateProfile(ctx, profileID, req.Nickname, req.About, req.AvatarPath)
	if err != nil {
		return UserDTO{}, err
	}
	return mapUser(updated), nil
}
