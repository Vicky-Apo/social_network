package chat

import "time"

// SendMessageRequest represents a request to send a message.
type SendMessageRequest struct {
	RecipientID *int64  `json:"recipient_id,omitempty"` // For direct messages
	GroupID     *int64  `json:"group_id,omitempty"`     // For group messages
	Content     *string `json:"content,omitempty"`
	MediaPath   *string `json:"media_path,omitempty"`
}

// MessageDTO represents a message response.
type MessageDTO struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	Content        *string   `json:"content,omitempty"`
	MediaPath      *string   `json:"media_path,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationDTO represents a conversation response.
type ConversationDTO struct {
	ID           int64        `json:"id"`
	Type         string       `json:"type"`
	OtherUserID  *int64       `json:"other_user_id,omitempty"`  // For direct conversations
	GroupID      *int64       `json:"group_id,omitempty"`       // For group conversations
	LastMessage  *MessageDTO  `json:"last_message,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

// ConversationWithMessagesDTO includes conversation details and messages.
type ConversationWithMessagesDTO struct {
	Conversation ConversationDTO `json:"conversation"`
	Messages     []MessageDTO    `json:"messages"`
}
