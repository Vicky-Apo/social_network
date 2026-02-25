# Project Askings

This document consolidates the project requirements for the social network (Facebook-like) platform. It is split into **Backend** and **Frontend** so both sides can track scope, constraints, and acceptance criteria.

## Backend Askings (Focus)

### Core Features
- Authentication with **sessions + cookies** (no token-only auth).
- Users, profiles, and privacy (public/private).
- Followers and follow-requests with accept/decline workflow.
- Posts with privacy levels and comments; support media attachments.
- Groups with invitations, join requests, member-only posts, and group events.
- Notifications for follow requests, group invites, join requests, and new group events.
- Real-time chat via WebSockets (private and group chat).

### Data Storage
- Use **PostgreSQL** (explicit requirement for this project instance).
- Design schema based on an ERD with explicit constraints and indexes.
- Store image/GIF metadata in the DB and the binary files on the filesystem.

### Migrations
- Provide versioned SQL migrations.
- Keep migrations in a dedicated folder in the backend (similar structure to the expected tree).
- Migrations must run automatically on backend startup.

### Auth Requirements
- Registration fields:
  - Required: Email, Password, First Name, Last Name, Date of Birth.
  - Optional but present in the form: Avatar/Image, Nickname, About Me.
- Login persists via session cookie until explicit logout.

### Followers
- Follow requests are required for private profiles.
- Public profiles auto-accept follow requests.
- Unfollow only allowed if already following.

### Profiles
- Profile includes: user info (except password), activity, posts, followers, following.
- Profile privacy:
  - Public: visible to all users.
  - Private: visible only to followers.
- Users can toggle their own profile privacy.

### Posts and Comments
- Logged-in users can create posts and comments.
- Posts and comments can include **JPEG, PNG, GIF**.
- Post privacy:
  - `public`: visible to all users.
  - `almost private`: visible to followers of author.
  - `private`: visible only to specific followers chosen by author.

### Groups
- Users can create groups (title + description).
- Group membership:
  - Invite-based (invited users must accept).
  - Join-request flow (creator accepts/declines).
  - Members can invite others.
- Group feed: posts/comments visible only to members.

### Group Events
- Group members can create events (title, description, date/time).
- RSVP options at minimum: Going / Not going.

### Chat (WebSockets)
- Direct messages allowed if at least one user follows the other.
- Messages delivered instantly if recipient follows sender or recipient is public.
- Emojis supported.
- Each group has a shared chat room for members.

### Notifications
- Show notifications on every page (distinct from chat messages).
- Minimum events to notify:
  - Follow request (private profile).
  - Group invitation.
  - Join request to group creator.
  - Group event created (for members).
- Additional notifications allowed.

### Containerization
- Backend has its own Docker image.
- Backend exposes ports needed for the frontend and external clients.

### Allowed Packages (Backend)
- Standard Go packages
- Gorilla WebSocket
- golang-migrate (or equivalent)
- sqlite3 (allowed but not used here)
- bcrypt
- gofrs/uuid or google/uuid

## Frontend Askings (Summary)

### Core Responsibilities
- Implement UI/UX for all backend features:
  - Auth (register/login/logout)
  - Profiles with privacy toggle
  - Follow requests + followers/following lists
  - Post creation, comments, media upload
  - Group creation, browsing, membership workflows
  - Group events + RSVP
  - Notifications list + unread indicators
  - Private chat + group chat (WebSockets)

### Framework
- Must use a JS framework (e.g., Next.js, Vue, Svelte, Mithril).

### UX/Performance
- Responsive layout for mobile and desktop.
- Efficient API usage, avoid over-fetching.
- Clear separation of notifications vs chat messages.

### Containerization
- Frontend has its own Docker image.
- Exposes the port needed to serve the client app.
- Communicates with backend over HTTP + WebSockets.

---

If you want, I can expand the **Frontend Askings** into detailed pages/routes and data flows once we lock the backend APIs.
