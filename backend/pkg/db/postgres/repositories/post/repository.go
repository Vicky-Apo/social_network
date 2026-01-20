package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainpost "social-network/backend/internal/domain/post"
)

// Repository implements the post repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres post repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all posts ordered by creation time (newest first).
func (r *Repository) List(ctx context.Context) ([]domainpost.Post, error) {
	const query = `
		SELECT id, author_id, content, privacy, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.Privacy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	return posts, nil
}

// GetByID returns a post by ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	const query = `
		SELECT id, author_id, content, privacy, created_at, updated_at
		FROM posts
		WHERE id = $1
	`
	var p domainpost.Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.AuthorID,
		&p.Content,
		&p.Privacy,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainpost.Post{}, domainpost.ErrNotFound
		}
		return domainpost.Post{}, fmt.Errorf("get post: %w", err)
	}
	return p, nil
}
