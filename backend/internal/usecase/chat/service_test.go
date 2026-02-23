package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	domainchat "social-network/backend/internal/domain/chat"
	domaingroup "social-network/backend/internal/domain/group"
	"social-network/backend/pkg/logger"
)

type fakeChatRepo struct {
	members       map[int64]map[int64]bool
	conversations map[int64]domainchat.Conversation
	messages      map[int64][]domainchat.Message
	groupConv     map[int64]int64
	unread        map[int64]map[int64]int
	directConv    map[[2]int64]int64
	nextMsgID     int64
	markReadCalls int
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{
		members:       make(map[int64]map[int64]bool),
		conversations: make(map[int64]domainchat.Conversation),
		messages:      make(map[int64][]domainchat.Message),
		groupConv:     make(map[int64]int64),
		unread:        make(map[int64]map[int64]int),
		directConv:    make(map[[2]int64]int64),
		nextMsgID:     1,
	}
}

func (r *fakeChatRepo) IsMember(ctx context.Context, conversationID, userID int64) (bool, error) {
	return r.members[conversationID][userID], nil
}

func (r *fakeChatRepo) GetConversationMembers(ctx context.Context, conversationID int64) ([]int64, error) {
	m := r.members[conversationID]
	out := make([]int64, 0, len(m))
	for id := range m {
		out = append(out, id)
	}
	return out, nil
}
func (r *fakeChatRepo) GetConversationMembersMap(ctx context.Context, conversationIDs []int64) (map[int64][]int64, error) {
	out := make(map[int64][]int64)
	for _, id := range conversationIDs {
		m := r.members[id]
		if len(m) == 0 {
			continue
		}
		for userID := range m {
			out[id] = append(out[id], userID)
		}
	}
	return out, nil
}

func (r *fakeChatRepo) GetMessagesByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]domainchat.Message, error) {
	msgs := r.messages[conversationID]
	if len(msgs) > limit {
		return msgs[:limit], nil
	}
	return msgs, nil
}
func (r *fakeChatRepo) GetLastMessages(ctx context.Context, conversationIDs []int64) (map[int64]domainchat.Message, error) {
	out := make(map[int64]domainchat.Message)
	for _, id := range conversationIDs {
		msgs := r.messages[id]
		if len(msgs) > 0 {
			out[id] = msgs[0]
		}
	}
	return out, nil
}

func (r *fakeChatRepo) GetConversationByID(ctx context.Context, id int64) (domainchat.Conversation, error) {
	conv, ok := r.conversations[id]
	if !ok {
		return domainchat.Conversation{}, domainchat.ErrConversationNotFound
	}
	return conv, nil
}

func (r *fakeChatRepo) ListUserConversations(ctx context.Context, userID int64, limit, offset int) ([]domainchat.Conversation, error) {
	var out []domainchat.Conversation
	for id, conv := range r.conversations {
		if r.members[id][userID] {
			out = append(out, conv)
		}
	}
	return out, nil
}

func (r *fakeChatRepo) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	return r.unread[userID], nil
}

func (r *fakeChatRepo) GetGroupIDByConversationID(ctx context.Context, conversationID int64) (*int64, error) {
	for gid, cid := range r.groupConv {
		if cid == conversationID {
			return &gid, nil
		}
	}
	return nil, nil
}
func (r *fakeChatRepo) GetGroupConversationMap(ctx context.Context, conversationIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64)
	for gid, cid := range r.groupConv {
		for _, id := range conversationIDs {
			if id == cid {
				out[cid] = gid
			}
		}
	}
	return out, nil
}

func (r *fakeChatRepo) GetGroupConversationID(ctx context.Context, groupID int64) (int64, error) {
	cid, ok := r.groupConv[groupID]
	if !ok {
		return 0, domainchat.ErrConversationNotFound
	}
	return cid, nil
}

func (r *fakeChatRepo) MarkAsRead(ctx context.Context, conversationID, userID int64) error {
	r.markReadCalls++
	return nil
}

// unused interface methods
func (r *fakeChatRepo) GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 int64) (domainchat.Conversation, error) {
	key := directKey(userID1, userID2)
	if id, ok := r.directConv[key]; ok {
		return r.conversations[id], nil
	}
	id := int64(len(r.conversations) + 1)
	conv := domainchat.Conversation{ID: id, Type: domainchat.ConversationTypeDirect, CreatedAt: time.Now()}
	r.directConv[key] = id
	r.conversations[id] = conv
	r.members[id] = map[int64]bool{userID1: true, userID2: true}
	return conv, nil
}
func (r *fakeChatRepo) AddMember(ctx context.Context, conversationID, userID int64, role domainchat.ConversationRole) error {
	return nil
}
func (r *fakeChatRepo) CreateMessage(ctx context.Context, conversationID, senderID int64, content *string, mediaPath *string) (domainchat.Message, error) {
	msg := domainchat.Message{
		ID:             r.nextMsgID,
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MediaPath:      mediaPath,
		CreatedAt:      time.Now(),
	}
	r.nextMsgID++
	r.messages[conversationID] = append(r.messages[conversationID], msg)
	return msg, nil
}
func (r *fakeChatRepo) GetMessageByID(ctx context.Context, id int64) (domainchat.Message, error) {
	return domainchat.Message{}, nil
}
func (r *fakeChatRepo) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	return 0, nil
}
func (r *fakeChatRepo) HasMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	return false, nil
}
func (r *fakeChatRepo) AddMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	return nil
}
func (r *fakeChatRepo) RemoveMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	return nil
}
func (r *fakeChatRepo) ToggleMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	return true, nil
}
func (r *fakeChatRepo) ListMessageReactions(ctx context.Context, messageID int64) ([]domainchat.MessageReaction, error) {
	return nil, nil
}

