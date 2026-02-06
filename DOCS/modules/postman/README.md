# Postman Collection - Social Network API

This directory contains Postman collection and environment files for testing the Social Network API.

## Files

- `social-network.postman_collection.json` - Main API collection with all endpoints
- `social-network.postman_environment.json` - Environment variables for local development

## Import to Postman

1. Open Postman
2. Click **Import** button
3. Select both JSON files
4. Select the **Social Network Local** environment from the dropdown

## WebSocket Testing - Quick Start

**IMPORTANT:** WebSocket requests cannot be imported. Follow these steps to test WebSocket:

### 1. Login First
Run **Auth → Login User A** to get a session cookie.

### 2. Create WebSocket Request
1. In Postman, click **New** (or press Ctrl/Cmd + N)
2. Select **WebSocket Request**
3. Enter URL: `ws://localhost:8080/ws`
4. Click **Connect** button
5. You should see connection established and receive:
```json
{
  "type": "connected",
  "payload": {
    "user_id": 1,
    "status": "connected"
  }
}
```

### 3. Send Your First Message
In the message compose area at the bottom, type:
```json
{
  "type": "chat_message",
  "payload": {
    "recipient_id": 2,
    "content": "Hello from WebSocket!"
  }
}
```
Click **Send** and you should receive a confirmation response.

**That's it!** See the detailed guide below for more message types and advanced testing.

## WebSocket Implementation Analysis

### Architecture

The WebSocket implementation consists of:

- **Handler** (`handler.go`): Manages WebSocket upgrade, authentication, and initial connection
- **Client** (`client.go`): Represents individual WebSocket connections with read/write pumps
- **Hub** (`hub.go`): Central message broker managing all active connections and routing messages
- **Message** (`message.go`): Defines message types and payload structures

### Key Features

1. **Authentication**: Session cookie-based (same as HTTP endpoints)
2. **Multi-device support**: Users can have multiple WebSocket connections simultaneously
3. **Direct messages**: 1-on-1 private messaging
4. **Group messages**: Broadcast to all group members
5. **Typing indicators**: Real-time typing status (currently logged, can be extended)
6. **Connection management**: Automatic ping/pong for keep-alive
7. **Error handling**: Structured error messages with error codes

### Connection Flow

```
Client                          Server
  |                               |
  |-------- HTTP Upgrade -------->|
  |     (with session cookie)     |
  |                               |
  |<------ 101 Switching ---------|
  |         Protocols             |
  |                               |
  |<----- Connected Message ------|
  | {"type":"connected",...}      |
  |                               |
  |<======= WebSocket Open =======|
```

### Message Format

All WebSocket messages follow this structure:

```json
{
  "type": "message_type",
  "payload": { ... }
}
```

### Message Types

#### Client → Server

**1. chat_message** - Send a message
```json
{
  "type": "chat_message",
  "payload": {
    "recipient_id": 2,           // For direct messages (optional)
    "group_id": 1,               // For group messages (optional)
    "content": "Hello!",         // Message content (required)
    "media_path": "/path.jpg"    // Media attachment (optional)
  }
}
```
**Note**: Specify either `recipient_id` OR `group_id`, not both.

**2. typing** - Send typing indicator
```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 1,
    "is_typing": true
  }
}
```

#### Server → Client

**1. connected** - Connection confirmation
```json
{
  "type": "connected",
  "payload": {
    "user_id": 1,
    "status": "connected"
  }
}
```

**2. chat_message** - Message delivery
```json
{
  "type": "chat_message",
  "payload": {
    "id": 123,
    "conversation_id": 1,
    "sender_id": 2,
    "content": "Hello!",
    "media_path": null,
    "created_at": "2026-02-03T10:00:00Z"
  }
}
```

**3. typing** - Typing indicator
```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 1,
    "is_typing": true
  }
}
```

**4. error** - Error message
```json
{
  "type": "error",
  "payload": {
    "message": "invalid message format",
    "code": "PARSE_ERROR"
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `PARSE_ERROR` | Invalid JSON format or payload structure |
| `UNKNOWN_TYPE` | Unknown message type |
| `SEND_ERROR` | Failed to send message (check message for details) |

### Hub Architecture

The Hub maintains connections in a map structure:
```
Hub
├── clients: map[userID]map[*Client]bool
│   ├── user1
│   │   ├── client_connection_1
│   │   └── client_connection_2  (multi-device)
│   └── user2
│       └── client_connection_1
├── register: chan *Client
└── unregister: chan *Client
```

This allows:
- Multiple devices per user
- Efficient message routing to all user's devices
- Online/offline status tracking

## Testing WebSocket with Postman

### Important Note

WebSocket requests **cannot be exported/imported** in Postman collection files. You must create them manually in Postman.

### Prerequisites

1. Postman version 10.0 or higher (WebSocket support)
2. Backend server running on `http://localhost:8080`
3. Active user session (login first)

### Step-by-Step Testing Guide

#### 1. Setup - Create Two Users

Run these requests in order:

1. **Auth → Register User A**
   - Creates user: `usera@example.com`

2. **Auth → Register User B**
   - Creates user: `userb@example.com`

3. **Auth → Login User A**
   - Logs in User A
   - Session cookie is automatically saved

