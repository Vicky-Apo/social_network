package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	usecasechat "social-network/backend/internal/usecase/chat"
	usecasenotification "social-network/backend/internal/usecase/notification"
	"social-network/backend/pkg/logger"
)

type fakeValidator struct {
	tokens map[string]int64
}

func (f fakeValidator) ValidateSession(ctx context.Context, token string) (int64, error) {
	if id, ok := f.tokens[token]; ok {
		return id, nil
	}
	return 0, context.Canceled
}

type fakeLimiter struct{}

func (f fakeLimiter) IsAllowed(key string) bool { return true }

type denyLimiter struct{}

func (d denyLimiter) IsAllowed(key string) bool { return false }

type fakePresence struct {
	network map[int64][]int64
}

func (p fakePresence) GetFollowNetwork(ctx context.Context, userID int64) ([]int64, error) {
	return p.network[userID], nil
}

type fakeChatService struct {
	sendFunc   func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error)
	recipients map[int64][]int64
	unreads    map[int64]map[int64]int
}

func (f *fakeChatService) SendMessage(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
	return f.sendFunc(ctx, senderID, req)
}
func (f *fakeChatService) GetConversationRecipients(ctx context.Context, userID, conversationID int64) ([]int64, error) {
	return f.recipients[conversationID], nil
}
func (f *fakeChatService) MarkAsRead(ctx context.Context, userID, conversationID int64) error {
	return nil
}
func (f *fakeChatService) GetUnreadConversations(ctx context.Context, userID int64) (map[int64]int, error) {
	if f.unreads == nil {
		return map[int64]int{}, nil
	}
	return f.unreads[userID], nil
}

func TestWebSocketChat_DirectMessageDelivered(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			content := "hi"
			return usecasechat.MessageDTO{
				ID:             1,
				ConversationID: 100,
				SenderID:       senderID,
				Content:        &content,
				CreatedAt:      time.Now(),
			}, []int64{2}, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1, "t2": 2}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	c2 := dialWS(t, server.URL, "t2")
	defer c2.Close()

	consumeUntilType(t, c1, MessageTypeConnected)
	consumeUntilType(t, c2, MessageTypeConnected)

	payload := usecasechat.SendMessageRequest{RecipientID: ptrInt64(2), Content: ptrString("hello")}
	msg := mustWSMessage(t, MessageTypeChat, payload)
	if err := c1.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	got1 := consumeUntilType(t, c1, MessageTypeChat)
	got2 := consumeUntilType(t, c2, MessageTypeChat)

	if got1.Type != MessageTypeChat || got2.Type != MessageTypeChat {
		t.Fatalf("expected chat messages")
	}
}

func TestWebSocketChat_GroupMessageDelivered(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			content := "group"
			return usecasechat.MessageDTO{
				ID:             2,
				ConversationID: 200,
				SenderID:       senderID,
				Content:        &content,
				CreatedAt:      time.Now(),
			}, []int64{2, 3}, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1, "t2": 2, "t3": 3}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	c2 := dialWS(t, server.URL, "t2")
	defer c2.Close()
	c3 := dialWS(t, server.URL, "t3")
	defer c3.Close()

	consumeUntilType(t, c1, MessageTypeConnected)
	consumeUntilType(t, c2, MessageTypeConnected)
	consumeUntilType(t, c3, MessageTypeConnected)

	payload := usecasechat.SendMessageRequest{GroupID: ptrInt64(10), Content: ptrString("hi group")}
	msg := mustWSMessage(t, MessageTypeChat, payload)
	if err := c1.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	consumeUntilType(t, c2, MessageTypeChat)
	consumeUntilType(t, c3, MessageTypeChat)
}

