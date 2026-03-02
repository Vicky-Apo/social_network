package comment

import (
	"context"
	"errors"
	"strings"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

// Service handles comment business logic
type Service struct {
	repo     domaincomment.Repository
	postRepo domainpost.Repository
	access   AccessService
	notifier Notifier
}

// AccessService provides centralized access checks.
type AccessService interface {
	CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error)
}

// Notifier allows emitting notifications without coupling to transport details.
type Notifier interface {
	CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error)
}

// NewService creates a comment service
func NewService(repo domaincomment.Repository, postRepo domainpost.Repository, access AccessService, notifier Notifier) *Service {
	return &Service{repo: repo, postRepo: postRepo, access: access, notifier: notifier}
}

// Create creates a new comment
func (s *Service) Create(ctx context.Context, req CreateCommentRequest) (CommentDTO, error) {
	if s.access == nil {
		return CommentDTO{}, errors.New("access service not configured")
	}
	if strings.TrimSpace(req.Content) == "" && strings.TrimSpace(req.MediaPath) == "" {
		return CommentDTO{}, errors.New("content or media is required")
	}
	ok, err := s.access.CanViewPost(ctx, req.AuthorID, req.PostID)
	if err != nil {
		return CommentDTO{}, err
	}
	if !ok {
		return CommentDTO{}, ErrForbidden
	}
	comment := domaincomment.Comment{
		PostID:    req.PostID,
		AuthorID:  req.AuthorID,
		Content:   req.Content,
		MediaPath: req.MediaPath,
	}

	created, err := s.repo.Create(ctx, comment)
	if err != nil {
		return CommentDTO{}, err
	}

	// Emit notification to post author
	if s.notifier != nil {
		post, err := s.postRepo.GetByID(ctx, req.PostID)
		if err == nil && post.AuthorID != req.AuthorID {
			_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
				UserID:     post.AuthorID,
				ActorID:    &req.AuthorID,
				Type:       "comment_on_post",
				EntityType: "post",
				EntityID:   req.PostID,
				Metadata: map[string]any{
					"comment_id": created.ID,
				},
			})
		}
	}

	return mapComment(created), nil
}

// GetByPostID gets all comments for a post
func (s *Service) GetByPostID(ctx context.Context, postID, viewerID int64, limit, offset int) ([]CommentDTO, error) {
	if s.access == nil {
		return nil, errors.New("access service not configured")
	}
	ok, err := s.access.CanViewPost(ctx, viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	comments, err := s.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, err
	}

	return mapComments(comments), nil
}

// Update updates an existing comment. Only the author can update.
func (s *Service) Update(ctx context.Context, commentID, authorID int64, req UpdateCommentRequest) (CommentDTO, error) {
	existing, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return CommentDTO{}, err
	}
	if existing.AuthorID != authorID {
		return CommentDTO{}, ErrForbidden
	}

	content := existing.Content
	if req.Content != nil {
		content = strings.TrimSpace(*req.Content)
	}
	mediaPath := existing.MediaPath
	if req.MediaPath != nil {
		trimmed := strings.TrimSpace(*req.MediaPath)
		mediaPath = trimmed
	}

	if strings.TrimSpace(content) == "" && strings.TrimSpace(mediaPath) == "" {
		return CommentDTO{}, errors.New("content or media is required")
	}

	updated, err := s.repo.Update(ctx, domaincomment.Comment{
		ID:        existing.ID,
		PostID:    existing.PostID,
		AuthorID:  existing.AuthorID,
		Content:   content,
		MediaPath: mediaPath,
	})
	if err != nil {
		return CommentDTO{}, err
	}
	return mapComment(updated), nil
}

// Delete removes a comment. Only the author can delete.
func (s *Service) Delete(ctx context.Context, commentID, authorID int64) error {
	existing, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if existing.AuthorID != authorID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, commentID)
}

// ErrForbidden is returned when a viewer cannot access comments.
var ErrForbidden = errors.New("comment access forbidden")
