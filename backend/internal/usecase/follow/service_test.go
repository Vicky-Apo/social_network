package follow

import (
	"context"
	"errors"
	"testing"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

type fakeUserRepo struct {
	users map[int64]domainuser.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[int64]domainuser.User)}
}

func (r *fakeUserRepo) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domainuser.User{}, domainuser.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return r.GetByID(ctx, id)
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
func (r *fakeUserRepo) ListUsers(ctx context.Context) ([]domainuser.User, error) { return nil, nil }
func (r *fakeUserRepo) SearchUsers(ctx context.Context, query string) ([]domainuser.User, error) {
	return nil, nil
}

// fake follow repo

type fakeFollowRepo struct {
	following map[[2]int64]bool
	requests  map[[2]int64]domainfollow.FollowRequest
	nextReqID int64
}

func newFakeFollowRepo() *fakeFollowRepo {
	return &fakeFollowRepo{following: make(map[[2]int64]bool), requests: make(map[[2]int64]domainfollow.FollowRequest), nextReqID: 1}
}

func (r *fakeFollowRepo) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return r.following[[2]int64{followerID, followingID}], nil
}

func (r *fakeFollowRepo) RequestExists(ctx context.Context, requesterID, targetID int64) (bool, error) {
	_, ok := r.requests[[2]int64{requesterID, targetID}]
	return ok, nil
}

func (r *fakeFollowRepo) CreateRequest(ctx context.Context, requesterID, targetID int64) (domainfollow.FollowRequest, error) {
	id := r.nextReqID
	r.nextReqID++
	req := domainfollow.FollowRequest{ID: id, RequesterID: requesterID, TargetID: targetID, Status: "pending"}
	r.requests[[2]int64{requesterID, targetID}] = req
	return req, nil
}

func (r *fakeFollowRepo) GetRequestByID(ctx context.Context, id int64) (domainfollow.FollowRequest, error) {
	for _, req := range r.requests {
		if req.ID == id {
			return req, nil
		}
	}
	return domainfollow.FollowRequest{}, domainfollow.ErrRequestNotFound
}

func (r *fakeFollowRepo) UpdateRequestStatus(ctx context.Context, id int64, status string) error {
	for key, req := range r.requests {
		if req.ID == id {
			req.Status = status
			r.requests[key] = req
			return nil
		}
	}
	return domainfollow.ErrRequestNotFound
}

func (r *fakeFollowRepo) ListRequestsByTarget(ctx context.Context, targetID int64) ([]domainfollow.FollowRequest, error) {
	return nil, nil
}

func (r *fakeFollowRepo) ListRequestsByRequester(ctx context.Context, requesterID int64) ([]domainfollow.FollowRequest, error) {
	return nil, nil
}

func (r *fakeFollowRepo) CreateFollow(ctx context.Context, followerID, followingID int64) error {
	r.following[[2]int64{followerID, followingID}] = true
	return nil
}

func (r *fakeFollowRepo) DeleteFollow(ctx context.Context, followerID, followingID int64) error {
	delete(r.following, [2]int64{followerID, followingID})
	return nil
}

func (r *fakeFollowRepo) GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error) {
	return nil, nil
}

// fake notifier

type testNotifier struct {
	calls      int
	lastUserID int64
}

func (n *testNotifier) CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error) {
	n.calls++
	n.lastUserID = req.UserID
	return usecasenotification.NotificationDTO{}, nil
}

func TestRequestFollow_PublicUserAutoFollow(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1, IsPublic: true}
	userRepo.users[2] = domainuser.User{ID: 2, IsPublic: true}

	followRepo := newFakeFollowRepo()
	notify := &testNotifier{}

	svc := NewService(userRepo, followRepo, notify)

	res, err := svc.RequestFollow(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "followed" {
		t.Fatalf("expected followed, got %s", res.Status)
	}
	if !followRepo.following[[2]int64{1, 2}] {
		t.Fatalf("expected follow created")
	}
}

func TestRequestFollow_PrivateUserCreatesRequest(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1, IsPublic: true}
	userRepo.users[2] = domainuser.User{ID: 2, IsPublic: false}

	followRepo := newFakeFollowRepo()
	notify := &testNotifier{}

	svc := NewService(userRepo, followRepo, notify)

	res, err := svc.RequestFollow(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "requested" {
		t.Fatalf("expected requested, got %s", res.Status)
	}
	if notify.calls != 1 {
		t.Fatalf("expected notification")
	}
}

func TestUpdateRequest_AcceptCreatesFollow(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1, IsPublic: true}
	userRepo.users[2] = domainuser.User{ID: 2, IsPublic: false}

	followRepo := newFakeFollowRepo()
	req, _ := followRepo.CreateRequest(context.Background(), 1, 2)

	svc := NewService(userRepo, followRepo, nil)

	err := svc.UpdateRequest(context.Background(), req.ID, 2, "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !followRepo.following[[2]int64{1, 2}] {
		t.Fatalf("expected follow created")
	}
}

func TestUpdateRequest_Forbidden(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1}
	userRepo.users[2] = domainuser.User{ID: 2}

	followRepo := newFakeFollowRepo()
	req, _ := followRepo.CreateRequest(context.Background(), 1, 2)

	svc := NewService(userRepo, followRepo, nil)

	err := svc.UpdateRequest(context.Background(), req.ID, 1, "accepted")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
