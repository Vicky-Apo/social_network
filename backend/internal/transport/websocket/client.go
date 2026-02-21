package websocket

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	usecasechat "social-network/backend/internal/usecase/chat"
	"social-network/backend/pkg/logger"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Size of the send channel buffer.
	sendBufferSize = 256
)

// Client represents a WebSocket client connection.
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	userID      int64
	send        chan []byte
	chatService ChatService
	limiter     MessageRateLimiter
	log         logger.Logger

	typingMu     sync.Mutex
	typingTimers map[int64]*time.Timer // conversation_id -> timer
	closed       bool
}

// NewClient creates a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn, userID int64, chatService ChatService, limiter MessageRateLimiter, log logger.Logger) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		userID:      userID,
		send:        make(chan []byte, sendBufferSize),
		chatService: chatService,
		limiter:     limiter,
		log:         log.WithFields(logger.F("user_id", userID)),
		typingTimers: make(map[int64]*time.Timer),
	}
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *Client) readPump() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		c.clearTypingStates()
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Debug("websocket read error", logger.F("error", err.Error()))
			}
			break
		}
		c.handleMessage(ctx, message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages.
func (c *Client) handleMessage(ctx context.Context, data []byte) {
	if !c.limiter.IsAllowed("ws:" + strconv.FormatInt(c.userID, 10)) {
		c.log.Debug("websocket rate limit exceeded")
		c.sendError("rate limit exceeded, please try again later", "RATE_LIMIT")
		return
	}

	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("invalid message format", "PARSE_ERROR")
		return
	}

	switch msg.Type {
	case MessageTypeChat:
		c.handleChatMessage(ctx, msg.Payload)
	case MessageTypeTyping:
		c.handleTypingIndicator(ctx, msg.Payload)
	case MessageTypeMarkRead:
		c.handleMarkRead(ctx, msg.Payload)
	default:
		c.sendError("unknown message type", "UNKNOWN_TYPE")
	}
}

// handleChatMessage processes a chat message.
func (c *Client) handleChatMessage(ctx context.Context, payload json.RawMessage) {
	var req usecasechat.SendMessageRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		c.sendError("invalid chat message payload", "PARSE_ERROR")
		return
	}

	msgDTO, recipientIDs, err := c.chatService.SendMessage(ctx, c.userID, req)
	if err != nil {
		c.log.Debug("failed to send message", logger.F("error", err.Error()))
		c.sendError(err.Error(), "SEND_ERROR")
		return
	}

	responseData, err := NewWSMessage(MessageTypeChat, msgDTO)
	if err != nil {
		c.log.Error("failed to create response message", err)
		return
	}

	// Send to sender (confirmation) and to recipients
	c.trySend(responseData)
	c.hub.SendToUsers(recipientIDs, responseData)
	c.pushUnreadCounts(ctx, recipientIDs, msgDTO.ConversationID)

	c.log.Debug("message sent",
		logger.F("message_id", msgDTO.ID),
		logger.F("conversation_id", msgDTO.ConversationID),
		logger.F("recipients", len(recipientIDs)),
	)
}

func (c *Client) pushUnreadCounts(ctx context.Context, recipientIDs []int64, conversationID int64) {
	for _, userID := range recipientIDs {
		unreadMap, err := c.chatService.GetUnreadConversations(ctx, userID)
		if err != nil {
			continue
		}
		count, ok := unreadMap[conversationID]
		if !ok {
			count = 0
		}
		msg, err := NewWSMessage(MessageTypeUnreadCounts, []UnreadCountItem{
			{ConversationID: conversationID, UnreadCount: count},
		})
		if err != nil {
			continue
		}
		c.hub.SendToUser(userID, msg)
	}
}

