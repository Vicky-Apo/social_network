package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domainuser "social-network/backend/internal/domain/user"
	usecaseuser "social-network/backend/internal/usecase/user"
	"social-network/backend/pkg/logger"
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

func TestUserList_Success(t *testing.T) {
	repo := &fakeUserRepo{users: []domainuser.User{{ID: 1, FirstName: "A"}}}
	svc := usecaseuser.NewService(repo)
	h := NewUserHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()
	handler := authWrap(http.HandlerFunc(h.ListUsers), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
