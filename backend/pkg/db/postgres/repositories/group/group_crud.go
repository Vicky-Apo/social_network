package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domaingroup "social-network/backend/internal/domain/group"
)

// List returns groups filtered by optional query.
func (r *Repository) List(ctx context.Context, query string, limit, offset int) ([]domaingroup.Group, error) {
	const stmt = `
		SELECT id, creator_id, title, description, created_at, updated_at
		FROM groups
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, stmt, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var groups []domaingroup.Group
	for rows.Next() {
		var g domaingroup.Group
		var desc sql.NullString
		if err := rows.Scan(
			&g.ID,
			&g.CreatorID,
			&g.Title,
			&desc,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		if desc.Valid {
			g.Description = &desc.String
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	return groups, nil
}

// Create inserts a new group.
func (r *Repository) Create(ctx context.Context, group domaingroup.Group) (domaingroup.Group, error) {
	const stmt = `
		INSERT INTO groups (creator_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	var desc any
	if group.Description != nil {
		desc = *group.Description
	}
	if err := r.db.QueryRowContext(ctx, stmt, group.CreatorID, group.Title, desc).Scan(
		&group.ID,
		&group.CreatedAt,
		&group.UpdatedAt,
	); err != nil {
		return domaingroup.Group{}, fmt.Errorf("create group: %w", err)
	}
	return group, nil
}

// Update updates a group.
func (r *Repository) Update(ctx context.Context, group domaingroup.Group) (domaingroup.Group, error) {
	const stmt = `
		UPDATE groups
		SET title = $1, description = $2
		WHERE id = $3
	`
	var desc any
	if group.Description != nil {
		desc = *group.Description
	}
	result, err := r.db.ExecContext(ctx, stmt, group.Title, desc, group.ID)
	if err != nil {
		return domaingroup.Group{}, fmt.Errorf("update group: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domaingroup.Group{}, fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.Group{}, domaingroup.ErrGroupNotFound
	}
	return r.GetByID(ctx, group.ID)
}

// Delete removes a group by ID.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM groups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return domaingroup.ErrGroupNotFound
	}
	return nil
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
