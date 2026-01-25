package post

import "time"

// PostDTO is the application-facing representation of a post.
type PostDTO struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	MediaPath *string   `json:"media_path,omitempty"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CommentCount int64  `json:"comment_count"`
	LikeCount    int64  `json:"like_count"`
	DislikeCount int64  `json:"dislike_count"`
}

// CreatePostRequest is the request to create a post.
type CreatePostRequest struct {
	Content     string  `json:"content"`
	MediaPath   *string `json:"media_path,omitempty"`
	Privacy     string  `json:"privacy"`
	CategoryIDs []int64 `json:"category_ids,omitempty"`
}
