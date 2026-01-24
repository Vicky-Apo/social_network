/* =========================
   TRIGGERS
   ========================= */

DROP TRIGGER IF EXISTS trg_create_group_conversation ON groups;
DROP TRIGGER IF EXISTS trg_prevent_group_post_categories ON post_categories;

DROP TRIGGER IF EXISTS trg_events_updated_at ON events;
DROP TRIGGER IF EXISTS trg_messages_updated_at ON messages;
DROP TRIGGER IF EXISTS trg_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS trg_posts_updated_at ON posts;
DROP TRIGGER IF EXISTS trg_group_join_requests_updated_at ON group_join_requests;
DROP TRIGGER IF EXISTS trg_group_invitations_updated_at ON group_invitations;
DROP TRIGGER IF EXISTS trg_groups_updated_at ON groups;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;


/* =========================
   FUNCTIONS
   ========================= */

DROP FUNCTION IF EXISTS get_message_reaction_summary(BIGINT);
DROP FUNCTION IF EXISTS get_comment_reaction_summary(BIGINT);
DROP FUNCTION IF EXISTS get_post_reaction_summary(BIGINT);
DROP FUNCTION IF EXISTS get_unread_notifications_count(BIGINT);
DROP FUNCTION IF EXISTS create_group_conversation();
DROP FUNCTION IF EXISTS set_updated_at();
DROP FUNCTION IF EXISTS prevent_group_post_categories();


/* =========================
   INDEXES
   ========================= */

DROP INDEX IF EXISTS idx_notifications_user_unread;
DROP INDEX IF EXISTS idx_notifications_user;
DROP INDEX IF EXISTS idx_conversation_members_user;
DROP INDEX IF EXISTS idx_messages_conversation_created_at;
DROP INDEX IF EXISTS idx_message_reactions_message;
DROP INDEX IF EXISTS idx_comment_reactions_comment;
DROP INDEX IF EXISTS idx_post_reactions_post;
DROP INDEX IF EXISTS idx_comments_post_created_at;
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_group_created_at;
DROP INDEX IF EXISTS idx_posts_author_created_at;


/* =========================
   TABLES
   ========================= */

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS group_conversations;
DROP TABLE IF EXISTS message_reactions;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_members;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS event_responses;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS comment_reactions;
DROP TABLE IF EXISTS post_reactions;
DROP TABLE IF EXISTS post_categories;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS post_allowed_users;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS group_join_requests;
DROP TABLE IF EXISTS group_invitations;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS follow_requests;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;


/* =========================
   TYPES
   ========================= */

DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS reaction_type;
DROP TYPE IF EXISTS event_response;
DROP TYPE IF EXISTS conversation_role;
DROP TYPE IF EXISTS conversation_type;
DROP TYPE IF EXISTS post_visibility;
