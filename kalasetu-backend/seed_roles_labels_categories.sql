-- This seed file initializes the essential master lookup data (user roles, artist labels, and event/post categories) required for onboarding, profiling, and content filtering across Kalasetu.

BEGIN;

INSERT INTO roles (role) VALUES
  ('Artist'),
  ('Event Organizer'),
  ('Audience')
ON CONFLICT (role) DO NOTHING;

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
