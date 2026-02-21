package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	usecasepost "social-network/backend/internal/usecase/post"
	"social-network/backend/pkg/logger"
)

func TestPostCreate_Unauthorized(t *testing.T) {
	h := NewPostHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/posts", nil)
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakePostRepo struct{}

func (r *fakePostRepo) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return []domainpost.Post{{ID: 1, AuthorID: 2}}, nil
}
func (r *fakePostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{}, nil
}
func (r *fakePostRepo) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	return domainpost.Post{}, nil
}
func (r *fakePostRepo) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, nil
}

type fakePostUserRepo struct{}

func (r *fakePostUserRepo) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakePostUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakePostUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	return nil
}
func (r *fakePostUserRepo) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakePostUserRepo) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakePostUserRepo) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakePostUserRepo) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakePostUserRepo) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakePostUserRepo) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}

func TestPostList_Success(t *testing.T) {
	repo := &fakePostRepo{}
	userRepo := &fakePostUserRepo{}
	svc := usecasepost.NewService(repo, userRepo, nil, logger.NewDefault(false))
	h := NewPostHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
