-- Add updated_at column to post_reactions
ALTER TABLE post_reactions 
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Add updated_at column to comment_reactions
ALTER TABLE comment_reactions 
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Create indexes for better query performance
CREATE INDEX idx_post_reactions_updated_at ON post_reactions(updated_at);
CREATE INDEX idx_comment_reactions_updated_at ON comment_reactions(updated_at);
