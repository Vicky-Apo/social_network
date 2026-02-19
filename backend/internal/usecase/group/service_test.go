package group

import (
	"context"
	"errors"
	"testing"
	"time"

	domaingroup "social-network/backend/internal/domain/group"
	domainfollow "social-network/backend/internal/domain/follow"
	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	usecaseaccess "social-network/backend/internal/usecase/access"
	"social-network/backend/pkg/logger"
)

type fakeGroupRepo struct {
	nextGroupID int64
	nextInviteID int64
	nextJoinID   int64
	nextEventID  int64

	groups       map[int64]domaingroup.Group
	members      map[int64]map[int64]time.Time
	invitations  map[int64]domaingroup.GroupInvitation
	joinRequests map[int64]domaingroup.GroupJoinRequest
	events       map[int64]domaingroup.GroupEvent
	eventResp    map[int64]map[int64]domaingroup.GroupEventResponse
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{
		nextGroupID: 1,
		nextInviteID: 1,
		nextJoinID:  1,
		nextEventID: 1,
		groups:      make(map[int64]domaingroup.Group),
		members:     make(map[int64]map[int64]time.Time),
		invitations: make(map[int64]domaingroup.GroupInvitation),
		joinRequests: make(map[int64]domaingroup.GroupJoinRequest),
		events:      make(map[int64]domaingroup.GroupEvent),
		eventResp:   make(map[int64]map[int64]domaingroup.GroupEventResponse),
	}
}

