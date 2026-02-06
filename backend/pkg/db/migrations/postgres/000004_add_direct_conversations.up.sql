-- Adds a canonical pair key to direct conversations to prevent duplicates
-- under concurrent "get or create" requests.
ALTER TABLE conversations
ADD COLUMN direct_pair_key TEXT;

UPDATE conversations c
SET direct_pair_key = sub.pair_key
FROM (
  SELECT
    cm.conversation_id,
    MIN(cm.user_id)::TEXT || ':' || MAX(cm.user_id)::TEXT AS pair_key
  FROM conversation_members cm
  JOIN conversations c2 ON c2.id = cm.conversation_id
  WHERE c2.type = 'direct'
  GROUP BY cm.conversation_id
  HAVING COUNT(*) = 2
) AS sub
WHERE c.id = sub.conversation_id;

CREATE UNIQUE INDEX idx_direct_conversation_pair
ON conversations (direct_pair_key)
WHERE type = 'direct' AND direct_pair_key IS NOT NULL;

DROP TABLE IF EXISTS direct_conversations;