func TestWebSocketChat_UnreadCountsPushed(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			content := "hi"
			return usecasechat.MessageDTO{
				ID:             3,
				ConversationID: 300,
				SenderID:       senderID,
				Content:        &content,
				CreatedAt:      time.Now(),
			}, []int64{2}, nil
		},
		unreads: map[int64]map[int64]int{
			2: {300: 1},
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1, "t2": 2}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	c2 := dialWS(t, server.URL, "t2")
	defer c2.Close()

	consumeUntilType(t, c1, MessageTypeConnected)
	consumeUntilType(t, c2, MessageTypeConnected)

	payload := usecasechat.SendMessageRequest{RecipientID: ptrInt64(2), Content: ptrString("hello")}
	msg := mustWSMessage(t, MessageTypeChat, payload)
	if err := c1.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	consumeUntilTypes(t, c2, MessageTypeChat, MessageTypeUnreadCounts)
}

func TestWebSocketNotificationPublisher(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			return usecasechat.MessageDTO{}, nil, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	consumeUntilType(t, c1, MessageTypeConnected)

	pub := NewNotificationPublisher(hub)
	if err := pub.Publish(context.Background(), 1, sampleNotificationDTO()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	consumeUntilType(t, c1, MessageTypeNotification)
}

func TestWebSocketTypingIndicatorBroadcast(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		recipients: map[int64][]int64{
			400: {2},
		},
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			return usecasechat.MessageDTO{}, nil, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1, "t2": 2}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	c2 := dialWS(t, server.URL, "t2")
	defer c2.Close()

	consumeUntilType(t, c1, MessageTypeConnected)
	consumeUntilType(t, c2, MessageTypeConnected)

	payload := TypingPayload{ConversationID: 400, IsTyping: true}
	msg := mustWSMessage(t, MessageTypeTyping, payload)
	if err := c1.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	typing := consumeUntilType(t, c2, MessageTypeTyping)
	var indicator TypingIndicatorPayload
	if err := json.Unmarshal(typing.Payload, &indicator); err != nil {
		t.Fatalf("unmarshal typing payload: %v", err)
	}
	if indicator.UserID != 1 || indicator.ConversationID != 400 || !indicator.IsTyping {
		t.Fatalf("unexpected typing payload")
	}
}

func TestWebSocketMarkRead(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			return usecasechat.MessageDTO{}, nil, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()

	consumeUntilType(t, c1, MessageTypeConnected)

	payload := MarkReadPayload{ConversationID: 500}
	msg := mustWSMessage(t, MessageTypeMarkRead, payload)
	if err := c1.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	// No error should be sent back.
	ensureNoType(t, c1, MessageTypeError)
}

func TestWebSocketPresenceOnlineOffline(t *testing.T) {
	presence := fakePresence{
		network: map[int64][]int64{
			1: {2},
		},
	}
	hub := NewHub(presence, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			return usecasechat.MessageDTO{}, nil, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1, "t2": 2}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c2 := dialWS(t, server.URL, "t2")
	defer c2.Close()
	consumeUntilType(t, c2, MessageTypeConnected)

	c1 := dialWS(t, server.URL, "t1")
	consumeUntilType(t, c1, MessageTypeConnected)

	online := consumeUntilType(t, c2, MessageTypeUserOnline)
	var payload UserPresencePayload
	if err := json.Unmarshal(online.Payload, &payload); err != nil {
		t.Fatalf("unmarshal online payload: %v", err)
	}
	if payload.UserID != 1 {
		t.Fatalf("expected user 1 online")
	}

	_ = c1.Close()
	offline := consumeUntilType(t, c2, MessageTypeUserOffline)
	if err := json.Unmarshal(offline.Payload, &payload); err != nil {
		t.Fatalf("unmarshal offline payload: %v", err)
	}
	if payload.UserID != 1 {
		t.Fatalf("expected user 1 offline")
	}
}

