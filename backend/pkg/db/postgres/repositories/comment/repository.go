package comment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domaincomment "social-network/backend/internal/domain/comment"
)

// Repository implements comment repository using Postgres
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new comment repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new comment
func (r *Repository) Create(ctx context.Context, comment domaincomment.Comment) (domaincomment.Comment, error) {
	query := `
		INSERT INTO comments (post_id, author_id, content, media_path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.AuthorID,
		comment.Content,
		comment.MediaPath,
		now,
		now,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return domaincomment.Comment{}, fmt.Errorf("create comment: %w", err)
	}

	return comment, nil
}

// GetByPostID gets all comments for a post
func (r *Repository) GetByPostID(ctx context.Context, postID int64) ([]domaincomment.Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.author_id, c.content, c.media_path, c.created_at, c.updated_at,
		       COALESCE(rc.like_count, 0) AS like_count,
		       COALESCE(rc.dislike_count, 0) AS dislike_count
		FROM comments c
		LEFT JOIN (
			SELECT comment_id,
			       COUNT(*) FILTER (WHERE reaction = 'like') AS like_count,
			       COUNT(*) FILTER (WHERE reaction = 'dislike') AS dislike_count
			FROM comment_reactions
			GROUP BY comment_id
		) rc ON rc.comment_id = c.id
		WHERE c.post_id = $1
		ORDER BY c.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("get comments by post: %w", err)
	}
	defer rows.Close()

	var comments []domaincomment.Comment
	for rows.Next() {
		var c domaincomment.Comment
		err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.AuthorID,
			&c.Content,
			&c.MediaPath,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.LikeCount,
			&c.DislikeCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}

	return comments, nil
}

// GetByID gets a comment by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (domaincomment.Comment, error) {
	query := `
		SELECT id, post_id, author_id, content, media_path, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	var c domaincomment.Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.PostID,
		&c.AuthorID,
		&c.Content,
		&c.MediaPath,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaincomment.Comment{}, domaincomment.ErrNotFound
		}
		return domaincomment.Comment{}, fmt.Errorf("get comment: %w", err)
	}

	return c, nil
}

// Delete deletes a comment
func (r *Repository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rows == 0 {
		return domaincomment.ErrNotFound
	}

	return nil
}
