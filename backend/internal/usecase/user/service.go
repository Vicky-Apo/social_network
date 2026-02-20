package user

import (
	"context"
	"fmt"

	domainuser "social-network/backend/internal/domain/user"
)

// Service orchestrates user listing/searching.
type Service struct {
	repo domainuser.Repository
}

// NewService builds a user service with the given repository.
func NewService(repo domainuser.Repository) *Service {
	return &Service{repo: repo}
}

// ListUsers returns users as lightweight DTOs with pagination and access filtering.
func (s *Service) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]UserListItemDTO, error) {
	users, err := s.repo.ListUsers(ctx, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return mapUsers(users), nil
}

// SearchUsers searches users by first name, last name, or nickname.
func (s *Service) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]UserListItemDTO, error) {
	users, err := s.repo.SearchUsers(ctx, viewerID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	return mapUsers(users), nil
}

func mapUsers(users []domainuser.User) []UserListItemDTO {
	out := make([]UserListItemDTO, 0, len(users))
	for _, u := range users {
		out = append(out, UserListItemDTO{
			ID:         u.ID,
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			Nickname:   u.Nickname,
			AvatarPath: u.AvatarPath,
		})
	}
	return out
}
