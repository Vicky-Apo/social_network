package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainchat "social-network/backend/internal/domain/chat"
)

// Repository implements the chat repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres chat repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetOrCreateDirectConversation returns an existing direct conversation between two users
// or creates a new one if it doesn't exist.
func (r *Repository) GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 int64) (domainchat.Conversation, error) {
	// Ensure consistent ordering
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}

	// Try to find existing conversation
	const findQuery = `
		SELECT c.id, c.type, c.created_at
		FROM conversations c
		JOIN conversation_members cm1 ON cm1.conversation_id = c.id AND cm1.user_id = $1
		JOIN conversation_members cm2 ON cm2.conversation_id = c.id AND cm2.user_id = $2
		WHERE c.type = 'direct'
		LIMIT 1
	`
	var conv domainchat.Conversation
	var convType string
	err := r.db.QueryRowContext(ctx, findQuery, userID1, userID2).Scan(&conv.ID, &convType, &conv.CreatedAt)
	if err == nil {
		conv.Type = domainchat.ConversationType(convType)
		return conv, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domainchat.Conversation{}, fmt.Errorf("find conversation: %w", err)
	}

	// Create new conversation with transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainchat.Conversation{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const insertConv = `
		INSERT INTO conversations (type)
		VALUES ('direct')
		RETURNING id, type, created_at
	`
	if err := tx.QueryRowContext(ctx, insertConv).Scan(&conv.ID, &convType, &conv.CreatedAt); err != nil {
		return domainchat.Conversation{}, fmt.Errorf("create conversation: %w", err)
	}
	conv.Type = domainchat.ConversationType(convType)

	const insertMember = `
		INSERT INTO conversation_members (conversation_id, user_id, role)
		VALUES ($1, $2, 'member')
	`
	if _, err := tx.ExecContext(ctx, insertMember, conv.ID, userID1); err != nil {
		return domainchat.Conversation{}, fmt.Errorf("add member 1: %w", err)
	}
	if _, err := tx.ExecContext(ctx, insertMember, conv.ID, userID2); err != nil {
		return domainchat.Conversation{}, fmt.Errorf("add member 2: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domainchat.Conversation{}, fmt.Errorf("commit: %w", err)
	}

	return conv, nil
}

// GetConversationByID returns a conversation by ID.
func (r *Repository) GetConversationByID(ctx context.Context, id int64) (domainchat.Conversation, error) {
	const query = `
		SELECT id, type, created_at
		FROM conversations
		WHERE id = $1
	`
	var conv domainchat.Conversation
	var convType string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&conv.ID, &convType, &conv.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainchat.Conversation{}, domainchat.ErrConversationNotFound
		}
		return domainchat.Conversation{}, fmt.Errorf("get conversation: %w", err)
	}
	conv.Type = domainchat.ConversationType(convType)
	return conv, nil
}

// GetGroupConversationID returns the conversation ID for a group.
func (r *Repository) GetGroupConversationID(ctx context.Context, groupID int64) (int64, error) {
	const query = `
		SELECT conversation_id
		FROM group_conversations
		WHERE group_id = $1
	`
	var convID int64
	err := r.db.QueryRowContext(ctx, query, groupID).Scan(&convID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domainchat.ErrConversationNotFound
		}
		return 0, fmt.Errorf("get group conversation: %w", err)
	}
	return convID, nil
}

// ListUserConversations returns all conversations for a user.
func (r *Repository) ListUserConversations(ctx context.Context, userID int64) ([]domainchat.Conversation, error) {
	const query = `
		SELECT c.id, c.type, c.created_at
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []domainchat.Conversation
	for rows.Next() {
		var conv domainchat.Conversation
		var convType string
		if err := rows.Scan(&conv.ID, &convType, &conv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conv.Type = domainchat.ConversationType(convType)
		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return conversations, nil
}

// IsMember checks if a user is a member of a conversation.
func (r *Repository) IsMember(ctx context.Context, conversationID, userID int64) (bool, error) {
	const query = `
		SELECT 1
		FROM conversation_members
		WHERE conversation_id = $1 AND user_id = $2
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check membership: %w", err)
	}
	return true, nil
}

// GetConversationMembers returns all member user IDs for a conversation.
func (r *Repository) GetConversationMembers(ctx context.Context, conversationID int64) ([]int64, error) {
	const query = `
		SELECT user_id
		FROM conversation_members
		WHERE conversation_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return members, nil
}

// AddMember adds a user to a conversation.
func (r *Repository) AddMember(ctx context.Context, conversationID, userID int64, role domainchat.ConversationRole) error {
	const query = `
		INSERT INTO conversation_members (conversation_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (conversation_id, user_id) DO NOTHING
	`
	if _, err := r.db.ExecContext(ctx, query, conversationID, userID, string(role)); err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

// CreateMessage creates a new message in a conversation.
func (r *Repository) CreateMessage(ctx context.Context, conversationID, senderID int64, content *string, mediaPath *string) (domainchat.Message, error) {
	const query = `
		INSERT INTO messages (conversation_id, sender_id, content, media_path)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, sender_id, content, media_path, created_at, updated_at
	`
	var msg domainchat.Message
	var msgContent, msgMedia sql.NullString
	err := r.db.QueryRowContext(ctx, query, conversationID, senderID, content, mediaPath).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msgContent,
		&msgMedia,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return domainchat.Message{}, fmt.Errorf("create message: %w", err)
	}
	if msgContent.Valid {
		msg.Content = &msgContent.String
	}
	if msgMedia.Valid {
		msg.MediaPath = &msgMedia.String
	}
	return msg, nil
}

// GetMessagesByConversation returns messages for a conversation with pagination.
func (r *Repository) GetMessagesByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]domainchat.Message, error) {
	const query = `
		SELECT id, conversation_id, sender_id, content, media_path, created_at, updated_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []domainchat.Message
	for rows.Next() {
		var msg domainchat.Message
		var msgContent, msgMedia sql.NullString
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msgContent,
			&msgMedia,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		if msgContent.Valid {
			msg.Content = &msgContent.String
		}
		if msgMedia.Valid {
			msg.MediaPath = &msgMedia.String
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return messages, nil
}

// GetMessageByID returns a message by ID.
func (r *Repository) GetMessageByID(ctx context.Context, id int64) (domainchat.Message, error) {
	const query = `
		SELECT id, conversation_id, sender_id, content, media_path, created_at, updated_at
		FROM messages
		WHERE id = $1
	`
	var msg domainchat.Message
	var msgContent, msgMedia sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msgContent,
		&msgMedia,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainchat.Message{}, domainchat.ErrMessageNotFound
		}
		return domainchat.Message{}, fmt.Errorf("get message: %w", err)
	}
	if msgContent.Valid {
		msg.Content = &msgContent.String
	}
	if msgMedia.Valid {
		msg.MediaPath = &msgMedia.String
	}
	return msg, nil
}
