package comment

import (
	"context"
	"errors"
	"testing"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
)

type fakeCommentRepo struct {
	created domaincomment.Comment
}

func (r *fakeCommentRepo) Create(ctx context.Context, c domaincomment.Comment) (domaincomment.Comment, error) {
	r.created = c
	c.ID = 1
	return c, nil
}

func (r *fakeCommentRepo) GetByPostID(ctx context.Context, postID int64, limit, offset int) ([]domaincomment.Comment, error) {
	return nil, nil
}

func (r *fakeCommentRepo) GetByID(ctx context.Context, id int64) (domaincomment.Comment, error) {
	return domaincomment.Comment{}, domaincomment.ErrNotFound
}

func (r *fakeCommentRepo) Update(ctx context.Context, c domaincomment.Comment) (domaincomment.Comment, error) {
	return c, nil
}

func (r *fakeCommentRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

// fake access

type fakeAccess struct {
	canView bool
}

func (f *fakeAccess) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	return f.canView, nil
}

func TestCreate_ForbiddenWhenNoAccess(t *testing.T) {
	repo := &fakeCommentRepo{}
	postRepo := &fakePostRepo{}
	access := &fakeAccess{canView: false}

	svc := NewService(repo, postRepo, access, nil)
	_, err := svc.Create(context.Background(), CreateCommentRequest{PostID: 1, AuthorID: 2, Content: "hi"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestGetByPostID_ForbiddenWhenNoAccess(t *testing.T) {
	repo := &fakeCommentRepo{}
	postRepo := &fakePostRepo{}
	access := &fakeAccess{canView: false}

	svc := NewService(repo, postRepo, access, nil)
	_, err := svc.GetByPostID(context.Background(), 1, 2, 20, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

// fake post repo

type fakePostRepo struct{}

func (r *fakePostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{ID: id, AuthorID: 1}, nil
}

func (r *fakePostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, nil
}

func (r *fakePostRepo) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) ListGroupsOnly(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) ListPublicOnly(ctx context.Context, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}

func (r *fakePostRepo) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	return domainpost.Post{}, nil
}
func (r *fakePostRepo) Update(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	return domainpost.Post{}, nil
}
func (r *fakePostRepo) Delete(ctx context.Context, id int64) error { return nil }

func (r *fakePostRepo) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}

func (r *fakePostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
