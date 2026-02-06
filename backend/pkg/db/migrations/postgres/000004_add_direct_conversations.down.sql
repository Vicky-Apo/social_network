DROP INDEX IF EXISTS idx_direct_conversation_pair;

ALTER TABLE conversations
DROP COLUMN direct_pair_key;
