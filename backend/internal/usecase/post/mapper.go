package post

import domainpost "social-network/backend/internal/domain/post"

func mapPosts(posts []domainpost.Post) []PostDTO {
	out := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, mapPost(p))
	}
	return out
}

func mapPost(p domainpost.Post) PostDTO {
	return PostDTO{
		ID:               p.ID,
		AuthorID:         p.AuthorID,
		GroupID:          p.GroupID,
		AuthorFirstName:  p.AuthorFirstName,
		AuthorLastName:   p.AuthorLastName,
		AuthorNickname:   p.AuthorNickname,
		AuthorAvatarPath: p.AuthorAvatarPath,
		Content:          p.Content,
		MediaPath:        p.MediaPath,
		Privacy:          p.Privacy,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		CommentCount:     p.CommentCount,
		LikeCount:        p.LikeCount,
		DislikeCount:     p.DislikeCount,
	}
}
