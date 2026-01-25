package reaction

import "context"

// Repository defines data access for reactions
type Repository interface {
	// Post reactions
	AddPostReaction(ctx context.Context, reaction PostReaction) error
	RemovePostReaction(ctx context.Context, postID, userID int64) error
	GetPostReactions(ctx context.Context, postID int64) ([]PostReaction, error)
	
	// Comment reactions
	AddCommentReaction(ctx context.Context, reaction CommentReaction) error
	RemoveCommentReaction(ctx context.Context, commentID, userID int64) error
	GetCommentReactions(ctx context.Context, commentID int64) ([]CommentReaction, error)
}