CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

CREATE TABLE post_categories (
  post_id BIGINT NOT NULL,
  category_id BIGINT NOT NULL,
  PRIMARY KEY (post_id, category_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

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
