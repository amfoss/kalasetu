-- =====================================================================
-- Seed profile + social data for the "Test User" (id = 1) on the local DB
-- Run: psql -h 127.0.0.1 -p 5433 -U postgres -d kalasetu_db -f seed_test_user.sql
-- =====================================================================

BEGIN;

-- 1) Update the test user's basics (real username, bio, location, avatar)
UPDATE users
   SET user_name = 'testuser',
       bio = 'Digital artist and illustrator. I draw characters that tell stories.',
       location = 'Kochi, India',
       profile_picture = 'https://i.pravatar.cc/300?img=12',
       updated_at = NOW()
 WHERE id = 1;

-- 2) Posts (these count as the profile's artworks / recent posts).
--    Order: most recent first -> ord 1..5.
INSERT INTO posts (user_id, content, created_at) VALUES
  (1, 'Just wrapped up this piece — a mercenary captain for an indie game. Took about 14 hours total. #characterdesign', NOW() - interval '2 hours'),
  (1, 'Work in progress: trying a warm palette for sunset lighting studies. C&C welcome!', NOW() - interval '1 day'),
  (1, 'Sketch dump from the weekend. Pencils, coffee, repeat. ✍️', NOW() - interval '3 days'),
  (1, 'Final render for a commission I finally shipped. The client loved it! 🙌', NOW() - interval '6 days'),
  (1, 'New brushes in the pack — painting foliage is about to get a lot less painful.', NOW() - interval '12 days');

-- 3) Post media (artwork images shown in the profile grid / feed)
INSERT INTO post_media (post_id, object_key, media_type, sort_order)
SELECT p.id, 'https://picsum.photos/seed/kala' || p.id || '/900/900', 'image/jpeg', 0
FROM posts p WHERE p.user_id = 1
ON CONFLICT DO NOTHING;

-- 4) Likes from other users on the test user's posts
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 1
)
INSERT INTO likes (user_id, parent_type, parent_id)
SELECT l.liker, 'post', op.id
FROM (VALUES
  (2, 1), (3, 1), (4, 1), (5, 1),
  (2, 2), (3, 2),
  (4, 3), (5, 3), (2, 3),
  (3, 4),
  (4, 5), (5, 5), (2, 5), (3, 5)
) AS l(liker, ord)
JOIN ordered_posts op ON op.ord = l.ord
WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = l.liker)
  AND NOT EXISTS (
    SELECT 1 FROM likes e
    WHERE e.user_id = l.liker AND e.parent_type = 'post' AND e.parent_id = op.id
  );

-- 5) Comments on the posts
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 1
)
INSERT INTO comments (user_id, parent_type, parent_id, content)
SELECT l.commenter, 'post', op.id, l.msg
FROM (VALUES
  (2, 1, 'The fabric movement in the cape is gorgeous!'),
  (3, 1, 'Would love to see the process timelapse.'),
  (4, 2, 'That warm light is coming together really nicely.'),
  (5, 2, 'The rim light sells it. Keep going!'),
  (2, 3, 'That first sketch is my favourite.'),
  (3, 4, 'Congrats on shipping it! Looks clean as always.')
) AS l(commenter, ord, msg)
JOIN ordered_posts op ON op.ord = l.ord
WHERE NOT EXISTS (
  SELECT 1 FROM comments e
  WHERE e.user_id = l.commenter AND e.parent_type = 'post' AND e.parent_id = op.id
    AND e.content = l.msg
);

-- 6) Followers / following (new user_follows table)
INSERT INTO user_follows (follower_id, following_id) VALUES
  (2, 1), (3, 1), (4, 1), (5, 1),   -- 4 followers of the test user
  (1, 2), (1, 3)                    -- test user follows 2 others
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- 7) Skills (reuse the labels system as the profile's skills list)
INSERT INTO labels (label_name) VALUES
  ('Digital Art'), ('Illustration'), ('Character Design')
ON CONFLICT (label_name) DO NOTHING;

INSERT INTO user_labels (user_id, label_id)
SELECT 1, l.id FROM labels l WHERE l.label_name IN ('Digital Art', 'Illustration', 'Character Design')
ON CONFLICT (user_id, label_id) DO NOTHING;

-- 8) Achievements
INSERT INTO achievements (user_id, title, description, icon_type) VALUES
  (1, 'Top Creator', 'Ranked in the top 5% of creators this month', 'TOP_CREATOR'),
  (1, 'Community Favourite', 'Reached 50 total likes across your artworks', 'FOLLOWERS'),
  (1, 'Featured Artist', 'Your artwork was featured on the home feed', 'FEATURED')
ON CONFLICT DO NOTHING;

COMMIT;