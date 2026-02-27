package message_reaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	domainchat "social-network/backend/internal/domain/chat"
)

// Service errors.
var (
	ErrInvalidEmoji = errors.New("invalid emoji")
	ErrForbidden    = errors.New("message reaction forbidden")
)

// Service handles message reaction business logic.
type Service struct {
	repo domainchat.Repository
}

// NewService creates a message reaction service.
func NewService(repo domainchat.Repository) *Service {
	return &Service{repo: repo}
}

// ToggleReaction adds or removes a reaction to a message.
func (s *Service) ToggleReaction(ctx context.Context, userID, messageID int64, emoji string) (string, []string, error) {
	clean := strings.TrimSpace(emoji)
	if clean == "" {
		return "", nil, ErrInvalidEmoji
	}
	if utf8.RuneCountInString(clean) > 8 {
		return "", nil, ErrInvalidEmoji
	}

	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return "", nil, err
	}

	isMember, err := s.repo.IsMember(ctx, msg.ConversationID, userID)
	if err != nil {
		return "", nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return "", nil, ErrForbidden
	}

	status, removed, err := s.repo.ToggleMessageReaction(ctx, messageID, userID, clean)
	if err != nil {
		return "", nil, err
	}
	return status, removed, nil
}

// GetRecipients returns the conversation id and member ids for a message.
func (s *Service) GetRecipients(ctx context.Context, userID, messageID int64) (int64, []int64, error) {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return 0, nil, err
	}

	isMember, err := s.repo.IsMember(ctx, msg.ConversationID, userID)
	if err != nil {
		return 0, nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return 0, nil, ErrForbidden
	}

	members, err := s.repo.GetConversationMembers(ctx, msg.ConversationID)
	if err != nil {
		return 0, nil, err
	}
	return msg.ConversationID, members, nil
}

// ListReactions returns reactions for a message.
func (s *Service) ListReactions(ctx context.Context, userID, messageID int64) ([]MessageReactionDTO, error) {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	isMember, err := s.repo.IsMember(ctx, msg.ConversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, ErrForbidden
	}

	reactions, err := s.repo.ListMessageReactions(ctx, messageID)
	if err != nil {
		return nil, err
	}

	out := make([]MessageReactionDTO, 0, len(reactions))
	for _, r := range reactions {
		out = append(out, MessageReactionDTO{
			MessageID: r.MessageID,
			UserID:    r.UserID,
			Emoji:     r.Emoji,
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}
