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
	query := baseSelect() + "\n" + baseFrom() + "\n" + baseJoins(1) + "\n" +
		"WHERE " + visibilityWhereForFeed(1) + "\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $2 OFFSET $3"
	rows, err := r.db.QueryContext(ctx, query, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		var mediaPath sql.NullString
		var nickname sql.NullString
		var avatarPath sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.AuthorFirstName,
			&p.AuthorLastName,
			&nickname,
			&avatarPath,
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
		p.AuthorNickname = nullableString(nickname)
		p.AuthorAvatarPath = nullableString(avatarPath)
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
		SELECT p.id, p.author_id, u.first_name, u.last_name, u.nickname, u.avatar_path,
		       p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(cc.comment_count, 0) AS comment_count,
		       COALESCE(rc.like_count, 0) AS like_count,
		       COALESCE(rc.dislike_count, 0) AS dislike_count
		FROM posts p
		JOIN users u ON u.id = p.author_id
		LEFT JOIN post_comment_counts cc ON cc.post_id = p.id
		LEFT JOIN post_reaction_counts rc ON rc.post_id = p.id
		WHERE p.id = $1
	`
	var p domainpost.Post
	var mediaPath sql.NullString
	var nickname sql.NullString
	var avatarPath sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.AuthorID,
		&p.AuthorFirstName,
		&p.AuthorLastName,
		&nickname,
		&avatarPath,
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
	p.AuthorNickname = nullableString(nickname)
	p.AuthorAvatarPath = nullableString(avatarPath)
	return p, nil
}

// Create inserts a new post and optional categories.
func (r *Repository) Create(ctx context.Context, post domainpost.Post, categoryIDs []int64, allowedUserIDs []int64) (domainpost.Post, error) {
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

	if len(allowedUserIDs) > 0 {
		const allowedQuery = `
			INSERT INTO post_allowed_users (post_id, user_id)
			VALUES ($1, $2)
		`
		for _, userID := range allowedUserIDs {
			if userID <= 0 {
				return domainpost.Post{}, fmt.Errorf("invalid allowed user id")
			}
			if _, err = tx.ExecContext(ctx, allowedQuery, post.ID, userID); err != nil {
				return domainpost.Post{}, fmt.Errorf("insert allowed user: %w", err)
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
	query := baseSelect() + "\n" + baseFrom() + "\n" + baseJoins(2) + "\n" +
		"WHERE p.author_id = $1 AND " + visibilityWhereForAuthor(3, 4) + "\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $5 OFFSET $6"
	rows, err := r.db.QueryContext(ctx, query, authorID, viewerID, isFollower, isOwner, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		var mediaPath sql.NullString
		var nickname sql.NullString
		var avatarPath sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.AuthorFirstName,
			&p.AuthorLastName,
			&nickname,
			&avatarPath,
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
		p.AuthorNickname = nullableString(nickname)
		p.AuthorAvatarPath = nullableString(avatarPath)
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	return posts, nil
}

// ListByCategory returns posts for a category with pagination.
func (r *Repository) ListByCategory(ctx context.Context, categoryID, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	query := baseSelect() + "\n" + baseFrom() + "\n" +
		"JOIN post_categories pc ON pc.post_id = p.id\n" +
		baseJoins(2) + "\n" +
		"WHERE pc.category_id = $1 AND " + visibilityWhereForFeed(2) + "\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $3 OFFSET $4"
	rows, err := r.db.QueryContext(ctx, query, categoryID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by category: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		var p domainpost.Post
		var mediaPath sql.NullString
		var nickname sql.NullString
		var avatarPath sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.AuthorFirstName,
			&p.AuthorLastName,
			&nickname,
			&avatarPath,
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
		p.AuthorNickname = nullableString(nickname)
		p.AuthorAvatarPath = nullableString(avatarPath)
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

func baseSelect() string {
	return `
		SELECT p.id, p.author_id, u.first_name, u.last_name, u.nickname, u.avatar_path,
		       p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(cc.comment_count, 0) AS comment_count,
		       COALESCE(rc.like_count, 0) AS like_count,
		       COALESCE(rc.dislike_count, 0) AS dislike_count
	`
}

func baseFrom() string {
	return `
		FROM posts p
		JOIN users u ON u.id = p.author_id
	`
}

func baseJoins(viewerParam int) string {
	return fmt.Sprintf(`
		LEFT JOIN follows f ON f.follower_id = $%d AND f.following_id = p.author_id
		LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = $%d
		LEFT JOIN post_comment_counts cc ON cc.post_id = p.id
		LEFT JOIN post_reaction_counts rc ON rc.post_id = p.id
	`, viewerParam, viewerParam)
}

func visibilityWhereForFeed(viewerParam int) string {
	return fmt.Sprintf(`
		(
			-- Always allow viewing your own posts
			p.author_id = $%d
			-- Public posts should be visible regardless of profile visibility
			OR p.visibility = 'public'
			-- Followers-only posts require an actual follow relationship
			OR (p.visibility = 'followers' AND f.follower_id IS NOT NULL)
			-- Private posts require explicit allow-list access
			OR (p.visibility = 'private' AND pau.user_id IS NOT NULL)
		)
	`, viewerParam)
}

func visibilityWhereForAuthor(isFollowerParam, isOwnerParam int) string {
	return fmt.Sprintf(`
		(
			$%d = TRUE
			OR p.visibility = 'public'
			OR ($%d = TRUE AND p.visibility = 'followers')
			OR (p.visibility = 'private' AND pau.user_id IS NOT NULL)
		)
	`, isOwnerParam, isFollowerParam)
}
