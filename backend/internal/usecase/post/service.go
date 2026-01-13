package post

import (
	"context"

	domainpost "social-network/backend/internal/domain/post"
)

// Service orchestrates post-related use cases.
type Service struct {
	repo domainpost.Repository
}

// NewService builds a post service with the given repository.
func NewService(repo domainpost.Repository) *Service {
	return &Service{repo: repo}
}

// List returns all posts as DTOs.
func (s *Service) List(ctx context.Context) ([]PostDTO, error) {
	posts, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return mapPosts(posts), nil
}

// GetByID returns a single post as a DTO.
func (s *Service) GetByID(ctx context.Context, id int64) (PostDTO, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return PostDTO{}, err
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