func (r *fakeGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	g, ok := r.groups[id]
	if !ok {
		return domaingroup.Group{}, domaingroup.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepo) List(ctx context.Context, query string, limit, offset int) ([]domaingroup.Group, error) {
	out := make([]domaingroup.Group, 0, len(r.groups))
	for _, g := range r.groups {
		out = append(out, g)
	}
	return out, nil
}

func (r *fakeGroupRepo) Create(ctx context.Context, group domaingroup.Group) (domaingroup.Group, error) {
	group.ID = r.nextGroupID
	r.nextGroupID++
	now := time.Now().UTC()
	group.CreatedAt = now
	group.UpdatedAt = now
	r.groups[group.ID] = group
	return group, nil
}

func (r *fakeGroupRepo) Update(ctx context.Context, group domaingroup.Group) (domaingroup.Group, error) {
	if _, ok := r.groups[group.ID]; !ok {
		return domaingroup.Group{}, domaingroup.ErrGroupNotFound
	}
	group.UpdatedAt = time.Now().UTC()
	r.groups[group.ID] = group
	return group, nil
}

func (r *fakeGroupRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := r.groups[id]; !ok {
		return domaingroup.ErrGroupNotFound
	}
	delete(r.groups, id)
	return nil
}

func (r *fakeGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	members, ok := r.members[groupID]
	if !ok {
		return false, nil
	}
	_, ok = members[userID]
	return ok, nil
}

func (r *fakeGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	members := r.members[groupID]
	out := make([]int64, 0, len(members))
	for id := range members {
		out = append(out, id)
	}
	return out, nil
}

func (r *fakeGroupRepo) ListMembers(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupMember, error) {
	members := r.members[groupID]
	out := make([]domaingroup.GroupMember, 0, len(members))
	for id, joined := range members {
		out = append(out, domaingroup.GroupMember{
			GroupID: groupID,
			UserID:  id,
			JoinedAt: joined,
		})
	}
	return out, nil
}

func (r *fakeGroupRepo) AddMember(ctx context.Context, groupID, userID int64) error {
	if _, ok := r.members[groupID]; !ok {
		r.members[groupID] = make(map[int64]time.Time)
	}
	r.members[groupID][userID] = time.Now().UTC()
	return nil
}

func (r *fakeGroupRepo) RemoveMember(ctx context.Context, groupID, userID int64) error {
	members, ok := r.members[groupID]
	if !ok {
		return domaingroup.ErrNotMember
	}
	if _, ok := members[userID]; !ok {
		return domaingroup.ErrNotMember
	}
	delete(members, userID)
	return nil
}

func (r *fakeGroupRepo) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	inv := domaingroup.GroupInvitation{
		ID:        r.nextInviteID,
		GroupID:   groupID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	r.nextInviteID++
	r.invitations[inv.ID] = inv
	return inv, nil
}

func (r *fakeGroupRepo) GetInvitationByID(ctx context.Context, id int64) (domaingroup.GroupInvitation, error) {
	inv, ok := r.invitations[id]
	if !ok {
		return domaingroup.GroupInvitation{}, domaingroup.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeGroupRepo) ListInvitationsByInvitee(ctx context.Context, inviteeID int64, limit, offset int) ([]domaingroup.GroupInvitation, error) {
	var out []domaingroup.GroupInvitation
	for _, inv := range r.invitations {
		if inv.InviteeID == inviteeID {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (r *fakeGroupRepo) ListInvitationsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupInvitation, error) {
	var out []domaingroup.GroupInvitation
	for _, inv := range r.invitations {
		if inv.GroupID == groupID {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (r *fakeGroupRepo) DeleteInvitation(ctx context.Context, id int64) error {
	if _, ok := r.invitations[id]; !ok {
		return domaingroup.ErrInvitationNotFound
	}
	delete(r.invitations, id)
	return nil
}

func (r *fakeGroupRepo) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	req := domaingroup.GroupJoinRequest{
		ID:        r.nextJoinID,
		GroupID:   groupID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	r.nextJoinID++
	r.joinRequests[req.ID] = req
	return req, nil
}

func (r *fakeGroupRepo) GetJoinRequestByID(ctx context.Context, id int64) (domaingroup.GroupJoinRequest, error) {
	req, ok := r.joinRequests[id]
	if !ok {
		return domaingroup.GroupJoinRequest{}, domaingroup.ErrJoinRequestNotFound
	}
	return req, nil
}

func (r *fakeGroupRepo) ListJoinRequestsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupJoinRequest, error) {
	var out []domaingroup.GroupJoinRequest
	for _, req := range r.joinRequests {
		if req.GroupID == groupID {
			out = append(out, req)
		}
	}
	return out, nil
}

func (r *fakeGroupRepo) DeleteJoinRequest(ctx context.Context, id int64) error {
	if _, ok := r.joinRequests[id]; !ok {
		return domaingroup.ErrJoinRequestNotFound
	}
	delete(r.joinRequests, id)
	return nil
}

func (r *fakeGroupRepo) CreateEvent(ctx context.Context, event domaingroup.GroupEvent) (domaingroup.GroupEvent, error) {
	event.ID = r.nextEventID
	r.nextEventID++
	now := time.Now().UTC()
	event.CreatedAt = now
	event.UpdatedAt = now
	r.events[event.ID] = event
	return event, nil
}

func (r *fakeGroupRepo) GetEventByID(ctx context.Context, id int64) (domaingroup.GroupEvent, error) {
	ev, ok := r.events[id]
	if !ok {
		return domaingroup.GroupEvent{}, domaingroup.ErrEventNotFound
	}
	return ev, nil
}

func (r *fakeGroupRepo) ListEventsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupEvent, error) {
	var out []domaingroup.GroupEvent
	for _, ev := range r.events {
		if ev.GroupID == groupID {
			out = append(out, ev)
		}
	}
	return out, nil
}

func (r *fakeGroupRepo) UpdateEvent(ctx context.Context, event domaingroup.GroupEvent) (domaingroup.GroupEvent, error) {
	if _, ok := r.events[event.ID]; !ok {
		return domaingroup.GroupEvent{}, domaingroup.ErrEventNotFound
	}
	event.UpdatedAt = time.Now().UTC()
	r.events[event.ID] = event
	return event, nil
}

func (r *fakeGroupRepo) DeleteEvent(ctx context.Context, id int64) error {
	if _, ok := r.events[id]; !ok {
		return domaingroup.ErrEventNotFound
	}
	delete(r.events, id)
	return nil
}

func (r *fakeGroupRepo) UpsertEventResponse(ctx context.Context, eventID, userID int64, response string) error {
	if _, ok := r.eventResp[eventID]; !ok {
		r.eventResp[eventID] = make(map[int64]domaingroup.GroupEventResponse)
	}
	now := time.Now().UTC()
	r.eventResp[eventID][userID] = domaingroup.GroupEventResponse{
		EventID:     eventID,
		UserID:      userID,
		Response:    response,
		RespondedAt: &now,
	}
	return nil
}

func (r *fakeGroupRepo) GetEventResponse(ctx context.Context, eventID, userID int64) (domaingroup.GroupEventResponse, error) {
	evMap, ok := r.eventResp[eventID]
	if !ok {
		return domaingroup.GroupEventResponse{}, domaingroup.ErrEventResponseNotFound
	}
	resp, ok := evMap[userID]
	if !ok {
		return domaingroup.GroupEventResponse{}, domaingroup.ErrEventResponseNotFound
	}
	return resp, nil
}

func (r *fakeGroupRepo) ListEventResponses(ctx context.Context, eventID int64) ([]domaingroup.GroupEventResponse, error) {
	evMap := r.eventResp[eventID]
	var out []domaingroup.GroupEventResponse
	for _, resp := range evMap {
		out = append(out, resp)
	}
	return out, nil
}

type fakeUserRepo struct {
	users map[int64]domainuser.User
}

func (r *fakeUserRepo) GetByID(ctx context.Context, id int64) (domainuser.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domainuser.User{}, domainuser.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error { return errors.New("not implemented") }
func (r *fakeUserRepo) CountFollowers(ctx context.Context, userID int64) (int64, error) { return 0, errors.New("not implemented") }
func (r *fakeUserRepo) CountFollowing(ctx context.Context, userID int64) (int64, error) { return 0, errors.New("not implemented") }
func (r *fakeUserRepo) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) { return nil, errors.New("not implemented") }
func (r *fakeUserRepo) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) { return nil, errors.New("not implemented") }
func (r *fakeUserRepo) ListUsers(ctx context.Context) ([]domainuser.User, error) { return nil, errors.New("not implemented") }
func (r *fakeUserRepo) SearchUsers(ctx context.Context, query string) ([]domainuser.User, error) { return nil, errors.New("not implemented") }

type fakeFollowRepo struct{}

func (r *fakeFollowRepo) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return false, errors.New("not implemented")
}
func (r *fakeFollowRepo) RequestExists(ctx context.Context, requesterID, targetID int64) (bool, error) {
	return false, errors.New("not implemented")
}
func (r *fakeFollowRepo) CreateRequest(ctx context.Context, requesterID, targetID int64) (domainfollow.FollowRequest, error) {
	return domainfollow.FollowRequest{}, errors.New("not implemented")
}
func (r *fakeFollowRepo) GetRequestByID(ctx context.Context, id int64) (domainfollow.FollowRequest, error) {
	return domainfollow.FollowRequest{}, errors.New("not implemented")
}
func (r *fakeFollowRepo) UpdateRequestStatus(ctx context.Context, id int64, status string) error {
	return errors.New("not implemented")
}
func (r *fakeFollowRepo) ListRequestsByTarget(ctx context.Context, targetID int64) ([]domainfollow.FollowRequest, error) {
	return nil, errors.New("not implemented")
}
func (r *fakeFollowRepo) ListRequestsByRequester(ctx context.Context, requesterID int64) ([]domainfollow.FollowRequest, error) {
	return nil, errors.New("not implemented")
}
func (r *fakeFollowRepo) CreateFollow(ctx context.Context, followerID, followingID int64) error {
	return errors.New("not implemented")
}
func (r *fakeFollowRepo) DeleteFollow(ctx context.Context, followerID, followingID int64) error {
	return errors.New("not implemented")
}
func (r *fakeFollowRepo) GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error) {
	return nil, errors.New("not implemented")
}

type fakePostRepo struct{}

func (r *fakePostRepo) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, errors.New("not implemented")
}
func (r *fakePostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	return domainpost.Post{}, errors.New("not implemented")
}
func (r *fakePostRepo) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	return domainpost.Post{}, errors.New("not implemented")
}
func (r *fakePostRepo) Update(ctx context.Context, post domainpost.Post, allowedUserIDs *[]int64) (domainpost.Post, error) {
	return domainpost.Post{}, errors.New("not implemented")
}
func (r *fakePostRepo) Delete(ctx context.Context, id int64) error { return errors.New("not implemented") }
func (r *fakePostRepo) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, errors.New("not implemented")
}
func (r *fakePostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, errors.New("not implemented")
}
func (r *fakePostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	return false, errors.New("not implemented")
}

