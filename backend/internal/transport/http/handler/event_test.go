package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainevent "social-network/backend/internal/domain/event"
	domaingroup "social-network/backend/internal/domain/group"
	usecaseevent "social-network/backend/internal/usecase/event"
	"social-network/backend/pkg/logger"
)

func TestEventCreate_Unauthorized(t *testing.T) {
	h := NewEventHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/groups/1/events", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeEventRepo struct{}

func (r *fakeEventRepo) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	return e, nil
}
func (r *fakeEventRepo) GetByID(ctx context.Context, id int64) (domainevent.Event, error) {
	return domainevent.Event{ID: id, GroupID: 1}, nil
}
func (r *fakeEventRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainevent.Event, error) {
	return []domainevent.Event{{ID: 1, GroupID: groupID, Title: "Meet", EventTime: time.Now()}}, nil
}
func (r *fakeEventRepo) Update(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	return e, nil
}
func (r *fakeEventRepo) Delete(ctx context.Context, id int64) error {
	return nil
}
func (r *fakeEventRepo) UpsertResponse(ctx context.Context, eventID, userID int64, response string) (domainevent.EventResponse, error) {
	return domainevent.EventResponse{}, nil
}
func (r *fakeEventRepo) ListResponses(ctx context.Context, eventID int64) ([]domainevent.EventResponseInfo, error) {
	return nil, nil
}

type fakeEventGroupRepo struct{}

func (r *fakeEventGroupRepo) Create(ctx context.Context, creatorID int64, title string, description *string) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeEventGroupRepo) List(ctx context.Context, userID int64, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeEventGroupRepo) Search(ctx context.Context, userID int64, query string, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeEventGroupRepo) GetWithMeta(ctx context.Context, userID, id int64) (domaingroup.GroupSummary, error) {
	return domaingroup.GroupSummary{}, nil
}
func (r *fakeEventGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeEventGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return true, nil
}
func (r *fakeEventGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeEventGroupRepo) ListMembers(ctx context.Context, groupID int64) ([]domaingroup.GroupMemberInfo, error) {
	return nil, nil
}
func (r *fakeEventGroupRepo) AddMember(ctx context.Context, groupID, userID int64) error {
	return nil
}
func (r *fakeEventGroupRepo) RemoveMember(ctx context.Context, groupID, userID int64) error {
	return nil
}
func (r *fakeEventGroupRepo) InvitationExists(ctx context.Context, groupID, inviteeID int64) (bool, error) {
	return false, nil
}
func (r *fakeEventGroupRepo) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	return domaingroup.GroupInvitation{}, nil
}
func (r *fakeEventGroupRepo) GetInvitationByID(ctx context.Context, id int64) (domaingroup.GroupInvitation, error) {
	return domaingroup.GroupInvitation{}, nil
}
func (r *fakeEventGroupRepo) ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]domaingroup.GroupInvitation, error) {
	return nil, nil
}
func (r *fakeEventGroupRepo) DeleteInvitation(ctx context.Context, id int64) error { return nil }
func (r *fakeEventGroupRepo) JoinRequestExists(ctx context.Context, groupID, userID int64) (bool, error) {
	return false, nil
}
func (r *fakeEventGroupRepo) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	return domaingroup.GroupJoinRequest{}, nil
}
func (r *fakeEventGroupRepo) GetJoinRequestByID(ctx context.Context, id int64) (domaingroup.GroupJoinRequest, error) {
	return domaingroup.GroupJoinRequest{}, nil
}
func (r *fakeEventGroupRepo) ListJoinRequestsByGroup(ctx context.Context, groupID int64) ([]domaingroup.GroupJoinRequest, error) {
	return nil, nil
}
func (r *fakeEventGroupRepo) DeleteJoinRequest(ctx context.Context, id int64) error { return nil }

type fakeEventAccess struct{}

func (f fakeEventAccess) CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return true, nil
}
func (f fakeEventAccess) CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return true, nil
}

func TestEventList_Success(t *testing.T) {
	repo := &fakeEventRepo{}
	groupRepo := &fakeEventGroupRepo{}
	svc := usecaseevent.NewService(repo, groupRepo, fakeEventAccess{}, nil, logger.NewDefault(false))
	h := NewEventHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/groups/1/events", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.ListByGroup), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
