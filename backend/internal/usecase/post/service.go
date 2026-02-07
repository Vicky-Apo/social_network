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

// GetByID returns a single post as a DTO.
func (s *Service) GetByID(ctx context.Context, id int64, viewerID int64) (PostDTO, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			s.log.Debug("post not found", logger.F("post_id", id))
			return PostDTO{}, err
		}
		s.log.Error("failed to get post", err, logger.F("post_id", id))
		return PostDTO{}, fmt.Errorf("get post: %w", err)
	}
	if s.access == nil {
		return PostDTO{}, errors.New("access service not configured")
	}
	ok, err := s.access.CanViewPost(ctx, viewerID, id)
	if err != nil {
		return PostDTO{}, err
	}
	if !ok {
		return PostDTO{}, ErrForbidden
	}
	return mapPost(post), nil
}

// Create creates a new post with optional categories.
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

	if privacy == "private" {
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
		Content:   req.Content,
		MediaPath: req.MediaPath,
		Privacy:   privacy,
	}

	created, err := s.repo.Create(ctx, post, req.CategoryIDs, req.AllowedUserIDs)
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
		if !ok && !isOwner && !isFollower {
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

// ListByCategory returns posts for a category with pagination.
func (s *Service) ListByCategory(ctx context.Context, categoryID, viewerID int64, limit, offset int) ([]PostDTO, error) {
	posts, err := s.repo.ListByCategory(ctx, categoryID, viewerID, limit, offset)
	if err != nil {
		s.log.Error("failed to list posts by category", err, logger.F("category_id", categoryID))
		return nil, fmt.Errorf("list posts by category: %w", err)
	}
	s.log.Debug("posts listed by category", logger.F("category_id", categoryID), logger.F("count", len(posts)))
	return mapPosts(posts), nil
}

// ErrForbidden is returned when a viewer cannot access a post.
var ErrForbidden = errors.New("post access forbidden")