func TestWebSocketRateLimitError(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			return usecasechat.MessageDTO{}, nil, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		denyLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	consumeUntilType(t, c1, MessageTypeConnected)

	payload := usecasechat.SendMessageRequest{RecipientID: ptrInt64(2), Content: ptrString("hello")}
	msg := mustWSMessage(t, MessageTypeChat, payload)
	if err := c1.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	errMsg := consumeUntilType(t, c1, MessageTypeError)
	var errPayload ErrorPayload
	if err := json.Unmarshal(errMsg.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload.Code != "RATE_LIMIT" {
		t.Fatalf("expected RATE_LIMIT code, got %s", errPayload.Code)
	}
}

func TestWebSocketInvalidPayloadError(t *testing.T) {
	hub := NewHub(fakePresence{}, logger.NewDefault(false))
	go hub.Run()
	t.Cleanup(hub.Stop)

	chatSvc := &fakeChatService{
		sendFunc: func(ctx context.Context, senderID int64, req usecasechat.SendMessageRequest) (usecasechat.MessageDTO, []int64, error) {
			return usecasechat.MessageDTO{}, nil, nil
		},
	}

	handler := NewHandler(
		hub,
		chatSvc,
		fakeLimiter{},
		fakeValidator{tokens: map[string]int64{"t1": 1}},
		"session_token",
		false,
		"*",
		logger.NewDefault(false),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	c1 := dialWS(t, server.URL, "t1")
	defer c1.Close()
	consumeUntilType(t, c1, MessageTypeConnected)

	if err := c1.WriteMessage(websocket.TextMessage, []byte("not-json")); err != nil {
		t.Fatalf("write message: %v", err)
	}

	errMsg := consumeUntilType(t, c1, MessageTypeError)
	var errPayload ErrorPayload
	if err := json.Unmarshal(errMsg.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload.Code != "PARSE_ERROR" {
		t.Fatalf("expected PARSE_ERROR code, got %s", errPayload.Code)
	}
}

func dialWS(t *testing.T, baseURL, token string) *websocket.Conn {
	t.Helper()
	u := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws"
	header := http.Header{}
	header.Add("Cookie", "session_token="+token)
	conn, _, err := websocket.DefaultDialer.Dial(u, header)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	return conn
}

func consumeUntilType(t *testing.T, conn *websocket.Conn, want string) WSMessage {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read message: %v", err)
		}
		for _, raw := range splitMessages(data) {
			var msg WSMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				t.Fatalf("unmarshal ws: %v", err)
			}
			if msg.Type == want {
				return msg
			}
		}
	}
}

func consumeUntilTypes(t *testing.T, conn *websocket.Conn, wants ...string) map[string]WSMessage {
	t.Helper()
	pending := make(map[string]struct{}, len(wants))
	for _, w := range wants {
		pending[w] = struct{}{}
	}
	seen := make(map[string]WSMessage, len(wants))

	deadline := time.Now().Add(2 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	for len(pending) > 0 {
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read message: %v", err)
		}
		for _, raw := range splitMessages(data) {
			var msg WSMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				t.Fatalf("unmarshal ws: %v", err)
			}
			if _, ok := pending[msg.Type]; ok {
				seen[msg.Type] = msg
				delete(pending, msg.Type)
			}
		}
	}
	return seen
}

func splitMessages(data []byte) [][]byte {
	parts := strings.Split(string(data), "\n")
	out := make([][]byte, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, []byte(p))
	}
	return out
}

func mustWSMessage(t *testing.T, typ string, payload interface{}) []byte {
	t.Helper()
	msg, err := NewWSMessage(typ, payload)
	if err != nil {
		t.Fatalf("new ws message: %v", err)
	}
	return msg
}

func ptrInt64(v int64) *int64 { return &v }

func ptrString(v string) *string { return &v }

func sampleNotificationDTO() usecasenotification.NotificationDTO {
	return usecasenotification.NotificationDTO{
		ID:         1,
		UserID:     1,
		Type:       "follow_request",
		EntityType: "follow_request",
		EntityID:   2,
		IsRead:     false,
	}
}

func ensureNoType(t *testing.T, conn *websocket.Conn, typ string) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, data, err := conn.ReadMessage()
	if err != nil {
		return
	}
	for _, raw := range splitMessages(data) {
		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("unmarshal ws: %v", err)
		}
		if msg.Type == typ {
			t.Fatalf("unexpected message type %s", typ)
		}
	}
}
