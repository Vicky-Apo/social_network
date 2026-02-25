package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
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

	directKey := fmt.Sprintf("%d:%d", userID1, userID2)

	// Try to find existing conversation
	const findQuery = `
		SELECT id, type, created_at
		FROM conversations
		WHERE type = 'direct' AND direct_pair_key = $1
		LIMIT 1
	`
	var conv domainchat.Conversation
	var convType string
	err := r.db.QueryRowContext(ctx, findQuery, directKey).Scan(&conv.ID, &convType, &conv.CreatedAt)
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
		INSERT INTO conversations (type, direct_pair_key)
		VALUES ('direct', $1)
		RETURNING id, type, created_at
	`
	if err := tx.QueryRowContext(ctx, insertConv, directKey).Scan(&conv.ID, &convType, &conv.CreatedAt); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			_ = tx.Rollback()
			var existing domainchat.Conversation
			var existingType string
			if err := r.db.QueryRowContext(ctx, findQuery, directKey).Scan(&existing.ID, &existingType, &existing.CreatedAt); err != nil {
				return domainchat.Conversation{}, fmt.Errorf("find conversation after conflict: %w", err)
			}
			existing.Type = domainchat.ConversationType(existingType)
			return existing, nil
		}
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

// GetGroupIDByConversationID returns the group ID for a conversation, if any.
func (r *Repository) GetGroupIDByConversationID(ctx context.Context, conversationID int64) (*int64, error) {
	const query = `
		SELECT group_id
		FROM group_conversations
		WHERE conversation_id = $1
	`
	var groupID int64
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(&groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get group id: %w", err)
	}
	return &groupID, nil
}

// ListUserConversations returns conversations for a user with pagination.
func (r *Repository) ListUserConversations(ctx context.Context, userID int64, limit, offset int) ([]domainchat.Conversation, error) {
	const query = `
		SELECT c.id, c.type, c.created_at
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
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

// GetConversationMembersMap returns members grouped by conversation ID.
func (r *Repository) GetConversationMembersMap(ctx context.Context, conversationIDs []int64) (map[int64][]int64, error) {
	out := make(map[int64][]int64)
	if len(conversationIDs) == 0 {
		return out, nil
	}
	const query = `
		SELECT conversation_id, user_id
		FROM conversation_members
		WHERE conversation_id = ANY($1)
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(conversationIDs))
	if err != nil {
		return nil, fmt.Errorf("list conversation members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var convID, userID int64
		if err := rows.Scan(&convID, &userID); err != nil {
			return nil, fmt.Errorf("scan conversation member: %w", err)
		}
		out[convID] = append(out[convID], userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list conversation members: %w", err)
	}
	return out, nil
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

// GetLastMessages returns the latest message per conversation.
func (r *Repository) GetLastMessages(ctx context.Context, conversationIDs []int64) (map[int64]domainchat.Message, error) {
	out := make(map[int64]domainchat.Message)
	if len(conversationIDs) == 0 {
		return out, nil
	}
	const query = `
		SELECT DISTINCT ON (conversation_id)
			id, conversation_id, sender_id, content, media_path, created_at, updated_at
		FROM messages
		WHERE conversation_id = ANY($1)
		ORDER BY conversation_id, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(conversationIDs))
	if err != nil {
		return nil, fmt.Errorf("get last messages: %w", err)
	}
	defer rows.Close()

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
			return nil, fmt.Errorf("scan last message: %w", err)
		}
		if msgContent.Valid {
			msg.Content = &msgContent.String
		}
		if msgMedia.Valid {
			msg.MediaPath = &msgMedia.String
		}
		out[msg.ConversationID] = msg
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get last messages: %w", err)
	}
	return out, nil
}

