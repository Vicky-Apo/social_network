package reaction

import "time"

// ReactionDTO represents a reaction
type ReactionDTO struct {
	UserID    int64     `json:"user_id"`
	Reaction  string    `json:"reaction"`
	CreatedAt time.Time `json:"created_at"`
}

// AddReactionRequest is the request to add a reaction
type AddReactionRequest struct {
	UserID   int64  `json:"user_id"`
	Reaction string `json:"reaction"` // "like" or "dislike"
}
