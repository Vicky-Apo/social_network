package handler

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
	domainreaction "social-network/backend/internal/domain/reaction"
	usecasereaction "social-network/backend/internal/usecase/reaction"
	"social-network/backend/pkg/logger"
)

func TestReactionAddPost_Unauthorized(t *testing.T) {
	h := NewReactionHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/posts/1/reactions", nil)
	rr := httptest.NewRecorder()
	h.AddPostReaction(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeReactionRepo struct{}

func (r *fakeReactionRepo) AddPostReaction(ctx context.Context, reaction domainreaction.PostReaction) error {
	return nil
}
func (r *fakeReactionRepo) RemovePostReaction(ctx context.Context, postID, userID int64) error {
	return nil
}
func (r *fakeReactionRepo) GetPostReactions(ctx context.Context, postID int64) ([]domainreaction.PostReaction, error) {
	return nil, nil
}
func (r *fakeReactionRepo) GetPostReaction(ctx context.Context, postID, userID int64) (domainreaction.PostReaction, error) {
	return domainreaction.PostReaction{}, sql.ErrNoRows
}
func (r *fakeReactionRepo) AddCommentReaction(ctx context.Context, reaction domainreaction.CommentReaction) error {
	return nil
}
func (r *fakeReactionRepo) RemoveCommentReaction(ctx context.Context, commentID, userID int64) error {
	return nil
}
func (r *fakeReactionRepo) GetCommentReactions(ctx context.Context, commentID int64) ([]domainreaction.CommentReaction, error) {
	return nil, nil
}
func (r *fakeReactionRepo) GetCommentReaction(ctx context.Context, commentID, userID int64) (domainreaction.CommentReaction, error) {
	return domainreaction.CommentReaction{}, sql.ErrNoRows
}

type fakeReactionPostRepo struct{}

func (r *fakeReactionPostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{ID: id, AuthorID: 2}, nil
}
func (r *fakeReactionPostRepo) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakeReactionPostRepo) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	return domainpost.Post{}, nil
}
func (r *fakeReactionPostRepo) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakeReactionPostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakeReactionPostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, nil
}

type fakeReactionCommentRepo struct{}

func (r *fakeReactionCommentRepo) Create(ctx context.Context, comment domaincomment.Comment) (domaincomment.Comment, error) {
	return domaincomment.Comment{}, nil
}
func (r *fakeReactionCommentRepo) GetByPostID(ctx context.Context, postID int64, limit, offset int) ([]domaincomment.Comment, error) {
	return nil, nil
}
func (r *fakeReactionCommentRepo) GetByID(ctx context.Context, id int64) (domaincomment.Comment, error) {
	return domaincomment.Comment{ID: id, PostID: 1, AuthorID: 2}, nil
}
func (r *fakeReactionCommentRepo) Delete(ctx context.Context, id int64) error { return nil }

type fakeReactionAccess struct{}

func (f fakeReactionAccess) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	return true, nil
}

func TestReactionAddPost_Success(t *testing.T) {
	svc := usecasereaction.NewService(&fakeReactionRepo{}, &fakeReactionPostRepo{}, &fakeReactionCommentRepo{}, fakeReactionAccess{}, nil)
	h := NewReactionHandler(svc, logger.NewDefault(false))

	req := newJSONRequest(t, http.MethodPost, "/posts/1/reactions", map[string]string{"reaction": "like"})
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.AddPostReaction), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
