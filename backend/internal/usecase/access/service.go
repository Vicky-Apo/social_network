package access

import (
	"context"
	"errors"
	"fmt"

	domainfollow "social-network/backend/internal/domain/follow"
	domaingroup "social-network/backend/internal/domain/group"
	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/pkg/logger"
)

// ErrAccessDenied indicates a policy denial.
var ErrAccessDenied = errors.New("access denied")

// Service centralizes access/relationship rules.
type Service struct {
	userRepo   domainuser.Repository
	followRepo domainfollow.Repository
	postRepo   domainpost.Repository
	groupRepo  domaingroup.Repository
	log        logger.Logger
}

// NewService builds an access service.
func NewService(
	userRepo domainuser.Repository,
	followRepo domainfollow.Repository,
	postRepo domainpost.Repository,
	groupRepo domaingroup.Repository,
	log logger.Logger,
) *Service {
	return &Service{
		userRepo:   userRepo,
		followRepo: followRepo,
		postRepo:   postRepo,
		groupRepo:  groupRepo,
		log:        log.WithFields(logger.F("service", "access")),
	}
}

// IsFollowing returns true if followerID follows followingID.
func (s *Service) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return s.followRepo.IsFollowing(ctx, followerID, followingID)
}

// CanViewProfile checks if viewer can access owner profile.
func (s *Service) CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error) {
	if viewerID == ownerID {
		return true, nil
	}
	user, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return false, err
	}
	if user.IsPublic {
		return true, nil
	}
	if viewerID == 0 {
		return false, nil
	}
	follows, err := s.followRepo.IsFollowing(ctx, viewerID, ownerID)
	if err != nil {
		return false, fmt.Errorf("check follow: %w", err)
	}
	return follows, nil
}

// CanViewPost checks if viewer can access a post.
func (s *Service) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return false, err
	}
	if post.GroupID != nil {
		if viewerID == 0 {
			return false, nil
		}
		if viewerID == post.AuthorID {
			return true, nil
		}
		return s.groupRepo.IsMember(ctx, *post.GroupID, viewerID)
	}
	if viewerID == post.AuthorID {
		return true, nil
	}
	if post.Privacy == "public" {
		return true, nil
	}
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil {
		return false, err
	}
	if viewerID == 0 {
		return author.IsPublic && post.Privacy == "public", nil
	}
	if !author.IsPublic {
		follows, err := s.followRepo.IsFollowing(ctx, viewerID, post.AuthorID)
		if err != nil {
			return false, fmt.Errorf("check follow: %w", err)
		}
		if !follows {
			return false, nil
		}
	}
	switch post.Privacy {
	case "followers":
		follows, err := s.followRepo.IsFollowing(ctx, viewerID, post.AuthorID)
		if err != nil {
			return false, fmt.Errorf("check follow: %w", err)
		}
		return follows, nil
	case "private":
		allowed, err := s.postRepo.IsUserAllowed(ctx, post.ID, viewerID)
		if err != nil {
			return false, fmt.Errorf("check allowed users: %w", err)
		}
		return allowed, nil
	default:
		return false, nil
	}
}

// CanSendDirectMessage returns true if at least one follows the other.
func (s *Service) CanSendDirectMessage(ctx context.Context, senderID, receiverID int64) (bool, error) {
	follows, err := s.followRepo.IsFollowing(ctx, senderID, receiverID)
	if err != nil {
		return false, fmt.Errorf("check sender follows: %w", err)
	}
	if follows {
		return true, nil
	}
	reverse, err := s.followRepo.IsFollowing(ctx, receiverID, senderID)
	if err != nil {
		return false, fmt.Errorf("check receiver follows: %w", err)
	}
	return reverse, nil
}

// CanViewGroup returns true if user is a member of the group.
func (s *Service) CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}

// CanPostInGroup returns true if user is a member of the group.
func (s *Service) CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}

// CanChatInGroup returns true if user is a member of the group.
func (s *Service) CanChatInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}

// CanInviteToGroup returns true if user is a member of the group.
func (s *Service) CanInviteToGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}

// CanApproveGroupJoin returns true if user is the group creator.
func (s *Service) CanApproveGroupJoin(ctx context.Context, userID, groupID int64) (bool, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return false, err
	}
	return group.CreatorID == userID, nil
}
