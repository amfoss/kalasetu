ALTER TABLE events
    DROP COLUMN IF EXISTS banner_key,
    DROP COLUMN IF EXISTS banner_content_type,
    DROP COLUMN IF EXISTS banner_size;