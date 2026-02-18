package comment

import "time"

// Comment represents a comment on a post
type Comment struct {
	ID           int64
	PostID       int64
	AuthorID     int64
	AuthorFirstName string
	AuthorLastName  string
	AuthorNickname  *string
	AuthorAvatarPath *string
	Content      string
	MediaPath    string
	LikeCount    int64
	DislikeCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
