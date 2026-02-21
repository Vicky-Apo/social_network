# Chat & WebSocket API (Frontend Guide)

This document explains how to connect to the real-time chat, send and receive messages, and handle unread message tracking.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

WebSocket endpoint:

- `ws://localhost:8080/ws`

## Cookie-based sessions

The WebSocket connection authenticates using the same session cookie as all other endpoints. The browser sends the cookie automatically on the upgrade request — no extra headers or options are needed. Make sure the user is logged in before opening the connection.

## REST Endpoints

These endpoints expose conversations and messages over HTTP.

### List conversations

`GET /conversations?limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "type": "direct",
      "other_user_id": 2,
      "last_message": {
        "id": 42,
        "conversation_id": 5,
        "sender_id": 2,
        "content": "Hey there!",
        "media_path": null,
        "created_at": "2025-01-24T12:34:56Z"
      },
      "unread_count": 3,
      "created_at": "2025-01-24T12:00:00Z"
    }
  ]
}
```

### Get conversation

`GET /conversations/{id}`

### List messages

`GET /conversations/{id}/messages?limit=20&offset=0`

### Mark conversation as read

`PATCH /conversations/{id}/read`

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "read"
  }
}
```

### Unread counts

`GET /conversations/unread-counts`

Response (200):

```json
{
  "success": true,
  "data": {
    "5": 3,
    "12": 1
  }
}
```

## Connecting

Open a WebSocket connection to `/ws`. The server validates your session cookie during the HTTP upgrade. If the cookie is missing or invalid, the upgrade is rejected with `401`.

On a successful connection the server sends two messages in order:

1. A `connected` confirmation.
2. An `unread_counts` message listing conversations that have unread messages (only sent if there are any).

## Message format

All messages use the same JSON envelope, both directions:

```json
{
  "type": "message_type",
  "payload": { ... }
}
```

## Server → Client messages

### connected

Sent immediately after a successful WebSocket upgrade.

```json
{
  "type": "connected",
  "payload": {
    "user_id": 1,
    "status": "connected"
  }
}
```

### unread_counts

Sent after connection if the user has unread messages in any conversation. Not sent if everything is read.
Additionally, the server pushes an `unread_counts` update to recipients whenever a new message is delivered so message badges can update in real time.

```json
{
  "type": "unread_counts",
  "payload": [
    { "conversation_id": 5, "unread_count": 3 },
    { "conversation_id": 12, "unread_count": 1 }
  ]
}
```

### chat_message

Sent when a new message arrives in a conversation the user belongs to. Also sent back to the sender as a delivery confirmation after they send a message.

```json
{
  "type": "chat_message",
  "payload": {
    "id": 42,
    "conversation_id": 5,
    "sender_id": 2,
    "content": "Hey there!",
    "media_path": null,
    "created_at": "2025-01-24T12:34:56Z"
  }
}
```

### typing

Sent when another member of a conversation starts or stops typing.

```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 5,
    "user_id": 2,
    "is_typing": true
  }
}
```
Notes:
- Typing is ephemeral and may be auto-cleared by the server after a short inactivity window (~5s).
- If a client disconnects while typing, the server will send a final `is_typing: false` for that user.

### user_online / user_offline

Sent when a user in your follow network comes online or goes offline. On connection, you also receive a `user_online` for each contact that is already online at that moment.

```json
{
  "type": "user_online",
  "payload": { "user_id": 3 }
}
```

```json
{
  "type": "user_offline",
  "payload": { "user_id": 3 }
}
```

### error

Sent when a request you made over the WebSocket fails.

```json
{
  "type": "error",
  "payload": {
    "message": "rate limit exceeded, please try again later",
    "code": "RATE_LIMIT"
  }
}
```

Error codes:

- `RATE_LIMIT` — too many messages sent in a short period.
- `PARSE_ERROR` — the message payload could not be parsed.
- `SEND_ERROR` — the message could not be delivered (e.g. no follow relationship, not a group member, empty content).
- `MARK_READ_ERROR` — the mark-as-read request failed (e.g. not a member of the conversation).
- `UNKNOWN_TYPE` — unrecognised message type.

## Client → Server messages

### chat_message

Send a new message. Use `recipient_id` for a direct message or `group_id` for a group message. Provide one or both of `content` and `media_path`.

Direct message:

```json
{
  "type": "chat_message",
  "payload": {
    "recipient_id": 2,
    "content": "Hey there!",
    "media_path": null
  }
}
```

Group message:

```json
{
  "type": "chat_message",
  "payload": {
    "group_id": 7,
    "content": "Hello group!"
  }
}
```

Notes:
- Direct messages require that at least one of the two users follows the other.
- Direct messages are also allowed if the recipient has a public profile.
- Group messages require that you are a member of the group.
- `content` or `media_path` must be provided — both cannot be empty/null.
- The server echoes the message back to you as a `chat_message` confirmation and forwards it to the other participants.

### typing

Notify other members of a conversation that you are typing (or have stopped).

```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 5,
    "is_typing": true
  }
}
```

Notes:
- Send `is_typing: false` when the user stops typing.
- You must be a member of the conversation.
- The server will auto-clear typing after a short inactivity window (~5s).
- If a client disconnects while typing, the server sends a final `is_typing: false` for that user.

### mark_read

Mark a conversation as fully read. This clears the unread count for that conversation. Send this when the user opens or views a conversation.

```json
{
  "type": "mark_read",
  "payload": {
    "conversation_id": 5
  }
}
```

Notes:
- You must be a member of the conversation.
- The read pointer advances to the latest message in the conversation at the time of the request.
- The server does not push an `unread_counts` update after `mark_read`. The client should update local unread state or refetch `GET /conversations/unread-counts`.

## Unread message tracking

Each conversation tracks a per-user read position using message IDs.

- When you send a message, your own read position advances automatically — your own messages never show as unread.
- When you receive a `chat_message` from someone else, it is unread until you send `mark_read` for that conversation.
- On WebSocket connect, the server pushes your current unread counts via `unread_counts`. Use this to populate notification badges immediately on app load.
- On each new incoming message, the server also pushes an `unread_counts` update for that conversation.
- If a recipient is offline when a message is sent, it is persisted to the database. The `unread_counts` message on their next connection lets them know new messages are waiting.

## Online Presence (Frontend-Only, No New Endpoints)

This project uses WebSocket presence events only. No REST endpoint is required.

### How it works

- The server sends `user_online` / `user_offline` events for **your follow network** (both directions).
- On connect, the server sends `user_online` events for all currently online contacts in your follow network.

### Frontend strategy (minimal approach)

1. Maintain a `Set` of online user IDs.
2. On `user_online`, add the ID. On `user_offline`, remove it.
3. For **followers/following lists**, online status is accurate.
4. For **group members**, online status is only accurate for members who are in your follow network.
   - If a member is not in your follow network, you cannot know their presence without backend changes.
   - Show “unknown/offline” for those users if needed.

This keeps backend unchanged and still provides live online/offline UX where it is supported by current events.

## React fetch example

```ts
const WS_BASE = import.meta.env.VITE_WS_BASE_URL; // e.g. ws://localhost:8080

