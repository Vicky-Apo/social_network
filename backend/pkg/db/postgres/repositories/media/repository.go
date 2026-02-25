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
	const query = `
		SELECT type, id, post_id, conversation_id, user_id
		FROM (
			SELECT 'post'::text AS type, p.id, NULL::bigint AS post_id, NULL::bigint AS conversation_id, NULL::bigint AS user_id
			FROM posts p
			WHERE p.media_path = $1
			UNION ALL
			SELECT 'comment'::text AS type, c.id, c.post_id, NULL::bigint, NULL::bigint
			FROM comments c
			WHERE c.media_path = $1
			UNION ALL
			SELECT 'message'::text AS type, m.id, NULL::bigint, m.conversation_id, NULL::bigint
			FROM messages m
			WHERE m.media_path = $1
			UNION ALL
			SELECT 'avatar'::text AS type, u.id, NULL::bigint, NULL::bigint, u.id
			FROM users u
			WHERE u.avatar_path = $1
		) matches
		LIMIT 1
	`

	var typ string
	var id, postID, conversationID, userID sql.NullInt64
	if err := r.db.QueryRowContext(ctx, query, path).Scan(&typ, &id, &postID, &conversationID, &userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainmedia.MediaRef{}, domainmedia.ErrNotFound
		}
		return domainmedia.MediaRef{}, fmt.Errorf("lookup media: %w", err)
	}

	switch typ {
	case "post":
		if !id.Valid {
			return domainmedia.MediaRef{}, domainmedia.ErrNotFound
		}
		return domainmedia.MediaRef{Type: domainmedia.MediaTypePost, PostID: id.Int64}, nil
	case "comment":
		if !id.Valid || !postID.Valid {
			return domainmedia.MediaRef{}, domainmedia.ErrNotFound
		}
		return domainmedia.MediaRef{Type: domainmedia.MediaTypeComment, CommentID: id.Int64, PostID: postID.Int64}, nil
	case "message":
		if !id.Valid || !conversationID.Valid {
			return domainmedia.MediaRef{}, domainmedia.ErrNotFound
		}
		return domainmedia.MediaRef{Type: domainmedia.MediaTypeMessage, MessageID: id.Int64, ConversationID: conversationID.Int64}, nil
	case "avatar":
		if !userID.Valid {
			return domainmedia.MediaRef{}, domainmedia.ErrNotFound
		}
		return domainmedia.MediaRef{Type: domainmedia.MediaTypeAvatar, UserID: userID.Int64}, nil
	default:
		return domainmedia.MediaRef{}, domainmedia.ErrNotFound
	}
}
