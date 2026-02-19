package message_reaction

import "time"

// MessageReactionDTO represents a reaction on a message.
type MessageReactionDTO struct {
	MessageID int64     `json:"message_id"`
	UserID    int64     `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// ToggleReactionRequest represents a request to toggle a reaction.
type ToggleReactionRequest struct {
	Emoji string `json:"emoji"`
}
