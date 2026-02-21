package user

import (
	"context"
	"errors"
	"testing"

	domainuser "social-network/backend/internal/domain/user"
)

type fakeUserRepo struct {
	users []domainuser.User
	err   error
}

func (r *fakeUserRepo) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]domainuser.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.users, nil
}
func (r *fakeUserRepo) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]domainuser.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.users, nil
}

// unused interface methods
func (r *fakeUserRepo) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakeUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakeUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	return nil
}
func (r *fakeUserRepo) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakeUserRepo) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakeUserRepo) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}

func TestListUsers_MapsDTO(t *testing.T) {
	repo := &fakeUserRepo{
		users: []domainuser.User{
			{ID: 1, FirstName: "A", LastName: "One"},
			{ID: 2, FirstName: "B", LastName: "Two"},
		},
	}
	svc := NewService(repo)

	items, err := svc.ListUsers(context.Background(), 1, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].ID != 1 || items[1].ID != 2 {
		t.Fatalf("unexpected items")
	}
}

func TestSearchUsers_MapsDTO(t *testing.T) {
	repo := &fakeUserRepo{
		users: []domainuser.User{
			{ID: 3, FirstName: "C", LastName: "Three"},
		},
	}
	svc := NewService(repo)

	items, err := svc.SearchUsers(context.Background(), 1, "c", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 3 {
		t.Fatalf("unexpected items")
	}
}

func TestUserService_PropagatesRepoError(t *testing.T) {
	repo := &fakeUserRepo{err: errors.New("boom")}
	svc := NewService(repo)

	if _, err := svc.ListUsers(context.Background(), 1, 20, 0); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := svc.SearchUsers(context.Background(), 1, "x", 20, 0); err == nil {
		t.Fatalf("expected error")
	}
}
