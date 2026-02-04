package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domaingroup "social-network/backend/internal/domain/group"
)

// Repository implements the group repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres group repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByID returns a group by ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (domaingroup.Group, error) {
	const query = `
		SELECT id, creator_id, title, description, created_at, updated_at
		FROM groups
		WHERE id = $1
	`
	var g domaingroup.Group
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&g.ID,
		&g.CreatorID,
		&g.Title,
		&desc,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.Group{}, domaingroup.ErrGroupNotFound
		}
		return domaingroup.Group{}, fmt.Errorf("get group: %w", err)
	}
	if desc.Valid {
		g.Description = &desc.String
	}
	return g, nil
}

// IsMember checks if a user is a member of a group.
func (r *Repository) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	const query = `
		SELECT 1
		FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check membership: %w", err)
	}
	return true, nil
}

// GetMemberIDs returns all member user IDs for a group.
func (r *Repository) GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	const query = `
		SELECT user_id
		FROM group_members
		WHERE group_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return members, nil
}
