/* =========================
   ENUM TYPES
   ========================= */
CREATE TYPE post_visibility AS ENUM ('public', 'followers', 'private');
CREATE TYPE conversation_type AS ENUM ('direct', 'private_group', 'group');
CREATE TYPE conversation_role AS ENUM ('member', 'admin');
CREATE TYPE event_response AS ENUM ('going', 'not_going');
CREATE TYPE notification_type AS ENUM ('follow_request', 'group_invitation', 'group_join_request', 'event_created');

/* =========================
   USERS
   ========================= */
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,

  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  date_of_birth DATE NOT NULL,

  avatar_path TEXT,
  nickname TEXT,
  about TEXT,

  is_public BOOLEAN NOT NULL DEFAULT FALSE,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

/* =========================
   FOLLOW SYSTEM
   ========================= */
CREATE TABLE follow_requests (
  requester_id BIGINT NOT NULL,
  target_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (requester_id, target_id),
  CHECK (requester_id <> target_id),

  FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE follows (
  follower_id BIGINT NOT NULL,
  following_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (follower_id, following_id),
  CHECK (follower_id <> following_id),

  FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   GROUPS
   ========================= */
CREATE TABLE groups (
  id BIGSERIAL PRIMARY KEY,
  creator_id BIGINT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_members (
  group_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (group_id, user_id),
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_invitations (
  group_id BIGINT NOT NULL,
  invited_user_id BIGINT NOT NULL,
  invited_by_user_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (group_id, invited_user_id),
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (invited_user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (invited_by_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_join_requests (
  group_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (group_id, user_id),
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   POSTS & COMMENTS
   ========================= */
CREATE TABLE posts (
  id BIGSERIAL PRIMARY KEY,
  author_id BIGINT NOT NULL,
  group_id BIGINT,
  content TEXT,
  media_path TEXT,

  visibility post_visibility NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE TABLE post_allowed_users (
  post_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,

  PRIMARY KEY (post_id, user_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE comments (
  id BIGSERIAL PRIMARY KEY,
  post_id BIGINT NOT NULL,
  author_id BIGINT NOT NULL,
  content TEXT,
  media_path TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   EVENTS
   ========================= */
CREATE TABLE events (
  id BIGSERIAL PRIMARY KEY,
  group_id BIGINT NOT NULL,
  creator_id BIGINT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  event_time TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE event_responses (
  event_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  response event_response NOT NULL,
  responded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (event_id, user_id),
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   UNIFIED CHAT
   ========================= */
CREATE TABLE conversations (
  id BIGSERIAL PRIMARY KEY,
  type conversation_type NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE conversation_members (
  conversation_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role conversation_role NOT NULL DEFAULT 'member',
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (conversation_id, user_id),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE messages (
  id BIGSERIAL PRIMARY KEY,
  conversation_id BIGINT NOT NULL,
  sender_id BIGINT NOT NULL,
  content TEXT,
  emoji TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   NOTIFICATIONS
   ========================= */
CREATE TABLE notifications (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type notification_type NOT NULL,

  entity_type TEXT NOT NULL,
  entity_id BIGINT NOT NULL,

  metadata JSONB NOT NULL DEFAULT '{}',
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   INDEXES
   ========================= */
CREATE INDEX idx_unread_notifications
ON notifications(user_id)
WHERE is_read = false;

CREATE INDEX idx_messages_conversation
ON messages(conversation_id);
