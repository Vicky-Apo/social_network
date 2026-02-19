package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domaingroup "social-network/backend/internal/domain/group"
	usecasegroup "social-network/backend/internal/usecase/group"
	"social-network/backend/pkg/logger"
)

func TestGroupCreate_Unauthorized(t *testing.T) {
	h := NewGroupHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/groups", nil)
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeGroupRepo struct{}

func (r *fakeGroupRepo) Create(ctx context.Context, creatorID int64, title string, description *string) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeGroupRepo) List(ctx context.Context, userID int64, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return []domaingroup.GroupSummary{
		{Group: domaingroup.Group{ID: 1, Title: "G", CreatedAt: time.Now(), UpdatedAt: time.Now()}, MemberCount: 1, IsMember: true},
	}, nil
}
func (r *fakeGroupRepo) Search(ctx context.Context, userID int64, query string, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeGroupRepo) GetWithMeta(ctx context.Context, userID, id int64) (domaingroup.GroupSummary, error) {
	return domaingroup.GroupSummary{}, nil
}
func (r *fakeGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return false, nil
}
func (r *fakeGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeGroupRepo) ListMembers(ctx context.Context, groupID int64) ([]domaingroup.GroupMemberInfo, error) {
	return nil, nil
}
func (r *fakeGroupRepo) AddMember(ctx context.Context, groupID, userID int64) error    { return nil }
func (r *fakeGroupRepo) RemoveMember(ctx context.Context, groupID, userID int64) error { return nil }
func (r *fakeGroupRepo) InvitationExists(ctx context.Context, groupID, inviteeID int64) (bool, error) {
	return false, nil
}
func (r *fakeGroupRepo) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	return domaingroup.GroupInvitation{}, nil
}
func (r *fakeGroupRepo) GetInvitationByID(ctx context.Context, id int64) (domaingroup.GroupInvitation, error) {
	return domaingroup.GroupInvitation{}, nil
}
func (r *fakeGroupRepo) ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]domaingroup.GroupInvitation, error) {
	return nil, nil
}
func (r *fakeGroupRepo) DeleteInvitation(ctx context.Context, id int64) error { return nil }
func (r *fakeGroupRepo) JoinRequestExists(ctx context.Context, groupID, userID int64) (bool, error) {
	return false, nil
}
func (r *fakeGroupRepo) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	return domaingroup.GroupJoinRequest{}, nil
}
func (r *fakeGroupRepo) GetJoinRequestByID(ctx context.Context, id int64) (domaingroup.GroupJoinRequest, error) {
	return domaingroup.GroupJoinRequest{}, nil
}
func (r *fakeGroupRepo) ListJoinRequestsByGroup(ctx context.Context, groupID int64) ([]domaingroup.GroupJoinRequest, error) {
	return nil, nil
}
func (r *fakeGroupRepo) DeleteJoinRequest(ctx context.Context, id int64) error { return nil }

func TestGroupList_Success(t *testing.T) {
	repo := &fakeGroupRepo{}
	svc := usecasegroup.NewService(repo, nil, nil, logger.NewDefault(false))
	h := NewGroupHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.List), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
