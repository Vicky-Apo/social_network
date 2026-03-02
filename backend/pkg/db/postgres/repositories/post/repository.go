package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domainpost "social-network/backend/internal/domain/post"
	reposhared "social-network/backend/pkg/db/postgres/repositories/shared"
)

// Repository implements the post repository using Postgres.
type Repository struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

// NewRepository builds a new Postgres post repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func scanPost(scanner rowScanner) (domainpost.Post, error) {
	var p domainpost.Post
	var groupID sql.NullInt64
	var groupTitle sql.NullString
	var mediaPath sql.NullString
	var nickname sql.NullString
	var avatarPath sql.NullString
	if err := scanner.Scan(
		&p.ID,
		&p.AuthorID,
		&groupID,
		&groupTitle,
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
		return domainpost.Post{}, err
	}
	p.GroupID = nullableInt64(groupID)
	p.GroupTitle = reposhared.NullableString(groupTitle)
	p.MediaPath = reposhared.NullableString(mediaPath)
	p.AuthorNickname = reposhared.NullableString(nickname)
	p.AuthorAvatarPath = reposhared.NullableString(avatarPath)
	return p, nil
}

// List returns all posts ordered by creation time (newest first).
func (r *Repository) List(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	query := visibilityCTEForFeed(1) + "\n" +
		baseSelect() + "\n" + baseFrom() + "\n" + baseJoinsCounts() + "\n" +
		"JOIN visible_posts vp ON vp.id = p.id\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $2 OFFSET $3"
	rows, err := r.db.QueryContext(ctx, query, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	return posts, nil
}

// Count returns total posts visible in the feed for a viewer.
func (r *Repository) Count(ctx context.Context, viewerID int64) (int, error) {
	query := visibilityCTEForFeed(1) + `
		SELECT COUNT(*)
		FROM visible_posts
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query, viewerID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count posts: %w", err)
	}
	return total, nil
}

// ListPublicOnly returns only public personal posts (no group posts). Visible to all users.
func (r *Repository) ListPublicOnly(ctx context.Context, limit, offset int) ([]domainpost.Post, error) {
	query := baseSelect() + "\n" + baseFrom() + "\n" +
		"LEFT JOIN post_comment_counts cc ON cc.post_id = p.id\n" +
		"LEFT JOIN post_reaction_counts rc ON rc.post_id = p.id\n" +
		"WHERE p.group_id IS NULL AND p.visibility = 'public'\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $1 OFFSET $2"
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list public posts: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list public posts: %w", err)
	}
	return posts, nil
}

// CountPublicOnly returns total public personal posts (no group posts).
func (r *Repository) CountPublicOnly(ctx context.Context) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM posts p
		WHERE p.group_id IS NULL AND p.visibility = 'public'
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, fmt.Errorf("count public posts: %w", err)
	}
	return total, nil
}

// ListGroupsOnly returns only group posts for groups the viewer is a member of.
func (r *Repository) ListGroupsOnly(ctx context.Context, viewerID int64, limit, offset int) ([]domainpost.Post, error) {
	query := baseSelect() + "\n" + baseFrom() + "\n" + baseJoinsCounts() + "\n" +
		"JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = $1\n" +
		"WHERE p.group_id IS NOT NULL\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $2 OFFSET $3"
	rows, err := r.db.QueryContext(ctx, query, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list group posts: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list group posts: %w", err)
	}
	return posts, nil
}

// CountGroupsOnly returns total group posts for groups the viewer is a member of.
func (r *Repository) CountGroupsOnly(ctx context.Context, viewerID int64) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM posts p
		JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = $1
		WHERE p.group_id IS NOT NULL
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query, viewerID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count group posts: %w", err)
	}
	return total, nil
}

// GetByID returns a post by ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (domainpost.Post, error) {
	const query = `
		SELECT p.id, p.author_id, p.group_id, g.title AS group_title, u.first_name, u.last_name, u.nickname, u.avatar_path,
		       p.content, p.media_path, p.visibility, p.created_at, p.updated_at,
		       COALESCE(cc.comment_count, 0) AS comment_count,
		       COALESCE(rc.like_count, 0) AS like_count,
		       COALESCE(rc.dislike_count, 0) AS dislike_count
		FROM posts p
		JOIN users u ON u.id = p.author_id
		LEFT JOIN groups g ON g.id = p.group_id
		LEFT JOIN post_comment_counts cc ON cc.post_id = p.id
		LEFT JOIN post_reaction_counts rc ON rc.post_id = p.id
		WHERE p.id = $1
	`
	var p domainpost.Post
	p, err := scanPost(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainpost.Post{}, domainpost.ErrNotFound
		}
		return domainpost.Post{}, fmt.Errorf("get post: %w", err)
	}
	return p, nil
}

// Create inserts a new post.
func (r *Repository) Create(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainpost.Post{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO posts (author_id, group_id, content, media_path, visibility)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	var mediaPath any
	if post.MediaPath != nil && strings.TrimSpace(*post.MediaPath) != "" {
		mediaPath = *post.MediaPath
	}
	var groupID any
	if post.GroupID != nil {
		groupID = *post.GroupID
	}

	if err = tx.QueryRowContext(ctx, query, post.AuthorID, groupID, post.Content, mediaPath, post.Privacy).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		return domainpost.Post{}, fmt.Errorf("create post: %w", err)
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

// Update updates an existing post and its allowed users.
func (r *Repository) Update(ctx context.Context, post domainpost.Post, allowedUserIDs []int64) (domainpost.Post, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainpost.Post{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const query = `
		UPDATE posts
		SET content = $1,
		    media_path = $2,
		    visibility = $3,
		    updated_at = NOW()
		WHERE id = $4
	`
	var mediaPath any
	if post.MediaPath != nil && strings.TrimSpace(*post.MediaPath) != "" {
		mediaPath = *post.MediaPath
	}
	res, err := tx.ExecContext(ctx, query, post.Content, mediaPath, post.Privacy, post.ID)
	if err != nil {
		return domainpost.Post{}, fmt.Errorf("update post: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return domainpost.Post{}, fmt.Errorf("update post: %w", err)
	}
	if rows == 0 {
		return domainpost.Post{}, domainpost.ErrNotFound
	}

	// Refresh allowed users
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_allowed_users WHERE post_id = $1`, post.ID); err != nil {
		return domainpost.Post{}, fmt.Errorf("delete allowed users: %w", err)
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

	updated, err := r.GetByID(ctx, post.ID)
	if err != nil {
		return domainpost.Post{}, err
	}
	return updated, nil
}

// Delete removes a post by ID.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM posts WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if rows == 0 {
		return domainpost.ErrNotFound
	}
	return nil
}

