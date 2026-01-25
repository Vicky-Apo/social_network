/* =========================
   ENUM TYPES
   ========================= */

CREATE TYPE post_visibility AS ENUM (
  'public',
  'followers',
  'private'
);

CREATE TYPE conversation_type AS ENUM (
  'direct',         -- exactly 2 users
  'private_group',  -- 3+ users, non-group chat
  'group'           -- group chat
);

CREATE TYPE conversation_role AS ENUM (
  'member',
  'admin'
);

CREATE TYPE event_response AS ENUM (
  'pending',
  'going',
  'not_going'
);

CREATE TYPE reaction_type AS ENUM (
  'like',
  'dislike'
);

CREATE TYPE notification_type AS ENUM (
  'follow_request',
  'group_invitation',
  'group_join_request',
  'event_created',
  'post_reaction',
  'comment_reaction',
  'comment_on_post'
);


/* =========================
   USERS & AUTH
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

CREATE TABLE sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  session_token TEXT NOT NULL,
  user_agent TEXT,
  ip_address TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);


/* =========================
   FOLLOWERS
   ========================= */

CREATE TABLE follow_requests (
  id BIGSERIAL PRIMARY KEY,
  requester_id BIGINT NOT NULL,
  target_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE follows (
  follower_id BIGINT NOT NULL,
  following_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (follower_id, following_id),
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
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
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
  id BIGSERIAL PRIMARY KEY,
  group_id BIGINT NOT NULL,
  inviter_id BIGINT NOT NULL,
  invitee_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (invitee_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_join_requests (
  id BIGSERIAL PRIMARY KEY,
  group_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
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
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
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
  content TEXT NOT NULL,
  media_path TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

/* =========================
   CATEGORIES (NON-GROUP POSTS)
   ========================= */

CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

INSERT INTO categories (name, description) VALUES
  ('Programming & Software Development', 'Languages, frameworks, algorithms, design patterns, code reviews.'),
  ('Web Development', 'Frontend, backend, APIs, performance, accessibility, browsers.'),
  ('DevOps & Infrastructure', 'Linux, Docker, Kubernetes, CI/CD, cloud, monitoring, scaling.'),
  ('Databases & Data Engineering', 'SQL/NoSQL, schema design, migrations, performance, backups.'),
  ('Cybersecurity & Privacy', 'Vulnerabilities, authentication, encryption, secure coding, audits.'),
  ('AI, Machine Learning & Data Science', 'Models, training, inference, tooling, real-world applications.'),
  ('Operating Systems & Low-Level Tech', 'Linux, kernels, memory, processes, networking internals.'),
  ('Hardware & Embedded Systems', 'CPUs, GPUs, IoT, microcontrollers, performance tuning.'),
  ('Tools, Editors & Productivity', 'IDEs, CLIs, workflows, automation, developer ergonomics.'),
  ('Architecture, Scalability & System Design', 'Distributed systems, microservices, trade-offs, failures.');

CREATE TABLE post_categories (
  post_id BIGINT NOT NULL,
  category_id BIGINT NOT NULL,
  PRIMARY KEY (post_id, category_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);


/* =========================
   REACTIONS
   ========================= */

CREATE TABLE post_reactions (
  post_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  reaction reaction_type NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (post_id, user_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE comment_reactions (
  comment_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  reaction reaction_type NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (comment_id, user_id),
  FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE event_responses (
  event_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  response event_response NOT NULL DEFAULT 'pending',
  responded_at TIMESTAMPTZ,
  PRIMARY KEY (event_id, user_id),
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);


/* =========================
   CONVERSATIONS & CHAT
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
  media_path TEXT, -- for images or files
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE message_reactions (
  message_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  emoji TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (message_id, user_id, emoji),
  FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_conversations (
  group_id BIGINT PRIMARY KEY,
  conversation_id BIGINT NOT NULL UNIQUE,
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
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
  metadata JSONB,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);



/*================CONSTRAINTS ==============*/

-- UNIQUE CONSTRAINTS 
ALTER TABLE follow_requests
ADD CONSTRAINT unique_follow_request
UNIQUE (requester_id, target_id);

ALTER TABLE group_invitations
ADD CONSTRAINT unique_group_invitation
UNIQUE (group_id, invitee_id);

ALTER TABLE group_join_requests
ADD CONSTRAINT unique_group_join_request
UNIQUE (group_id, user_id);


-- CHECK CONSTRAINTS
ALTER TABLE posts
ADD CONSTRAINT post_content_or_media_check
CHECK (content IS NOT NULL OR media_path IS NOT NULL);

ALTER TABLE messages
ADD CONSTRAINT message_content_or_media_check
CHECK (content IS NOT NULL OR media_path IS NOT NULL);

ALTER TABLE message_reactions
ADD CONSTRAINT emoji_length_check
CHECK (char_length(emoji) <= 8);


--LOGICAL INVARIANT (TRIGGER)

--ensure that only non-group posts can have categories assigned
CREATE OR REPLACE FUNCTION prevent_group_post_categories()
RETURNS TRIGGER AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM posts
    WHERE id = NEW.post_id
      AND group_id IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'Categories are only allowed for non-group posts';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_prevent_group_post_categories
BEFORE INSERT ON post_categories
FOR EACH ROW
EXECUTE FUNCTION prevent_group_post_categories();


/* =========================
   UPDATED_AT TRIGGER
   ========================= */

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;



/* USERS */
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* GROUPS */
CREATE TRIGGER trg_groups_updated_at
BEFORE UPDATE ON groups
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* GROUP INVITATIONS */
CREATE TRIGGER trg_group_invitations_updated_at
BEFORE UPDATE ON group_invitations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* GROUP JOIN REQUESTS */
CREATE TRIGGER trg_group_join_requests_updated_at
BEFORE UPDATE ON group_join_requests
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* POSTS */
CREATE TRIGGER trg_posts_updated_at
BEFORE UPDATE ON posts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* COMMENTS */
CREATE TRIGGER trg_comments_updated_at
BEFORE UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* MESSAGES */
CREATE TRIGGER trg_messages_updated_at
BEFORE UPDATE ON messages
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

/* EVENTS */
CREATE TRIGGER trg_events_updated_at
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();






/* ========== INDEXES ==================== */


--- POST INDEXES
CREATE INDEX idx_posts_author_created_at
ON posts (author_id, created_at DESC);

CREATE INDEX idx_posts_group_created_at
ON posts (group_id, created_at DESC);

CREATE INDEX idx_posts_created_at
ON posts (created_at DESC);


--COMMENTS INDEXES
CREATE INDEX idx_comments_post_created_at
ON comments (post_id, created_at ASC);


--REACTIONS INDEXES
CREATE INDEX idx_post_reactions_post
ON post_reactions (post_id);

CREATE INDEX idx_comment_reactions_comment
ON comment_reactions (comment_id);

CREATE INDEX idx_message_reactions_message
ON message_reactions (message_id);


--CHAT INDEXES
CREATE INDEX idx_messages_conversation_created_at
ON messages (conversation_id, created_at DESC);

CREATE INDEX idx_conversation_members_user
ON conversation_members (user_id);

--NOTIFICATIONS INDEXES
CREATE INDEX idx_notifications_user
ON notifications (user_id);

CREATE INDEX idx_notifications_user_unread
ON notifications (user_id)
WHERE is_read = false;



/* AUTO-CREATE GROUP CHAT */

CREATE OR REPLACE FUNCTION create_group_conversation()
RETURNS TRIGGER AS $$
DECLARE
  conv_id BIGINT;
BEGIN
  INSERT INTO conversations (type)
  VALUES ('group')
  RETURNING id INTO conv_id;

  INSERT INTO group_conversations (group_id, conversation_id)
  VALUES (NEW.id, conv_id);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_create_group_conversation
AFTER INSERT ON groups
FOR EACH ROW
EXECUTE FUNCTION create_group_conversation();



/* ========== FUNCTIONS ==================== */
-- functions are creating repeatable logic for common queries

/*
  Returns the number of unread notifications for a user.
  Solves:
  - notification bell badge count
  - fast unread check using partial index
*/
CREATE OR REPLACE FUNCTION get_unread_notifications_count(p_user_id BIGINT)
RETURNS INTEGER AS $$
BEGIN
  RETURN (
    SELECT COUNT(*)
    FROM notifications
    WHERE user_id = p_user_id
      AND is_read = false
  );
END;
$$ LANGUAGE plpgsql STABLE;


/*
  Returns aggregated reaction counts for a post.
  Solves:
  - like/dislike counters under posts
  - avoids repeating GROUP BY logic in application code
*/
CREATE OR REPLACE FUNCTION get_post_reaction_summary(p_post_id BIGINT)
RETURNS TABLE (
  reaction reaction_type,
  count BIGINT
) AS $$
BEGIN
  RETURN QUERY
  SELECT r.reaction, COUNT(*)::BIGINT
  FROM post_reactions r
  WHERE r.post_id = p_post_id
  GROUP BY r.reaction;
END;
$$ LANGUAGE plpgsql STABLE;


/*
  Returns aggregated reaction counts for a comment.
  Solves:
  - like/dislike counters under comments
  - consistent reaction summary logic
*/
CREATE OR REPLACE FUNCTION get_comment_reaction_summary(p_comment_id BIGINT)
RETURNS TABLE (
  reaction reaction_type,
  count BIGINT
) AS $$
BEGIN
  RETURN QUERY
  SELECT r.reaction, COUNT(*)::BIGINT
  FROM comment_reactions r
  WHERE r.comment_id = p_comment_id
  GROUP BY r.reaction;
END;
$$ LANGUAGE plpgsql STABLE;


/*
  Returns aggregated emoji reaction counts for a message.
  Solves:
  - emoji reaction rendering in chats
  - avoids per-message aggregation logic in application code
*/
CREATE OR REPLACE FUNCTION get_message_reaction_summary(p_message_id BIGINT)
RETURNS TABLE (
  emoji TEXT,
  count BIGINT
) AS $$
BEGIN
  RETURN QUERY
  SELECT mr.emoji, COUNT(*)::BIGINT
  FROM message_reactions mr
  WHERE mr.message_id = p_message_id
  GROUP BY mr.emoji;
END;
$$ LANGUAGE plpgsql STABLE;
