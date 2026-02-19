package profile

import (
	"context"
	"errors"
	"testing"
	"time"

	domainuser "social-network/backend/internal/domain/user"
)

type fakeUserRepo struct {
	users              map[int64]domainuser.User
	setVisibilityCalls int
	updateProfileCalls int
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
	r.updateProfileCalls++
	u, ok := r.users[id]
	if !ok {
		return domainuser.User{}, domainuser.ErrNotFound
	}
	if nickname != nil {
		u.Nickname = nickname
	}
	if about != nil {
		u.About = about
	}
	if avatarPath != nil {
		u.AvatarPath = avatarPath
	}
	r.users[id] = u
	return u, nil
}

func (r *fakeUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	r.setVisibilityCalls++
	u, ok := r.users[id]
	if !ok {
		return domainuser.ErrNotFound
	}
	u.IsPublic = isPublic
	r.users[id] = u
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

// fake access service

type fakeAccess struct {
	canView      bool
	isFollowing  bool
	isFollowedBy bool
	canViewErr   error
	followErr    error
}

func (f *fakeAccess) CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error) {
	return f.canView, f.canViewErr
}

func (f *fakeAccess) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	if f.followErr != nil {
		return false, f.followErr
	}
	if followerID == 1 && followingID == 2 {
		return f.isFollowing, nil
	}
	if followerID == 2 && followingID == 1 {
		return f.isFollowedBy, nil
	}
	return false, nil
}

func TestGetProfile_LimitedForPrivate(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users[2] = domainuser.User{ID: 2, FirstName: "User", LastName: "Beta", IsPublic: false}

	access := &fakeAccess{canView: false}
	svc := NewService(repo, access)

	profile, err := svc.GetProfile(context.Background(), 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Limited != true {
		t.Fatalf("expected limited profile")
	}
}

func TestGetProfile_FullForOwner(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users[1] = domainuser.User{ID: 1, FirstName: "Jane", LastName: "Doe", DateOfBirth: time.Now().AddDate(-20, 0, 0), IsPublic: false}

	access := &fakeAccess{canView: true}
	svc := NewService(repo, access)

	profile, err := svc.GetProfile(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Limited {
		t.Fatalf("expected full profile")
	}
}

func TestSetVisibility_ForbidNonOwner(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users[1] = domainuser.User{ID: 1}

	svc := NewService(repo, nil)
	if err := svc.SetVisibility(context.Background(), 1, 2, true); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users[1] = domainuser.User{ID: 1, FirstName: "Jane", LastName: "Doe"}
	svc := NewService(repo, nil)

	nickname := "jdoe"
	about := "hello"
	avatar := "/uploads/avatar/x.png"

	updated, err := svc.UpdateProfile(context.Background(), 1, 1, UpdateProfileRequest{
		Nickname:   &nickname,
		About:      &about,
		AvatarPath: &avatar,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Nickname == nil || *updated.Nickname != "jdoe" {
		t.Fatalf("expected nickname updated")
	}
}

func TestUpdateProfile_ForbidNonOwner(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users[1] = domainuser.User{ID: 1}
	svc := NewService(repo, nil)

	_, err := svc.UpdateProfile(context.Background(), 1, 2, UpdateProfileRequest{})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
