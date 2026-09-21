-- =====================================================================
-- Seed profile + social data for testuser1@gmail.com on local DB
-- Password for all created users: Test@1234 (bcrypt)
-- Run: psql -h 127.0.0.1 -p 5433 -U postgres -d kalasetu_db -f seed_test_user1.sql
-- =====================================================================

BEGIN;

-- 0) Support users so likes/comments/followers have real owners
INSERT INTO users (id, email, password, name, bio, profile_picture, location) VALUES
  (1, 'testuser1@gmail.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Test User One', 'Digital artist and illustrator. I draw characters that tell stories.', 'https://i.pravatar.cc/300?img=12', 'Kochi, India'),
  (2, 'support1@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Maya', 'Concept artist, painter.', NULL, NULL),
  (3, 'support2@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Arjun', 'Sculptor working in mixed media.', NULL, NULL),
  (4, 'support3@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Sara', 'Ceramicist.', NULL, NULL),
  (5, 'support4@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Kiran', 'Photographer.', NULL, NULL),
  (6, 'support5@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Nikhil', 'UI/UX designer.', NULL, NULL),
  (7, 'support6@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Anjali', 'Street artist.', NULL, NULL),
  (8, 'support7@test.com', '$2a$10$L3MyJkUxHa7r18rW/7Zspewz1gDDEoxdU7sXV/P3HWbplohZ1ntjm', 'Ravi', 'Illustrator.', NULL, NULL)
ON CONFLICT (email) DO UPDATE
  SET password = EXCLUDED.password,
      name = EXCLUDED.name,
      bio = EXCLUDED.bio,
      profile_picture = EXCLUDED.profile_picture,
      location = EXCLUDED.location;

-- Tidy the test user's basics (re-run safe)
UPDATE users
   SET bio = 'Digital artist and illustrator. I draw characters that tell stories.',
       location = 'Kochi, India',
       profile_picture = 'https://i.pravatar.cc/300?img=12',
       updated_at = NOW()
 WHERE id = 1;

-- 1) Reset previously seeded content so this script is safe to re-run
DELETE FROM achievements WHERE user_id = 1;
DELETE FROM likes     WHERE parent_type = 'post' AND parent_id IN (SELECT id FROM posts WHERE user_id = 1);
DELETE FROM comments  WHERE parent_type = 'post' AND parent_id IN (SELECT id FROM posts WHERE user_id = 1);
DELETE FROM post_media WHERE post_id IN (SELECT id FROM posts WHERE user_id = 1);
DELETE FROM posts WHERE user_id = 1;
DELETE FROM user_follows WHERE follower_id = 1 OR following_id = 1;
DELETE FROM user_labels WHERE user_id = 1;

-- 2) Posts (artworks / recent posts), most recent first -> ord 1..5
INSERT INTO posts (user_id, content, created_at) VALUES
  (1, 'Night sketch — a girl and her fox, done in one take. Sometimes restraint is everything.', NOW() - interval '3 hours'),
  (1, 'The mural proposal for the metro station got approved! Colors picking up next week.', NOW() - interval '1 day'),
  (1, 'Character studies for the graphic novel. This cast keeps writing itself.', NOW() - interval '3 days'),
  (1, 'Inktober D20 — sailors, storms and a single lighthouse. Page 4 of 10.', NOW() - interval '6 days'),
  (1, 'First big canvas of the season is drying. Acrylics on linen, 4 days of work.', NOW() - interval '12 days');

-- 3) Post media
INSERT INTO post_media (post_id, object_key, media_type, sort_order)
SELECT p.id, 'https://picsum.photos/seed/ink' || p.id || '/900/900', 'image/jpeg', 0
FROM posts p WHERE p.user_id = 1
ON CONFLICT DO NOTHING;

-- 4) Likes (21 total)
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 1
)
INSERT INTO likes (user_id, parent_type, parent_id)
SELECT l.liker, 'post', op.id
FROM (VALUES
  (2, 1), (3, 1), (4, 1), (5, 1), (8, 1),   -- post 1: 5 likes
  (2, 2), (6, 2), (7, 2), (8, 2),            -- post 2: 4 likes
  (3, 3), (4, 3), (6, 3), (7, 3),            -- post 3: 4 likes
  (2, 4), (5, 4), (8, 4),                    -- post 4: 3 likes
  (2, 5), (3, 5), (4, 5), (5, 5), (7, 5)     -- post 5: 5 likes
) AS l(liker, ord)
JOIN ordered_posts op ON op.ord = l.ord
WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = l.liker)
  AND NOT EXISTS (
    SELECT 1 FROM likes e
    WHERE e.user_id = l.liker AND e.parent_type = 'post' AND e.parent_id = op.id
  );

-- 5) Comments (7 total)
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 1
)
INSERT INTO comments (user_id, parent_type, parent_id, content)
SELECT l.commenter, 'post', op.id, l.msg
FROM (VALUES
  (3, 1, 'That minimal line work is gorgeous.'),
  (5, 1, 'The fox’s eye says everything.'),
  (6, 2, 'Metro station murals deserve more love — great news!'),
  (4, 3, 'This cast has such strong voice.'),
  (7, 4, 'Sailor #2 is my favourite so far.'),
  (2, 5, 'Linen canvas for acrylics is a great combo.'),
  (8, 5, 'Would love to see this in person.')
) AS l(commenter, ord, msg)
JOIN ordered_posts op ON op.ord = l.ord
WHERE NOT EXISTS (
  SELECT 1 FROM comments e
  WHERE e.user_id = l.commenter AND e.parent_type = 'post' AND e.parent_id = op.id
    AND e.content = l.msg
);

-- 6) Followers / following
INSERT INTO user_follows (follower_id, following_id) VALUES
  (2, 1), (3, 1), (4, 1), (5, 1), (6, 1), (7, 1), (8, 1),   -- 7 followers
  (1, 2), (1, 3), (1, 7), (1, 8)                             -- follows 4 others
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- 7) Skills
INSERT INTO labels (label_name) VALUES
  ('Digital Art'), ('Illustration'), ('Public Art')
ON CONFLICT (label_name) DO NOTHING;

INSERT INTO user_labels (user_id, label_id)
SELECT 1, l.id FROM labels l WHERE l.label_name IN ('Digital Art', 'Illustration', 'Public Art')
ON CONFLICT (user_id, label_id) DO NOTHING;

-- 8) Achievements
INSERT INTO achievements (user_id, title, description, icon_type)
SELECT 1, t.title, t.description, t.icon_type
FROM (VALUES
  ('Top Creator', 'Ranked in the top 5% of creators this month', 'TOP_CREATOR'),
  ('Community Favourite', 'Reached 50 total likes across your artworks', 'FOLLOWERS'),
  ('Featured Artist', 'Your work was featured on the home feed', 'FEATURED')
) AS t(title, description, icon_type)
WHERE NOT EXISTS (
  SELECT 1 FROM achievements e
  WHERE e.user_id = 1 AND e.title = t.title
);

COMMIT;