package event

import (
	"context"
	"errors"
	"testing"
	"time"

	domainevent "social-network/backend/internal/domain/event"
	domaingroup "social-network/backend/internal/domain/group"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

type fakeEventRepo struct {
	events    map[int64]domainevent.Event
	responses map[int64][]domainevent.EventResponseInfo
	nextID    int64
}

func newFakeEventRepo() *fakeEventRepo {
	return &fakeEventRepo{events: make(map[int64]domainevent.Event), responses: make(map[int64][]domainevent.EventResponseInfo), nextID: 1}
}

func (r *fakeEventRepo) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	e.ID = r.nextID
	r.nextID++
	e.CreatedAt = time.Now()
	r.events[e.ID] = e
	return e, nil
}

func (r *fakeEventRepo) GetByID(ctx context.Context, id int64) (domainevent.Event, error) {
	e, ok := r.events[id]
	if !ok {
		return domainevent.Event{}, domainevent.ErrNotFound
	}
	return e, nil
}

func (r *fakeEventRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainevent.Event, error) {
	return nil, nil
}

func (r *fakeEventRepo) UpsertResponse(ctx context.Context, eventID, userID int64, response string) (domainevent.EventResponse, error) {
	return domainevent.EventResponse{EventID: eventID, UserID: userID, Response: response, RespondedAt: timePtr(time.Now())}, nil
}

func (r *fakeEventRepo) ListResponses(ctx context.Context, eventID int64) ([]domainevent.EventResponseInfo, error) {
	return r.responses[eventID], nil
}

// fake group repo

type fakeGroupRepo struct {
	memberIDs map[int64][]int64
	groups    map[int64]domaingroup.Group
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{memberIDs: make(map[int64][]int64), groups: make(map[int64]domaingroup.Group)}
}

func (r *fakeGroupRepo) Create(ctx context.Context, creatorID int64, title string, description *string) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeGroupRepo) List(ctx context.Context, userID int64, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeGroupRepo) Search(ctx context.Context, userID int64, query string, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeGroupRepo) GetWithMeta(ctx context.Context, userID, id int64) (domaingroup.GroupSummary, error) {
	return domaingroup.GroupSummary{}, nil
}
func (r *fakeGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	g, ok := r.groups[id]
	if !ok {
		return domaingroup.Group{}, domaingroup.ErrGroupNotFound
	}
	return g, nil
}
func (r *fakeGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return true, nil
}
func (r *fakeGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return r.memberIDs[groupID], nil
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

// fake access

type fakeAccess struct {
	canView bool
	canPost bool
}

func (f *fakeAccess) CanViewGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canView, nil
}
func (f *fakeAccess) CanPostInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canPost, nil
}

// fake notifier

type testNotifier struct {
	calls      int
	lastUserID int64
	lastType   string
}

func (n *testNotifier) CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error) {
	n.calls++
	n.lastUserID = req.UserID
	n.lastType = req.Type
	return usecasenotification.NotificationDTO{}, nil
}

func TestCreateEvent_ForbiddenForNonMember(t *testing.T) {
	repo := newFakeEventRepo()
	groups := newFakeGroupRepo()
	access := &fakeAccess{canPost: false}

	svc := NewService(repo, groups, access, nil)
	_, err := svc.CreateEvent(context.Background(), 1, 1, CreateEventRequest{Title: "Meet", EventTime: time.Now()})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestCreateEvent_NotifiesMembers(t *testing.T) {
	repo := newFakeEventRepo()
	groups := newFakeGroupRepo()
	groups.memberIDs[1] = []int64{1, 2, 3}
	access := &fakeAccess{canPost: true}
	notify := &testNotifier{}

	svc := NewService(repo, groups, access, notify)
	_, err := svc.CreateEvent(context.Background(), 1, 1, CreateEventRequest{Title: "Meet", EventTime: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notify.calls == 0 {
		t.Fatalf("expected notifications")
	}
}

func TestRespond_InvalidResponse(t *testing.T) {
	repo := newFakeEventRepo()
	groups := newFakeGroupRepo()
	access := &fakeAccess{canView: true}

	svc := NewService(repo, groups, access, nil)
	_, err := svc.Respond(context.Background(), 1, 1, "maybe")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("expected invalid response, got %v", err)
	}
}

func timePtr(t time.Time) *time.Time { return &t }
