package comment

import domaincomment "social-network/backend/internal/domain/comment"

func mapComments(comments []domaincomment.Comment) []CommentDTO {
	out := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		out = append(out, mapComment(c))
	}
	return out
}

func mapComment(c domaincomment.Comment) CommentDTO {
	return CommentDTO{
		ID:           c.ID,
		PostID:       c.PostID,
		AuthorID:     c.AuthorID,
		AuthorFirstName: c.AuthorFirstName,
		AuthorLastName:  c.AuthorLastName,
		AuthorNickname:  c.AuthorNickname,
		AuthorAvatarPath: c.AuthorAvatarPath,
		Content:      c.Content,
		MediaPath:    c.MediaPath,
		LikeCount:    c.LikeCount,
		DislikeCount: c.DislikeCount,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
