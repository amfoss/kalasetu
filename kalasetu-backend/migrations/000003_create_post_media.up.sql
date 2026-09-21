CREATE TABLE post_media (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    object_key VARCHAR(1024) NOT NULL,
    media_type VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_post_media_post_id ON post_media (post_id);

-- post_media is now the single source of truth for post media; the old
-- single-media columns on posts are obsolete.
ALTER TABLE posts
    DROP COLUMN IF EXISTS media_type,
    DROP COLUMN IF EXISTS media_uri;