package post

import (
	"context"
	"errors"
	"testing"

	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/pkg/logger"
)

type fakePostRepo struct {
	created           domainpost.Post
	listByGroupCalled bool
}

func (r *fakePostRepo) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) Count(ctx context.Context, viewerID int64) (int, error) { return 0, nil }
func (r *fakePostRepo) ListGroupsOnly(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) CountGroupsOnly(ctx context.Context, viewerID int64) (int, error) { return 0, nil }
func (r *fakePostRepo) ListPublicOnly(ctx context.Context, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) CountPublicOnly(ctx context.Context) (int, error) { return 0, nil }
func (r *fakePostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{ID: id, AuthorID: 1, Privacy: "public"}, nil
}
func (r *fakePostRepo) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	r.created = post
	post.ID = 1
	return post, nil
}
func (r *fakePostRepo) Update(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	r.created = post
	return post, nil
}
func (r *fakePostRepo) Delete(ctx context.Context, id int64) error { return nil }
func (r *fakePostRepo) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) CountByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool) (int, error) {
	return 0, nil
}
func (r *fakePostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	r.listByGroupCalled = true
	return nil, nil
}
func (r *fakePostRepo) CountByGroup(ctx context.Context, groupID int64) (int, error) { return 0, nil }
func (r *fakePostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, nil
}

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

func (r *fakeUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error { return nil }
func (r *fakeUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return r.GetByID(ctx, id)
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
func (r *fakeUserRepo) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]domainuser.User, error) { return nil, nil }
func (r *fakeUserRepo) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}

// fake access

type fakeAccess struct {
	canPostInGroup bool
	canViewGroup   bool
	isFollowing    bool
}

func (f *fakeAccess) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return f.isFollowing, nil
}
func (f *fakeAccess) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	return true, nil
}
func (f *fakeAccess) CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error) {
	return true, nil
}
func (f *fakeAccess) CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canPostInGroup, nil
}
func (f *fakeAccess) CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canViewGroup, nil
}

func TestCreate_GroupPostForbidden(t *testing.T) {
	repo := &fakePostRepo{}
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1}

	svc := NewService(repo, userRepo, &fakeAccess{canPostInGroup: false}, logger.NewDefault(false))

	groupID := int64(10)
	_, err := svc.Create(context.Background(), 1, CreatePostRequest{GroupID: &groupID, Content: "hello", Privacy: "public"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestCreate_PrivateRequiresAllowedUsers(t *testing.T) {
	repo := &fakePostRepo{}
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1}

	svc := NewService(repo, userRepo, &fakeAccess{isFollowing: true}, logger.NewDefault(false))
	_, err := svc.Create(context.Background(), 1, CreatePostRequest{Content: "hello", Privacy: "private"})
	if err == nil {
		t.Fatalf("expected error for missing allowed_user_ids")
	}
}

func TestListByGroup_Forbidden(t *testing.T) {
	repo := &fakePostRepo{}
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1}

	svc := NewService(repo, userRepo, &fakeAccess{canViewGroup: false}, logger.NewDefault(false))
	_, err := svc.ListByGroup(context.Background(), 1, 1, 20, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestListByGroup_AllowsMembers(t *testing.T) {
	repo := &fakePostRepo{}
	userRepo := newFakeUserRepo()
	userRepo.users[1] = domainuser.User{ID: 1}

	svc := NewService(repo, userRepo, &fakeAccess{canViewGroup: true}, logger.NewDefault(false))
	_, err := svc.ListByGroup(context.Background(), 1, 1, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.listByGroupCalled {
		t.Fatalf("expected list by group called")
	}
}
