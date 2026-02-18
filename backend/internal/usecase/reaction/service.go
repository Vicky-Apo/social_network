package reaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
	domainreaction "social-network/backend/internal/domain/reaction"
	usecasenotification "social-network/backend/internal/usecase/notification"
)

// Service handles reaction business logic
type Service struct {
	repo        domainreaction.Repository
	postRepo    domainpost.Repository
	commentRepo domaincomment.Repository
	notifier    Notifier
	access      AccessService
}

// Notifier allows emitting notifications without coupling to transport details.
type Notifier interface {
	CreateForUser(ctx context.Context, req usecasenotification.CreateRequest) (usecasenotification.NotificationDTO, error)
}

// AccessService provides centralized access checks.
type AccessService interface {
	CanViewPost(ctx context.Context, viewerID, postID int64) (bool, error)
}

// ErrForbidden is returned when a viewer cannot access a post/comment.
var ErrForbidden = errors.New("reaction access forbidden")

// NewService creates a reaction service
func NewService(
	repo domainreaction.Repository,
	postRepo domainpost.Repository,
	commentRepo domaincomment.Repository,
	notifier Notifier,
	access AccessService,
) *Service {
	return &Service{
		repo:        repo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		notifier:    notifier,
		access:      access,
	}
}

// AddPostReaction adds a reaction to a post
func (s *Service) AddPostReaction(ctx context.Context, postID int64, req AddReactionRequest) (string, error) {
	if req.UserID <= 0 {
		return "", fmt.Errorf("invalid user id")
	}
	if s.access == nil {
		return "", errors.New("access service not configured")
	}
	canView, err := s.access.CanViewPost(ctx, req.UserID, postID)
	if err != nil {
		return "", err
	}
	if !canView {
		return "", ErrForbidden
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
		s.emitPostReactionNotification(ctx, postID, req.UserID, req.Reaction, "updated")
		return "updated", nil
	}
	s.emitPostReactionNotification(ctx, postID, req.UserID, req.Reaction, "added")
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
	if s.access == nil {
		return "", errors.New("access service not configured")
	}
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return "", err
	}
	canView, err := s.access.CanViewPost(ctx, req.UserID, comment.PostID)
	if err != nil {
		return "", err
	}
	if !canView {
		return "", ErrForbidden
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
		s.emitCommentReactionNotification(ctx, commentID, req.UserID, req.Reaction, "updated")
		return "updated", nil
	}
	s.emitCommentReactionNotification(ctx, commentID, req.UserID, req.Reaction, "added")
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

func (s *Service) emitPostReactionNotification(ctx context.Context, postID, actorID int64, reaction string, action string) {
	if s.notifier == nil || s.postRepo == nil {
		return
	}
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil || post.AuthorID == actorID {
		return
	}
	_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
		UserID:     post.AuthorID,
		ActorID:    &actorID,
		Type:       "post_reaction",
		EntityType: "post",
		EntityID:   post.ID,
		Metadata: map[string]any{
			"reaction": reaction,
			"action":   action,
		},
	})
}

func (s *Service) emitCommentReactionNotification(ctx context.Context, commentID, actorID int64, reaction string, action string) {
	if s.notifier == nil || s.commentRepo == nil {
		return
	}
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil || comment.AuthorID == actorID {
		return
	}
	_, _ = s.notifier.CreateForUser(ctx, usecasenotification.CreateRequest{
		UserID:     comment.AuthorID,
		ActorID:    &actorID,
		Type:       "comment_reaction",
		EntityType: "comment",
		EntityID:   comment.ID,
		Metadata: map[string]any{
			"reaction": reaction,
			"action":   action,
			"post_id":  comment.PostID,
		},
	})
}
