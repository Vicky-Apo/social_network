package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
	usecasefollow "social-network/backend/internal/usecase/follow"
	"social-network/backend/pkg/logger"
)

func TestFollowCreateRequest_Unauthorized(t *testing.T) {
	h := NewFollowHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/follow-requests", nil)
	rr := httptest.NewRecorder()
	h.CreateRequest(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeFollowUserRepo struct {
	users map[int64]domainuser.User
}

func (r *fakeFollowUserRepo) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domainuser.User{}, domainuser.ErrNotFound
	}
	return u, nil
}
func (r *fakeFollowUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakeFollowUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	return nil
}
func (r *fakeFollowUserRepo) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakeFollowUserRepo) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakeFollowUserRepo) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeFollowUserRepo) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeFollowUserRepo) ListUsers(ctx context.Context) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeFollowUserRepo) SearchUsers(ctx context.Context, query string) ([]domainuser.User, error) {
	return nil, nil
}

type fakeFollowRepo struct{}

func (r *fakeFollowRepo) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return false, nil
}
func (r *fakeFollowRepo) RequestExists(ctx context.Context, requesterID, targetID int64) (bool, error) {
	return false, nil
}
func (r *fakeFollowRepo) CreateRequest(ctx context.Context, requesterID, targetID int64) (domainfollow.FollowRequest, error) {
	return domainfollow.FollowRequest{ID: 1, RequesterID: requesterID, TargetID: targetID, Status: "pending", CreatedAt: time.Now()}, nil
}
func (r *fakeFollowRepo) GetRequestByID(ctx context.Context, id int64) (domainfollow.FollowRequest, error) {
	return domainfollow.FollowRequest{}, nil
}
func (r *fakeFollowRepo) UpdateRequestStatus(ctx context.Context, id int64, status string) error {
	return nil
}
func (r *fakeFollowRepo) ListRequestsByTarget(ctx context.Context, targetID int64) ([]domainfollow.FollowRequest, error) {
	return nil, nil
}
func (r *fakeFollowRepo) ListRequestsByRequester(ctx context.Context, requesterID int64) ([]domainfollow.FollowRequest, error) {
	return nil, nil
}
func (r *fakeFollowRepo) CreateFollow(ctx context.Context, followerID, followingID int64) error {
	return nil
}
func (r *fakeFollowRepo) DeleteFollow(ctx context.Context, followerID, followingID int64) error {
	return nil
}
func (r *fakeFollowRepo) GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error) {
	return nil, nil
}

func TestFollowCreateRequest_Success(t *testing.T) {
	userRepo := &fakeFollowUserRepo{
		users: map[int64]domainuser.User{
			1: {ID: 1, IsPublic: false},
			2: {ID: 2, IsPublic: true},
		},
	}
	followRepo := &fakeFollowRepo{}
	svc := usecasefollow.NewService(userRepo, followRepo, nil)
	h := NewFollowHandler(svc, logger.NewDefault(false))

	req := newJSONRequest(t, http.MethodPost, "/follow-requests", map[string]int64{"target_id": 2})
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.CreateRequest), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
