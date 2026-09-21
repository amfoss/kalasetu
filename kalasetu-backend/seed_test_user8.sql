-- =====================================================================
-- Seed profile + social data for navtest@test.com (id = 8) on local DB
-- Run: psql -h 127.0.0.1 -p 5433 -U postgres -d kalasetu_db -f seed_test_user8.sql
-- =====================================================================

BEGIN;

-- 1) Update the test user's basics
UPDATE users
   SET user_name = 'navtest',
       bio = 'Folk artist from Kerala. Mural practice, natural pigments and big walls.',
       location = 'Thrissur, India',
       profile_picture = 'https://i.pravatar.cc/300?img=47',
       updated_at = NOW()
 WHERE id = 8;

-- 2) Posts (artworks / recent posts), most recent first -> ord 1..5
INSERT INTO posts (user_id, content, created_at) VALUES
  (8, 'Mural study in the traditional style — mixing the tints for the base coat today.', NOW() - interval '5 hours'),
  (8, 'Finished a festival commission! Folk art panels for a wedding invitation.', NOW() - interval '2 days'),
  (8, 'Practising elephants again. The curve of the trunk is the hardest part.', NOW() - interval '5 days'),
  (8, 'Natural pigment tests — turmeric, indigo and clay on cotton cloth this week.', NOW() - interval '9 days'),
  (8, 'First ever public mural! 12ft wall right on the college campus 🎉', NOW() - interval '16 days');

-- 3) Post media
INSERT INTO post_media (post_id, object_key, media_type, sort_order)
SELECT p.id, 'https://picsum.photos/seed/mural' || p.id || '/900/900', 'image/jpeg', 0
FROM posts p WHERE p.user_id = 8
ON CONFLICT DO NOTHING;

-- 4) Likes
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 8
)
INSERT INTO likes (user_id, parent_type, parent_id)
SELECT l.liker, 'post', op.id
FROM (VALUES
  (1, 1), (2, 1), (3, 1), (12, 1),
  (1, 2), (4, 2), (5, 2), (12, 2),
  (2, 3), (3, 3), (5, 3),
  (1, 4), (4, 4), (12, 4),
  (2, 5), (3, 5), (5, 5), (12, 5)
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
  FROM posts WHERE user_id = 8
)
INSERT INTO comments (user_id, parent_type, parent_id, content)
SELECT l.commenter, 'post', op.id, l.msg
FROM (VALUES
  (1, 1, 'The base coat looks so rich already.'),
  (12, 1, 'Love that you mix the pigments yourself.'),
  (3, 2, 'Bold colour choices — festival energy!'),
  (5, 3, 'The trunk study is coming along nicely.'),
  (4, 4, 'Natural dyes are a whole different vibe.'),
  (12, 5, 'Waiting for the dedication photos 😄'),
  (2, 5, 'That wall deserved you.')
) AS l(commenter, ord, msg)
JOIN ordered_posts op ON op.ord = l.ord
WHERE NOT EXISTS (
  SELECT 1 FROM comments e
  WHERE e.user_id = l.commenter AND e.parent_type = 'post' AND e.parent_id = op.id
    AND e.content = l.msg
);

-- 6) Followers / following
INSERT INTO user_follows (follower_id, following_id) VALUES
  (1, 8), (2, 8), (3, 8), (4, 8), (5, 8), (12, 8),   -- 6 followers
  (8, 1), (8, 2), (8, 12)                              -- follows 3 others
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- 7) Skills
INSERT INTO labels (label_name) VALUES
  ('Folk Art'), ('Mural Painting'), ('Traditional Painting')
ON CONFLICT (label_name) DO NOTHING;

INSERT INTO user_labels (user_id, label_id)
SELECT 8, l.id FROM labels l WHERE l.label_name IN ('Folk Art', 'Mural Painting', 'Traditional Painting')
ON CONFLICT (user_id, label_id) DO NOTHING;

-- 8) Achievements
INSERT INTO achievements (user_id, title, description, icon_type) VALUES
  (8, 'Featured Artist', 'Your work was featured on the home feed', 'FEATURED'),
  (8, 'Community Favourite', 'Reached 50 total likes across your artworks', 'FOLLOWERS'),
  (8, 'Top Creator', 'Ranked in the top 5% of creators this month', 'TOP_CREATOR')
ON CONFLICT DO NOTHING;

COMMIT;