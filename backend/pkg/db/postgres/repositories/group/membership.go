package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domaingroup "social-network/backend/internal/domain/group"
)

// ListMembers returns members for a group with pagination.
func (r *Repository) ListMembers(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupMember, error) {
	const query = `
		SELECT group_id, user_id, joined_at
		FROM group_members
		WHERE group_id = $1
		ORDER BY joined_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	defer rows.Close()

	var members []domaingroup.GroupMember
	for rows.Next() {
		var m domaingroup.GroupMember
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan group member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	return members, nil
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

// AddMember inserts a group member.
func (r *Repository) AddMember(ctx context.Context, groupID, userID int64) error {
	const query = `
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`
	if _, err := r.db.ExecContext(ctx, query, groupID, userID); err != nil {
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

// RemoveMember removes a member from a group.
func (r *Repository) RemoveMember(ctx context.Context, groupID, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`, groupID, userID)
	if err != nil {
		return fmt.Errorf("remove group member: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.ErrNotMember
	}
	return nil
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