// ListByAuthor returns posts for a specific author with pagination.
func (r *Repository) ListByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool, limit, offset int) ([]domainpost.Post, error) {
	query := visibilityCTEForAuthor(2, 3, 4) + "\n" +
		baseSelect() + "\n" + baseFrom() + "\n" + baseJoinsCounts() + "\n" +
		"JOIN visible_posts vp ON vp.id = p.id\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $5 OFFSET $6"
	rows, err := r.db.QueryContext(ctx, query, authorID, viewerID, isFollower, isOwner, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	return posts, nil
}

// CountByAuthor returns total posts for a specific author respecting viewer access.
func (r *Repository) CountByAuthor(ctx context.Context, authorID, viewerID int64, isFollower, isOwner bool) (int, error) {
	query := visibilityCTEForAuthor(2, 3, 4) + `
		SELECT COUNT(*)
		FROM visible_posts
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query, authorID, viewerID, isFollower, isOwner).Scan(&total); err != nil {
		return 0, fmt.Errorf("count posts by author: %w", err)
	}
	return total, nil
}

// ListByGroup returns posts for a group with pagination.
func (r *Repository) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]domainpost.Post, error) {
	query := baseSelect() + "\n" + baseFrom() + "\n" +
		"LEFT JOIN post_comment_counts cc ON cc.post_id = p.id\n" +
		"LEFT JOIN post_reaction_counts rc ON rc.post_id = p.id\n" +
		"WHERE p.group_id = $1\n" +
		"ORDER BY p.created_at DESC\n" +
		"LIMIT $2 OFFSET $3"

	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by group: %w", err)
	}
	defer rows.Close()

	var posts []domainpost.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts by group: %w", err)
	}
	return posts, nil
}

// CountByGroup returns total posts for a group.
func (r *Repository) CountByGroup(ctx context.Context, groupID int64) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM posts p
		WHERE p.group_id = $1
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query, groupID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count posts by group: %w", err)
	}
	return total, nil
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

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func baseSelect() string {
	return `
		SELECT p.id, p.author_id, p.group_id, g.title AS group_title, u.first_name, u.last_name, u.nickname, u.avatar_path,
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
		LEFT JOIN groups g ON g.id = p.group_id
	`
}

func baseJoinsCounts() string {
	return `
		LEFT JOIN post_comment_counts cc ON cc.post_id = p.id
		LEFT JOIN post_reaction_counts rc ON rc.post_id = p.id
	`
}

func visibilityWhereForFeed(viewerParam int) string {
	return fmt.Sprintf(`
		(
			p.visibility = 'public'
			OR (
				(
					u.is_public = TRUE
					OR p.author_id = $%d
					OR f.follower_id IS NOT NULL
				)
				AND (
					p.author_id = $%d
					OR (p.visibility = 'followers' AND f.follower_id IS NOT NULL)
					OR (p.visibility = 'private' AND pau.user_id IS NOT NULL)
				)
			)
		)
	`, viewerParam, viewerParam)
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

func visibilityCTEForFeed(viewerParam int) string {
	return fmt.Sprintf(`
		WITH visible_posts AS (
			SELECT p.id
			FROM posts p
			JOIN users u ON u.id = p.author_id
			LEFT JOIN follows f ON f.follower_id = $%d AND f.following_id = p.author_id
			LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = $%d
			LEFT JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = $%d
			WHERE (p.group_id IS NULL AND %s)
			   OR (p.group_id IS NOT NULL AND gm.user_id IS NOT NULL)
		)
	`, viewerParam, viewerParam, viewerParam, visibilityWhereForFeed(viewerParam))
}

func visibilityCTEForAuthor(viewerParam, isFollowerParam, isOwnerParam int) string {
	return fmt.Sprintf(`
		WITH visible_posts AS (
			SELECT p.id
			FROM posts p
			JOIN users u ON u.id = p.author_id
			LEFT JOIN follows f ON f.follower_id = $%d AND f.following_id = p.author_id
			LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = $%d
			LEFT JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = $%d
			WHERE p.author_id = $1 AND (
				(p.group_id IS NULL AND %s)
				OR (p.group_id IS NOT NULL AND gm.user_id IS NOT NULL)
			)
		)
	`, viewerParam, viewerParam, viewerParam, visibilityWhereForAuthor(isFollowerParam, isOwnerParam))
}
