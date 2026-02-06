package websocket

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"social-network/backend/internal/transport/http/middleware"
	usecasechat "social-network/backend/internal/usecase/chat"
	"social-network/backend/pkg/logger"
)

// ChatService defines the chat operations the WebSocket layer depends on.
type ChatService interface {
	SendMessage(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error)
	GetConversationRecipients(ctx context.Context, userID, conversationID int64) ([]int64, error)
	MarkAsRead(ctx context.Context, userID, conversationID int64) error
	GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error)
}

// MessageRateLimiter defines the rate limiting check the WebSocket layer depends on.
type MessageRateLimiter interface {
	IsAllowed(key string) bool
}

// Handler handles WebSocket upgrade requests.
type Handler struct {
	hub         *Hub
	chatService ChatService
	limiter     MessageRateLimiter
	validator   middleware.SessionValidator
	cookieName  string
	corsEnabled bool
	origins     string
	upgrader    websocket.Upgrader
	log         logger.Logger
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	hub *Hub,
	chatService ChatService,
	limiter MessageRateLimiter,
	validator middleware.SessionValidator,
	cookieName string,
	corsEnabled bool,
	allowedOrigins string,
	log logger.Logger,
) *Handler {
	h := &Handler{
		hub:         hub,
		chatService: chatService,
		limiter:     limiter,
		validator:   validator,
		cookieName:  cookieName,
		corsEnabled: corsEnabled,
		origins:     allowedOrigins,
		log:         log.WithFields(logger.F("handler", "websocket")),
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

// ServeHTTP handles the WebSocket upgrade request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract session token from cookie
	cookie, err := r.Cookie(h.cookieName)
	if err != nil {
		h.log.Debug("missing session cookie", logger.F("error", err.Error()))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate session
	userID, err := h.validator.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		h.log.Debug("invalid session", logger.F("error", err.Error()))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("failed to upgrade connection", err)
		return
	}

	// Create client
	client := NewClient(h.hub, conn, userID, h.chatService, h.limiter, h.log)

	// Register client with hub
	h.hub.register <- client

	// Send connected message
	connectedMsg, err := NewWSMessage(MessageTypeConnected, ConnectedPayload{
		UserID: userID,
		Status: "connected",
	})
	if err == nil {
		client.send <- connectedMsg
	}

	// Start client pumps in separate goroutines
	go client.writePump()
	go client.readPump()

	// Push unread counts in background so the upgrade doesn't block on a DB round-trip
	go func() {
		unreadMap, err := h.chatService.GetUnreadConversations(context.Background(), userID)
		if err != nil {
			h.log.Debug("failed to get unread conversations",
				logger.F("user_id", userID),
				logger.F("error", err.Error()),
			)
			return
		}
		if len(unreadMap) == 0 {
			return
		}

		items := make([]UnreadCountItem, 0, len(unreadMap))
		for convID, count := range unreadMap {
			items = append(items, UnreadCountItem{ConversationID: convID, UnreadCount: count})
		}

		unreadMsg, err := NewWSMessage(MessageTypeUnreadCounts, items)
		if err != nil {
			return
		}
		client.trySend(unreadMsg)
	}()

	h.log.Info("client connected",
		logger.F("user_id", userID),
		logger.F("remote_addr", r.RemoteAddr),
	)
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	if !h.corsEnabled {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	if h.origins == "*" {
		return true
	}

	for _, allowed := range strings.Split(h.origins, ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}
