package post

import "time"

// Post represents a social post in the domain layer.
type Post struct {
	ID               int64
	AuthorID         int64
	GroupID          *int64
	GroupTitle       *string
	AuthorFirstName  string
	AuthorLastName   string
	AuthorNickname   *string
	AuthorAvatarPath *string
	Content          string
	MediaPath        *string
	Privacy          string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CommentCount     int64
	LikeCount        int64
	DislikeCount     int64
}
