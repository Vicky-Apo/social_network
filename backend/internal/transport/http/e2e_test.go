//go:build integration

package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"

	"social-network/backend/internal/config"
	transporthandler "social-network/backend/internal/transport/http/handler"
	"social-network/backend/internal/transport/http/middleware"
	transportws "social-network/backend/internal/transport/websocket"
	accessusecase "social-network/backend/internal/usecase/access"
	authusecase "social-network/backend/internal/usecase/auth"
	chatusecase "social-network/backend/internal/usecase/chat"
	commentusecase "social-network/backend/internal/usecase/comment"
	eventusecase "social-network/backend/internal/usecase/event"
	followusecase "social-network/backend/internal/usecase/follow"
	groupusecase "social-network/backend/internal/usecase/group"
	messagereactionusecase "social-network/backend/internal/usecase/message_reaction"
	notificationusecase "social-network/backend/internal/usecase/notification"
	postusecase "social-network/backend/internal/usecase/post"
	profileusecase "social-network/backend/internal/usecase/profile"
	reactionusecase "social-network/backend/internal/usecase/reaction"
	userusecase "social-network/backend/internal/usecase/user"
	"social-network/backend/pkg/db/postgres"
	authrepo "social-network/backend/pkg/db/postgres/repositories/auth"
	chatrepo "social-network/backend/pkg/db/postgres/repositories/chat"
	commentrepo "social-network/backend/pkg/db/postgres/repositories/comment"
	eventrepo "social-network/backend/pkg/db/postgres/repositories/event"
	followrepo "social-network/backend/pkg/db/postgres/repositories/follow"
	grouprepo "social-network/backend/pkg/db/postgres/repositories/group"
	notificationrepo "social-network/backend/pkg/db/postgres/repositories/notification"
	postrepo "social-network/backend/pkg/db/postgres/repositories/post"
	reactionrepo "social-network/backend/pkg/db/postgres/repositories/reaction"
	userrepo "social-network/backend/pkg/db/postgres/repositories/user"
	"social-network/backend/pkg/logger"
)

const sessionCookieName = "session_token"

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type httpResp struct {
	StatusCode int
	Body       []byte
	Cookies    []*http.Cookie
}