func newServiceWithAccess(t *testing.T) (*Service, *fakeGroupRepo) {
	t.Helper()
	groupRepo := newFakeGroupRepo()
	userRepo := &fakeUserRepo{
		users: map[int64]domainuser.User{
			1: {ID: 1},
			2: {ID: 2},
			3: {ID: 3},
		},
	}
	access := usecaseaccess.NewService(userRepo, &fakeFollowRepo{}, &fakePostRepo{}, groupRepo, logger.NewDefault(false))
	service := NewService(groupRepo, userRepo, access, nil)
	return service, groupRepo
}

func TestCreateGroupAddsCreatorMember(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{
		Title: "Gophers",
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if group.ID == 0 {
		t.Fatalf("expected group id")
	}
	isMember, _ := repo.IsMember(context.Background(), group.ID, 1)
	if !isMember {
		t.Fatalf("creator should be member")
	}
}

func TestUpdateGroupForbidden(t *testing.T) {
	service, _ := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	_, err = service.Update(context.Background(), group.ID, 2, UpdateGroupRequest{Title: strPtr("New")})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestInviteAndAcceptAddsMember(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	inv, err := service.Invite(context.Background(), group.ID, 1, 2)
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if err := service.RespondInvitation(context.Background(), inv.ID, 2, "accepted"); err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	isMember, _ := repo.IsMember(context.Background(), group.ID, 2)
	if !isMember {
		t.Fatalf("invitee should be member after accept")
	}
}

func TestInviteExistingMemberInvalid(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := repo.AddMember(context.Background(), group.ID, 2); err != nil {
		t.Fatalf("add member: %v", err)
	}
	_, err = service.Invite(context.Background(), group.ID, 1, 2)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestJoinRequestAlreadyMemberInvalid(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := repo.AddMember(context.Background(), group.ID, 2); err != nil {
		t.Fatalf("add member: %v", err)
	}
	_, err = service.RequestJoin(context.Background(), group.ID, 2)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestJoinRequestApproveAddsMember(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	req, err := service.RequestJoin(context.Background(), group.ID, 2)
	if err != nil {
		t.Fatalf("request join: %v", err)
	}
	if err := service.RespondJoinRequest(context.Background(), group.ID, req.ID, 1, "accepted"); err != nil {
		t.Fatalf("approve join: %v", err)
	}
	isMember, _ := repo.IsMember(context.Background(), group.ID, 2)
	if !isMember {
		t.Fatalf("requester should be member after approval")
	}
}

func TestCreateEventRequiresMembership(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	_, err = service.CreateEvent(context.Background(), group.ID, 2, CreateEventRequest{
		Title:     "Meetup",
		EventTime: timePtr(time.Now().Add(24 * time.Hour)),
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if err := repo.AddMember(context.Background(), group.ID, 2); err != nil {
		t.Fatalf("add member: %v", err)
	}
	_, err = service.CreateEvent(context.Background(), group.ID, 2, CreateEventRequest{
		Title:     "Meetup",
		EventTime: timePtr(time.Now().Add(24 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
}

func TestRSVPInvalidResponse(t *testing.T) {
	service, repo := newServiceWithAccess(t)
	group, err := service.Create(context.Background(), 1, CreateGroupRequest{Title: "Gophers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := repo.AddMember(context.Background(), group.ID, 1); err != nil {
		t.Fatalf("add member: %v", err)
	}
	event, err := service.CreateEvent(context.Background(), group.ID, 1, CreateEventRequest{
		Title:     "Meetup",
		EventTime: timePtr(time.Now().Add(24 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if err := service.RSVP(context.Background(), group.ID, event.ID, 1, "maybe"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func strPtr(v string) *string { return &v }
func timePtr(v time.Time) *time.Time { return &v }
