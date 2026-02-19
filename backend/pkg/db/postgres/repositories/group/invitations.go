package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domaingroup "social-network/backend/internal/domain/group"
)

// CreateInvitation inserts a new group invitation.
func (r *Repository) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	const query = `
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	inv := domaingroup.GroupInvitation{
		GroupID:   groupID,
		InviterID: inviterID,
		InviteeID: inviteeID,
	}
	if err := r.db.QueryRowContext(ctx, query, groupID, inviterID, inviteeID).Scan(
		&inv.ID,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	); err != nil {
		return domaingroup.GroupInvitation{}, fmt.Errorf("create invitation: %w", err)
	}
	return inv, nil
}

// GetInvitationByID returns a group invitation by ID.
func (r *Repository) GetInvitationByID(ctx context.Context, id int64) (domaingroup.GroupInvitation, error) {
	const query = `
		SELECT id, group_id, inviter_id, invitee_id, created_at, updated_at
		FROM group_invitations
		WHERE id = $1
	`
	var inv domaingroup.GroupInvitation
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID,
		&inv.GroupID,
		&inv.InviterID,
		&inv.InviteeID,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupInvitation{}, domaingroup.ErrInvitationNotFound
		}
		return domaingroup.GroupInvitation{}, fmt.Errorf("get invitation: %w", err)
	}
	return inv, nil
}

// ListInvitationsByInvitee returns invitations for an invitee.
func (r *Repository) ListInvitationsByInvitee(ctx context.Context, inviteeID int64, limit, offset int) ([]domaingroup.GroupInvitation, error) {
	const query = `
		SELECT id, group_id, inviter_id, invitee_id, created_at, updated_at
		FROM group_invitations
		WHERE invitee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, inviteeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	var invs []domaingroup.GroupInvitation
	for rows.Next() {
		var inv domaingroup.GroupInvitation
		if err := rows.Scan(
			&inv.ID,
			&inv.GroupID,
			&inv.InviterID,
			&inv.InviteeID,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invs = append(invs, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	return invs, nil
}

// ListInvitationsByGroup returns invitations for a group.
func (r *Repository) ListInvitationsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domaingroup.GroupInvitation, error) {
	const query = `
		SELECT id, group_id, inviter_id, invitee_id, created_at, updated_at
		FROM group_invitations
		WHERE group_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list group invitations: %w", err)
	}
	defer rows.Close()

	var invs []domaingroup.GroupInvitation
	for rows.Next() {
		var inv domaingroup.GroupInvitation
		if err := rows.Scan(
			&inv.ID,
			&inv.GroupID,
			&inv.InviterID,
			&inv.InviteeID,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invs = append(invs, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list group invitations: %w", err)
	}
	return invs, nil
}

// DeleteInvitation removes an invitation by ID.
func (r *Repository) DeleteInvitation(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM group_invitations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete invitation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.ErrInvitationNotFound
	}
	return nil
}