func TestE2EFlow_TwoUsers(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	client := server.Client()

	// register + login user1
	user1ID := registerUser(t, client, server.URL, "u1@example.com")
	user2ID := registerUser(t, client, server.URL, "u2@example.com")
	if user1ID == 0 || user2ID == 0 {
		t.Fatalf("expected user ids")
	}

	u1Cookie := loginUser(t, client, server.URL, "u1@example.com", "password123")
	u2Cookie := loginUser(t, client, server.URL, "u2@example.com", "password123")

	setProfileVisibility(t, client, server.URL, u1Cookie, user1ID, true)

	// user1 creates post
	uploadPath := uploadFile(t, client, server.URL, u1Cookie, "post")
	postID := createPostWithMedia(t, client, server.URL, u1Cookie, "hello world", uploadPath)

	// user2 reacts + comments
	addReaction(t, client, server.URL, u2Cookie, postID, "like")
	addComment(t, client, server.URL, u2Cookie, postID, "nice post")

	// switch profile to private to trigger follow request flow
	setProfileVisibility(t, client, server.URL, u1Cookie, user1ID, false)

	// follow request (user1 private by default)
	reqID := followRequest(t, client, server.URL, u2Cookie, user1ID)
	acceptFollowRequest(t, client, server.URL, u1Cookie, reqID)

	// profile view
	getProfile(t, client, server.URL, u2Cookie, user1ID)

	// group flow
	groupID := createGroup(t, client, server.URL, u1Cookie, "group one")
	invID := inviteToGroup(t, client, server.URL, u1Cookie, groupID, user2ID)
	acceptGroupInvitation(t, client, server.URL, u2Cookie, invID)
	createGroupPost(t, client, server.URL, u1Cookie, groupID, "group post")
	listGroupPosts(t, client, server.URL, u2Cookie, groupID)

	// event flow
	eventID := createGroupEvent(t, client, server.URL, u1Cookie, groupID)
	respondEvent(t, client, server.URL, u2Cookie, eventID, "going")

	// notifications list
	listNotifications(t, client, server.URL, u1Cookie)
	listNotifications(t, client, server.URL, u2Cookie)

	// chat via websocket
	msgID := wsDirectMessage(t, server.URL, u1Cookie, u2Cookie, user1ID)
	addMessageReaction(t, client, server.URL, u1Cookie, msgID, "😀")
	listMessageReactions(t, client, server.URL, u1Cookie, msgID)
}

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	db := openTestDB(t)
	cleanupTables(t, db)

	log := logger.NewDefault(false)

	authCfg := config.AuthConfig{
		BcryptCost:          bcrypt.MinCost,
		SessionTokenBytes:   16,
		SessionDurationDays: 1,
		SessionCookieName:   sessionCookieName,
		SessionMaxAge:       86400,
		MinPasswordLength:   6,
		MinUserAge:          13,
	}

	authRepository := authrepo.NewRepository(db)
	postRepository := postrepo.NewRepository(db)
	commentRepository := commentrepo.NewRepository(db)
	eventRepository := eventrepo.NewRepository(db)
	reactionRepository := reactionrepo.NewRepository(db)
	userRepository := userrepo.NewRepository(db)
	followRepository := followrepo.NewRepository(db)
	chatRepository := chatrepo.NewRepository(db)
	groupRepository := grouprepo.NewRepository(db)
	notificationRepository := notificationrepo.NewRepository(db)

	wsHub := transportws.NewHub(followRepository, log)
	go wsHub.Run()

	authService := authusecase.NewService(authRepository, authCfg, log)
	notificationPublisher := transportws.NewNotificationPublisher(wsHub)
	notificationService := notificationusecase.NewService(notificationRepository, notificationPublisher, log)
	accessService := accessusecase.NewService(userRepository, followRepository, postRepository, groupRepository, log)
	postService := postusecase.NewService(postRepository, userRepository, accessService, log)
	commentService := commentusecase.NewService(commentRepository, postRepository, accessService, notificationService)
	reactionService := reactionusecase.NewService(reactionRepository, postRepository, commentRepository, notificationService)
	profileService := profileusecase.NewService(userRepository, accessService)
	followService := followusecase.NewService(userRepository, followRepository, notificationService)
	userService := userusecase.NewService(userRepository)
	chatService := chatusecase.NewService(chatRepository, groupRepository, accessService, log)
	groupService := groupusecase.NewService(groupRepository, accessService, notificationService)
	eventService := eventusecase.NewService(eventRepository, groupRepository, accessService, notificationService)
	messageReactionService := messagereactionusecase.NewService(chatRepository)

	authHandlerCfg := transporthandler.AuthHandlerConfig{
		CookieName: authCfg.SessionCookieName,
		MaxAge:     authCfg.SessionMaxAge,
	}
	authHandler := transporthandler.NewAuthHandler(authService, log, authHandlerCfg)
	postHandler := transporthandler.NewPostHandler(postService, log)
	commentHandler := transporthandler.NewCommentHandler(commentService, log)
	reactionHandler := transporthandler.NewReactionHandler(reactionService, log)
	profileHandler := transporthandler.NewProfileHandler(profileService, postService, log)
	followHandler := transporthandler.NewFollowHandler(followService, log)
	userHandler := transporthandler.NewUserHandler(userService, log)
	notificationHandler := transporthandler.NewNotificationHandler(notificationService, log)
	groupHandler := transporthandler.NewGroupHandler(groupService, log)
	eventHandler := transporthandler.NewEventHandler(eventService, log)
	chatHandler := transporthandler.NewChatHandler(chatService, log)
	messageReactionHandler := transporthandler.NewMessageReactionHandler(messageReactionService, log)
	uploadHandler := transporthandler.NewUploadHandler(t.TempDir(), 5*1024*1024, log)

	authMiddleware := middleware.Auth(authService, authCfg.SessionCookieName, log)
	rl := middleware.NewRateLimiter(1000, false, log)
	cors := middleware.CORS(config.CORSConfig{Enabled: false})
	sec := middleware.SecurityHeaders(config.SecurityHeadersConfig{Enabled: false})

	mw := Middlewares{
		Auth:            authMiddleware,
		RateLimit:       middleware.RateLimit(rl),
		CORS:            cors,
		SecurityHeaders: sec,
	}

	wsHandler := transportws.NewHandler(
		wsHub,
		chatService,
		rl,
		authService,
		authCfg.SessionCookieName,
		false,
		"*",
		log,
	)

	router := NewRouter(
		postHandler,
		authHandler,
		commentHandler,
		reactionHandler,
		profileHandler,
		followHandler,
		userHandler,
		notificationHandler,
		groupHandler,
		eventHandler,
		chatHandler,
		messageReactionHandler,
		uploadHandler,
		nil,
		"",
		wsHandler,
		mw,
	)

	server := httptest.NewServer(router)
	cleanup := func() {
		server.Close()
		wsHub.Stop()
		rl.Stop()
		db.Close()
	}
	return server, cleanup
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	migrations := os.Getenv("MIGRATIONS_PATH")
	if migrations == "" {
		t.Skip("MIGRATIONS_PATH not set")
	}
	abs, err := filepath.Abs(migrations)
	if err != nil {
		t.Fatalf("migrations path: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		abs = filepath.Join(findModuleRoot(t), migrations)
	}

	db, err := postgres.Open(postgres.WithDefaults(url))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := postgres.ApplyMigrations(db, "file://"+abs); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

func cleanupTables(t *testing.T, db *sql.DB) {
	t.Helper()
	tables := []string{
		"message_reactions",
		"messages",
		"conversation_members",
		"group_conversations",
		"conversations",
		"event_responses",
		"events",
		"group_join_requests",
		"group_invitations",
		"group_members",
		"groups",
		"post_categories",
		"post_allowed_users",
		"comment_reactions",
		"comments",
		"post_reactions",
		"posts",
		"follow_requests",
		"follows",
		"notifications",
		"sessions",
		"users",
	}
	for _, table := range tables {
		if _, err := db.ExecContext(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}

func registerUser(t *testing.T, client *http.Client, baseURL, email string) int64 {
	t.Helper()
	payload := map[string]string{
		"email":         email,
		"password":      "password123",
		"first_name":    "Test",
		"last_name":     "User",
		"date_of_birth": "01/01/2000",
	}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/auth/register", payload, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status: %d", resp.StatusCode)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	decodeData(t, resp.Body, &data)
	return data.ID
}

func loginUser(t *testing.T, client *http.Client, baseURL, email, password string) *http.Cookie {
	t.Helper()
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/auth/login", payload, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status: %d", resp.StatusCode)
	}
	for _, c := range resp.Cookies {
		if c.Name == sessionCookieName {
			return c
		}
	}
	t.Fatalf("session cookie not found")
	return nil
}

