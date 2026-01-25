package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
func (r *Repository) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	const query = `
		SELECT p.id, p.author_id, p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(c.comment_count, 0) AS comment_count,
		       COALESCE(l.like_count, 0) AS like_count,
		       COALESCE(d.dislike_count, 0) AS dislike_count
		FROM posts p
		JOIN users u ON u.id = p.author_id
		LEFT JOIN follows f ON f.follower_id = $1 AND f.following_id = p.author_id
		LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = $1
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		) c ON c.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS like_count
			FROM post_reactions
			WHERE reaction = 'like'
			GROUP BY post_id
		) l ON l.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS dislike_count
			FROM post_reactions
			WHERE reaction = 'dislike'
			GROUP BY post_id
		) d ON d.post_id = p.id
		WHERE (
			u.is_public = TRUE
			OR p.author_id = $1
			OR f.follower_id IS NOT NULL
		)
		AND (
			p.author_id = $1
			OR p.visibility = 'public'
			OR (p.visibility = 'followers' AND f.follower_id IS NOT NULL)
			OR (p.visibility = 'private' AND pau.user_id IS NOT NULL)
		)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		var mediaPath sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Content,
			&mediaPath,
			&p.Privacy,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.CommentCount,
			&p.LikeCount,
			&p.DislikeCount,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		p.MediaPath = nullableString(mediaPath)
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
		SELECT p.id, p.author_id, p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(c.comment_count, 0) AS comment_count,
		       COALESCE(l.like_count, 0) AS like_count,
		       COALESCE(d.dislike_count, 0) AS dislike_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		) c ON c.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS like_count
			FROM post_reactions
			WHERE reaction = 'like'
			GROUP BY post_id
		) l ON l.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS dislike_count
			FROM post_reactions
			WHERE reaction = 'dislike'
			GROUP BY post_id
		) d ON d.post_id = p.id
		WHERE p.id = $1
	`
	var p domainpost.Post
	var mediaPath sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.AuthorID,
		&p.Content,
		&mediaPath,
		&p.Privacy,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.CommentCount,
		&p.LikeCount,
		&p.DislikeCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainpost.Post{}, domainpost.ErrNotFound
		}
		return domainpost.Post{}, fmt.Errorf("get post: %w", err)
	}
	p.MediaPath = nullableString(mediaPath)
	return p, nil
}

// Create inserts a new post and optional categories.
func (r *Repository) Create(ctx context.Context, post domainpost.Post, categoryIDs []int64) (domainpost.Post, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainpost.Post{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const query = `
		INSERT INTO posts (author_id, content, media_path, visibility)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	var mediaPath any
	if post.MediaPath != nil && strings.TrimSpace(*post.MediaPath) != "" {
		mediaPath = *post.MediaPath
	}

	if err = tx.QueryRowContext(ctx, query, post.AuthorID, post.Content, mediaPath, post.Privacy).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		return domainpost.Post{}, fmt.Errorf("create post: %w", err)
	}

	if len(categoryIDs) > 0 {
		const catQuery = `
			INSERT INTO post_categories (post_id, category_id)
			VALUES ($1, $2)
		`
		for _, categoryID := range categoryIDs {
			if categoryID <= 0 {
				return domainpost.Post{}, fmt.Errorf("invalid category id")
			}
			if _, err = tx.ExecContext(ctx, catQuery, post.ID, categoryID); err != nil {
				return domainpost.Post{}, fmt.Errorf("insert post category: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return domainpost.Post{}, fmt.Errorf("commit post: %w", err)
	}

	return post, nil
}

// ListByAuthor returns posts for a specific author with pagination.
func (r *Repository) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	const query = `
		SELECT p.id, p.author_id, p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(c.comment_count, 0) AS comment_count,
		       COALESCE(l.like_count, 0) AS like_count,
		       COALESCE(d.dislike_count, 0) AS dislike_count
		FROM posts p
		LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = $2
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		) c ON c.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS like_count
			FROM post_reactions
			WHERE reaction = 'like'
			GROUP BY post_id
		) l ON l.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS dislike_count
			FROM post_reactions
			WHERE reaction = 'dislike'
			GROUP BY post_id
		) d ON d.post_id = p.id
		WHERE p.author_id = $1 AND (
			$4 = TRUE
			OR p.visibility = 'public'
			OR ($3 = TRUE AND p.visibility = 'followers')
			OR (p.visibility = 'private' AND pau.user_id IS NOT NULL)
		)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.db.QueryContext(ctx, query, authorID, viewerID, isFollower, isOwner, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		var mediaPath sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Content,
			&mediaPath,
			&p.Privacy,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.CommentCount,
			&p.LikeCount,
			&p.DislikeCount,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		p.MediaPath = nullableString(mediaPath)
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	return posts, nil
}

// ListByCategory returns posts for a category with pagination.
func (r *Repository) ListByCategory(ctx context.Context, categoryID, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	const query = `
		SELECT p.id, p.author_id, p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(c.comment_count, 0) AS comment_count,
		       COALESCE(l.like_count, 0) AS like_count,
		       COALESCE(d.dislike_count, 0) AS dislike_count
		FROM posts p
		JOIN post_categories pc ON pc.post_id = p.id
		JOIN users u ON u.id = p.author_id
		LEFT JOIN follows f ON f.follower_id = $2 AND f.following_id = p.author_id
		LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = $2
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		) c ON c.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS like_count
			FROM post_reactions
			WHERE reaction = 'like'
			GROUP BY post_id
		) l ON l.post_id = p.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS dislike_count
			FROM post_reactions
			WHERE reaction = 'dislike'
			GROUP BY post_id
		) d ON d.post_id = p.id
		WHERE pc.category_id = $1
		AND (
			u.is_public = TRUE
			OR p.author_id = $2
			OR f.follower_id IS NOT NULL
		)
		AND (
			p.author_id = $2
			OR p.visibility = 'public'
			OR (p.visibility = 'followers' AND f.follower_id IS NOT NULL)
			OR (p.visibility = 'private' AND pau.user_id IS NOT NULL)
		)
		ORDER BY p.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, categoryID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by category: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		var mediaPath sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Content,
			&mediaPath,
			&p.Privacy,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.CommentCount,
			&p.LikeCount,
			&p.DislikeCount,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		p.MediaPath = nullableString(mediaPath)
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts by category: %w", err)
	}
	return posts, nil
}

// IsUserAllowed checks if a user is in the allowed list for a private post.
func (r *Repository) IsUserAllowed(ctx context.Context, postID, userID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM post_allowed_users
			WHERE post_id = $1 AND user_id = $2
		)
	`
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, postID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check post allowed users: %w", err)
	}
	return exists, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
