package profile

import (
	"context"
	"errors"

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
	if s.access == nil {
		return errors.New("access service not configured")
	}
	ok, err := s.access.CanViewProfile(ctx, viewerID, user.ID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
