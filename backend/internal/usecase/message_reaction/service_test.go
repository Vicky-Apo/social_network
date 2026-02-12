package message_reaction

import (
	"context"
	"errors"
	"testing"

	domainchat "social-network/backend/internal/domain/chat"
)

type fakeChatRepo struct {
	message   domainchat.Message
	isMember  bool
	reactions map[[3]string]bool
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{reactions: make(map[[3]string]bool)}
}

func (r *fakeChatRepo) GetMessageByID(ctx context.Context, id int64) (domainchat.Message, error) {
	if r.message.ID == 0 {
		return domainchat.Message{}, domainchat.ErrMessageNotFound
	}
	return r.message, nil
}

func (r *fakeChatRepo) IsMember(ctx context.Context, conversationID, userID int64) (bool, error) {
	return r.isMember, nil
}

func (r *fakeChatRepo) HasMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	key := [3]string{itoa(messageID), itoa(userID), emoji}
	return r.reactions[key], nil
}

func (r *fakeChatRepo) AddMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	key := [3]string{itoa(messageID), itoa(userID), emoji}
	r.reactions[key] = true
	return nil
}

func (r *fakeChatRepo) RemoveMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	key := [3]string{itoa(messageID), itoa(userID), emoji}
	delete(r.reactions, key)
	return nil
}

func (r *fakeChatRepo) ListMessageReactions(ctx context.Context, messageID int64) ([]domainchat.MessageReaction, error) {
	return []domainchat.MessageReaction{{MessageID: messageID, UserID: 1, Emoji: "😀"}}, nil
}

// unused interface methods
func (r *fakeChatRepo) GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 int64) (domainchat.Conversation, error) {
	return domainchat.Conversation{}, nil
}
func (r *fakeChatRepo) GetConversationByID(ctx context.Context, id int64) (domainchat.Conversation, error) {
	return domainchat.Conversation{}, nil
}
func (r *fakeChatRepo) GetGroupConversationID(ctx context.Context, groupID int64) (int64, error) {
	return 0, nil
}
func (r *fakeChatRepo) GetGroupIDByConversationID(ctx context.Context, conversationID int64) (*int64, error) {
	return nil, nil
}
func (r *fakeChatRepo) ListUserConversations(ctx context.Context, userID int64) ([]domainchat.Conversation, error) {
	return nil, nil
}
func (r *fakeChatRepo) GetConversationMembers(ctx context.Context, conversationID int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeChatRepo) AddMember(ctx context.Context, conversationID, userID int64, role domainchat.ConversationRole) error {
	return nil
}
func (r *fakeChatRepo) CreateMessage(ctx context.Context, conversationID, senderID int64, content *string, mediaPath *string) (domainchat.Message, error) {
	return domainchat.Message{}, nil
}
func (r *fakeChatRepo) GetMessagesByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]domainchat.Message, error) {
	return nil, nil
}
func (r *fakeChatRepo) MarkAsRead(ctx context.Context, conversationID, userID int64) error {
	return nil
}
func (r *fakeChatRepo) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	return 0, nil
}
func (r *fakeChatRepo) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func TestToggleReaction_ForbiddenForNonMember(t *testing.T) {
	repo := newFakeChatRepo()
	repo.message = domainchat.Message{ID: 1, ConversationID: 10}
	repo.isMember = false

	svc := NewService(repo)
	_, err := svc.ToggleReaction(context.Background(), 1, 1, "😀")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestToggleReaction_AddRemove(t *testing.T) {
	repo := newFakeChatRepo()
	repo.message = domainchat.Message{ID: 1, ConversationID: 10}
	repo.isMember = true

	svc := NewService(repo)
	status, err := svc.ToggleReaction(context.Background(), 1, 1, "😀")
	if err != nil || status != "added" {
		t.Fatalf("expected added, got %v, %v", status, err)
	}
	status, err = svc.ToggleReaction(context.Background(), 1, 1, "😀")
	if err != nil || status != "removed" {
		t.Fatalf("expected removed, got %v, %v", status, err)
	}
}

func TestToggleReaction_InvalidEmoji(t *testing.T) {
	repo := newFakeChatRepo()
	repo.message = domainchat.Message{ID: 1, ConversationID: 10}
	repo.isMember = true

	svc := NewService(repo)
	_, err := svc.ToggleReaction(context.Background(), 1, 1, "")
	if !errors.Is(err, ErrInvalidEmoji) {
		t.Fatalf("expected invalid emoji, got %v", err)
	}
}

func TestToggleReaction_EmojiTooLong(t *testing.T) {
	repo := newFakeChatRepo()
	repo.message = domainchat.Message{ID: 1, ConversationID: 10}
	repo.isMember = true

	svc := NewService(repo)
	_, err := svc.ToggleReaction(context.Background(), 1, 1, "😀😀😀😀😀😀😀😀😀")
	if !errors.Is(err, ErrInvalidEmoji) {
		t.Fatalf("expected invalid emoji, got %v", err)
	}
}

func TestListReactions_Forbidden(t *testing.T) {
	repo := newFakeChatRepo()
	repo.message = domainchat.Message{ID: 1, ConversationID: 10}
	repo.isMember = false

	svc := NewService(repo)
	_, err := svc.ListReactions(context.Background(), 1, 1)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestListReactions_Success(t *testing.T) {
	repo := newFakeChatRepo()
	repo.message = domainchat.Message{ID: 1, ConversationID: 10}
	repo.isMember = true

	svc := NewService(repo)
	reactions, err := svc.ListReactions(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reactions) != 1 || reactions[0].Emoji != "😀" {
		t.Fatalf("unexpected reactions")
	}
}

func itoa(v int64) string {
	return string(rune('a' + v))
}
