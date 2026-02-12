package media

import (
	"context"
	"errors"
	"fmt"

	domainmedia "social-network/backend/internal/domain/media"
)

// AccessService provides access checks for posts and profiles.
type AccessService interface {
	CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error)
	CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error)
}

// ChatAccess provides conversation membership checks.
type ChatAccess interface {
	IsMember(ctx context.Context, conversationID, userID int64) (bool, error)
}

// Service authorizes access to uploaded media.
type Service struct {
	repo   domainmedia.Repository
	access AccessService
	chat   ChatAccess
}

// NewService builds a media access service.
func NewService(repo domainmedia.Repository, access AccessService, chat ChatAccess) *Service {
	return &Service{repo: repo, access: access, chat: chat}
}

// CanAccess checks whether a user can access a media path.
func (s *Service) CanAccess(ctx context.Context, userID int64, path string) (bool, error) {
	ref, err := s.repo.FindByPath(ctx, path)
	if err != nil {
		if errors.Is(err, domainmedia.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	switch ref.Type {
	case domainmedia.MediaTypePost:
		if s.access == nil {
			return false, errors.New("access service not configured")
		}
		ok, err := s.access.CanViewPost(ctx, userID, ref.PostID)
		if err != nil {
			return false, fmt.Errorf("check post access: %w", err)
		}
		return ok, nil
	case domainmedia.MediaTypeComment:
		if s.access == nil {
			return false, errors.New("access service not configured")
		}
		ok, err := s.access.CanViewPost(ctx, userID, ref.PostID)
		if err != nil {
			return false, fmt.Errorf("check comment post access: %w", err)
		}
		return ok, nil
	case domainmedia.MediaTypeMessage:
		if s.chat == nil {
			return false, errors.New("chat access not configured")
		}
		ok, err := s.chat.IsMember(ctx, ref.ConversationID, userID)
		if err != nil {
			return false, fmt.Errorf("check conversation access: %w", err)
		}
		return ok, nil
	case domainmedia.MediaTypeAvatar:
		if s.access == nil {
			return false, errors.New("access service not configured")
		}
		ok, err := s.access.CanViewProfile(ctx, userID, ref.UserID)
		if err != nil {
			return false, fmt.Errorf("check profile access: %w", err)
		}
		return ok, nil
	default:
		return false, nil
	}
}
