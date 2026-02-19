package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainchat "social-network/backend/internal/domain/chat"
	usecasemessagereaction "social-network/backend/internal/usecase/message_reaction"
	"social-network/backend/pkg/logger"
)

func TestMessageReactionToggle_Unauthorized(t *testing.T) {
	h := NewMessageReactionHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/messages/1/reactions", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.Toggle(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeMessageReactionChatRepo struct{}

func (r *fakeMessageReactionChatRepo) GetMessageByID(ctx context.Context, id int64) (domainchat.Message, error) {
	return domainchat.Message{ID: id, ConversationID: 10}, nil
}
func (r *fakeMessageReactionChatRepo) IsMember(ctx context.Context, conversationID, userID int64) (bool, error) {
	return true, nil
}
func (r *fakeMessageReactionChatRepo) HasMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	return false, nil
}
func (r *fakeMessageReactionChatRepo) AddMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	return nil
}
func (r *fakeMessageReactionChatRepo) RemoveMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	return nil
}
func (r *fakeMessageReactionChatRepo) ToggleMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	return true, nil
}
func (r *fakeMessageReactionChatRepo) ListMessageReactions(ctx context.Context, messageID int64) ([]domainchat.MessageReaction, error) {
	return []domainchat.MessageReaction{{MessageID: messageID, UserID: 2, Emoji: "😀", CreatedAt: time.Now()}}, nil
}
func (r *fakeMessageReactionChatRepo) GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 int64) (domainchat.Conversation, error) {
	return domainchat.Conversation{}, nil
}
func (r *fakeMessageReactionChatRepo) GetConversationByID(ctx context.Context, id int64) (domainchat.Conversation, error) {
	return domainchat.Conversation{}, nil
}
func (r *fakeMessageReactionChatRepo) GetGroupConversationID(ctx context.Context, groupID int64) (int64, error) {
	return 0, nil
}
func (r *fakeMessageReactionChatRepo) GetGroupIDByConversationID(ctx context.Context, conversationID int64) (*int64, error) {
	return nil, nil
}
func (r *fakeMessageReactionChatRepo) GetGroupConversationMap(ctx context.Context, conversationIDs []int64) (map[int64]int64, error) {
	return map[int64]int64{}, nil
}
func (r *fakeMessageReactionChatRepo) ListUserConversations(ctx context.Context, userID int64) ([]domainchat.Conversation, error) {
	return nil, nil
}
func (r *fakeMessageReactionChatRepo) GetConversationMembers(ctx context.Context, conversationID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeMessageReactionChatRepo) GetConversationMembersMap(ctx context.Context, conversationIDs []int64) (map[int64][]int64, error) {
	return map[int64][]int64{}, nil
}
func (r *fakeMessageReactionChatRepo) AddMember(ctx context.Context, conversationID, userID int64, role domainchat.ConversationRole) error {
	return nil
}
func (r *fakeMessageReactionChatRepo) CreateMessage(ctx context.Context, conversationID, senderID int64, content *string, mediaPath *string) (domainchat.Message, error) {
	return domainchat.Message{}, nil
}
func (r *fakeMessageReactionChatRepo) GetMessagesByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]domainchat.Message, error) {
	return nil, nil
}
func (r *fakeMessageReactionChatRepo) GetLastMessages(ctx context.Context, conversationIDs []int64) (map[int64]domainchat.Message, error) {
	return map[int64]domainchat.Message{}, nil
}
func (r *fakeMessageReactionChatRepo) MarkAsRead(ctx context.Context, conversationID, userID int64) error {
	return nil
}
func (r *fakeMessageReactionChatRepo) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	return 0, nil
}
func (r *fakeMessageReactionChatRepo) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func TestMessageReactionList_Success(t *testing.T) {
	svc := usecasemessagereaction.NewService(&fakeMessageReactionChatRepo{})
	h := NewMessageReactionHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/messages/1/reactions", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.List), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
