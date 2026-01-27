package reaction

import (
	"context"
	"database/sql"
	"fmt"

	domainreaction "social-network/backend/internal/domain/reaction"
)

// Service handles reaction business logic
type Service struct {
	repo domainreaction.Repository
}

// NewService creates a reaction service
func NewService(repo domainreaction.Repository) *Service {
	return &Service{repo: repo}
}

// AddPostReaction adds a reaction to a post
func (s *Service) AddPostReaction(ctx context.Context, postID int64, req AddReactionRequest) (string, error) {
	if req.UserID <= 0 {
		return "", fmt.Errorf("invalid user id")
	}
	reactionType := domainreaction.ReactionType(req.Reaction)

	if reactionType != domainreaction.Like && reactionType != domainreaction.Dislike {
		return "", fmt.Errorf("invalid reaction type: %s", req.Reaction)
	}

	existing, err := s.repo.GetPostReaction(ctx, postID, req.UserID)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err == nil {
		if existing.Reaction == reactionType {
			if err := s.repo.RemovePostReaction(ctx, postID, req.UserID); err != nil {
				return "", err
			}
			return "removed", nil
		}
	}

	reaction := domainreaction.PostReaction{
		PostID:   postID,
		UserID:   req.UserID,
		Reaction: reactionType,
	}

	if err := s.repo.AddPostReaction(ctx, reaction); err != nil {
		return "", err
	}
	// Check if we updated an existing reaction or added a new one
	if existing.Reaction != "" {
		return "updated", nil
	}
	return "added", nil
}

// RemovePostReaction removes a reaction from a post
func (s *Service) RemovePostReaction(ctx context.Context, postID, userID int64) error {
	return s.repo.RemovePostReaction(ctx, postID, userID)
}

// GetPostReactions gets all reactions for a post
func (s *Service) GetPostReactions(ctx context.Context, postID int64) ([]ReactionDTO, error) {
	reactions, err := s.repo.GetPostReactions(ctx, postID)
	if err != nil {
		return nil, err
	}

	return mapPostReactions(reactions), nil
}

// AddCommentReaction adds a reaction to a comment
func (s *Service) AddCommentReaction(ctx context.Context, commentID int64, req AddReactionRequest) (string, error) {
	if req.UserID <= 0 {
		return "", fmt.Errorf("invalid user id")
	}
	reactionType := domainreaction.ReactionType(req.Reaction)

	if reactionType != domainreaction.Like && reactionType != domainreaction.Dislike {
		return "", fmt.Errorf("invalid reaction type: %s", req.Reaction)
	}

	existing, err := s.repo.GetCommentReaction(ctx, commentID, req.UserID)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err == nil {
		if existing.Reaction == reactionType {
			if err := s.repo.RemoveCommentReaction(ctx, commentID, req.UserID); err != nil {
				return "", err
			}
			return "removed", nil
		}
	}

	reaction := domainreaction.CommentReaction{
		CommentID: commentID,
		UserID:    req.UserID,
		Reaction:  reactionType,
	}

	if err := s.repo.AddCommentReaction(ctx, reaction); err != nil {
		return "", err
	}
	// Check if we updated an existing reaction or added a new one
	if existing.Reaction != "" {
		return "updated", nil
	}
	return "added", nil
}

// GetCommentReactions gets all reactions for a comment
func (s *Service) GetCommentReactions(ctx context.Context, commentID int64) ([]ReactionDTO, error) {
	reactions, err := s.repo.GetCommentReactions(ctx, commentID)
	if err != nil {
		return nil, err
	}

	return mapCommentReactions(reactions), nil
}

func mapPostReactions(reactions []domainreaction.PostReaction) []ReactionDTO {
	out := make([]ReactionDTO, 0, len(reactions))
	for _, r := range reactions {
		out = append(out, ReactionDTO{
			UserID:    r.UserID,
			Reaction:  string(r.Reaction),
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	return out
}

func mapCommentReactions(reactions []domainreaction.CommentReaction) []ReactionDTO {
	out := make([]ReactionDTO, 0, len(reactions))
	for _, r := range reactions {
		out = append(out, ReactionDTO{
			UserID:    r.UserID,
			Reaction:  string(r.Reaction),
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	return out
}
