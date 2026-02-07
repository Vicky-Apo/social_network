package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domainnotification "social-network/backend/internal/domain/notification"
)

// Repository implements notification repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new notification repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new notification.
func (r *Repository) Create(ctx context.Context, n domainnotification.Notification) (domainnotification.Notification, error) {
	const query = `
		INSERT INTO notifications (user_id, type, entity_type, entity_id, metadata, actor_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_read, created_at, read_at, actor_id
	`

	var (
		readAt sql.NullTime
		actor  sql.NullInt64
	)
	if n.ActorID != nil {
		actor = sql.NullInt64{Int64: *n.ActorID, Valid: true}
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		n.UserID,
		string(n.Type),
		n.EntityType,
		n.EntityID,
		n.Metadata,
		actor,
	).Scan(&n.ID, &n.IsRead, &n.CreatedAt, &readAt, &actor)
	if err != nil {
		return domainnotification.Notification{}, fmt.Errorf("create notification: %w", err)
	}
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}
	if actor.Valid {
		v := actor.Int64
		n.ActorID = &v
	}
	return n, nil
}

// ListByUser returns notifications for a user.
func (r *Repository) ListByUser(ctx context.Context, userID int64, limit, offset int, unreadOnly bool) ([]domainnotification.Notification, error) {
	query := `
		SELECT id, user_id, actor_id, type, entity_type, entity_id, metadata, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []any{userID}
	if unreadOnly {
		query += " AND is_read = false"
	}
	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var out []domainnotification.Notification
	for rows.Next() {
		var n domainnotification.Notification
		var actor sql.NullInt64
		var readAt sql.NullTime
		var metadata []byte
		var typeStr string
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&actor,
			&typeStr,
			&n.EntityType,
			&n.EntityID,
			&metadata,
			&n.IsRead,
			&readAt,
			&n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		if typeStr != "" {
			n.Type = domainnotification.NotificationType(typeStr)
		}
		if actor.Valid {
			v := actor.Int64
			n.ActorID = &v
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}
		if metadata != nil {
			n.Metadata = metadata
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list notifications rows: %w", err)
	}
	return out, nil
}

// MarkRead marks a single notification as read.
func (r *Repository) MarkRead(ctx context.Context, userID, notificationID int64, readAt time.Time) error {
	const query = `
		UPDATE notifications
		SET is_read = true,
		    read_at = COALESCE(read_at, $3)
		WHERE id = $1 AND user_id = $2
	`
	res, err := r.db.ExecContext(ctx, query, notificationID, userID, readAt)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return domainnotification.ErrNotFound
	}
	return nil
}

// MarkAllRead marks all unread notifications for a user as read.
func (r *Repository) MarkAllRead(ctx context.Context, userID int64, readAt time.Time) (int64, error) {
	const query = `
		UPDATE notifications
		SET is_read = true,
		    read_at = COALESCE(read_at, $2)
		WHERE user_id = $1 AND is_read = false
	`
	res, err := r.db.ExecContext(ctx, query, userID, readAt)
	if err != nil {
		return 0, fmt.Errorf("mark all read: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("mark all read rows: %w", err)
	}
	return rows, nil
}

// UnreadCount returns the number of unread notifications for a user.
func (r *Repository) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`
	var count int64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}
