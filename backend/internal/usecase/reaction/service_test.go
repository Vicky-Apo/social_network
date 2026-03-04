package reaction

import (
	"context"
	"database/sql"
	"testing"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
	domainreaction "social-network/backend/internal/domain/reaction"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

type fakeReactionRepo struct {
	postReactions    map[[2]int64]domainreaction.PostReaction
	commentReactions map[[2]int64]domainreaction.CommentReaction
}

func newFakeReactionRepo() *fakeReactionRepo {
	return &fakeReactionRepo{
		postReactions:    make(map[[2]int64]domainreaction.PostReaction),
		commentReactions: make(map[[2]int64]domainreaction.CommentReaction),
	}
}

func (r *fakeReactionRepo) GetPostReaction(ctx context.Context, postID, userID int64) (domainreaction.PostReaction, error) {
	if v, ok := r.postReactions[[2]int64{postID, userID}]; ok {
		return v, nil
	}
	return domainreaction.PostReaction{}, sqlErrNoRows
}

func (r *fakeReactionRepo) AddPostReaction(ctx context.Context, reaction domainreaction.PostReaction) error {
	r.postReactions[[2]int64{reaction.PostID, reaction.UserID}] = reaction
	return nil
}

func (r *fakeReactionRepo) RemovePostReaction(ctx context.Context, postID, userID int64) error {
	delete(r.postReactions, [2]int64{postID, userID})
	return nil
}

func (r *fakeReactionRepo) GetPostReactions(ctx context.Context, postID int64) ([]domainreaction.PostReaction, error) {
	out := make([]domainreaction.PostReaction, 0)
	for key, val := range r.postReactions {
		if key[0] == postID {
			out = append(out, val)
		}
	}
	return out, nil
}

func (r *fakeReactionRepo) GetCommentReaction(ctx context.Context, commentID, userID int64) (domainreaction.CommentReaction, error) {
	if v, ok := r.commentReactions[[2]int64{commentID, userID}]; ok {
		return v, nil
	}
	return domainreaction.CommentReaction{}, sqlErrNoRows
}

func (r *fakeReactionRepo) AddCommentReaction(ctx context.Context, reaction domainreaction.CommentReaction) error {
	r.commentReactions[[2]int64{reaction.CommentID, reaction.UserID}] = reaction
	return nil
}

func (r *fakeReactionRepo) RemoveCommentReaction(ctx context.Context, commentID, userID int64) error {
	delete(r.commentReactions, [2]int64{commentID, userID})
	return nil
}

func (r *fakeReactionRepo) GetCommentReactions(ctx context.Context, commentID int64) ([]domainreaction.CommentReaction, error) {
	out := make([]domainreaction.CommentReaction, 0)
	for key, val := range r.commentReactions {
		if key[0] == commentID {
			out = append(out, val)
		}
	}
	return out, nil
}

// fake post repo

type fakePostRepo struct{ authorID int64 }

func (r *fakePostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{ID: id, AuthorID: r.authorID}, nil
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
func (r *fakePostRepo) CountByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool) (int, error) {
	return 0, nil
}
func (r *fakePostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) CountByGroup(ctx context.Context, groupID int64) (int, error) { return 0, nil }
func (r *fakePostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, nil
}

// fake comment repo

type fakeCommentRepo struct{ authorID int64 }

func (r *fakeCommentRepo) GetByID(ctx context.Context, id int64) (domaincomment.Comment, error) {
	return domaincomment.Comment{ID: id, AuthorID: r.authorID}, nil
}

func (r *fakeCommentRepo) Create(ctx context.Context, comment domaincomment.Comment) (domaincomment.Comment, error) {
	return domaincomment.Comment{}, nil
}
func (r *fakeCommentRepo) GetByPostID(ctx context.Context, postID int64, limit, offset int) ([]domaincomment.Comment, error) {
	return nil, nil
}
func (r *fakeCommentRepo) CountByPostID(ctx context.Context, postID int64) (int, error) { return 0, nil }
func (r *fakeCommentRepo) Delete(ctx context.Context, id int64) error { return nil }
func (r *fakeCommentRepo) Update(ctx context.Context, comment domaincomment.Comment) (domaincomment.Comment, error) {
	return domaincomment.Comment{}, nil
}

// fake notifier

type testNotifier struct {
	calls        int
	lastType     string
	lastUserID   int64
	lastMetadata map[string]any
}

type fakeAccess struct{}

func (f fakeAccess) CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error) {
	return true, nil
}

func (n *testNotifier) CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error) {
	n.calls++
	n.lastType = req.Type
	n.lastUserID = req.UserID
	n.lastMetadata = req.Metadata
	return usecasenotification.NotificationDTO{}, nil
}

var sqlErrNoRows = sql.ErrNoRows

