package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domaingroup "social-network/backend/internal/domain/group"
)

// CreateEvent inserts a new event.
func (r *Repository) CreateEvent(ctx context.Context, event domaingroup.GroupEvent) (domaingroup.GroupEvent, error) {
	const query = `
		INSERT INTO events (group_id, creator_id, title, description, event_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	var desc any
	if event.Description != nil {
		desc = *event.Description
	}
	if err := r.db.QueryRowContext(ctx, query, event.GroupID, event.CreatorID, event.Title, desc, event.EventTime).Scan(
		&event.ID,
		&event.CreatedAt,
		&event.UpdatedAt,
	); err != nil {
		return domaingroup.GroupEvent{}, fmt.Errorf("create event: %w", err)
	}
	return event, nil
}

// GetEventByID returns an event by ID.
func (r *Repository) GetEventByID(ctx context.Context, id int64) (domaingroup.GroupEvent, error) {
	const query = `
		SELECT id, group_id, creator_id, title, description, event_time, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	var ev domaingroup.GroupEvent
	var desc sql.NullString
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ev.ID,
		&ev.GroupID,
		&ev.CreatorID,
		&ev.Title,
		&desc,
		&ev.EventTime,
		&ev.CreatedAt,
		&ev.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupEvent{}, domaingroup.ErrEventNotFound
		}
		return domaingroup.GroupEvent{}, fmt.Errorf("get event: %w", err)
	}
	if desc.Valid {
		ev.Description = &desc.String
	}
	return ev, nil
}

// ListEventsByGroup returns events for a group.
func (r *Repository) ListEventsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupEvent, error) {
	const query = `
		SELECT id, group_id, creator_id, title, description, event_time, created_at, updated_at
		FROM events
		WHERE group_id = $1
		ORDER BY event_time DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []domaingroup.GroupEvent
	for rows.Next() {
		var ev domaingroup.GroupEvent
		var desc sql.NullString
		if err := rows.Scan(
			&ev.ID,
			&ev.GroupID,
			&ev.CreatorID,
			&ev.Title,
			&desc,
			&ev.EventTime,
			&ev.CreatedAt,
			&ev.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		if desc.Valid {
			ev.Description = &desc.String
		}
		events = append(events, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return events, nil
}

// UpdateEvent updates an event.
func (r *Repository) UpdateEvent(ctx context.Context, event domaingroup.GroupEvent) (domaingroup.GroupEvent, error) {
	const query = `
		UPDATE events
		SET title = $1, description = $2, event_time = $3
		WHERE id = $4
	`
	var desc any
	if event.Description != nil {
		desc = *event.Description
	}
	result, err := r.db.ExecContext(ctx, query, event.Title, desc, event.EventTime, event.ID)
	if err != nil {
		return domaingroup.GroupEvent{}, fmt.Errorf("update event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domaingroup.GroupEvent{}, fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.GroupEvent{}, domaingroup.ErrEventNotFound
	}
	return r.GetEventByID(ctx, event.ID)
}

// DeleteEvent removes an event by ID.
func (r *Repository) DeleteEvent(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.ErrEventNotFound
	}
	return nil
}

// UpsertEventResponse sets or updates a user's response to an event.
func (r *Repository) UpsertEventResponse(ctx context.Context, eventID, userID int64, response string) error {
	const query = `
		INSERT INTO event_responses (event_id, user_id, response, responded_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (event_id, user_id)
		DO UPDATE SET response = EXCLUDED.response, responded_at = EXCLUDED.responded_at
	`
	respondedAt := time.Now().UTC()
	if _, err := r.db.ExecContext(ctx, query, eventID, userID, response, respondedAt); err != nil {
		return fmt.Errorf("upsert event response: %w", err)
	}
	return nil
}

// GetEventResponse returns a user's response to an event.
func (r *Repository) GetEventResponse(ctx context.Context, eventID, userID int64) (domaingroup.GroupEventResponse, error) {
	const query = `
		SELECT event_id, user_id, response, responded_at
		FROM event_responses
		WHERE event_id = $1 AND user_id = $2
	`
	var resp domaingroup.GroupEventResponse
	var respondedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, eventID, userID).Scan(
		&resp.EventID,
		&resp.UserID,
		&resp.Response,
		&respondedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupEventResponse{}, domaingroup.ErrEventResponseNotFound
		}
		return domaingroup.GroupEventResponse{}, fmt.Errorf("get event response: %w", err)
	}
	if respondedAt.Valid {
		resp.RespondedAt = &respondedAt.Time
	}
	return resp, nil
}

// ListEventResponses returns all responses for an event.
func (r *Repository) ListEventResponses(ctx context.Context, eventID int64) ([]domaingroup.GroupEventResponse, error) {
	const query = `
		SELECT event_id, user_id, response, responded_at
		FROM event_responses
		WHERE event_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("list event responses: %w", err)
	}
	defer rows.Close()

	var responses []domaingroup.GroupEventResponse
	for rows.Next() {
		var resp domaingroup.GroupEventResponse
		var respondedAt sql.NullTime
		if err := rows.Scan(
			&resp.EventID,
			&resp.UserID,
			&resp.Response,
			&respondedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event response: %w", err)
		}
		if respondedAt.Valid {
			resp.RespondedAt = &respondedAt.Time
		}
		responses = append(responses, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list event responses: %w", err)
	}
	return responses, nil
}
