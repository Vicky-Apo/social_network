BEGIN;

-- =====================================================
-- Users
-- =====================================================
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    profile_visibility TEXT NOT NULL DEFAULT 'public', -- public | private
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- =====================================================
-- Follows
-- =====================================================
CREATE TABLE follows (
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id)
);

CREATE INDEX idx_follows_followee
ON follows(followee_id);


-- =====================================================
-- Posts
-- =====================================================
CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    visibility TEXT NOT NULL, -- public | followers | selected | profile
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_author_created
ON posts(author_id, created_at DESC);

CREATE INDEX idx_posts_created
ON posts(created_at DESC);


-- Selected viewers (for visibility = selected)
CREATE TABLE post_allowed_viewers (
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    viewer_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, viewer_id)
);

CREATE INDEX idx_pav_viewer
ON post_allowed_viewers(viewer_id);


-- =====================================================
-- Groups
-- =====================================================
CREATE TABLE groups (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE group_members (
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role TEXT DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_members_user
ON group_members(user_id);


CREATE TABLE group_posts (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    author_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_group_posts_created
ON group_posts(group_id, created_at DESC);


-- =====================================================
-- Reactions
-- =====================================================
CREATE TABLE reactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_reactions_unique
ON reactions(user_id, post_id);


-- =====================================================
-- Conversations / Messages
-- =====================================================
CREATE TABLE conversations (
    id BIGSERIAL PRIMARY KEY,
    is_direct BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE conversation_participants (
    conversation_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id BIGINT,
    PRIMARY KEY (conversation_id, user_id)
);


CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    body TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- =====================================================
-- Notifications
-- =====================================================
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    actor_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    entity_id BIGINT,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_created
ON notifications(user_id, created_at DESC);


COMMIT;