package post

import (
	"context"
	"errors"
	"fmt"

	domainpost "social-network/backend/internal/domain/post"
	"social-network/backend/pkg/logger"
)

// Service orchestrates post-related use cases.
type Service struct {
	repo domainpost.Repository
	log  logger.Logger
}

// NewService builds a post service with the given repository.
func NewService(repo domainpost.Repository, log logger.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log.WithFields(logger.F("service", "post")),
	}
}

// List returns all posts as DTOs.
func (s *Service) List(ctx context.Context) ([]PostDTO, error) {
	posts, err := s.repo.List(ctx)
	if err != nil {
		s.log.Error("failed to list posts", err)
		return nil, fmt.Errorf("list posts: %w", err)
	}
	s.log.Debug("posts listed", logger.F("count", len(posts)))
	return mapPosts(posts), nil
}

// GetByID returns a single post as a DTO.
func (s *Service) GetByID(ctx context.Context, id int64) (PostDTO, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			s.log.Debug("post not found", logger.F("post_id", id))
			return PostDTO{}, err
		}
		s.log.Error("failed to get post", err, logger.F("post_id", id))
		return PostDTO{}, fmt.Errorf("get post: %w", err)
	}
	return mapPost(post), nil
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
		Privacy:   p.Privacy,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
