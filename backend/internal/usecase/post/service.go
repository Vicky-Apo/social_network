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
	CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error)
	CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error)
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
		return PostDTO{}, ErrInvalidRequest
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

// CreateGroupPost creates a new group post if the author is a member.
func (s *Service) CreateGroupPost(ctx context.Context, authorID, groupID int64, req CreatePostRequest) (PostDTO, error) {
	if s.access == nil {
		return PostDTO{}, errors.New("access service not configured")
	}
	ok, err := s.access.CanPostInGroup(ctx, authorID, groupID)
	if err != nil {
		return PostDTO{}, err
	}
	if !ok {
		return PostDTO{}, ErrForbidden
	}
	if strings.TrimSpace(req.Content) == "" && (req.MediaPath == nil || strings.TrimSpace(*req.MediaPath) == "") {
		return PostDTO{}, errors.New("content or media_path is required")
	}

	post := domainpost.Post{
		AuthorID:  authorID,
		GroupID:   &groupID,
		Content:   req.Content,
		MediaPath: req.MediaPath,
		Privacy:   "public",
	}
	created, err := s.repo.Create(ctx, post, nil)
	if err != nil {
		s.log.Error("failed to create group post", err, logger.F("author_id", authorID), logger.F("group_id", groupID))
		return PostDTO{}, fmt.Errorf("create group post: %w", err)
	}
	if s.userRepo != nil {
		if author, err := s.userRepo.GetByID(ctx, authorID); err == nil {
			created.AuthorFirstName = author.FirstName
			created.AuthorLastName = author.LastName
			created.AuthorNickname = author.Nickname
			created.AuthorAvatarPath = author.AvatarPath
		}
	}
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

// ListByGroup returns posts for a group if viewer is a member.
func (s *Service) ListByGroup(ctx context.Context, groupID, viewerID int64, limit, offset int) ([]PostDTO, error) {
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	ok, err := s.access.CanViewGroup(ctx, viewerID, groupID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	posts, err := s.repo.ListByGroup(ctx, groupID, limit, offset)
	if err != nil {
		s.log.Error("failed to list posts by group", err, logger.F("group_id", groupID))
		return nil, fmt.Errorf("list posts by group: %w", err)
	}
	s.log.Debug("posts listed by group", logger.F("group_id", groupID), logger.F("count", len(posts)))
	return mapPosts(posts), nil
}


// ErrForbidden is returned when a viewer cannot access a post.
var ErrForbidden = errors.New("post access forbidden")

// ErrInvalidRequest is returned when a request is invalid.
var ErrInvalidRequest = errors.New("invalid post request")

// Update updates a post if the actor is the author.
func (s *Service) Update(ctx context.Context, postID, actorID int64, req UpdatePostRequest) (PostDTO, error) {
	if req.Content == nil && req.MediaPath == nil && req.Privacy == nil && req.AllowedUserIDs == nil {
		return PostDTO{}, ErrInvalidRequest
	}

	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			return PostDTO{}, err
		}
		s.log.Error("failed to load post for update", err, logger.F("post_id", postID))
		return PostDTO{}, fmt.Errorf("get post: %w", err)
	}
	if post.AuthorID != actorID {
		return PostDTO{}, ErrForbidden
	}

	newContent := post.Content
	if req.Content != nil {
		newContent = *req.Content
	}
	newMediaPath := post.MediaPath
	if req.MediaPath != nil {
		trimmed := strings.TrimSpace(*req.MediaPath)
		if trimmed == "" {
			newMediaPath = nil
		} else {
			newMediaPath = req.MediaPath
		}
	}

	newPrivacy := post.Privacy
	if req.Privacy != nil {
		privacy := strings.TrimSpace(*req.Privacy)
		if privacy == "" {
			return PostDTO{}, ErrInvalidRequest
		}
		switch privacy {
		case "public", "followers", "private":
			newPrivacy = privacy
		default:
			return PostDTO{}, ErrInvalidRequest
		}
	}

	if strings.TrimSpace(newContent) == "" && (newMediaPath == nil || strings.TrimSpace(*newMediaPath) == "") {
		return PostDTO{}, fmt.Errorf("%w: content or media_path is required", ErrInvalidRequest)
	}

	var allowedUserIDsUpdate *[]int64
	if newPrivacy == "private" {
		if req.AllowedUserIDs != nil {
			if len(*req.AllowedUserIDs) == 0 {
				return PostDTO{}, ErrInvalidRequest
			}
			seen := make(map[int64]struct{}, len(*req.AllowedUserIDs))
			deduped := make([]int64, 0, len(*req.AllowedUserIDs))
			for _, allowedID := range *req.AllowedUserIDs {
				if allowedID <= 0 {
					return PostDTO{}, ErrInvalidRequest
				}
				if allowedID == actorID {
					return PostDTO{}, ErrInvalidRequest
				}
				if _, exists := seen[allowedID]; exists {
					continue
				}
				seen[allowedID] = struct{}{}
				if s.access == nil {
					return PostDTO{}, errors.New("access service not configured")
				}
				follows, err := s.access.IsFollowing(ctx, allowedID, actorID)
				if err != nil {
					return PostDTO{}, fmt.Errorf("check allowed users: %w", err)
				}
				if !follows {
					return PostDTO{}, ErrInvalidRequest
				}
				deduped = append(deduped, allowedID)
			}
			allowedUserIDsUpdate = &deduped
		} else if post.Privacy != "private" {
			return PostDTO{}, ErrInvalidRequest
		}
	} else if post.Privacy == "private" || req.AllowedUserIDs != nil {
		empty := []int64{}
		allowedUserIDsUpdate = &empty
	}

	post.Content = newContent
	post.MediaPath = newMediaPath
	post.Privacy = newPrivacy

	updated, err := s.repo.Update(ctx, post, allowedUserIDsUpdate)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			return PostDTO{}, err
		}
		s.log.Error("failed to update post", err, logger.F("post_id", postID))
		return PostDTO{}, fmt.Errorf("update post: %w", err)
	}

	return mapPost(updated), nil
}

// Delete removes a post if the actor is the author.
func (s *Service) Delete(ctx context.Context, postID, actorID int64) error {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.AuthorID != actorID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, postID)
}
