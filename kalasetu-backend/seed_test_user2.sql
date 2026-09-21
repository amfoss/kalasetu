-- =====================================================================
-- Seed profile + social data for testuser2@gmail.com (id = 12) on local DB
-- Run: psql -h 127.0.0.1 -p 5433 -U postgres -d kalasetu_db -f seed_test_user2.sql
-- =====================================================================

BEGIN;

-- 1) Update the test user's basics
UPDATE users
   SET user_name = 'testuser2',
       bio = 'Photographer capturing little moments in big cities. Film + digital.',
       location = 'Bengaluru, India',
       profile_picture = 'https://i.pravatar.cc/300?img=32',
       updated_at = NOW()
 WHERE id = 12;

-- 2) Posts (artworks / recent posts), most recent first -> ord 1..5
INSERT INTO posts (user_id, content, created_at) VALUES
  (12, 'Golden hour through the glass — my favourite corner of the city right now. #streetphotography', NOW() - interval '3 hours'),
  (12, 'Film roll from the weekend finally developed. Grain so good it hurts.', NOW() - interval '2 days'),
  (12, 'Long exposure study. 30s, f/11, no ND filter — just patience.', NOW() - interval '4 days'),
  (12, 'Portrait session with @testuser for an upcoming project. Sneak peek only 😉', NOW() - interval '7 days'),
  (12, 'Rainy rooftop series, take three. Umbrellas in focus.', NOW() - interval '14 days');

-- 3) Post media
INSERT INTO post_media (post_id, object_key, media_type, sort_order)
SELECT p.id, 'https://picsum.photos/seed/photo' || p.id || '/900/900', 'image/jpeg', 0
FROM posts p WHERE p.user_id = 12
ON CONFLICT DO NOTHING;

-- 4) Likes from other (existing) users on the test user's posts
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 12
)
INSERT INTO likes (user_id, parent_type, parent_id)
SELECT l.liker, 'post', op.id
FROM (VALUES
  (1, 1), (2, 1), (3, 1), (4, 1), (5, 1),
  (1, 2), (3, 2), (5, 2),
  (2, 3), (4, 3), (5, 3),
  (1, 4), (3, 4),
  (2, 5), (4, 5), (5, 5)
) AS l(liker, ord)
JOIN ordered_posts op ON op.ord = l.ord
WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = l.liker)
  AND NOT EXISTS (
    SELECT 1 FROM likes e
    WHERE e.user_id = l.liker AND e.parent_type = 'post' AND e.parent_id = op.id
  );

-- 5) Comments
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 12
)
INSERT INTO comments (user_id, parent_type, parent_id, content)
SELECT l.commenter, 'post', op.id, l.msg
FROM (VALUES
  (1, 1, 'The colour grading here is so soft. Lovely.'),
  (3, 1, 'That glass reflection framing is perfect.'),
  (2, 2, 'What stock did you use? The grain looks amazing.'),
  (5, 2, 'Film days are the best days.'),
  (4, 3, '30 seconds of pure patience, worth it!'),
  (1, 4, 'Tagged! Can''''t wait for the final set.'),
  (3, 5, 'Umbrellas in focus — such a mood.')
) AS l(commenter, ord, msg)
JOIN ordered_posts op ON op.ord = l.ord
WHERE NOT EXISTS (
  SELECT 1 FROM comments e
  WHERE e.user_id = l.commenter AND e.parent_type = 'post' AND e.parent_id = op.id
    AND e.content = l.msg
);

-- 6) Followers / following
INSERT INTO user_follows (follower_id, following_id) VALUES
  (1, 12), (2, 12), (3, 12), (4, 12), (5, 12),   -- 5 followers
  (12, 1), (12, 2), (12, 3)                       -- follows 3 others
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- 7) Skills
INSERT INTO labels (label_name) VALUES
  ('Photography'), ('Street Photography'), ('Film Photography')
ON CONFLICT (label_name) DO NOTHING;

INSERT INTO user_labels (user_id, label_id)
SELECT 12, l.id FROM labels l WHERE l.label_name IN ('Photography', 'Street Photography', 'Film Photography')
ON CONFLICT (user_id, label_id) DO NOTHING;

-- 8) Achievements
INSERT INTO achievements (user_id, title, description, icon_type) VALUES
  (12, 'Featured Artist', 'Your work was featured on the home feed', 'FEATURED'),
  (12, 'Top Creator', 'Ranked in the top 5% of creators this month', 'TOP_CREATOR'),
  (12, 'Rising Star', 'Gained your first 50 followers', 'FOLLOWERS')
ON CONFLICT DO NOTHING;

COMMIT;