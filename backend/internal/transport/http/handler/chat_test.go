package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainchat "social-network/backend/internal/domain/chat"
	domaingroup "social-network/backend/internal/domain/group"
	usecasechat "social-network/backend/internal/usecase/chat"
	"social-network/backend/pkg/logger"
)

func TestChatListConversations_Unauthorized(t *testing.T) {
	h := NewChatHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	rr := httptest.NewRecorder()
	h.ListConversations(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeChatRepo struct{}

func (r *fakeChatRepo) ListUserConversations(ctx context.Context, userID int64) ([]domainchat.Conversation, error) {
	return []domainchat.Conversation{{ID: 1, Type: domainchat.ConversationTypeDirect, CreatedAt: time.Now()}}, nil
}
func (r *fakeChatRepo) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	return map[int64]int{1: 1}, nil
}
func (r *fakeChatRepo) GetConversationMembers(ctx context.Context, conversationID int64) ([]int64, error) {
	return []int64{1, 2}, nil
}
func (r *fakeChatRepo) GetGroupIDByConversationID(ctx context.Context, conversationID int64) (*int64, error) {
	return nil, nil
}
func (r *fakeChatRepo) GetMessagesByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]domainchat.Message, error) {
	return []domainchat.Message{{ID: 10, ConversationID: conversationID, SenderID: 2, CreatedAt: time.Now()}}, nil
}
func (r *fakeChatRepo) IsMember(ctx context.Context, conversationID, userID int64) (bool, error) {
	return true, nil
}
func (r *fakeChatRepo) GetConversationByID(ctx context.Context, id int64) (domainchat.Conversation, error) {
	return domainchat.Conversation{}, nil
}
func (r *fakeChatRepo) GetGroupConversationID(ctx context.Context, groupID int64) (int64, error) {
	return 0, nil
}
func (r *fakeChatRepo) GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 int64) (domainchat.Conversation, error) {
	return domainchat.Conversation{}, nil
}
func (r *fakeChatRepo) AddMember(ctx context.Context, conversationID, userID int64, role domainchat.ConversationRole) error {
	return nil
}
func (r *fakeChatRepo) CreateMessage(ctx context.Context, conversationID, senderID int64, content *string, mediaPath *string) (domainchat.Message, error) {
	return domainchat.Message{}, nil
}
func (r *fakeChatRepo) GetMessageByID(ctx context.Context, id int64) (domainchat.Message, error) {
	return domainchat.Message{}, nil
}
func (r *fakeChatRepo) MarkAsRead(ctx context.Context, conversationID, userID int64) error {
	return nil
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
func (r *fakeChatRepo) ListMessageReactions(ctx context.Context, messageID int64) ([]domainchat.MessageReaction, error) {
	return nil, nil
}

type fakeChatGroupRepo struct{}

func (r *fakeChatGroupRepo) Create(ctx context.Context, creatorID int64, title string, description *string) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeChatGroupRepo) List(ctx context.Context, userID int64, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeChatGroupRepo) Search(ctx context.Context, userID int64, query string, limit, offset int) ([]domaingroup.GroupSummary, error) {
	return nil, nil
}
func (r *fakeChatGroupRepo) GetWithMeta(ctx context.Context, userID, id int64) (domaingroup.GroupSummary, error) {
	return domaingroup.GroupSummary{}, nil
}
func (r *fakeChatGroupRepo) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	return domaingroup.Group{}, nil
}
func (r *fakeChatGroupRepo) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return true, nil
}
func (r *fakeChatGroupRepo) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeChatGroupRepo) ListMembers(ctx context.Context, groupID int64) ([]domaingroup.GroupMemberInfo, error) {
	return nil, nil
}
func (r *fakeChatGroupRepo) AddMember(ctx context.Context, groupID, userID int64) error { return nil }
func (r *fakeChatGroupRepo) RemoveMember(ctx context.Context, groupID, userID int64) error {
	return nil
}
func (r *fakeChatGroupRepo) InvitationExists(ctx context.Context, groupID, inviteeID int64) (bool, error) {
	return false, nil
}
func (r *fakeChatGroupRepo) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	return domaingroup.GroupInvitation{}, nil
}
func (r *fakeChatGroupRepo) GetInvitationByID(ctx context.Context, id int64) (domaingroup.GroupInvitation, error) {
	return domaingroup.GroupInvitation{}, nil
}
func (r *fakeChatGroupRepo) ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]domaingroup.GroupInvitation, error) {
	return nil, nil
}
func (r *fakeChatGroupRepo) DeleteInvitation(ctx context.Context, id int64) error { return nil }
func (r *fakeChatGroupRepo) JoinRequestExists(ctx context.Context, groupID, userID int64) (bool, error) {
	return false, nil
}
func (r *fakeChatGroupRepo) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	return domaingroup.GroupJoinRequest{}, nil
}
func (r *fakeChatGroupRepo) GetJoinRequestByID(ctx context.Context, id int64) (domaingroup.GroupJoinRequest, error) {
	return domaingroup.GroupJoinRequest{}, nil
}
func (r *fakeChatGroupRepo) ListJoinRequestsByGroup(ctx context.Context, groupID int64) ([]domaingroup.GroupJoinRequest, error) {
	return nil, nil
}
func (r *fakeChatGroupRepo) DeleteJoinRequest(ctx context.Context, id int64) error { return nil }

type fakeChatAccess struct{}

func (f fakeChatAccess) CanSendDirectMessage(ctx context.Context, senderID, receiverID int64) (bool, error) {
	return true, nil
}
func (f fakeChatAccess) CanChatInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	return true, nil
}

func TestChatListConversations_Success(t *testing.T) {
	repo := &fakeChatRepo{}
	groupRepo := &fakeChatGroupRepo{}
	svc := usecasechat.NewService(repo, groupRepo, fakeChatAccess{}, logger.NewDefault(false))
	h := NewChatHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.ListConversations), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
