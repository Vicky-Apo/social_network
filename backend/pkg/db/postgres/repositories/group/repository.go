package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

// Create creates a group, its initial member, and the group conversation.
func (r *Repository) Create(ctx context.Context, creatorID int64, title string, description *string) (domaingroup.Group, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domaingroup.Group{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const insertGroup = `
		INSERT INTO groups (creator_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id, creator_id, title, description, created_at, updated_at
	`
	var g domaingroup.Group
	var desc sql.NullString
	if err := tx.QueryRowContext(ctx, insertGroup, creatorID, title, description).Scan(
		&g.ID,
		&g.CreatorID,
		&g.Title,
		&desc,
		&g.CreatedAt,
		&g.UpdatedAt,
	); err != nil {
		return domaingroup.Group{}, fmt.Errorf("create group: %w", err)
	}
	if desc.Valid {
		g.Description = &desc.String
	}

	const insertMember = `
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
	`
	if _, err := tx.ExecContext(ctx, insertMember, g.ID, creatorID); err != nil {
		return domaingroup.Group{}, fmt.Errorf("add creator member: %w", err)
	}

	const insertConversation = `
		INSERT INTO conversations (type)
		VALUES ('group')
		RETURNING id
	`
	var conversationID int64
	if err := tx.QueryRowContext(ctx, insertConversation).Scan(&conversationID); err != nil {
		return domaingroup.Group{}, fmt.Errorf("create group conversation: %w", err)
	}

	const insertGroupConversation = `
		INSERT INTO group_conversations (group_id, conversation_id)
		VALUES ($1, $2)
	`
	if _, err := tx.ExecContext(ctx, insertGroupConversation, g.ID, conversationID); err != nil {
		return domaingroup.Group{}, fmt.Errorf("link group conversation: %w", err)
	}

	const insertConversationMember = `
		INSERT INTO conversation_members (conversation_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`
	if _, err := tx.ExecContext(ctx, insertConversationMember, conversationID, creatorID); err != nil {
		return domaingroup.Group{}, fmt.Errorf("add conversation member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domaingroup.Group{}, fmt.Errorf("commit: %w", err)
	}

	return g, nil
}

// List returns groups with membership metadata.
func (r *Repository) List(ctx context.Context, userID int64, limit, offset int) ([]domaingroup.GroupSummary, error) {
	const query = `
		SELECT g.id, g.creator_id, g.title, g.description, g.created_at, g.updated_at,
		       COALESCE(m.member_count, 0) AS member_count,
		       CASE WHEN gm.user_id IS NULL THEN false ELSE true END AS is_member
		FROM groups g
		LEFT JOIN (
			SELECT group_id, COUNT(*) AS member_count
			FROM group_members
			GROUP BY group_id
		) m ON m.group_id = g.id
		LEFT JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = $1
		ORDER BY g.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var out []domaingroup.GroupSummary
	for rows.Next() {
		item, err := scanGroupSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	return out, nil
}

// Search returns groups filtered by query with membership metadata.
func (r *Repository) Search(ctx context.Context, userID int64, queryText string, limit, offset int) ([]domaingroup.GroupSummary, error) {
	const query = `
		SELECT g.id, g.creator_id, g.title, g.description, g.created_at, g.updated_at,
		       COALESCE(m.member_count, 0) AS member_count,
		       CASE WHEN gm.user_id IS NULL THEN false ELSE true END AS is_member
		FROM groups g
		LEFT JOIN (
			SELECT group_id, COUNT(*) AS member_count
			FROM group_members
			GROUP BY group_id
		) m ON m.group_id = g.id
		LEFT JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = $1
		WHERE g.title ILIKE $2 OR g.description ILIKE $2
		ORDER BY g.created_at DESC
		LIMIT $3 OFFSET $4
	`
	pattern := "%" + strings.TrimSpace(queryText) + "%"
	rows, err := r.db.QueryContext(ctx, query, userID, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search groups: %w", err)
	}
	defer rows.Close()

	var out []domaingroup.GroupSummary
	for rows.Next() {
		item, err := scanGroupSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search groups: %w", err)
	}
	return out, nil
}

// GetWithMeta returns a group by ID with member count and membership flag.
func (r *Repository) GetWithMeta(ctx context.Context, userID, id int64) (domaingroup.GroupSummary, error) {
	const query = `
		SELECT g.id, g.creator_id, g.title, g.description, g.created_at, g.updated_at,
		       COALESCE(m.member_count, 0) AS member_count,
		       CASE WHEN gm.user_id IS NULL THEN false ELSE true END AS is_member
		FROM groups g
		LEFT JOIN (
			SELECT group_id, COUNT(*) AS member_count
			FROM group_members
			GROUP BY group_id
		) m ON m.group_id = g.id
		LEFT JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = $1
		WHERE g.id = $2
	`
	row := r.db.QueryRowContext(ctx, query, userID, id)
	item, err := scanGroupSummary(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupSummary{}, domaingroup.ErrGroupNotFound
		}
		return domaingroup.GroupSummary{}, fmt.Errorf("get group: %w", err)
	}
	return item, nil
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

// ListMembers returns member info for a group.
func (r *Repository) ListMembers(ctx context.Context, groupID int64) ([]domaingroup.GroupMemberInfo, error) {
	const query = `
		SELECT u.id, u.first_name, u.last_name, u.nickname, u.avatar_path, gm.joined_at
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var out []domaingroup.GroupMemberInfo
	for rows.Next() {
		var item domaingroup.GroupMemberInfo
		var nickname sql.NullString
		var avatar sql.NullString
		if err := rows.Scan(&item.UserID, &item.FirstName, &item.LastName, &nickname, &avatar, &item.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		if nickname.Valid {
			item.Nickname = &nickname.String
		}
		if avatar.Valid {
			item.AvatarPath = &avatar.String
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return out, nil
}

// AddMember adds a user to a group and the group conversation.
func (r *Repository) AddMember(ctx context.Context, groupID, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const insertMember = `
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
	`
	if _, err := tx.ExecContext(ctx, insertMember, groupID, userID); err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	const getConversation = `
		SELECT conversation_id
		FROM group_conversations
		WHERE group_id = $1
	`
	var conversationID int64
	if err := tx.QueryRowContext(ctx, getConversation, groupID).Scan(&conversationID); err != nil {
		return fmt.Errorf("get group conversation: %w", err)
	}

	const insertConversationMember = `
		INSERT INTO conversation_members (conversation_id, user_id, role)
		VALUES ($1, $2, 'member')
	`
	if _, err := tx.ExecContext(ctx, insertConversationMember, conversationID, userID); err != nil {
		return fmt.Errorf("add conversation member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// RemoveMember removes a user from a group and the group conversation.
func (r *Repository) RemoveMember(ctx context.Context, groupID, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const deleteMember = `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`
	res, err := tx.ExecContext(ctx, deleteMember, groupID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return domaingroup.ErrNotMember
	}

	const getConversation = `
		SELECT conversation_id
		FROM group_conversations
		WHERE group_id = $1
	`
	var conversationID int64
	if err := tx.QueryRowContext(ctx, getConversation, groupID).Scan(&conversationID); err != nil {
		return fmt.Errorf("get group conversation: %w", err)
	}

	const deleteConversationMember = `
		DELETE FROM conversation_members
		WHERE conversation_id = $1 AND user_id = $2
	`
	if _, err := tx.ExecContext(ctx, deleteConversationMember, conversationID, userID); err != nil {
		return fmt.Errorf("remove conversation member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// InvitationExists checks if an invitation exists for a user and group.
func (r *Repository) InvitationExists(ctx context.Context, groupID, inviteeID int64) (bool, error) {
	const query = `
		SELECT 1
		FROM group_invitations
		WHERE group_id = $1 AND invitee_id = $2
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, groupID, inviteeID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check invitation: %w", err)
	}
	return true, nil
}

// CreateInvitation creates a group invitation.
func (r *Repository) CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (domaingroup.GroupInvitation, error) {
	const query = `
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id)
		VALUES ($1, $2, $3)
		RETURNING id, group_id, inviter_id, invitee_id, created_at, updated_at
	`
	var inv domaingroup.GroupInvitation
	if err := r.db.QueryRowContext(ctx, query, groupID, inviterID, inviteeID).Scan(
		&inv.ID,
		&inv.GroupID,
		&inv.InviterID,
		&inv.InviteeID,
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
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID,
		&inv.GroupID,
		&inv.InviterID,
		&inv.InviteeID,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupInvitation{}, domaingroup.ErrInvitationNotFound
		}
		return domaingroup.GroupInvitation{}, fmt.Errorf("get invitation: %w", err)
	}
	return inv, nil
}

// ListInvitationsByInvitee returns invitations for an invitee.
func (r *Repository) ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]domaingroup.GroupInvitation, error) {
	const query = `
		SELECT id, group_id, inviter_id, invitee_id, created_at, updated_at
		FROM group_invitations
		WHERE invitee_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, inviteeID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	var out []domaingroup.GroupInvitation
	for rows.Next() {
		var inv domaingroup.GroupInvitation
		if err := rows.Scan(&inv.ID, &inv.GroupID, &inv.InviterID, &inv.InviteeID, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		out = append(out, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	return out, nil
}

// DeleteInvitation deletes a group invitation.
func (r *Repository) DeleteInvitation(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM group_invitations
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete invitation: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return domaingroup.ErrInvitationNotFound
	}
	return nil
}

// JoinRequestExists checks if a join request exists.
func (r *Repository) JoinRequestExists(ctx context.Context, groupID, userID int64) (bool, error) {
	const query = `
		SELECT 1
		FROM group_join_requests
		WHERE group_id = $1 AND user_id = $2
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check join request: %w", err)
	}
	return true, nil
}

// CreateJoinRequest creates a group join request.
func (r *Repository) CreateJoinRequest(ctx context.Context, groupID, userID int64) (domaingroup.GroupJoinRequest, error) {
	const query = `
		INSERT INTO group_join_requests (group_id, user_id)
		VALUES ($1, $2)
		RETURNING id, group_id, user_id, created_at, updated_at
	`
	var req domaingroup.GroupJoinRequest
	if err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(
		&req.ID,
		&req.GroupID,
		&req.UserID,
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
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.GroupID,
		&req.UserID,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domaingroup.GroupJoinRequest{}, domaingroup.ErrJoinRequestNotFound
		}
		return domaingroup.GroupJoinRequest{}, fmt.Errorf("get join request: %w", err)
	}
	return req, nil
}

// ListJoinRequestsByGroup returns join requests for a group.
func (r *Repository) ListJoinRequestsByGroup(ctx context.Context, groupID int64) ([]domaingroup.GroupJoinRequest, error) {
	const query = `
		SELECT id, group_id, user_id, created_at, updated_at
		FROM group_join_requests
		WHERE group_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	defer rows.Close()

	var out []domaingroup.GroupJoinRequest
	for rows.Next() {
		var req domaingroup.GroupJoinRequest
		if err := rows.Scan(&req.ID, &req.GroupID, &req.UserID, &req.CreatedAt, &req.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan join request: %w", err)
		}
		out = append(out, req)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	return out, nil
}

// DeleteJoinRequest deletes a join request.
func (r *Repository) DeleteJoinRequest(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM group_join_requests
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete join request: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return domaingroup.ErrJoinRequestNotFound
	}
	return nil
}

func scanGroupSummary(scanner interface{ Scan(dest ...any) error }) (domaingroup.GroupSummary, error) {
	var item domaingroup.GroupSummary
	var desc sql.NullString
	if err := scanner.Scan(
		&item.ID,
		&item.CreatorID,
		&item.Title,
		&desc,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.MemberCount,
		&item.IsMember,
	); err != nil {
		return domaingroup.GroupSummary{}, err
	}
	if desc.Valid {
		item.Description = &desc.String
	}
	return item, nil
}
