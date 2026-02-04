package chat

import "time"

// ConversationType represents the type of conversation.
type ConversationType string

const (
	ConversationTypeDirect       ConversationType = "direct"
	ConversationTypePrivateGroup ConversationType = "private_group"
	ConversationTypeGroup        ConversationType = "group"
)

// ConversationRole represents a member's role in a conversation.
type ConversationRole string

const (
	RoleMember ConversationRole = "member"
	RoleAdmin  ConversationRole = "admin"
)

// Conversation represents a chat conversation.
type Conversation struct {
	ID        int64
	Type      ConversationType
	CreatedAt time.Time
}

// Message represents a chat message.
type Message struct {
	ID             int64
	ConversationID int64
	SenderID       int64
	Content        *string
	MediaPath      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ConversationMember represents a member of a conversation.
type ConversationMember struct {
	ConversationID int64
	UserID         int64
	Role           ConversationRole
	JoinedAt       time.Time
}
