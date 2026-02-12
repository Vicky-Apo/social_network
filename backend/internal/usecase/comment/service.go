package comment

import (
	"context"
	"errors"

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
	if s.access != nil {
		ok, err := s.access.CanViewPost(ctx, req.AuthorID, req.PostID)
		if err != nil {
			return CommentDTO{}, err
		}
		if !ok {
			return CommentDTO{}, ErrForbidden
		}
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
	if s.notifier != nil && s.postRepo != nil {
		post, err := s.postRepo.GetByID(ctx, created.PostID)
		if err == nil && post.AuthorID != created.AuthorID {
			_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
				UserID:     post.AuthorID,
				ActorID:    &created.AuthorID,
				Type:       "comment_on_post",
				EntityType: "post",
				EntityID:   post.ID,
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
	if s.access != nil {
		ok, err := s.access.CanViewPost(ctx, viewerID, postID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	comments, err := s.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, err
	}

	return mapComments(comments), nil
}

// ErrForbidden is returned when a viewer cannot access comments.
var ErrForbidden = errors.New("comment access forbidden")
