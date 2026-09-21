-- =====================================================================
-- Seed Roles, Labels, and Categories for Kalasetu
-- Run: PGPASSWORD=souri@123 psql -h kalasetu.postgres.database.azure.com -U kalasetuamfoss -d kalasetu_db -f seed_roles_labels_categories.sql
-- =====================================================================

BEGIN;

-- 1. Roles (used in onboarding role selection)
INSERT INTO roles (role) VALUES
  ('Artist'),
  ('Event Organizer'),
  ('Audience')
ON CONFLICT (role) DO NOTHING;

-- 2. Labels (used in onboarding artist category preferences / user labels)
INSERT INTO labels (label_name) VALUES
  ('Music'),
  ('Dance'),
  ('Theatre'),
  ('Stand-up Comedy'),
  ('Painting'),
  ('Photography'),
  ('Digital Art'),
  ('Graphic Design'),
  ('Live Shows'),
  ('Workshops'),
  ('Exhibitions'),
  ('Festivals'),
  ('Meetups'),
  ('Networking'),
  ('Local Communities'),
  ('Cultural Events'),
  ('Illustration'),
  ('Character Design'),
  ('Public Art'),
  ('Sculpture'),
  ('Ceramics')
ON CONFLICT (label_name) DO NOTHING;

-- 3. Categories (used for post and event classification)
INSERT INTO categories (category_name) VALUES
  ('Performing Arts'),
  ('Visual & Creative Arts'),
  ('Events & Experiences'),
  ('Social & Community'),
  ('Music'),
  ('Dance'),
  ('Theatre'),
  ('Stand-up Comedy'),
  ('Painting'),
  ('Photography'),
  ('Digital Art'),
  ('Graphic Design'),
  ('Live Shows'),
  ('Workshops'),
  ('Exhibitions'),
  ('Festivals'),
  ('Meetups'),
  ('Networking'),
  ('Local Communities'),
  ('Cultural Events'),
  ('Illustration'),
  ('Character Design')
ON CONFLICT (category_name) DO NOTHING;

COMMIT;
