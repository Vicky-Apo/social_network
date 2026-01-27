package comment

import (
	"context"

	domaincomment "social-network/backend/internal/domain/comment"
)

// Service handles comment business logic
type Service struct {
	repo domaincomment.Repository
}

// NewService creates a comment service
func NewService(repo domaincomment.Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new comment
func (s *Service) Create(ctx context.Context, req CreateCommentRequest) (CommentDTO, error) {
	comment := domaincomment.Comment{
		PostID:   req.PostID,
		AuthorID: req.AuthorID,
		Content:  req.Content,
		MediaPath: req.MediaPath,
	}

	created, err := s.repo.Create(ctx, comment)
	if err != nil {
		return CommentDTO{}, err
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
