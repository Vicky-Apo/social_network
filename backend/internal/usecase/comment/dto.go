package comment

import "time"

// CommentDTO is the API representation of a comment
type CommentDTO struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	MediaPath string    `json:"media_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCommentRequest is the request to create a comment
type CreateCommentRequest struct {
	PostID   int64  `json:"post_id"`
	AuthorID int64  `json:"author_id"`
	Content  string `json:"content"`
	MediaPath string `json:"media_path,omitempty"`
}
