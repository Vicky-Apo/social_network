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
		return ProfileDTO{}, err
	}

	followers, err := s.userRepo.CountFollowers(ctx, profileID)
	if err != nil {
		return ProfileDTO{}, fmt.Errorf("count followers: %w", err)
	}
	following, err := s.userRepo.CountFollowing(ctx, profileID)
	if err != nil {
		return ProfileDTO{}, fmt.Errorf("count following: %w", err)
	}

	return ProfileDTO{
		User:           mapUser(user),
		FollowersCount: followers,
		FollowingCount: following,
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
func (s *Service) SetVisibility(ctx context.Context, profileID int64, isPublic bool) error {
	if err := s.userRepo.SetVisibility(ctx, profileID, isPublic); err != nil {
		return err
	}
	return nil
}

func (s *Service) ensureAccess(ctx context.Context, user domainuser.User, viewerID int64) error {
	if user.IsPublic || viewerID == user.ID {
		return nil
	}
	if viewerID == 0 {
		return ErrForbidden
	}
	follows, err := s.followRepo.IsFollowing(ctx, viewerID, user.ID)
	if err != nil {
		return fmt.Errorf("check follow: %w", err)
	}
	if !follows {
		return ErrForbidden
	}
	return nil
}

func mapUsers(users []domainuser.User) []UserDTO {
	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, mapUser(u))
	}
	return out
}

func mapUser(u domainuser.User) UserDTO {
	return UserDTO{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DateOfBirth: u.DateOfBirth,
		AvatarPath:  u.AvatarPath,
		Nickname:    u.Nickname,
		About:       u.About,
		IsPublic:    u.IsPublic,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
