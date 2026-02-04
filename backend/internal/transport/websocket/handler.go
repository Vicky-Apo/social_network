package websocket

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"social-network/backend/internal/transport/http/middleware"
	usecasechat "social-network/backend/internal/usecase/chat"
	"social-network/backend/pkg/logger"
)

// ChatService defines the chat operations the WebSocket layer depends on.
type ChatService interface {
	SendMessage(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error)
	GetConversationRecipients(ctx context.Context, userID, conversationID int64) ([]int64, error)
}

// MessageRateLimiter defines the rate limiting check the WebSocket layer depends on.
type MessageRateLimiter interface {
	IsAllowed(key string) bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Configure allowed origins in production
		return true
	},
}

// Handler handles WebSocket upgrade requests.
type Handler struct {
	hub         *Hub
	chatService ChatService
	limiter     MessageRateLimiter
	validator   middleware.SessionValidator
	cookieName  string
	log         logger.Logger
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	hub *Hub,
	chatService ChatService,
	limiter MessageRateLimiter,
	validator middleware.SessionValidator,
	cookieName string,
	log logger.Logger,
) *Handler {
	return &Handler{
		hub:         hub,
		chatService: chatService,
		limiter:     limiter,
		validator:   validator,
		cookieName:  cookieName,
		log:         log.WithFields(logger.F("handler", "websocket")),
	}
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
	conn, err := upgrader.Upgrade(w, r, nil)
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

	h.log.Info("client connected",
		logger.F("user_id", userID),
		logger.F("remote_addr", r.RemoteAddr),
	)

	// Start client pumps in separate goroutines
	go client.writePump()
	go client.readPump()
}
