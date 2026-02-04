package follow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainfollow "social-network/backend/internal/domain/follow"
)

// Repository implements the follow repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres follow repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// IsFollowing checks if followerID follows followingID.
func (r *Repository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	const query = `
		SELECT 1
		FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, followerID, followingID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check follow: %w", err)
	}
	return true, nil
}

// RequestExists checks if a follow request already exists.
func (r *Repository) RequestExists(ctx context.Context, requesterID, targetID int64) (bool, error) {
	const query = `
		SELECT 1
		FROM follow_requests
		WHERE requester_id = $1 AND target_id = $2 AND status = 'pending'
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, requesterID, targetID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check follow request: %w", err)
	}
	return true, nil
}

// CreateRequest creates a follow request.
func (r *Repository) CreateRequest(ctx context.Context, requesterID, targetID int64) (domainfollow.FollowRequest, error) {
	const query = `
		INSERT INTO follow_requests (requester_id, target_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, requester_id, target_id, status, created_at
	`
	var req domainfollow.FollowRequest
	if err := r.db.QueryRowContext(ctx, query, requesterID, targetID).Scan(
		&req.ID,
		&req.RequesterID,
		&req.TargetID,
		&req.Status,
		&req.CreatedAt,
	); err != nil {
		return domainfollow.FollowRequest{}, fmt.Errorf("create follow request: %w", err)
	}
	return req, nil
}

// GetRequestByID returns a follow request by ID.
func (r *Repository) GetRequestByID(ctx context.Context, id int64) (domainfollow.FollowRequest, error) {
	const query = `
		SELECT id, requester_id, target_id, status, created_at
		FROM follow_requests
		WHERE id = $1
	`
	var req domainfollow.FollowRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.RequesterID,
		&req.TargetID,
		&req.Status,
		&req.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainfollow.FollowRequest{}, domainfollow.ErrRequestNotFound
		}
		return domainfollow.FollowRequest{}, fmt.Errorf("get follow request: %w", err)
	}
	return req, nil
}

// UpdateRequestStatus updates a follow request status.
func (r *Repository) UpdateRequestStatus(ctx context.Context, id int64, status string) error {
	const query = `
		UPDATE follow_requests
		SET status = $2
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update follow request: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return domainfollow.ErrRequestNotFound
	}
	return nil
}

// ListRequestsByTarget returns pending follow requests for a target user.
func (r *Repository) ListRequestsByTarget(ctx context.Context, targetID int64) ([]domainfollow.FollowRequest, error) {
	const query = `
		SELECT id, requester_id, target_id, status, created_at
		FROM follow_requests
		WHERE target_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, targetID)
	if err != nil {
		return nil, fmt.Errorf("list follow requests: %w", err)
	}
	defer rows.Close()

	var requests []domainfollow.FollowRequest
	for rows.Next() {
		var req domainfollow.FollowRequest
		if err := rows.Scan(&req.ID, &req.RequesterID, &req.TargetID, &req.Status, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("list follow requests: %w", err)
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list follow requests: %w", err)
	}
	return requests, nil
}

// ListRequestsByRequester returns pending follow requests created by a requester.
func (r *Repository) ListRequestsByRequester(ctx context.Context, requesterID int64) ([]domainfollow.FollowRequest, error) {
	const query = `
		SELECT id, requester_id, target_id, status, created_at
		FROM follow_requests
		WHERE requester_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, requesterID)
	if err != nil {
		return nil, fmt.Errorf("list follow requests: %w", err)
	}
	defer rows.Close()

	var requests []domainfollow.FollowRequest
	for rows.Next() {
		var req domainfollow.FollowRequest
		if err := rows.Scan(&req.ID, &req.RequesterID, &req.TargetID, &req.Status, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("list follow requests: %w", err)
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list follow requests: %w", err)
	}
	return requests, nil
}

// CreateFollow creates a follow relationship.
func (r *Repository) CreateFollow(ctx context.Context, followerID, followingID int64) error {
	const query = `
		INSERT INTO follows (follower_id, following_id)
		VALUES ($1, $2)
	`
	if _, err := r.db.ExecContext(ctx, query, followerID, followingID); err != nil {
		return fmt.Errorf("create follow: %w", err)
	}
	return nil
}

// GetFollowNetwork returns all user IDs connected to userID by a follow in either direction.
func (r *Repository) GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error) {
	const query = `
		SELECT following_id FROM follows WHERE follower_id = $1
		UNION
		SELECT follower_id FROM follows WHERE following_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get follow network: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan follow network: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("follow network rows: %w", err)
	}
	return ids, nil
}

// DeleteFollow removes a follow relationship.
func (r *Repository) DeleteFollow(ctx context.Context, followerID, followingID int64) error {
	const query = `
		DELETE FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`
	if _, err := r.db.ExecContext(ctx, query, followerID, followingID); err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}
	return nil
}
