package profile

import (
	"context"
	"fmt"

	domainuser "social-network/backend/internal/domain/user"
)

// ensureAccess checks whether the viewer can access the target user's profile.
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
