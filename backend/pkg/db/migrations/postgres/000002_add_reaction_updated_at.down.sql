-- Drop indexes
DROP INDEX IF EXISTS idx_post_reactions_updated_at;
DROP INDEX IF EXISTS idx_comment_reactions_updated_at;

-- Remove updated_at columns
ALTER TABLE post_reactions DROP COLUMN IF EXISTS updated_at;
ALTER TABLE comment_reactions DROP COLUMN IF EXISTS updated_at;
