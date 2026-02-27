package post

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/pkg/logger"
)

// Service orchestrates post-related use cases.
type Service struct {
	repo     domainpost.Repository
	userRepo domainuser.Repository
	access   AccessService
	log      logger.Logger
}

// AccessService provides centralized access checks.
type AccessService interface {
	IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error)
	CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error)
	CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error)
	CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error)
	CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error)
}

// NewService builds a post service with the given repository.
func NewService(repo domainpost.Repository, userRepo domainuser.Repository, access AccessService, log logger.Logger) *Service {
	return &Service{
		repo:     repo,
		userRepo: userRepo,
		access:   access,
		log:      log.WithFields(logger.F("service", "post")),
	}
}

// List returns paginated posts as DTOs.
func (s *Service) List(ctx context.Context, viewerID int64, limit, offset int) ([]PostDTO, error) {
	posts, err := s.repo.List(ctx, viewerID, limit, offset)
	if err != nil {
		s.log.Error("failed to list posts", err)
		return nil, fmt.Errorf("list posts: %w", err)
	}
	s.log.Debug("posts listed", logger.F("count", len(posts)))
	return mapPosts(posts), nil
}

// ListGroupsOnly returns only group posts for groups the viewer is a member of.
func (s *Service) ListGroupsOnly(ctx context.Context, viewerID int64, limit, offset int) ([]PostDTO, error) {
	posts, err := s.repo.ListGroupsOnly(ctx, viewerID, limit, offset)
	if err != nil {
		s.log.Error("failed to list group posts", err)
		return nil, fmt.Errorf("list group posts: %w", err)
	}
	s.log.Debug("group posts listed", logger.F("count", len(posts)))
	return mapPosts(posts), nil
}

// Create creates a new post.
func (s *Service) Create(ctx context.Context, authorID int64, req CreatePostRequest) (PostDTO, error) {
	privacy := strings.TrimSpace(req.Privacy)
	if privacy == "" {
		privacy = "public"
	}
	switch privacy {
	case "public", "followers", "private":
	default:
		return PostDTO{}, errors.New("invalid privacy")
	}
	if strings.TrimSpace(req.Content) == "" && (req.MediaPath == nil || strings.TrimSpace(*req.MediaPath) == "") {
		return PostDTO{}, errors.New("content or media_path is required")
	}

	if req.GroupID != nil {
		if *req.GroupID <= 0 {
			return PostDTO{}, errors.New("invalid group_id")
		}
		if s.access == nil {
			return PostDTO{}, errors.New("access service not configured")
		}
		canPost, err := s.access.CanPostInGroup(ctx, authorID, *req.GroupID)
		if err != nil {
			return PostDTO{}, fmt.Errorf("check group access: %w", err)
		}
		if !canPost {
			return PostDTO{}, ErrForbidden
		}
		if len(req.AllowedUserIDs) > 0 {
			return PostDTO{}, errors.New("allowed_user_ids are not allowed for group posts")
		}
		// Normalize group post visibility to public (group access is enforced separately).
		privacy = "public"
	}

	if privacy == "private" {
		if req.GroupID != nil {
			return PostDTO{}, errors.New("private posts are not allowed in groups")
		}
		if len(req.AllowedUserIDs) == 0 {
			return PostDTO{}, errors.New("allowed_user_ids is required for private posts")
		}
		seen := make(map[int64]struct{}, len(req.AllowedUserIDs))
		for _, allowedID := range req.AllowedUserIDs {
			if allowedID <= 0 {
				return PostDTO{}, errors.New("allowed_user_ids must be positive integers")
			}
			if allowedID == authorID {
				return PostDTO{}, errors.New("author cannot be in allowed_user_ids")
			}
			if _, exists := seen[allowedID]; exists {
				continue
			}
			seen[allowedID] = struct{}{}
			if s.access == nil {
				return PostDTO{}, errors.New("access service not configured")
			}
			follows, err := s.access.IsFollowing(ctx, allowedID, authorID)
			if err != nil {
				return PostDTO{}, fmt.Errorf("check allowed users: %w", err)
			}
			if !follows {
				return PostDTO{}, errors.New("allowed_user_ids must be followers of the author")
			}
		}
	} else if len(req.AllowedUserIDs) > 0 {
		// Ignore allowed_user_ids unless privacy is private
		req.AllowedUserIDs = nil
	}

	post := domainpost.Post{
		AuthorID:  authorID,
		GroupID:   req.GroupID,
		Content:   req.Content,
		MediaPath: req.MediaPath,
		Privacy:   privacy,
	}

	created, err := s.repo.Create(ctx, post, req.AllowedUserIDs)
	if err != nil {
		s.log.Error("failed to create post", err, logger.F("author_id", authorID))
		return PostDTO{}, fmt.Errorf("create post: %w", err)
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		s.log.Error("failed to load author", err, logger.F("author_id", authorID))
		return PostDTO{}, fmt.Errorf("get author: %w", err)
	}
	created.AuthorFirstName = author.FirstName
	created.AuthorLastName = author.LastName
	created.AuthorNickname = author.Nickname
	created.AuthorAvatarPath = author.AvatarPath

	return mapPost(created), nil
}