func TestAddPostReaction_Adds(t *testing.T) {
	repo := newFakeReactionRepo()
	postRepo := &fakePostRepo{authorID: 2}
	commentRepo := &fakeCommentRepo{authorID: 3}
	notify := &testNotifier{}

	svc := NewService(repo, postRepo, commentRepo, fakeAccess{}, notify)
	status, err := svc.AddPostReaction(context.Background(), 10, AddReactionRequest{UserID: 1, Reaction: "like"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "added" {
		t.Fatalf("expected added, got %s", status)
	}
	if notify.calls != 0 {
		t.Fatalf("expected no notification")
	}
}

func TestAddCommentReaction_Adds(t *testing.T) {
	repo := newFakeReactionRepo()
	postRepo := &fakePostRepo{authorID: 2}
	commentRepo := &fakeCommentRepo{authorID: 4}
	notify := &testNotifier{}

	svc := NewService(repo, postRepo, commentRepo, fakeAccess{}, notify)
	status, err := svc.AddCommentReaction(context.Background(), 10, AddReactionRequest{UserID: 1, Reaction: "like"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "added" {
		t.Fatalf("expected added, got %s", status)
	}
	if notify.calls != 0 {
		t.Fatalf("expected no notification")
	}
}

func TestAddPostReaction_RemovesWhenSame(t *testing.T) {
	repo := newFakeReactionRepo()
	repo.postReactions[[2]int64{10, 1}] = domainreaction.PostReaction{PostID: 10, UserID: 1, Reaction: domainreaction.Like}
	svc := NewService(repo, &fakePostRepo{authorID: 2}, &fakeCommentRepo{}, fakeAccess{}, nil)

	status, err := svc.AddPostReaction(context.Background(), 10, AddReactionRequest{UserID: 1, Reaction: "like"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "removed" {
		t.Fatalf("expected removed, got %s", status)
	}
	if _, ok := repo.postReactions[[2]int64{10, 1}]; ok {
		t.Fatalf("expected reaction removed")
	}
}

func TestAddPostReaction_UpdatesWhenDifferent(t *testing.T) {
	repo := newFakeReactionRepo()
	repo.postReactions[[2]int64{10, 1}] = domainreaction.PostReaction{PostID: 10, UserID: 1, Reaction: domainreaction.Like}
	notify := &testNotifier{}
	svc := NewService(repo, &fakePostRepo{authorID: 2}, &fakeCommentRepo{}, fakeAccess{}, notify)

	status, err := svc.AddPostReaction(context.Background(), 10, AddReactionRequest{UserID: 1, Reaction: "dislike"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "updated" {
		t.Fatalf("expected updated, got %s", status)
	}
	if repo.postReactions[[2]int64{10, 1}].Reaction != domainreaction.Dislike {
		t.Fatalf("expected reaction updated")
	}
	if notify.calls != 1 {
		t.Fatalf("expected notification")
	}
}

func TestAddPostReaction_InvalidInput(t *testing.T) {
	repo := newFakeReactionRepo()
	svc := NewService(repo, &fakePostRepo{}, &fakeCommentRepo{}, fakeAccess{}, nil)

	if _, err := svc.AddPostReaction(context.Background(), 10, AddReactionRequest{UserID: 0, Reaction: "like"}); err == nil {
		t.Fatalf("expected error for invalid user")
	}
	if _, err := svc.AddPostReaction(context.Background(), 10, AddReactionRequest{UserID: 1, Reaction: "laugh"}); err == nil {
		t.Fatalf("expected error for invalid reaction")
	}
}

func TestAddCommentReaction_RemovesWhenSame(t *testing.T) {
	repo := newFakeReactionRepo()
	repo.commentReactions[[2]int64{20, 1}] = domainreaction.CommentReaction{CommentID: 20, UserID: 1, Reaction: domainreaction.Dislike}
	svc := NewService(repo, &fakePostRepo{}, &fakeCommentRepo{authorID: 3}, fakeAccess{}, nil)

	status, err := svc.AddCommentReaction(context.Background(), 20, AddReactionRequest{UserID: 1, Reaction: "dislike"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "removed" {
		t.Fatalf("expected removed, got %s", status)
	}
	if _, ok := repo.commentReactions[[2]int64{20, 1}]; ok {
		t.Fatalf("expected reaction removed")
	}
}

func TestAddCommentReaction_DoesNotNotifySelf(t *testing.T) {
	repo := newFakeReactionRepo()
	notify := &testNotifier{}
	svc := NewService(repo, &fakePostRepo{}, &fakeCommentRepo{authorID: 1}, fakeAccess{}, notify)

	_, err := svc.AddCommentReaction(context.Background(), 20, AddReactionRequest{UserID: 1, Reaction: "like"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notify.calls != 0 {
		t.Fatalf("expected no notification for self reaction")
	}
}

func TestGetPostReactions_Maps(t *testing.T) {
	repo := newFakeReactionRepo()
	repo.postReactions[[2]int64{10, 1}] = domainreaction.PostReaction{PostID: 10, UserID: 1, Reaction: domainreaction.Like}
	repo.postReactions[[2]int64{10, 2}] = domainreaction.PostReaction{PostID: 10, UserID: 2, Reaction: domainreaction.Dislike}
	svc := NewService(repo, nil, nil, fakeAccess{}, nil)

	reactions, err := svc.GetPostReactions(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reactions) != 2 {
		t.Fatalf("expected 2 reactions")
	}
}

func TestGetCommentReactions_Maps(t *testing.T) {
	repo := newFakeReactionRepo()
	repo.commentReactions[[2]int64{20, 1}] = domainreaction.CommentReaction{CommentID: 20, UserID: 1, Reaction: domainreaction.Like}
	svc := NewService(repo, nil, nil, fakeAccess{}, nil)

	reactions, err := svc.GetCommentReactions(context.Background(), 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reactions) != 1 {
		t.Fatalf("expected 1 reaction")
	}
}