4. **Auth → Login User B**
   - Open request in new tab/window
   - Use a different Postman window or browser to maintain separate sessions

#### 2. Connect to WebSocket

1. Open **WebSocket → Connect to WebSocket**
2. Click **Connect**
3. You should receive a connection confirmation:
   ```json
   {
     "type": "connected",
     "payload": {
       "user_id": 1,
       "status": "connected"
     }
   }
   ```

#### 3. Send Direct Message

1. Make sure you're connected as User A
2. Open **WebSocket → Send Direct Message**
3. Update `recipientId` in environment to User B's ID (usually `2`)
4. Click **Send**
5. You should receive confirmation with the message details

**Expected response:**
```json
{
  "type": "chat_message",
  "payload": {
    "id": 1,
    "conversation_id": 1,
    "sender_id": 1,
    "content": "Hello! This is a direct message from Postman",
    "created_at": "2026-02-03T..."
  }
}
```

#### 4. Test Multi-User Chat

To test real-time messaging between users:

1. **User A**: Connect to WebSocket (connection 1)
2. **User B**: Open new Postman window, login as User B, connect to WebSocket (connection 2)
3. **User A**: Send message to User B
4. **User B**: Should receive the message in real-time in their WebSocket connection

#### 5. Send Group Message

1. Update `groupId` in environment variables
2. Open **WebSocket → Send Group Message**
3. Click **Send**
4. All group members will receive the message

#### 6. Test Typing Indicators

1. Open **WebSocket → Send Typing Indicator (Start)**
2. Update `conversationId` to your conversation ID
3. Click **Send**
4. Send **Send Typing Indicator (Stop)** when done typing

**Note**: Typing indicators are currently logged but not broadcast. The backend can be extended to broadcast to conversation members.

#### 7. Test Error Handling

Try these requests to see error responses:

1. **WebSocket → Invalid Message (Error Test)**
   - Sends unknown message type
   - Expect: `UNKNOWN_TYPE` error

2. **WebSocket → Malformed JSON (Error Test)**
   - Sends invalid JSON
   - Expect: `PARSE_ERROR` error

### Environment Variables

Key WebSocket variables in the environment:

| Variable | Default | Description |
|----------|---------|-------------|
| `wsUrl` | `ws://localhost:8080` | WebSocket base URL |
| `recipientId` | `2` | Target user for direct messages |
| `groupId` | `1` | Target group for group messages |
| `conversationId` | `1` | Conversation ID for typing indicators |

### Tips

1. **Cookie Management**: Postman automatically handles session cookies. Make sure you're logged in before connecting.

2. **Multiple Connections**: To test multi-device scenarios, open multiple Postman windows with different users.

3. **Message Persistence**: Messages are saved to the database and can be retrieved via HTTP endpoints.

4. **Connection Issues**: If connection fails:
   - Check if you're logged in (session cookie exists)
   - Verify backend is running on port 8080
   - Check browser console for errors

5. **Real-time Testing**: Keep the WebSocket connection open to receive messages from other users in real-time.

## HTTP Endpoints Related to Chat

While WebSocket is for real-time messaging, you may need these HTTP endpoints:

- **GET** `/conversations` - List user's conversations
- **GET** `/conversations/{id}` - Get conversation details
- **GET** `/conversations/{id}/messages` - Get message history
- **POST** `/groups` - Create a group
- **GET** `/groups/{id}` - Get group details

(Note: Add these endpoints to the collection as needed)

## Technical Implementation Details

### Backend Components

1. **Server Setup** (`server.go:142-146`)
   - Creates Hub instance
   - Starts Hub in goroutine
   - Initializes WebSocket handler with dependencies

2. **Router Configuration** (`router.go:86`)
   - WebSocket endpoint at `/ws`
   - No auth middleware (authentication handled internally)
   - Subject to CORS and rate limiting

3. **Connection Lifecycle**
   - Client connects → Handler validates session
   - Handler upgrades connection → Creates Client instance
   - Client registers with Hub → Starts read/write pumps
   - Client disconnects → Unregisters from Hub → Cleans up resources

4. **Message Flow**
   - Client sends message → readPump receives → handleMessage processes
   - Service layer processes (saves to DB, determines recipients)
   - Response sent to sender (confirmation)
   - Hub broadcasts to recipients → writePump sends to each connection

5. **Concurrency Safety**
   - Hub uses mutex for thread-safe client map access
   - Separate goroutines for reading and writing
   - Buffered channels prevent blocking
   - Graceful handling of closed connections

## Future Enhancements

Potential improvements to the WebSocket implementation:

1. **Typing Indicators Broadcasting**: Extend typing indicator handler to broadcast to conversation members
2. **Read Receipts**: Add message read status tracking
3. **Online Status**: Expose user online/offline status via HTTP endpoint
4. **Message Editing**: Support for editing sent messages
5. **Message Reactions**: Real-time emoji reactions
6. **File Uploads**: Direct file upload via WebSocket (with chunking)
7. **Voice Messages**: Support for audio message attachments
8. **Push Notifications**: Integration for offline users
9. **Connection Limits**: Rate limiting for message sending
10. **Admin Features**: Broadcast messages, user kick/ban

## Support

For issues or questions:
- Check the backend logs for detailed error messages
- Ensure all migrations are applied
- Verify database connection
- Check CORS configuration if connecting from browser
