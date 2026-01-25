package post

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainfollow "social-network/backend/internal/domain/follow"
	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/pkg/logger"
)

// Service orchestrates post-related use cases.
type Service struct {
	repo       domainpost.Repository
	userRepo   domainuser.Repository
	followRepo domainfollow.Repository
	log        logger.Logger
}

// NewService builds a post service with the given repository.
func NewService(repo domainpost.Repository, userRepo domainuser.Repository, followRepo domainfollow.Repository, log logger.Logger) *Service {
	return &Service{
		repo:       repo,
		userRepo:   userRepo,
		followRepo: followRepo,
		log:        log.WithFields(logger.F("service", "post")),
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
	if ok, err := s.canViewPost(ctx, post, viewerID); err != nil {
		return PostDTO{}, err
	} else if !ok {
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

	post := domainpost.Post{
		AuthorID:  authorID,
		Content:   req.Content,
		MediaPath: req.MediaPath,
		Privacy:   privacy,
	}

	created, err := s.repo.Create(ctx, post, req.CategoryIDs)
	if err != nil {
		s.log.Error("failed to create post", err, logger.F("author_id", authorID))
		return PostDTO{}, fmt.Errorf("create post: %w", err)
	}

	return mapPost(created), nil
}

// ListByAuthor returns posts for a specific author with pagination.
func (s *Service) ListByAuthor(ctx context.Context, authorID, viewerID int64, limit, offset int) ([]PostDTO, error) {
	user, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}
	isOwner := viewerID != 0 && viewerID == authorID
	isFollower := false
	if viewerID != 0 && !isOwner {
		isFollower, err = s.followRepo.IsFollowing(ctx, viewerID, authorID)
		if err != nil {
			return nil, fmt.Errorf("check follow: %w", err)
		}
	}
	if !user.IsPublic && !isOwner && !isFollower {
		return nil, ErrForbidden
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

func mapPosts(posts []domainpost.Post) []PostDTO {
	out := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, mapPost(p))
	}
	return out
}

func mapPost(p domainpost.Post) PostDTO {
	return PostDTO{
		ID:        p.ID,
		AuthorID:  p.AuthorID,
		Content:   p.Content,
		MediaPath: p.MediaPath,
		Privacy:   p.Privacy,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		CommentCount: p.CommentCount,
		LikeCount:    p.LikeCount,
		DislikeCount: p.DislikeCount,
	}
}

// ErrForbidden is returned when a viewer cannot access a post.
var ErrForbidden = errors.New("post access forbidden")

func (s *Service) canViewPost(ctx context.Context, post domainpost.Post, viewerID int64) (bool, error) {
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil {
		return false, err
	}
	if viewerID == 0 {
		return author.IsPublic && post.Privacy == "public", nil
	}
	if viewerID == post.AuthorID {
		return true, nil
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
	case "public":
		return true, nil
	case "followers":
		follows, err := s.followRepo.IsFollowing(ctx, viewerID, post.AuthorID)
		if err != nil {
			return false, fmt.Errorf("check follow: %w", err)
		}
		return follows, nil
	case "private":
		allowed, err := s.repo.IsUserAllowed(ctx, post.ID, viewerID)
		if err != nil {
			return false, fmt.Errorf("check allowed users: %w", err)
		}
		return allowed, nil
	default:
		return false, nil
	}
}
