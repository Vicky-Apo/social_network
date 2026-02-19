package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domaingroup "social-network/backend/internal/domain/group"
)

// CreateJoinRequest inserts a new join request.
func (r *Repository) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	const query = `
		INSERT INTO group_join_requests (group_id, user_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`
	req := domaingroup.GroupJoinRequest{
		GroupID: groupID,
		UserID:  userID,
	}
	if err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(
		&req.ID,
		&req.CreatedAt,
		&req.UpdatedAt,
	); err != nil {
		return domaingroup.GroupJoinRequest{}, fmt.Errorf("create join request: %w", err)
	}
	return req, nil
}

// GetJoinRequestByID returns a join request by ID.
func (r *Repository) GetJoinRequestByID(ctx context.Context, id int64) (domaingroup.GroupJoinRequest, error) {
	const query = `
		SELECT id, group_id, user_id, created_at, updated_at
		FROM group_join_requests
		WHERE id = $1
	`
	var req domaingroup.GroupJoinRequest
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.GroupID,
		&req.UserID,
		&req.CreatedAt,
		&req.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupJoinRequest{}, domaingroup.ErrJoinRequestNotFound
		}
		return domaingroup.GroupJoinRequest{}, fmt.Errorf("get join request: %w", err)
	}
	return req, nil
}

// ListJoinRequestsByGroup returns join requests for a group.
func (r *Repository) ListJoinRequestsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupJoinRequest, error) {
	const query = `
		SELECT id, group_id, user_id, created_at, updated_at
		FROM group_join_requests
		WHERE group_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	defer rows.Close()

	var reqs []domaingroup.GroupJoinRequest
	for rows.Next() {
		var req domaingroup.GroupJoinRequest
		if err := rows.Scan(
			&req.ID,
			&req.GroupID,
			&req.UserID,
			&req.CreatedAt,
			&req.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan join request: %w", err)
		}
		reqs = append(reqs, req)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	return reqs, nil
}

// DeleteJoinRequest removes a join request by ID.
func (r *Repository) DeleteJoinRequest(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM group_join_requests WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete join request: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.ErrJoinRequestNotFound
	}
	return nil
}
