-- =====================================================================
-- Seed profile + social data for ftest@tester.com (id = 20) on local DB
-- Run: psql -h 127.0.0.1 -p 5433 -U postgres -d kalasetu_db -f seed_test_user20.sql
-- =====================================================================

BEGIN;

-- 1) Update the test user's basics
UPDATE users
   SET bio = 'Sculptor working in clay and stone. Currently obsessed with gesture and light.',
       location = 'Chennai, India',
       profile_picture = 'https://i.pravatar.cc/300?img=60',
       updated_at = NOW()
 WHERE id = 20;

-- 2) Reset previously seeded content so this script is safe to re-run
DELETE FROM achievements WHERE user_id = 20;
DELETE FROM likes     WHERE parent_type = 'post' AND parent_id IN (SELECT id FROM posts WHERE user_id = 20);
DELETE FROM comments  WHERE parent_type = 'post' AND parent_id IN (SELECT id FROM posts WHERE user_id = 20);
DELETE FROM post_media WHERE post_id IN (SELECT id FROM posts WHERE user_id = 20);
DELETE FROM posts WHERE user_id = 20;

-- 3) Posts (artworks / recent posts), most recent first -> ord 1..5
INSERT INTO posts (user_id, content, created_at) VALUES
  (20, 'Clay bust in progress — big gestures today. The shoulders finally snapped into place.', NOW() - interval '4 hours'),
  (20, 'Bronze cast update: releasing the mold was brutal, but the first pulls look clean.', NOW() - interval '1 day'),
  (20, 'Sketches before I commit anything to stone. Proportion > everything.', NOW() - interval '3 days'),
  (20, 'Group show next month — did the whole curation wall plan today. So excited.', NOW() - interval '7 days'),
  (20, 'First ever ceramic set is out of the kiln! Glazing day tomorrow ☕', NOW() - interval '13 days');

-- 4) Post media
INSERT INTO post_media (post_id, object_key, media_type, sort_order)
SELECT p.id, 'https://picsum.photos/seed/stone' || p.id || '/900/900', 'image/jpeg', 0
FROM posts p WHERE p.user_id = 20
ON CONFLICT DO NOTHING;

-- 5) Likes
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 20
)
INSERT INTO likes (user_id, parent_type, parent_id)
SELECT l.liker, 'post', op.id
FROM (VALUES
  (1, 1), (2, 1), (3, 1), (8, 1), (12, 1),
  (1, 2), (4, 2), (5, 2), (12, 2),
  (2, 3), (3, 3), (5, 3), (8, 3),
  (1, 4), (4, 4), (12, 4),
  (2, 5), (3, 5), (5, 5), (8, 5), (12, 5)
) AS l(liker, ord)
JOIN ordered_posts op ON op.ord = l.ord
WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = l.liker)
  AND NOT EXISTS (
    SELECT 1 FROM likes e
    WHERE e.user_id = l.liker AND e.parent_type = 'post' AND e.parent_id = op.id
  );

-- 6) Comments
WITH ordered_posts AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id DESC) AS ord
  FROM posts WHERE user_id = 20
)
INSERT INTO comments (user_id, parent_type, parent_id, content)
SELECT l.commenter, 'post', op.id, l.msg
FROM (VALUES
  (1, 1, 'That gesture is alive. Great work.'),
  (12, 1, 'The shoulders make the whole stance.'),
  (3, 2, 'Mold release day is always a gamble! Glad it paid off.'),
  (5, 3, 'Patience wins with stone, always.'),
  (8, 4, 'Would love the exhibition timings once confirmed.'),
  (2, 5, 'Ceramics era begins! 🎨'),
  (4, 5, 'The glaze combinations look promising.')
) AS l(commenter, ord, msg)
JOIN ordered_posts op ON op.ord = l.ord
WHERE NOT EXISTS (
  SELECT 1 FROM comments e
  WHERE e.user_id = l.commenter AND e.parent_type = 'post' AND e.parent_id = op.id
    AND e.content = l.msg
);

-- 7) Followers / following
INSERT INTO user_follows (follower_id, following_id) VALUES
  (1, 20), (2, 20), (3, 20), (4, 20), (5, 20), (8, 20), (12, 20),   -- 7 followers
  (20, 1), (20, 2), (20, 8), (20, 12)                                 -- follows 4 others
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- 8) Skills
INSERT INTO labels (label_name) VALUES
  ('Sculpture'), ('Ceramics'), ('Public Art')
ON CONFLICT (label_name) DO NOTHING;

INSERT INTO user_labels (user_id, label_id)
SELECT 20, l.id FROM labels l WHERE l.label_name IN ('Sculpture', 'Ceramics', 'Public Art')
ON CONFLICT (user_id, label_id) DO NOTHING;

-- 9) Achievements
INSERT INTO achievements (user_id, title, description, icon_type)
SELECT 20, t.title, t.description, t.icon_type
FROM (VALUES
  ('Top Creator', 'Ranked in the top 5% of creators this month', 'TOP_CREATOR'),
  ('Community Favourite', 'Reached 50 total likes across your artworks', 'FOLLOWERS'),
  ('Featured Artist', 'Your work was featured on the home feed', 'FEATURED')
) AS t(title, description, icon_type)
WHERE NOT EXISTS (
  SELECT 1 FROM achievements e
  WHERE e.user_id = 20 AND e.title = t.title
);

COMMIT;