/* =========================
   RECREATE POST CATEGORIES
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