// fake group repo

type fakeGroupRepo struct {
	members  map[int64][]int64
	isMember bool
}

func (r *fakeGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
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
func (r *fakeGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return r.isMember, nil
}
func (r *fakeGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return r.members[groupID], nil
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
	canSend  bool
	canGroup bool
}

func (f *fakeAccess) CanSendDirectMessage(ctx context.Context, senderID, receiverID int64) (bool, error) {
	return f.canSend, nil
}
func (f *fakeAccess) CanChatInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return f.canGroup, nil
}

func TestListConversations_GroupIncludesGroupID(t *testing.T) {
	repo := newFakeChatRepo()
	repo.conversations[1] = domainchat.Conversation{ID: 1, Type: domainchat.ConversationTypeGroup, CreatedAt: time.Now()}
	repo.members[1] = map[int64]bool{1: true}
	repo.groupConv[42] = 1
	repo.unread[1] = map[int64]int{1: 2}

	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{}, logger.NewDefault(false))
	convs, err := svc.ListConversations(context.Background(), 1, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(convs) != 1 || convs[0].GroupID == nil || *convs[0].GroupID != 42 {
		t.Fatalf("expected group_id in conversation")
	}
}

func TestGetConversation_ByID_Forbidden(t *testing.T) {
	repo := newFakeChatRepo()
	repo.conversations[1] = domainchat.Conversation{ID: 1, Type: domainchat.ConversationTypeDirect, CreatedAt: time.Now()}
	repo.members[1] = map[int64]bool{2: true}

	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{}, logger.NewDefault(false))
	_, err := svc.GetConversationByID(context.Background(), 1, 1)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestGetConversationMessages_Forbidden(t *testing.T) {
	repo := newFakeChatRepo()
	repo.members[1] = map[int64]bool{2: true}

	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{}, logger.NewDefault(false))
	_, err := svc.GetConversationMessages(context.Background(), 1, 1, 20, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestSendMessage_InvalidRequest(t *testing.T) {
	repo := newFakeChatRepo()
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{}, logger.NewDefault(false))

	_, _, err := svc.SendMessage(context.Background(), 1, SendMessageRequest{})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request, got %v", err)
	}
}

func TestSendDirectMessage_Empty(t *testing.T) {
	repo := newFakeChatRepo()
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{canSend: true}, logger.NewDefault(false))

	_, _, err := svc.SendDirectMessage(context.Background(), 1, 2, nil, nil)
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("expected empty message error, got %v", err)
	}
}

func TestSendDirectMessage_Forbidden(t *testing.T) {
	repo := newFakeChatRepo()
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{canSend: false}, logger.NewDefault(false))

	content := "hi"
	_, _, err := svc.SendDirectMessage(context.Background(), 1, 2, &content, nil)
	if !errors.Is(err, ErrCannotMessage) {
		t.Fatalf("expected cannot message, got %v", err)
	}
}

func TestSendDirectMessage_Success(t *testing.T) {
	repo := newFakeChatRepo()
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{canSend: true}, logger.NewDefault(false))

	content := "hi"
	msg, recipients, err := svc.SendDirectMessage(context.Background(), 1, 2, &content, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.SenderID != 1 || msg.Content == nil || *msg.Content != "hi" {
		t.Fatalf("unexpected message dto")
	}
	if len(recipients) != 1 || recipients[0] != 2 {
		t.Fatalf("unexpected recipients")
	}
	if repo.markReadCalls == 0 {
		t.Fatalf("expected mark as read")
	}
}

func TestSendGroupMessage_NotMember(t *testing.T) {
	repo := newFakeChatRepo()
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{canGroup: false}, logger.NewDefault(false))

	content := "hi"
	_, _, err := svc.SendGroupMessage(context.Background(), 1, 10, &content, nil)
	if !errors.Is(err, ErrNotGroupMember) {
		t.Fatalf("expected not group member, got %v", err)
	}
}

func TestSendGroupMessage_SuccessRecipients(t *testing.T) {
	repo := newFakeChatRepo()
	repo.groupConv[10] = 1
	repo.conversations[1] = domainchat.Conversation{ID: 1, Type: domainchat.ConversationTypeGroup, CreatedAt: time.Now()}
	repo.members[1] = map[int64]bool{1: true, 2: true, 3: true}

	groups := &fakeGroupRepo{members: map[int64][]int64{10: {1, 2, 3}}}
	svc := NewService(repo, groups, &fakeAccess{canGroup: true}, logger.NewDefault(false))

	content := "hello group"
	_, recipients, err := svc.SendGroupMessage(context.Background(), 1, 10, &content, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients")
	}
}

func TestMarkAsRead_Forbidden(t *testing.T) {
	repo := newFakeChatRepo()
	repo.members[1] = map[int64]bool{2: true}
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{}, logger.NewDefault(false))

	if err := svc.MarkAsRead(context.Background(), 1, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestGetConversationRecipients_FiltersUser(t *testing.T) {
	repo := newFakeChatRepo()
	repo.members[1] = map[int64]bool{1: true, 2: true, 3: true}
	svc := NewService(repo, &fakeGroupRepo{}, &fakeAccess{}, logger.NewDefault(false))

	recipients, err := svc.GetConversationRecipients(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients")
	}
}

func directKey(a, b int64) [2]int64 {
	if a < b {
		return [2]int64{a, b}
	}
	return [2]int64{b, a}
}
