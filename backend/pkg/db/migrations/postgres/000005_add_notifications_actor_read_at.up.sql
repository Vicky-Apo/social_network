ALTER TABLE notifications
ADD COLUMN actor_id BIGINT,
ADD COLUMN read_at TIMESTAMPTZ;

ALTER TABLE notifications
ADD CONSTRAINT notifications_actor_fk
FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL;

-- Backfill read_at for already-read notifications
UPDATE notifications
SET read_at = created_at
WHERE is_read = true AND read_at IS NULL;

-- Optional indexes for future filtering
CREATE INDEX IF NOT EXISTS idx_notifications_actor
ON notifications (actor_id);

CREATE INDEX IF NOT EXISTS idx_notifications_read_at
ON notifications (read_at);