// handleTypingIndicator processes a typing indicator message.
func (c *Client) handleTypingIndicator(ctx context.Context, payload json.RawMessage) {
	var typingPayload TypingPayload
	if err := json.Unmarshal(payload, &typingPayload); err != nil {
		c.sendError("invalid typing payload", "PARSE_ERROR")
		return
	}

	recipientIDs, err := c.chatService.GetConversationRecipients(ctx, c.userID, typingPayload.ConversationID)
	if err != nil {
		// User is not authorized or conversation not found
		c.log.Debug("unauthorized typing indicator", logger.F("error", err.Error()))
		return
	}

	// Create typing message with sender ID included
	typingData, err := NewWSMessage(MessageTypeTyping, TypingIndicatorPayload{
		ConversationID: typingPayload.ConversationID,
		UserID:         c.userID,
		IsTyping:       typingPayload.IsTyping,
	})
	if err != nil {
		c.log.Error("failed to create typing message", err)
		return
	}

	// Broadcast to all conversation members except sender
	c.hub.SendToUsers(recipientIDs, typingData)

	c.log.Debug("typing indicator sent",
		logger.F("conversation_id", typingPayload.ConversationID),
		logger.F("user_id", c.userID),
		logger.F("is_typing", typingPayload.IsTyping),
		logger.F("recipients", len(recipientIDs)),
	)

	c.trackTypingTimeout(typingPayload.ConversationID, typingPayload.IsTyping)
}

func (c *Client) trackTypingTimeout(conversationID int64, isTyping bool) {
	c.typingMu.Lock()
	defer c.typingMu.Unlock()

	if c.closed {
		return
	}

	if !isTyping {
		if t, ok := c.typingTimers[conversationID]; ok {
			t.Stop()
			delete(c.typingTimers, conversationID)
		}
		return
	}

	if t, ok := c.typingTimers[conversationID]; ok {
		t.Stop()
	}

	c.typingTimers[conversationID] = time.AfterFunc(5*time.Second, func() {
		c.typingMu.Lock()
		if c.closed {
			c.typingMu.Unlock()
			return
		}
		delete(c.typingTimers, conversationID)
		c.typingMu.Unlock()

		recipientIDs, err := c.chatService.GetConversationRecipients(context.Background(), c.userID, conversationID)
		if err != nil {
			return
		}
		typingData, err := NewWSMessage(MessageTypeTyping, TypingIndicatorPayload{
			ConversationID: conversationID,
			UserID:         c.userID,
			IsTyping:       false,
		})
		if err != nil {
			return
		}
		c.hub.SendToUsers(recipientIDs, typingData)
	})
}

func (c *Client) clearTypingStates() {
	c.typingMu.Lock()
	if c.closed {
		c.typingMu.Unlock()
		return
	}
	c.closed = true
	conversations := make([]int64, 0, len(c.typingTimers))
	for convID, t := range c.typingTimers {
		t.Stop()
		conversations = append(conversations, convID)
	}
	c.typingTimers = make(map[int64]*time.Timer)
	c.typingMu.Unlock()

	for _, convID := range conversations {
		recipientIDs, err := c.chatService.GetConversationRecipients(context.Background(), c.userID, convID)
		if err != nil {
			continue
		}
		typingData, err := NewWSMessage(MessageTypeTyping, TypingIndicatorPayload{
			ConversationID: convID,
			UserID:         c.userID,
			IsTyping:       false,
		})
		if err != nil {
			continue
		}
		c.hub.SendToUsers(recipientIDs, typingData)
	}
}

// handleMarkRead processes a mark-as-read request.
func (c *Client) handleMarkRead(ctx context.Context, payload json.RawMessage) {
	var req MarkReadPayload
	if err := json.Unmarshal(payload, &req); err != nil {
		c.sendError("invalid mark_read payload", "PARSE_ERROR")
		return
	}

	if err := c.chatService.MarkAsRead(ctx, c.userID, req.ConversationID); err != nil {
		c.log.Debug("failed to mark as read", logger.F("error", err.Error()))
		c.sendError(err.Error(), "MARK_READ_ERROR")
		return
	}

	c.log.Debug("conversation marked as read",
		logger.F("conversation_id", req.ConversationID),
	)
}

// sendError sends an error message to the client.
func (c *Client) sendError(message, code string) {
	errorData, err := NewErrorMessage(message, code)
	if err != nil {
		c.log.Error("failed to create error message", err)
		return
	}
	c.trySend(errorData)
}

// trySend attempts a non-blocking send and guards against closed-channel panics.
func (c *Client) trySend(message []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.log.Debug("websocket send on closed channel")
		}
	}()
	select {
	case c.send <- message:
	default:
	}
}
