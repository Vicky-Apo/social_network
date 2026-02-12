package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainevent "social-network/backend/internal/domain/event"
)

// Repository implements the event repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres event repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new event.
func (r *Repository) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	const query = `
		INSERT INTO events (group_id, creator_id, title, description, event_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	if err := r.db.QueryRowContext(ctx, query, e.GroupID, e.CreatorID, e.Title, e.Description, e.EventTime).Scan(
		&e.ID,
		&e.CreatedAt,
		&e.UpdatedAt,
	); err != nil {
		return domainevent.Event{}, fmt.Errorf("create event: %w", err)
	}
	return e, nil
}

// GetByID returns an event by ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (domainevent.Event, error) {
	const query = `
		SELECT id, group_id, creator_id, title, description, event_time, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	var e domainevent.Event
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID,
		&e.GroupID,
		&e.CreatorID,
		&e.Title,
		&desc,
		&e.EventTime,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainevent.Event{}, domainevent.ErrNotFound
		}
		return domainevent.Event{}, fmt.Errorf("get event: %w", err)
	}
	if desc.Valid {
		e.Description = &desc.String
	}
	return e, nil
}

// ListByGroup returns events for a group with pagination.
func (r *Repository) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainevent.Event, error) {
	const query = `
		SELECT id, group_id, creator_id, title, description, event_time, created_at, updated_at
		FROM events
		WHERE group_id = $1
		ORDER BY event_time ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var out []domainevent.Event
	for rows.Next() {
		var e domainevent.Event
		var desc sql.NullString
		if err := rows.Scan(
			&e.ID,
			&e.GroupID,
			&e.CreatorID,
			&e.Title,
			&desc,
			&e.EventTime,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		if desc.Valid {
			e.Description = &desc.String
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return out, nil
}

// UpsertResponse inserts or updates a user's response to an event.
func (r *Repository) UpsertResponse(ctx context.Context, eventID, userID int64, response string) (domainevent.EventResponse, error) {
	const query = `
		INSERT INTO event_responses (event_id, user_id, response, responded_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (event_id, user_id)
		DO UPDATE SET response = EXCLUDED.response, responded_at = now()
		RETURNING event_id, user_id, response, responded_at
	`
	var er domainevent.EventResponse
	var respondedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, eventID, userID, response).Scan(
		&er.EventID,
		&er.UserID,
		&er.Response,
		&respondedAt,
	); err != nil {
		return domainevent.EventResponse{}, fmt.Errorf("upsert response: %w", err)
	}
	if respondedAt.Valid {
		er.RespondedAt = &respondedAt.Time
	}
	return er, nil
}

// ListResponses returns responses for an event.
func (r *Repository) ListResponses(ctx context.Context, eventID int64) ([]domainevent.EventResponseInfo, error) {
	const query = `
		SELECT er.event_id, er.user_id, u.first_name, u.last_name, u.nickname, u.avatar_path,
		       er.response, er.responded_at
		FROM event_responses er
		JOIN users u ON u.id = er.user_id
		WHERE er.event_id = $1
		ORDER BY er.responded_at DESC NULLS LAST
	`
	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("list responses: %w", err)
	}
	defer rows.Close()

	var out []domainevent.EventResponseInfo
	for rows.Next() {
		var item domainevent.EventResponseInfo
		var nickname sql.NullString
		var avatar sql.NullString
		var respondedAt sql.NullTime
		if err := rows.Scan(
			&item.EventID,
			&item.UserID,
			&item.FirstName,
			&item.LastName,
			&nickname,
			&avatar,
			&item.Response,
			&respondedAt,
		); err != nil {
			return nil, fmt.Errorf("scan response: %w", err)
		}
		if nickname.Valid {
			item.Nickname = &nickname.String
		}
		if avatar.Valid {
			item.AvatarPath = &avatar.String
		}
		if respondedAt.Valid {
			item.RespondedAt = &respondedAt.Time
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list responses: %w", err)
	}
	return out, nil
}
