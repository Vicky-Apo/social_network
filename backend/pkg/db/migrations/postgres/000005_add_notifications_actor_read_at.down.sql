ALTER TABLE notifications
DROP CONSTRAINT IF EXISTS notifications_actor_fk;

DROP INDEX IF EXISTS idx_notifications_actor;
DROP INDEX IF EXISTS idx_notifications_read_at;

ALTER TABLE notifications
DROP COLUMN IF EXISTS actor_id,
DROP COLUMN IF EXISTS read_at;