// Update updates a post content/media/visibility. Only the author can update.
func (s *Service) Update(ctx context.Context, postID, authorID int64, req UpdatePostRequest) (PostDTO, error) {
	existing, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			return PostDTO{}, err
		}
		return PostDTO{}, fmt.Errorf("get post: %w", err)
	}
	if existing.AuthorID != authorID {
		return PostDTO{}, ErrForbidden
	}

	content := existing.Content
	if req.Content != nil {
		content = strings.TrimSpace(*req.Content)
	}

	var mediaPath *string
	if req.MediaPath == nil {
		mediaPath = existing.MediaPath
	} else {
		trimmed := strings.TrimSpace(*req.MediaPath)
		if trimmed == "" {
			mediaPath = nil
		} else {
			mediaPath = &trimmed
		}
	}

	privacy := existing.Privacy
	if req.Privacy != nil {
		privacy = strings.TrimSpace(*req.Privacy)
	}

	if strings.TrimSpace(content) == "" && (mediaPath == nil || strings.TrimSpace(*mediaPath) == "") {
		return PostDTO{}, errors.New("content or media_path is required")
	}

	// Group posts: restrict privacy changes and allowed users.
	if existing.GroupID != nil {
		if req.Privacy != nil || len(req.AllowedUserIDs) > 0 {
			return PostDTO{}, errors.New("cannot change privacy or allowed users for group posts")
		}
		privacy = "public"
	}

	switch privacy {
	case "public", "followers", "private":
	default:
		return PostDTO{}, errors.New("invalid privacy")
	}

	allowed := req.AllowedUserIDs
	if privacy == "private" {
		if len(allowed) == 0 {
			return PostDTO{}, errors.New("allowed_user_ids is required for private posts")
		}
		seen := make(map[int64]struct{}, len(allowed))
		for _, allowedID := range allowed {
			if allowedID <= 0 {
				return PostDTO{}, errors.New("allowed_user_ids must be positive integers")
			}
			if allowedID == authorID {
				return PostDTO{}, errors.New("author cannot be in allowed_user_ids")
			}
			if _, exists := seen[allowedID]; exists {
				continue
			}
			seen[allowedID] = struct{}{}
			if s.access == nil {
				return PostDTO{}, errors.New("access service not configured")
			}
			follows, err := s.access.IsFollowing(ctx, allowedID, authorID)
			if err != nil {
				return PostDTO{}, fmt.Errorf("check allowed users: %w", err)
			}
			if !follows {
				return PostDTO{}, errors.New("allowed_user_ids must be followers of the author")
			}
		}
	} else {
		allowed = nil
	}

	updatedPost := domainpost.Post{
		ID:        existing.ID,
		AuthorID:  existing.AuthorID,
		GroupID:   existing.GroupID,
		Content:   content,
		MediaPath: mediaPath,
		Privacy:   privacy,
	}

	updated, err := s.repo.Update(ctx, updatedPost, allowed)
	if err != nil {
		return PostDTO{}, err
	}
	return mapPost(updated), nil
}

// Delete removes a post. Only the author can delete.
func (s *Service) Delete(ctx context.Context, postID, authorID int64) error {
	existing, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if existing.AuthorID != authorID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, postID)
}

// ListByGroup returns posts for a group with pagination.
func (s *Service) ListByGroup(ctx context.Context, groupID, viewerID int64, limit, offset int) ([]PostDTO, error) {
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	canView, err := s.access.CanViewGroup(ctx, viewerID, groupID)
	if err != nil {
		return nil, fmt.Errorf("check group access: %w", err)
	}
	if !canView {
		return nil, ErrForbidden
	}
	posts, err := s.repo.ListByGroup(ctx, groupID, limit, offset)
	if err != nil {
		s.log.Error("failed to list posts by group", err, logger.F("group_id", groupID))
		return nil, fmt.Errorf("list posts by group: %w", err)
	}
	return mapPosts(posts), nil
}

// ListByAuthor returns posts for a specific author with pagination.
func (s *Service) ListByAuthor(ctx context.Context, authorID, viewerID int64, limit, offset int) ([]PostDTO, error) {
	isOwner := viewerID != 0 && viewerID == authorID
	isFollower := false
	if viewerID != 0 && !isOwner {
		if s.access == nil {
			return nil, errors.New("access service not configured")
		}
		var err error
		isFollower, err = s.access.IsFollowing(ctx, viewerID, authorID)
		if err != nil {
			return nil, fmt.Errorf("check follow: %w", err)
		}
	}
	if s.access != nil {
		ok, err := s.access.CanViewProfile(ctx, viewerID, authorID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	} else {
		user, err := s.userRepo.GetByID(ctx, authorID)
		if err != nil {
			return nil, err
		}
		if !user.IsPublic && !isOwner && !isFollower {
			return nil, ErrForbidden
		}
	}

	posts, err := s.repo.ListByAuthor(ctx, authorID, viewerID, isFollower, isOwner, limit, offset)
	if err != nil {
		s.log.Error("failed to list posts by author", err, logger.F("author_id", authorID))
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	s.log.Debug("posts listed by author", logger.F("author_id", authorID), logger.F("count", len(posts)))
	return mapPosts(posts), nil
}

// ErrForbidden is returned when a viewer cannot access a post.
var ErrForbidden = errors.New("post access forbidden")
