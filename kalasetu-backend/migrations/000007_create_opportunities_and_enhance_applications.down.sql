ALTER TABLE opportunities
    DROP COLUMN IF EXISTS event_id,
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS categories,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS total_positions,
    DROP COLUMN IF EXISTS status;

ALTER TABLE applications
    DROP COLUMN IF EXISTS event_id,
    DROP COLUMN IF EXISTS applicant_name,
    DROP COLUMN IF EXISTS applicant_email,
    DROP COLUMN IF EXISTS applicant_phone,
    DROP COLUMN IF EXISTS description;
