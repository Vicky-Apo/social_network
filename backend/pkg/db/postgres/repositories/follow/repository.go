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
		WHERE requester_id = $1 AND target_id = $2
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
		INSERT INTO follow_requests (requester_id, target_id)
		VALUES ($1, $2)
		RETURNING id, requester_id, target_id, created_at, updated_at
	`
	var req domainfollow.FollowRequest
	if err := r.db.QueryRowContext(ctx, query, requesterID, targetID).Scan(
		&req.ID,
		&req.RequesterID,
		&req.TargetID,
		&req.CreatedAt,
		&req.UpdatedAt,
	); err != nil {
		return domainfollow.FollowRequest{}, fmt.Errorf("create follow request: %w", err)
	}
	return req, nil
}

// GetRequestByID returns a follow request by ID.
func (r *Repository) GetRequestByID(ctx context.Context, id int64) (domainfollow.FollowRequest, error) {
	const query = `
		SELECT id, requester_id, target_id, created_at, updated_at
		FROM follow_requests
		WHERE id = $1
	`
	var req domainfollow.FollowRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.RequesterID,
		&req.TargetID,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainfollow.FollowRequest{}, domainfollow.ErrRequestNotFound
		}
		return domainfollow.FollowRequest{}, fmt.Errorf("get follow request: %w", err)
	}
	return req, nil
}

// DeleteRequest deletes a follow request by ID.
func (r *Repository) DeleteRequest(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM follow_requests
		WHERE id = $1
	`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete follow request: %w", err)
	}
	return nil
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
