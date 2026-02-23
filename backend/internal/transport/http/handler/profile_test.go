package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainuser "social-network/backend/internal/domain/user"
	usecaseprofile "social-network/backend/internal/usecase/profile"
	"social-network/backend/pkg/logger"
)

func TestProfileGet_Unauthorized(t *testing.T) {
	h := NewProfileHandler(nil, nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeProfileUserRepo struct {
	users map[int64]domainuser.User
}

func (r *fakeProfileUserRepo) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return domainuser.User{}, domainuser.ErrNotFound
}
func (r *fakeProfileUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakeProfileUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	return nil
}
func (r *fakeProfileUserRepo) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return 1, nil
}
func (r *fakeProfileUserRepo) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	return 2, nil
}
func (r *fakeProfileUserRepo) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeProfileUserRepo) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeProfileUserRepo) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeProfileUserRepo) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}

type fakeProfileAccess struct{}

func (f fakeProfileAccess) CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error) {
	return true, nil
}
func (f fakeProfileAccess) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return false, nil
}

func TestProfileGet_Success(t *testing.T) {
	user := domainuser.User{
		ID:          1,
		FirstName:   "A",
		LastName:    "B",
		DateOfBirth: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		IsPublic:    true,
	}
	repo := &fakeProfileUserRepo{users: map[int64]domainuser.User{1: user}}
	svc := usecaseprofile.NewService(repo, fakeProfileAccess{})
	h := NewProfileHandler(svc, nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/profiles/1", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.GetProfile), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
