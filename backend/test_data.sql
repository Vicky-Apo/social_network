-- Test data for comments and reactions feature
-- Run this with: psql -h localhost -p 5432 -U vapostol -d social_network -f test_data.sql

-- Insert test user
INSERT INTO users (id, email, password_hash, first_name, last_name, date_of_birth, nickname, about_me, is_public, created_at, updated_at)
VALUES (1, 'test@example.com', '$2a$10$hash', 'Test', 'User', '1990-01-01', 'testuser', 'Test user for development', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert test post
INSERT INTO posts (id, user_id, content, privacy, created_at, updated_at)
VALUES (1, 1, 'This is a test post for testing comments and reactions!', 'public', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

SELECT 'Test data inserted successfully!' as result;