let socket: WebSocket;

export function connectChat(onMessage: (msg: { type: string; payload: any }) => void) {
  socket = new WebSocket(`${WS_BASE}/ws`);

  socket.onopen = () => {
    console.log("WebSocket connected");
  };

  socket.onmessage = (event: MessageEvent) => {
    const msg = JSON.parse(event.data);
    onMessage(msg);
  };

  socket.onerror = (err) => {
    console.error("WebSocket error", err);
  };

  socket.onclose = () => {
    console.log("WebSocket closed");
  };
}

export function sendChatMessage(payload: {
  recipient_id?: number;
  group_id?: number;
  content?: string;
  media_path?: string;
}) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;
  socket.send(JSON.stringify({ type: "chat_message", payload }));
}

export function sendTyping(conversationId: number, isTyping: boolean) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;
  socket.send(JSON.stringify({
    type: "typing",
    payload: { conversation_id: conversationId, is_typing: isTyping },
  }));
}

export function markConversationRead(conversationId: number) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;
  socket.send(JSON.stringify({
    type: "mark_read",
    payload: { conversation_id: conversationId },
  }));
}

export function disconnectChat() {
  if (socket) socket.close();
}
```

## REST endpoints (conversation history)

These endpoints complement WebSockets by providing conversation lists and history.

### List conversations

`GET /conversations`

Notes:
- Returns direct and group conversations.
- Group conversations include `group_id` in the response.

### Get conversation

`GET /conversations/{id}`

### List conversation messages

`GET /conversations/{id}/messages?limit=20&offset=0`

### Mark conversation read

`PATCH /conversations/{id}/read`

### Unread counts (HTTP)

`GET /conversations/unread-counts`

## Message reactions (emoji)

### Toggle message reaction

`POST /messages/{id}/reactions`

Request body (JSON):

```json
{
  "emoji": "😀"
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "added"
  }
}
```

Notes:
- If the same emoji is sent again, the reaction is removed (`status: "removed"`).
- You must be a member of the conversation containing the message.
- Emoji length is limited to 8 characters.

### List message reactions

`GET /messages/{id}/reactions`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "message_id": 12,
      "user_id": 5,
      "emoji": "😀",
      "created_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

## Notes

- The WebSocket connection requires a valid session cookie — the user must be logged in before connecting.
- The server rate-limits WebSocket messages per user. If the limit is hit, an `error` message with code `RATE_LIMIT` is sent back.
- Messages are persisted to the database regardless of whether the recipient is online. Offline recipients will see the messages when they next open the conversation.
- Conversation history fetching (loading past messages) is available over HTTP via `/conversations/{id}/messages`.
- Typing indicators are ephemeral UI state. The frontend should still apply a local timeout (e.g., 4–6s) in case the client does not receive a stop-typing event.
