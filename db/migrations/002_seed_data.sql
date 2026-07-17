-- +goose Up
-- +goose StatementBegin

-- Services
INSERT INTO services (num, name, name_html, description, duration_mins, price_pence, active) VALUES
  ('01', 'Precision Cut', 'Precision Cut',
   'Most booked. Skin fades, tapers and modern styles, finished clean and designed to sit well as it grows out.',
   30, 2500, true),
  ('02', 'Precision Cut + Beard', 'Precision Cut <span class="plus">+</span> Beard',
   'Skin fades, tapers and modern styles with beard work. Detailed, balanced finish — designed to sit well as it grows.',
   40, 3000, true),
  ('03', 'Classic + Beard', 'Classic <span class="plus">+</span> Beard',
   'A quick, simple cut with beard work included. Clipper grades or basic scissor work — no skin fades or tight tapers.',
   30, 2500, true),
  ('04', 'Classic Cut', 'Classic Cut',
   'Grades 0.5 — 8. A quick, simple, no fuss cut. Clipper with basic scissor work on top; for fades or tight tapers, please book a Precision Cut.',
   20, 2000, true),
  ('05', 'Under 16', 'Under 16',
   'Standard cuts only. Skin fades must be booked as a Precision Cut.',
   20, 1500, true),
  ('06', 'Senior Gent 65+', 'Senior Gent <span class="dot">·</span> 65+',
   'A standard cut for the regulars who''ve been sitting in chairs longer than we have.',
   20, 1500, true);

-- Nigel — founder
INSERT INTO users (id, email, password_hash, first_name, role)
VALUES (
  '11111111-1111-1111-1111-111111111101',
  'nigel@reformbarber.com',
  crypt('reform2024', gen_salt('bf', 12)),
  'Nigel',
  'barber'
);

INSERT INTO barbers (id, user_id, name, title, bio, num, active)
VALUES (
  '22222222-2222-2222-2222-222222222201',
  '11111111-1111-1111-1111-111111111101',
  'Nigel',
  'Founder',
  'Built the room, sets the standard. Carries the original following into the new chair.',
  '01',
  true
);

-- Nigel's weekly schedule (dow: 0=Sun, 1=Mon … 6=Sat)
INSERT INTO barber_schedules (barber_id, dow, open_time, close_time) VALUES
  ('22222222-2222-2222-2222-222222222201', 1, '09:30', '16:30'),
  ('22222222-2222-2222-2222-222222222201', 2, '09:00', '16:30'),
  ('22222222-2222-2222-2222-222222222201', 3, '09:00', '16:30'),
  ('22222222-2222-2222-2222-222222222201', 4, '09:30', '19:00'),
  ('22222222-2222-2222-2222-222222222201', 5, '09:00', '16:30'),
  ('22222222-2222-2222-2222-222222222201', 6, '07:30', '14:30');

-- Dev accounts matching .dev-passwords, so every role is testable locally.

-- Admin — site owner/administrator. No barber profile; books like a customer.
-- Keyed ON CONFLICT(email) so if this address was already registered through
-- the live app (e.g. the site owner signing up as a normal customer first),
-- we only promote its role to admin and leave the real password/name alone
-- rather than overwriting a live account with dev seed data.
INSERT INTO users (id, email, password_hash, first_name, role)
VALUES (
  '67827264-40fb-4b4b-ab52-a114e40a3dfd',
  'rc07jnr@gmail.com',
  crypt('reformAdmin', gen_salt('bf', 12)),
  'Admin',
  'admin'
)
ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role;

-- Founder dev account (distinct from Nigel, who is seeded with role 'barber'
-- above) so the founder dashboard has a real account to log in with.
INSERT INTO users (id, email, password_hash, first_name, last_name, role)
VALUES (
  '10b16b92-b811-47f1-9018-cb0518c4a17a',
  'nigel.jones@reformbarbers.com',
  crypt('reformFounder', gen_salt('bf', 12)),
  'Nigel',
  'Jones',
  'founder'
)
ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM barber_schedules WHERE barber_id = '22222222-2222-2222-2222-222222222201';
DELETE FROM barbers WHERE id = '22222222-2222-2222-2222-222222222201';
-- Note: if the admin row was an ON CONFLICT promotion of a pre-existing
-- account (different id, real password), it's intentionally left alone here.
DELETE FROM users WHERE id IN (
  '11111111-1111-1111-1111-111111111101',
  '67827264-40fb-4b4b-ab52-a114e40a3dfd',
  '10b16b92-b811-47f1-9018-cb0518c4a17a'
);
DELETE FROM services WHERE num IN ('01', '02', '03', '04', '05', '06');
-- +goose StatementEnd
