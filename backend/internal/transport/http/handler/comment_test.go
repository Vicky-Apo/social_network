package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domaincomment "social-network/backend/internal/domain/comment"
	usecasecomment "social-network/backend/internal/usecase/comment"
	"social-network/backend/pkg/logger"
)

func TestCommentCreate_Unauthorized(t *testing.T) {
	h := NewCommentHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/posts/1/comments", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeCommentRepo struct{}

func (r *fakeCommentRepo) Create(ctx context.Context, comment domaincomment.Comment) (domaincomment.Comment, error) {
	comment.ID = 1
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = comment.CreatedAt
	return comment, nil
}
func (r *fakeCommentRepo) GetByPostID(ctx context.Context, postID int64, limit, offset int) ([]domaincomment.Comment, error) {
	return []domaincomment.Comment{{ID: 1, PostID: postID, AuthorID: 2, Content: "c"}}, nil
}
func (r *fakeCommentRepo) GetByID(ctx context.Context, id int64) (domaincomment.Comment, error) {
	return domaincomment.Comment{}, nil
}
func (r *fakeCommentRepo) Delete(ctx context.Context, id int64) error { return nil }

func TestCommentGetByPostID_Success(t *testing.T) {
	repo := &fakeCommentRepo{}
	access := &fakeCommentAccess{canView: true}
	svc := usecasecomment.NewService(repo, nil, access, nil)
	h := NewCommentHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/posts/1/comments", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.GetByPostID), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

type fakeCommentAccess struct {
	canView bool
}

func (a *fakeCommentAccess) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	return a.canView, nil
}
