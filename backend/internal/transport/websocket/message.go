package websocket

import (
	"encoding/json"
)

// Message types for WebSocket communication.
const (
	MessageTypeChat        = "chat_message"
	MessageTypeTyping      = "typing"
	MessageTypeError       = "error"
	MessageTypeConnected   = "connected"
	MessageTypeUserOnline  = "user_online"
	MessageTypeUserOffline = "user_offline"
)

// WSMessage represents a WebSocket message envelope.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// TypingPayload represents the payload for typing indicator sent by client.
type TypingPayload struct {
	ConversationID int64 `json:"conversation_id"`
	IsTyping       bool  `json:"is_typing"`
}

// TypingIndicatorPayload represents the payload for typing indicator broadcast to recipients.
// Includes the user_id so recipients know WHO is typing.
type TypingIndicatorPayload struct {
	ConversationID int64 `json:"conversation_id"`
	UserID         int64 `json:"user_id"`
	IsTyping       bool  `json:"is_typing"`
}

// ErrorPayload represents an error message payload.
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ConnectedPayload represents the connected message payload.
type ConnectedPayload struct {
	UserID int64  `json:"user_id"`
	Status string `json:"status"`
}

// UserPresencePayload represents an online/offline status change for a user.
type UserPresencePayload struct {
	UserID int64 `json:"user_id"`
}

// NewWSMessage creates a new WebSocket message with the given type and payload.
func NewWSMessage(msgType string, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg := WSMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}
	return json.Marshal(msg)
}

// NewErrorMessage creates a new error WebSocket message.
func NewErrorMessage(message, code string) ([]byte, error) {
	return NewWSMessage(MessageTypeError, ErrorPayload{
		Message: message,
		Code:    code,
	})
}

