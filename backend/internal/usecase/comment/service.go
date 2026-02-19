package comment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

// Service handles comment business logic
type Service struct {
	repo     domaincomment.Repository
	postRepo domainpost.Repository
	userRepo domainuser.Repository
	notifier Notifier
}

// Notifier allows emitting notifications without coupling to transport details.
type Notifier interface {
	CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error)
}

// NewService creates a comment service
func NewService(repo domaincomment.Repository, postRepo domainpost.Repository, userRepo domainuser.Repository, notifier Notifier) *Service {
	return &Service{repo: repo, postRepo: postRepo, userRepo: userRepo, notifier: notifier}
}

// GetByID returns a comment by ID (no author enrichment).
func (s *Service) GetByID(ctx context.Context, commentID int64) (domaincomment.Comment, error) {
	return s.repo.GetByID(ctx, commentID)
}

// Create creates a new comment
func (s *Service) Create(ctx context.Context, req CreateCommentRequest) (CommentDTO, error) {
	if strings.TrimSpace(req.Content) == "" && strings.TrimSpace(req.MediaPath) == "" {
		return CommentDTO{}, ErrInvalidRequest
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
	if s.userRepo != nil {
		if author, err := s.userRepo.GetByID(ctx, req.AuthorID); err == nil {
			created.AuthorFirstName = author.FirstName
			created.AuthorLastName = author.LastName
			created.AuthorNickname = author.Nickname
			created.AuthorAvatarPath = author.AvatarPath
		}
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
func (s *Service) GetByPostID(ctx context.Context, postID int64, limit, offset int) ([]CommentDTO, error) {
	comments, err := s.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, err
	}

	return mapComments(comments), nil
}

// Update updates an existing comment.
func (s *Service) Update(ctx context.Context, commentID, actorID int64, req UpdateCommentRequest) (CommentDTO, error) {
	if req.Content == nil && req.MediaPath == nil {
		return CommentDTO{}, ErrInvalidRequest
	}

	comment, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return CommentDTO{}, err
	}
	if comment.AuthorID != actorID {
		return CommentDTO{}, ErrForbidden
	}

	newContent := comment.Content
	if req.Content != nil {
		newContent = *req.Content
	}
	newMediaPath := comment.MediaPath
	if req.MediaPath != nil {
		newMediaPath = *req.MediaPath
	}

	if strings.TrimSpace(newContent) == "" && strings.TrimSpace(newMediaPath) == "" {
		return CommentDTO{}, fmt.Errorf("%w: content or media_path is required", ErrInvalidRequest)
	}

	comment.Content = newContent
	comment.MediaPath = newMediaPath

	updated, err := s.repo.Update(ctx, comment)
	if err != nil {
		return CommentDTO{}, err
	}
	if s.userRepo != nil {
		if author, err := s.userRepo.GetByID(ctx, updated.AuthorID); err == nil {
			updated.AuthorFirstName = author.FirstName
			updated.AuthorLastName = author.LastName
			updated.AuthorNickname = author.Nickname
			updated.AuthorAvatarPath = author.AvatarPath
		}
	}

	return mapComment(updated), nil
}

// Delete removes a comment if the actor is the author.
func (s *Service) Delete(ctx context.Context, commentID, actorID int64) error {
	comment, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment.AuthorID != actorID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, commentID)
}

// ErrForbidden is returned when a user cannot modify a comment.
var ErrForbidden = errors.New("comment access forbidden")

// ErrInvalidRequest is returned when an update request is invalid.
var ErrInvalidRequest = errors.New("invalid comment request")
