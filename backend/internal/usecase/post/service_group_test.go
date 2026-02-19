package post

import (
	"context"
	"errors"
	"testing"

	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/pkg/logger"
)

type accessStub struct {
	canViewGroup bool
	canPostGroup bool
}

func (a accessStub) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return false, nil
}
func (a accessStub) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	return false, nil
}
func (a accessStub) CanViewProfile(ctx context.Context, viewerID, ownerID int64) (bool, error) {
	return false, nil
}
func (a accessStub) CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return a.canViewGroup, nil
}
func (a accessStub) CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return a.canPostGroup, nil
}

type postRepoStub struct {
	posts []domainpost.Post
}

func (r *postRepoStub) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, errors.New("not implemented")
}
func (r *postRepoStub) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{}, errors.New("not implemented")
}
func (r *postRepoStub) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	post.ID = 1
	return post, nil
}
func (r *postRepoStub) Update(ctx context.Context, post domainpost.Post, allowedUserIDs *[]int64) (domainpost.Post, error) {
	return domainpost.Post{}, errors.New("not implemented")
}
func (r *postRepoStub) Delete(ctx context.Context, id int64) error { return errors.New("not implemented") }
func (r *postRepoStub) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, errors.New("not implemented")
}
func (r *postRepoStub) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return r.posts, nil
}
func (r *postRepoStub) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, errors.New("not implemented")
}

type userRepoStub struct{}

func (r *userRepoStub) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	return domainuser.User{ID: id, FirstName: "Test", LastName: "User"}, nil
}
func (r *userRepoStub) SetVisibility(ctx context.Context, id int64, isPublic bool) error { return errors.New("not implemented") }
func (r *userRepoStub) CountFollowers(ctx context.Context, userID int64) (int64, error) { return 0, errors.New("not implemented") }
func (r *userRepoStub) CountFollowing(ctx context.Context, userID int64) (int64, error) { return 0, errors.New("not implemented") }
func (r *userRepoStub) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) { return nil, errors.New("not implemented") }
func (r *userRepoStub) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) { return nil, errors.New("not implemented") }
func (r *userRepoStub) ListUsers(ctx context.Context) ([]domainuser.User, error) { return nil, errors.New("not implemented") }
func (r *userRepoStub) SearchUsers(ctx context.Context, query string) ([]domainuser.User, error) { return nil, errors.New("not implemented") }

func TestCreateGroupPostForbidden(t *testing.T) {
	service := NewService(&postRepoStub{}, &userRepoStub{}, accessStub{canPostGroup: false}, logger.NewDefault(false))
	_, err := service.CreateGroupPost(context.Background(), 1, 10, CreatePostRequest{Content: "Hello", Privacy: "public"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateGroupPostSuccess(t *testing.T) {
	service := NewService(&postRepoStub{}, &userRepoStub{}, accessStub{canPostGroup: true}, logger.NewDefault(false))
	post, err := service.CreateGroupPost(context.Background(), 1, 10, CreatePostRequest{Content: "Hello", Privacy: "public"})
	if err != nil {
		t.Fatalf("create group post: %v", err)
	}
	if post.GroupID == nil || *post.GroupID != 10 {
		t.Fatalf("expected group_id 10")
	}
}

func TestListByGroupForbidden(t *testing.T) {
	service := NewService(&postRepoStub{}, &userRepoStub{}, accessStub{canViewGroup: false}, logger.NewDefault(false))
	_, err := service.ListByGroup(context.Background(), 10, 2, 20, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListByGroupSuccess(t *testing.T) {
	repo := &postRepoStub{
		posts: []domainpost.Post{
			{ID: 1, GroupID: ptrInt64(10), AuthorID: 1, Content: "Hello", Privacy: "public"},
		},
	}
	service := NewService(repo, &userRepoStub{}, accessStub{canViewGroup: true}, logger.NewDefault(false))
	posts, err := service.ListByGroup(context.Background(), 10, 2, 20, 0)
	if err != nil {
		t.Fatalf("list by group: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
}

func ptrInt64(v int64) *int64 { return &v }
