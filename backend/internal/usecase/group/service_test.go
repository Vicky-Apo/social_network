package group

import (
	"context"
	"errors"
	"testing"
	"time"

	domaingroup "social-network/backend/internal/domain/group"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

type fakeGroupRepo struct {
	groups       map[int64]domaingroup.Group
	members      map[[2]int64]bool
	invites      map[int64]domaingroup.GroupInvitation
	joinReqs     map[int64]domaingroup.GroupJoinRequest
	nextInviteID int64
	nextJoinID   int64
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{
		groups:       make(map[int64]domaingroup.Group),
		members:      make(map[[2]int64]bool),
		invites:      make(map[int64]domaingroup.GroupInvitation),
		joinReqs:     make(map[int64]domaingroup.GroupJoinRequest),
		nextInviteID: 1,
		nextJoinID:   1,
	}
}

func (r *fakeGroupRepo) Create(ctx context.Context, creatorID int64, title string, description *string) (domaingroup.Group, error) {
	g := domaingroup.Group{ID: 1, CreatorID: creatorID, Title: title, Description: description, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.groups[g.ID] = g
	r.members[[2]int64{g.ID, creatorID}] = true
	return g, nil
}

func (r *fakeGroupRepo) List(ctx context.Context, userID int64, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeGroupRepo) Search(ctx context.Context, userID int64, query string, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeGroupRepo) GetWithMeta(ctx context.Context, userID, id int64) (domaingroup.GroupSummary, error) {
	g, ok := r.groups[id]
	if !ok {
		return domaingroup.GroupSummary{}, domaingroup.ErrGroupNotFound
	}
	return domaingroup.GroupSummary{Group: g}, nil
}

func (r *fakeGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	g, ok := r.groups[id]
	if !ok {
		return domaingroup.Group{}, domaingroup.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return r.members[[2]int64{groupID, userID}], nil
}

func (r *fakeGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeGroupRepo) ListMembers(ctx context.Context, groupID int64) ([]domaingroup.GroupMemberInfo, error) {
	return nil, nil
}

func (r *fakeGroupRepo) AddMember(ctx context.Context, groupID, userID int64) error {
	r.members[[2]int64{groupID, userID}] = true
	return nil
}

func (r *fakeGroupRepo) RemoveMember(ctx context.Context, groupID, userID int64) error {
	if !r.members[[2]int64{groupID, userID}] {
		return domaingroup.ErrNotMember
	}
	delete(r.members, [2]int64{groupID, userID})
	return nil
}

func (r *fakeGroupRepo) InvitationExists(ctx context.Context, groupID, inviteeID int64) (bool, error) {
	for _, inv := range r.invites {
		if inv.GroupID == groupID && inv.InviteeID == inviteeID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeGroupRepo) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	id := r.nextInviteID
	r.nextInviteID++
	inv := domaingroup.GroupInvitation{ID: id, GroupID: groupID, InviterID: inviterID, InviteeID: inviteeID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.invites[id] = inv
	return inv, nil
}

func (r *fakeGroupRepo) GetInvitationByID(ctx context.Context, id int64) (domaingroup.GroupInvitation, error) {
	inv, ok := r.invites[id]
	if !ok {
		return domaingroup.GroupInvitation{}, domaingroup.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeGroupRepo) ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]domaingroup.GroupInvitation, error) {
	return nil, nil
}

func (r *fakeGroupRepo) DeleteInvitation(ctx context.Context, id int64) error {
	if _, ok := r.invites[id]; !ok {
		return domaingroup.ErrInvitationNotFound
	}
	delete(r.invites, id)
	return nil
}

func (r *fakeGroupRepo) JoinRequestExists(ctx context.Context, groupID, userID int64) (bool, error) {
	for _, req := range r.joinReqs {
		if req.GroupID == groupID && req.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeGroupRepo) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	id := r.nextJoinID
	r.nextJoinID++
	req := domaingroup.GroupJoinRequest{ID: id, GroupID: groupID, UserID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.joinReqs[id] = req
	return req, nil
}

func (r *fakeGroupRepo) GetJoinRequestByID(ctx context.Context, id int64) (domaingroup.GroupJoinRequest, error) {
	req, ok := r.joinReqs[id]
	if !ok {
		return domaingroup.GroupJoinRequest{}, domaingroup.ErrJoinRequestNotFound
	}
	return req, nil
}

func (r *fakeGroupRepo) ListJoinRequestsByGroup(ctx context.Context, groupID int64) ([]domaingroup.GroupJoinRequest, error) {
	return nil, nil
}

func (r *fakeGroupRepo) DeleteJoinRequest(ctx context.Context, id int64) error {
	if _, ok := r.joinReqs[id]; !ok {
		return domaingroup.ErrJoinRequestNotFound
	}
	delete(r.joinReqs, id)
	return nil
}

// fake access service

type fakeAccess struct {
	canInvite  bool
	canApprove bool
	canView    bool
	canPost    bool
}

func (f *fakeAccess) CanInviteToGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canInvite, nil
}
func (f *fakeAccess) CanApproveGroupJoin(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canApprove, nil
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

func TestInviteToGroup_NotMemberForbidden(t *testing.T) {
	repo := newFakeGroupRepo()
	repo.groups[1] = domaingroup.Group{ID: 1, CreatorID: 1}

	svc := NewService(repo, &fakeAccess{canInvite: false}, nil)
	_, err := svc.InviteToGroup(context.Background(), 2, 1, 3)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestInviteToGroup_SendsNotification(t *testing.T) {
	repo := newFakeGroupRepo()
	repo.groups[1] = domaingroup.Group{ID: 1, CreatorID: 1}
	notify := &testNotifier{}

	svc := NewService(repo, &fakeAccess{canInvite: true}, notify)
	_, err := svc.InviteToGroup(context.Background(), 1, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notify.calls != 1 || notify.lastType != "group_invitation" {
		t.Fatalf("expected group_invitation notification")
	}
}

func TestRequestJoin_SendsNotificationToCreator(t *testing.T) {
	repo := newFakeGroupRepo()
	repo.groups[1] = domaingroup.Group{ID: 1, CreatorID: 99}
	notify := &testNotifier{}

	svc := NewService(repo, &fakeAccess{canPost: true}, notify)
	_, err := svc.RequestJoin(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notify.calls != 1 || notify.lastUserID != 99 {
		t.Fatalf("expected notification to creator")
	}
}