func createPost(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, content string) int64 {
	t.Helper()
	payload := map[string]any{
		"content": content,
		"privacy": "public",
	}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/posts", payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create post status: %d", resp.StatusCode)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	decodeData(t, resp.Body, &data)
	return data.ID
}

func createPostWithMedia(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, content, mediaPath string) int64 {
	t.Helper()
	payload := map[string]any{
		"content":    content,
		"media_path": mediaPath,
		"privacy":    "public",
	}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/posts", payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create post status: %d", resp.StatusCode)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	decodeData(t, resp.Body, &data)
	return data.ID
}

func addReaction(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, postID int64, reaction string) {
	t.Helper()
	payload := map[string]string{"reaction": reaction}
	url := baseURL + "/posts/" + itoa(postID) + "/reactions"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add reaction status: %d", resp.StatusCode)
	}
}

func addComment(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, postID int64, content string) {
	t.Helper()
	payload := map[string]string{"content": content}
	url := baseURL + "/posts/" + itoa(postID) + "/comments"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add comment status: %d", resp.StatusCode)
	}
}

func followRequest(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, targetID int64) int64 {
	t.Helper()
	payload := map[string]int64{"target_id": targetID}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/follow-requests", payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("follow request status: %d", resp.StatusCode)
	}
	var data struct {
		Request *struct {
			ID int64 `json:"id"`
		} `json:"request"`
	}
	decodeData(t, resp.Body, &data)
	if data.Request == nil {
		t.Fatalf("expected request")
	}
	return data.Request.ID
}

