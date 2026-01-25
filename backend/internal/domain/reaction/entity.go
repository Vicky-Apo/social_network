package reaction

import "time"

// ReactionType represents like or dislike
type ReactionType string

const (
	Like    ReactionType = "like"
	Dislike ReactionType = "dislike"
)

// PostReaction represents a reaction to a post
type PostReaction struct {
	PostID    int64
	UserID    int64
	Reaction  ReactionType
	CreatedAt time.Time
}

// CommentReaction represents a reaction to a comment
type CommentReaction struct {
	CommentID int64
	UserID    int64
	Reaction  ReactionType
	CreatedAt time.Time
}
