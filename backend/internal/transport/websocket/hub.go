package websocket

import (
	"context"
	"sync"

	"social-network/backend/pkg/logger"
)

// PresenceNotifier provides the follow network needed to broadcast online/offline events.
type PresenceNotifier interface {
	GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error)
}

// Hub maintains the set of active clients and broadcasts messages to clients.
type Hub struct {
	// clients maps userID to a set of client connections (supports multi-device)
	clients map[int64]map[*Client]bool

	// register requests from clients
	register chan *Client

	// unregister requests from clients
	unregister chan *Client

	// done signals the hub to stop
	done chan struct{}

	// notifier provides follow-network data for presence broadcasts
	notifier PresenceNotifier

	// mutex for thread-safe access to clients map
	mu sync.RWMutex

	// logger
	log logger.Logger
}

// NewHub creates a new Hub instance.
func NewHub(notifier PresenceNotifier, log logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
		notifier:   notifier,
		log:        log.WithFields(logger.F("component", "websocket_hub")),
	}
}

// Run starts the hub's main loop processing register/unregister requests.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case <-h.done:
			h.mu.Lock()
			for userID, clients := range h.clients {
				for client := range clients {
					close(client.send)
				}
				delete(h.clients, userID)
			}
			h.mu.Unlock()
			h.log.Info("websocket hub stopped")
			return
		}
	}
}

// Stop signals the hub to close all client connections and exit.
func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	isFirstConnection := len(h.clients[client.userID]) == 0
	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]bool)
	}
	h.clients[client.userID][client] = true
	connections := len(h.clients[client.userID])
	h.mu.Unlock()

	h.log.Debug("client registered",
		logger.F("user_id", client.userID),
		logger.F("connections", connections),
	)

	// Send this client the current online status of all its contacts
	go h.sendInitialPresence(client)

	// Notify contacts that this user just came online
	if isFirstConnection {
		go h.broadcastPresence(client.userID, true)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	var (
		isLastConnection bool
		remaining        int
	)
	if clients, ok := h.clients[client.userID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
			remaining = len(clients)
			if remaining == 0 {
				delete(h.clients, client.userID)
				isLastConnection = true
			}
		}
	}
	h.mu.Unlock()

	h.log.Debug("client unregistered",
		logger.F("user_id", client.userID),
		logger.F("remaining", remaining),
	)

	// Notify contacts that this user just went offline
	if isLastConnection {
		go h.broadcastPresence(client.userID, false)
	}
}

// broadcastPresence notifies all contacts of userID that they went online or offline.
func (h *Hub) broadcastPresence(userID int64, online bool) {
	targets, err := h.notifier.GetFollowNetwork(context.Background(), userID)
	if err != nil {
		h.log.Debug("failed to get follow network for presence",
			logger.F("user_id", userID),
			logger.F("error", err.Error()),
		)
		return
	}
	if len(targets) == 0 {
		return
	}

	msgType := MessageTypeUserOffline
	if online {
		msgType = MessageTypeUserOnline
	}

	msg, err := NewWSMessage(msgType, UserPresencePayload{UserID: userID})
	if err != nil {
		h.log.Error("failed to create presence message", err)
		return
	}

	h.SendToUsers(targets, msg)
	h.log.Debug("presence broadcast",
		logger.F("user_id", userID),
		logger.F("online", online),
		logger.F("targets", len(targets)),
	)
}

// sendInitialPresence sends user_online messages to a newly connected client
// for each of its contacts that are currently online.
func (h *Hub) sendInitialPresence(client *Client) {
	targets, err := h.notifier.GetFollowNetwork(context.Background(), client.userID)
	if err != nil {
		h.log.Debug("failed to get follow network for initial presence",
			logger.F("user_id", client.userID),
			logger.F("error", err.Error()),
		)
		return
	}

	for _, targetID := range targets {
		if !h.IsUserOnline(targetID) {
			continue
		}
		msg, err := NewWSMessage(MessageTypeUserOnline, UserPresencePayload{UserID: targetID})
		if err != nil {
			continue
		}
		client.trySend(msg)
	}
}

// SendToUser sends a message to all connections for a specific user.
func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				// Client's send buffer is full, skip
				h.log.Debug("client send buffer full",
					logger.F("user_id", userID),
				)
			}
		}
	}
}

// SendToUsers sends a message to all connections for multiple users.
func (h *Hub) SendToUsers(userIDs []int64, message []byte) {
	for _, userID := range userIDs {
		h.SendToUser(userID, message)
	}
}

// IsUserOnline checks if a user has any active WebSocket connections.
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetOnlineUserCount returns the number of users with active connections.
func (h *Hub) GetOnlineUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetTotalConnectionCount returns the total number of active connections.
func (h *Hub) GetTotalConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}
