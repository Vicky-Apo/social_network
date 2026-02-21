package chat

import (
	"context"
	"errors"
	"fmt"

	domainchat "social-network/backend/internal/domain/chat"
	domaingroup "social-network/backend/internal/domain/group"
	"social-network/backend/pkg/logger"
)

// Errors for the chat service.
var (
	ErrCannotMessage  = errors.New("cannot send message: no follow relationship")
	ErrNotGroupMember = errors.New("user is not a member of this group")
	ErrInvalidRequest = errors.New("invalid message request: must specify recipient_id or group_id")
	ErrEmptyMessage   = errors.New("message content cannot be empty")
	ErrForbidden      = errors.New("access denied to this conversation")
)

// Service orchestrates chat-related use cases.
type Service struct {
	chatRepo  domainchat.Repository
	groupRepo domaingroup.Repository
	access    AccessService
	log       logger.Logger
}

// AccessService provides centralized access checks.
type AccessService interface {
	CanSendDirectMessage(ctx context.Context, senderID, receiverID int64) (bool, error)
	CanChatInGroup(ctx context.Context, userID, groupID int64) (bool, error)
}

// NewService builds a chat service with the given repositories.
func NewService(
	chatRepo domainchat.Repository,
	groupRepo domaingroup.Repository,
	access AccessService,
	log logger.Logger,
) *Service {
	return &Service{
		chatRepo:  chatRepo,
		groupRepo: groupRepo,
		access:    access,
		log:       log.WithFields(logger.F("service", "chat")),
	}
}

// CanMessage checks if senderID can send a direct message to recipientID.
// Returns true if at least one follows the other.
func (s *Service) CanMessage(ctx context.Context, senderID, recipientID int64) (bool, error) {
	if s.access == nil {
		return false, errors.New("access service not configured")
	}
	return s.access.CanSendDirectMessage(ctx, senderID, recipientID)
}

// SendDirectMessage sends a message from sender to recipient.
func (s *Service) SendDirectMessage(ctx context.Context, senderID, recipientID int64, content *string, mediaPath *string) (MessageDTO, []int64, error) {
	if content == nil && mediaPath == nil {
		return MessageDTO{}, nil, ErrEmptyMessage
	}

	// Verify messaging permission
	canMessage, err := s.CanMessage(ctx, senderID, recipientID)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("check permission: %w", err)
	}
	if !canMessage {
		return MessageDTO{}, nil, ErrCannotMessage
	}

	// Get or create conversation
	conv, err := s.chatRepo.GetOrCreateDirectConversation(ctx, senderID, recipientID)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("get conversation: %w", err)
	}

	msg, err := s.chatRepo.CreateMessage(ctx, conv.ID, senderID, content, mediaPath)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("create message: %w", err)
	}

	// Sender has seen their own message — advance their read pointer
	if err := s.chatRepo.MarkAsRead(ctx, conv.ID, senderID); err != nil {
		s.log.Debug("failed to update sender read position",
			logger.F("sender_id", senderID),
			logger.F("conversation_id", conv.ID),
			logger.F("error", err.Error()),
		)
	}

	recipientIDs := []int64{recipientID}

	s.log.Debug("direct message sent",
		logger.F("sender_id", senderID),
		logger.F("recipient_id", recipientID),
		logger.F("message_id", msg.ID),
	)

	return mapMessage(msg), recipientIDs, nil
}

// SendGroupMessage sends a message to a group chat.
func (s *Service) SendGroupMessage(ctx context.Context, senderID, groupID int64, content *string, mediaPath *string) (MessageDTO, []int64, error) {
	if content == nil && mediaPath == nil {
		return MessageDTO{}, nil, ErrEmptyMessage
	}

	// Verify sender is a group member
	if s.access == nil {
		return MessageDTO{}, nil, errors.New("access service not configured")
	}
	isMember, err := s.access.CanChatInGroup(ctx, senderID, groupID)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return MessageDTO{}, nil, ErrNotGroupMember
	}

	// Get group conversation
	convID, err := s.chatRepo.GetGroupConversationID(ctx, groupID)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("get group conversation: %w", err)
	}

	msg, err := s.chatRepo.CreateMessage(ctx, convID, senderID, content, mediaPath)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("create message: %w", err)
	}

	// Sender has seen their own message — advance their read pointer
	if err := s.chatRepo.MarkAsRead(ctx, convID, senderID); err != nil {
		s.log.Debug("failed to update sender read position",
			logger.F("sender_id", senderID),
			logger.F("conversation_id", convID),
			logger.F("error", err.Error()),
		)
	}

	// Get all group members for WebSocket delivery (except sender)
	memberIDs, err := s.groupRepo.GetMemberIDs(ctx, groupID)
	if err != nil {
		return MessageDTO{}, nil, fmt.Errorf("get members: %w", err)
	}

	// Filter out sender
	recipientIDs := make([]int64, 0, len(memberIDs)-1)
	for _, id := range memberIDs {
		if id != senderID {
			recipientIDs = append(recipientIDs, id)
		}
	}

	s.log.Debug("group message sent",
		logger.F("sender_id", senderID),
		logger.F("group_id", groupID),
		logger.F("message_id", msg.ID),
		logger.F("recipients", len(recipientIDs)),
	)

	return mapMessage(msg), recipientIDs, nil
}

