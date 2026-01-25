package reaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	domainreaction "social-network/backend/internal/domain/reaction"
)

// Repository implements reaction repository using Postgres
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new reaction repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// AddPostReaction adds or updates a reaction to a post
func (r *Repository) AddPostReaction(ctx context.Context, reaction domainreaction.PostReaction) error {
	query := `
		INSERT INTO post_reactions (post_id, user_id, reaction, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (post_id, user_id)
		DO UPDATE SET reaction = $3, created_at = $4
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reaction.PostID,
		reaction.UserID,
		reaction.Reaction,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("add post reaction: %w", err)
	}

	return nil
}

// RemovePostReaction removes a reaction from a post
func (r *Repository) RemovePostReaction(ctx context.Context, postID, userID int64) error {
	query := `DELETE FROM post_reactions WHERE post_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("remove post reaction: %w", err)
	}

	return nil
}

// GetPostReactions gets all reactions for a post
func (r *Repository) GetPostReactions(ctx context.Context, postID int64) ([]domainreaction.PostReaction, error) {
	query := `
		SELECT post_id, user_id, reaction, created_at
		FROM post_reactions
		WHERE post_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("get post reactions: %w", err)
	}
	defer rows.Close()

	var reactions []domainreaction.PostReaction
	for rows.Next() {
		var pr domainreaction.PostReaction
		var reactionStr string

		err := rows.Scan(&pr.PostID, &pr.UserID, &reactionStr, &pr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan post reaction: %w", err)
		}

		pr.Reaction = domainreaction.ReactionType(reactionStr)
		reactions = append(reactions, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get post reactions: %w", err)
	}

	return reactions, nil
}

// AddCommentReaction adds or updates a reaction to a comment
func (r *Repository) AddCommentReaction(ctx context.Context, reaction domainreaction.CommentReaction) error {
	query := `
		INSERT INTO comment_reactions (comment_id, user_id, reaction, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (comment_id, user_id)
		DO UPDATE SET reaction = $3, created_at = $4
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reaction.CommentID,
		reaction.UserID,
		reaction.Reaction,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("add comment reaction: %w", err)
	}

	return nil
}

// RemoveCommentReaction removes a reaction from a comment
func (r *Repository) RemoveCommentReaction(ctx context.Context, commentID, userID int64) error {
	query := `DELETE FROM comment_reactions WHERE comment_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, commentID, userID)
	if err != nil {
		return fmt.Errorf("remove comment reaction: %w", err)
	}

	return nil
}

// GetCommentReactions gets all reactions for a comment
func (r *Repository) GetCommentReactions(ctx context.Context, commentID int64) ([]domainreaction.CommentReaction, error) {
	query := `
		SELECT comment_id, user_id, reaction, created_at
		FROM comment_reactions
		WHERE comment_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, commentID)
	if err != nil {
		return nil, fmt.Errorf("get comment reactions: %w", err)
	}
	defer rows.Close()

	var reactions []domainreaction.CommentReaction
	for rows.Next() {
		var cr domainreaction.CommentReaction
		var reactionStr string

		err := rows.Scan(&cr.CommentID, &cr.UserID, &reactionStr, &cr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan comment reaction: %w", err)
		}

		cr.Reaction = domainreaction.ReactionType(reactionStr)
		reactions = append(reactions, cr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get comment reactions: %w", err)
	}

	return reactions, nil
}
