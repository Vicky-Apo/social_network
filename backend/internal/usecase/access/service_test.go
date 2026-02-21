package access

import (
	"context"
	"errors"
	"testing"

	domainfollow "social-network/backend/internal/domain/follow"
	domaingroup "social-network/backend/internal/domain/group"
	domainpost "social-network/backend/internal/domain/post"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/pkg/logger"
)

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
func (r *fakeUserRepo) UpdateProfile(ctx context.Context, id int64, nickname, about, avatarPath *string) (domainuser.User, error) {
	return domainuser.User{}, nil
}
func (r *fakeUserRepo) SetVisibility(ctx context.Context, id int64, isPublic bool) error {
	return nil
}
func (r *fakeUserRepo) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakeUserRepo) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}
func (r *fakeUserRepo) ListFollowers(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) ListFollowing(ctx context.Context, userID int64) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) ListUsers(ctx context.Context, viewerID int64, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) SearchUsers(ctx context.Context, viewerID int64, query string, limit, offset int) ([]domainuser.User, error) {
	return nil, nil
}

type fakeFollowRepo struct {
	follows map[[2]int64]bool
}

func (r *fakeFollowRepo) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return r.follows[[2]int64{followerID, followingID}], nil
}
func (r *fakeFollowRepo) RequestExists(ctx context.Context, requesterID, targetID int64) (bool, error) {
	return false, nil
}
func (r *fakeFollowRepo) CreateRequest(ctx context.Context, requesterID, targetID int64) (domainfollow.FollowRequest, error) {
	return domainfollow.FollowRequest{}, nil
}
func (r *fakeFollowRepo) GetRequestByID(ctx context.Context, id int64) (domainfollow.FollowRequest, error) {
	return domainfollow.FollowRequest{}, nil
}
func (r *fakeFollowRepo) UpdateRequestStatus(ctx context.Context, id int64, status string) error {
	return nil
}
func (r *fakeFollowRepo) ListRequestsByTarget(ctx context.Context, targetID int64) ([]domainfollow.FollowRequest, error) {
	return nil, nil
}
func (r *fakeFollowRepo) ListRequestsByRequester(ctx context.Context, requesterID int64) ([]domainfollow.FollowRequest, error) {
	return nil, nil
}
func (r *fakeFollowRepo) CreateFollow(ctx context.Context, followerID, followingID int64) error {
	return nil
}
func (r *fakeFollowRepo) DeleteFollow(ctx context.Context, followerID, followingID int64) error {
	return nil
}
func (r *fakeFollowRepo) GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error) {
	return nil, nil
}

type fakePostRepo struct {
	posts        map[int64]domainpost.Post
	allowed      map[[2]int64]bool
	getByIDErr   error
	isAllowedErr error
}

func (r *fakePostRepo) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	if r.getByIDErr != nil {
		return domainpost.Post{}, r.getByIDErr
	}
	p, ok := r.posts[id]
	if !ok {
		return domainpost.Post{}, domainpost.ErrNotFound
	}
	return p, nil
}
func (r *fakePostRepo) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	if r.isAllowedErr != nil {
		return false, r.isAllowedErr
	}
	return r.allowed[[2]int64{postID, userID}], nil
}
func (r *fakePostRepo) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	return domainpost.Post{}, nil
}
func (r *fakePostRepo) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	return nil, nil
}

type fakeGroupRepo struct {
	members map[[2]int64]bool
	groups  map[int64]domaingroup.Group
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
	return r.members[[2]int64{groupID, userID}], nil
}
func (r *fakeGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeGroupRepo) ListMembers(ctx context.Context, groupID int64) ([]domaingroup.GroupMemberInfo, error) {
	return nil, nil
}
func (r *fakeGroupRepo) AddMember(ctx context.Context, groupID, userID int64) error {
	return nil
}
func (r *fakeGroupRepo) RemoveMember(ctx context.Context, groupID, userID int64) error {
	return nil
}
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

func TestCanViewProfile(t *testing.T) {
	users := map[int64]domainuser.User{
		1: {ID: 1, IsPublic: true},
		2: {ID: 2, IsPublic: false},
	}
	follows := map[[2]int64]bool{
		{3, 2}: true,
	}

	svc := NewService(
		&fakeUserRepo{users: users},
		&fakeFollowRepo{follows: follows},
		&fakePostRepo{},
		&fakeGroupRepo{},
		logger.NewDefault(false),
	)

	ok, err := svc.CanViewProfile(context.Background(), 1, 1)
	if err != nil || !ok {
		t.Fatalf("expected owner can view")
	}

	ok, err = svc.CanViewProfile(context.Background(), 0, 1)
	if err != nil || !ok {
		t.Fatalf("expected public view")
	}

	ok, err = svc.CanViewProfile(context.Background(), 0, 2)
	if err != nil || ok {
		t.Fatalf("expected private no view")
	}

	ok, err = svc.CanViewProfile(context.Background(), 3, 2)
	if err != nil || !ok {
		t.Fatalf("expected follower view")
	}
}