// GetGroupConversationMap returns group_id keyed by conversation_id.
func (r *Repository) GetGroupConversationMap(ctx context.Context, conversationIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64)
	if len(conversationIDs) == 0 {
		return out, nil
	}
	const query = `
		SELECT conversation_id, group_id
		FROM group_conversations
		WHERE conversation_id = ANY($1)
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(conversationIDs))
	if err != nil {
		return nil, fmt.Errorf("get group conversations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var convID, groupID int64
		if err := rows.Scan(&convID, &groupID); err != nil {
			return nil, fmt.Errorf("scan group conversation: %w", err)
		}
		out[convID] = groupID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get group conversations: %w", err)
	}
	return out, nil
}

// MarkAsRead sets last_read_message_id to the latest message in the conversation for the given user.
func (r *Repository) MarkAsRead(ctx context.Context, conversationID, userID int64) error {
	const query = `
		UPDATE conversation_members
		SET last_read_message_id = (SELECT MAX(id) FROM messages WHERE conversation_id = $1)
		WHERE conversation_id = $1 AND user_id = $2
	`
	if _, err := r.db.ExecContext(ctx, query, conversationID, userID); err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

// GetUnreadCount returns the number of messages in the conversation that the user has not yet read.
func (r *Repository) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM messages m
		WHERE m.conversation_id = $1
		  AND m.id > COALESCE(
		        (SELECT last_read_message_id FROM conversation_members WHERE conversation_id = $1 AND user_id = $2),
		        0
		      )
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

// GetUnreadConversations returns a map of conversation_id → unread message count
// for every conversation the user belongs to that has at least one unread message.
func (r *Repository) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	const query = `
		SELECT cm.conversation_id, COUNT(m.id) AS unread_count
		FROM conversation_members cm
		JOIN messages m ON m.conversation_id = cm.conversation_id
		WHERE cm.user_id = $1
		  AND m.id > COALESCE(cm.last_read_message_id, 0)
		GROUP BY cm.conversation_id
		HAVING COUNT(m.id) > 0
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get unread conversations: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]int)
	for rows.Next() {
		var convID int64
		var count int
		if err := rows.Scan(&convID, &count); err != nil {
			return nil, fmt.Errorf("scan unread: %w", err)
		}
		result[convID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
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

// HasMessageReaction checks if a user already reacted with the emoji.
func (r *Repository) HasMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	const query = `
		SELECT 1
		FROM message_reactions
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, messageID, userID, emoji).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check message reaction: %w", err)
	}
	return true, nil
}

// AddMessageReaction inserts a reaction to a message.
func (r *Repository) AddMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	const query = `
		INSERT INTO message_reactions (message_id, user_id, emoji)
		VALUES ($1, $2, $3)
	`
	if _, err := r.db.ExecContext(ctx, query, messageID, userID, emoji); err != nil {
		return fmt.Errorf("add message reaction: %w", err)
	}
	return nil
}

// RemoveMessageReaction removes a reaction from a message.
func (r *Repository) RemoveMessageReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	const query = `
		DELETE FROM message_reactions
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`
	if _, err := r.db.ExecContext(ctx, query, messageID, userID, emoji); err != nil {
		return fmt.Errorf("remove message reaction: %w", err)
	}
	return nil
}

// ToggleMessageReaction atomically adds or removes a reaction.
func (r *Repository) ToggleMessageReaction(ctx context.Context, messageID, userID int64, emoji string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT id FROM messages WHERE id = $1 FOR UPDATE`, messageID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domainchat.ErrMessageNotFound
		}
		return false, fmt.Errorf("lock message: %w", err)
	}

	err = tx.QueryRowContext(ctx, `
		SELECT 1
		FROM message_reactions
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`, messageID, userID, emoji).Scan(&exists)
	if err == nil {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM message_reactions
			WHERE message_id = $1 AND user_id = $2 AND emoji = $3
		`, messageID, userID, emoji); err != nil {
			return false, fmt.Errorf("remove message reaction: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit: %w", err)
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("check message reaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO message_reactions (message_id, user_id, emoji)
		VALUES ($1, $2, $3)
	`, messageID, userID, emoji); err != nil {
		return false, fmt.Errorf("add message reaction: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return true, nil
}

// ListMessageReactions returns reactions for a message.
func (r *Repository) ListMessageReactions(ctx context.Context, messageID int64) ([]domainchat.MessageReaction, error) {
	const query = `
		SELECT message_id, user_id, emoji, created_at
		FROM message_reactions
		WHERE message_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("list message reactions: %w", err)
	}
	defer rows.Close()

	var out []domainchat.MessageReaction
	for rows.Next() {
		var rct domainchat.MessageReaction
		if err := rows.Scan(&rct.MessageID, &rct.UserID, &rct.Emoji, &rct.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message reaction: %w", err)
		}
		out = append(out, rct)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list message reactions: %w", err)
	}
	return out, nil
}
