package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainmedia "social-network/backend/internal/domain/media"
)

// Repository implements media lookups using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a media repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindByPath finds the entity that owns a given media path.
func (r *Repository) FindByPath(ctx context.Context, path string) (domainmedia.MediaRef, error) {
	var (
		id            int64
		postID        int64
		conversation  int64
	)

	// Posts
	err := r.db.QueryRowContext(ctx, `SELECT id FROM posts WHERE media_path = $1`, path).Scan(&id)
	if err == nil {
		return domainmedia.MediaRef{Type: domainmedia.MediaTypePost, PostID: id}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domainmedia.MediaRef{}, fmt.Errorf("lookup post media: %w", err)
	}

	// Comments
	err = r.db.QueryRowContext(ctx, `SELECT id, post_id FROM comments WHERE media_path = $1`, path).Scan(&id, &postID)
	if err == nil {
		return domainmedia.MediaRef{Type: domainmedia.MediaTypeComment, CommentID: id, PostID: postID}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domainmedia.MediaRef{}, fmt.Errorf("lookup comment media: %w", err)
	}

	// Messages
	err = r.db.QueryRowContext(ctx, `SELECT id, conversation_id FROM messages WHERE media_path = $1`, path).Scan(&id, &conversation)
	if err == nil {
		return domainmedia.MediaRef{Type: domainmedia.MediaTypeMessage, MessageID: id, ConversationID: conversation}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domainmedia.MediaRef{}, fmt.Errorf("lookup message media: %w", err)
	}

	// Avatars
	err = r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE avatar_path = $1`, path).Scan(&id)
	if err == nil {
		return domainmedia.MediaRef{Type: domainmedia.MediaTypeAvatar, UserID: id}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domainmedia.MediaRef{}, fmt.Errorf("lookup avatar media: %w", err)
	}

	return domainmedia.MediaRef{}, domainmedia.ErrNotFound
}