func TestCanViewPost_GroupMembership(t *testing.T) {
	groupID := int64(9)
	post := domainpost.Post{ID: 1, AuthorID: 10, GroupID: &groupID}

	svc := NewService(
		&fakeUserRepo{users: map[int64]domainuser.User{10: {ID: 10, IsPublic: true}}},
		&fakeFollowRepo{follows: map[[2]int64]bool{}},
		&fakePostRepo{posts: map[int64]domainpost.Post{1: post}},
		&fakeGroupRepo{members: map[[2]int64]bool{{groupID, 1}: true}},
		logger.NewDefault(false),
	)

	ok, err := svc.CanViewPost(context.Background(), 1, 1)
	if err != nil || !ok {
		t.Fatalf("expected member can view group post")
	}

	ok, err = svc.CanViewPost(context.Background(), 2, 1)
	if err != nil || ok {
		t.Fatalf("expected non-member denied")
	}
}

func TestCanViewPost_Privacy(t *testing.T) {
	postPublic := domainpost.Post{ID: 1, AuthorID: 5, Privacy: "public"}
	postFollowers := domainpost.Post{ID: 2, AuthorID: 5, Privacy: "followers"}
	postPrivate := domainpost.Post{ID: 3, AuthorID: 5, Privacy: "private"}

	users := map[int64]domainuser.User{
		5: {ID: 5, IsPublic: false},
	}
	follows := map[[2]int64]bool{
		{7, 5}: true,
	}
	allowed := map[[2]int64]bool{
		{3, 7}: true,
	}

	svc := NewService(
		&fakeUserRepo{users: users},
		&fakeFollowRepo{follows: follows},
		&fakePostRepo{posts: map[int64]domainpost.Post{
			1: postPublic, 2: postFollowers, 3: postPrivate,
		}, allowed: allowed},
		&fakeGroupRepo{},
		logger.NewDefault(false),
	)

	ok, err := svc.CanViewPost(context.Background(), 6, 1)
	if err != nil || ok {
		t.Fatalf("expected non-follower cannot view private author")
	}

	ok, err = svc.CanViewPost(context.Background(), 7, 1)
	if err != nil || !ok {
		t.Fatalf("expected follower can view public")
	}

	ok, err = svc.CanViewPost(context.Background(), 7, 2)
	if err != nil || !ok {
		t.Fatalf("expected follower can view followers post")
	}

	ok, err = svc.CanViewPost(context.Background(), 7, 3)
	if err != nil || !ok {
		t.Fatalf("expected allowed user can view private post")
	}
}

func TestCanSendDirectMessage(t *testing.T) {
	follows := map[[2]int64]bool{
		{1, 2}: false,
		{2, 1}: true,
	}
	svc := NewService(
		&fakeUserRepo{users: map[int64]domainuser.User{2: {ID: 2, IsPublic: false}}},
		&fakeFollowRepo{follows: follows},
		&fakePostRepo{},
		&fakeGroupRepo{},
		logger.NewDefault(false),
	)

	ok, err := svc.CanSendDirectMessage(context.Background(), 1, 2)
	if err != nil || !ok {
		t.Fatalf("expected can message when reverse follow exists")
	}
}

func TestCanSendDirectMessage_PublicProfile(t *testing.T) {
	svc := NewService(
		&fakeUserRepo{users: map[int64]domainuser.User{2: {ID: 2, IsPublic: true}}},
		&fakeFollowRepo{follows: map[[2]int64]bool{}},
		&fakePostRepo{},
		&fakeGroupRepo{},
		logger.NewDefault(false),
	)
	ok, err := svc.CanSendDirectMessage(context.Background(), 1, 2)
	if err != nil || !ok {
		t.Fatalf("expected can message public profile")
	}
}

func TestCanApproveGroupJoin(t *testing.T) {
	svc := NewService(
		&fakeUserRepo{},
		&fakeFollowRepo{},
		&fakePostRepo{},
		&fakeGroupRepo{groups: map[int64]domaingroup.Group{10: {ID: 10, CreatorID: 1}}},
		logger.NewDefault(false),
	)
	ok, err := svc.CanApproveGroupJoin(context.Background(), 1, 10)
	if err != nil || !ok {
		t.Fatalf("expected creator can approve")
	}
	ok, err = svc.CanApproveGroupJoin(context.Background(), 2, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected non-creator denied")
	}
}

func TestCanViewPost_ErrorPaths(t *testing.T) {
	svc := NewService(
		&fakeUserRepo{},
		&fakeFollowRepo{},
		&fakePostRepo{getByIDErr: errors.New("boom")},
		&fakeGroupRepo{},
		logger.NewDefault(false),
	)
	if _, err := svc.CanViewPost(context.Background(), 1, 1); err == nil {
		t.Fatalf("expected error")
	}
}