func acceptFollowRequest(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, reqID int64) {
	t.Helper()
	payload := map[string]string{"status": "accepted"}
	url := baseURL + "/follow-requests/" + itoa(reqID)
	resp := doJSON(t, client, http.MethodPatch, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("accept follow status: %d", resp.StatusCode)
	}
}

func getProfile(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, userID int64) {
	t.Helper()
	url := baseURL + "/profiles/" + itoa(userID)
	resp := doJSON(t, client, http.MethodGet, url, nil, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get profile status: %d", resp.StatusCode)
	}
}

func setProfileVisibility(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, userID int64, isPublic bool) {
	t.Helper()
	payload := map[string]bool{"is_public": isPublic}
	url := baseURL + "/profiles/" + itoa(userID) + "/visibility"
	resp := doJSON(t, client, http.MethodPatch, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("set visibility status: %d", resp.StatusCode)
	}
}

func createGroup(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, title string) int64 {
	t.Helper()
	payload := map[string]string{"title": title}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/groups", payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create group status: %d", resp.StatusCode)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	decodeData(t, resp.Body, &data)
	return data.ID
}

func inviteToGroup(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, groupID, inviteeID int64) int64 {
	t.Helper()
	payload := map[string]int64{"invitee_id": inviteeID}
	url := baseURL + "/groups/" + itoa(groupID) + "/invitations"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("invite status: %d", resp.StatusCode)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	decodeData(t, resp.Body, &data)
	return data.ID
}

func acceptGroupInvitation(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, invID int64) {
	t.Helper()
	payload := map[string]string{"status": "accepted"}
	url := baseURL + "/group-invitations/" + itoa(invID)
	resp := doJSON(t, client, http.MethodPatch, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("accept invitation status: %d", resp.StatusCode)
	}
}

func createGroupPost(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, groupID int64, content string) {
	t.Helper()
	payload := map[string]any{"content": content, "privacy": "public"}
	url := baseURL + "/groups/" + itoa(groupID) + "/posts"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create group post status: %d", resp.StatusCode)
	}
}

func listGroupPosts(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, groupID int64) {
	t.Helper()
	url := baseURL + "/groups/" + itoa(groupID) + "/posts"
	resp := doJSON(t, client, http.MethodGet, url, nil, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list group posts status: %d", resp.StatusCode)
	}
}

func createGroupEvent(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, groupID int64) int64 {
	t.Helper()
	payload := map[string]any{
		"title":       "Meetup",
		"event_time":  time.Now().Add(24 * time.Hour),
		"description": "desc",
	}
	url := baseURL + "/groups/" + itoa(groupID) + "/events"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create event status: %d", resp.StatusCode)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	decodeData(t, resp.Body, &data)
	return data.ID
}

func respondEvent(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, eventID int64, response string) {
	t.Helper()
	payload := map[string]string{"response": response}
	url := baseURL + "/events/" + itoa(eventID) + "/responses"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("respond event status: %d", resp.StatusCode)
	}
}

func listNotifications(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie) {
	t.Helper()
	resp := doJSON(t, client, http.MethodGet, baseURL+"/notifications", nil, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list notifications status: %d", resp.StatusCode)
	}
}

