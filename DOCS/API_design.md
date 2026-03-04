# Social Network API Plan (REST + WebSockets + optional GraphQL)

This document is a realistic implementation plan for the project using:

- **REST** for core HTTP operations (commands + most reads)
- **WebSockets** for real-time events (chat, notifications, presence)
- **GraphQL (optional / later)** for complex aggregated reads (feed/profile) once REST becomes painful

---

## 0) Guiding Principles

1. **Keep writes boring**: use REST for create/update/delete.
2. **Push events, don’t do business logic over WS**: WebSockets deliver real-time updates; REST still validates and persists.
3. **Start simple, evolve safely**: GraphQL is an optimization layer for reads, not a requirement.
4. **Auth-first**: sessions + cookies for HTTP; WebSocket authenticates using the same session cookie (or a short-lived WS token).

---

## 1) Phase Plan (Realistic)

### Phase 1 — MVP (REST + WebSockets for chat only)

**Goal:** working social network with login, profiles, follows, posts, groups, events; plus real-time chat.

Deliverables:

- REST: auth, profiles, follows, posts/comments, groups/events, notifications CRUD (basic).
- WebSockets: private + group chat messages, notifications in real time in real-time.
- Minimal notification polling (REST) if needed.

### Phase 2 — Real-time polish (WebSockets for notifications + presence)

**Goal:** “feels live” with notifications + online (online users) indicators.

Deliverables:

- WebSockets: notification events, unread counters, presence (online/offline), typing indicators.
- REST: notification list, mark-read endpoints.

### Phase 3 — Performance + DX (Optional GraphQL Read Layer)

**Goal:** reduce frontend over-fetching and simplify complex screens (feed, profile page).

Deliverables:

- GraphQL read-only endpoint for feed/profile aggregation.
- Keep REST for all writes (commands).

---

## 2) REST API Design

### 2.1 Conventions

- Base URL: `/api/v1`
- JSON payloads
- Use **cookies + session** for auth (project requirement).
- Standard status codes:
  - `200 OK`, `201 Created`, `204 No Content`
  - `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`
- Pagination:
  - `?limit=20&cursor=<opaque>`
- Sorting/filtering:
  - `?sort=created_at_desc`
- Media upload:
  - `multipart/form-data` with server-stored files and `media_path` in DB.

### 2.2 Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET  /auth/me` (returns current user + session state)

**Notes**

- Session cookie: `HttpOnly`, `Secure` (when using HTTPS), `SameSite=Lax` (or `Strict` if feasible).

### 2.3 Profiles

- `GET   /users/:id` (public profile view; enforce private/public rules)
- `PATCH /users/me` (update profile fields)
- `POST  /users/me/avatar` (upload avatar)
- `PATCH /users/me/privacy` (toggle `is_public`)

### 2.4 Follow system

- `POST   /users/:id/follow-request` (creates request or auto-follow if target is public)
- `POST   /users/:id/unfollow`
- `GET    /users/:id/followers`
- `GET    /users/:id/following`

Requests management:

- `GET    /follow-requests` (incoming)
- `POST   /follow-requests/:requester_id/accept`
- `POST   /follow-requests/:requester_id/decline`

### 2.5 Posts & Feed

Create:

- `POST /posts` (user post; visibility in body)
- `POST /groups/:group_id/posts` (group post)

Read:

- `GET  /feed` (viewer feed; respects visibility rules)
- `GET  /users/:id/posts`
- `GET  /groups/:id/posts`
- `GET  /posts/:id`

Update/delete (optional):

- `PATCH  /posts/:id`
- `DELETE /posts/:id`

Visibility rules (server-side):

- `public` → everyone logged-in can view
- `followers` → only followers of author
- `private` → only `post_allowed_users` + author

### 2.6 Comments

- `POST /posts/:id/comments`
- `GET  /posts/:id/comments`
- `DELETE /comments/:id` (author/admin rules)

### 2.7 Groups

- `POST /groups`
- `GET  /groups` (browse)
- `GET  /groups/:id`
- `POST /groups/:id/invite` (invited_user_id)
- `POST /groups/:id/join-request`
- `POST /groups/:id/join-request/accept` (creator only)
- `POST /groups/:id/join-request/decline` (creator only)
- `POST /groups/:id/invitations/accept`
- `POST /groups/:id/invitations/decline`
- `POST /groups/:id/leave`

### 2.8 Events