// SendMessage handles sending either a direct or group message based on the request.
func (s *Service) SendMessage(ctx context.Context, senderID int64, req SendMessageRequest) (MessageDTO, []int64, error) {
	if req.RecipientID != nil {
		return s.SendDirectMessage(ctx, senderID, *req.RecipientID, req.Content, req.MediaPath)
	}
	if req.GroupID != nil {
		return s.SendGroupMessage(ctx, senderID, *req.GroupID, req.Content, req.MediaPath)
	}
	return MessageDTO{}, nil, ErrInvalidRequest
}

// GetConversationMessages returns messages for a conversation with pagination.
func (s *Service) GetConversationMessages(ctx context.Context, userID, conversationID int64, limit, offset int) ([]MessageDTO, error) {
	// Verify user is a member
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, ErrForbidden
	}

	messages, err := s.chatRepo.GetMessagesByConversation(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	dtos := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		dtos[i] = mapMessage(msg)
	}
	return dtos, nil
}

// ListConversations returns conversations for a user with pagination.
func (s *Service) ListConversations(ctx context.Context, userID int64, limit, offset int) ([]ConversationDTO, error) {
	conversations, err := s.chatRepo.ListUserConversations(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	// Fetch all unread counts in a single query
	unreadMap, err := s.chatRepo.GetUnreadConversations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get unread counts: %w", err)
	}

	conversationIDs := make([]int64, 0, len(conversations))
	for _, conv := range conversations {
		conversationIDs = append(conversationIDs, conv.ID)
	}

	memberMap, err := s.chatRepo.GetConversationMembersMap(ctx, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("get conversation members: %w", err)
	}

	groupMap, err := s.chatRepo.GetGroupConversationMap(ctx, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("get group conversations: %w", err)
	}

	lastMap, err := s.chatRepo.GetLastMessages(ctx, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("get last messages: %w", err)
	}

	dtos := make([]ConversationDTO, len(conversations))
	for i, conv := range conversations {
		dto := ConversationDTO{
			ID:          conv.ID,
			Type:        string(conv.Type),
			UnreadCount: unreadMap[conv.ID], // 0 if not present
			CreatedAt:   conv.CreatedAt,
		}

		// For direct conversations, get the other user
		if conv.Type == domainchat.ConversationTypeDirect {
			members := memberMap[conv.ID]
			for _, memberID := range members {
				if memberID != userID {
					dto.OtherUserID = &memberID
					break
				}
			}
		} else if conv.Type == domainchat.ConversationTypeGroup {
			if groupID, ok := groupMap[conv.ID]; ok {
				dto.GroupID = &groupID
			}
		}

		// Get last message
		if msg, ok := lastMap[conv.ID]; ok {
			lastMsg := mapMessage(msg)
			dto.LastMessage = &lastMsg
		}

		dtos[i] = dto
	}
	return dtos, nil
}

// MarkAsRead advances the read pointer for a conversation to the latest message.
func (s *Service) MarkAsRead(ctx context.Context, userID, conversationID int64) error {
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return ErrForbidden
	}
	return s.chatRepo.MarkAsRead(ctx, conversationID, userID)
}

// GetUnreadConversations returns a map of conversation_id → unread count for the user.
func (s *Service) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	return s.chatRepo.GetUnreadConversations(ctx, userID)
}

// GetConversationByID returns a conversation by ID after verifying access.
func (s *Service) GetConversationByID(ctx context.Context, userID, conversationID int64) (ConversationDTO, error) {
	// Verify user is a member
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return ConversationDTO{}, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return ConversationDTO{}, ErrForbidden
	}

	conv, err := s.chatRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return ConversationDTO{}, fmt.Errorf("get conversation: %w", err)
	}

	dto := ConversationDTO{
		ID:        conv.ID,
		Type:      string(conv.Type),
		CreatedAt: conv.CreatedAt,
	}

	// For direct conversations, get the other user
	if conv.Type == domainchat.ConversationTypeDirect {
		members, err := s.chatRepo.GetConversationMembers(ctx, conv.ID)
		if err == nil {
			for _, memberID := range members {
				if memberID != userID {
					dto.OtherUserID = &memberID
					break
				}
			}
		}
	} else if conv.Type == domainchat.ConversationTypeGroup {
		if groupID, err := s.chatRepo.GetGroupIDByConversationID(ctx, conv.ID); err == nil {
			dto.GroupID = groupID
		}
	}

	return dto, nil
}

// GetConversationRecipients returns all members of a conversation except the specified userID.
// Used for broadcasting typing indicators and other real-time events.
func (s *Service) GetConversationRecipients(ctx context.Context, userID, conversationID int64) ([]int64, error) {
	// Verify user is a member
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, ErrForbidden
	}

	// Get all members
	members, err := s.chatRepo.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	// Filter out the requesting user
	recipients := make([]int64, 0, len(members)-1)
	for _, memberID := range members {
		if memberID != userID {
			recipients = append(recipients, memberID)
		}
	}

	return recipients, nil
}

func mapMessage(msg domainchat.Message) MessageDTO {
	return MessageDTO{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		MediaPath:      msg.MediaPath,
		CreatedAt:      msg.CreatedAt,
	}
}