func wsDirectMessage(t *testing.T, baseURL string, u1Cookie, u2Cookie *http.Cookie, user1ID int64) int64 {
	t.Helper()
	c1 := dialWS(t, baseURL, u1Cookie)
	defer c1.Close()
	c2 := dialWS(t, baseURL, u2Cookie)
	defer c2.Close()

	consumeUntilType(t, c1, transportws.MessageTypeConnected)
	consumeUntilType(t, c2, transportws.MessageTypeConnected)

	payload := chatusecase.SendMessageRequest{
		RecipientID: &user1ID,
		Content:     ptrString("hello"),
	}
	msg := mustWSMessage(t, transportws.MessageTypeChat, payload)
	if err := c2.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("ws write: %v", err)
	}
	chatMsg := consumeUntilType(t, c1, transportws.MessageTypeChat)
	var dto chatusecase.MessageDTO
	if err := json.Unmarshal(chatMsg.Payload, &dto); err != nil {
		t.Fatalf("unmarshal chat payload: %v", err)
	}
	if dto.ID == 0 {
		t.Fatalf("expected message id")
	}
	return dto.ID
}

func uploadFile(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, kind string) string {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if kind != "" {
		if err := writer.WriteField("kind", kind); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = part.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	_, _ = part.Write(bytes.Repeat([]byte{0x00}, 10))
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/uploads", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("upload request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload status: %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read upload response: %v", err)
	}
	var api apiResponse
	if err := json.Unmarshal(raw, &api); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if !api.Success {
		t.Fatalf("upload api error: %s", api.Error)
	}
	var data struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(api.Data, &data); err != nil {
		t.Fatalf("decode upload data: %v", err)
	}
	if data.Path == "" {
		t.Fatalf("expected upload path")
	}
	return data.Path
}

func addMessageReaction(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, messageID int64, emoji string) {
	t.Helper()
	payload := map[string]string{"emoji": emoji}
	url := baseURL + "/messages/" + itoa(messageID) + "/reactions"
	resp := doJSON(t, client, http.MethodPost, url, payload, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add message reaction status: %d", resp.StatusCode)
	}
}

func listMessageReactions(t *testing.T, client *http.Client, baseURL string, cookie *http.Cookie, messageID int64) {
	t.Helper()
	url := baseURL + "/messages/" + itoa(messageID) + "/reactions"
	resp := doJSON(t, client, http.MethodGet, url, nil, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list message reactions status: %d", resp.StatusCode)
	}
}

func doJSON(t *testing.T, client *http.Client, method, url string, payload any, cookies []*http.Cookie) *httpResp {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode json: %v", err)
		}
	}
	req, err := http.NewRequest(method, url, &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return &httpResp{
		StatusCode: resp.StatusCode,
		Body:       data,
		Cookies:    resp.Cookies(),
	}
}

func decodeData(t *testing.T, data []byte, out any) {
	t.Helper()
	var api apiResponse
	if err := json.Unmarshal(data, &api); err != nil {
		t.Fatalf("decode api response: %v", err)
	}
	if !api.Success {
		t.Fatalf("api error: %s", api.Error)
	}
	if err := json.Unmarshal(api.Data, out); err != nil {
		t.Fatalf("decode data: %v", err)
	}
}

func dialWS(t *testing.T, baseURL string, cookie *http.Cookie) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws"
	header := http.Header{}
	header.Add("Cookie", cookie.Name+"="+cookie.Value)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	return conn
}

func consumeUntilType(t *testing.T, conn *websocket.Conn, want string) transportws.WSMessage {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read message: %v", err)
		}
		for _, raw := range splitMessages(data) {
			var msg transportws.WSMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				t.Fatalf("unmarshal ws: %v", err)
			}
			if msg.Type == want {
				return msg
			}
		}
	}
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
	msg, err := transportws.NewWSMessage(typ, payload)
	if err != nil {
		t.Fatalf("new ws message: %v", err)
	}
	return msg
}

func itoa(id int64) string { return strconv.FormatInt(id, 10) }

func ptrString(v string) *string { return &v }