- `POST /groups/:id/events`
- `GET  /groups/:id/events`
- `POST /events/:id/respond` (going / not_going)
- `GET  /events/:id/responses` (members only)

### 2.9 Notifications

- `GET   /notifications?unread_only=true`
- `POST  /notifications/:id/read`
- `POST  /notifications/read-all`

**Note:** even if you push notifications via WebSocket, you still need REST to fetch the list/history.

---

## 3) WebSockets Plan

### 3.1 Connection

- Endpoint: `GET /ws`
- Auth:
  - preferred: reuse session cookie during WS handshake
  - alternative: REST endpoint `POST /ws/token` returning short-lived token, then connect with `?token=...`

### 3.2 Message Envelope (recommended)

All WS frames use a common envelope:

```json
{
  "type": "message.send",
  "request_id": "uuid-optional",
  "payload": {}
}
```

Server responses:

```json
{
  "type": "message.sent",
  "request_id": "same-if-provided",
  "payload": {}
}
```

### 3.3 Events to support (Phase 1 → 2)

#### Chat (Phase 1)

Client → Server:

- `conversation.join` { "conversation_id": 123 }
- `message.send` { "conversation_id": 123, "content": "...", "emoji": null }

Server → Clients:

- `message.new` { "conversation_id": 123, "message": { ... } }

Rules:

- Sender must be a member of the conversation.
- For direct chats: enforce “can chat if following relationship exists OR recipient is public” (as per project spec).

#### Notifications (Phase 2)

Server → Client:

- `notification.new` { "notification": { ... } }
- `notification.count` { "unread": 5 }

Client → Server (optional):

- `notification.read` { "notification_id": 55 }

#### Presence & typing (Phase 2)

Server → Clients:

- `presence.update` { "user_id": 7, "status": "online" }
- `typing` { "conversation_id": 123, "user_id": 7, "is_typing": true }

### 3.4 Reliability

- Persist messages in DB first, then broadcast.
- If broadcast fails, DB is still source of truth.
- Client should be able to re-sync:
  - `GET /conversations/:id/messages?limit=50&before=<ts>` (REST)

---

## 4) GraphQL Plan (Optional / Later)

### 4.1 What GraphQL should be used for

**Read-heavy aggregated screens**, e.g.:

- home feed with nested author + comment counts
- profile page: user + follower counts + last posts + relationship state
- group page: group details + members + posts + events summary

### 4.2 What GraphQL should NOT be used for (in this project)

- authentication
- posting messages
- follow/unfollow
- creating posts/comments
  Those remain REST to keep security and logic straightforward.

### 4.3 Endpoint

- `POST /graphql`

### 4.4 Example Queries (read-only)

Feed query (example shape):

```graphql
query Feed($cursor: String, $limit: Int!) {
  feed(cursor: $cursor, limit: $limit) {
    items {
      id
      createdAt
      visibility
      content
      mediaPath
      author {
        id
        nickname
        avatarPath
      }
      commentCount
    }
    nextCursor
  }
}
```

Profile page:

```graphql
query Profile($userId: ID!) {
  user(id: $userId) {
    id
    firstName
    lastName
    nickname
    about
    isPublic
    followersCount
    followingCount
    recentPosts(limit: 10) {
      id
      content
      createdAt
    }
  }
}
```

### 4.5 Security model

- Use the same session cookie as REST.
- Resolvers must apply the same visibility rules as REST handlers.
- Avoid N+1 queries by using batching (DataLoader pattern) or SQL joins.

---

## 5) Suggested Deliverable Checklist

### MVP (Phase 1) — must-haves

- [ ] Sessions + cookies: register/login/logout
- [ ] Profile public/private enforcement
- [ ] Follow requests + follows
- [ ] Posts with visibility + comments
- [ ] Groups + invitations + join requests
- [ ] Events + responses
- [ ] WebSocket chat (direct + group + private multi-chat)

### Phase 2 — quality

- [ ] WebSocket notifications + unread count
- [ ] Presence + typing
- [ ] Better pagination & indexes

### Phase 3 — optional

- [ ] GraphQL feed/profile aggregation
- [ ] Strict performance tuning and caching

---

## 6) Practical “Don’t Shoot Yourself” Notes

- Don’t implement GraphQL in the MVP unless REST becomes clearly painful.
- Keep WebSocket payloads small and versioned (type names are your versioning tool).
- Always keep REST “source of truth” endpoints for re-sync and debugging.
- Treat authorization rules as a first-class module (middleware/service), not scattered `if`s.

---
