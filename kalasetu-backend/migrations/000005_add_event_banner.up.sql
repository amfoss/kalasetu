-- An event can have one optional banner image. Only the S3 object key and its
-- metadata are stored in PostgreSQL; the image bytes live in object storage.
ALTER TABLE events
    ADD COLUMN banner_key TEXT,
    ADD COLUMN banner_content_type VARCHAR(100),
    ADD COLUMN banner_size BIGINT;