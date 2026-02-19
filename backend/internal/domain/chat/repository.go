package chat

import (
	"context"
	"errors"
)

// Errors for the chat domain.
var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
	ErrNotMember            = errors.New("user is not a member of this conversation")
)

// Repository defines the data access contract for chat operations.
type Repository interface {
	// Conversations
	GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 int64) (Conversation, error)
	GetConversationByID(ctx context.Context, id int64) (Conversation, error)
	GetGroupConversationID(ctx context.Context, groupID int64) (int64, error)
	GetGroupIDByConversationID(ctx context.Context, conversationID int64) (*int64, error)
	ListUserConversations(ctx context.Context, userID int64) ([]Conversation, error)
	IsMember(ctx context.Context, conversationID, userID int64) (bool, error)
	GetConversationMembers(ctx context.Context, conversationID int64) ([]int64, error)
	GetConversationMembersMap(ctx context.Context, conversationIDs []int64) (map[int64][]int64, error)
	AddMember(ctx context.Context, conversationID, userID int64, role ConversationRole) error

	// Messages
	CreateMessage(ctx context.Context, conversationID, senderID int64, content *string, mediaPath *string) (Message, error)
	GetMessagesByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]Message, error)
	GetLastMessages(ctx context.Context, conversationIDs []int64) (map[int64]Message, error)
	GetMessageByID(ctx context.Context, id int64) (Message, error)
	HasMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error)
	AddMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error
	RemoveMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error
	ToggleMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error)
	ListMessageReactions(ctx context.Context, messageID int64) ([]MessageReaction, error)

	// Read tracking
	MarkAsRead(ctx context.Context, conversationID, userID int64) error
	GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error)
	GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error)

	// Group conversation mapping
	GetGroupConversationMap(ctx context.Context, conversationIDs []int64) (map[int64]int64, error)
}
