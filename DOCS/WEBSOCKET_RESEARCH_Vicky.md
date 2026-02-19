# WebSocket Research: Messages, Group Chat & Notifications

## TL;DR (Quick Summary)

**What we need:** Real-time chat (1-on-1 and groups) + live notifications
**Best approach:** Native Go WebSockets with Gorilla WebSocket library
**Why:** Simple, efficient, works with our existing auth, no extra infrastructure needed
**Complexity:** Medium (but totally doable!)

---

## Table of Contents

1. [What is WebSocket? (Simple Explanation)](#1-what-is-websocket-simple-explanation)
2. [Why WebSocket for Our Project?](#2-why-websocket-for-our-project)
3. [WebSocket Libraries: Comparing Options](#3-websocket-libraries-comparing-options)
4. [Recommended Approach](#4-recommended-approach)
5. [How It Will Work (Architecture)](#5-how-it-will-work-architecture)
6. [Implementation Plan (Step by Step)](#6-implementation-plan-step-by-step)
7. [Alternative Approaches (And Why We Skip Them)](#7-alternative-approaches-and-why-we-skip-them)
8. [Security Considerations](#8-security-considerations)
9. [Resources & Learning Materials](#9-resources--learning-materials)

---

## 1. What is WebSocket? (Simple Explanation)

### The Problem with Regular HTTP

Imagine you're texting a friend:

- **HTTP (regular web):** You send a message, wait for reply, send another, wait again...
  It's like sending letters back and forth. Slow and annoying!

- **WebSocket:** Open a phone call where both can talk anytime!
  Connection stays open, messages flow instantly both ways.

### Key Differences

| Feature | HTTP | WebSocket |
|---------|------|-----------|
| Connection | Open → Request → Close | Open once, stays open |
| Direction | Client asks, server responds | Both can send anytime |
| Overhead | Headers every time (heavy) | Small frames (light) |
| Real-time | No (polling needed) | Yes (instant updates) |

### When to Use WebSocket?

✅ **Perfect for:**
- Chat messages (instant delivery)
- Group conversations (multiple users)
- Notifications (live updates)
- Typing indicators
- Online/offline status
- Read receipts

❌ **Not needed for:**
- Loading user profiles (HTTP is fine)
- Creating posts (HTTP works)
- File uploads (HTTP better)

---

## 2. Why WebSocket for Our Project?

### Our Current Setup (From Codebase Review)

✅ **We already have:**
- Go backend with custom HTTP server
- Session-based authentication (cookies)
- PostgreSQL with conversation/message tables ready
- User context extraction middleware
- Next.js frontend with React

✅ **What's missing:**
- WebSocket connection handler
- Real-time message broadcasting
- Live notification delivery

### What We Need to Build

**Three main features:**

1. **Direct Messages (1-on-1 chat)**
   - User A sends message to User B
   - User B receives instantly (if online)
   - If offline, stored in DB, delivered on login

2. **Group Chat**
   - Multiple users in one conversation
   - Everyone receives messages instantly
   - Works with our existing groups table

3. **Live Notifications**
   - Follow requests, reactions, comments
   - Pop up instantly without page refresh
   - Uses existing notifications table

---

## 3. WebSocket Libraries: Comparing Options

### Option 1: Gorilla WebSocket (RECOMMENDED)

**What is it?**
The most popular Go WebSocket library. Battle-tested, simple, reliable.

**Pros:**
- ✅ Easy to learn (great docs)
- ✅ Works with standard Go `net/http`
- ✅ Fits our existing architecture
- ✅ Active maintenance (still updated)
- ✅ Good examples available
- ✅ No breaking changes from our setup

**Cons:**
- ⚠️ Need to build connection management ourselves
- ⚠️ Need to handle message broadcasting manually

**Installation:**
```bash
go get github.com/gorilla/websocket
```

**Code Example (Simple):**
```go
// Upgrade HTTP to WebSocket
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Configure properly for production
    },
}

conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
    log.Println(err)
    return
}
defer conn.Close()

// Read messages
for {
    messageType, p, err := conn.ReadMessage()
    if err != nil {
        break
    }
    // Handle message
}
```

**Best for:** Our project! Simple, efficient, no extra complexity.

---

### Option 2: Melody

**What is it?**
A wrapper around Gorilla WebSocket with built-in broadcasting and session management.

**Pros:**
- ✅ Simpler than raw Gorilla
- ✅ Built-in broadcast functionality
- ✅ Session management included

**Cons:**
- ⚠️ Less control over connections
- ⚠️ Opinionated structure (might not fit our clean architecture)
- ⚠️ Smaller community than Gorilla

**Verdict:** Good option if we want less control, but Gorilla gives us more flexibility.

---

### Option 3: Socket.io with Go adapter

**What is it?**
Popular library from Node.js world, has unofficial Go ports.

**Pros:**
- ✅ Fallback to polling if WebSocket unavailable
- ✅ Rooms and namespaces built-in
- ✅ Familiar to JavaScript developers

**Cons:**
- ❌ Not idiomatic Go
- ❌ Unofficial Go ports (less stable)
- ❌ Adds unnecessary complexity
- ❌ Heavier than native WebSocket

**Verdict:** Overkill for our needs. Stick with native Go solutions.

---

### Option 4: nhooyr.io/websocket

**What is it?**
Modern, minimal WebSocket library for Go. Newer than Gorilla.

**Pros:**
- ✅ Cleaner API than Gorilla
- ✅ Better context.Context support
- ✅ More idiomatic Go
- ✅ Better compression support

**Cons:**
- ⚠️ Smaller community
- ⚠️ Less Stack Overflow answers
- ⚠️ Fewer examples

**Verdict:** Great library, but Gorilla has more learning resources (better for learning).

---

### Final Verdict: Use Gorilla WebSocket

**Why?**
- Most documentation and tutorials
- Works perfectly with our setup
- Easy to find help when stuck
- Battle-tested by thousands of projects
- Simple enough to learn

---

## 4. Recommended Approach

### Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    FRONTEND (Next.js)                    │
│                                                           │
│  ┌──────────────┐    ┌──────────────┐   ┌─────────────┐│
│  │  Chat Page   │    │ Notification │   │ Group Chat  ││
│  │              │    │    Bell      │   │    Page     ││
│  └──────┬───────┘    └──────┬───────┘   └──────┬──────┘│
│         │                   │                   │        │
│         └───────────────────┼───────────────────┘        │
│                             │                            │
│                   ┌─────────▼──────────┐                 │
│                   │  WebSocket Client  │                 │
│                   │  (React hook)      │                 │
│                   └─────────┬──────────┘                 │
└─────────────────────────────┼──────────────────────────┘
                              │
                    WebSocket Connection
                              │
┌─────────────────────────────▼──────────────────────────┐
│                    BACKEND (Go)                         │
│                                                          │
│  ┌────────────────────────────────────────────────────┐│
│  │         WebSocket Handler (/ws)                    ││
│  │  - Authenticate user (session cookie)              ││
│  │  - Upgrade HTTP → WebSocket                        ││
│  │  - Store connection in ConnectionManager           ││
│  └────────────────┬───────────────────────────────────┘│
│                   │                                      │
│  ┌────────────────▼───────────────────────────────────┐│
│  │         Connection Manager                         ││
│  │  - Map[UserID] → []*websocket.Conn                ││
│  │  - Send message to specific user(s)               ││
│  │  - Broadcast to conversation/group                ││
│  └────────────────┬───────────────────────────────────┘│
│                   │                                      │
│  ┌────────────────▼───────────────────────────────────┐│
│  │         Message Handler                            ││
│  │  - Parse incoming messages                         ││
│  │  - Save to database (messages table)              ││
│  │  - Broadcast to recipients                        ││
│  └────────────────┬───────────────────────────────────┘│
│                   │                                      │
│  ┌────────────────▼───────────────────────────────────┐│
│  │         Message Use Case                           ││
│  │  - Business logic                                  ││
│  │  - Check permissions                               ││
│  │  - Create notifications                            ││
│  └────────────────┬───────────────────────────────────┘│
│                   │                                      │
│  ┌────────────────▼───────────────────────────────────┐│
│  │         Message Repository (PostgreSQL)            ││
│  │  - INSERT into messages                            ││
│  │  - SELECT from conversations                       ││
│  └────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
```

### Key Components

**1. Connection Manager (Backend)**
- Keeps track of who's online
- Maps User ID → WebSocket Connection(s)
- Handles sending messages to specific users
- Manages connection lifecycle (connect/disconnect)

**2. Message Types**
```json
// Client → Server
{
  "type": "send_message",
  "conversation_id": 123,
  "content": "Hey there!"
}

// Server → Client
{
  "type": "new_message",
  "message": {
    "id": 456,
    "conversation_id": 123,
    "author_id": 789,
    "content": "Hey there!",
    "created_at": "2026-01-26T..."
  }
}

// Notification
{
  "type": "notification",
  "notification": {
    "id": 111,
    "type": "comment_on_post",
    "read": false,
    "metadata": {...}
  }
}
```

**3. Frontend Hook (React)**
```typescript
// Custom React hook for WebSocket
function useWebSocket() {
  const [messages, setMessages] = useState([]);
  const ws = useRef(null);

  useEffect(() => {
    // Connect to WebSocket
    ws.current = new WebSocket('ws://localhost:8080/ws');

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'new_message') {
        setMessages(prev => [...prev, data.message]);
      }
    };

    return () => ws.current.close();
  }, []);

  const sendMessage = (conversationId, content) => {
    ws.current.send(JSON.stringify({
      type: 'send_message',
      conversation_id: conversationId,
      content: content
    }));
  };

  return { messages, sendMessage };
}
```

---

## 5. How It Will Work (Architecture)

### Flow: Sending a Direct Message

**Step-by-step:**

1. **User A types message in chat**
   - Frontend: `sendMessage(conversationId, "Hello!")`

2. **Browser sends via WebSocket**
   - Client → Server: JSON message packet

3. **Backend receives message**
   - WebSocket handler parses JSON
   - Extracts user ID from connection context

4. **Business logic validation**
   - Check: Is User A member of this conversation?
   - Check: Is conversation active?

5. **Save to database**
   - INSERT into `messages` table
   - Get message ID back

6. **Find recipients**
   - Query `conversation_members` table
   - Get all member user IDs

7. **Broadcast to online users**
   - Connection Manager looks up User B's connection
   - If online: Send message immediately
   - If offline: Skip (they'll load from DB on login)

8. **User B receives message**
   - WebSocket client receives JSON
   - React updates UI instantly
   - Message appears in chat!

### Flow: Group Chat

**Same as direct message, but:**
- Multiple recipients (3+ users)
- Broadcast to ALL online members
- Each receives identical message

### Flow: Notifications

**Example: Someone likes your post**

1. **User B likes User A's post**
   - HTTP POST `/posts/123/reactions` (existing endpoint)

2. **Backend creates reaction**
   - INSERT into `post_reactions` table

3. **Trigger notification**
   - INSERT into `notifications` table
   - Get User A's ID (post author)

4. **Check if User A is online**
   - Connection Manager checks for active connection

5. **Send via WebSocket**
   - If online: Push notification immediately
   - If offline: They'll see red badge on next login

6. **Frontend updates**
   - Notification bell shows red dot
   - Count increments
   - Toast notification pops up

---

## 6. Implementation Plan (Step by Step)

### Phase 1: Backend Foundation

**Task 1.1: Add Gorilla WebSocket**
```bash
cd backend
go get github.com/gorilla/websocket
```

**Task 1.2: Create Connection Manager**
- File: `internal/transport/websocket/manager.go`
- Track online users
- Send/broadcast methods

**Task 1.3: Create WebSocket Handler**
- File: `internal/transport/websocket/handler.go`
- Upgrade HTTP connection
- Authenticate user (session cookie)
- Register connection with manager

**Task 1.4: Message Routing**
- Parse incoming message types
- Route to appropriate handler

**Task 1.5: Database Integration**
- Use existing message repository
- Save messages to DB
- Query conversation members

---

### Phase 2: Message Features

**Task 2.1: Direct Messages**
- Handle "send_message" type
- Validate sender is in conversation
- Broadcast to recipient

**Task 2.2: Group Chat**
- Same as direct, but multiple recipients
- Broadcast to all conversation members

**Task 2.3: Message History**
- HTTP endpoint: `GET /conversations/:id/messages`
- Load last 50 messages on chat open
- Pagination for older messages

**Task 2.4: Typing Indicators (Optional)**
- Broadcast "user is typing..." to conversation
- Don't save to DB (ephemeral)

---

### Phase 3: Notifications

**Task 3.1: Notification Broadcaster**
- When notification created, check if user online
- Send via WebSocket if available

**Task 3.2: Mark as Read**
- HTTP endpoint: `PATCH /notifications/:id/read`
- Update via WebSocket too

**Task 3.3: Notification Types**
- Follow requests
- Comment replies
- Reactions
- Group invites
- Event reminders

---

### Phase 4: Frontend Integration

**Task 4.1: WebSocket Context**
- Create React Context for WebSocket
- Connect on app load (if authenticated)
- Reconnect on disconnect

**Task 4.2: Chat UI**
- Message list component
- Input field
- Send button
- Auto-scroll to bottom

**Task 4.3: Notification Bell**
- Show unread count
- Dropdown with recent notifications
- Click to mark as read

**Task 4.4: Online Status**
- Show green dot for online users
- Gray dot for offline

---

### Phase 5: Polish & Testing

**Task 5.1: Error Handling**
- Connection drops: Auto-reconnect
- Failed messages: Retry or show error

**Task 5.2: Security**
- Rate limiting on WebSocket messages
- Input validation (max message length)
- XSS prevention (sanitize HTML)

**Task 5.3: Testing**
- Unit tests for Connection Manager
- Integration tests for message flow
- Frontend: Test with mock WebSocket

---

## 7. Alternative Approaches (And Why We Skip Them)

### Alternative 1: Polling (Old School)

**How it works:**
- Frontend asks "Any new messages?" every 2 seconds
- Backend checks database, returns new messages

**Pros:**
- Simple to implement
- Works everywhere (no special server support)

**Cons:**
- ❌ Wasteful (constant empty requests)
- ❌ Delayed messages (up to 2 second lag)
- ❌ High server load with many users
- ❌ Battery drain on mobile

**Verdict:** No. WebSocket is way better.

---

### Alternative 2: Server-Sent Events (SSE)

**How it works:**
- Server pushes updates to client
- One-way only (server → client)

**Pros:**
- ✅ Simpler than WebSocket
- ✅ Auto-reconnect built-in
- ✅ Works over HTTP

**Cons:**
- ❌ One-way only (need separate HTTP for sending)
- ❌ Less efficient than WebSocket
- ❌ Browser connection limits (6 per domain)

**Verdict:** Could work for notifications only, but WebSocket handles both chat + notifications.

---

### Alternative 3: Third-Party Services

**Examples:** Pusher, Ably, Firebase, PubNub

**How it works:**
- Pay external service to handle WebSockets
- They manage connections, you just send/receive

**Pros:**
- ✅ Zero infrastructure setup
- ✅ Scalability handled for you
- ✅ Built-in features (presence, typing, etc.)

**Cons:**
- ❌ Costs money (per message/connection)
- ❌ Vendor lock-in
- ❌ Data privacy concerns (messages through third-party)
- ❌ Need internet to test (no local dev)

**Verdict:** Overkill for our project. We can handle this ourselves!

---

### Alternative 4: Redis Pub/Sub

**How it works:**
- Backend instances publish messages to Redis
- Subscribers receive and broadcast via WebSocket

**When to use:**
- Multiple backend servers (horizontal scaling)
- Users connected to different servers need to chat

**Current situation:**
- We have 1 backend server (for now)
- Don't need Redis yet!

**Verdict:** Future optimization when we scale. Not needed now.

---

## 8. Security Considerations

### 1. Authentication

**Problem:** Anyone could connect to WebSocket endpoint!

**Solution:**
```go
// Authenticate BEFORE upgrading to WebSocket
userID, ok := middleware.GetUserID(r.Context())
if !ok {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

// Upgrade with user context
conn, err := upgrader.Upgrade(w, r, nil)
// ... store userID with connection
```

**Key points:**
- ✅ Validate session token from cookie
- ✅ Extract user ID before upgrade
- ✅ Attach user ID to connection object
- ✅ Reject if not authenticated

---

### 2. Message Validation

**Problems:**
- XSS attacks (malicious HTML/JavaScript in messages)
- Message bombing (spam 1000 messages/second)
- Oversized messages (send 10MB text)

**Solutions:**

**A. Input Sanitization**
```go
// Max message length
const MaxMessageLength = 5000 // characters

if len(content) > MaxMessageLength {
    return errors.New("message too long")
}

// Sanitize HTML (on frontend too!)
// Use library like bluemonday if allowing formatted text
```

**B. Rate Limiting**
```go
// Per-user rate limit
// Example: 10 messages per second max
type RateLimiter struct {
    messages map[int][]time.Time // userID -> timestamps
}

func (rl *RateLimiter) Allow(userID int) bool {
    // Check if user exceeded limit
    // Remove old timestamps
    // Return true/false
}
```

**C. Permission Checks**
```go
// Before sending message:
// 1. Is user member of this conversation?
// 2. Is conversation active (not deleted)?
// 3. Is user blocked by recipient?
```

---

### 3. Connection Management

**Problems:**
- Memory leaks (connections not cleaned up)
- Duplicate connections (user opens multiple tabs)
- Stale connections (user disconnected but still tracked)

**Solutions:**

**A. Cleanup on Disconnect**
```go
defer func() {
    conn.Close()
    manager.Unregister(userID, conn)
}()
```

**B. Multiple Connections per User**
```go
// Allow multiple connections (phone + laptop)
type ConnectionManager struct {
    connections map[int][]*websocket.Conn // userID -> []conn
}

// Send to ALL user's connections
func (cm *ConnectionManager) SendToUser(userID int, msg []byte) {
    for _, conn := range cm.connections[userID] {
        conn.WriteMessage(websocket.TextMessage, msg)
    }
}
```

**C. Heartbeat (Ping/Pong)**
```go
// Detect dead connections
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
            return // Connection dead, cleanup
        }
    }
}
```

---

### 4. CORS & Origin Checks

**Problem:** Malicious website connects to our WebSocket!

**Solution:**
```go
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        // Allow only our frontend
        return origin == "http://localhost:3000" ||
               origin == "https://yourdomain.com"
    },
}
```

---

### 5. Data Privacy

**Considerations:**
- Messages stored in database (encrypted at rest?)
- WebSocket traffic (use WSS in production, not WS)
- Admin access (can admins read private messages?)

**Best practices:**
- ✅ Use `wss://` (WebSocket Secure) in production
- ✅ HTTPS for all HTTP endpoints too
- ✅ Consider end-to-end encryption for sensitive chats
- ✅ Clear data retention policy

---

## 9. Resources & Learning Materials

### Official Documentation

**Gorilla WebSocket:**
- 📖 Docs: https://pkg.go.dev/github.com/gorilla/websocket
- 💻 GitHub: https://github.com/gorilla/websocket
- 📝 Examples: https://github.com/gorilla/websocket/tree/master/examples

**WebSocket Protocol (RFC 6455):**
- 📖 Spec: https://datatracker.ietf.org/doc/html/rfc6455
- (You don't need to read this, but it's there if curious!)

---

### Tutorials & Guides

**Text Tutorials:**
1. "Building Real-Time Chat with Go and WebSockets"
   - https://tutorialedge.net/golang/go-websocket-tutorial/

2. "Gorilla WebSocket Chat Example"
   - https://github.com/gorilla/websocket/tree/master/examples/chat
   - Full working chat app example!

3. "WebSocket Best Practices"
   - https://ably.com/topic/websocket-best-practices

**Video Tutorials:**
1. "Real-Time Chat with Go, React, and WebSockets" (YouTube)
   - Search: "golang websocket chat tutorial"
   - Many great videos available

2. "WebSockets in 100 Seconds" by Fireship
   - Quick intro to WebSocket concept

---

### Frontend Resources

**React WebSocket:**
1. Custom hook example:
   - https://github.com/robtaussig/react-use-websocket

2. Tutorial:
   - "Using WebSockets in React"
   - https://blog.logrocket.com/websockets-tutorial-how-to-go-real-time-with-node-and-react/

---

### Example Projects

**Similar to ours:**
1. **Gorilla Chat Example** (Simple, great for learning)
   - https://github.com/gorilla/websocket/tree/master/examples/chat

2. **Go Chat Application** (More complete)
   - Search GitHub: "golang websocket chat"
   - Many open-source examples available

3. **Real-World Example:**
   - Slack, Discord, WhatsApp Web all use WebSocket!
   - Can inspect with browser DevTools → Network → WS

---

### Testing Tools

**Browser DevTools:**
- Chrome/Firefox → Network tab → WS filter
- See all WebSocket messages in real-time!

**Postman:**
- Can test WebSocket connections
- Send/receive messages manually

**wscat (CLI tool):**
```bash
npm install -g wscat
wscat -c ws://localhost:8080/ws
```

---

### Books (If You Want Deep Dive)

1. **"Network Programming with Go"** by Jan Newmarch
   - Chapter on WebSockets

2. **"Web Development with Go"** by Shiju Varghese
   - Covers real-time apps

---

## Final Thoughts

### What Makes This Approach Good?

✅ **Simple:** Using standard Go patterns we already have
✅ **Efficient:** WebSocket is lightweight and fast
✅ **Scalable:** Can add Redis Pub/Sub later if needed
✅ **Secure:** Works with our existing auth system
✅ **Learnable:** Tons of resources and examples
✅ **No vendor lock-in:** We own the code

---

### Success Criteria

**You'll know it's working when:**
1. ✅ User A sends message, User B sees it instantly (if online)
2. ✅ Group messages delivered to all members
3. ✅ Notifications pop up without page refresh
4. ✅ Connection stays alive (doesn't disconnect randomly)
5. ✅ Works on multiple devices (phone + laptop)
6. ✅ Messages saved to DB even if recipient offline

---

### Next Steps

**After reading this document:**

1. **Discuss with team:**
   - Does everyone agree with this approach?
   - Any concerns or questions?
   - Who works on backend vs frontend?

2. **Start small:**
   - Build basic WebSocket connection first
   - Then add direct messages
   - Then group chat
   - Then notifications
   - Polish last!

3. **Ask for help when stuck:**
   - Check examples in Gorilla repo
   - Search Stack Overflow
   - Ask teammates
   - Learning is a process!

---

### Remember

- 🧠 **ADHD-friendly tip:** Implement one feature at a time. Don't try to build everything at once!
- 💪 **You got this:** Thousands of developers have built this before. If they can, you can too!
- 🐛 **Bugs are normal:** Things won't work perfectly first try. That's part of coding!
- 🎯 **Focus on learning:** This is a great learning opportunity!

---

**Questions?**
Save this doc, share with team, and start building! 🚀

---

**Last Updated:** January 26, 2026
**Version:** 1.0
**Status:** Ready for team review
